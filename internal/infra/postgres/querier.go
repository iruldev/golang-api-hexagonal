// Package postgres provides PostgreSQL database adapters and repository implementations.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// PoolQuerier wraps Pooler to implement domain.Querier.
// Use this for regular database operations outside transactions.
type PoolQuerier struct {
	pool Pooler
}

// NewPoolQuerier creates a new PoolQuerier from a Pooler.
func NewPoolQuerier(pool Pooler) domain.Querier {
	return &PoolQuerier{pool: pool}
}

// Exec executes a query that doesn't return rows.
func (q *PoolQuerier) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	pool := q.pool.Pool()
	if pool == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return pool.Exec(ctx, sql, args...)
}

// Query executes a query that returns rows.
func (q *PoolQuerier) Query(ctx context.Context, sql string, args ...any) (any, error) {
	pool := q.pool.Pool()
	if pool == nil {
		return nil, fmt.Errorf("database not connected")
	}
	return pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row.
func (q *PoolQuerier) QueryRow(ctx context.Context, sql string, args ...any) any {
	pool := q.pool.Pool()
	if pool == nil {
		// Used in context where we can't easily return error (QueryRow returns Row scanner).
		// We return a dummy scanner that returns error on Scan.
		return errRow{err: fmt.Errorf("database not connected")}
	}
	return pool.QueryRow(ctx, sql, args...)
}

type errRow struct {
	err error
}

func (e errRow) Scan(dest ...any) error { return e.err }

// TxQuerier wraps pgx.Tx to implement domain.Querier.
// Use this within transactions via TxManager.WithTx.
type TxQuerier struct {
	tx pgx.Tx
}

// NewTxQuerier creates a new TxQuerier from a pgx.Tx.
func NewTxQuerier(tx pgx.Tx) domain.Querier {
	return &TxQuerier{tx: tx}
}

// Exec executes a query that doesn't return rows within the transaction.
func (q *TxQuerier) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	return q.tx.Exec(ctx, sql, args...)
}

// Query executes a query that returns rows within the transaction.
func (q *TxQuerier) Query(ctx context.Context, sql string, args ...any) (any, error) {
	return q.tx.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row within the transaction.
func (q *TxQuerier) QueryRow(ctx context.Context, sql string, args ...any) any {
	return q.tx.QueryRow(ctx, sql, args...)
}
