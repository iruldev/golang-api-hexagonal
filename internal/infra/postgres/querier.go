// Package postgres provides PostgreSQL database adapters and repository implementations.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// rowScanner is a minimal interface for scanning a single row.
// Used for type assertions in repository code.
type rowScanner interface {
	Scan(dest ...any) error
}

// rowsScanner is a minimal interface for iterating over multiple rows.
// Used for type assertions in repository code.
type rowsScanner interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...any) error
}

// PoolQuerier wraps pgxpool.Pool to implement domain.Querier.
// Use this for regular database operations outside transactions.
type PoolQuerier struct {
	pool *pgxpool.Pool
}

// NewPoolQuerier creates a new PoolQuerier from a pgxpool.Pool.
func NewPoolQuerier(pool *pgxpool.Pool) domain.Querier {
	return &PoolQuerier{pool: pool}
}

// Exec executes a query that doesn't return rows.
func (q *PoolQuerier) Exec(ctx context.Context, sql string, args ...any) (any, error) {
	return q.pool.Exec(ctx, sql, args...)
}

// Query executes a query that returns rows.
func (q *PoolQuerier) Query(ctx context.Context, sql string, args ...any) (any, error) {
	return q.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row.
func (q *PoolQuerier) QueryRow(ctx context.Context, sql string, args ...any) any {
	return q.pool.QueryRow(ctx, sql, args...)
}

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
