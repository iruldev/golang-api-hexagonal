package domain

import "context"

// Querier is an abstraction for database operations that works with both
// connection pools and transactions. Implementations convert between
// this interface and driver-specific types.
//
// The any return types allow infra layer to return driver-specific types.
// Domain layer doesn't need to know the concrete types.
type Querier interface {
	// Exec executes a query that doesn't return rows (INSERT, UPDATE, DELETE).
	Exec(ctx context.Context, sql string, args ...any) (any, error)

	// Query executes a query that returns rows.
	Query(ctx context.Context, sql string, args ...any) (any, error)

	// QueryRow executes a query that returns at most one row.
	QueryRow(ctx context.Context, sql string, args ...any) any
}
