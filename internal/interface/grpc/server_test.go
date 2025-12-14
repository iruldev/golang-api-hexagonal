package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	tests := []struct {
		name           string
		cfg            *config.GRPCConfig
		wantReflection bool
	}{
		{
			name: "reflection enabled",
			cfg: &config.GRPCConfig{
				Enabled:           true,
				Port:              0, // Random port for testing
				ReflectionEnabled: true,
			},
			wantReflection: true,
		},
		{
			name: "reflection disabled",
			cfg: &config.GRPCConfig{
				Enabled:           true,
				Port:              0,
				ReflectionEnabled: false,
			},
			wantReflection: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := observability.NewNopLoggerInterface()

			server := NewServer(tt.cfg, logger)

			assert.NotNil(t, server)
			assert.NotNil(t, server.GRPCServer())
			assert.Equal(t, tt.cfg, server.cfg)
		})
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	cfg := &config.GRPCConfig{
		Enabled:           true,
		Port:              0, // Use random available port
		ReflectionEnabled: true,
	}
	logger := observability.NewNopLoggerInterface()

	server := NewServer(cfg, logger)
	require.NotNil(t, server)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start(context.Background())
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Gracefully shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	assert.NoError(t, err)

	// Check that Start returned without error after shutdown
	select {
	case startErr := <-errChan:
		// Start should return nil after GracefulStop
		assert.NoError(t, startErr)
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestServer_Stop(t *testing.T) {
	cfg := &config.GRPCConfig{
		Enabled:           true,
		Port:              0,
		ReflectionEnabled: false,
	}
	logger := observability.NewNopLoggerInterface()

	server := NewServer(cfg, logger)
	require.NotNil(t, server)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start(context.Background())
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Immediately stop
	server.Stop()

	// Check that Start returned
	select {
	case <-errChan:
		// Server stopped
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop within timeout")
	}
}

func TestServer_WithUnaryInterceptors(t *testing.T) {
	cfg := &config.GRPCConfig{
		Enabled: true,
		Port:    0,
	}
	logger := observability.NewNopLoggerInterface()

	interceptorCalled := false
	testInterceptor := func(ctx context.Context, req interface{}, info interface{}, handler interface{}) (interface{}, error) {
		interceptorCalled = true
		return nil, nil
	}
	_ = testInterceptor
	_ = interceptorCalled

	// Test that WithUnaryInterceptors doesn't panic
	server := NewServer(cfg, logger, WithUnaryInterceptors())

	assert.NotNil(t, server)
	assert.NotNil(t, server.GRPCServer())
}

func TestServer_WithStreamInterceptors(t *testing.T) {
	cfg := &config.GRPCConfig{
		Enabled: true,
		Port:    0,
	}
	logger := observability.NewNopLoggerInterface()

	// Test that WithStreamInterceptors doesn't panic
	server := NewServer(cfg, logger, WithStreamInterceptors())

	assert.NotNil(t, server)
	assert.NotNil(t, server.GRPCServer())
}
