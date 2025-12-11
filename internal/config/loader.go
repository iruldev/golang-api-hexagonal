package config

import (
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

// envPrefixes maps environment variable prefixes to config paths.
var envPrefixes = map[string]string{
	"APP_":  "app",
	"DB_":   "db",
	"OTEL_": "otel",
	"LOG_":  "log",
}

// Load loads configuration from environment variables.
// Environment variables are mapped using prefixes:
//   - APP_ -> app.* (e.g., APP_HTTP_PORT -> app.http_port)
//   - DB_ -> db.* (e.g., DB_HOST -> db.host)
//   - OTEL_ -> otel.* (e.g., OTEL_SERVICE_NAME -> otel.service_name)
//   - LOG_ -> log.* (e.g., LOG_LEVEL -> log.level)
func Load() (*Config, error) {
	k := koanf.New(".")

	// Load env vars for each prefix
	for prefix, path := range envPrefixes {
		if err := loadEnvPrefix(k, prefix, path); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadEnvPrefix loads environment variables with the given prefix.
func loadEnvPrefix(k *koanf.Koanf, prefix, path string) error {
	return k.Load(env.Provider(prefix, ".", func(s string) string {
		return path + "." + strings.ToLower(strings.TrimPrefix(s, prefix))
	}), nil)
}
