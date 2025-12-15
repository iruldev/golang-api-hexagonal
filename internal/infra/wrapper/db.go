package wrapper

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Default timeout constants
const (
	// DefaultQueryTimeout is applied when context has no deadline for DB operations.
	DefaultQueryTimeout = 30 * time.Second
	// DefaultHTTPTimeout is applied when context has no deadline for HTTP requests.
	DefaultHTTPTimeout = 30 * time.Second
)

// Querier interface defines the methods from pgxpool.Pool used by wrappers.
// This allows easier testing with mock implementations.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// Query wraps pool.Query with context timeout enforcement.
// If ctx has no deadline, DefaultQueryTimeout is applied.
// Returns immediately if context is already cancelled.
// The returned Rows are wrapped to ensure the timeout context is cancelled
// only when the Rows are closed.
func Query(ctx context.Context, pool Querier, query string, args ...any) (pgx.Rows, error) {
	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Add timeout only if no deadline is set
	var cancel context.CancelFunc
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, DefaultQueryTimeout)
	}

	rows, err := pool.Query(ctx, query, args...)
	if err != nil {
		if cancel != nil {
			cancel()
		}
		return nil, err
	}

	// If we created a cancel function, wrap the rows to ensure it's called on Close
	if cancel != nil {
		return &cancelRows{
			Rows:   rows,
			cancel: cancel,
		}, nil
	}

	return rows, nil
}

// cancelRows wraps pgx.Rows to call a specific cancel function on Close.
type cancelRows struct {
	pgx.Rows
	cancel context.CancelFunc
}

func (r *cancelRows) Close() {
	defer r.cancel()
	r.Rows.Close()
}

// QueryRow wraps pool.QueryRow with context timeout enforcement.
// If ctx has no deadline, DefaultQueryTimeout is applied.
// Returns a row that will cancel the context after Scan() is called.
func QueryRow(ctx context.Context, pool Querier, query string, args ...any) pgx.Row {
	// Add timeout only if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultQueryTimeout)
		return &timeoutRow{
			Row:    pool.QueryRow(ctx, query, args...),
			cancel: cancel,
		}
	}

	return pool.QueryRow(ctx, query, args...)
}

// timeoutRow wraps pgx.Row to ensure cancel is called after Scan.
type timeoutRow struct {
	pgx.Row
	cancel context.CancelFunc
}

func (r *timeoutRow) Scan(dest ...any) error {
	defer r.cancel()
	return r.Row.Scan(dest...)
}

// QueryRowWithCancel wraps pool.QueryRow and returns the cancel function.
// The caller MUST call cancel() after scanning the row.
// If ctx has no deadline, DefaultQueryTimeout is applied.
func QueryRowWithCancel(ctx context.Context, pool Querier, query string, args ...any) (pgx.Row, context.CancelFunc) {
	cancel := func() {} // noop cancel by default

	// Add timeout only if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, DefaultQueryTimeout)
	}

	return pool.QueryRow(ctx, query, args...), cancel
}

// Exec wraps pool.Exec with context timeout enforcement.
// If ctx has no deadline, DefaultQueryTimeout is applied.
// Returns immediately if context is already cancelled.
func Exec(ctx context.Context, pool Querier, query string, args ...any) (pgconn.CommandTag, error) {
	// Return early if context is already cancelled
	if err := ctx.Err(); err != nil {
		return pgconn.CommandTag{}, err
	}

	// Add timeout only if no deadline is set
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, DefaultQueryTimeout)
		defer cancel()
	}

	return pool.Exec(ctx, query, args...)
}

// PoolQuerier adapts *pgxpool.Pool to the Querier interface.
// This helper ensures type safety when passing the pool to wrapper functions.
func PoolQuerier(pool *pgxpool.Pool) Querier {
	return pool
}
