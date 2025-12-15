package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

func TestErrorHandler(t *testing.T) {
	t.Run("normal handler continues without error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		ErrorHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("recovers from panic with string", func(t *testing.T) {
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("something went wrong")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		ErrorHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		// Verify response is Envelope format
		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Error == nil {
			t.Error("expected error in response")
		}

		if envelope.Error != nil && envelope.Error.Code != domainerrors.CodeInternalError {
			t.Errorf("expected code %s, got %s", domainerrors.CodeInternalError, envelope.Error.Code)
		}
	})

	t.Run("recovers from panic with error", func(t *testing.T) {
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic(http.ErrAbortHandler)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		ErrorHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("includes trace_id in error response", func(t *testing.T) {
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("panic with trace")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		// Add request ID to context
		ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace-123")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		ErrorHandler(handler).ServeHTTP(rec, req)

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Meta == nil {
			t.Fatal("expected meta in response")
		}

		if envelope.Meta.TraceID != "test-trace-123" {
			t.Errorf("expected trace_id %q, got %q", "test-trace-123", envelope.Meta.TraceID)
		}
	})

	t.Run("includes unknown trace_id when not in context", func(t *testing.T) {
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("panic without trace")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		ErrorHandler(handler).ServeHTTP(rec, req)

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Meta == nil {
			t.Fatal("expected meta in response")
		}

		if envelope.Meta.TraceID != response.UnknownTraceID {
			t.Errorf("expected trace_id %q, got %q", response.UnknownTraceID, envelope.Meta.TraceID)
		}
	})

	t.Run("error response message does not expose internal details", func(t *testing.T) {
		handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			panic("SENSITIVE: db connection failed at host=secret-db:5432")
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		ErrorHandler(handler).ServeHTTP(rec, req)

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Error == nil {
			t.Fatal("expected error in response")
		}

		// Message should be generic, not expose internal details
		if envelope.Error.Message != "internal server error" {
			t.Errorf("expected generic message, got %q", envelope.Error.Message)
		}
	})
}

func TestErrorHandler_ContentType(t *testing.T) {
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	ErrorHandler(handler).ServeHTTP(rec, req)

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}
