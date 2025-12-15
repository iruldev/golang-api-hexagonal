package postgres

import (
	"context"
	"time"
)

// QueryContext returns a context with query timeout.
// Uses the provided timeout or DefaultQueryTimeout (30s) if zero.
//
// Deprecated: Use wrapper.Query, wrapper.QueryRow, or wrapper.Exec from
// internal/infra/wrapper package instead. The wrapper package provides
// additional benefits:
//   - Automatic timeout only when no deadline is set (preserves existing deadlines)
//   - Early return if context is already cancelled
//   - Consistent interface for DB, HTTP, and Redis operations
//
// Example migration:
//
//	// Old:
//	ctx, cancel := postgres.QueryContext(r.Context(), cfg.Database.QueryTimeout)
//	defer cancel()
//	rows, err := pool.Query(ctx, "SELECT * FROM users")
//
//	// New:
//	rows, err := wrapper.Query(ctx, pool, "SELECT * FROM users")
func QueryContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultQueryTimeout
	}
	return context.WithTimeout(ctx, timeout)
}
