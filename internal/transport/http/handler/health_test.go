package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthHandler_ServeHTTP(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rec.Code)

	// Check content type
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// Check response body
	var resp HealthResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "ok", resp.Data.Status)
}

// =============================================================================
// Story 3.1: Liveness Probe Unit Tests (AC: #1, #2)
// =============================================================================

// TestLivenessHandler_ServeHTTP verifies the liveness probe handler behavior.
// This tests AC1: /healthz returns 200 OK with minimal body.
func TestLivenessHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name               string
		method             string
		expectedStatus     int
		expectedStatusText string
	}{
		{
			name:               "GET request returns 200 OK with alive status",
			method:             http.MethodGet,
			expectedStatus:     http.StatusOK,
			expectedStatusText: "alive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewLivenessHandler()

			req := httptest.NewRequest(tt.method, "/healthz", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			// AC1: 200 OK status code
			assert.Equal(t, tt.expectedStatus, rec.Code, "Expected status code %d", tt.expectedStatus)

			// AC1: Correct Content-Type header
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Expected application/json Content-Type")

			// AC1: Minimal JSON response body
			var resp LivenessResponse
			err := json.NewDecoder(rec.Body).Decode(&resp)
			require.NoError(t, err, "Response body should be valid JSON")

			assert.Equal(t, tt.expectedStatusText, resp.Status, "Response status should be 'alive'")
		})
	}
}

// TestLivenessHandler_ResponseStructure verifies the response structure is flat (no nested data).
func TestLivenessHandler_ResponseStructure(t *testing.T) {
	handler := NewLivenessHandler()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Verify response is flat JSON with only "status" field
	var rawResp map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&rawResp)
	require.NoError(t, err)

	// Should only have "status" field (no "data" wrapper like /health)
	assert.Len(t, rawResp, 1, "Response should have exactly 1 field")
	_, hasStatus := rawResp["status"]
	assert.True(t, hasStatus, "Response should have 'status' field")
	_, hasData := rawResp["data"]
	assert.False(t, hasData, "Response should NOT have 'data' wrapper")
}

// TestLivenessHandler_NoDependencyChecks verifies the handler doesn't perform any dependency checks.
// This is a structural validation - the handler is intentionally stateless.
func TestLivenessHandler_NoDependencyChecks(t *testing.T) {
	// Liveness handler has no dependencies injected - this is by design
	handler := NewLivenessHandler()

	// Handler should work without any setup
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Should succeed immediately
	assert.Equal(t, http.StatusOK, rec.Code)
}

// =============================================================================
// Story 3.1: Liveness Probe Benchmark Tests (AC: #1, #3)
// =============================================================================

// HeaderMap is a simple map for headers to avoid allocations in benchmarks
type benchHeaderMap map[string][]string

func (h benchHeaderMap) Set(key, value string) {
	h[key] = []string{value}
}
func (h benchHeaderMap) Add(key, value string) {
	h[key] = append(h[key], value)
}
func (h benchHeaderMap) Del(key string) {
	delete(h, key)
}
func (h benchHeaderMap) Get(key string) string {
	if v, ok := h[key]; ok && len(v) > 0 {
		return v[0]
	}
	return ""
}
func (h benchHeaderMap) Values(key string) []string {
	return h[key]
}

// benchResponseWriter is a mock http.ResponseWriter that avoids allocations
type benchResponseWriter struct {
	header http.Header
	code   int
}

func newBenchResponseWriter() *benchResponseWriter {
	return &benchResponseWriter{
		header: make(http.Header, 2), // Pre-allocate space for expected headers
	}
}

func (w *benchResponseWriter) Header() http.Header {
	return w.header
}

func (w *benchResponseWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func (w *benchResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

// BenchmarkLivenessHandler validates performance requirements.
// Target: <10ms p99 response time, minimal allocations.
func BenchmarkLivenessHandler(b *testing.B) {
	handler := NewLivenessHandler()

	// Create request once
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	// Use our zero-alloc writer
	w := newBenchResponseWriter()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(w, req)

		if w.code != http.StatusOK {
			b.Fatalf("Expected 200 OK, got %d", w.code)
		}
	}
}

// BenchmarkLivenessHandler_Parallel validates performance under concurrent load.
// This simulates K8s kubelet polling from multiple nodes.
func BenchmarkLivenessHandler_Parallel(b *testing.B) {
	handler := NewLivenessHandler()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				b.Fatalf("Expected 200 OK, got %d", rec.Code)
			}
		}
	})
}
