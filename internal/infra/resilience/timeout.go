package resilience

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// Timeout provides context-based timeout functionality.
// It wraps operations with a configured timeout duration.
type Timeout interface {
	// Do executes the given function with timeout.
	// It returns ErrTimeoutExceeded (RES-003) if the operation times out.
	Do(ctx context.Context, fn func(ctx context.Context) error) error

	// Name returns the name of this timeout for metrics/logging.
	Name() string

	// Duration returns the configured timeout duration.
	Duration() time.Duration
}

// timeout wraps context.WithTimeout with metrics and logging.
type timeout struct {
	name     string
	duration time.Duration
	metrics  *TimeoutMetrics
	logger   *slog.Logger
}

// TimeoutOption configures a timeout.
type TimeoutOption func(*timeoutOptions)

type timeoutOptions struct {
	metrics *TimeoutMetrics
	logger  *slog.Logger
}

// WithTimeoutMetrics sets the metrics for the timeout.
// If m is nil, metrics will not be recorded (noop behavior).
func WithTimeoutMetrics(m *TimeoutMetrics) TimeoutOption {
	return func(o *timeoutOptions) {
		if m != nil {
			o.metrics = m
		}
		// If nil, keep the default (nil) - metrics are optional
	}
}

// WithTimeoutLogger sets the logger for the timeout.
// If l is nil, the default logger (slog.Default()) will be used.
func WithTimeoutLogger(l *slog.Logger) TimeoutOption {
	return func(o *timeoutOptions) {
		if l != nil {
			o.logger = l
		}
		// If nil, keep the default logger set in NewTimeout
	}
}

// NewTimeout creates a new timeout wrapper with the given name and duration.
// Options can be used to configure metrics and logging.
func NewTimeout(name string, duration time.Duration, opts ...TimeoutOption) Timeout {
	options := &timeoutOptions{
		metrics: nil,
		logger:  slog.Default(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return &timeout{
		name:     name,
		duration: duration,
		metrics:  options.metrics,
		logger:   options.logger,
	}
}

// Do executes the given function with a timeout.
// If the operation times out, it returns ErrTimeoutExceeded (RES-003).
// If the parent context is cancelled, the cancellation is propagated.
// Context cancellation (context.Canceled) is NOT wrapped as timeout error.
func (t *timeout) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	start := time.Now()

	// Apply timeout
	ctx, cancel := context.WithTimeout(ctx, t.duration)
	defer cancel()

	// Execute operation
	err := fn(ctx)

	duration := time.Since(start)

	// Handle result
	if err != nil {
		// Check specifically for DeadlineExceeded (timeout)
		// context.Canceled should NOT be wrapped as timeout error
		if errors.Is(err, context.DeadlineExceeded) {
			t.recordMetrics("timeout", duration)
			t.logTimeout(duration)
			return NewTimeoutExceededError(err)
		}

		// Other errors (including context.Canceled) are passed through
		t.recordMetrics("error", duration)
		return err
	}

	// Success
	t.recordMetrics("success", duration)
	t.logSuccess(duration)
	return nil
}

// Name returns the name of this timeout.
func (t *timeout) Name() string {
	return t.name
}

// Duration returns the configured timeout duration.
func (t *timeout) Duration() time.Duration {
	return t.duration
}

// recordMetrics records the operation result to Prometheus metrics.
func (t *timeout) recordMetrics(result string, duration time.Duration) {
	if t.metrics != nil {
		t.metrics.RecordOperation(t.name, result, duration.Seconds())
	}
}

// logSuccess logs a successful timeout operation at DEBUG level.
func (t *timeout) logSuccess(duration time.Duration) {
	t.logger.Debug("operation completed within timeout",
		"name", t.name,
		"timeout_duration", t.duration.String(),
		"actual_duration_ms", duration.Milliseconds(),
		"result", "success",
	)
}

// logTimeout logs a timeout exceeded at WARN level.
func (t *timeout) logTimeout(duration time.Duration) {
	t.logger.Warn("operation exceeded timeout",
		"name", t.name,
		"timeout_duration", t.duration.String(),
		"actual_duration_ms", duration.Milliseconds(),
		"result", "timeout",
	)
}

// TimeoutPresets provides pre-configured timeouts from TimeoutConfig.
// It allows easy access to common timeout configurations.
type TimeoutPresets struct {
	database    Timeout
	externalAPI Timeout
	defaultT    Timeout
	opts        []TimeoutOption
}

// NewTimeoutPresets creates presets from TimeoutConfig.
// Options are applied to all created timeouts.
func NewTimeoutPresets(cfg TimeoutConfig, opts ...TimeoutOption) *TimeoutPresets {
	return &TimeoutPresets{
		database:    NewTimeout("database", cfg.Database, opts...),
		externalAPI: NewTimeout("external_api", cfg.ExternalAPI, opts...),
		defaultT:    NewTimeout("default", cfg.Default, opts...),
		opts:        opts,
	}
}

// ForDatabase returns a timeout configured for database operations.
func (p *TimeoutPresets) ForDatabase() Timeout {
	return p.database
}

// ForExternalAPI returns a timeout configured for external API calls.
func (p *TimeoutPresets) ForExternalAPI() Timeout {
	return p.externalAPI
}

// Default returns the default timeout.
func (p *TimeoutPresets) Default() Timeout {
	return p.defaultT
}

// ForOperation creates a custom timeout for a specific operation.
// This is useful for one-off timeouts that don't fit the predefined categories.
func (p *TimeoutPresets) ForOperation(name string, d time.Duration) Timeout {
	return NewTimeout(name, d, p.opts...)
}

// DatabaseDuration returns the database timeout duration.
func (p *TimeoutPresets) DatabaseDuration() time.Duration {
	return p.database.Duration()
}

// ExternalAPIDuration returns the external API timeout duration.
func (p *TimeoutPresets) ExternalAPIDuration() time.Duration {
	return p.externalAPI.Duration()
}

// DefaultDuration returns the default timeout duration.
func (p *TimeoutPresets) DefaultDuration() time.Duration {
	return p.defaultT.Duration()
}

// DoWithTimeout executes a function that returns data with timeout.
// This is a helper function for functions that return both a result and an error.
func DoWithTimeout[T any](t Timeout, ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	err := t.Do(ctx, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = fn(ctx)
		return innerErr
	})
	return result, err
}
