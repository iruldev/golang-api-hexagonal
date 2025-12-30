// Package logger provides cross-cutting logging types for use across layers.
// This package exists to allow the app layer to use structured logging
// without directly importing infrastructure packages or log/slog.
//
// Architecture rationale:
//   - Domain layer: No logging (pure business logic)
//   - App layer: Can use this package for logging type references
//   - Transport layer: Can use this package or log/slog directly
//   - Infra layer: Implements concrete loggers using this type
//
// The Logger type is a type alias for slog.Logger, allowing seamless
// integration with Go's standard structured logging while maintaining
// clean architecture boundaries.
package logger

import (
	"context"
	"log/slog"

	sharedctx "github.com/iruldev/golang-api-hexagonal/internal/shared/context"
)

// Logger is a type alias for slog.Logger.
// This allows layers that cannot import log/slog directly (like app layer)
// to reference the logger type through this shared package.
//
// Usage in app layer:
//
//	type MyUseCase struct {
//	    logger *logger.Logger
//	}
//
// Usage in infra layer (creating loggers):
//
//	func NewLogger() *logger.Logger {
//	    return slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	}
type Logger = slog.Logger

// Attr is a type alias for slog.Attr for structured logging attributes.
type Attr = slog.Attr

// Level is a type alias for slog.Level for log levels.
type Level = slog.Level

// Log level constants.
const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

// Attribute constructors - re-exported from slog for convenience.
var (
	String   = slog.String
	Int      = slog.Int
	Int64    = slog.Int64
	Float64  = slog.Float64
	Bool     = slog.Bool
	Duration = slog.Duration
	Time     = slog.Time
	Any      = slog.Any
	Group    = slog.Group
)

// Log key constants for consistent log field names across the application.
const (
	KeyService   = "service"
	KeyEnv       = "env"
	KeyRequestID = "requestId"
	KeyTraceID   = "traceId"
	KeySpanID    = "spanId"
	KeyMethod    = "method"
	KeyRoute     = "route"
	KeyStatus    = "status"
	KeyDuration  = "duration_ms"
	KeyBytes     = "bytes"
)

// FromContext returns a logger enriched with request_id, trace_id, and span_id from context.
// If any ID is not present in context, that field is omitted from the logger.
// This enables request correlation across all log entries in a request lifecycle.
//
// Usage:
//
//	logger := logger.FromContext(ctx, baseLogger)
//	logger.Info("processing request")
func FromContext(ctx context.Context, base *Logger) *Logger {
	enriched := base
	if requestID := sharedctx.GetRequestID(ctx); requestID != "" {
		enriched = enriched.With(KeyRequestID, requestID)
	}
	if traceID := sharedctx.GetTraceID(ctx); traceID != "" && traceID != sharedctx.EmptyTraceID {
		enriched = enriched.With(KeyTraceID, traceID)
	}
	if spanID := sharedctx.GetSpanID(ctx); spanID != "" && spanID != sharedctx.EmptySpanID {
		enriched = enriched.With(KeySpanID, spanID)
	}
	return enriched
}
