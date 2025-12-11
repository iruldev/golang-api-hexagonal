package observability_test

import (
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger_Production(t *testing.T) {
	cfg := &config.LogConfig{
		Level:  "info",
		Format: "json",
	}

	logger, err := observability.NewLogger(cfg, "production")
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_Development(t *testing.T) {
	cfg := &config.LogConfig{
		Level:  "debug",
		Format: "console",
	}

	logger, err := observability.NewLogger(cfg, "development")
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_Staging(t *testing.T) {
	cfg := &config.LogConfig{
		Level:  "warn",
		Format: "json",
	}

	logger, err := observability.NewLogger(cfg, "staging")
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := &config.LogConfig{
		Level:  "invalid",
		Format: "json",
	}

	// Should not error, defaults to info
	logger, err := observability.NewLogger(cfg, "development")
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewNopLogger(t *testing.T) {
	logger := observability.NewNopLogger()
	assert.NotNil(t, logger)
}
