package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_Success(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "postgres://user:pass@localhost:5432/testdb", cfg.DatabaseURL)
}

func TestLoad_InvalidRateLimitRPS(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("RATE_LIMIT_RPS", "0")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "RATE_LIMIT_RPS")
	assert.Contains(t, err.Error(), "greater than 0")
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port, "PORT should default to 8080")
	assert.Equal(t, "info", cfg.LogLevel, "LOG_LEVEL should default to info")
	assert.Equal(t, "development", cfg.Env, "ENV should default to development")
	assert.Equal(t, "golang-api-hexagonal", cfg.ServiceName, "SERVICE_NAME should default to golang-api-hexagonal")
	assert.Equal(t, "https://api.example.com/problems/", cfg.ProblemBaseURL, "PROBLEM_BASE_URL should default to https://api.example.com/problems/")
	assert.Equal(t, 100, cfg.RateLimitRPS, "RATE_LIMIT_RPS should default to 100")
	assert.False(t, cfg.TrustProxy, "TRUST_PROXY should default to false")
}

func TestLoad_CustomValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("ENV", "production")
	t.Setenv("SERVICE_NAME", "my-custom-service")
	t.Setenv("PROBLEM_BASE_URL", "https://my-custom-service.example/problems/")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "my-custom-service", cfg.ServiceName)
	assert.Equal(t, "https://my-custom-service.example/problems/", cfg.ProblemBaseURL)
	// We didn't set them in env, so they should be defaults, but let's test setting them separately if needed or just assume existing test is fine.
	// Actually, let's update the test to set them.
}

func TestLoad_InvalidProblemBaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("PROBLEM_BASE_URL", "not-a-url")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PROBLEM_BASE_URL")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_ProblemBaseURLMustEndWithSlash(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("PROBLEM_BASE_URL", "https://example.com/problems")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PROBLEM_BASE_URL")
	assert.Contains(t, err.Error(), "trailing slash")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_LogLevelUppercase(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("LOG_LEVEL", "WARN")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "warn", cfg.LogLevel, "LOG_LEVEL should be normalized to lowercase")
}

func TestLoad_MissingRequired(t *testing.T) {
	// Don't set DATABASE_URL, and ensure it's unset in case it's in the process env
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("PORT", "not-a-number")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_InvalidPortRange(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("PORT", "0")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid PORT")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_InvalidEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("ENV", "dev")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ENV")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("LOG_LEVEL", "verbose")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid LOG_LEVEL")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestLoad_InvalidServiceName(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("SERVICE_NAME", "   ")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SERVICE_NAME")
	assert.Contains(t, err.Error(), "config.Load")
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"development env", "development", true},
		{"production env", "production", false},
		{"staging env", "staging", false},
		{"empty env", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			assert.Equal(t, tt.want, cfg.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{"production env", "production", true},
		{"development env", "development", false},
		{"staging env", "staging", false},
		{"empty env", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Env: tt.env}
			assert.Equal(t, tt.want, cfg.IsProduction())
		})
	}
}

func TestLoad_InvalidAuditRedactEmail(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("AUDIT_REDACT_EMAIL", "invalid")

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AUDIT_REDACT_EMAIL")
	assert.Contains(t, err.Error(), "'full' or 'partial'")
}

func TestLoad_AuditRedactEmailValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"full mode", "full"},
		{"partial mode", "partial"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
			t.Setenv("AUDIT_REDACT_EMAIL", tt.value)

			cfg, err := Load()

			require.NoError(t, err)
			assert.Equal(t, tt.value, cfg.AuditRedactEmail)
		})
	}
}

func TestLoad_AuditRedactEmailDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	// Don't set AUDIT_REDACT_EMAIL to test default

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "full", cfg.AuditRedactEmail, "AUDIT_REDACT_EMAIL should default to 'full'")
}
