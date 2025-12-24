// Package ctxutil provides context utility functions for storing and retrieving
// request-scoped values such as JWT claims and Request IDs.
package ctxutil

import "context"

// requestIDKey is the unexported type for the context key to prevent collisions.
type requestIDKey struct{}

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}

// SetRequestID returns a new context with the given request ID.
func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}
