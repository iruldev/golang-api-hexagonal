package app

import (
	"net/http"
	"net/http/httptest"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulShutdown_CleanExit(t *testing.T) {
	// Arrange: Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create an http.Server from the test server
	httpServer := &http.Server{
		Addr:    server.Listener.Addr().String(),
		Handler: server.Config.Handler,
	}

	done := make(chan error, 1)

	// Act: Start shutdown in goroutine and send signal
	go func() {
		// Give GracefulShutdown time to set up signal handler
		time.Sleep(50 * time.Millisecond)
		// Trigger shutdown by sending SIGINT to self
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGINT); err != nil {
			t.Errorf("Failed to send signal: %v", err)
		}
	}()

	go GracefulShutdown(httpServer, done)

	// Assert: Shutdown should complete without error
	select {
	case err := <-done:
		assert.NoError(t, err, "Shutdown should complete without error")
	case <-time.After(5 * time.Second):
		t.Fatal("Shutdown timed out")
	}
}

func TestShutdownTimeout_IsCorrect(t *testing.T) {
	// Assert: ShutdownTimeout should be 30 seconds per NFR5
	require.Equal(t, 30*time.Second, ShutdownTimeout, "ShutdownTimeout should be 30 seconds per NFR5")
}
