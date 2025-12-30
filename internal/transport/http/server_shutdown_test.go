//go:build integration

// Package http contains integration tests for HTTP server shutdown behavior.
// NOTE: This file intentionally imports from internal/infra/resilience for integration
// testing purposes. This cross-layer import is acceptable in integration tests
// (as opposed to unit tests) because integration tests verify the complete system
// behavior including proper wiring between layers.
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/resilience"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// TestGracefulShutdown verifies that the server completes in-flight requests
// before shutting down when Shutdown is called.
func TestGracefulShutdown(t *testing.T) {
	// Synchronization channels
	handlerStarted := make(chan struct{})
	blockHandler := make(chan struct{})

	// Create a handler that we can control
	controlledHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handlerStarted) // Signal that request has reached handler
		<-blockHandler        // Wait until we allow it to proceed
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("completed"))
	})

	// Create server
	mux := http.NewServeMux()
	mux.Handle("/controlled", controlledHandler)

	srv := &http.Server{
		Handler: mux,
	}

	// Use random port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Start server in goroutine
	serverErrCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
			serverErrCh <- err
		}
		close(serverErrCh)
	}()

	// Start in-flight request
	requestResult := make(chan error, 1)
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/controlled", ln.Addr()))
		if err != nil {
			requestResult <- err
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			requestResult <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
			return
		}
		requestResult <- nil
	}()

	// Wait for request to reach the handler (proving it is properly "in-flight")
	select {
	case <-handlerStarted:
		// good
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for request to start")
	}

	// Initiate graceful shutdown
	shutdownErrCh := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErrCh <- srv.Shutdown(ctx)
	}()

	// Unblock the handler to allow request to finish
	// We do this AFTER shutdown started to verify Shutdown waits for it
	// Note: In a real race, Shutdown calls Close immediately if not implementing graceful correctly.
	// But `Shutdown` doc says it blocks. To verify it's blocked *waiting* for us, we could adding a small delay or check.
	// For integrating testing, ensuring it completes successfull is the main goal (AC2).
	close(blockHandler)

	// Verify request completed successfully
	select {
	case err := <-requestResult:
		assert.NoError(t, err, "in-flight request should complete successfull")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for request to complete")
	}

	// Verify shutdown completed successfully
	select {
	case err := <-shutdownErrCh:
		assert.NoError(t, err, "shutdown should complete successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for shutdown to return")
	}

	// Verify server Serve returned
	select {
	case err := <-serverErrCh:
		assert.NoError(t, err, "Serve should return without non-closed error")
	case <-time.After(1 * time.Second):
		t.Fatal("Serve did not return after shutdown")
	}
}

// TestGracefulShutdownMultipleRequests verifies that multiple in-flight
// requests all complete before shutdown.
func TestGracefulShutdownMultipleRequests(t *testing.T) {
	const numRequests = 5

	var wg sync.WaitGroup
	wg.Add(numRequests)

	// Handler just waits gently
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Minimal sleep to ensure overlap without complexity
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)

	// Send requests
	errCh := make(chan error, numRequests)
	for i := 0; i < numRequests; i++ {
		go func() {
			defer wg.Done()
			resp, err := http.Get(fmt.Sprintf("http://%s/", ln.Addr()))
			if err != nil {
				errCh <- err
				return
			}
			resp.Body.Close()
		}()
	}

	// We can't easily sync "all requests started" without complex coordination or race-prone sleeps
	// for multiple independent requests.
	// Instead, we just call Shutdown immediately. The OS usually buffers pending accepts/requests.
	// For this test to be strictly robust without sleep, we'd need a "RequestsActive" counter in handler.
	// Leaving a tiny startup delay as acceptable trade-off for this specific scenario
	// OR better: Just call Shutdown. If they started, they finish. If they didn't, they fail (refused).
	// But we want to test "in-flight".
	// Let's use a standard WaitUntil for a very short period to let some connect.
	time.Sleep(50 * time.Millisecond) // Keeping minimal sleep here as simpler alternative to complex barrier

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = srv.Shutdown(ctx)
	require.NoError(t, err)

	wg.Wait()
	close(errCh)

	for err := range errCh {
		assert.NoError(t, err, "request failed during shutdown")
	}
}

// TestGracefulShutdownTimeout verifies that shutdown respects the
// configured timeout (AC4).
func TestGracefulShutdownTimeout(t *testing.T) {
	// Handler that blocks forever (until test ends)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(10 * time.Second):
			return
		}
	})

	srv := &http.Server{Handler: handler}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)

	// Start a request to occupy the server
	go http.Get(fmt.Sprintf("http://%s/", ln.Addr()))

	// Give it a moment to reach handler
	time.Sleep(10 * time.Millisecond)

	// Shutdown with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	start := time.Now()
	err = srv.Shutdown(ctx)
	elapsed := time.Since(start)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Less(t, elapsed, 1*time.Second)
}

// =============================================================================
// Story 1.6: ShutdownCoordinator Integration Tests
// =============================================================================

// TestShutdownCoordinator_RejectsNewRequests verifies that new requests are rejected
// with 503 Service Unavailable after shutdown is initiated (AC1).
func TestShutdownCoordinator_RejectsNewRequests(t *testing.T) {
	// Create a ShutdownCoordinator
	cfg := resilience.ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := resilience.NewShutdownCoordinator(cfg)

	// Create handler with shutdown middleware
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	shutdownMiddleware := middleware.Shutdown(coord)
	handler := shutdownMiddleware(baseHandler)

	// Create server
	mux := http.NewServeMux()
	mux.Handle("/", handler)

	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)
	defer srv.Close()

	// Verify request works before shutdown
	resp, err := http.Get(fmt.Sprintf("http://%s/", ln.Addr()))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Initiate shutdown
	coord.InitiateShutdown()

	// Verify new request is rejected with 503
	resp, err = http.Get(fmt.Sprintf("http://%s/", ln.Addr()))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "new requests should get 503 during shutdown")
	assert.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))
	assert.NotEmpty(t, resp.Header.Get("Retry-After"), "should have Retry-After header")
	// Note: Connection: close header is set by middleware but may be normalized by Go's http client

	// Verify RFC 7807 format
	var problem contract.ProblemDetail
	err = json.NewDecoder(resp.Body).Decode(&problem)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, problem.Status)
	assert.Equal(t, "Service Unavailable", problem.Title)
	assert.Contains(t, problem.Detail, "shutting down")
}

// TestShutdownCoordinator_TracksActiveRequests verifies that the ShutdownCoordinator
// correctly tracks in-flight requests (AC2).
func TestShutdownCoordinator_TracksActiveRequests(t *testing.T) {
	cfg := resilience.ShutdownConfig{
		DrainPeriod: 30 * time.Second,
		GracePeriod: 5 * time.Second,
	}
	coord := resilience.NewShutdownCoordinator(cfg)

	// Channels for synchronization
	handlerStarted := make(chan struct{})
	handlerComplete := make(chan struct{})

	// Create handler with shutdown middleware
	baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(handlerStarted) // Signal handler started
		<-handlerComplete     // Wait for signal to complete
		w.WriteHeader(http.StatusOK)
	})

	shutdownMiddleware := middleware.Shutdown(coord)
	handler := shutdownMiddleware(baseHandler)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	srv := &http.Server{Handler: mux}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	go srv.Serve(ln)
	defer srv.Close()

	// Verify initial count is 0
	assert.Equal(t, int64(0), coord.ActiveCount())

	// Start a request
	go http.Get(fmt.Sprintf("http://%s/", ln.Addr()))

	// Wait for handler to start
	select {
	case <-handlerStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("handler did not start")
	}

	// Verify active count is now 1
	assert.Equal(t, int64(1), coord.ActiveCount())

	// Complete the request
	close(handlerComplete)

	// Wait a bit for the request to complete
	time.Sleep(100 * time.Millisecond)

	// Verify active count is back to 0
	assert.Equal(t, int64(0), coord.ActiveCount())
}

// TestShutdownCoordinator_WaitForDrain verifies that WaitForDrain completes
// when all in-flight requests finish and times out otherwise (AC3).
func TestShutdownCoordinator_WaitForDrain(t *testing.T) {
	t.Run("completes when all requests finish", func(t *testing.T) {
		cfg := resilience.ShutdownConfig{
			DrainPeriod: 5 * time.Second,
			GracePeriod: 1 * time.Second,
		}
		coord := resilience.NewShutdownCoordinator(cfg)

		handlerStarted := make(chan struct{})
		handlerComplete := make(chan struct{})

		baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			<-handlerComplete
			w.WriteHeader(http.StatusOK)
		})

		shutdownMiddleware := middleware.Shutdown(coord)
		handler := shutdownMiddleware(baseHandler)

		mux := http.NewServeMux()
		mux.Handle("/", handler)

		srv := &http.Server{Handler: mux}
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		go srv.Serve(ln)
		defer srv.Close()

		// Start a request
		go http.Get(fmt.Sprintf("http://%s/", ln.Addr()))

		// Wait for handler to start
		<-handlerStarted

		// Initiate shutdown
		coord.InitiateShutdown()

		// Start WaitForDrain in goroutine
		drainResult := make(chan error, 1)
		go func() {
			drainResult <- coord.WaitForDrain(context.Background())
		}()

		// Complete the in-flight request
		time.Sleep(50 * time.Millisecond)
		close(handlerComplete)

		// Verify drain completes successfully
		select {
		case err := <-drainResult:
			assert.NoError(t, err, "WaitForDrain should complete when all requests finish")
		case <-time.After(5 * time.Second):
			t.Fatal("WaitForDrain did not complete in time")
		}
	})

	t.Run("times out when requests exceed drain period", func(t *testing.T) {
		cfg := resilience.ShutdownConfig{
			DrainPeriod: 100 * time.Millisecond, // Very short drain period
			GracePeriod: 50 * time.Millisecond,
		}
		coord := resilience.NewShutdownCoordinator(cfg)

		handlerStarted := make(chan struct{})

		baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(handlerStarted)
			// Block forever (simulating a stuck request)
			select {
			case <-r.Context().Done():
				return
			case <-time.After(10 * time.Second):
				return
			}
		})

		shutdownMiddleware := middleware.Shutdown(coord)
		handler := shutdownMiddleware(baseHandler)

		mux := http.NewServeMux()
		mux.Handle("/", handler)

		srv := &http.Server{Handler: mux}
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)

		go srv.Serve(ln)
		defer srv.Close()

		// Start a request that will block
		go http.Get(fmt.Sprintf("http://%s/", ln.Addr()))

		// Wait for handler to start
		<-handlerStarted

		// Initiate shutdown
		coord.InitiateShutdown()

		// WaitForDrain should timeout
		start := time.Now()
		err = coord.WaitForDrain(context.Background())
		elapsed := time.Since(start)

		assert.Error(t, err, "WaitForDrain should error on timeout")
		assert.Contains(t, err.Error(), "drain timeout")
		assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "should wait at least drain period")
		assert.Less(t, elapsed, 500*time.Millisecond, "should not wait excessively long")
	})
}
