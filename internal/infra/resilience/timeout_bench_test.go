package resilience

import (
	"context"
	"testing"
	"time"
)

// Timeout Benchmark Tests
//
// Performance Targets (NFR-PERF):
// - Setup overhead: <100ns per operation
//
// Baseline Results (Apple M1):
// - BenchmarkTimeout_Setup: ~36ns/op, 64 B/op, 2 allocs/op
// - BenchmarkTimeout_Do_Success: ~450-600ns/op, ~306 B/op, 7 allocs/op
// - BenchmarkTimeoutPresets_ForDatabase: ~0.3ns/op, 0 B/op, 0 allocs/op (cached access)
// - BenchmarkTimeoutPresets_ForOperation: ~35ns/op, 64 B/op, 2 allocs/op (creates new timeout)
// - BenchmarkDoWithTimeout: ~450ns/op, ~352 B/op, 9 allocs/op
//
// Note: Benchmark results may vary across platforms and Go versions.

// BenchmarkTimeout_Setup measures the overhead of creating a timeout.
func BenchmarkTimeout_Setup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = NewTimeout("bench", 5*time.Second)
	}
}

// BenchmarkTimeout_Do_Success measures the overhead of a successful timeout operation.
func BenchmarkTimeout_Do_Success(b *testing.B) {
	to := NewTimeout("bench", 5*time.Second)
	ctx := context.Background()
	fn := func(ctx context.Context) error { return nil }

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = to.Do(ctx, fn)
	}
}

// BenchmarkTimeout_Do_Success_WithMetrics measures overhead with metrics enabled.
func BenchmarkTimeout_Do_Success_WithMetrics(b *testing.B) {
	metrics := NoopTimeoutMetrics()
	to := NewTimeout("bench", 5*time.Second, WithTimeoutMetrics(metrics))
	ctx := context.Background()
	fn := func(ctx context.Context) error { return nil }

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = to.Do(ctx, fn)
	}
}

// BenchmarkTimeoutPresets_ForDatabase measures preset access overhead.
func BenchmarkTimeoutPresets_ForDatabase(b *testing.B) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}
	presets := NewTimeoutPresets(cfg)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = presets.ForDatabase()
	}
}

// BenchmarkTimeoutPresets_ForOperation measures custom timeout creation overhead.
func BenchmarkTimeoutPresets_ForOperation(b *testing.B) {
	cfg := TimeoutConfig{
		Default:     30 * time.Second,
		Database:    5 * time.Second,
		ExternalAPI: 10 * time.Second,
	}
	presets := NewTimeoutPresets(cfg)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = presets.ForOperation("custom", 1*time.Second)
	}
}

// BenchmarkDoWithTimeout measures generic helper overhead.
func BenchmarkDoWithTimeout(b *testing.B) {
	to := NewTimeout("bench", 5*time.Second)
	ctx := context.Background()
	fn := func(ctx context.Context) (int, error) { return 42, nil }

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = DoWithTimeout[int](to, ctx, fn)
	}
}
