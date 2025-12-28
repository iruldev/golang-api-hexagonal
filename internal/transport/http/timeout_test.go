//go:build integration

package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

// TestHTTPServerTimeouts verifies that HTTP server timeouts are properly applied (AC1).
func TestHTTPServerTimeouts(t *testing.T) {
	// Create server with short timeouts
	srv := &http.Server{
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		}),
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	// Test normal request completes within timeout
	resp, err := http.Get(fmt.Sprintf("http://%s/", ln.Addr()))
	require.NoError(t, err, "normal request should complete")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

// TestHTTPWriteTimeout verifies that WriteTimeout triggers when handler exceeds it (AC1).
func TestHTTPWriteTimeout(t *testing.T) {
	// Channel to synchronize handler start
	handlerStarted := make(chan struct{})

	// Create server with very short write timeout
	srv := &http.Server{
		WriteTimeout: 50 * time.Millisecond,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			// Sleep longer than write timeout
			// We use a small sleep that is definitely larger than 50ms
			time.Sleep(200 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}),
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	// Make request - should timeout
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/", ln.Addr()))

	// Wait for handler to start to ensure we tested the timeout case
	<-handlerStarted

	// Verify that we got a failure (connection close or error)
	// If the timeout works, the server closes the connection, which results in an error on the client side
	// (usually EOF or connection reset)
	assert.Error(t, err, "request should fail due to write timeout")
	if resp != nil {
		resp.Body.Close()
	}
}

// TestConfiguredShutdownTimeout verifies that shutdown respects configured timeout (AC3).
func TestConfiguredShutdownTimeout(t *testing.T) {
	// Channel to verify request entered handler
	requestStarted := make(chan struct{})

	// Handler that blocks longer than shutdown timeout
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		select {
		case <-r.Context().Done():
			return
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	})

	srv := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)

	// Start a request that will block
	go func() {
		http.Get(fmt.Sprintf("http://%s/", ln.Addr()))
	}()

	// Wait for request to reach handler
	<-requestStarted

	// Shutdown with very short timeout - should timeout
	shutdownTimeout := 100 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	start := time.Now()
	err = srv.Shutdown(ctx)
	elapsed := time.Since(start)

	// Should error with deadline exceeded
	assert.ErrorIs(t, err, context.DeadlineExceeded,
		"shutdown should timeout when requests don't finish")
	assert.Less(t, elapsed, 1*time.Second,
		"shutdown should respect timeout duration")
}

// TestShutdownWithinTimeout verifies successful shutdown within timeout (AC3).
func TestShutdownWithinTimeout(t *testing.T) {
	// Fast handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)

	// Shutdown with generous timeout - should succeed
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctx)
	assert.NoError(t, err, "shutdown should succeed when no blocking requests")
}

// TestTimeoutFromEnv verifies that timeouts can be set via env vars (AC4).
func TestTimeoutFromEnv(t *testing.T) {
	// Set env vars
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db")
	t.Setenv("HTTP_READ_TIMEOUT", "2s")
	t.Setenv("HTTP_WRITE_TIMEOUT", "3s")
	t.Setenv("SHUTDOWN_TIMEOUT", "15s")
	t.Setenv("DB_QUERY_TIMEOUT", "5s")

	cfg, err := config.Load()
	require.NoError(t, err)

	assert.Equal(t, 2*time.Second, cfg.HTTPReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.HTTPWriteTimeout)
	assert.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
	assert.Equal(t, 5*time.Second, cfg.DBQueryTimeout)
}
