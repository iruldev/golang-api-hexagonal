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
	Kafka         KafkaConfig         `koanf:"kafka"`
	RabbitMQ      RabbitMQConfig      `koanf:"rabbitmq"`
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

// KafkaConfig holds Kafka producer settings.
type KafkaConfig struct {
	Enabled      bool          `koanf:"enabled"`       // KAFKA_ENABLED, default: false
	Brokers      []string      `koanf:"brokers"`       // KAFKA_BROKERS, default: localhost:9092
	ClientID     string        `koanf:"client_id"`     // KAFKA_CLIENT_ID, default: golang-api-hexagonal
	Timeout      time.Duration `koanf:"timeout"`       // KAFKA_PRODUCER_TIMEOUT, default: 10s
	RequiredAcks string        `koanf:"required_acks"` // KAFKA_PRODUCER_REQUIRED_ACKS: all, local, none
	// TLS/SASL config placeholders for production
	TLSEnabled    bool   `koanf:"tls_enabled"`    // KAFKA_TLS_ENABLED, default: false
	SASLEnabled   bool   `koanf:"sasl_enabled"`   // KAFKA_SASL_ENABLED, default: false
	SASLUsername  string `koanf:"sasl_username"`  // KAFKA_SASL_USERNAME
	SASLPassword  string `koanf:"sasl_password"`  // KAFKA_SASL_PASSWORD
	SASLMechanism string `koanf:"sasl_mechanism"` // KAFKA_SASL_MECHANISM: PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
}

// IsEnabled returns true if Kafka is enabled.
func (c *KafkaConfig) IsEnabled() bool {
	return c.Enabled
}

// RabbitMQConfig holds RabbitMQ connection settings.
type RabbitMQConfig struct {
	Enabled       bool          `koanf:"enabled"`        // RABBITMQ_ENABLED, default: false
	URL           string        `koanf:"url"`            // RABBITMQ_URL, default: amqp://guest:guest@localhost:5672/
	Exchange      string        `koanf:"exchange"`       // RABBITMQ_EXCHANGE, default: events
	ExchangeType  string        `koanf:"exchange_type"`  // RABBITMQ_EXCHANGE_TYPE: direct, topic, fanout, headers
	Durable       bool          `koanf:"durable"`        // RABBITMQ_DURABLE, default: true
	PrefetchCount int           `koanf:"prefetch_count"` // RABBITMQ_PREFETCH_COUNT, default: 10 (for future consumer)
	Timeout       time.Duration `koanf:"timeout"`        // RABBITMQ_TIMEOUT, default: 10s
	// TLS config placeholder for production
	TLSEnabled bool `koanf:"tls_enabled"` // RABBITMQ_TLS_ENABLED, default: false
}

// IsEnabled returns true if RabbitMQ is enabled.
func (c *RabbitMQConfig) IsEnabled() bool {
	return c.Enabled
}
