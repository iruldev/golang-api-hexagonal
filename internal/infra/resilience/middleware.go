// Package resilience provides resilience patterns for service operations.
// This file implements the ResilienceWrapper for composable resilience patterns.

package resilience

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ResilienceWrapper composes multiple resilience patterns into a single execution wrapper.
// It applies patterns in the correct order: CircuitBreaker → Retry → Timeout (outermost to innermost).
//
// Composition flow (execution order):
//  1. Check circuit breaker state (outermost) - fast-fail if circuit is open
//  2. Apply retry logic with backoff - retry on transient failures
//  3. Apply timeout to each attempt (innermost) - limit each attempt duration
//
// All operations are traced with OpenTelemetry spans for observability.
//
// ResilienceWrapper implementations are safe for concurrent use from multiple goroutines.
type ResilienceWrapper interface {
	// Execute wraps the given function with configured resilience patterns.
	// The name parameter is used for circuit breaker identification and tracing.
	Execute(ctx context.Context, name string, fn func(ctx context.Context) error) error
}

// resilienceWrapper implements the ResilienceWrapper interface.
type resilienceWrapper struct {
	cbFactory func(name string) CircuitBreaker
	retrier   Retrier
	timeout   Timeout
	bulkhead  Bulkhead
	tracer    trace.Tracer
	logger    *slog.Logger
}

// ResilienceWrapperOption configures a ResilienceWrapper.
type ResilienceWrapperOption func(*resilienceWrapperOptions)

type resilienceWrapperOptions struct {
	cbFactory func(name string) CircuitBreaker
	retrier   Retrier
	timeout   Timeout
	bulkhead  Bulkhead
	tracer    trace.Tracer
	logger    *slog.Logger
}

// WithCircuitBreakerFactory sets the circuit breaker factory for the wrapper.
// If factory is nil, circuit breaker protection is not applied.
func WithCircuitBreakerFactory(factory func(name string) CircuitBreaker) ResilienceWrapperOption {
	return func(o *resilienceWrapperOptions) {
		if factory != nil {
			o.cbFactory = factory
		}
	}
}

// WithWrapperRetrier sets the retrier for the wrapper.
// If r is nil, retry logic is not applied.
func WithWrapperRetrier(r Retrier) ResilienceWrapperOption {
	return func(o *resilienceWrapperOptions) {
		if r != nil {
			o.retrier = r
		}
	}
}

// WithWrapperTimeout sets the timeout for the wrapper.
// If t is nil, timeout is not applied.
func WithWrapperTimeout(t Timeout) ResilienceWrapperOption {
	return func(o *resilienceWrapperOptions) {
		if t != nil {
			o.timeout = t
		}
	}
}

// WithWrapperBulkhead sets the bulkhead for the wrapper.
// If b is nil, bulkhead protection is not applied.
func WithWrapperBulkhead(b Bulkhead) ResilienceWrapperOption {
	return func(o *resilienceWrapperOptions) {
		if b != nil {
			o.bulkhead = b
		}
	}
}

// WithWrapperTracer sets the OpenTelemetry tracer for the wrapper.
// If tracer is nil, a default tracer named "resilience" is used.
func WithWrapperTracer(tracer trace.Tracer) ResilienceWrapperOption {
	return func(o *resilienceWrapperOptions) {
		if tracer != nil {
			o.tracer = tracer
		}
	}
}

// WithWrapperLogger sets the logger for the wrapper.
// If l is nil, the default logger (slog.Default()) is used.
func WithWrapperLogger(l *slog.Logger) ResilienceWrapperOption {
	return func(o *resilienceWrapperOptions) {
		if l != nil {
			o.logger = l
		}
	}
}

// NewResilienceWrapper creates a new ResilienceWrapper with the given options.
// All resilience components (circuit breaker, retrier, timeout, bulkhead) are optional.
// Components that are not provided will be skipped during execution.
func NewResilienceWrapper(opts ...ResilienceWrapperOption) ResilienceWrapper {
	options := &resilienceWrapperOptions{
		cbFactory: nil,
		retrier:   nil,
		timeout:   nil,
		bulkhead:  nil,
		tracer:    otel.Tracer("resilience"),
		logger:    slog.Default(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return &resilienceWrapper{
		cbFactory: options.cbFactory,
		retrier:   options.retrier,
		timeout:   options.timeout,
		bulkhead:  options.bulkhead,
		tracer:    options.tracer,
		logger:    options.logger,
	}
}

// Execute wraps the given function with configured resilience patterns.
// The composition order is: CircuitBreaker → Retry → Timeout (outermost to innermost).
// Bulkhead is applied at the outermost level if configured.
//
// Each layer adds protection:
//   - Bulkhead: limits concurrent executions to prevent resource exhaustion
//   - CircuitBreaker: fast-fails when downstream is unhealthy
//   - Retry: retries transient failures with exponential backoff
//   - Timeout: limits duration of each individual attempt
func (w *resilienceWrapper) Execute(ctx context.Context, name string, fn func(ctx context.Context) error) error {
	start := time.Now()

	// Start tracing span
	ctx, span := w.tracer.Start(ctx, "resilience.Execute",
		trace.WithAttributes(
			attribute.String("operation", name),
		),
	)
	defer span.End()

	// Build the execution chain from innermost to outermost
	// Final chain: Bulkhead → CircuitBreaker → Retry → Timeout → fn

	// Innermost: the actual operation
	operation := fn

	// Wrap with timeout if configured (innermost wrapper)
	operation = w.wrapTimeout(operation, span)

	// Wrap with retry if configured
	operation = w.wrapRetry(operation, span)

	// Wrap with circuit breaker if configured
	operation = w.wrapCircuitBreaker(name, operation, span)

	// Wrap with bulkhead if configured (outermost)
	operation = w.wrapBulkhead(operation, span)

	// Execute the composed operation
	err := operation(ctx)

	duration := time.Since(start)

	// Record result in span
	w.recordResult(span, name, err, duration)

	return err
}

func (w *resilienceWrapper) wrapTimeout(next func(ctx context.Context) error, span trace.Span) func(ctx context.Context) error {
	if w.timeout == nil {
		return next
	}

	return func(ctx context.Context) error {
		span.AddEvent("timeout.start", trace.WithAttributes(
			attribute.String("component", "timeout"),
			attribute.String("duration", w.timeout.Duration().String()),
		))
		err := w.timeout.Do(ctx, next)
		if err != nil {
			span.AddEvent("timeout.error", trace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		return err
	}
}

func (w *resilienceWrapper) wrapRetry(next func(ctx context.Context) error, span trace.Span) func(ctx context.Context) error {
	if w.retrier == nil {
		return next
	}

	return func(ctx context.Context) error {
		span.AddEvent("retry.start", trace.WithAttributes(
			attribute.String("component", "retry"),
			attribute.String("retrier", w.retrier.Name()),
		))
		err := w.retrier.Do(ctx, next)
		if err != nil {
			span.AddEvent("retry.exhausted", trace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		return err
	}
}

func (w *resilienceWrapper) wrapCircuitBreaker(name string, next func(ctx context.Context) error, span trace.Span) func(ctx context.Context) error {
	if w.cbFactory == nil {
		return next
	}

	cb := w.cbFactory(name)
	return func(ctx context.Context) error {
		span.AddEvent("circuit_breaker.check", trace.WithAttributes(
			attribute.String("component", "circuit_breaker"),
			attribute.String("name", cb.Name()),
			attribute.String("state", string(cb.State())),
		))
		_, err := cb.Execute(ctx, func() (any, error) {
			return nil, next(ctx)
		})
		if err != nil {
			span.AddEvent("circuit_breaker.error", trace.WithAttributes(
				attribute.String("error", err.Error()),
				attribute.String("state", string(cb.State())),
			))
		}
		return err
	}
}

func (w *resilienceWrapper) wrapBulkhead(next func(ctx context.Context) error, span trace.Span) func(ctx context.Context) error {
	if w.bulkhead == nil {
		return next
	}

	return func(ctx context.Context) error {
		span.AddEvent("bulkhead.acquire", trace.WithAttributes(
			attribute.String("component", "bulkhead"),
			attribute.String("name", w.bulkhead.Name()),
			attribute.Int("active_count", w.bulkhead.ActiveCount()),
			attribute.Int("waiting_count", w.bulkhead.WaitingCount()),
		))
		err := w.bulkhead.Do(ctx, next)
		if err != nil {
			span.AddEvent("bulkhead.error", trace.WithAttributes(
				attribute.String("error", err.Error()),
			))
		}
		return err
	}
}

func (w *resilienceWrapper) recordResult(span trace.Span, name string, err error, duration time.Duration) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(
			attribute.String("error.type", errorType(err)),
			attribute.Float64("duration_seconds", duration.Seconds()),
		)
		w.logger.Debug("resilience wrapper operation failed",
			"name", name,
			"duration_ms", duration.Milliseconds(),
			"error", err.Error(),
			"error_type", errorType(err),
		)
	} else {
		span.SetStatus(codes.Ok, "success")
		span.SetAttributes(
			attribute.Float64("duration_seconds", duration.Seconds()),
		)
		w.logger.Debug("resilience wrapper operation succeeded",
			"name", name,
			"duration_ms", duration.Milliseconds(),
		)
	}
}

// errorType returns a string categorizing the error type for metrics/tracing.
func errorType(err error) string {
	if err == nil {
		return "none"
	}

	// Check for resilience-specific errors
	switch {
	case IsCircuitOpen(err):
		return "circuit_open"
	case IsBulkheadFull(err):
		return "bulkhead_full"
	case IsTimeoutExceeded(err):
		return "timeout"
	case IsMaxRetriesExceeded(err):
		return "max_retries"
	default:
		return "unknown"
	}
}

// CircuitBreakerFactory creates named circuit breakers with independent state.
// Each name returns a distinct circuit breaker with its own failure tracking.
type CircuitBreakerFactory func(name string) CircuitBreaker

// NewCircuitBreakerFactory creates a factory for named circuit breakers.
// The factory uses the provided configuration and options to create each breaker.
// Created circuit breakers are cached by name, so calling with the same name
// returns the same instance.
func NewCircuitBreakerFactory(
	cfg CircuitBreakerConfig,
	opts ...CircuitBreakerOption,
) CircuitBreakerFactory {
	cache := make(map[string]CircuitBreaker)

	return func(name string) CircuitBreaker {
		if cb, ok := cache[name]; ok {
			return cb
		}

		cb := NewCircuitBreaker(name, cfg, opts...)
		cache[name] = cb
		return cb
	}
}

// CircuitBreakerPreset defines preset configurations for circuit breakers.
type CircuitBreakerPreset string

const (
	// CBPresetDatabase is a preset for database operations.
	CBPresetDatabase CircuitBreakerPreset = "database"
	// CBPresetExternalAPI is a preset for external API calls.
	CBPresetExternalAPI CircuitBreakerPreset = "external_api"
	// CBPresetDefault is the default preset for generic operations.
	CBPresetDefault CircuitBreakerPreset = "default"
)

// CircuitBreakerPresets provides pre-configured circuit breakers.
type CircuitBreakerPresets struct {
	factory CircuitBreakerFactory
}

// NewCircuitBreakerPresets creates presets with the given configuration and options.
func NewCircuitBreakerPresets(cfg CircuitBreakerConfig, opts ...CircuitBreakerOption) *CircuitBreakerPresets {
	return &CircuitBreakerPresets{
		factory: NewCircuitBreakerFactory(cfg, opts...),
	}
}

// ForDatabase returns a circuit breaker for database operations.
func (p *CircuitBreakerPresets) ForDatabase() CircuitBreaker {
	return p.factory(string(CBPresetDatabase))
}

// ForExternalAPI returns a circuit breaker for external API calls.
func (p *CircuitBreakerPresets) ForExternalAPI() CircuitBreaker {
	return p.factory(string(CBPresetExternalAPI))
}

// Default returns the default circuit breaker.
func (p *CircuitBreakerPresets) Default() CircuitBreaker {
	return p.factory(string(CBPresetDefault))
}

// ForOperation returns a circuit breaker for a named operation.
func (p *CircuitBreakerPresets) ForOperation(name string) CircuitBreaker {
	return p.factory(name)
}

// Factory returns the underlying factory function for use with ResilienceWrapper.
func (p *CircuitBreakerPresets) Factory() CircuitBreakerFactory {
	return p.factory
}
