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
	"go.uber.org/goleak"
)

// TestContextCancellation verifies that HTTP requests are properly cancelled
// when the request context is cancelled (AC1, AC2).
func TestContextCancellation(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Synchronization channels
	handlerStarted := make(chan struct{})
	handlerDone := make(chan struct{})

	// Create a handler that respects context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handlerStarted)
		ctx := r.Context()
		select {
		case <-ctx.Done():
			// Context cancelled - clean up and return
			close(handlerDone)
			return
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	})

	srv := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	// Create request with cancellable context (AC1)
	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("http://%s/", ln.Addr()), nil)
	require.NoError(t, err)

	// Start request in goroutine
	errCh := make(chan error, 1)
	go func() {
		_, err := http.DefaultClient.Do(req)
		errCh <- err
	}()

	// Wait for handler to start
	select {
	case <-handlerStarted:
		// good
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for handler to start")
	}

	// Cancel context mid-request (AC2)
	cancel()

	// Verify request was cancelled
	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled, "request should be cancelled")
	case <-time.After(2 * time.Second):
		t.Fatal("request did not cancel in time")
	}

	// Wait for handler to finish cleanup
	select {
	case <-handlerDone:
		// good
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not finish cleanup")
	}
}

// TestContextCancellationNoGoroutineLeaks verifies that no goroutines are leaked
// when context is cancelled (AC4).
func TestContextCancellationNoGoroutineLeaks(t *testing.T) {
	defer goleak.VerifyNone(t)

	// Channel to signal that handler has received the request
	ready := make(chan struct{})

	// Create a handler that blocks until context is cancelled
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Signal that we've reached the handler
		select {
		case ready <- struct{}{}:
		case <-r.Context().Done():
			// Request might be cancelled before we send ready
		}
		<-r.Context().Done()
	})

	srv := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	// Run multiple cancel cycles to stress test
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		req, err := http.NewRequestWithContext(ctx, "GET",
			fmt.Sprintf("http://%s/", ln.Addr()), nil)
		require.NoError(t, err)

		errCh := make(chan error, 1)
		go func() {
			_, err := http.DefaultClient.Do(req)
			errCh <- err
		}()

		// Wait for handler to confirm it's running
		select {
		case <-ready:
			// Handler is running, now we can cancel
			cancel()
		case <-time.After(2 * time.Second):
			cancel() // cleanup
			t.Fatalf("iteration %d: timed out waiting for handler to start", i)
		}

		// Wait for request to return
		select {
		case <-errCh:
			// good
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: request did not cancel", i)
		}
	}
}
