package resilience

import (
	"github.com/prometheus/client_golang/prometheus"
)

// CircuitBreakerMetrics provides Prometheus metrics for circuit breaker monitoring.
type CircuitBreakerMetrics struct {
	// state tracks the current state of each circuit breaker using {name, state} labels.
	// Each state (closed, open, half-open) is a separate time series with value 1 (active) or 0 (inactive).
	state *prometheus.GaugeVec

	// transitions counts state transitions.
	transitions *prometheus.CounterVec

	// operationDuration measures the duration of operations executed through the circuit breaker.
	operationDuration *prometheus.HistogramVec
}

// NewCircuitBreakerMetrics creates and registers circuit breaker metrics with the given registry.
// If registry is nil, a new registry is created.
func NewCircuitBreakerMetrics(registry *prometheus.Registry) *CircuitBreakerMetrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	state := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of the circuit breaker (1=active, 0=inactive for each state label)",
		},
		[]string{"name", "state"},
	)

	transitions := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_transitions_total",
			Help: "Total number of circuit breaker state transitions",
		},
		[]string{"name", "from", "to"},
	)

	operationDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "circuit_breaker_operation_duration_seconds",
			Help: "Duration of operations executed through the circuit breaker",
			Buckets: []float64{
				0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
			},
		},
		[]string{"name", "result"},
	)

	// Register metrics with registry.
	// Errors are intentionally ignored as they indicate metrics are already registered,
	// which is expected when creating multiple circuit breakers in the same process.
	_ = registry.Register(state)
	_ = registry.Register(transitions)
	_ = registry.Register(operationDuration)

	return &CircuitBreakerMetrics{
		state:             state,
		transitions:       transitions,
		operationDuration: operationDuration,
	}
}

// SetState updates the state gauge for a circuit breaker.
// Sets the active state to 1 and all other states to 0.
// state: 0=closed, 1=open, 2=half-open
func (m *CircuitBreakerMetrics) SetState(name string, state int) {
	// Set all states to 0 first
	m.state.WithLabelValues(name, "closed").Set(0)
	m.state.WithLabelValues(name, "open").Set(0)
	m.state.WithLabelValues(name, "half-open").Set(0)

	// Set the active state to 1
	switch state {
	case 0:
		m.state.WithLabelValues(name, "closed").Set(1)
	case 1:
		m.state.WithLabelValues(name, "open").Set(1)
	case 2:
		m.state.WithLabelValues(name, "half-open").Set(1)
	}
}

// RecordTransition increments the transition counter for a circuit breaker.
func (m *CircuitBreakerMetrics) RecordTransition(name, from, to string) {
	m.transitions.WithLabelValues(name, from, to).Inc()
}

// RecordOperationDuration records the duration of an operation and its result.
// result should be one of: "success", "failure", "rejected"
func (m *CircuitBreakerMetrics) RecordOperationDuration(name, result string, durationSeconds float64) {
	m.operationDuration.WithLabelValues(name, result).Observe(durationSeconds)
}

// Reset resets all metrics. Useful for testing.
func (m *CircuitBreakerMetrics) Reset() {
	m.state.Reset()
	m.transitions.Reset()
	m.operationDuration.Reset()
}

// NoopCircuitBreakerMetrics returns a no-op metrics implementation for testing.
func NoopCircuitBreakerMetrics() *CircuitBreakerMetrics {
	return NewCircuitBreakerMetrics(prometheus.NewRegistry())
}

// RetryMetrics provides Prometheus metrics for retry monitoring.
type RetryMetrics struct {
	// operationTotal counts retry operations by name, result, and attempt count.
	operationTotal *prometheus.CounterVec

	// attemptTotal counts individual retry attempts by name and result.
	attemptTotal *prometheus.CounterVec

	// durationSeconds measures total duration of retry operations.
	durationSeconds *prometheus.HistogramVec
}

// NewRetryMetrics creates and registers retry metrics with the given registry.
// If registry is nil, a new registry is created.
func NewRetryMetrics(registry *prometheus.Registry) *RetryMetrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	operationTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_operation_total",
			Help: "Total number of retry operations by attempt count",
		},
		[]string{"name", "attempts"},
	)

	attemptTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_attempts_total",
			Help: "Total number of completed retry operations by result. Labels: name=retrier name, result=success|failure|exhausted. Incremented once per operation completion, not per individual retry attempt.",
		},
		[]string{"name", "result"},
	)

	durationSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "retry_duration_seconds",
			Help: "Duration of retry operations including all attempts",
			Buckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0,
			},
		},
		[]string{"name", "result"},
	)

	// Register metrics with registry.
	// Errors are intentionally ignored as they indicate metrics are already registered,
	// which is expected when creating multiple retriers in the same process.
	_ = registry.Register(operationTotal)
	_ = registry.Register(attemptTotal)
	_ = registry.Register(durationSeconds)

	return &RetryMetrics{
		operationTotal:  operationTotal,
		attemptTotal:    attemptTotal,
		durationSeconds: durationSeconds,
	}
}

// RecordOperation records a retry operation completion.
// result should be one of: "success", "failure", "exhausted"
func (m *RetryMetrics) RecordOperation(name, result string, attempts int, durationSeconds float64) {
	m.attemptTotal.WithLabelValues(name, result).Inc()
	m.operationTotal.WithLabelValues(name, itoa(attempts)).Inc()
	m.durationSeconds.WithLabelValues(name, result).Observe(durationSeconds)
}

// Reset resets all metrics. Useful for testing.
func (m *RetryMetrics) Reset() {
	m.operationTotal.Reset()
	m.attemptTotal.Reset()
	m.durationSeconds.Reset()
}

// NoopRetryMetrics returns a no-op metrics implementation for testing.
func NoopRetryMetrics() *RetryMetrics {
	return NewRetryMetrics(prometheus.NewRegistry())
}

// itoa converts an integer to a string for metric labels.
func itoa(i int) string {
	// Simple implementation for small numbers used in retry counts
	if i < 0 {
		return "-" + itoa(-i)
	}
	if i < 10 {
		return string(rune('0' + i))
	}
	return itoa(i/10) + string(rune('0'+i%10))
}

// TimeoutMetrics provides Prometheus metrics for timeout monitoring.
type TimeoutMetrics struct {
	// operationTotal counts timeout operations by name and result.
	operationTotal *prometheus.CounterVec

	// durationSeconds measures duration of operations with timeout.
	durationSeconds *prometheus.HistogramVec
}

// NewTimeoutMetrics creates and registers timeout metrics with the given registry.
// If registry is nil, a new registry is created.
func NewTimeoutMetrics(registry *prometheus.Registry) *TimeoutMetrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	operationTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "timeout_total",
			Help: "Total number of timeout operations by result. Labels: name=timeout name, result=success|timeout|error.",
		},
		[]string{"name", "result"},
	)

	durationSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "timeout_duration_seconds",
			Help: "Duration of operations executed with timeout wrapper",
			Buckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0,
			},
		},
		[]string{"name", "result"},
	)

	// Register metrics with registry.
	// Errors are intentionally ignored as they indicate metrics are already registered,
	// which is expected when creating multiple timeouts in the same process.
	_ = registry.Register(operationTotal)
	_ = registry.Register(durationSeconds)

	return &TimeoutMetrics{
		operationTotal:  operationTotal,
		durationSeconds: durationSeconds,
	}
}

// RecordOperation records a timeout operation completion.
// result should be one of: "success", "timeout", "error"
func (m *TimeoutMetrics) RecordOperation(name, result string, durationSeconds float64) {
	m.operationTotal.WithLabelValues(name, result).Inc()
	m.durationSeconds.WithLabelValues(name, result).Observe(durationSeconds)
}

// Reset resets all metrics. Useful for testing.
func (m *TimeoutMetrics) Reset() {
	m.operationTotal.Reset()
	m.durationSeconds.Reset()
}

// NoopTimeoutMetrics returns a no-op metrics implementation for testing.
func NoopTimeoutMetrics() *TimeoutMetrics {
	return NewTimeoutMetrics(prometheus.NewRegistry())
}

// BulkheadMetrics provides Prometheus metrics for bulkhead monitoring.
type BulkheadMetrics struct {
	// active tracks the current number of active executions per bulkhead.
	active *prometheus.GaugeVec

	// waiting tracks the current number of waiting operations per bulkhead.
	waiting *prometheus.GaugeVec

	// operationTotal counts bulkhead operations by name and result.
	operationTotal *prometheus.CounterVec

	// waitDurationSeconds measures time spent waiting for a slot.
	waitDurationSeconds *prometheus.HistogramVec
}

// NewBulkheadMetrics creates and registers bulkhead metrics with the given registry.
// If registry is nil, a new registry is created.
func NewBulkheadMetrics(registry *prometheus.Registry) *BulkheadMetrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	active := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_active",
			Help: "Current number of active executions in the bulkhead",
		},
		[]string{"name"},
	)

	waiting := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bulkhead_waiting",
			Help: "Current number of operations waiting for a slot in the bulkhead",
		},
		[]string{"name"},
	)

	operationTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bulkhead_total",
			Help: "Total number of bulkhead operations by result. Labels: name=bulkhead name, result=success|error|rejected|cancelled.",
		},
		[]string{"name", "result"},
	)

	waitDurationSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "bulkhead_wait_duration_seconds",
			Help: "Time spent waiting for a slot in the bulkhead",
			Buckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
			},
		},
		[]string{"name"},
	)

	// Register metrics with registry.
	// Errors are intentionally ignored as they indicate metrics are already registered,
	// which is expected when creating multiple bulkheads in the same process.
	_ = registry.Register(active)
	_ = registry.Register(waiting)
	_ = registry.Register(operationTotal)
	_ = registry.Register(waitDurationSeconds)

	return &BulkheadMetrics{
		active:              active,
		waiting:             waiting,
		operationTotal:      operationTotal,
		waitDurationSeconds: waitDurationSeconds,
	}
}

// SetActive updates the active execution count for a bulkhead.
func (m *BulkheadMetrics) SetActive(name string, count int) {
	m.active.WithLabelValues(name).Set(float64(count))
}

// SetWaiting updates the waiting operation count for a bulkhead.
func (m *BulkheadMetrics) SetWaiting(name string, count int) {
	m.waiting.WithLabelValues(name).Set(float64(count))
}

// RecordOperation records a bulkhead operation completion.
// result should be one of: "success", "error", "rejected", "cancelled"
func (m *BulkheadMetrics) RecordOperation(name, result string) {
	m.operationTotal.WithLabelValues(name, result).Inc()
}

// RecordWaitDuration records time spent waiting for a slot.
func (m *BulkheadMetrics) RecordWaitDuration(name string, durationSeconds float64) {
	m.waitDurationSeconds.WithLabelValues(name).Observe(durationSeconds)
}

// Reset resets all metrics. Useful for testing.
func (m *BulkheadMetrics) Reset() {
	m.active.Reset()
	m.waiting.Reset()
	m.operationTotal.Reset()
	m.waitDurationSeconds.Reset()
}

// NoopBulkheadMetrics returns a no-op metrics implementation for testing.
func NoopBulkheadMetrics() *BulkheadMetrics {
	return NewBulkheadMetrics(prometheus.NewRegistry())
}
