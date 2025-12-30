package resilience

import (
	"context"
	"testing"
)

// BenchmarkBulkhead_SlotAcquisition measures the overhead of slot acquisition
// with no contention.
//
// Target performance: <500ns per operation
//
// Typical results (platform-dependent, will vary based on CPU and system load):
// - Apple M1: ~140-160ns/op with no contention (immediate slot acquisition)
// - Results may vary significantly on different platforms
// - 1 allocs/op (closure allocation)
func BenchmarkBulkhead_SlotAcquisition(b *testing.B) {
	bh := NewBulkhead("bench", BulkheadConfig{MaxConcurrent: 100, MaxWaiting: 100})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = bh.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkBulkhead_WithMetrics measures overhead with metrics enabled.
//
// Typical results (platform-dependent, will vary based on CPU and system load):
// - Apple M1: ~250-300ns/op with metrics (gauge/counter updates)
// - Results may vary significantly on different platforms
// - 1 allocs/op
func BenchmarkBulkhead_WithMetrics(b *testing.B) {
	metrics := NoopBulkheadMetrics()
	bh := NewBulkhead("bench", BulkheadConfig{MaxConcurrent: 100, MaxWaiting: 100},
		WithBulkheadMetrics(metrics))
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = bh.Do(ctx, func(ctx context.Context) error {
			return nil
		})
	}
}

// BenchmarkBulkhead_ConcurrentLowContention tests concurrent access with
// plenty of capacity.
//
// Typical results (platform-dependent, will vary based on CPU and system load):
// - Apple M1: ~200-250ns/op per goroutine
// - Results may vary significantly on different platforms
func BenchmarkBulkhead_ConcurrentLowContention(b *testing.B) {
	bh := NewBulkhead("bench", BulkheadConfig{MaxConcurrent: 50, MaxWaiting: 100})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bh.Do(ctx, func(ctx context.Context) error {
				return nil
			})
		}
	})
}

// BenchmarkBulkhead_ConcurrentHighContention tests concurrent access with
// limited capacity, causing waiting.
//
// Typical results (platform-dependent, will vary based on CPU and system load):
// - Apple M1: ~400-500ns/op (variable based on contention level)
// - Results may vary significantly on different platforms
// - Tests the waiting queue mechanism performance
func BenchmarkBulkhead_ConcurrentHighContention(b *testing.B) {
	bh := NewBulkhead("bench", BulkheadConfig{MaxConcurrent: 2, MaxWaiting: 100})
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = bh.Do(ctx, func(ctx context.Context) error {
				return nil
			})
		}
	})
}

// BenchmarkBulkhead_Creation measures bulkhead creation overhead.
//
// Typical results (platform-dependent, will vary based on CPU and system load):
// - Apple M1: ~50-80ns/op (channel allocation)
// - Results may vary significantly on different platforms
// - 3 allocs/op
func BenchmarkBulkhead_Creation(b *testing.B) {
	cfg := BulkheadConfig{MaxConcurrent: 10, MaxWaiting: 100}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = NewBulkhead("bench", cfg)
	}
}
