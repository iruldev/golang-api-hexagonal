package ctxutil

import (
	"context"

	sharedctx "github.com/iruldev/golang-api-hexagonal/internal/shared/context"
)

// Empty ID constants - re-exported from shared/context for backward compatibility.
const (
	// EmptyTraceID is a 32-character string of zeros.
	EmptyTraceID = sharedctx.EmptyTraceID
	// EmptySpanID is a 16-character string of zeros.
	EmptySpanID = sharedctx.EmptySpanID
)

// GetTraceID retrieves the trace ID from the context.
// Returns an empty string if no trace ID is present or if tracing is disabled.
func GetTraceID(ctx context.Context) string {
	return sharedctx.GetTraceID(ctx)
}

// SetTraceID returns a new context with the given trace ID.
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return sharedctx.SetTraceID(ctx, traceID)
}

// GetSpanID retrieves the span ID from the context.
// Returns an empty string if no span ID is present or if tracing is disabled.
func GetSpanID(ctx context.Context) string {
	return sharedctx.GetSpanID(ctx)
}

// SetSpanID returns a new context with the given span ID.
func SetSpanID(ctx context.Context, spanID string) context.Context {
	return sharedctx.SetSpanID(ctx, spanID)
}
