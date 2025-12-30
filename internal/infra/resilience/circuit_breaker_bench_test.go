package resilience

import (
	"context"
	"errors"
	"testing"
	"time"
)

// BenchmarkCircuitBreaker_Execute_Closed measures the overhead of executing
// an operation through a closed circuit breaker.
// Target: <1μs overhead per operation
func BenchmarkCircuitBreaker_Execute_Closed(b *testing.B) {
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig())
	ctx := context.Background()
	fn := func() (any, error) { return nil, nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cb.Execute(ctx, fn)
	}
}

// BenchmarkCircuitBreaker_Execute_Closed_WithMetrics measures overhead with metrics enabled.
func BenchmarkCircuitBreaker_Execute_Closed_WithMetrics(b *testing.B) {
	metrics := NoopCircuitBreakerMetrics()
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig(), WithMetrics(metrics))
	ctx := context.Background()
	fn := func() (any, error) { return nil, nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cb.Execute(ctx, fn)
	}
}

// BenchmarkCircuitBreaker_Execute_Open measures the overhead of rejecting
// requests when the circuit is open.
func BenchmarkCircuitBreaker_Execute_Open(b *testing.B) {
	cfg := CircuitBreakerConfig{
		MaxRequests:      1,
		Interval:         10 * time.Second,
		Timeout:          24 * time.Hour, // Long timeout to keep circuit open during benchmark
		FailureThreshold: 1,
	}
	cb := NewCircuitBreaker("bench", cfg)
	ctx := context.Background()

	// Trip the circuit
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("fail")
	})

	fn := func() (any, error) { return nil, nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cb.Execute(ctx, fn)
	}
}

// BenchmarkCircuitBreaker_State measures the overhead of checking the circuit state.
func BenchmarkCircuitBreaker_State(b *testing.B) {
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.State()
	}
}

// BenchmarkCircuitBreaker_Name measures the overhead of getting the circuit name.
func BenchmarkCircuitBreaker_Name(b *testing.B) {
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cb.Name()
	}
}

// BenchmarkCircuitBreaker_Execute_Parallel measures concurrent execution performance.
func BenchmarkCircuitBreaker_Execute_Parallel(b *testing.B) {
	cb := NewCircuitBreaker("bench", DefaultCircuitBreakerConfig())
	ctx := context.Background()
	fn := func() (any, error) { return nil, nil }

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = cb.Execute(ctx, fn)
		}
	})
}

// BenchmarkCircuitBreakerMetrics_SetState measures the overhead of updating state metrics.
func BenchmarkCircuitBreakerMetrics_SetState(b *testing.B) {
	metrics := NoopCircuitBreakerMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.SetState("bench", 0)
	}
}

// BenchmarkCircuitBreakerMetrics_RecordTransition measures the overhead of recording transitions.
func BenchmarkCircuitBreakerMetrics_RecordTransition(b *testing.B) {
	metrics := NoopCircuitBreakerMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordTransition("bench", "closed", "open")
	}
}

// BenchmarkCircuitBreakerMetrics_RecordOperationDuration measures the overhead of recording durations.
func BenchmarkCircuitBreakerMetrics_RecordOperationDuration(b *testing.B) {
	metrics := NoopCircuitBreakerMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordOperationDuration("bench", "success", 0.001)
	}
}
