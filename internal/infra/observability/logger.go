// Package observability provides logging, tracing, and metrics utilities.
package observability

import (
	"context"
	"log/slog"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// Logger key constants for consistent log field names.
const (
	LogKeyService   = "service"
	LogKeyEnv       = "env"
	LogKeyRequestID = "requestId"
	LogKeyTraceID   = "traceId"
	LogKeySpanID    = "spanId"
	LogKeyMethod    = "method"
	LogKeyRoute     = "route"
	LogKeyStatus    = "status"
	LogKeyDuration  = "duration_ms"
	LogKeyBytes     = "bytes"
)

// NewLogger creates a structured JSON logger with default attributes.
// The logger includes service and environment fields on every log entry.
// Log level is controlled via the LOG_LEVEL configuration.
func NewLogger(cfg *config.Config) *slog.Logger {
	level := parseLogLevel(cfg.LogLevel)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	// Add default attributes that appear on every log entry
	logger := slog.New(handler).With(
		LogKeyService, cfg.ServiceName,
		LogKeyEnv, cfg.Env,
	)

	return logger
}

// parseLogLevel converts a log level string to slog.Level.
// Defaults to Info level for unknown values.
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Any wraps a value for structured logging, similar to slog.Any
// but ensuring consistent serialization (e.g., for slices).
func Any(key string, value interface{}) slog.Attr {
	return slog.Any(key, value)
}

// LoggerFromContext returns a logger enriched with request_id, trace_id, and span_id from context.
// If any ID is not present in context, that field is omitted from the logger.
// This enables request correlation across all log entries in a request lifecycle.
func LoggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	enriched := base
	if requestID := ctxutil.GetRequestID(ctx); requestID != "" {
		enriched = enriched.With(LogKeyRequestID, requestID)
	}
	if traceID := ctxutil.GetTraceID(ctx); traceID != "" && traceID != ctxutil.EmptyTraceID {
		enriched = enriched.With(LogKeyTraceID, traceID)
	}
	if spanID := ctxutil.GetSpanID(ctx); spanID != "" && spanID != ctxutil.EmptySpanID {
		enriched = enriched.With(LogKeySpanID, spanID)
	}
	return enriched
}
