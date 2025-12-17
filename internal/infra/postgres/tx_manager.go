package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// TxManager implements domain.TxManager for PostgreSQL transactions.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a new TxManager from a pgxpool.Pool.
func NewTxManager(pool *pgxpool.Pool) domain.TxManager {
	return &TxManager{pool: pool}
}

// WithTx executes the given function within a database transaction.
// If fn returns an error or panics, the transaction is rolled back.
// If fn succeeds, the transaction is committed.
func (m *TxManager) WithTx(ctx context.Context, fn func(tx domain.Querier) error) (err error) {
	const op = "TxManager.WithTx"

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin: %w", op, err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic; re-panic after rollback attempt
			_ = tx.Rollback(ctx)
			panic(p)
		}
		if err != nil {
			// Rollback on error
			_ = tx.Rollback(ctx)
			return
		}
		// Commit on success
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("%s: commit: %w", op, commitErr)
		}
	}()

	return fn(NewTxQuerier(tx))
}
