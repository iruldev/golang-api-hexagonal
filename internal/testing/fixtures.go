package testing

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupTestDatabase creates a connection pool to the test database
// and runs migrations. Returns the pool and a cleanup function.
func SetupTestDatabase(ctx context.Context, dsn string) (*pgxpool.Pool, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("create pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("ping database: %w", err)
	}

	// Run migrations (create tables for test)
	if err := runTestMigrations(ctx, pool); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("run migrations: %w", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup, nil
}

// runTestMigrations creates the necessary tables for integration tests.
// This mirrors the production migrations but is self-contained for tests.
func runTestMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Create notes table (matches production schema from db/migrations)
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS notes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			title VARCHAR(255) NOT NULL,
			content TEXT,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create notes table: %w", err)
	}

	return nil
}

// CleanupTestData removes all test data from tables.
// Call this between tests for isolation.
func CleanupTestData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, "TRUNCATE TABLE notes CASCADE")
	return err
}
