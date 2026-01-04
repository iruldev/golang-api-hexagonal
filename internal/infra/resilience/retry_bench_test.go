package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

// BenchmarkRetrier_Setup measures the overhead of creating a new retrier.
// Target: <10ns per setup (though this is the setup creation, not the backoff setup).
func BenchmarkRetrier_Setup(b *testing.B) {
	cfg := DefaultRetryConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewRetrier("bench", cfg)
	}
}

// BenchmarkRetrier_Do_Success measures the overhead of a successful operation with no retry.
func BenchmarkRetrier_Do_Success(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
	r := NewRetrier("bench-success", cfg)
	ctx := context.Background()
	fn := func(ctx context.Context) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Do(ctx, fn)
	}
}

// BenchmarkRetrier_Do_OneRetry measures the overhead of an operation that succeeds after one retry.
// This includes the backoff delay, so actual time will be dominated by sleep.
func BenchmarkRetrier_Do_OneRetry(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Microsecond, // Very short delay for benchmark
		MaxDelay:     10 * time.Microsecond,
		Multiplier:   2.0,
	}
	r := NewRetrier("bench-retry", cfg)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempt := 0
		fn := func(ctx context.Context) error {
			attempt++
			if attempt < 2 {
				return errors.New("transient")
			}
			return nil
		}
		_ = r.Do(ctx, fn)
	}
}

// BenchmarkRetrier_Do_Success_WithMetrics measures overhead with metrics enabled.
func BenchmarkRetrier_Do_Success_WithMetrics(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
	metrics := NoopRetryMetrics()
	r := NewRetrier("bench-metrics", cfg, WithRetryMetrics(metrics))
	ctx := context.Background()
	fn := func(ctx context.Context) error { return nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Do(ctx, fn)
	}
}

// BenchmarkDefaultIsRetryable measures the overhead of error classification.
func BenchmarkDefaultIsRetryable(b *testing.B) {
	err := errors.New("some error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultIsRetryable(err)
	}
}

// BenchmarkDefaultIsRetryable_ContextCanceled measures context.Canceled check overhead.
func BenchmarkDefaultIsRetryable_ContextCanceled(b *testing.B) {
	err := context.Canceled

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultIsRetryable(err)
	}
}

// BenchmarkDefaultIsRetryable_DeadlineExceeded measures context.DeadlineExceeded check overhead.
func BenchmarkDefaultIsRetryable_DeadlineExceeded(b *testing.B) {
	err := context.DeadlineExceeded

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DefaultIsRetryable(err)
	}
}

// BenchmarkDoWithResult measures the generic wrapper overhead.
func BenchmarkDoWithResult(b *testing.B) {
	cfg := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
	r := NewRetrier("bench-result", cfg)
	ctx := context.Background()
	fn := func(ctx context.Context) (int, error) { return 42, nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DoWithResult(r, ctx, fn)
	}
}
