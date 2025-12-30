package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockShutdownCoordinator is a mock implementation for testing.
// Only implements the methods required by the ShutdownCoordinator interface.
type mockShutdownCoordinator struct {
	shuttingDown bool
	activeCount  int64
}

func (m *mockShutdownCoordinator) IncrementActive() bool {
	if m.shuttingDown {
		return false
	}
	m.activeCount++
	return true
}

func (m *mockShutdownCoordinator) DecrementActive() {
	m.activeCount--
}

func TestShutdownMiddleware(t *testing.T) {
	tests := []struct {
		name              string
		shuttingDown      bool
		expectedStatus    int
		expectedActive    int64
		expectNextCalled  bool
		expectRetryAfter  bool
		expectProblemJSON bool
	}{
		{
			name:              "allows request when not shutting down",
			shuttingDown:      false,
			expectedStatus:    http.StatusOK,
			expectedActive:    0, // Decremented after request
			expectNextCalled:  true,
			expectRetryAfter:  false,
			expectProblemJSON: false,
		},
		{
			name:              "rejects request when shutting down",
			shuttingDown:      true,
			expectedStatus:    http.StatusServiceUnavailable,
			expectedActive:    0, // Never incremented
			expectNextCalled:  false,
			expectRetryAfter:  true,
			expectProblemJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord := &mockShutdownCoordinator{shuttingDown: tt.shuttingDown}
			nextCalled := false

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			})

			handler := Shutdown(coord)(next)

			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, tt.expectedActive, coord.activeCount)
			assert.Equal(t, tt.expectNextCalled, nextCalled)

			if tt.expectRetryAfter {
				assert.NotEmpty(t, rec.Header().Get("Retry-After"))
				assert.Equal(t, "close", rec.Header().Get("Connection"))
			}

			if tt.expectProblemJSON {
				assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))
			}
		})
	}
}

func TestShutdownMiddleware_ResponseBody(t *testing.T) {
	coord := &mockShutdownCoordinator{shuttingDown: true}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next should not be called")
	})

	handler := Shutdown(coord)(next)

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	// Type URL comes from contract package's configurable base URL
	require.Contains(t, body, `"type":`)
	require.Contains(t, body, `service-unavailable`)
	require.Contains(t, body, `"title":"Service Unavailable"`)
	require.Contains(t, body, `"status":503`)
	require.Contains(t, body, `"detail":"Server is shutting down. Please retry later."`)
	require.Contains(t, body, `"instance":"/api/users"`)
	require.Contains(t, body, `"code":"SYS-002"`)
}

func TestShutdownMiddleware_TracksActiveCount(t *testing.T) {
	coord := &mockShutdownCoordinator{shuttingDown: false}

	activeCountDuringRequest := int64(0)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture active count during request processing
		activeCountDuringRequest = coord.activeCount
		w.WriteHeader(http.StatusOK)
	})

	handler := Shutdown(coord)(next)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Active count should be 1 during the request
	assert.Equal(t, int64(1), activeCountDuringRequest)

	// Active count should be 0 after the request
	assert.Equal(t, int64(0), coord.activeCount)
}

// Note: writeShutdownProblemJSON has a json.Marshal error fallback branch that cannot
// be easily tested because ProblemDetail only contains JSON-serializable types.
// The fallback is defensive code for extremely rare scenarios (memory corruption, etc.)
// and is now RFC 7807 compliant using contract.ProblemTypeURL().
