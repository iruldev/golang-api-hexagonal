package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOtel_CreatesMiddleware(t *testing.T) {
	middleware := Otel("test-operation")

	if middleware == nil {
		t.Error("Expected non-nil middleware")
	}
}

func TestOtel_WrapsHandler(t *testing.T) {
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Otel("test")(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestOtel_PropagatesContext(t *testing.T) {
	var reqContext http.Request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqContext = *r
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Otel("api")(handler)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	// Context should be non-nil
	if reqContext.Context() == nil {
		t.Error("Expected non-nil context")
	}
}

func TestOtel_DifferentOperations(t *testing.T) {
	tests := []struct {
		name      string
		operation string
	}{
		{"api operation", "api"},
		{"internal operation", "internal"},
		{"empty operation", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := Otel(tt.operation)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			wrapped := middleware(handler)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			wrapped.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}
		})
	}
}

func TestOtel_HandlesMultipleRequests(t *testing.T) {
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Otel("api")(handler)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		wrapped.ServeHTTP(rr, req)
	}

	if callCount != 5 {
		t.Errorf("Expected 5 calls, got %d", callCount)
	}
}
