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
	return fmt.Sprintf("config validation failed:\n- %s", strings.Join(e.Errors, "\n- "))
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
	errs = append(errs, c.validateRedis()...)
	errs = append(errs, c.validateAsynq()...)

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

// validateRedis checks Redis configuration (Story 8.1).
// Note: Redis is optional - validation only runs if REDIS_HOST is set.
// When REDIS_HOST is set, the port must be valid.
func (c *Config) validateRedis() []string {
	var errs []string

	// Redis is optional - only validate if host is configured
	if c.Redis.Host == "" {
		return errs // Skip validation if Redis not configured
	}

	// If Redis is configured, port must be valid
	if c.Redis.Port < 0 || c.Redis.Port > 65535 {
		errs = append(errs, "REDIS_PORT must be between 0 and 65535")
	}
	if c.Redis.DB < 0 || c.Redis.DB > 15 {
		errs = append(errs, "REDIS_DB must be between 0 and 15")
	}
	if c.Redis.PoolSize < 0 {
		errs = append(errs, "REDIS_POOL_SIZE must be >= 0")
	}
	if c.Redis.MinIdleConns < 0 {
		errs = append(errs, "REDIS_MIN_IDLE_CONNS must be >= 0")
	}

	return errs
}

// validateAsynq checks Asynq worker configuration (Story 8.2).
// Asynq is optional - validation only runs if any Asynq config is set.
func (c *Config) validateAsynq() []string {
	var errs []string

	if c.Asynq.Concurrency < 0 {
		errs = append(errs, "ASYNQ_CONCURRENCY must be >= 0")
	}
	if c.Asynq.ShutdownTimeout < 0 {
		errs = append(errs, "ASYNQ_SHUTDOWN_TIMEOUT must be >= 0")
	}

	return errs
}
