package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTempConfigFile creates a temporary config file for testing.
func createTempConfigFile(t *testing.T, ext, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "config."+ext)
	err := os.WriteFile(filePath, []byte(content), 0600)
	require.NoError(t, err)
	return filePath
}

func TestLoad_FromYAMLFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, "yaml", `
app:
  http_port: 8080
  name: test-from-yaml
db:
  host: db.example.com
  port: 5432
`)
	t.Setenv("APP_CONFIG_FILE", tmpFile)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.App.HTTPPort)
	assert.Equal(t, "test-from-yaml", cfg.App.Name)
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
}

func TestLoad_FromJSONFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, "json", `{
  "app": {
    "http_port": 9000,
    "name": "test-from-json"
  },
  "db": {
    "host": "json-db.example.com"
  }
}`)
	t.Setenv("APP_CONFIG_FILE", tmpFile)

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9000, cfg.App.HTTPPort)
	assert.Equal(t, "test-from-json", cfg.App.Name)
	assert.Equal(t, "json-db.example.com", cfg.Database.Host)
}

func TestLoad_EnvOverridesFile(t *testing.T) {
	tmpFile := createTempConfigFile(t, "yaml", `
app:
  http_port: 8080
  name: from-file
db:
  host: file-db.example.com
`)
	t.Setenv("APP_CONFIG_FILE", tmpFile)
	t.Setenv("APP_HTTP_PORT", "9090")         // Override!
	t.Setenv("DB_HOST", "env-db.example.com") // Override!

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.App.HTTPPort)                  // Env wins
	assert.Equal(t, "from-file", cfg.App.Name)               // File value preserved
	assert.Equal(t, "env-db.example.com", cfg.Database.Host) // Env wins
}

func TestLoad_NoConfigFile(t *testing.T) {
	// No APP_CONFIG_FILE set, only env vars
	t.Setenv("APP_HTTP_PORT", "8080")
	t.Setenv("DB_HOST", "localhost")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.App.HTTPPort)
	assert.Equal(t, "localhost", cfg.Database.Host)
}

func TestLoad_FileNotFound(t *testing.T) {
	t.Setenv("APP_CONFIG_FILE", "/nonexistent/config.yaml")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to load config file")
}

func TestLoad_UnsupportedFormat(t *testing.T) {
	tmpFile := createTempConfigFile(t, "toml", `[app]
http_port = 8080
`)
	t.Setenv("APP_CONFIG_FILE", tmpFile)

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "unsupported config file format")
}
