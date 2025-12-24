package observability

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_JSONFormat(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer

	cfg := &config.Config{
		LogLevel:    "info",
		ServiceName: "test-service",
		Env:         "test",
	}

	// Create logger with custom output (test version)
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler).With(
		LogKeyService, cfg.ServiceName,
		LogKeyEnv, cfg.Env,
	)

	// Log a test message
	logger.Info("test message", "key", "value")

	// Parse the JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err, "log output should be valid JSON")

	// Verify required fields
	assert.Equal(t, "INFO", logEntry["level"])
	assert.Equal(t, "test message", logEntry["msg"])
	assert.Equal(t, "test-service", logEntry["service"])
	assert.Equal(t, "test", logEntry["env"])
	assert.Equal(t, "value", logEntry["key"])
	assert.NotEmpty(t, logEntry["time"], "time field should be present")
}

func TestNewLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name      string
		logLevel  string
		logFn     func(*slog.Logger)
		wantEmpty bool
	}{
		{
			name:      "debug level logs info",
			logLevel:  "debug",
			logFn:     func(l *slog.Logger) { l.Info("test") },
			wantEmpty: false,
		},
		{
			name:      "warn level filters info",
			logLevel:  "warn",
			logFn:     func(l *slog.Logger) { l.Info("test") },
			wantEmpty: true,
		},
		{
			name:      "error level filters warn",
			logLevel:  "error",
			logFn:     func(l *slog.Logger) { l.Warn("test") },
			wantEmpty: true,
		},
		{
			name:      "error level logs error",
			logLevel:  "error",
			logFn:     func(l *slog.Logger) { l.Error("test") },
			wantEmpty: false,
		},
		{
			name:      "unknown level defaults to info",
			logLevel:  "unknown",
			logFn:     func(l *slog.Logger) { l.Info("test") },
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer

			level := parseLogLevel(tt.logLevel)
			handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				Level: level,
			})
			logger := slog.New(handler)

			tt.logFn(logger)

			if tt.wantEmpty {
				assert.Empty(t, buf.String(), "log should be filtered")
			} else {
				assert.NotEmpty(t, buf.String(), "log should be emitted")
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"info", slog.LevelInfo},
		{"warn", slog.LevelWarn},
		{"error", slog.LevelError},
		{"unknown", slog.LevelInfo}, // default
		{"", slog.LevelInfo},        // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLogLevel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewLogger_RequiredFieldsPresent(t *testing.T) {
	var buf bytes.Buffer

	cfg := &config.Config{
		LogLevel:    "info",
		ServiceName: "my-service",
		Env:         "production",
	}

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler).With(
		LogKeyService, cfg.ServiceName,
		LogKeyEnv, cfg.Env,
	)

	logger.Info("test")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// RFC 3339 format time
	timeVal, ok := logEntry["time"].(string)
	assert.True(t, ok, "time should be a string")
	assert.NotEmpty(t, timeVal)

	// Level
	assert.Equal(t, "INFO", logEntry["level"])

	// Service and env
	assert.Equal(t, "my-service", logEntry["service"])
	assert.Equal(t, "production", logEntry["env"])
}

func TestLoggerFromContext_WithRequestID(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	baseLogger := slog.New(handler)

	// Set request ID in context
	ctx := ctxutil.SetRequestID(context.Background(), "test-request-123")

	// Get enriched logger
	enrichedLogger := LoggerFromContext(ctx, baseLogger)

	// Log a message
	enrichedLogger.Info("test message")

	// Parse and verify
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "test-request-123", logEntry["requestId"], "request_id should be present in log")
	assert.Equal(t, "test message", logEntry["msg"])
}

func TestLoggerFromContext_WithoutRequestID(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	baseLogger := slog.New(handler)

	// Empty context (no request ID)
	ctx := context.Background()

	// Get logger (should be base logger unchanged)
	resultLogger := LoggerFromContext(ctx, baseLogger)

	// Log a message
	resultLogger.Info("test message")

	// Parse and verify
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// request_id should NOT be present
	_, hasRequestID := logEntry["requestId"]
	assert.False(t, hasRequestID, "request_id should NOT be present when context is empty")
	assert.Equal(t, "test message", logEntry["msg"])
}

func TestLoggerFromContext_WithTraceID(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	baseLogger := slog.New(handler)

	// Set trace ID in context
	ctx := ctxutil.SetTraceID(context.Background(), "abc123def456789012345678901234")

	// Get enriched logger
	enrichedLogger := LoggerFromContext(ctx, baseLogger)

	// Log a message
	enrichedLogger.Info("test message")

	// Parse and verify
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "abc123def456789012345678901234", logEntry["traceId"], "trace_id should be present in log")
}

func TestLoggerFromContext_WithAllIDs(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	baseLogger := slog.New(handler)

	// Set all IDs in context
	ctx := context.Background()
	ctx = ctxutil.SetRequestID(ctx, "req-123")
	ctx = ctxutil.SetTraceID(ctx, "trace-abc-def")
	ctx = ctxutil.SetSpanID(ctx, "span-xyz-123")

	// Get enriched logger
	enrichedLogger := LoggerFromContext(ctx, baseLogger)
	enrichedLogger.Info("test message")

	// Parse and verify
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "req-123", logEntry["requestId"])
	assert.Equal(t, "trace-abc-def", logEntry["traceId"])
	assert.Equal(t, "span-xyz-123", logEntry[LogKeySpanID])
}

func TestLoggerFromContext_ZeroTraceIDFiltered(t *testing.T) {
	var buf bytes.Buffer

	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	baseLogger := slog.New(handler)

	// Set zero trace ID (should be filtered)
	ctx := ctxutil.SetTraceID(context.Background(), "00000000000000000000000000000000")

	enrichedLogger := LoggerFromContext(ctx, baseLogger)
	enrichedLogger.Info("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Zero trace ID should NOT appear
	_, hasTraceID := logEntry["traceId"]
	assert.False(t, hasTraceID, "zero trace_id should NOT be present")
}
