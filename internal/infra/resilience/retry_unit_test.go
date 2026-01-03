package resilience

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Basic retry logic tests.

func TestRetrier_Do(t *testing.T) {
	tests := []struct {
		name         string
		cfg          RetryConfig
		failCount    int // how many times fn will fail before succeeding
		wantAttempts int // expected total attempts
		wantErr      error
	}{
		{
			name:         "succeeds on first attempt",
			cfg:          DefaultRetryConfig(),
			failCount:    0,
			wantAttempts: 1,
			wantErr:      nil,
		},
		{
			name:         "succeeds after 1 retry",
			cfg:          DefaultRetryConfig(),
			failCount:    1,
			wantAttempts: 2,
			wantErr:      nil,
		},
		{
			name:         "succeeds after 2 retries",
			cfg:          DefaultRetryConfig(),
			failCount:    2,
			wantAttempts: 3,
			wantErr:      nil,
		},
		{
			name: "exhausts max attempts",
			cfg: RetryConfig{
				MaxAttempts:  3,
				InitialDelay: 1 * time.Millisecond,
				MaxDelay:     10 * time.Millisecond,
				Multiplier:   2.0,
			},
			failCount:    10, // always fail
			wantAttempts: 3,
			wantErr:      ErrMaxRetriesExceeded,
		},
		{
			name: "single attempt allowed",
			cfg: RetryConfig{
				MaxAttempts:  1,
				InitialDelay: 1 * time.Millisecond,
				MaxDelay:     10 * time.Millisecond,
				Multiplier:   2.0,
			},
			failCount:    1,
			wantAttempts: 1,
			wantErr:      ErrMaxRetriesExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRetrier("test", tt.cfg)
			ctx := context.Background()

			var attempts int32
			fn := func(ctx context.Context) error {
				attempt := atomic.AddInt32(&attempts, 1)
				if int(attempt) <= tt.failCount {
					return errors.New("transient error")
				}
				return nil
			}

			err := r.Do(ctx, fn)

			assert.Equal(t, tt.wantAttempts, int(atomic.LoadInt32(&attempts)))
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRetrier_Do_NonRetryableError(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	// Use custom retryable function that returns false
	r := NewRetrier("test-nonretryable", cfg, WithRetryableFunc(func(err error) bool {
		return false // Never retry
	}))
	ctx := context.Background()

	var attempts int32
	fn := func(ctx context.Context) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("permanent error")
	}

	err := r.Do(ctx, fn)

	// Should stop after first attempt with non-retryable error
	assert.Equal(t, 1, int(atomic.LoadInt32(&attempts)))
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrMaxRetriesExceeded)
}

func TestRetrier_Do_CustomRetryableFunc(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	specificErr := errors.New("specific retryable error")
	otherErr := errors.New("other error")

	// Custom function that only retries specific error
	r := NewRetrier("test-custom", cfg, WithRetryableFunc(func(err error) bool {
		return errors.Is(err, specificErr)
	}))
	ctx := context.Background()

	t.Run("retries specific error", func(t *testing.T) {
		var attempts int32
		fn := func(ctx context.Context) error {
			attempt := atomic.AddInt32(&attempts, 1)
			if int(attempt) < 3 {
				return specificErr
			}
			return nil
		}

		err := r.Do(ctx, fn)
		assert.NoError(t, err)
		assert.Equal(t, 3, int(atomic.LoadInt32(&attempts)))
	})

	t.Run("does not retry other error", func(t *testing.T) {
		var attempts int32
		fn := func(ctx context.Context) error {
			atomic.AddInt32(&attempts, 1)
			return otherErr
		}

		err := r.Do(ctx, fn)
		assert.Error(t, err)
		assert.Equal(t, 1, int(atomic.LoadInt32(&attempts)))
	})
}

func TestRetrier_Name(t *testing.T) {
	r := NewRetrier("my-retrier", DefaultRetryConfig())
	assert.Equal(t, "my-retrier", r.Name())
}

func TestRetrier_WithLogger(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	logger := slog.Default()
	r := NewRetrier("test-logger", cfg, WithRetryLogger(logger))
	ctx := context.Background()

	var attempts int32
	fn := func(ctx context.Context) error {
		attempt := atomic.AddInt32(&attempts, 1)
		if int(attempt) < 2 {
			return errors.New("transient")
		}
		return nil
	}

	err := r.Do(ctx, fn)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(atomic.LoadInt32(&attempts)))
}

func TestRetrier_WithNilLogger(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	// Should not panic with nil logger - uses default
	r := NewRetrier("test-nil-logger", cfg, WithRetryLogger(nil))
	ctx := context.Background()

	var attempts int32
	fn := func(ctx context.Context) error {
		attempt := atomic.AddInt32(&attempts, 1)
		if int(attempt) < 2 {
			return errors.New("transient")
		}
		return nil
	}

	err := r.Do(ctx, fn)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(atomic.LoadInt32(&attempts)))
}

func TestRetrier_WithNilRetryableFunc(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	// Should not panic with nil retryable func - uses DefaultIsRetryable
	r := NewRetrier("test-nil-retryable", cfg, WithRetryableFunc(nil))
	ctx := context.Background()

	var attempts int32
	fn := func(ctx context.Context) error {
		attempt := atomic.AddInt32(&attempts, 1)
		if int(attempt) < 2 {
			return errors.New("transient")
		}
		return nil
	}

	err := r.Do(ctx, fn)
	assert.NoError(t, err)
	assert.Equal(t, 2, int(atomic.LoadInt32(&attempts)))
}

func TestDefaultIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error is not retryable",
			err:      nil,
			expected: false,
		},
		{
			name:     "context.Canceled is not retryable",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "context.DeadlineExceeded is retryable",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "wrapped context.Canceled is not retryable",
			err:      errors.Join(errors.New("wrap"), context.Canceled),
			expected: false,
		},
		{
			name:     "generic error is retryable by default",
			err:      errors.New("some error"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DefaultIsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsRetryable(t *testing.T) {
	// IsRetryable is just an alias for DefaultIsRetryable
	assert.False(t, IsRetryable(nil))
	assert.False(t, IsRetryable(context.Canceled))
	assert.True(t, IsRetryable(context.DeadlineExceeded))
	assert.True(t, IsRetryable(errors.New("some error")))
}

func TestDefaultRetryConfig(t *testing.T) {
	cfg := DefaultRetryConfig()

	assert.Equal(t, DefaultRetryMaxAttempts, cfg.MaxAttempts)
	assert.Equal(t, DefaultRetryInitialDelay, cfg.InitialDelay)
	assert.Equal(t, DefaultRetryMaxDelay, cfg.MaxDelay)
	assert.Equal(t, DefaultRetryMultiplier, cfg.Multiplier)
}

// Custom retryable error for testing
type retryableErr struct {
	retryable bool
	msg       string
}

func (e *retryableErr) Error() string {
	return e.msg
}

func (e *retryableErr) Retryable() bool {
	return e.retryable
}

func TestDefaultIsRetryable_RetryableInterface(t *testing.T) {
	t.Run("respects Retryable() true", func(t *testing.T) {
		err := &retryableErr{retryable: true, msg: "retryable error"}
		assert.True(t, DefaultIsRetryable(err))
	})

	t.Run("respects Retryable() false", func(t *testing.T) {
		err := &retryableErr{retryable: false, msg: "non-retryable error"}
		assert.False(t, DefaultIsRetryable(err))
	})
}

// Custom temporary error for testing
type tempErr struct {
	temp bool
	msg  string
}

func (e *tempErr) Error() string {
	return e.msg
}

func (e *tempErr) Temporary() bool {
	return e.temp
}

func TestDefaultIsRetryable_TemporaryInterface(t *testing.T) {
	t.Run("respects Temporary() true", func(t *testing.T) {
		err := &tempErr{temp: true, msg: "temporary error"}
		assert.True(t, DefaultIsRetryable(err))
	})

	t.Run("respects Temporary() false", func(t *testing.T) {
		err := &tempErr{temp: false, msg: "permanent error"}
		assert.False(t, DefaultIsRetryable(err))
	})
}

func TestNewRetryMetrics_NilRegistry(t *testing.T) {
	// Should not panic with nil registry
	metrics := NewRetryMetrics(nil)
	assert.NotNil(t, metrics)
	assert.NotNil(t, metrics.operationTotal)
	assert.NotNil(t, metrics.attemptTotal)
	assert.NotNil(t, metrics.durationSeconds)
}

func TestRetryMetrics_RecordOperation(t *testing.T) {
	metrics := NewRetryMetrics(prometheus.NewRegistry())

	// Should not panic
	metrics.RecordOperation("test", "success", 1, 0.1)
	metrics.RecordOperation("test", "failure", 2, 0.2)
	metrics.RecordOperation("test", "exhausted", 3, 0.3)
}

func TestRetryMetrics_Reset(t *testing.T) {
	metrics := NewRetryMetrics(prometheus.NewRegistry())
	metrics.RecordOperation("test", "success", 1, 0.1)

	// Should not panic
	metrics.Reset()
}

func TestNoopRetryMetrics(t *testing.T) {
	metrics := NoopRetryMetrics()
	assert.NotNil(t, metrics)

	// Should not panic when recording
	metrics.RecordOperation("test", "success", 1, 0.1)
	metrics.Reset()
}

func Test_itoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{123, "123"},
		{-1, "-1"},
		{-42, "-42"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := itoa(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetrier_Do_Metrics(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	metrics := NewRetryMetrics(prometheus.NewRegistry())
	r := NewRetrier("test-metrics", cfg, WithRetryMetrics(metrics))
	ctx := context.Background()

	t.Run("records success metrics", func(t *testing.T) {
		metrics.Reset()
		var attempts int32
		fn := func(ctx context.Context) error {
			attempt := atomic.AddInt32(&attempts, 1)
			if int(attempt) < 2 {
				return errors.New("transient")
			}
			return nil
		}

		err := r.Do(ctx, fn)
		require.NoError(t, err)
		// Metrics should be recorded - we can't easily assert values without
		// exposing the metrics internals, but we verify no panic occurred
	})

	t.Run("records exhausted metrics", func(t *testing.T) {
		metrics.Reset()
		fn := func(ctx context.Context) error {
			return errors.New("always fail")
		}

		err := r.Do(ctx, fn)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrMaxRetriesExceeded)
	})
}
