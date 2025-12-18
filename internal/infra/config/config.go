// Package config provides environment-based configuration loading.
package config

import (
	"fmt"
	"net/url"
	"strings"

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
