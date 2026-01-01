// Package postgres provides PostgreSQL database connectivity and repositories.
// This file implements the background cleanup service for expired idempotency records.
package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/sqlcgen"
)

// DefaultCleanupInterval is the default interval for cleanup job execution.
const DefaultCleanupInterval = time.Hour

// IdempotencyCleanerMetrics provides Prometheus metrics for the idempotency cleanup job.
type IdempotencyCleanerMetrics struct {
	// recordsDeleted tracks total records deleted by the cleanup job.
	recordsDeleted prometheus.Counter

	// cleanupDuration measures the duration of each cleanup execution.
	cleanupDuration prometheus.Histogram

	// cleanupErrors counts cleanup job errors.
	cleanupErrors prometheus.Counter
}

// NewIdempotencyCleanerMetrics creates and registers cleanup metrics with the given registry.
// If registry is nil, a new registry is created.
func NewIdempotencyCleanerMetrics(registry *prometheus.Registry) *IdempotencyCleanerMetrics {
	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	recordsDeleted := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "idempotency_cleanup_records_deleted_total",
			Help: "Total number of expired idempotency records deleted by the cleanup job",
		},
	)

	cleanupDuration := prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name: "idempotency_cleanup_duration_seconds",
			Help: "Duration of idempotency cleanup job execution",
			Buckets: []float64{
				0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0,
			},
		},
	)

	cleanupErrors := prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "idempotency_cleanup_errors_total",
			Help: "Total number of errors during idempotency cleanup job execution",
		},
	)

	// Register metrics with registry.
	// Errors are intentionally ignored as they indicate metrics are already registered.
	_ = registry.Register(recordsDeleted)
	_ = registry.Register(cleanupDuration)
	_ = registry.Register(cleanupErrors)

	return &IdempotencyCleanerMetrics{
		recordsDeleted:  recordsDeleted,
		cleanupDuration: cleanupDuration,
		cleanupErrors:   cleanupErrors,
	}
}

// NoopIdempotencyCleanerMetrics returns a no-op metrics implementation for testing.
func NoopIdempotencyCleanerMetrics() *IdempotencyCleanerMetrics {
	return NewIdempotencyCleanerMetrics(prometheus.NewRegistry())
}

// IdempotencyCleanerConfig holds configuration for the cleaner.
type IdempotencyCleanerConfig struct {
	// Interval is the duration between cleanup runs. Default: 1 hour.
	Interval time.Duration
}

// IdempotencyCleaner periodically removes expired idempotency records.
// It runs as a background goroutine and can be gracefully stopped.
type IdempotencyCleaner struct {
	pool     Pooler
	interval time.Duration
	logger   *slog.Logger
	metrics  *IdempotencyCleanerMetrics
	stopCh   chan struct{}
	doneCh   chan struct{}
}

// NewIdempotencyCleaner creates a new cleaner instance.
func NewIdempotencyCleaner(
	pool Pooler,
	cfg IdempotencyCleanerConfig,
	logger *slog.Logger,
	registry *prometheus.Registry,
) *IdempotencyCleaner {
	if cfg.Interval == 0 {
		cfg.Interval = DefaultCleanupInterval
	}

	metrics := NewIdempotencyCleanerMetrics(registry)

	return &IdempotencyCleaner{
		pool:     pool,
		interval: cfg.Interval,
		logger:   logger,
		metrics:  metrics,
		stopCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}
}

// Start begins the periodic cleanup loop.
// This method returns immediately; the cleanup runs in a background goroutine.
func (c *IdempotencyCleaner) Start(ctx context.Context) error {
	go c.run(ctx)
	c.logger.Info("idempotency cleaner started",
		"interval", c.interval.String(),
	)
	return nil
}

// Stop gracefully stops the cleaner.
// It waits for the current cleanup cycle to complete or until ctx is cancelled.
func (c *IdempotencyCleaner) Stop(ctx context.Context) error {
	close(c.stopCh)
	select {
	case <-c.doneCh:
		c.logger.Info("idempotency cleaner stopped")
		return nil
	case <-ctx.Done():
		c.logger.Warn("idempotency cleaner stop timed out")
		return ctx.Err()
	}
}

func (c *IdempotencyCleaner) run(ctx context.Context) {
	defer close(c.doneCh)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	c.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.cleanup(ctx)
		}
	}
}

func (c *IdempotencyCleaner) cleanup(ctx context.Context) {
	start := time.Now()

	pool := c.pool.Pool()
	if pool == nil {
		c.logger.Warn("idempotency cleanup skipped: database not connected")
		return
	}

	queries := sqlcgen.New(pool)

	deleted, err := queries.DeleteExpiredIdempotencyKeys(ctx)
	duration := time.Since(start)

	if err != nil {
		c.logger.Error("idempotency cleanup failed",
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		c.metrics.cleanupErrors.Inc()
		return
	}

	c.metrics.recordsDeleted.Add(float64(deleted))
	c.metrics.cleanupDuration.Observe(duration.Seconds())

	// Log only if records were deleted to reduce noise
	if deleted > 0 {
		c.logger.Info("idempotency cleanup completed",
			"deleted", deleted,
			"duration_ms", duration.Milliseconds(),
		)
	}
}
