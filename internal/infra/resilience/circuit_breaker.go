package resilience

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/sony/gobreaker"
)

// State represents the circuit breaker state.
type State string

const (
	// StateClosed indicates the circuit breaker is closed and requests are allowed.
	StateClosed State = "closed"
	// StateOpen indicates the circuit breaker is open and requests are rejected.
	StateOpen State = "open"
	// StateHalfOpen indicates the circuit breaker is half-open and limited requests are allowed.
	StateHalfOpen State = "half-open"
)

// stateToInt converts State to an integer for metrics.
func stateToInt(s State) int {
	switch s {
	case StateClosed:
		return 0
	case StateOpen:
		return 1
	case StateHalfOpen:
		return 2
	default:
		return 0
	}
}

// goStateToState converts gobreaker.State to our State type.
func goStateToState(s gobreaker.State) State {
	switch s {
	case gobreaker.StateClosed:
		return StateClosed
	case gobreaker.StateOpen:
		return StateOpen
	case gobreaker.StateHalfOpen:
		return StateHalfOpen
	default:
		return StateClosed
	}
}

// CircuitBreaker provides circuit breaker pattern functionality.
// It protects against cascading failures by temporarily blocking
// requests to failing services.
type CircuitBreaker interface {
	// Execute runs the given function with circuit breaker protection.
	// It returns ErrCircuitOpen (RES-001) if the circuit is open.
	Execute(ctx context.Context, fn func() (any, error)) (any, error)

	// State returns the current state of the circuit breaker.
	State() State

	// Name returns the name of this circuit breaker.
	Name() string
}

// circuitBreaker wraps gobreaker.CircuitBreaker with metrics and logging.
type circuitBreaker struct {
	name    string
	breaker *gobreaker.CircuitBreaker
	metrics *CircuitBreakerMetrics
	logger  *slog.Logger
}

// CircuitBreakerOption configures a circuit breaker.
type CircuitBreakerOption func(*circuitBreakerOptions)

type circuitBreakerOptions struct {
	metrics *CircuitBreakerMetrics
	logger  *slog.Logger
}

// WithMetrics sets the metrics for the circuit breaker.
func WithMetrics(m *CircuitBreakerMetrics) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		o.metrics = m
	}
}

// WithLogger sets the logger for the circuit breaker.
func WithLogger(l *slog.Logger) CircuitBreakerOption {
	return func(o *circuitBreakerOptions) {
		o.logger = l
	}
}

// NewCircuitBreaker creates a new circuit breaker with the given name and configuration.
// The circuit breaker will open when the number of consecutive failures reaches the
// configured threshold (FailureThreshold).
func NewCircuitBreaker(name string, cfg CircuitBreakerConfig, opts ...CircuitBreakerOption) CircuitBreaker {
	options := &circuitBreakerOptions{
		metrics: nil,
		logger:  slog.Default(),
	}

	for _, opt := range opts {
		opt(options)
	}

	cb := &circuitBreaker{
		name:    name,
		metrics: options.metrics,
		logger:  options.logger,
	}

	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: uint32(cfg.MaxRequests),
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= uint32(cfg.FailureThreshold)
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			cb.onStateChange(name, from, to)
		},
	}

	cb.breaker = gobreaker.NewCircuitBreaker(settings)

	// Initialize metrics with closed state
	if cb.metrics != nil {
		cb.metrics.SetState(name, stateToInt(StateClosed))
	}

	return cb
}

// Execute runs the given function with circuit breaker protection.
// If the circuit is open, it returns ErrCircuitOpen immediately.
// The context is passed through for cancellation support.
func (cb *circuitBreaker) Execute(ctx context.Context, fn func() (any, error)) (any, error) {
	start := time.Now()

	result, err := cb.breaker.Execute(func() (any, error) {
		// Check context cancellation before executing
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return fn()
	})

	duration := time.Since(start).Seconds()

	// Handle circuit open error
	if errors.Is(err, gobreaker.ErrOpenState) {
		if cb.metrics != nil {
			cb.metrics.RecordOperationDuration(cb.name, "rejected", duration)
		}
		return nil, NewCircuitOpenError(err)
	}

	// Handle too many requests error (circuit is half-open and max requests exceeded)
	if errors.Is(err, gobreaker.ErrTooManyRequests) {
		if cb.metrics != nil {
			cb.metrics.RecordOperationDuration(cb.name, "rejected", duration)
		}
		return nil, NewCircuitOpenError(err)
	}

	// Record metrics for success/failure
	if cb.metrics != nil {
		if err != nil {
			cb.metrics.RecordOperationDuration(cb.name, "failure", duration)
		} else {
			cb.metrics.RecordOperationDuration(cb.name, "success", duration)
		}
	}

	return result, err
}

// State returns the current state of the circuit breaker.
func (cb *circuitBreaker) State() State {
	return goStateToState(cb.breaker.State())
}

// Name returns the name of this circuit breaker.
func (cb *circuitBreaker) Name() string {
	return cb.name
}

// onStateChange is called when the circuit breaker state changes.
func (cb *circuitBreaker) onStateChange(name string, from, to gobreaker.State) {
	fromState := goStateToState(from)
	toState := goStateToState(to)

	// Update metrics
	if cb.metrics != nil {
		cb.metrics.SetState(name, stateToInt(toState))
		cb.metrics.RecordTransition(name, string(fromState), string(toState))
	}

	// Log state change
	// Use INFO level for significant transitions (closed→open, any→closed)
	// Use DEBUG level for half-open transitions
	logLevel := slog.LevelDebug
	if to == gobreaker.StateOpen || to == gobreaker.StateClosed {
		logLevel = slog.LevelInfo
	}

	cb.logger.Log(context.Background(), logLevel, "circuit breaker state changed",
		"name", name,
		"previous_state", string(fromState),
		"new_state", string(toState),
		"timestamp", time.Now().Format(time.RFC3339),
	)
}

// DefaultCircuitBreakerConfig returns a CircuitBreakerConfig with sensible defaults.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxRequests:      DefaultCBMaxRequests,
		Interval:         DefaultCBInterval,
		Timeout:          DefaultCBTimeout,
		FailureThreshold: DefaultCBFailureThreshold,
	}
}
