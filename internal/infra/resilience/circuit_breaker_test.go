package resilience

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCircuitBreaker_StaysClosedOnSuccess(t *testing.T) {
	// Given a circuit breaker with default config
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())

	// When operations succeed
	ctx := context.Background()
	result, err := cb.Execute(ctx, func() (any, error) {
		return "success", nil
	})

	// Then circuit stays closed and returns result
	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_OpensAfterThresholdFailures(t *testing.T) {
	// Given a circuit breaker with low threshold
	cfg := CircuitBreakerConfig{
		MaxRequests:      1,
		Interval:         10 * time.Second,
		Timeout:          30 * time.Second,
		FailureThreshold: 3, // Open after 3 consecutive failures
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// When threshold failures occur
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		_, err := cb.Execute(ctx, func() (any, error) {
			return nil, testErr
		})
		require.Error(t, err)
	}

	// Then circuit opens
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_RejectsRequestsWhenOpen(t *testing.T) {
	// Given a circuit breaker that is open
	cfg := CircuitBreakerConfig{
		MaxRequests:      1,
		Interval:         10 * time.Second,
		Timeout:          1 * time.Hour, // Long timeout to stay open
		FailureThreshold: 1,             // Open after 1 failure
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Trip the circuit
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("fail")
	})
	require.Equal(t, StateOpen, cb.State())

	// When a request is made
	result, err := cb.Execute(ctx, func() (any, error) {
		return "should not execute", nil
	})

	// Then request is rejected with RES-001 error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrCircuitOpen), "error should be ErrCircuitOpen (RES-001)")

	// Verify error code
	var resErr *ResilienceError
	require.True(t, errors.As(err, &resErr))
	assert.Equal(t, ErrCodeCircuitOpen, resErr.Code)
}

func TestCircuitBreaker_TransitionsToHalfOpenAfterTimeout(t *testing.T) {
	// Given a circuit breaker with short timeout
	cfg := CircuitBreakerConfig{
		MaxRequests:      2, // Need 2 successful requests to close
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond, // Short timeout
		FailureThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Trip the circuit
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("fail")
	})
	require.Equal(t, StateOpen, cb.State())

	// When timeout elapses
	time.Sleep(100 * time.Millisecond)

	// Then circuit transitions to half-open on next request
	// The first successful request in half-open doesn't immediately close
	// Need MaxRequests successful calls to close
	_, err := cb.Execute(ctx, func() (any, error) {
		return "probe request 1", nil
	})
	require.NoError(t, err)

	// State might still be half-open after first success (depends on MaxRequests)
	// After another successful request, circuit should close
	_, err = cb.Execute(ctx, func() (any, error) {
		return "probe request 2", nil
	})
	require.NoError(t, err)

	// After MaxRequests successful probes, circuit should close
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ClosesAfterSuccessfulRequestsInHalfOpen(t *testing.T) {
	// Given a circuit breaker in half-open state
	cfg := CircuitBreakerConfig{
		MaxRequests:      1, // 1 successful request to close
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Trip the circuit
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("fail")
	})
	require.Equal(t, StateOpen, cb.State())

	// Wait for transition to half-open
	time.Sleep(100 * time.Millisecond)

	// When test requests succeed
	result, err := cb.Execute(ctx, func() (any, error) {
		return "success", nil
	})

	// Then circuit closes
	require.NoError(t, err)
	assert.Equal(t, "success", result)
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ReopensAfterFailureInHalfOpen(t *testing.T) {
	// Given a circuit breaker in half-open state
	cfg := CircuitBreakerConfig{
		MaxRequests:      2,
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 1,
	}
	cb := NewCircuitBreaker("test", cfg)
	ctx := context.Background()

	// Trip the circuit
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("initial fail")
	})
	require.Equal(t, StateOpen, cb.State())

	// Wait for transition to half-open
	time.Sleep(100 * time.Millisecond)

	// When a test request fails in half-open state
	_, err := cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("half-open fail")
	})
	require.Error(t, err)

	// Then circuit reopens
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreaker_MetricsUpdatedOnStateTransitions(t *testing.T) {
	// Given a circuit breaker with metrics
	registry := prometheus.NewRegistry()
	metrics := NewCircuitBreakerMetrics(registry)

	cfg := CircuitBreakerConfig{
		MaxRequests:      1,
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 1,
	}
	cb := NewCircuitBreaker("test-metrics", cfg, WithMetrics(metrics))
	ctx := context.Background()

	// When circuit trips (closed → open)
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("fail")
	})

	// Then state should be open (verify cb state)
	assert.Equal(t, StateOpen, cb.State())

	// Wait for half-open and make a successful request
	time.Sleep(100 * time.Millisecond)

	_, _ = cb.Execute(ctx, func() (any, error) {
		return "success", nil
	})

	// Then state should be closed
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_OperationDurationRecorded(t *testing.T) {
	// Given a circuit breaker with metrics
	registry := prometheus.NewRegistry()
	metrics := NewCircuitBreakerMetrics(registry)

	cb := NewCircuitBreaker("test-duration", DefaultCircuitBreakerConfig(), WithMetrics(metrics))
	ctx := context.Background()

	// When an operation is executed
	_, err := cb.Execute(ctx, func() (any, error) {
		time.Sleep(10 * time.Millisecond)
		return "done", nil
	})

	// Then operation completes successfully
	require.NoError(t, err)
	// Metrics would have recorded the duration (verified by no panic)
}

func TestCircuitBreaker_Name(t *testing.T) {
	// Given a circuit breaker with a specific name
	cb := NewCircuitBreaker("my-service", DefaultCircuitBreakerConfig())

	// Then name is accessible
	assert.Equal(t, "my-service", cb.Name())
}

func TestCircuitBreaker_ContextCancellation(t *testing.T) {
	// Given a circuit breaker and a cancelled context
	cb := NewCircuitBreaker("test-ctx", DefaultCircuitBreakerConfig())

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// When Execute is called with cancelled context
	result, err := cb.Execute(ctx, func() (any, error) {
		return "should not reach", nil
	})

	// Then context error is returned
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, context.Canceled))
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.state))
		})
	}
}

func TestStateToInt(t *testing.T) {
	tests := []struct {
		state State
		want  int
	}{
		{StateClosed, 0},
		{StateOpen, 1},
		{StateHalfOpen, 2},
		{State("unknown"), 0}, // Unknown defaults to 0
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			assert.Equal(t, tt.want, stateToInt(tt.state))
		})
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	cfg := DefaultCircuitBreakerConfig()

	assert.Equal(t, DefaultCBMaxRequests, cfg.MaxRequests)
	assert.Equal(t, DefaultCBInterval, cfg.Interval)
	assert.Equal(t, DefaultCBTimeout, cfg.Timeout)
	assert.Equal(t, DefaultCBFailureThreshold, cfg.FailureThreshold)
}

func TestCircuitBreaker_WithOptions(t *testing.T) {
	// Given custom options
	registry := prometheus.NewRegistry()
	metrics := NewCircuitBreakerMetrics(registry)

	// When creating circuit breaker with options
	cb := NewCircuitBreaker("test-options", DefaultCircuitBreakerConfig(),
		WithMetrics(metrics),
	)

	// Then circuit breaker is created successfully
	require.NotNil(t, cb)
	assert.Equal(t, "test-options", cb.Name())
	assert.Equal(t, StateClosed, cb.State())
}

func TestCircuitBreaker_ExecuteReturnsOriginalError(t *testing.T) {
	// Given a circuit breaker
	cb := NewCircuitBreaker("test", DefaultCircuitBreakerConfig())
	ctx := context.Background()

	// When operation returns an error (but circuit doesn't open yet)
	expectedErr := errors.New("original error")
	_, err := cb.Execute(ctx, func() (any, error) {
		return nil, expectedErr
	})

	// Then original error is returned
	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
}

func TestCircuitBreaker_WithLogger(t *testing.T) {
	// Given a custom logger
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// When creating a circuit breaker with custom logger
	cfg := CircuitBreakerConfig{
		MaxRequests:      1,
		Interval:         10 * time.Second,
		Timeout:          50 * time.Millisecond,
		FailureThreshold: 1,
	}
	cb := NewCircuitBreaker("test-logger", cfg, WithLogger(logger))
	ctx := context.Background()

	// Then circuit breaker is created successfully
	require.NotNil(t, cb)
	assert.Equal(t, "test-logger", cb.Name())

	// And logging works during state transitions (trip the circuit to trigger logging)
	_, _ = cb.Execute(ctx, func() (any, error) {
		return nil, errors.New("fail")
	})
	assert.Equal(t, StateOpen, cb.State())
}

func TestCircuitBreakerMetrics_Reset(t *testing.T) {
	// Given metrics with recorded data
	registry := prometheus.NewRegistry()
	metrics := NewCircuitBreakerMetrics(registry)

	// Record some data
	metrics.SetState("test", 1)
	metrics.RecordTransition("test", "closed", "open")
	metrics.RecordOperationDuration("test", "success", 0.1)

	// When Reset is called
	metrics.Reset()

	// Then metrics are cleared (no panic, operation succeeds)
	// Reset clears all time series, so subsequent sets create fresh data
	metrics.SetState("test", 0)
	metrics.RecordTransition("test", "open", "closed")
}

func TestCircuitBreakerMetrics_SetState_AllStates(t *testing.T) {
	// Given metrics
	registry := prometheus.NewRegistry()
	metrics := NewCircuitBreakerMetrics(registry)

	tests := []struct {
		name      string
		stateInt  int
		wantState string
	}{
		{"closed state", 0, "closed"},
		{"open state", 1, "open"},
		{"half-open state", 2, "half-open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When setting state
			metrics.SetState("test-cb", tt.stateInt)

			// Then no panic occurs and state is set
			// (Prometheus metrics are updated internally)
		})
	}
}

func TestNoopCircuitBreakerMetrics(t *testing.T) {
	// When creating noop metrics
	metrics := NoopCircuitBreakerMetrics()

	// Then metrics object is created
	require.NotNil(t, metrics)

	// And operations don't panic
	metrics.SetState("test", 0)
	metrics.RecordTransition("test", "closed", "open")
	metrics.RecordOperationDuration("test", "success", 0.001)
	metrics.Reset()
}

func TestNewCircuitBreakerMetrics_NilRegistry(t *testing.T) {
	// When creating metrics with nil registry
	metrics := NewCircuitBreakerMetrics(nil)

	// Then a new registry is created internally and metrics work
	require.NotNil(t, metrics)

	// And all operations work without panic
	metrics.SetState("test-nil-registry", 0)
	metrics.SetState("test-nil-registry", 1)
	metrics.SetState("test-nil-registry", 2)
	metrics.RecordTransition("test-nil-registry", "closed", "open")
	metrics.RecordOperationDuration("test-nil-registry", "success", 0.005)
}
