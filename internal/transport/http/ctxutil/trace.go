package ctxutil

import "context"

// traceIDKey is the unexported type for the context key to prevent collisions.
type traceIDKey struct{}

// spanIDKey is the unexported type for the span ID context key.
type spanIDKey struct{}

const (
	// EmptyTraceID is a 32-character string of zeros.
	EmptyTraceID = "00000000000000000000000000000000"
	// EmptySpanID is a 16-character string of zeros.
	EmptySpanID = "0000000000000000"
)

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
