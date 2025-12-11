package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// testConfig returns a minimal config for testing.
func testConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		Log: config.LogConfig{
			Level:  "debug",
			Format: "console",
		},
	}
}

func TestRegisterRoutes_HealthEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestRegisterRoutes_ExampleEndpoint(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Verify JSON response
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestRegisterRoutes_NotFound(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", RegisterRoutes)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/nonexistent", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rr.Code)
	}
}

func TestRegisterRoutes_MethodNotAllowed(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", RegisterRoutes)

	// POST to health endpoint should fail (only GET allowed)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestRegisterRoutes_APIv1Prefix(t *testing.T) {
	r := chi.NewRouter()
	r.Route("/api/v1", RegisterRoutes)

	// Without /api/v1 prefix should return 404
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for /health without prefix, got %d", rr.Code)
	}

	// With /api/v1 prefix should return 200
	req = httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr = httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /api/v1/health, got %d", rr.Code)
	}
}

// TestNewRouter_MiddlewareApplied verifies AC2: middleware chain is applied automatically.
// This test uses the full NewRouter to ensure middleware is wired correctly.
func TestNewRouter_MiddlewareApplied(t *testing.T) {
	cfg := testConfig()

	router := NewRouter(cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Verify RequestID middleware is applied (Story 3.2)
	requestID := rr.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("Expected X-Request-ID header to be set by middleware")
	}

	// Verify response is successful
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

// TestNewRouter_RequestIDMiddleware verifies that RequestID middleware generates unique IDs.
func TestNewRouter_RequestIDMiddleware(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	// First request
	req1 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	id1 := rr1.Header().Get("X-Request-ID")

	// Second request
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, req2)
	id2 := rr2.Header().Get("X-Request-ID")

	// IDs should be different
	if id1 == id2 {
		t.Errorf("Expected different request IDs, got same: %s", id1)
	}
}

// TestNewRouter_ExistingRequestID verifies that middleware uses existing X-Request-ID if provided.
func TestNewRouter_ExistingRequestID(t *testing.T) {
	cfg := testConfig()
	router := NewRouter(cfg)

	existingID := "my-custom-request-id"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", existingID)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	responseID := rr.Header().Get("X-Request-ID")
	if responseID != existingID {
		t.Errorf("Expected existing request ID %s, got %s", existingID, responseID)
	}
}

// TestNewRouter_RecoveryMiddleware verifies Recovery middleware catches panics.
func TestNewRouter_RecoveryMiddleware(t *testing.T) {
	// Create router with a panicking handler to test recovery
	r := chi.NewRouter()
	r.Use(middleware.Recovery(observability.NewNopLogger())) // Use NopLogger for test
	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	// Should not panic, recovery middleware catches it
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 after panic, got %d", rr.Code)
	}
}
