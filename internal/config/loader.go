package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// envPrefixes maps environment variable prefixes to config paths.
var envPrefixes = map[string]string{
	"APP_":  "app",
	"DB_":   "db",
	"OTEL_": "otel",
	"LOG_":  "log",
}

// Load loads configuration from optional file and environment variables.
// If APP_CONFIG_FILE is set, loads from that file first.
// Environment variables always override file values.
func Load() (*Config, error) {
	k := koanf.New(".")

	// Step 1: Load from config file if specified
	if configFile := os.Getenv("APP_CONFIG_FILE"); configFile != "" {
		if err := loadFromFile(k, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file %s: %w", configFile, err)
		}
	}

	// Step 2: Load env vars (overrides file values)
	for prefix, path := range envPrefixes {
		if err := loadEnvPrefix(k, prefix, path); err != nil {
			return nil, err
		}
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, err
	}

	// Validate config - fail fast on invalid configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// loadFromFile loads configuration from a YAML or JSON file.
func loadFromFile(k *koanf.Koanf, path string) error {
	// Check file exists first for clearer error message
	if _, err := os.Stat(path); err != nil {
		return err
	}

	ext := filepath.Ext(path)
	var parser koanf.Parser

	switch ext {
	case ".yaml", ".yml":
		parser = yaml.Parser()
	case ".json":
		parser = json.Parser()
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	return k.Load(file.Provider(path), parser)
}

// loadEnvPrefix loads environment variables with the given prefix.
func loadEnvPrefix(k *koanf.Koanf, prefix, path string) error {
	return k.Load(env.Provider(prefix, ".", func(s string) string {
		return path + "." + strings.ToLower(strings.TrimPrefix(s, prefix))
	}), nil)
}
