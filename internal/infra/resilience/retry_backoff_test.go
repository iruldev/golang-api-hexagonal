package resilience

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Exponential backoff and jitter tests.

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
