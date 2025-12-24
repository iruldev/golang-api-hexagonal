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
	// Production requires JWT (Story 2.3)
	t.Setenv("JWT_ENABLED", "true")
	t.Setenv("JWT_SECRET", "this-is-a-32-byte-secret-key!!@@")

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "production", cfg.Env)
	assert.Equal(t, "my-custom-service", cfg.ServiceName)
	assert.Equal(t, "https://my-custom-service.example/problems/", cfg.ProblemBaseURL)
	assert.True(t, cfg.JWTEnabled)
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

// =============================================================================
// Story 2.3: Production Guard Tests
// =============================================================================

// TestLoad_ProductionRequiresJWTEnabled tests AC #1, #2: production requires JWT_ENABLED=true
func TestLoad_ProductionRequiresJWTEnabled(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("ENV", "production")
	t.Setenv("JWT_ENABLED", "false") // Not enabled in production
	// JWT_SECRET is irrelevant here because we fail on JWT_ENABLED=false first

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ENV=production requires JWT_ENABLED=true")
}

// TestLoad_ProductionRequiresJWTSecret tests AC #1: production requires JWT_SECRET
func TestLoad_ProductionRequiresJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("ENV", "production")
	t.Setenv("JWT_ENABLED", "true")
	t.Setenv("JWT_SECRET", "") // Empty secret in production

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ENV=production requires JWT_SECRET to be set")
}

// TestLoad_ProductionWithValidJWT tests AC #2: production with valid JWT passes
func TestLoad_ProductionWithValidJWT(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("ENV", "production")
	t.Setenv("JWT_ENABLED", "true")
	t.Setenv("JWT_SECRET", "this-is-a-32-byte-secret-key!!@@")

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "production", cfg.Env)
	assert.True(t, cfg.JWTEnabled)
}

// TestLoad_DevelopmentAllowsNoJWT tests development can run without JWT
func TestLoad_DevelopmentAllowsNoJWT(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("ENV", "development")
	t.Setenv("JWT_ENABLED", "false")
	t.Setenv("JWT_SECRET", "") // Empty is OK in development

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "development", cfg.Env)
	assert.False(t, cfg.JWTEnabled)
}

// =============================================================================
// Story 2.4: JWT Secret Length Tests
// =============================================================================

// TestLoad_JWTSecretTooShort tests AC #1: JWT secret < 32 bytes fails
func TestLoad_JWTSecretTooShort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("JWT_ENABLED", "true")
	t.Setenv("JWT_SECRET", "short-secret-20-byte") // 20 bytes

	cfg, err := Load()

	assert.Nil(t, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least 32 bytes")
}

// TestLoad_JWTSecretExactly32Bytes tests boundary: 32 bytes is valid
func TestLoad_JWTSecretExactly32Bytes(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("JWT_ENABLED", "true")
	t.Setenv("JWT_SECRET", "12345678901234567890123456789012") // 32 bytes exactly

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.JWTEnabled)
}

// TestLoad_JWTSecretOver32Bytes tests JWT secret > 32 bytes is valid
func TestLoad_JWTSecretOver32Bytes(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("JWT_ENABLED", "true")
	t.Setenv("JWT_SECRET", "this-is-a-very-long-secret-key-that-exceeds-32-bytes") // 52 bytes

	cfg, err := Load()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.JWTEnabled)
}

// TestLoad_JWTSecretNormalization tests that whitespace is trimmed from secret
func TestLoad_JWTSecretNormalization(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("JWT_ENABLED", "true")
	// 32 bytes surrounded by spaces
	secret := "   12345678901234567890123456789012   "
	t.Setenv("JWT_SECRET", secret)

	cfg, err := Load()

	require.NoError(t, err)
	assert.Equal(t, "12345678901234567890123456789012", cfg.JWTSecret, "Secret should be trimmed")
	assert.Len(t, cfg.JWTSecret, 32)
}

// TestConfig_Redacted tests that sensitive fields are redacted
func TestConfig_Redacted(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://user:pass@localhost:5432/secret_db",
		JWTSecret:   "extremely-sensitive-secret-key-123",
		Port:        8080,
	}

	redacted := cfg.Redacted()

	assert.NotContains(t, redacted, "pass")
	assert.NotContains(t, redacted, "secret_db")
	assert.NotContains(t, redacted, "extremely-sensitive")
	assert.Contains(t, redacted, "[REDACTED]")
	assert.Contains(t, redacted, "8080")
}
