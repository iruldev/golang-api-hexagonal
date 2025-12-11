package config

import (
	"fmt"
	"strings"
)

// validAppEnvs defines the allowed values for APP_ENV.
var validAppEnvs = map[string]bool{
	"development": true,
	"staging":     true,
	"production":  true,
}

// ValidationError holds multiple configuration validation errors.
type ValidationError struct {
	Errors []string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("config validation failed: %s", strings.Join(e.Errors, "; "))
}

// Is supports errors.Is() pattern for type checking.
func (e *ValidationError) Is(target error) bool {
	_, ok := target.(*ValidationError)
	return ok
}

// Validate checks configuration for required fields and valid ranges.
// Returns ValidationError with all validation errors collected (not just first).
func (c *Config) Validate() error {
	var errs []string

	errs = append(errs, c.validateDatabase()...)
	errs = append(errs, c.validateApp()...)

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// validateDatabase checks database configuration.
// Note: Password is NOT required to support trust-based local development
// where PostgreSQL may be configured without password authentication.
func (c *Config) validateDatabase() []string {
	var errs []string

	if c.Database.Host == "" {
		errs = append(errs, "DB_HOST is required")
	}
	if c.Database.Port <= 0 || c.Database.Port > 65535 {
		errs = append(errs, "DB_PORT must be between 1 and 65535")
	}
	if c.Database.User == "" {
		errs = append(errs, "DB_USER is required")
	}
	if c.Database.Name == "" {
		errs = append(errs, "DB_NAME is required")
	}
	if c.Database.MaxOpenConns < 0 {
		errs = append(errs, "DB_MAX_OPEN_CONNS must be >= 0")
	}
	if c.Database.MaxIdleConns < 0 {
		errs = append(errs, "DB_MAX_IDLE_CONNS must be >= 0")
	}

	return errs
}

// validateApp checks application configuration.
func (c *Config) validateApp() []string {
	var errs []string

	if c.App.HTTPPort <= 0 || c.App.HTTPPort > 65535 {
		errs = append(errs, "APP_HTTP_PORT must be between 1 and 65535")
	}

	// App.Env validation (optional but if set, must be valid)
	if c.App.Env != "" && !validAppEnvs[c.App.Env] {
		errs = append(errs, "APP_ENV must be one of: development, staging, production")
	}

	return errs
}
