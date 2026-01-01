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

	// Database Pool Configuration (Story 5.1)
	// DBPoolMaxConns is the maximum number of connections in the pool. Default: 25.
	DBPoolMaxConns int32 `envconfig:"DB_POOL_MAX_CONNS" default:"25"`
	// DBPoolMinConns is the minimum number of connections in the pool. Default: 5.
	DBPoolMinConns int32 `envconfig:"DB_POOL_MIN_CONNS" default:"5"`
	// DBPoolMaxLifetime is the maximum lifetime of a connection. Default: 1h.
	DBPoolMaxLifetime time.Duration `envconfig:"DB_POOL_MAX_LIFETIME" default:"1h"`

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

	// Internal Server (Story 2.5a)
	// InternalPort is the port for internal endpoints like /metrics. Default: 8081.
	InternalPort int `envconfig:"INTERNAL_PORT" default:"8081"`
	// InternalBindAddress is the bind address for the internal server.
	// Default: "127.0.0.1" (loopback only) for security isolation.
	InternalBindAddress string `envconfig:"INTERNAL_BIND_ADDRESS" default:"127.0.0.1"`

	// Smoke Test Support (Hidden)
	// IgnoreDBStartupError allows starting the server without a valid DB connection.
	// Intended ONLY for smoke testing/build verification. Default: false.
	IgnoreDBStartupError bool `envconfig:"IGNORE_DB_STARTUP_ERROR" default:"false"`

	// Server Timeouts
	// HTTPReadTimeout is the maximum duration for reading the entire request, including the body. Default: 15s.
	HTTPReadTimeout time.Duration `envconfig:"HTTP_READ_TIMEOUT" default:"15s"`
	// HTTPWriteTimeout is the maximum duration before timing out writes of the response. Default: 15s.
	HTTPWriteTimeout time.Duration `envconfig:"HTTP_WRITE_TIMEOUT" default:"15s"`
	// HTTPIdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled. Default: 60s.
	HTTPIdleTimeout time.Duration `envconfig:"HTTP_IDLE_TIMEOUT" default:"60s"`
	// ShutdownTimeout is the duration to wait for graceful shutdown. Default: 30s.
	ShutdownTimeout time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
	// DBQueryTimeout is the default timeout for database queries. Default: 5s.
	DBQueryTimeout time.Duration `envconfig:"DB_QUERY_TIMEOUT" default:"5s"`
	// HTTPReadHeaderTimeout is the amount of time allowed to read request headers.
	// Default: 10s. This helps mitigate slowloris attacks.
	HTTPReadHeaderTimeout time.Duration `envconfig:"HTTP_READ_HEADER_TIMEOUT" default:"10s"`
	// HTTPMaxHeaderBytes is the maximum size of request headers.
	// Default: 1MB (1048576 bytes). This helps prevent header-based DoS attacks.
	HTTPMaxHeaderBytes int `envconfig:"HTTP_MAX_HEADER_BYTES" default:"1048576"`

	// Audit
	// AuditRedactEmail controls how email addresses are redacted in audit logs.
	// Options: "full" (default, replaces with [REDACTED]) or "partial" (shows first 2 chars + domain).
	AuditRedactEmail string `envconfig:"AUDIT_REDACT_EMAIL" default:"full"`

	// Resilience - Circuit Breaker
	// CBMaxRequests is the number of requests allowed in the half-open state. Default: 3.
	CBMaxRequests int `envconfig:"CB_MAX_REQUESTS" default:"3"`
	// CBInterval is the cyclic period for clearing internal counts. Default: 10s.
	CBInterval time.Duration `envconfig:"CB_INTERVAL" default:"10s"`
	// CBTimeout is the period to wait before transitioning from open to half-open. Default: 30s.
	CBTimeout time.Duration `envconfig:"CB_TIMEOUT" default:"30s"`
	// CBFailureThreshold is the number of failures to trip the circuit. Default: 5.
	CBFailureThreshold int `envconfig:"CB_FAILURE_THRESHOLD" default:"5"`

	// Resilience - Retry
	// RetryMaxAttempts is the maximum number of retry attempts. Default: 3.
	RetryMaxAttempts int `envconfig:"RETRY_MAX_ATTEMPTS" default:"3"`
	// RetryInitialDelay is the initial delay before the first retry. Default: 100ms.
	RetryInitialDelay time.Duration `envconfig:"RETRY_INITIAL_DELAY" default:"100ms"`
	// RetryMaxDelay is the maximum delay between retries. Default: 5s.
	RetryMaxDelay time.Duration `envconfig:"RETRY_MAX_DELAY" default:"5s"`
	// RetryMultiplier is the factor by which the delay increases after each retry. Default: 2.0.
	RetryMultiplier float64 `envconfig:"RETRY_MULTIPLIER" default:"2.0"`

	// Resilience - Timeout
	// TimeoutDefault is the default timeout for operations. Default: 30s.
	TimeoutDefault time.Duration `envconfig:"TIMEOUT_DEFAULT" default:"30s"`
	// TimeoutDatabase is the timeout for database operations. Default: 5s.
	TimeoutDatabase time.Duration `envconfig:"TIMEOUT_DATABASE" default:"5s"`
	// TimeoutExternalAPI is the timeout for external API calls. Default: 10s.
	TimeoutExternalAPI time.Duration `envconfig:"TIMEOUT_EXTERNAL_API" default:"10s"`

	// Resilience - Bulkhead (Story 1.5)
	// BulkheadMaxConcurrent is the maximum number of concurrent executions. Default: 10.
	BulkheadMaxConcurrent int `envconfig:"BULKHEAD_MAX_CONCURRENT" default:"10"`
	// BulkheadMaxWaiting is the maximum number of operations waiting for execution. Default: 100.
	BulkheadMaxWaiting int `envconfig:"BULKHEAD_MAX_WAITING" default:"100"`

	// Resilience - Graceful Shutdown (Story 1.6)
	// ShutdownDrainPeriod is the maximum time to wait for in-flight requests to complete. Default: 30s.
	ShutdownDrainPeriod time.Duration `envconfig:"SHUTDOWN_DRAIN_PERIOD" default:"30s"`
	// ShutdownGracePeriod is additional time after drain for cleanup operations. Default: 5s.
	ShutdownGracePeriod time.Duration `envconfig:"SHUTDOWN_GRACE_PERIOD" default:"5s"`

	// Idempotency (Story 2.5)
	// IdempotencyTTL is the time-to-live for idempotency records. Default: 24h.
	IdempotencyTTL time.Duration `envconfig:"IDEMPOTENCY_TTL" default:"24h"`
	// IdempotencyCleanupInterval is the interval between cleanup job runs. Default: 1h.
	IdempotencyCleanupInterval time.Duration `envconfig:"IDEMPOTENCY_CLEANUP_INTERVAL" default:"1h"`

	// Health Check (Story 3.4)
	// HealthCheckDBTimeout is the timeout for database health check. Default: 2s.
	HealthCheckDBTimeout time.Duration `envconfig:"HEALTH_CHECK_DB_TIMEOUT" default:"2s"`
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

	// Allow 0 for dynamic port allocation
	if c.Port < 0 || c.Port > 65535 {
		return fmt.Errorf("invalid PORT: must be between 0 and 65535")
	}
	// Story 2.5a: InternalPort validation
	// Allow 0 for dynamic port allocation
	if c.InternalPort < 0 || c.InternalPort > 65535 {
		return fmt.Errorf("invalid INTERNAL_PORT: must be between 0 and 65535")
	}
	// Only check collision if both are non-zero (if 0, OS assigns different ports)
	if c.InternalPort != 0 && c.InternalPort == c.Port {
		return fmt.Errorf("INTERNAL_PORT must differ from PORT")
	}
	// Check InternalBindAddress is valid IP
	if c.InternalBindAddress == "" {
		return fmt.Errorf("INTERNAL_BIND_ADDRESS cannot be empty")
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

	// Story 5.1: Database Pool Validation
	if c.DBPoolMaxConns < 1 {
		return fmt.Errorf("invalid DB_POOL_MAX_CONNS: must be greater than 0")
	}
	if c.DBPoolMinConns < 0 { // 0 is technically allowed by pgx (no idle conns), but let's allow it. config default is 5.
		return fmt.Errorf("invalid DB_POOL_MIN_CONNS: must be non-negative")
	}
	if c.DBPoolMinConns > c.DBPoolMaxConns {
		return fmt.Errorf("invalid DB_POOL_MIN_CONNS: must be less than or equal to DB_POOL_MAX_CONNS")
	}
	if c.DBPoolMaxLifetime <= 0 {
		return fmt.Errorf("invalid DB_POOL_MAX_LIFETIME: must be greater than 0")
	}

	// Server Timeouts Validation
	if c.DBQueryTimeout <= 0 {
		return fmt.Errorf("invalid DB_QUERY_TIMEOUT: must be greater than 0")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("invalid SHUTDOWN_TIMEOUT: must be greater than 0")
	}

	// Story 1.6: Graceful Shutdown validation
	if c.ShutdownDrainPeriod <= 0 {
		return fmt.Errorf("invalid SHUTDOWN_DRAIN_PERIOD: must be greater than 0")
	}
	if c.ShutdownGracePeriod < 0 {
		return fmt.Errorf("invalid SHUTDOWN_GRACE_PERIOD: must be non-negative")
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
