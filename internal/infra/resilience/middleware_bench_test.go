package resilience

import (
	"context"
	"testing"
	"time"
)

// BenchmarkResilienceWrapper_FullComposition benchmarks the overhead
// of the full resilience composition (CB → Retry → Timeout → Bulkhead).
func BenchmarkResilienceWrapper_FullComposition(b *testing.B) {
	// Create all components with real implementations
	cbConfig := DefaultCircuitBreakerConfig()
	cbPresets := NewCircuitBreakerPresets(cbConfig)

	retryConfig := DefaultRetryConfig()
	retryConfig.MaxAttempts = 1 // No actual retries for benchmark
	retrier := NewRetrier("benchmark", retryConfig)

	timeoutPresets := NewTimeoutPresets(TimeoutConfig{
		Default:     5 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	})

	bulkheadPresets := NewBulkheadPresets(BulkheadConfig{
		MaxConcurrent: 100,
		MaxWaiting:    100,
	})

	wrapper := NewResilienceWrapper(
		WithCircuitBreakerFactory(cbPresets.Factory()),
		WithWrapperRetrier(retrier),
		WithWrapperTimeout(timeoutPresets.Default()),
		WithWrapperBulkhead(bulkheadPresets.Default()),
	)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wrapper.Execute(ctx, "bench-op", func(ctx context.Context) error {
			return nil // Instant success
		})
	}
}

// BenchmarkResilienceWrapper_NoComponents benchmarks baseline without
// any resilience components.
func BenchmarkResilienceWrapper_NoComponents(b *testing.B) {
	wrapper := NewResilienceWrapper() // No components

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = wrapper.Execute(ctx, "bench-op", func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkCircuitBreaker_Execute benchmarks circuit breaker overhead.
func BenchmarkCircuitBreaker_Execute(b *testing.B) {
	cb := NewCircuitBreaker("benchmark", DefaultCircuitBreakerConfig())
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = cb.Execute(ctx, func() (any, error) {
			return nil, nil
		})
	}
}

// BenchmarkRetrier_Do benchmarks retrier overhead with single attempt.
func BenchmarkRetrier_Do(b *testing.B) {
	cfg := DefaultRetryConfig()
	cfg.MaxAttempts = 1 // Single attempt, no retries
	retrier := NewRetrier("benchmark", cfg)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = retrier.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkTimeout_Do benchmarks timeout wrapper overhead.
func BenchmarkTimeout_Do(b *testing.B) {
	timeout := NewTimeout("benchmark", 5*time.Second)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = timeout.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkBulkhead_Do benchmarks bulkhead overhead.
func BenchmarkBulkhead_Do(b *testing.B) {
	bulkhead := NewBulkhead("benchmark", BulkheadConfig{
		MaxConcurrent: 100,
		MaxWaiting:    100,
	})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = bulkhead.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkCircuitBreakerFactory benchmarks factory pattern overhead.
func BenchmarkCircuitBreakerFactory(b *testing.B) {
	factory := NewCircuitBreakerFactory(DefaultCircuitBreakerConfig())

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = factory("benchmark-op")
	}
}

// BenchmarkResilienceWrapper_Parallel benchmarks concurrent execution.
func BenchmarkResilienceWrapper_Parallel(b *testing.B) {
	cbPresets := NewCircuitBreakerPresets(DefaultCircuitBreakerConfig())
	retrier := NewRetrier("parallel", DefaultRetryConfig())
	timeoutPresets := NewTimeoutPresets(TimeoutConfig{
		Default:     5 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	})
	bulkheadPresets := NewBulkheadPresets(BulkheadConfig{
		MaxConcurrent: 100,
		MaxWaiting:    100,
	})

	wrapper := NewResilienceWrapper(
		WithCircuitBreakerFactory(cbPresets.Factory()),
		WithWrapperRetrier(retrier),
		WithWrapperTimeout(timeoutPresets.Default()),
		WithWrapperBulkhead(bulkheadPresets.Default()),
	)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			_ = wrapper.Execute(ctx, "parallel-op", func(ctx context.Context) error {
				return nil
			})
		}
	})
}
