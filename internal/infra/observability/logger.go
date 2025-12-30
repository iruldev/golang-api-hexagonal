// Package observability provides logging, tracing, and metrics utilities.
package observability

import (
	"context"
	"log/slog"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/logger"
)

// Logger key constants - re-exported from shared/logger for backward compatibility.
// Deprecated: Use logger.Key* constants from shared/logger package instead.
const (
	LogKeyService   = logger.KeyService
	LogKeyEnv       = logger.KeyEnv
	LogKeyRequestID = logger.KeyRequestID
	LogKeyTraceID   = logger.KeyTraceID
	LogKeySpanID    = logger.KeySpanID
	LogKeyMethod    = logger.KeyMethod
	LogKeyRoute     = logger.KeyRoute
	LogKeyStatus    = logger.KeyStatus
	LogKeyDuration  = logger.KeyDuration
	LogKeyBytes     = logger.KeyBytes
)

// NewLogger creates a structured JSON logger with default attributes.
// The logger includes service and environment fields on every log entry.
// Log level is controlled via the LOG_LEVEL configuration.
func NewLogger(cfg *config.Config) *logger.Logger {
	level := parseLogLevel(cfg.LogLevel)

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	// Add default attributes that appear on every log entry
	l := slog.New(handler).With(
		logger.KeyService, cfg.ServiceName,
		logger.KeyEnv, cfg.Env,
	)

	return l
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
// Deprecated: Use logger.Any from shared/logger package instead.
func Any(key string, value interface{}) slog.Attr {
	return slog.Any(key, value)
}

// LoggerFromContext returns a logger enriched with request_id, trace_id, and span_id from context.
// If any ID is not present in context, that field is omitted from the logger.
// This enables request correlation across all log entries in a request lifecycle.
//
// Deprecated: Use logger.FromContext from shared/logger package instead.
func LoggerFromContext(ctx context.Context, base *logger.Logger) *logger.Logger {
	return logger.FromContext(ctx, base)
}
