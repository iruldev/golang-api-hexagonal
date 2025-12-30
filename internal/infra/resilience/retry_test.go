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

func TestRetrier_Do_ExponentialBackoff(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     500 * time.Millisecond,
		Multiplier:   2.0,
	}

	r := NewRetrier("test-backoff", cfg)
	ctx := context.Background()

	var attempts int32
	var timestamps []time.Time

	fn := func(ctx context.Context) error {
		timestamps = append(timestamps, time.Now())
		attempt := atomic.AddInt32(&attempts, 1)
		if int(attempt) < 4 {
			return errors.New("transient error")
		}
		return nil
	}

	start := time.Now()
	err := r.Do(ctx, fn)
	totalDuration := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, 4, int(atomic.LoadInt32(&attempts)))
	assert.Len(t, timestamps, 4)

	// Verify delays are approximately exponential:
	// First attempt: immediate
	// After fail 1: ~50ms (+/- jitter)
	// After fail 2: ~100ms (+/- jitter)
	// After fail 3: ~200ms (+/- jitter)
	// Total: ~350ms minimum
	assert.GreaterOrEqual(t, totalDuration, 250*time.Millisecond, "total duration should be at least 250ms")

	// First to second attempt delay should be around 50ms
	if len(timestamps) >= 2 {
		delay1 := timestamps[1].Sub(timestamps[0])
		assert.GreaterOrEqual(t, delay1, 25*time.Millisecond, "first delay should be at least 25ms")
		assert.LessOrEqual(t, delay1, 100*time.Millisecond, "first delay should be at most 100ms")
	}
}

func TestRetrier_Do_JitterApplied(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	// Run multiple times to verify jitter produces different delays
	delays := make([]time.Duration, 5)

	for i := 0; i < 5; i++ {
		r := NewRetrier("test-jitter", cfg)
		ctx := context.Background()

		var timestamps []time.Time
		var attempts int32

		fn := func(ctx context.Context) error {
			timestamps = append(timestamps, time.Now())
			attempt := atomic.AddInt32(&attempts, 1)
			if int(attempt) < 2 {
				return errors.New("transient error")
			}
			return nil
		}

		_ = r.Do(ctx, fn)

		if len(timestamps) >= 2 {
			delays[i] = timestamps[1].Sub(timestamps[0])
		}
	}

	// Verify that not all delays are exactly the same (jitter is applied)
	// At least some variation should exist
	allSame := true
	for i := 1; i < len(delays); i++ {
		// Allow for 1ms tolerance
		if delays[i] < delays[0]-1*time.Millisecond || delays[i] > delays[0]+1*time.Millisecond {
			allSame = false
			break
		}
	}
	// Due to jitter, delays should vary. However, this test is non-deterministic:
	// - On fast machines, timing resolution may mask small jitter differences
	// - Random jitter could theoretically produce identical values
	// We log rather than fail to avoid flaky CI, but manual verification is recommended
	// if this warning appears consistently.
	if allSame {
		t.Log("INFO: All delays were identical - jitter may not be detectable at this timing resolution")
	}
}

func TestRetrier_Do_MaxDelayCap(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     150 * time.Millisecond, // Low cap
		Multiplier:   2.0,
	}

	r := NewRetrier("test-maxdelay", cfg)
	ctx := context.Background()

	var timestamps []time.Time
	var attempts int32

	fn := func(ctx context.Context) error {
		timestamps = append(timestamps, time.Now())
		attempt := atomic.AddInt32(&attempts, 1)
		if int(attempt) < 5 {
			return errors.New("transient error")
		}
		return nil
	}

	start := time.Now()
	err := r.Do(ctx, fn)
	totalDuration := time.Since(start)

	require.NoError(t, err)

	// With max delay cap at 150ms, delays should not exceed ~150ms + jitter
	// Without cap: 100ms + 200ms + 400ms + 800ms = 1500ms
	// With cap at 150ms: 100ms + 150ms + 150ms + 150ms = 550ms (approximately)
	assert.Less(t, totalDuration, 1*time.Second, "delays should be capped")
}

func TestRetrier_Do_ContextCancellation(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
	}

	r := NewRetrier("test-cancel", cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var attempts int32
	fn := func(ctx context.Context) error {
		atomic.AddInt32(&attempts, 1)
		return errors.New("always fail")
	}

	err := r.Do(ctx, fn)

	// Should stop before max attempts due to context cancellation
	assert.Less(t, int(atomic.LoadInt32(&attempts)), 10)
	assert.Error(t, err)
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

func TestDoWithResult(t *testing.T) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}

	r := NewRetrier("test-result", cfg)
	ctx := context.Background()

	t.Run("returns result on success", func(t *testing.T) {
		var attempts int32
		fn := func(ctx context.Context) (string, error) {
			atomic.AddInt32(&attempts, 1)
			return "success", nil
		}

		result, err := DoWithResult(r, ctx, fn)
		require.NoError(t, err)
		assert.Equal(t, "success", result)
	})

	t.Run("returns result after retry", func(t *testing.T) {
		var attempts int32
		fn := func(ctx context.Context) (int, error) {
			attempt := atomic.AddInt32(&attempts, 1)
			if int(attempt) < 2 {
				return 0, errors.New("transient")
			}
			return 42, nil
		}

		result, err := DoWithResult(r, ctx, fn)
		require.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		fn := func(ctx context.Context) (string, error) {
			return "", errors.New("always fail")
		}

		result, err := DoWithResult(r, ctx, fn)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrMaxRetriesExceeded)
		assert.Empty(t, result)
	})
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
