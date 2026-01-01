package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Story 3.3: Startup Probe Unit Tests (AC: #1, #2, #3)
// =============================================================================

// TestStartupHandler_BeforeReady verifies 503 is returned before MarkReady (AC1).
func TestStartupHandler_BeforeReady(t *testing.T) {
	handler := NewStartupHandler()

	req := httptest.NewRequest(http.MethodGet, "/startupz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// AC1: 503 Service Unavailable before ready
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "starting", resp["status"])
}

// TestStartupHandler_AfterReady verifies 200 is returned after MarkReady (AC2).
func TestStartupHandler_AfterReady(t *testing.T) {
	handler := NewStartupHandler()

	// Signal ready
	handler.MarkReady()

	req := httptest.NewRequest(http.MethodGet, "/startupz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// AC2: 200 OK after ready
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]string
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "ready", resp["status"])
}

// TestStartupHandler_IsReady verifies IsReady reflects state correctly.
func TestStartupHandler_IsReady(t *testing.T) {
	handler := NewStartupHandler()

	// Initially false
	assert.False(t, handler.IsReady())

	// Mark ready
	handler.MarkReady()

	// Now true
	assert.True(t, handler.IsReady())
}

// TestStartupHandler_ThreadSafety verifies concurrent access safety.
func TestStartupHandler_ThreadSafety(t *testing.T) {
	handler := NewStartupHandler()

	// Create request once
	req := httptest.NewRequest(http.MethodGet, "/startupz", nil)

	// Synchronization
	startCh := make(chan struct{})
	done := make(chan struct{})

	// Writer goroutine
	go func() {
		<-startCh // Wait for start signal
		handler.MarkReady()
		close(done)
	}()

	// Reader goroutines
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startCh // Wait for start signal
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}()
	}

	// Unleash all goroutines
	close(startCh)

	// Wait for completion
	wg.Wait()
	<-done

	assert.True(t, handler.IsReady())
}

// BenchmarkStartupHandler_Starting benchmarks the pre-ready state (zero allocs).
func BenchmarkStartupHandler_Starting(b *testing.B) {
	handler := NewStartupHandler()
	req := httptest.NewRequest(http.MethodGet, "/startupz", nil)
	w := newBenchResponseWriter()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
		if w.code != http.StatusServiceUnavailable {
			b.Fatalf("Expected 503, got %d", w.code)
		}
	}
}

// BenchmarkStartupHandler_Ready benchmarks the ready state (zero allocs).
func BenchmarkStartupHandler_Ready(b *testing.B) {
	handler := NewStartupHandler()
	handler.MarkReady()
	req := httptest.NewRequest(http.MethodGet, "/startupz", nil)
	w := newBenchResponseWriter()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)
		if w.code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.code)
		}
	}
}
