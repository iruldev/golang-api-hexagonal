package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_FromEnvVars(t *testing.T) {
	// Arrange: Set environment variables
	t.Setenv("APP_NAME", "test-app")
	t.Setenv("APP_ENV", "development")
	t.Setenv("APP_HTTP_PORT", "9090")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSL_MODE", "disable")
	t.Setenv("DB_MAX_OPEN_CONNS", "25")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("DB_CONN_MAX_LIFETIME", "1h")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	t.Setenv("OTEL_SERVICE_NAME", "test-service")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")

	// Act
	cfg, err := Load()

	// Assert
	require.NoError(t, err)

	// App config
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "development", cfg.App.Env)
	assert.Equal(t, 9090, cfg.App.HTTPPort)

	// Database config
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "postgres", cfg.Database.User)
	assert.Equal(t, "secret", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, 25, cfg.Database.MaxOpenConns)
	assert.Equal(t, 10, cfg.Database.MaxIdleConns)
	assert.Equal(t, time.Hour, cfg.Database.ConnMaxLifetime)

	// Observability config
	assert.Equal(t, "http://localhost:4317", cfg.Observability.ExporterEndpoint)
	assert.Equal(t, "test-service", cfg.Observability.ServiceName)

	// Log config
	assert.Equal(t, "debug", cfg.Log.Level)
	assert.Equal(t, "json", cfg.Log.Format)
}

func TestLoad_PartialEnvVars(t *testing.T) {
	// Arrange: Set only required env vars
	t.Setenv("APP_HTTP_PORT", "8080")
	t.Setenv("DB_HOST", "db.example.com")

	// Act
	cfg, err := Load()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.App.HTTPPort)
	assert.Equal(t, "db.example.com", cfg.Database.Host)

	// Unset values should be zero values
	assert.Equal(t, "", cfg.App.Name)
	assert.Equal(t, 0, cfg.Database.Port)
}

func TestLoad_EmptyEnv(t *testing.T) {
	// Arrange: No env vars set (use default test isolation)
	// Note: t.Setenv not called, so OS env vars may leak
	// This test verifies Load() doesn't fail on empty config

	// Act
	cfg, err := Load()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	// All values should be zero values
	assert.Equal(t, 0, cfg.App.HTTPPort)
	assert.Equal(t, "", cfg.Database.Host)
}
