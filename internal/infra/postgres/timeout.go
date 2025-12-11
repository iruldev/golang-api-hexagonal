package postgres

import (
	"context"
	"time"
)

// QueryContext returns a context with query timeout.
// Uses the provided timeout or DefaultQueryTimeout (30s) if zero.
//
// Example:
//
//	ctx, cancel := postgres.QueryContext(r.Context(), cfg.Database.QueryTimeout)
//	defer cancel()
//	rows, err := pool.Query(ctx, "SELECT * FROM users")
func QueryContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout == 0 {
		timeout = DefaultQueryTimeout
	}
	return context.WithTimeout(ctx, timeout)
}
