package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
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
