package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_MissingDBHost(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{Port: 5432, User: "test", Name: "testdb"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_HOST is required")
}

func TestValidate_MissingDBPort(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{Host: "localhost", User: "test", Name: "testdb"},
		// Port is 0 (zero value)
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_PORT must be between 1 and 65535")
}

func TestValidate_MissingDBUser(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{Host: "localhost", Port: 5432, Name: "testdb"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_USER is required")
}

func TestValidate_MissingDBName(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{Host: "localhost", Port: 5432, User: "test"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_NAME is required")
}

func TestValidate_InvalidHTTPPort_Negative(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: -1},
		Database: DatabaseConfig{Host: "localhost", Port: 5432, User: "test", Name: "testdb"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APP_HTTP_PORT must be between 1 and 65535")
}

func TestValidate_InvalidHTTPPort_TooHigh(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: 70000},
		Database: DatabaseConfig{Host: "localhost", Port: 5432, User: "test", Name: "testdb"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APP_HTTP_PORT must be between 1 and 65535")
}

func TestValidate_InvalidDBPort_TooHigh(t *testing.T) {
	cfg := &Config{
		App:      AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{Host: "localhost", Port: 70000, User: "test", Name: "testdb"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_PORT must be between 1 and 65535")
}

func TestValidate_InvalidAppEnv(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			HTTPPort: 8080,
			Env:      "invalid",
		},
		Database: DatabaseConfig{Host: "localhost", Port: 5432, User: "test", Name: "testdb"},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "APP_ENV must be one of: development, staging, production")
}

func TestValidate_ValidAppEnvValues(t *testing.T) {
	validEnvs := []string{"development", "staging", "production"}

	for _, env := range validEnvs {
		t.Run(env, func(t *testing.T) {
			cfg := &Config{
				App: AppConfig{
					Name:     "test-app",
					Env:      env,
					HTTPPort: 8080,
				},
				Database: DatabaseConfig{
					Host: "localhost",
					Port: 5432,
					User: "postgres",
					Name: "testdb",
				},
			}

			err := cfg.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestValidate_NegativeMaxOpenConns(t *testing.T) {
	cfg := &Config{
		App: AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "test",
			Name:         "testdb",
			MaxOpenConns: -5,
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_MAX_OPEN_CONNS must be >= 0")
}

func TestValidate_NegativeMaxIdleConns(t *testing.T) {
	cfg := &Config{
		App: AppConfig{HTTPPort: 8080},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "test",
			Name:         "testdb",
			MaxIdleConns: -5,
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DB_MAX_IDLE_CONNS must be >= 0")
}

func TestValidate_MultipleErrors(t *testing.T) {
	cfg := &Config{} // All empty/zero values

	err := cfg.Validate()
	require.Error(t, err)

	// Should collect ALL errors, not stop at first
	validErr, ok := err.(*ValidationError)
	require.True(t, ok, "error should be *ValidationError")
	// Expected errors: DB_HOST, DB_PORT, DB_USER, DB_NAME, APP_HTTP_PORT = 5 minimum
	assert.GreaterOrEqual(t, len(validErr.Errors), 5, "should collect at least 5 errors")

	// Verify all expected errors are present
	errStr := err.Error()
	assert.Contains(t, errStr, "DB_HOST is required")
	assert.Contains(t, errStr, "DB_PORT must be between 1 and 65535")
	assert.Contains(t, errStr, "DB_USER is required")
	assert.Contains(t, errStr, "DB_NAME is required")
	assert.Contains(t, errStr, "APP_HTTP_PORT must be between 1 and 65535")
}

func TestValidate_ValidConfig(t *testing.T) {
	cfg := &Config{
		App: AppConfig{
			Name:     "test-app",
			Env:      "development",
			HTTPPort: 8080,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Name:         "testdb",
			MaxOpenConns: 20,
			MaxIdleConns: 5,
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_ValidConfigMinimal(t *testing.T) {
	// Only required fields set
	cfg := &Config{
		App: AppConfig{
			HTTPPort: 8080,
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Name: "testdb",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_EmptyAppEnvIsValid(t *testing.T) {
	// Empty App.Env should be valid (optional field)
	cfg := &Config{
		App: AppConfig{
			HTTPPort: 8080,
			// Env is empty
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Name: "testdb",
		},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidationError_Is(t *testing.T) {
	err := &ValidationError{Errors: []string{"test error"}}

	// Test errors.Is() pattern works
	assert.True(t, errors.Is(err, &ValidationError{}))
}

func TestValidationError_ErrorMessage(t *testing.T) {
	err := &ValidationError{
		Errors: []string{"error1", "error2", "error3"},
	}

	msg := err.Error()
	assert.Contains(t, msg, "config validation failed:")
	assert.Contains(t, msg, "error1")
	assert.Contains(t, msg, "error2")
	assert.Contains(t, msg, "error3")
	assert.Contains(t, msg, "; ") // Semicolon separator
}

func TestLoad_WithInvalidConfig(t *testing.T) {
	// Test that Load() returns error for invalid config type
	// When APP_HTTP_PORT is "invalid", koanf returns a parsing error
	t.Setenv("APP_HTTP_PORT", "invalid") // koanf cannot parse as int
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_NAME", "testdb")

	cfg, err := Load()
	require.Error(t, err)
	assert.Nil(t, cfg)
	// koanf returns decode error for type mismatch
	assert.Contains(t, err.Error(), "cannot parse value as 'int'")
}

func TestLoad_WithValidConfig(t *testing.T) {
	// Test that Load() returns valid config when all required fields are set
	t.Setenv("APP_HTTP_PORT", "8080")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_NAME", "testdb")

	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 8080, cfg.App.HTTPPort)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test", cfg.Database.User)
	assert.Equal(t, "testdb", cfg.Database.Name)
}
