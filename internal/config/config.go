package config

import "time"

// Config holds all application configuration.
type Config struct {
	App           AppConfig           `koanf:"app"`
	Database      DatabaseConfig      `koanf:"db"`
	Observability ObservabilityConfig `koanf:"otel"`
	Log           LogConfig           `koanf:"log"`
}

// AppConfig holds application settings.
type AppConfig struct {
	Name     string `koanf:"name"`
	Env      string `koanf:"env"` // development, staging, production
	HTTPPort int    `koanf:"http_port"`
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
