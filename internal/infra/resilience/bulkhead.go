package resilience

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// Bulkhead provides concurrency limiting functionality using the bulkhead pattern.
// It limits the number of concurrent executions to prevent resource exhaustion.
//
// Bulkhead implementations are safe for concurrent use from multiple goroutines.
// All methods can be called simultaneously without external synchronization.
type Bulkhead interface {
	// Do executes the given function within bulkhead constraints.
	// It returns ErrBulkheadFull (RES-002) if capacity and waiting queue are full.
	Do(ctx context.Context, fn func(ctx context.Context) error) error

	// Name returns the name of this bulkhead for metrics/logging.
	Name() string

	// ActiveCount returns the current number of active executions.
	ActiveCount() int

	// WaitingCount returns the current number of waiting operations.
	WaitingCount() int
}

// bulkhead implements the Bulkhead interface using a buffered channel as semaphore.
type bulkhead struct {
	name       string
	maxConc    int
	maxWaiting int
	semaphore  chan struct{}
	metrics    *BulkheadMetrics
	logger     *slog.Logger

	// Atomic counters for thread-safe metrics
	active  atomic.Int64
	waiting atomic.Int64
}

// BulkheadOption configures a bulkhead.
type BulkheadOption func(*bulkheadOptions)

type bulkheadOptions struct {
	metrics *BulkheadMetrics
	logger  *slog.Logger
}

// WithBulkheadMetrics sets the metrics for the bulkhead.
// If m is nil, metrics will not be recorded (noop behavior).
func WithBulkheadMetrics(m *BulkheadMetrics) BulkheadOption {
	return func(o *bulkheadOptions) {
		if m != nil {
			o.metrics = m
		}
		// If nil, keep the default (nil) - metrics are optional
	}
}

// WithBulkheadLogger sets the logger for the bulkhead.
// If l is nil, the default logger (slog.Default()) will be used.
func WithBulkheadLogger(l *slog.Logger) BulkheadOption {
	return func(o *bulkheadOptions) {
		if l != nil {
			o.logger = l
		}
		// If nil, keep the default logger set in NewBulkhead
	}
}

// NewBulkhead creates a new bulkhead with the given name and configuration.
// Options can be used to configure metrics and logging.
// Panics if configuration is invalid (MaxConcurrent < 1 or MaxWaiting < 0).
func NewBulkhead(name string, cfg BulkheadConfig, opts ...BulkheadOption) Bulkhead {
	if cfg.MaxConcurrent < 1 {
		panic("resilience: bulkhead MaxConcurrent must be >= 1")
	}
	if cfg.MaxWaiting < 0 {
		panic("resilience: bulkhead MaxWaiting must be >= 0")
	}

	options := &bulkheadOptions{
		metrics: nil,
		logger:  slog.Default(),
	}

	for _, opt := range opts {
		opt(options)
	}

	return &bulkhead{
		name:       name,
		maxConc:    cfg.MaxConcurrent,
		maxWaiting: cfg.MaxWaiting,
		semaphore:  make(chan struct{}, cfg.MaxConcurrent),
		metrics:    options.metrics,
		logger:     options.logger,
	}
}

// Do executes the given function within bulkhead constraints.
// If the bulkhead is at capacity and waiting queue is full, returns ErrBulkheadFull (RES-002).
// If context is cancelled while waiting, returns the context error.
func (b *bulkhead) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	// Try to acquire slot immediately (non-blocking)
	select {
	case b.semaphore <- struct{}{}:
		// Got slot immediately
		return b.executeWithSlot(ctx, fn, false)
	default:
		// Semaphore full, need to wait or reject
	}

	// Check if waiting queue is full using atomic compare
	for {
		currentWaiting := b.waiting.Load()
		if currentWaiting >= int64(b.maxWaiting) {
			// Waiting queue is full, reject
			b.recordMetrics("rejected", false)
			b.logRejection()
			return NewBulkheadFullError(nil)
		}

		// Try to increment waiting count atomically
		if b.waiting.CompareAndSwap(currentWaiting, currentWaiting+1) {
			break
		}
		// CAS failed, retry the loop
	}

	waitStart := time.Now()

	// Wait for slot with context
	select {
	case b.semaphore <- struct{}{}:
		// Got slot after waiting
		b.waiting.Add(-1)
		waitDuration := time.Since(waitStart)
		return b.executeWithSlot(ctx, fn, true, waitDuration)
	case <-ctx.Done():
		// Context cancelled while waiting
		b.waiting.Add(-1)
		waitDuration := time.Since(waitStart)
		b.recordMetrics("cancelled", false)
		b.logCancelled(waitDuration)
		return ctx.Err()
	}
}

// executeWithSlot executes the function with the slot already acquired.
// It ensures the slot is released regardless of panics.
func (b *bulkhead) executeWithSlot(ctx context.Context, fn func(ctx context.Context) error, waited bool, waitDuration ...time.Duration) error {
	b.active.Add(1)
	start := time.Now()

	// Record wait duration if applicable
	var wd time.Duration
	if len(waitDuration) > 0 {
		wd = waitDuration[0]
		b.recordWaitDuration(wd)
	}

	// Ensure slot is released even on panic
	defer func() {
		<-b.semaphore
		b.active.Add(-1)
	}()

	// Execute the operation
	err := fn(ctx)

	duration := time.Since(start)

	if err != nil {
		b.recordMetrics("error", true)
		b.logError(err, waited, wd, duration)
		return err
	}

	b.recordMetrics("success", true)
	b.logSuccess(waited, wd, duration)
	return nil
}

// Name returns the name of this bulkhead.
func (b *bulkhead) Name() string {
	return b.name
}

// ActiveCount returns the current number of active executions.
func (b *bulkhead) ActiveCount() int {
	return int(b.active.Load())
}

// WaitingCount returns the current number of waiting operations.
func (b *bulkhead) WaitingCount() int {
	return int(b.waiting.Load())
}

// recordMetrics records the operation result to Prometheus metrics.
// postOp indicates if this is called after operation completion (success/error),
// in which case active count is adjusted to show post-operation state for consistency with logs.
func (b *bulkhead) recordMetrics(result string, postOp bool) {
	if b.metrics != nil {
		b.metrics.RecordOperation(b.name, result)
		active := int(b.active.Load())
		if postOp {
			active-- // Show post-operation state for consistency with logs
		}
		b.metrics.SetActive(b.name, active)
		b.metrics.SetWaiting(b.name, int(b.waiting.Load()))
	}
}

// recordWaitDuration records waiting time to Prometheus metrics.
func (b *bulkhead) recordWaitDuration(d time.Duration) {
	if b.metrics != nil {
		b.metrics.RecordWaitDuration(b.name, d.Seconds())
	}
}

// logSuccess logs a successful operation at DEBUG level.
// Note: active_count shows post-operation state (current - 1) since slot will be released.
func (b *bulkhead) logSuccess(waited bool, waitDuration, execDuration time.Duration) {
	b.logger.Debug("bulkhead operation completed",
		"name", b.name,
		"max_concurrent", b.maxConc,
		"active_count", b.active.Load()-1,
		"waiting_count", b.waiting.Load(),
		"waited", waited,
		"wait_duration_ms", waitDuration.Milliseconds(),
		"exec_duration_ms", execDuration.Milliseconds(),
		"result", "success",
	)
}

// logError logs a failed operation at DEBUG level.
// Note: active_count shows post-operation state (current - 1) since slot will be released.
func (b *bulkhead) logError(err error, waited bool, waitDuration, execDuration time.Duration) {
	b.logger.Debug("bulkhead operation failed",
		"name", b.name,
		"max_concurrent", b.maxConc,
		"active_count", b.active.Load()-1,
		"waiting_count", b.waiting.Load(),
		"waited", waited,
		"wait_duration_ms", waitDuration.Milliseconds(),
		"exec_duration_ms", execDuration.Milliseconds(),
		"result", "error",
		"error", err.Error(),
	)
}

// logRejection logs a rejected operation at WARN level.
func (b *bulkhead) logRejection() {
	b.logger.Warn("bulkhead operation rejected - capacity full",
		"name", b.name,
		"max_concurrent", b.maxConc,
		"max_waiting", b.maxWaiting,
		"active_count", b.active.Load(),
		"waiting_count", b.waiting.Load(),
		"result", "rejected",
	)
}

// logCancelled logs a cancelled operation at DEBUG level.
func (b *bulkhead) logCancelled(waitDuration time.Duration) {
	b.logger.Debug("bulkhead operation cancelled while waiting",
		"name", b.name,
		"max_concurrent", b.maxConc,
		"max_waiting", b.maxWaiting,
		"active_count", b.active.Load(),
		"waiting_count", b.waiting.Load(),
		"wait_duration_ms", waitDuration.Milliseconds(),
		"result", "cancelled",
	)
}

// BulkheadPresets provides pre-configured bulkheads from BulkheadConfig.
// It allows easy access to common bulkhead configurations.
type BulkheadPresets struct {
	database    Bulkhead
	externalAPI Bulkhead
	defaultB    Bulkhead
	opts        []BulkheadOption
}

// NewBulkheadPresets creates presets from BulkheadConfig.
// Options are applied to all created bulkheads.
// Note: All presets share the same config values by default.
// For different limits per operation type, use ForOperation.
func NewBulkheadPresets(cfg BulkheadConfig, opts ...BulkheadOption) *BulkheadPresets {
	return &BulkheadPresets{
		database:    NewBulkhead("database", cfg, opts...),
		externalAPI: NewBulkhead("external_api", cfg, opts...),
		defaultB:    NewBulkhead("default", cfg, opts...),
		opts:        opts,
	}
}

// ForDatabase returns a bulkhead configured for database operations.
func (p *BulkheadPresets) ForDatabase() Bulkhead {
	return p.database
}

// ForExternalAPI returns a bulkhead configured for external API calls.
func (p *BulkheadPresets) ForExternalAPI() Bulkhead {
	return p.externalAPI
}

// Default returns the default bulkhead.
func (p *BulkheadPresets) Default() Bulkhead {
	return p.defaultB
}

// ForOperation creates a custom bulkhead for a specific operation.
// This is useful for one-off bulkheads that don't fit the predefined categories.
func (p *BulkheadPresets) ForOperation(name string, maxConcurrent, maxWaiting int, opts ...BulkheadOption) Bulkhead {
	cfg := BulkheadConfig{MaxConcurrent: maxConcurrent, MaxWaiting: maxWaiting}
	// Combine preset options with additional options (safe copy to avoid modifying original)
	allOpts := make([]BulkheadOption, 0, len(p.opts)+len(opts))
	allOpts = append(allOpts, p.opts...)
	allOpts = append(allOpts, opts...)
	return NewBulkhead(name, cfg, allOpts...)
}

// DoWithBulkhead executes a function that returns data within bulkhead constraints.
// This is a helper function for functions that return both a result and an error.
func DoWithBulkhead[T any](b Bulkhead, ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	err := b.Do(ctx, func(ctx context.Context) error {
		var innerErr error
		result, innerErr = fn(ctx)
		return innerErr
	})
	return result, err
}
