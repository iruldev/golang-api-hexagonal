// Package observability provides logging, tracing, and metrics functionality.
package observability

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

// NewLogger creates a new zap logger based on configuration.
// Returns a production logger (JSON format) for production/staging environments,
// or a development logger (console format) otherwise.
func NewLogger(cfg *config.LogConfig, appEnv string) (*zap.Logger, error) {
	var zapConfig zap.Config

	// Use production config for production/staging
	if appEnv == "production" || appEnv == "staging" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Override format if specified in config
	switch cfg.Format {
	case "json":
		zapConfig.Encoding = "json"
	case "console":
		zapConfig.Encoding = "console"
	}

	// Set log level from config
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	return zapConfig.Build()
}

// NewNopLogger creates a no-op logger for testing.
func NewNopLogger() *zap.Logger {
	return zap.NewNop()
}
