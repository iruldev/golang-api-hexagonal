package resilience

import (
	"fmt"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

// Default values for resilience configuration.
const (
	// Circuit Breaker defaults
	DefaultCBMaxRequests      = 3
	DefaultCBInterval         = 10 * time.Second
	DefaultCBTimeout          = 30 * time.Second
	DefaultCBFailureThreshold = 5

	// Retry defaults
	DefaultRetryMaxAttempts  = 3
	DefaultRetryInitialDelay = 100 * time.Millisecond
	DefaultRetryMaxDelay     = 5 * time.Second
	DefaultRetryMultiplier   = 2.0

	// Timeout defaults
	DefaultTimeoutDefault     = 30 * time.Second
	DefaultTimeoutDatabase    = 5 * time.Second
	DefaultTimeoutExternalAPI = 10 * time.Second

	// Bulkhead defaults
	DefaultBulkheadMaxConcurrent = 10
	DefaultBulkheadMaxWaiting    = 100

	// Shutdown defaults (Story 1.6)
	DefaultShutdownDrainPeriod = 30 * time.Second
	DefaultShutdownGracePeriod = 5 * time.Second
)

// ResilienceConfig holds all resilience-related configuration.
type ResilienceConfig struct {
	CircuitBreaker CircuitBreakerConfig
	Retry          RetryConfig
	Timeout        TimeoutConfig
	Bulkhead       BulkheadConfig
	Shutdown       ShutdownConfig
}

// CircuitBreakerConfig holds configuration for circuit breaker pattern.
type CircuitBreakerConfig struct {
	// MaxRequests is the number of requests allowed in the half-open state.
	MaxRequests int
	// Interval is the cyclic period for clearing internal counts.
	Interval time.Duration
	// Timeout is the period to wait before transitioning from open to half-open.
	Timeout time.Duration
	// FailureThreshold is the number of failures to trip the circuit.
	FailureThreshold int
}

// RetryConfig holds configuration for retry with exponential backoff.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts.
	MaxAttempts int
	// InitialDelay is the initial delay before the first retry.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// Multiplier is the factor by which the delay increases after each retry.
	Multiplier float64
}

// TimeoutConfig holds configuration for operation timeouts.
type TimeoutConfig struct {
	// Default is the default timeout for operations.
	Default time.Duration
	// Database is the timeout for database operations.
	Database time.Duration
	// ExternalAPI is the timeout for external API calls.
	ExternalAPI time.Duration
}

// BulkheadConfig holds configuration for bulkhead pattern (Story 1.5).
type BulkheadConfig struct {
	// MaxConcurrent is the maximum number of concurrent executions.
	MaxConcurrent int
	// MaxWaiting is the maximum number of operations waiting for execution.
	MaxWaiting int
}

// ShutdownConfig holds configuration for graceful shutdown (Story 1.6).
type ShutdownConfig struct {
	// DrainPeriod is the maximum time to wait for in-flight requests to complete.
	// After this period, remaining requests will be forcefully terminated.
	DrainPeriod time.Duration
	// GracePeriod is additional time after drain for cleanup operations.
	GracePeriod time.Duration
}

// DefaultResilienceConfig returns a new ResilienceConfig with sensible defaults.
func DefaultResilienceConfig() ResilienceConfig {
	return ResilienceConfig{
		CircuitBreaker: CircuitBreakerConfig{
			MaxRequests:      DefaultCBMaxRequests,
			Interval:         DefaultCBInterval,
			Timeout:          DefaultCBTimeout,
			FailureThreshold: DefaultCBFailureThreshold,
		},
		Retry: RetryConfig{
			MaxAttempts:  DefaultRetryMaxAttempts,
			InitialDelay: DefaultRetryInitialDelay,
			MaxDelay:     DefaultRetryMaxDelay,
			Multiplier:   DefaultRetryMultiplier,
		},
		Timeout: TimeoutConfig{
			Default:     DefaultTimeoutDefault,
			Database:    DefaultTimeoutDatabase,
			ExternalAPI: DefaultTimeoutExternalAPI,
		},
		Bulkhead: BulkheadConfig{
			MaxConcurrent: DefaultBulkheadMaxConcurrent,
			MaxWaiting:    DefaultBulkheadMaxWaiting,
		},
		Shutdown: ShutdownConfig{
			DrainPeriod: DefaultShutdownDrainPeriod,
			GracePeriod: DefaultShutdownGracePeriod,
		},
	}
}

// NewResilienceConfig creates a ResilienceConfig from the main application Config.
func NewResilienceConfig(cfg *config.Config) ResilienceConfig {
	return ResilienceConfig{
		CircuitBreaker: CircuitBreakerConfig{
			MaxRequests:      cfg.CBMaxRequests,
			Interval:         cfg.CBInterval,
			Timeout:          cfg.CBTimeout,
			FailureThreshold: cfg.CBFailureThreshold,
		},
		Retry: RetryConfig{
			MaxAttempts:  cfg.RetryMaxAttempts,
			InitialDelay: cfg.RetryInitialDelay,
			MaxDelay:     cfg.RetryMaxDelay,
			Multiplier:   cfg.RetryMultiplier,
		},
		Timeout: TimeoutConfig{
			Default:     cfg.TimeoutDefault,
			Database:    cfg.TimeoutDatabase,
			ExternalAPI: cfg.TimeoutExternalAPI,
		},
		Bulkhead: BulkheadConfig{
			MaxConcurrent: cfg.BulkheadMaxConcurrent,
			MaxWaiting:    cfg.BulkheadMaxWaiting,
		},
		Shutdown: ShutdownConfig{
			DrainPeriod: cfg.ShutdownDrainPeriod,
			GracePeriod: cfg.ShutdownGracePeriod,
		},
	}
}

// Validate checks if the configuration is valid.
// It returns an error with a clear message if any field is invalid.
func (c *ResilienceConfig) Validate() error {
	if err := c.CircuitBreaker.validate(); err != nil {
		return fmt.Errorf("circuit breaker config: %w", err)
	}
	if err := c.Retry.validate(); err != nil {
		return fmt.Errorf("retry config: %w", err)
	}
	if err := c.Timeout.validate(); err != nil {
		return fmt.Errorf("timeout config: %w", err)
	}
	if err := c.Bulkhead.validate(); err != nil {
		return fmt.Errorf("bulkhead config: %w", err)
	}
	if err := c.Shutdown.validate(); err != nil {
		return fmt.Errorf("shutdown config: %w", err)
	}
	return nil
}

func (c *CircuitBreakerConfig) validate() error {
	if c.MaxRequests < 1 {
		return fmt.Errorf("max_requests must be greater than 0, got %d", c.MaxRequests)
	}
	if c.Interval <= 0 {
		return fmt.Errorf("interval must be greater than 0, got %s", c.Interval)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0, got %s", c.Timeout)
	}
	if c.FailureThreshold < 1 {
		return fmt.Errorf("failure_threshold must be greater than 0, got %d", c.FailureThreshold)
	}
	return nil
}

func (c *RetryConfig) validate() error {
	if c.MaxAttempts < 1 {
		return fmt.Errorf("max_attempts must be greater than 0, got %d", c.MaxAttempts)
	}
	if c.InitialDelay <= 0 {
		return fmt.Errorf("initial_delay must be greater than 0, got %s", c.InitialDelay)
	}
	if c.MaxDelay <= 0 {
		return fmt.Errorf("max_delay must be greater than 0, got %s", c.MaxDelay)
	}
	if c.MaxDelay < c.InitialDelay {
		return fmt.Errorf("max_delay must be greater than or equal to initial_delay, got max_delay=%s, initial_delay=%s", c.MaxDelay, c.InitialDelay)
	}
	if c.Multiplier < 1.0 {
		return fmt.Errorf("multiplier must be greater than or equal to 1.0, got %v", c.Multiplier)
	}
	return nil
}

func (c *TimeoutConfig) validate() error {
	if c.Default <= 0 {
		return fmt.Errorf("default timeout must be greater than 0, got %s", c.Default)
	}
	if c.Database <= 0 {
		return fmt.Errorf("database timeout must be greater than 0, got %s", c.Database)
	}
	if c.ExternalAPI <= 0 {
		return fmt.Errorf("external_api timeout must be greater than 0, got %s", c.ExternalAPI)
	}
	return nil
}

func (c *BulkheadConfig) validate() error {
	if c.MaxConcurrent < 1 {
		return fmt.Errorf("max_concurrent must be greater than 0, got %d", c.MaxConcurrent)
	}
	if c.MaxWaiting < 0 {
		return fmt.Errorf("max_waiting must be non-negative, got %d", c.MaxWaiting)
	}
	return nil
}

func (c *ShutdownConfig) validate() error {
	if c.DrainPeriod <= 0 {
		return fmt.Errorf("drain_period must be greater than 0, got %s", c.DrainPeriod)
	}
	if c.GracePeriod < 0 {
		return fmt.Errorf("grace_period must be non-negative, got %s", c.GracePeriod)
	}
	return nil
}
