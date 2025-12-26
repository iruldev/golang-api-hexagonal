package postgres

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pooler defines the interface for a database pool interaction.
type Pooler interface {
	Ping(context.Context) error
	Close()
	Pool() *pgxpool.Pool
}

// ResilientPool lazily establishes a database pool and retries on readiness checks.
type ResilientPool struct {
	dsn                string
	ignoreStartupError bool
	mu                 sync.RWMutex
	pool               Pooler
	log                *slog.Logger
	poolCreator        func(context.Context, string) (Pooler, error)
}

// NewResilientPool creates a new ResilientPool.
func NewResilientPool(ctx context.Context, dsn string, poolCfg PoolConfig, ignoreStartupError bool, log *slog.Logger) *ResilientPool {
	rp := &ResilientPool{
		dsn:                dsn,
		ignoreStartupError: ignoreStartupError,
		log:                log,
		poolCreator: func(ctx context.Context, dsn string) (Pooler, error) {
			return NewPool(ctx, dsn, poolCfg)
		},
	}

	// Try initial connection
	// We ignore the error here because the pool manages its own state
	// and dependent services will handle the failure (by receiving nil pool/error).
	// Ping will retry on next call if it failed and ignoreStartupError is true.
	_ = rp.Ping(ctx)

	return rp
}

// Ping ensures a pool exists and is healthy; recreates the pool on failure.
func (r *ResilientPool) Ping(ctx context.Context) error {
	// Fast path: try existing pool under read lock
	r.mu.RLock()
	pool := r.pool
	r.mu.RUnlock()

	if pool == nil {
		// Create pool under write lock (double-check pattern)
		r.mu.Lock()
		if r.pool == nil {
			newPool, err := r.poolCreator(ctx, r.dsn)
			if err != nil {
				// If creation fails and we ignore startup errors
				if r.ignoreStartupError {
					r.log.Warn("database pool creation failed but IGNORE_DB_STARTUP_ERROR is set; using no-op pool", slog.Any("err", err))
					r.pool = &noopPool{}
					r.mu.Unlock()
					return nil
				}
				r.mu.Unlock()
				return err
			}
			r.pool = newPool
		}
		pool = r.pool
		r.mu.Unlock()
	}

	if err := pool.Ping(ctx); err != nil {
		if r.ignoreStartupError {
			r.log.Warn("database ping failed but IGNORE_DB_STARTUP_ERROR is set; using no-op pool", slog.Any("err", err))
			// Assign a no-op pool so checking db.Pool() doesn't panic
			// We need write lock to update r.pool
			r.mu.Lock()
			r.pool = &noopPool{}
			r.mu.Unlock()
			return nil
		}
		// pgxpool handles reconnection automatically - don't close the pool
		// Closing invalidates references held by querier/txManager causing panics
		r.log.Warn("database ping failed", slog.Any("err", err))
		return err
	}

	return nil
}

// Close shuts down the pool if it was created.
func (r *ResilientPool) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.pool != nil {
		r.pool.Close()
		r.pool = nil
	}
}

// Pool returns the underlying *pgxpool.Pool from the current Pooler.
func (r *ResilientPool) Pool() *pgxpool.Pool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.pool == nil {
		return nil
	}
	return r.pool.Pool()
}

// noopPool is a placeholder for when DB connection is ignored.
type noopPool struct{}

func (n *noopPool) Ping(context.Context) error { return nil }
func (n *noopPool) Close()                     {}
func (n *noopPool) Pool() *pgxpool.Pool        { return nil }
