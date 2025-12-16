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
type envelopeResponse = response.TestEnvelopeResponse

func TestRequireRole(t *testing.T) {
	logger := observability.NewNopLoggerInterface()

	tests := []struct {
		name           string
		role           string
		claims         ctxutil.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has exact role",
			role:           "admin",
			claims:         ctxutil.Claims{Roles: []string{"admin"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Missing role",
			role:           "admin",
			claims:         ctxutil.Claims{Roles: []string{"user"}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeInsufficientRole,
		},
		{
			name:           "Error - No claims",
			role:           "admin",
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequireRole(logger, tt.role)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				var env envelopeResponse
				err := json.Unmarshal(rec.Body.Bytes(), &env)
				assert.NoError(t, err)
				assert.NotNil(t, env.Error)
				assert.Equal(t, tt.expectedCode, env.Error.Code)
				assert.NotNil(t, env.Meta)
				assert.NotEmpty(t, env.Meta.TraceID)
			}
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	logger := observability.NewNopLoggerInterface()

	tests := []struct {
		name           string
		roles          []string
		claims         ctxutil.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed - Has first role",
			roles:          []string{"admin", "editor"},
			claims:         ctxutil.Claims{Roles: []string{"admin"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Allowed - Has second role",
			roles:          []string{"admin", "editor"},
			claims:         ctxutil.Claims{Roles: []string{"editor"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied - Has none of roles",
			roles:          []string{"admin", "editor"},
			claims:         ctxutil.Claims{Roles: []string{"user"}, UserID: "user-3"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeInsufficientRole,
		},
		{
			name:           "Denied - Empty roles list",
			roles:          []string{},
			claims:         ctxutil.Claims{Roles: []string{"admin"}, UserID: "user-4"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeInsufficientRole,
		},
		{
			name:           "Error - No claims",
			roles:          []string{"admin"},
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Using Variadic signature as per plan
			handler := middleware.RequireAnyRole(logger, tt.roles...)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				var env envelopeResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &env)
				assert.Equal(t, tt.expectedCode, env.Error.Code)
				assert.NotNil(t, env.Meta)
				assert.NotEmpty(t, env.Meta.TraceID)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	logger := observability.NewNopLoggerInterface()

	tests := []struct {
		name           string
		perm           string
		claims         ctxutil.Claims
		hasContext     bool
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Allowed",
			perm:           "write",
			claims:         ctxutil.Claims{Permissions: []string{"write"}, UserID: "user-1"},
			hasContext:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Denied",
			perm:           "write",
			claims:         ctxutil.Claims{Permissions: []string{"read"}, UserID: "user-2"},
			hasContext:     true,
			expectedStatus: http.StatusForbidden,
			expectedCode:   domainerrors.CodeInsufficientPermission,
		},
		{
			name:           "Error - No claims",
			perm:           "write",
			hasContext:     false,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   domainerrors.CodeInternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := middleware.RequirePermission(logger, tt.perm)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				var env envelopeResponse
				_ = json.Unmarshal(rec.Body.Bytes(), &env)
				assert.Equal(t, tt.expectedCode, env.Error.Code)
				assert.NotNil(t, env.Meta)
				assert.NotEmpty(t, env.Meta.TraceID)
			}
		})
	}
}
