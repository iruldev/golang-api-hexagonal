package resilience

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/sethvargo/go-retry"
)

// Retrier provides retry with exponential backoff functionality.
// It automatically retries failed operations with configurable delays.
type Retrier interface {
	// Do executes the given function with retry logic.
	// It returns the last error if all attempts fail (wrapped as RES-004).
	Do(ctx context.Context, fn func(ctx context.Context) error) error

	// Name returns the name of this retrier for metrics/logging.
	Name() string
}

// retrier wraps go-retry with metrics and logging.
type retrier struct {
	name            string
	cfg             RetryConfig
	metrics         *RetryMetrics
	logger          *slog.Logger
	isRetryableFunc func(error) bool
}

// RetrierOption configures a retrier.
type RetrierOption func(*retrierOptions)

type retrierOptions struct {
	metrics         *RetryMetrics
	logger          *slog.Logger
	isRetryableFunc func(error) bool
}

// WithRetryMetrics sets the metrics for the retrier.
func WithRetryMetrics(m *RetryMetrics) RetrierOption {
	return func(o *retrierOptions) {
		o.metrics = m
	}
}

// WithRetryLogger sets the logger for the retrier.
// If l is nil, the default logger (slog.Default()) will be used.
func WithRetryLogger(l *slog.Logger) RetrierOption {
	return func(o *retrierOptions) {
		if l != nil {
			o.logger = l
		}
		// If nil, keep the default logger set in NewRetrier
	}
}

// WithRetryableFunc sets a custom function to determine if an error is retryable.
// If fn is nil, the default retryable function (DefaultIsRetryable) will be used.
func WithRetryableFunc(fn func(error) bool) RetrierOption {
	return func(o *retrierOptions) {
		if fn != nil {
			o.isRetryableFunc = fn
		}
		// If nil, keep the default isRetryableFunc set in NewRetrier
	}
}

// NewRetrier creates a new retrier with the given name and configuration.
// The retrier will retry failed operations with exponential backoff and jitter.
//
// Note: The underlying go-retry library uses base-2 exponential backoff (delay doubles
// each attempt). The RetryConfig.Multiplier field is validated but not currently used
// by this implementation - the multiplier is always 2.0. This field exists for API
// compatibility and potential future custom backoff implementations.
func NewRetrier(name string, cfg RetryConfig, opts ...RetrierOption) Retrier {
	options := &retrierOptions{
		metrics:         nil,
		logger:          slog.Default(),
		isRetryableFunc: DefaultIsRetryable,
	}

	for _, opt := range opts {
		opt(options)
	}

	return &retrier{
		name:            name,
		cfg:             cfg,
		metrics:         options.metrics,
		logger:          options.logger,
		isRetryableFunc: options.isRetryableFunc,
	}
}

// Do executes the given function with retry logic.
// It uses exponential backoff with jitter to determine retry delays.
// Context cancellation is respected and will stop retries immediately.
func (r *retrier) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	start := time.Now()
	attempt := 0
	var lastErr error

	// Build the backoff strategy:
	// 1. Start with exponential backoff from initial delay
	// 2. Add jitter to prevent thundering herd
	// 3. Cap delays at max delay
	// 4. Limit to max retries (MaxAttempts - 1, since first attempt is not a retry)
	backoff := retry.NewExponential(r.cfg.InitialDelay)

	// Calculate jitter as 25% of the delay to add randomization
	jitterDuration := r.cfg.InitialDelay / 4
	backoff = retry.WithJitter(jitterDuration, backoff)

	// Cap the delay at max delay
	backoff = retry.WithCappedDuration(r.cfg.MaxDelay, backoff)

	// Limit retries (WithMaxRetries takes number of RETRIES, not total attempts)
	// For MaxAttempts=3, we need 2 retries (first attempt + 2 retries)
	var maxRetries uint64
	if r.cfg.MaxAttempts > 1 {
		maxRetries = uint64(r.cfg.MaxAttempts - 1)
	}
	backoff = retry.WithMaxRetries(maxRetries, backoff)

	err := retry.Do(ctx, backoff, func(ctx context.Context) error {
		attempt++

		operationErr := fn(ctx)

		if operationErr == nil {
			// Success
			return nil
		}

		lastErr = operationErr

		// Check if context was cancelled
		if ctx.Err() != nil {
			return ctx.Err() // Don't retry on context cancellation
		}

		// Check if error is retryable
		if !r.isRetryableFunc(operationErr) {
			// Non-retryable error - stop immediately
			r.logger.Debug("non-retryable error, stopping retry",
				"name", r.name,
				"attempt", attempt,
				"error", operationErr,
			)
			return operationErr // Return unwrapped to stop retry
		}

		// Log retry attempt
		r.logger.Debug("operation failed, will retry",
			"name", r.name,
			"attempt", attempt,
			"max_attempts", r.cfg.MaxAttempts,
			"error", operationErr,
		)

		// Mark as retryable for go-retry
		return retry.RetryableError(operationErr)
	})

	duration := time.Since(start)

	// Handle the result
	if err == nil {
		// Success
		if r.metrics != nil {
			r.metrics.RecordOperation(r.name, "success", attempt, duration.Seconds())
		}
		if attempt > 1 {
			r.logger.Info("operation succeeded after retry",
				"name", r.name,
				"total_attempts", attempt,
				"duration_ms", duration.Milliseconds(),
			)
		}
		return nil
	}

	// Check if we exhausted all retries
	if attempt >= r.cfg.MaxAttempts {
		if r.metrics != nil {
			r.metrics.RecordOperation(r.name, "exhausted", attempt, duration.Seconds())
		}
		r.logger.Warn("max retries exceeded",
			"name", r.name,
			"total_attempts", attempt,
			"max_attempts", r.cfg.MaxAttempts,
			"duration_ms", duration.Milliseconds(),
			"last_error", lastErr,
		)
		return NewMaxRetriesExceededError(lastErr)
	}

	// Non-retryable error or context cancelled
	if r.metrics != nil {
		r.metrics.RecordOperation(r.name, "failure", attempt, duration.Seconds())
	}

	return err
}

// Name returns the name of this retrier.
func (r *retrier) Name() string {
	return r.name
}

// RetryableError is an interface for errors that indicate whether they are retryable.
type RetryableError interface {
	error
	Retryable() bool
}

// temporaryError is an interface for errors that indicate temporary failure.
type temporaryError interface {
	Temporary() bool
}

// DefaultIsRetryable returns true if the error should be retried.
// It checks for:
// - context.DeadlineExceeded (temporary timeout)
// - errors implementing Retryable() bool
// - errors implementing Temporary() bool
// - net.Error with Timeout() (network timeouts are retryable)
// It returns false for:
// - context.Canceled (user cancelled)
// - nil errors
func DefaultIsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Never retry cancelled context
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Retry deadline exceeded (temporary timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for RetryableError interface
	var retryable RetryableError
	if errors.As(err, &retryable) {
		return retryable.Retryable()
	}

	// Check for temporary error interface
	var tempErr temporaryError
	if errors.As(err, &tempErr) {
		return tempErr.Temporary()
	}

	// Check for net.Error with Timeout() - network timeouts are typically retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}

	// Default to retryable for unknown errors
	// This ensures transient failures are retried
	return true
}

// IsRetryable is a helper function that uses DefaultIsRetryable.
func IsRetryable(err error) bool {
	return DefaultIsRetryable(err)
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  DefaultRetryMaxAttempts,
		InitialDelay: DefaultRetryInitialDelay,
		MaxDelay:     DefaultRetryMaxDelay,
		Multiplier:   DefaultRetryMultiplier,
	}
}

// DoWithResult executes a function that returns data with retry logic.
// This is a helper function for functions that return both a result and an error.
func DoWithResult[T any](r Retrier, ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	err := r.Do(ctx, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = fn(ctx)
		return innerErr
	})
	return result, err
}
