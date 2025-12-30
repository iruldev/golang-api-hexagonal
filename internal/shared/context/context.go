// Package context provides cross-cutting context utilities for storing and
// retrieving request-scoped values such as Request IDs and Trace IDs.
//
// This package exists in the shared layer to allow both app and transport
// layers to access request context without violating architecture boundaries.
//
// Architecture rationale:
//   - Transport layer: Sets context values (request ID, trace ID, etc.)
//   - App layer: Reads context values for logging and tracing
//   - Infra layer: Uses context values for correlation
//
// Note: The transport/http/ctxutil package re-exports these functions
// for backward compatibility and convenience within the transport layer.
package context

import "context"

// Context key types - unexported to prevent collisions.
type (
	requestIDKey struct{}
	traceIDKey   struct{}
	spanIDKey    struct{}
)

// Empty ID constants for trace validation.
const (
	// EmptyTraceID is a 32-character string of zeros representing an invalid trace ID.
	EmptyTraceID = "00000000000000000000000000000000"
	// EmptySpanID is a 16-character string of zeros representing an invalid span ID.
	EmptySpanID = "0000000000000000"
)

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

// GetTraceID retrieves the trace ID from the context.
// Returns an empty string if no trace ID is present or if tracing is disabled.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey{}).(string); ok {
		return id
	}
	return ""
}

// SetTraceID returns a new context with the given trace ID.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// GetSpanID retrieves the span ID from the context.
// Returns an empty string if no span ID is present or if tracing is disabled.
func GetSpanID(ctx context.Context) string {
	if id, ok := ctx.Value(spanIDKey{}).(string); ok {
		return id
	}
	return ""
}

// SetSpanID returns a new context with the given span ID.
func SetSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey{}, spanID)
}
