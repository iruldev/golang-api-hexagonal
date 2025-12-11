// Package postgres provides PostgreSQL database connection management.
package postgres

import (
	"context"
	"fmt"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool creates a new PostgreSQL connection pool.
// It configures the pool based on the application config and verifies
// the connection by pinging the database.
//
// Returns:
//   - *pgxpool.Pool: The connection pool
//   - error: Connection error or nil
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}

	// Validate pool settings
	maxConns := cfg.Database.MaxOpenConns
	minConns := cfg.Database.MaxIdleConns

	if maxConns <= 0 {
		maxConns = 25 // default
	}
	if minConns <= 0 {
		minConns = 5 // default
	}
	if minConns > maxConns {
		minConns = maxConns // MinConns can't exceed MaxConns
	}

	// Configure pool settings from config
	poolConfig.MaxConns = int32(maxConns)
	poolConfig.MinConns = int32(minConns)
	poolConfig.MaxConnLifetime = cfg.Database.ConnMaxLifetime

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
