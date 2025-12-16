// Package postgres provides PostgreSQL database connectivity.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps a pgxpool.Pool and provides database operations.
type Pool struct {
	pool *pgxpool.Pool
}

// NewPool creates a new database connection pool.
func NewPool(ctx context.Context, databaseURL string) (*Pool, error) {
	const op = "postgres.NewPool"

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("%s: parse config: %w", op, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s: create pool: %w", op, err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: ping: %w", op, err)
	}

	return &Pool{pool: pool}, nil
}

// Ping verifies database connectivity.
func (p *Pool) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

// Close closes the database connection pool.
func (p *Pool) Close() {
	p.pool.Close()
}

// Pool returns the underlying pgxpool.Pool for queries.
func (p *Pool) Pool() *pgxpool.Pool {
	return p.pool
}
