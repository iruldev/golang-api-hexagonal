// Package ctxutil provides context utility functions for storing and retrieving
// request-scoped values such as JWT claims and Request IDs.
//
// This package re-exports functions from shared/context for backward compatibility
// and convenience within the transport layer.
package ctxutil

import (
	"context"

	sharedctx "github.com/iruldev/golang-api-hexagonal/internal/shared/context"
)

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	return sharedctx.GetRequestID(ctx)
}

// SetRequestID returns a new context with the given request ID.
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return sharedctx.SetRequestID(ctx, requestID)
}
