package resilience

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// discardLogger returns a logger that discards all output (for benchmarks).
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// BenchmarkShutdownCoordinator_IncrementDecrement measures the performance of
// request tracking operations which are called on every HTTP request.
func BenchmarkShutdownCoordinator_IncrementDecrement(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if coord.IncrementActive() {
			coord.DecrementActive()
		}
	}
}

// BenchmarkShutdownCoordinator_IncrementDecrement_WithMetrics measures overhead
// when Prometheus metrics are enabled.
func BenchmarkShutdownCoordinator_IncrementDecrement_WithMetrics(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	metrics := NewShutdownMetrics(prometheus.NewRegistry())
	coord := NewShutdownCoordinator(cfg, WithShutdownMetrics(metrics))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if coord.IncrementActive() {
			coord.DecrementActive()
		}
	}
}

// BenchmarkShutdownCoordinator_IncrementOnly measures raw increment performance.
func BenchmarkShutdownCoordinator_IncrementOnly(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord.IncrementActive()
	}
	b.StopTimer()

	// Cleanup: decrement all
	for i := 0; i < b.N; i++ {
		coord.DecrementActive()
	}
}

// BenchmarkShutdownCoordinator_ActiveCount measures the read operation performance.
func BenchmarkShutdownCoordinator_ActiveCount(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)
	coord.IncrementActive()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = coord.ActiveCount()
	}
}

// BenchmarkShutdownCoordinator_IsShuttingDown measures shutdown state check performance.
func BenchmarkShutdownCoordinator_IsShuttingDown(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = coord.IsShuttingDown()
	}
}

// BenchmarkShutdownCoordinator_Parallel measures concurrent request tracking.
func BenchmarkShutdownCoordinator_Parallel(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if coord.IncrementActive() {
				coord.DecrementActive()
			}
		}
	})
}

// BenchmarkShutdownCoordinator_Parallel_WithMetrics measures concurrent performance with metrics.
func BenchmarkShutdownCoordinator_Parallel_WithMetrics(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	metrics := NewShutdownMetrics(prometheus.NewRegistry())
	coord := NewShutdownCoordinator(cfg, WithShutdownMetrics(metrics))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if coord.IncrementActive() {
				coord.DecrementActive()
			}
		}
	})
}

// BenchmarkShutdownCoordinator_DuringShutdown measures rejection performance during shutdown.
func BenchmarkShutdownCoordinator_DuringShutdown(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg, WithShutdownLogger(discardLogger()))
	coord.InitiateShutdown()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		coord.IncrementActive() // Should return false
	}
}

// BenchmarkShutdownCoordinator_WaitForDrain_Immediate measures drain when no active requests.
func BenchmarkShutdownCoordinator_WaitForDrain_Immediate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := ShutdownConfig{
			DrainPeriod: 1 * time.Second,
			GracePeriod: 0,
		}
		coord := NewShutdownCoordinator(cfg, WithShutdownLogger(discardLogger()))
		coord.InitiateShutdown()
		_ = coord.WaitForDrain(context.Background())
	}
}

// BenchmarkShutdownCoordinator_Creation measures coordinator instantiation.
func BenchmarkShutdownCoordinator_Creation(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewShutdownCoordinator(cfg)
	}
}

// BenchmarkShutdownCoordinator_HighContention simulates high-load scenario.
func BenchmarkShutdownCoordinator_HighContention(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	// Pre-populate with some active requests
	for i := 0; i < 100; i++ {
		coord.IncrementActive()
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if coord.IncrementActive() {
				coord.DecrementActive()
			}
		}
	})

	b.StopTimer()
	// Cleanup
	for i := 0; i < 100; i++ {
		coord.DecrementActive()
	}
}

// BenchmarkShutdownMetrics_SetActiveRequests measures metrics update performance.
func BenchmarkShutdownMetrics_SetActiveRequests(b *testing.B) {
	metrics := NewShutdownMetrics(prometheus.NewRegistry())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.SetActiveRequests(int64(i % 100))
	}
}

// BenchmarkShutdownMetrics_RecordRejection measures rejection counter performance.
func BenchmarkShutdownMetrics_RecordRejection(b *testing.B) {
	metrics := NewShutdownMetrics(prometheus.NewRegistry())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordRejection()
	}
}

// BenchmarkShutdownMetrics_RecordShutdownDuration measures histogram observation.
func BenchmarkShutdownMetrics_RecordShutdownDuration(b *testing.B) {
	metrics := NewShutdownMetrics(prometheus.NewRegistry())
	duration := 5 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics.RecordShutdownDuration(duration, "success")
	}
}

// BenchmarkShutdownCoordinator_MixedWorkload simulates realistic request pattern.
func BenchmarkShutdownCoordinator_MixedWorkload(b *testing.B) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var wg sync.WaitGroup
		for pb.Next() {
			if coord.IncrementActive() {
				// Simulate brief request processing
				wg.Add(1)
				go func() {
					defer wg.Done()
					_ = coord.ActiveCount()
					_ = coord.IsShuttingDown()
					coord.DecrementActive()
				}()
			}
		}
		wg.Wait()
	})
}
