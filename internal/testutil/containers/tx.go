package containers

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx runs the function within a transaction that is rolled back after.
// Provides test isolation without truncate overhead.
//
// Example usage:
//
//	func TestCreateUser(t *testing.T) {
//	    pool := containers.NewPostgres(t)
//	    containers.Migrate(t, pool)
//
//	    containers.WithTx(t, pool, func(tx pgx.Tx) {
//	        // All changes here are rolled back after test
//	        repo := postgres.NewUserRepo(tx)
//	        // ... test
//	    })
//	}
func WithTx(t testing.TB, pool *pgxpool.Pool, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("failed to rollback: %v", err)
		}
	}()

	fn(tx)
}

// WithTxContext is like WithTx but uses the provided context.
func WithTxContext(t testing.TB, ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx)) {
	t.Helper()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			t.Errorf("failed to rollback: %v", err)
		}
	}()

	fn(tx)
}
