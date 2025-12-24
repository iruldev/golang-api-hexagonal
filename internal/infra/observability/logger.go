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

// LoggerFromContext returns a logger enriched with request_id from context.
// If request_id is not present in context, returns the base logger unchanged.
// This enables request correlation across all log entries in a request lifecycle.
func LoggerFromContext(ctx context.Context, base *slog.Logger) *slog.Logger {
	if requestID := ctxutil.GetRequestID(ctx); requestID != "" {
		return base.With(LogKeyRequestID, requestID)
	}
	return base
}
