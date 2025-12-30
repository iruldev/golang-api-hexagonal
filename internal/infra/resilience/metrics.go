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
