package resilience

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShutdownCoordinator(t *testing.T) {
	t.Run("creates with valid config", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 30 * time.Second,
			GracePeriod: 5 * time.Second,
		}

		coord := NewShutdownCoordinator(cfg)

		require.NotNil(t, coord)
		assert.False(t, coord.IsShuttingDown())
		assert.Equal(t, int64(0), coord.ActiveCount())
	})

	t.Run("panics with invalid config", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 0, // Invalid
			GracePeriod: 5 * time.Second,
		}

		assert.Panics(t, func() {
			NewShutdownCoordinator(cfg)
		})
	})

	t.Run("accepts options", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 30 * time.Second,
			GracePeriod: 5 * time.Second,
		}

		metrics := NewShutdownMetrics(prometheus.NewRegistry())
		coord := NewShutdownCoordinator(cfg, WithShutdownMetrics(metrics))

		require.NotNil(t, coord)
	})

	t.Run("handles nil metrics option gracefully", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 30 * time.Second,
			GracePeriod: 5 * time.Second,
		}

		coord := NewShutdownCoordinator(cfg, WithShutdownMetrics(nil))

		require.NotNil(t, coord)
	})

	t.Run("handles nil logger option gracefully", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 30 * time.Second,
			GracePeriod: 5 * time.Second,
		}

		coord := NewShutdownCoordinator(cfg, WithShutdownLogger(nil))

		require.NotNil(t, coord)
	})
}

func TestShutdownCoordinator_IncrementDecrement(t *testing.T) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}

	t.Run("tracks active requests correctly", func(t *testing.T) {
		coord := NewShutdownCoordinator(cfg)

		assert.Equal(t, int64(0), coord.ActiveCount())

		ok := coord.IncrementActive()
		assert.True(t, ok)
		assert.Equal(t, int64(1), coord.ActiveCount())

		ok = coord.IncrementActive()
		assert.True(t, ok)
		assert.Equal(t, int64(2), coord.ActiveCount())

		coord.DecrementActive()
		assert.Equal(t, int64(1), coord.ActiveCount())

		coord.DecrementActive()
		assert.Equal(t, int64(0), coord.ActiveCount())
	})

	t.Run("rejects increment during shutdown", func(t *testing.T) {
		coord := NewShutdownCoordinator(cfg)

		// Start one active request
		ok := coord.IncrementActive()
		assert.True(t, ok)

		// Initiate shutdown
		coord.InitiateShutdown()

		// New increment should be rejected
		ok = coord.IncrementActive()
		assert.False(t, ok)
		assert.Equal(t, int64(1), coord.ActiveCount()) // Still only 1

		// Decrement still works
		coord.DecrementActive()
		assert.Equal(t, int64(0), coord.ActiveCount())
	})
}

func TestShutdownCoordinator_IsShuttingDown(t *testing.T) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}

	t.Run("returns false initially", func(t *testing.T) {
		coord := NewShutdownCoordinator(cfg)
		assert.False(t, coord.IsShuttingDown())
	})

	t.Run("returns true after InitiateShutdown", func(t *testing.T) {
		coord := NewShutdownCoordinator(cfg)
		coord.InitiateShutdown()
		assert.True(t, coord.IsShuttingDown())
	})

	t.Run("InitiateShutdown is idempotent", func(t *testing.T) {
		coord := NewShutdownCoordinator(cfg)
		coord.InitiateShutdown()
		coord.InitiateShutdown() // Should not panic or cause issues
		assert.True(t, coord.IsShuttingDown())
	})
}

func TestShutdownCoordinator_WaitForDrain(t *testing.T) {
	t.Run("completes immediately when no active requests", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 1 * time.Second,
			GracePeriod: 0,
		}
		coord := NewShutdownCoordinator(cfg)
		coord.InitiateShutdown()

		ctx := context.Background()
		err := coord.WaitForDrain(ctx)

		assert.NoError(t, err)
	})

	t.Run("waits for active requests to complete", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 5 * time.Second,
			GracePeriod: 0,
		}
		coord := NewShutdownCoordinator(cfg)

		// Start an active request
		coord.IncrementActive()
		coord.InitiateShutdown()

		// Simulate request completing after short delay
		go func() {
			time.Sleep(100 * time.Millisecond)
			coord.DecrementActive()
		}()

		ctx := context.Background()
		start := time.Now()
		err := coord.WaitForDrain(ctx)
		elapsed := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, elapsed, 1*time.Second) // Should complete quickly
	})

	t.Run("times out after drain period", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 200 * time.Millisecond,
			GracePeriod: 0,
		}
		coord := NewShutdownCoordinator(cfg)

		// Start an active request that won't complete
		coord.IncrementActive()
		coord.InitiateShutdown()

		ctx := context.Background()
		start := time.Now()
		err := coord.WaitForDrain(ctx)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drain timeout")
		assert.Contains(t, err.Error(), "1 requests still active")
		// Should have waited approximately drain period
		assert.GreaterOrEqual(t, elapsed, 200*time.Millisecond)
		assert.Less(t, elapsed, 500*time.Millisecond)
	})

	t.Run("respects parent context cancellation", func(t *testing.T) {
		cfg := ShutdownConfig{
			DrainPeriod: 5 * time.Second,
			GracePeriod: 0,
		}
		coord := NewShutdownCoordinator(cfg)

		// Start an active request that won't complete
		coord.IncrementActive()
		coord.InitiateShutdown()

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := coord.WaitForDrain(ctx)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "drain timeout")
		// Should have timed out based on context, not drain period
		assert.Less(t, elapsed, 1*time.Second)
	})
}

func TestShutdownCoordinator_ConcurrentAccess(t *testing.T) {
	cfg := ShutdownConfig{
		DrainPeriod: 5 * time.Second,
		GracePeriod: 0,
	}
	coord := NewShutdownCoordinator(cfg)

	const numGoroutines = 100
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				if coord.IncrementActive() {
					// Simulate some work
					time.Sleep(time.Microsecond)
					coord.DecrementActive()
				}
			}
		}()
	}

	wg.Wait()

	// All requests should have completed
	assert.Equal(t, int64(0), coord.ActiveCount())
}

func TestShutdownCoordinator_ConcurrentShutdown(t *testing.T) {
	cfg := ShutdownConfig{
		DrainPeriod: 5 * time.Second,
		GracePeriod: 0,
	}
	// Use silent logger to avoid excessive WARN logs from intentional negative count guards
	silentLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	coord := NewShutdownCoordinator(cfg, WithShutdownLogger(silentLogger))

	// Start some active requests
	for i := 0; i < 10; i++ {
		coord.IncrementActive()
	}

	const numGoroutines = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Half try to initiate shutdown
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			coord.InitiateShutdown()
		}()
	}

	// Half try to complete requests
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			coord.DecrementActive()
		}()
	}

	wg.Wait()

	// All should be stable
	assert.True(t, coord.IsShuttingDown())
	// DecrementActive now guards against negative counts, so this is safe
}

func TestShutdownCoordinator_WithMetrics(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics := NewShutdownMetrics(registry)

	cfg := ShutdownConfig{
		DrainPeriod: 200 * time.Millisecond,
		GracePeriod: 0,
	}
	coord := NewShutdownCoordinator(cfg, WithShutdownMetrics(metrics))

	// Increment active
	ok := coord.IncrementActive()
	assert.True(t, ok)

	// Initiate shutdown
	coord.InitiateShutdown()

	// Try to increment (should be rejected)
	ok = coord.IncrementActive()
	assert.False(t, ok)

	// Decrement
	coord.DecrementActive()

	// Wait for drain
	err := coord.WaitForDrain(context.Background())
	assert.NoError(t, err)
}

func TestShutdownCoordinator_Config(t *testing.T) {
	cfg := ShutdownConfig{
		DrainPeriod: 45 * time.Second,
		GracePeriod: 10 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	// Config should return the original configuration
	returnedCfg := coord.Config()
	assert.Equal(t, cfg.DrainPeriod, returnedCfg.DrainPeriod)
	assert.Equal(t, cfg.GracePeriod, returnedCfg.GracePeriod)
}

func TestShutdownCoordinator_NegativeCountGuard(t *testing.T) {
	cfg := ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := NewShutdownCoordinator(cfg)

	// Decrement without any increment should be guarded
	coord.DecrementActive()

	// Active count should be 0, not negative
	assert.Equal(t, int64(0), coord.ActiveCount())
}
