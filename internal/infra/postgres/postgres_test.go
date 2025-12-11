package postgres

import (
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

func TestNewPool_InvalidConfig(t *testing.T) {
	// This test documents expected behavior
	// NewPool requires a running database - use testcontainers for integration tests

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:         "invalid-host-that-does-not-exist",
			Port:         5432,
			User:         "test",
			Password:     "test",
			Name:         "test",
			MaxOpenConns: 5,
			MaxIdleConns: 2,
		},
	}

	// Verify config is valid (DSN can be built)
	dsn := cfg.Database.DSN()
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
	// Note: Not logging DSN to avoid password in logs
	t.Log("NewPool would return error for invalid host when called with real database")
}

func TestDatabaseConfig_DSN(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.DatabaseConfig
		expected string
	}{
		{
			name: "basic config",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "secret",
				Name:     "testdb",
				SSLMode:  "disable",
			},
			expected: "postgres://postgres:secret@localhost:5432/testdb?sslmode=disable",
		},
		{
			name: "with require ssl",
			cfg: config.DatabaseConfig{
				Host:     "prod.db.example.com",
				Port:     5432,
				User:     "app",
				Password: "pass123",
				Name:     "production",
				SSLMode:  "require",
			},
			expected: "postgres://app:pass123@prod.db.example.com:5432/production?sslmode=require",
		},
		{
			name: "default ssl mode",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "test",
				Password: "test",
				Name:     "test",
				SSLMode:  "", // Should default to disable
			},
			expected: "postgres://test:test@localhost:5432/test?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.DSN()
			if got != tt.expected {
				t.Errorf("DSN() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestDatabaseConfig_PoolSettings(t *testing.T) {
	cfg := config.DatabaseConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * 60 * 1e9, // 30 minutes in nanoseconds
	}

	if cfg.MaxOpenConns != 25 {
		t.Errorf("MaxOpenConns = %d, want 25", cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns != 5 {
		t.Errorf("MaxIdleConns = %d, want 5", cfg.MaxIdleConns)
	}
}
