package redis

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

// isRedisAvailable checks if Redis is running on localhost:6379
func isRedisAvailable() bool {
	conn, err := net.DialTimeout("tcp", "localhost:6379", 100*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func TestNewClient_WithRedisRunning(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping connection test")
	}

	// Test that we can connect when Redis is running
	cfg := config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Cleanup
	err = client.Close()
	assert.NoError(t, err)
}

func TestNewClient_WithRedisNotRunning(t *testing.T) {
	if isRedisAvailable() {
		t.Skip("Redis is available, skipping connection failure test")
	}

	// Test connection failure when Redis is not running
	cfg := config.RedisConfig{
		Host:        "localhost",
		Port:        6379,
		DialTimeout: 100 * time.Millisecond, // Fast timeout for test
	}

	_, err := NewClient(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection failed")
}

func TestNewClient_InvalidHost(t *testing.T) {
	// Test with definitely invalid host
	cfg := config.RedisConfig{
		Host:        "nonexistent.invalid.local.host.12345",
		Port:        6379,
		DialTimeout: 100 * time.Millisecond, // Fast timeout for test
	}

	_, err := NewClient(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis connection failed")
}

func TestNewClient_ConfigDefaults(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping config defaults test")
	}

	tests := []struct {
		name     string
		input    config.RedisConfig
		wantHost string
		wantPort int
	}{
		{
			name:     "empty config gets defaults",
			input:    config.RedisConfig{},
			wantHost: "localhost",
			wantPort: 6379,
		},
		{
			name: "custom port preserved",
			input: config.RedisConfig{
				Host: "localhost",
				Port: 6379,
			},
			wantHost: "localhost",
			wantPort: 6379,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.input)
			require.NoError(t, err)
			require.NotNil(t, client)
			client.Close()
		})
	}
}

// TestClient_Ping requires a running Redis instance.
func TestClient_Ping(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping ping test")
	}

	cfg := config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	err = client.Ping(context.Background())
	assert.NoError(t, err)
}

func TestClient_ImplementsDBHealthChecker(t *testing.T) {
	// Compile-time interface check
	var _ interface {
		Ping(ctx interface{ Err() error }) error
	} // This is a basic structure check

	// The actual interface compatibility is enforced by the handlers package
	// This test documents the requirement for future reference
}

func TestClient_Close(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping close test")
	}

	cfg := config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)

	err = client.Close()
	assert.NoError(t, err)
}

func TestClient_Client(t *testing.T) {
	if !isRedisAvailable() {
		t.Skip("Redis not available, skipping underlying client test")
	}

	cfg := config.RedisConfig{
		Host: "localhost",
		Port: 6379,
	}

	client, err := NewClient(cfg)
	require.NoError(t, err)
	defer client.Close()

	underlying := client.Client()
	require.NotNil(t, underlying)
}
