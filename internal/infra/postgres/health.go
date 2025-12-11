// Package postgres provides PostgreSQL database infrastructure.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DBHealthChecker checks database health.
type DBHealthChecker interface {
	Ping(ctx context.Context) error
}

// PoolHealthChecker implements DBHealthChecker for pgxpool.
type PoolHealthChecker struct {
	pool *pgxpool.Pool
}

// NewPoolHealthChecker creates a new PoolHealthChecker.
func NewPoolHealthChecker(pool *pgxpool.Pool) *PoolHealthChecker {
	if pool == nil {
		return nil
	}
	return &PoolHealthChecker{pool: pool}
}

// Ping checks database connectivity.
func (c *PoolHealthChecker) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
