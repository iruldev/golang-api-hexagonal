// Package config provides environment-based configuration loading.
package config

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config holds all configuration values for the application.
// Required fields will cause startup failure if not provided.
// Optional fields have sensible defaults.
type Config struct {
	// Required - Database connection string
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`

	// Optional with defaults
	Port        int    `envconfig:"PORT" default:"8080"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`
	Env         string `envconfig:"ENV" default:"development"`
	ServiceName string `envconfig:"SERVICE_NAME" default:"golang-api-hexagonal"`

	// Error response contract (RFC 7807)
	ProblemBaseURL string `envconfig:"PROBLEM_BASE_URL" default:"https://api.example.com/problems/"`

	// OpenTelemetry
	OTELEnabled          bool   `envconfig:"OTEL_ENABLED" default:"false"`
	OTELExporterEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTELExporterInsecure bool   `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"false"`

	// HTTP request handling
	// MaxRequestSize is the maximum request body size in bytes. Default: 1MB (1048576 bytes).
	MaxRequestSize int64 `envconfig:"MAX_REQUEST_SIZE" default:"1048576"`

	// JWT Authentication
	// JWTEnabled enables JWT authentication for protected endpoints. Default: false.
	JWTEnabled bool `envconfig:"JWT_ENABLED" default:"false"`
	// JWTSecret is the secret key for JWT signing (required if JWTEnabled=true).
	JWTSecret string `envconfig:"JWT_SECRET"`
	// JWTIssuer is the expected issuer claim (optional).
	JWTIssuer string `envconfig:"JWT_ISSUER"`
	// JWTAudience is the expected audience claim (optional).
	JWTAudience string `envconfig:"JWT_AUDIENCE"`
	// JWTClockSkew is the tolerance for expired tokens (optional). Default: 0s.
	JWTClockSkew time.Duration `envconfig:"JWT_CLOCK_SKEW" default:"0s"`

	// Rate Limiting
	// RateLimitRPS is the rate limit in requests per second. Default: 100.
	RateLimitRPS int `envconfig:"RATE_LIMIT_RPS" default:"100"`
	// TrustProxy enables trusting X-Forwarded-For/X-Real-IP headers. Default: false.
	TrustProxy bool `envconfig:"TRUST_PROXY" default:"false"`

	// Audit
	// AuditRedactEmail controls how email addresses are redacted in audit logs.
	// Options: "full" (default, replaces with [REDACTED]) or "partial" (shows first 2 chars + domain).
	AuditRedactEmail string `envconfig:"AUDIT_REDACT_EMAIL" default:"full"`
}

// Redacted returns a safe string representation of the Config for logging.
func (c *Config) Redacted() string {
	safe := *c
	safe.DatabaseURL = "[REDACTED]"
	safe.JWTSecret = "[REDACTED]"
	return fmt.Sprintf("%+v", safe)
}

// Load reads configuration from environment variables.
// It returns an error if required fields are missing.
func Load() (*Config, error) {
	const op = "config.Load"

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required and cannot be empty")
	}

	if c.OTELEnabled && strings.TrimSpace(c.OTELExporterEndpoint) == "" {
		return fmt.Errorf("OTEL_ENABLED is true but OTEL_EXPORTER_OTLP_ENDPOINT is empty")
	}

	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid PORT: must be between 1 and 65535")
	}
	if strings.TrimSpace(c.ServiceName) == "" {
		return fmt.Errorf("invalid SERVICE_NAME: must not be empty")
	}

	c.LogLevel = strings.ToLower(strings.TrimSpace(c.LogLevel))
	c.Env = strings.ToLower(strings.TrimSpace(c.Env))
	// Fix: Normalize JWTSecret by trimming whitespace and updating the struct
	c.JWTSecret = strings.TrimSpace(c.JWTSecret)
	c.AuditRedactEmail = strings.ToLower(strings.TrimSpace(c.AuditRedactEmail))

	switch c.Env {
	case "development", "staging", "production", "test":
	default:
		return fmt.Errorf("invalid ENV: must be one of development, staging, production, test")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid LOG_LEVEL: must be one of debug, info, warn, error")
	}

	if err := validateProblemBaseURL(c.ProblemBaseURL); err != nil {
		return err
	}

	if c.MaxRequestSize < 1 {
		return fmt.Errorf("invalid MAX_REQUEST_SIZE: must be greater than 0")
	}

	// Production environment requires JWT authentication (Story 2.3, Option B - Strict)
	// This prevents accidentally running without auth in production.
	if c.Env == "production" {
		if !c.JWTEnabled {
			return fmt.Errorf("ENV=production requires JWT_ENABLED=true")
		}
		// The generic check below will ensure JWTSecret is set and valid,
		// but we prefer a specific error message for production empty secret.
		if c.JWTSecret == "" {
			return fmt.Errorf("ENV=production requires JWT_SECRET to be set")
		}
	}

	if c.JWTEnabled {
		if c.JWTSecret == "" {
			return fmt.Errorf("JWT_ENABLED is true but JWT_SECRET is empty")
		}
		if len(c.JWTSecret) < 32 {
			return fmt.Errorf("JWT_SECRET must be at least 32 bytes when JWT_ENABLED is true")
		}
	}

	if c.RateLimitRPS < 1 {
		return fmt.Errorf("invalid RATE_LIMIT_RPS: must be greater than 0")
	}

	switch c.AuditRedactEmail {
	case "full", "partial":
	default:
		return fmt.Errorf("invalid AUDIT_REDACT_EMAIL: must be 'full' or 'partial'")
	}

	return nil
}

func validateProblemBaseURL(raw string) error {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("invalid PROBLEM_BASE_URL: must not be empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return fmt.Errorf("invalid PROBLEM_BASE_URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("invalid PROBLEM_BASE_URL: must be an absolute URL (scheme + host)")
	}
	if !strings.HasSuffix(trimmed, "/") {
		return fmt.Errorf("invalid PROBLEM_BASE_URL: must end with a trailing slash")
	}
	return nil
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}
