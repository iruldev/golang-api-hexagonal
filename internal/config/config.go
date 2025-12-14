package config

import (
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	App           AppConfig           `koanf:"app"`
	Database      DatabaseConfig      `koanf:"db"`
	Redis         RedisConfig         `koanf:"redis"`
	Asynq         AsynqConfig         `koanf:"asynq"`
	GRPC          GRPCConfig          `koanf:"grpc"`
	Observability ObservabilityConfig `koanf:"otel"`
	Log           LogConfig           `koanf:"log"`
}

// GRPCConfig holds gRPC server settings.
type GRPCConfig struct {
	Enabled           bool `koanf:"enabled"`            // default: false
	Port              int  `koanf:"port"`               // default: 50051
	ReflectionEnabled bool `koanf:"reflection_enabled"` // default: true in dev, false in prod
}

// Environment constants for application environment.
const (
	EnvDevelopment = "development"
	EnvLocal       = "local"
	EnvStaging     = "staging"
	EnvProduction  = "production"
)

// AppConfig holds application settings.
type AppConfig struct {
	Name     string `koanf:"name"`
	Env      string `koanf:"env"` // development, local, staging, production
	HTTPPort int    `koanf:"http_port"`
}

// IsDevelopment returns true if the environment is development or local.
// Use this for enabling dev-only features like GraphQL Playground.
func (c *AppConfig) IsDevelopment() bool {
	return c.Env == EnvDevelopment || c.Env == EnvLocal
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string        `koanf:"host"`
	Port            int           `koanf:"port"`
	User            string        `koanf:"user"`
	Password        string        `koanf:"password"`
	Name            string        `koanf:"name"`
	SSLMode         string        `koanf:"ssl_mode"`
	MaxOpenConns    int           `koanf:"max_open_conns"`
	MaxIdleConns    int           `koanf:"max_idle_conns"`
	ConnMaxLifetime time.Duration `koanf:"conn_max_lifetime"`
	ConnTimeout     time.Duration `koanf:"conn_timeout"`  // connection timeout (default: 10s)
	QueryTimeout    time.Duration `koanf:"query_timeout"` // query timeout (default: 30s)
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" +
		strconv.Itoa(c.Port) + "/" + c.Name + "?sslmode=" + sslMode
}

// ObservabilityConfig holds OpenTelemetry settings.
type ObservabilityConfig struct {
	ExporterEndpoint string `koanf:"exporter_otlp_endpoint"`
	ServiceName      string `koanf:"service_name"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"` // json, console
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host         string        `koanf:"host"`
	Port         int           `koanf:"port"`
	Password     string        `koanf:"password"`
	DB           int           `koanf:"db"`
	PoolSize     int           `koanf:"pool_size"`
	MinIdleConns int           `koanf:"min_idle_conns"`
	DialTimeout  time.Duration `koanf:"dial_timeout"`
	ReadTimeout  time.Duration `koanf:"read_timeout"`
	WriteTimeout time.Duration `koanf:"write_timeout"`
}

// AsynqConfig holds Asynq worker settings.
type AsynqConfig struct {
	Concurrency     int           `koanf:"concurrency"`      // default: 10
	RetryMax        int           `koanf:"retry_max"`        // default: 3
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"` // default: 30s
}
