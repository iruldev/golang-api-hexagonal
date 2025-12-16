// Package config provides environment-based configuration loading.
package config

import (
	"fmt"
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

	// OpenTelemetry
	OTELEnabled          bool   `envconfig:"OTEL_ENABLED" default:"false"`
	OTELExporterEndpoint string `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	OTELExporterInsecure bool   `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"false"`
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
