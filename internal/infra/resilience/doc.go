// Package resilience provides centralized resilience configuration and error types.
//
// This package is part of the infrastructure layer and provides:
//   - ResilienceConfig: Centralized configuration for all resilience patterns
//   - Resilience error types: Structured errors with stable codes (RES-xxx)
//
// # Configuration
//
// The package supports configuration via environment variables:
//
//	# Circuit Breaker
//	CB_MAX_REQUESTS=3          # Requests allowed in half-open state
//	CB_INTERVAL=10s            # Cyclic period for clearing counts
//	CB_TIMEOUT=30s             # Time to wait before half-open
//	CB_FAILURE_THRESHOLD=5     # Failures to trip the breaker
//
//	# Retry
//	RETRY_MAX_ATTEMPTS=3       # Maximum retry attempts
//	RETRY_INITIAL_DELAY=100ms  # Initial backoff delay
//	RETRY_MAX_DELAY=5s         # Maximum backoff delay cap
//	RETRY_MULTIPLIER=2.0       # Exponential multiplier
//
//	# Timeout
//	TIMEOUT_DEFAULT=30s        # Default timeout for operations
//	TIMEOUT_DATABASE=5s        # Database operation timeout
//	TIMEOUT_EXTERNAL_API=10s   # External API call timeout
//
// # Error Codes
//
// | Code     | Name               | Description                              |
// |----------|--------------------| -----------------------------------------|
// | RES-001  | CircuitOpen        | Circuit breaker is open, requests rejected|
// | RES-002  | BulkheadFull       | Bulkhead capacity reached, request rejected|
// | RES-003  | TimeoutExceeded    | Operation timeout exceeded               |
// | RES-004  | MaxRetriesExceeded | Maximum retry attempts exhausted         |
//
// # Usage
//
// Configuration is loaded via the main config package and validated at startup:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err) // Fail fast on invalid config
//	}
//	resilienceCfg := resilience.NewResilienceConfig(cfg)
//	if err := resilienceCfg.Validate(); err != nil {
//	    log.Fatal(err)
//	}
package resilience
