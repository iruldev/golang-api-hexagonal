// Package resilience provides fault tolerance patterns for the application.
// This file implements graceful shutdown coordination (Story 1.6).
package resilience

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

// ShutdownCoordinator coordinates graceful shutdown across the application.
// It tracks in-flight requests and ensures they complete before shutdown.
// Thread-safe for concurrent request tracking.
//
// The coordinator manages DrainPeriod internally via WaitForDrain.
// GracePeriod from ShutdownConfig is intended for use by the caller (e.g., Fx lifecycle)
// to add additional cleanup time after drain completes. The caller should retrieve
// GracePeriod via Config() and apply it in their shutdown sequence.
type ShutdownCoordinator interface {
	// IncrementActive increments the active request counter.
	// Returns false if shutdown has been initiated (caller should reject the request).
	IncrementActive() bool

	// DecrementActive decrements the active request counter.
	DecrementActive()

	// ActiveCount returns the current number of active requests.
	ActiveCount() int64

	// IsShuttingDown returns true if shutdown has been initiated.
	IsShuttingDown() bool

	// InitiateShutdown starts the shutdown process.
	// After this call, IncrementActive will return false.
	InitiateShutdown()

	// WaitForDrain waits for all active requests to complete or drain timeout.
	// Returns nil if all requests completed, error if timeout.
	WaitForDrain(ctx context.Context) error

	// Config returns the shutdown configuration.
	// Callers can use Config().GracePeriod for additional cleanup time after drain.
	Config() ShutdownConfig
}

// ShutdownOption configures the ShutdownCoordinator.
type ShutdownOption func(*shutdownCoordinator)

// WithShutdownMetrics configures the ShutdownCoordinator to record Prometheus metrics.
func WithShutdownMetrics(m *ShutdownMetrics) ShutdownOption {
	return func(s *shutdownCoordinator) {
		if m != nil {
			s.metrics = m
		}
	}
}

// WithShutdownLogger configures the ShutdownCoordinator to use a custom logger.
func WithShutdownLogger(l *slog.Logger) ShutdownOption {
	return func(s *shutdownCoordinator) {
		if l != nil {
			s.logger = l
		}
	}
}

// shutdownCoordinator implements the ShutdownCoordinator interface.
type shutdownCoordinator struct {
	cfg            ShutdownConfig
	shuttingDown   atomic.Bool
	activeRequests atomic.Int64
	metrics        *ShutdownMetrics
	logger         *slog.Logger
}

// NewShutdownCoordinator creates a new ShutdownCoordinator with the given configuration.
func NewShutdownCoordinator(cfg ShutdownConfig, opts ...ShutdownOption) ShutdownCoordinator {
	// Validate config at construction time
	if err := cfg.validate(); err != nil {
		panic(fmt.Sprintf("invalid shutdown config: %v", err))
	}

	s := &shutdownCoordinator{
		cfg:    cfg,
		logger: slog.Default(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// IncrementActive increments the active request counter.
// Returns false if shutdown has been initiated (caller should reject the request).
// Thread-safe: uses increment-then-check pattern to avoid race conditions.
func (s *shutdownCoordinator) IncrementActive() bool {
	// Increment first to avoid race between check and increment.
	// If shutdown happens between these lines, we'll detect and rollback.
	newCount := s.activeRequests.Add(1)

	// Check if we're shutting down - if so, rollback the increment
	if s.shuttingDown.Load() {
		s.activeRequests.Add(-1) // Rollback
		s.logRejection()
		if s.metrics != nil {
			s.metrics.RecordRejection()
		}
		return false
	}

	if s.logger != nil {
		s.logger.Debug("request started",
			"active_requests", newCount,
		)
	}

	if s.metrics != nil {
		s.metrics.SetActiveRequests(newCount)
	}

	return true
}

// DecrementActive decrements the active request counter.
// Safe to call even if counter would go negative (guards against misuse).
func (s *shutdownCoordinator) DecrementActive() {
	newCount := s.activeRequests.Add(-1)

	// Guard against negative counts from mismatched calls
	if newCount < 0 {
		s.activeRequests.CompareAndSwap(newCount, 0)
		if s.logger != nil {
			s.logger.Warn("active request count went negative, reset to 0",
				"previous_count", newCount,
			)
		}
		newCount = 0
	}

	if s.logger != nil {
		s.logger.Debug("request completed",
			"active_requests", newCount,
		)
	}

	if s.metrics != nil {
		s.metrics.SetActiveRequests(newCount)
	}
}

// ActiveCount returns the current number of active requests.
func (s *shutdownCoordinator) ActiveCount() int64 {
	return s.activeRequests.Load()
}

// IsShuttingDown returns true if shutdown has been initiated.
func (s *shutdownCoordinator) IsShuttingDown() bool {
	return s.shuttingDown.Load()
}

// InitiateShutdown starts the shutdown process.
// After this call, IncrementActive will return false.
func (s *shutdownCoordinator) InitiateShutdown() {
	wasShuttingDown := s.shuttingDown.Swap(true)
	if wasShuttingDown {
		// Already shutting down
		return
	}

	if s.logger != nil {
		s.logger.Info("shutdown initiated",
			"drain_period", s.cfg.DrainPeriod,
			"active_requests", s.activeRequests.Load(),
		)
	}

	if s.metrics != nil {
		s.metrics.SetShutdownInProgress(true)
	}
}

// WaitForDrain waits for all active requests to complete or drain timeout.
// Returns nil if all requests completed, error if timeout.
func (s *shutdownCoordinator) WaitForDrain(ctx context.Context) error {
	start := time.Now()

	// Create drain context with timeout
	drainCtx, cancel := context.WithTimeout(ctx, s.cfg.DrainPeriod)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		currentActive := s.activeRequests.Load()
		if currentActive <= 0 {
			duration := time.Since(start)
			if s.logger != nil {
				s.logger.Info("drain completed, all requests finished",
					"duration", duration,
				)
			}
			if s.metrics != nil {
				s.metrics.RecordShutdownDuration(duration, "success")
			}
			return nil
		}

		select {
		case <-drainCtx.Done():
			remaining := s.activeRequests.Load()
			duration := time.Since(start)
			if s.logger != nil {
				s.logger.Warn("drain timeout, forcing shutdown",
					"remaining_requests", remaining,
					"duration", duration,
				)
			}
			if s.metrics != nil {
				s.metrics.RecordShutdownDuration(duration, "timeout")
			}
			return fmt.Errorf("drain timeout: %d requests still active", remaining)
		case <-ticker.C:
			if s.logger != nil {
				s.logger.Debug("waiting for drain",
					"active_requests", s.activeRequests.Load(),
					"elapsed", time.Since(start),
				)
			}
		}
	}
}

// logRejection logs when a request is rejected during shutdown.
func (s *shutdownCoordinator) logRejection() {
	if s.logger != nil {
		s.logger.Warn("request rejected during shutdown",
			"active_requests", s.activeRequests.Load(),
		)
	}
}

// Config returns the shutdown configuration.
// Callers can use Config().GracePeriod for additional cleanup time after drain.
func (s *shutdownCoordinator) Config() ShutdownConfig {
	return s.cfg
}
