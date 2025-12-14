package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		roles          []string
		claims         middleware.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has exact role",
			roles:          []string{"admin"},
			claims:         middleware.Claims{Roles: []string{"admin"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed - Has one of multiple required roles",
			roles:          []string{"admin", "editor"},
			claims:         middleware.Claims{Roles: []string{"editor"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing role",
			roles:          []string{"admin"},
			claims:         middleware.Claims{Roles: []string{"user"}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "ERR_INSUFFICIENT_ROLE",
		},
		{
			name:           "Denied - No roles in claims",
			roles:          []string{"admin"},
			claims:         middleware.Claims{UserID: "user-4"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "ERR_INSUFFICIENT_ROLE",
		},
		{
			name:           "Error - No claims in context (Server Error)",
			roles:          []string{"admin"},
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequireRole(tt.roles...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasContext {
				ctx := middleware.NewContext(req.Context(), tt.claims)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedCode != "" {
				// We expect the error response to follow standard error format
				// But for middleware testing, checking status code is often primary.
				// If we want to check body, we'd need to parse the JSON.
				// For now, let's assume status code is enough or verify the body contains the code.
				assert.Contains(t, rec.Body.String(), tt.expectedCode)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		perms          []string
		claims         middleware.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has required permission",
			perms:          []string{"note:read"},
			claims:         middleware.Claims{Permissions: []string{"note:read"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed - Has all required permissions",
			perms:          []string{"note:read", "note:write"},
			claims:         middleware.Claims{Permissions: []string{"note:read", "note:write"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing one of required permissions (AND logic)",
			perms:          []string{"note:read", "note:write"},
			claims:         middleware.Claims{Permissions: []string{"note:read"}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "ERR_INSUFFICIENT_PERMISSION",
		},
		{
			name:           "Denied - Missing permission",
			perms:          []string{"note:admin"},
			claims:         middleware.Claims{Permissions: []string{"note:read"}, UserID: "user-4"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "ERR_INSUFFICIENT_PERMISSION",
		},
		{
			name:           "Error - No claims in context",
			perms:          []string{"note:read"},
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequirePermission(tt.perms...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasContext {
				ctx := middleware.NewContext(req.Context(), tt.claims)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedCode != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedCode)
			}
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name           string
		perms          []string
		claims         middleware.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has one of required permissions",
			perms:          []string{"note:read", "note:unique"},
			claims:         middleware.Claims{Permissions: []string{"note:read"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing all required permissions",
			perms:          []string{"note:admin", "note:special"},
			claims:         middleware.Claims{Permissions: []string{"note:read"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   "ERR_INSUFFICIENT_PERMISSION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequireAnyPermission(tt.perms...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/", nil)
			if tt.hasContext {
				ctx := middleware.NewContext(req.Context(), tt.claims)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedCode != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedCode)
			}
		})
	}
}
