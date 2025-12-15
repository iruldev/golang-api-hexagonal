package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/stretchr/testify/assert"
)

// envelopeResponse is an alias to the shared test response type.
// This avoids code duplication across test files.
type envelopeResponse = response.TestEnvelopeResponse

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		roles          []string
		claims         ctxutil.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has exact role",
			roles:          []string{"admin"},
			claims:         ctxutil.Claims{Roles: []string{"admin"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed - Has one of multiple required roles",
			roles:          []string{"admin", "editor"},
			claims:         ctxutil.Claims{Roles: []string{"editor"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing role",
			roles:          []string{"admin"},
			claims:         ctxutil.Claims{Roles: []string{"user"}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden, // "FORBIDDEN"
		},
		{
			name:           "Denied - No roles in claims",
			roles:          []string{"admin"},
			claims:         ctxutil.Claims{UserID: "user-4"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden, // "FORBIDDEN"
		},
		{
			name:           "Error - No claims in context (Server Error)",
			roles:          []string{"admin"},
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError, // "INTERNAL_ERROR"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := observability.NewNopLoggerInterface()
			handler := middleware.RequireRole(tt.roles, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasContext {
				ctx := ctxutil.NewClaimsContext(req.Context(), tt.claims)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedCode != "" {
				// Parse response body as Envelope
				var env envelopeResponse
				err := json.Unmarshal(rec.Body.Bytes(), &env)
				assert.NoError(t, err)
				assert.NotNil(t, env.Error)
				assert.Equal(t, tt.expectedCode, env.Error.Code)
				// Verify trace_id is present in error responses (consistency with auth_test.go)
				assert.NotNil(t, env.Meta, "expected meta in error response")
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		perms          []string
		claims         ctxutil.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has required permission",
			perms:          []string{"note:read"},
			claims:         ctxutil.Claims{Permissions: []string{"note:read"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed - Has all required permissions",
			perms:          []string{"note:read", "note:write"},
			claims:         ctxutil.Claims{Permissions: []string{"note:read", "note:write"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing one of required permissions (AND logic)",
			perms:          []string{"note:read", "note:write"},
			claims:         ctxutil.Claims{Permissions: []string{"note:read"}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden, // "FORBIDDEN"
		},
		{
			name:           "Denied - Missing permission",
			perms:          []string{"note:admin"},
			claims:         ctxutil.Claims{Permissions: []string{"note:read"}, UserID: "user-4"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden, // "FORBIDDEN"
		},
		{
			name:           "Error - No claims in context",
			perms:          []string{"note:read"},
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError, // "INTERNAL_ERROR"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := observability.NewNopLoggerInterface()
			handler := middleware.RequirePermission(tt.perms, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasContext {
				ctx := ctxutil.NewClaimsContext(req.Context(), tt.claims)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedCode != "" {
				// Parse response body as Envelope
				var env envelopeResponse
				err := json.Unmarshal(rec.Body.Bytes(), &env)
				assert.NoError(t, err)
				assert.NotNil(t, env.Error)
				assert.Equal(t, tt.expectedCode, env.Error.Code)
				// Verify trace_id is present in error responses (consistency with auth_test.go)
				assert.NotNil(t, env.Meta, "expected meta in error response")
			}
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name           string
		perms          []string
		claims         ctxutil.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has one of required permissions",
			perms:          []string{"note:read", "note:unique"},
			claims:         ctxutil.Claims{Permissions: []string{"note:read"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing all required permissions",
			perms:          []string{"note:admin", "note:special"},
			claims:         ctxutil.Claims{Permissions: []string{"note:read"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden, // "FORBIDDEN"
		},
		{
			name:           "Denied - Empty permissions array",
			perms:          []string{"note:read"},
			claims:         ctxutil.Claims{Permissions: []string{}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeForbidden, // "FORBIDDEN"
		},
		{
			name:           "Error - No claims in context",
			perms:          []string{"note:read"},
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError, // "INTERNAL_ERROR"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := observability.NewNopLoggerInterface()
			handler := middleware.RequireAnyPermission(tt.perms, logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasContext {
				ctx := ctxutil.NewClaimsContext(req.Context(), tt.claims)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedCode != "" {
				// Parse response body as Envelope
				var env envelopeResponse
				err := json.Unmarshal(rec.Body.Bytes(), &env)
				assert.NoError(t, err)
				assert.NotNil(t, env.Error)
				assert.Equal(t, tt.expectedCode, env.Error.Code)
				// Verify trace_id is present in error responses (consistency with auth_test.go)
				assert.NotNil(t, env.Meta, "expected meta in error response")
			}
		})
	}
}
