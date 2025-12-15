package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

func TestWrapHandler(t *testing.T) {
	t.Run("handler returns nil - success path", func(t *testing.T) {
		handler := func(w http.ResponseWriter, r *http.Request) error {
			response.SuccessEnvelope(w, r.Context(), map[string]string{"status": "ok"})
			return nil
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace-id")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		WrapHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Error != nil {
			t.Error("expected no error in response")
		}
	})

	t.Run("handler returns DomainError", func(t *testing.T) {
		handler := func(_ http.ResponseWriter, _ *http.Request) error {
			return domainerrors.NewDomain(domainerrors.CodeNotFound, "user not found")
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace-id")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		WrapHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Error == nil {
			t.Fatal("expected error in response")
		}

		if envelope.Error.Code != domainerrors.CodeNotFound {
			t.Errorf("expected code %s, got %s", domainerrors.CodeNotFound, envelope.Error.Code)
		}

		if envelope.Error.Message != "user not found" {
			t.Errorf("expected message %q, got %q", "user not found", envelope.Error.Message)
		}
	})

	t.Run("handler returns DomainError with hint", func(t *testing.T) {
		handler := func(_ http.ResponseWriter, _ *http.Request) error {
			return domainerrors.NewDomainWithHint(
				domainerrors.CodeValidationError,
				"invalid email format",
				"email should be in format user@example.com",
			)
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace-id")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		WrapHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, rec.Code)
		}

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Error == nil {
			t.Fatal("expected error in response")
		}

		if envelope.Error.Hint != "email should be in format user@example.com" {
			t.Errorf("expected hint, got %q", envelope.Error.Hint)
		}
	})

	t.Run("handler returns legacy sentinel error", func(t *testing.T) {
		handler := func(_ http.ResponseWriter, _ *http.Request) error {
			return domain.ErrNotFound
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace-id")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		WrapHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("handler returns unknown error - maps to 500", func(t *testing.T) {
		handler := func(_ http.ResponseWriter, _ *http.Request) error {
			return errors.New("unexpected database error")
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace-id")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		WrapHandler(handler).ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})

	t.Run("trace_id included in error response", func(t *testing.T) {
		handler := func(_ http.ResponseWriter, _ *http.Request) error {
			return domainerrors.NewDomain(domainerrors.CodeNotFound, "not found")
		}

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxutil.NewRequestIDContext(req.Context(), "my-custom-trace")
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		WrapHandler(handler).ServeHTTP(rec, req)

		var envelope response.Envelope
		if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if envelope.Meta == nil {
			t.Fatal("expected meta in response")
		}

		if envelope.Meta.TraceID != "my-custom-trace" {
			t.Errorf("expected trace_id %q, got %q", "my-custom-trace", envelope.Meta.TraceID)
		}
	})
}

func TestWrapHandler_ErrorCodeMapping(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "NOT_FOUND",
			err:            domainerrors.NewDomain(domainerrors.CodeNotFound, "not found"),
			expectedStatus: http.StatusNotFound,
			expectedCode:   domainerrors.CodeNotFound,
		},
		{
			name:           "VALIDATION_ERROR",
			err:            domainerrors.NewDomain(domainerrors.CodeValidationError, "validation error"),
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   domainerrors.CodeValidationError,
		},
		{
			name:           "UNAUTHORIZED",
			err:            domainerrors.NewDomain(domainerrors.CodeUnauthorized, "unauthorized"),
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   domainerrors.CodeUnauthorized,
		},
		{
			name:           "FORBIDDEN",
			err:            domainerrors.NewDomain(domainerrors.CodeForbidden, "forbidden"),
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden,
		},
		{
			name:           "CONFLICT",
			err:            domainerrors.NewDomain(domainerrors.CodeConflict, "conflict"),
			expectedStatus: http.StatusConflict,
			expectedCode:   domainerrors.CodeConflict,
		},
		{
			name:           "INTERNAL_ERROR",
			err:            domainerrors.NewDomain(domainerrors.CodeInternalError, "internal error"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError,
		},
		{
			name:           "TIMEOUT",
			err:            domainerrors.NewDomain(domainerrors.CodeTimeout, "timeout"),
			expectedStatus: http.StatusGatewayTimeout,
			expectedCode:   domainerrors.CodeTimeout,
		},
		{
			name:           "RATE_LIMIT_EXCEEDED",
			err:            domainerrors.NewDomain(domainerrors.CodeRateLimitExceeded, "rate limit"),
			expectedStatus: http.StatusTooManyRequests,
			expectedCode:   domainerrors.CodeRateLimitExceeded,
		},
		{
			name:           "BAD_REQUEST",
			err:            domainerrors.NewDomain(domainerrors.CodeBadRequest, "bad request"),
			expectedStatus: http.StatusBadRequest,
			expectedCode:   domainerrors.CodeBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := func(_ http.ResponseWriter, _ *http.Request) error {
				return tt.err
			}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			ctx := ctxutil.NewRequestIDContext(req.Context(), "test-trace")
			req = req.WithContext(ctx)
			rec := httptest.NewRecorder()

			WrapHandler(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			var envelope response.Envelope
			if err := json.NewDecoder(rec.Body).Decode(&envelope); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if envelope.Error == nil {
				t.Fatal("expected error in response")
			}

			if envelope.Error.Code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, envelope.Error.Code)
			}
		})
	}
}
