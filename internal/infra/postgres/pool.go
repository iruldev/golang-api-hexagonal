// Package postgres provides PostgreSQL database connectivity.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps a pgxpool.Pool and provides database operations.
type Pool struct {
	pool *pgxpool.Pool
}

// PoolConfig holds database pool configuration.
type PoolConfig struct {
	// MaxConns is the maximum number of connections in the pool.
	MaxConns int32
	// MinConns is the minimum number of connections in the pool.
	MinConns int32
	// MaxConnLifetime is the maximum lifetime of a connection.
	MaxConnLifetime time.Duration
}

// NewPool creates a new database connection pool with the given configuration.
func NewPool(ctx context.Context, databaseURL string, poolCfg PoolConfig) (*Pool, error) {
	const op = "postgres.NewPool"

	config, err := getPGXPoolConfig(databaseURL, poolCfg)
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

// getPGXPoolConfig creates a pgxpool.Config with the applied custom settings.
// Exposed/Extracted for unit testing the logic without needing a DB connection.
func getPGXPoolConfig(databaseURL string, poolCfg PoolConfig) (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Apply pool configuration (Story 5.1)
	if poolCfg.MaxConns > 0 {
		config.MaxConns = poolCfg.MaxConns
	}
	if poolCfg.MinConns > 0 {
		config.MinConns = poolCfg.MinConns
	}
	if poolCfg.MaxConnLifetime > 0 {
		config.MaxConnLifetime = poolCfg.MaxConnLifetime
	}

	return config, nil
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
