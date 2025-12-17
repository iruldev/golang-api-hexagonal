package domain

import "context"

// TxManager provides transaction management for use cases that need
// atomicity across multiple repository operations.
//
// The implementation wraps driver-specific transaction handling and
// provides the transaction as a Querier to repository methods.
type TxManager interface {
	// WithTx executes the given function within a transaction.
	// If fn returns an error, the transaction is rolled back.
	// If fn succeeds, the transaction is committed.
	WithTx(ctx context.Context, fn func(tx Querier) error) error
}
