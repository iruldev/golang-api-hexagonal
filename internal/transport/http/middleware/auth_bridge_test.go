//go:build !integration

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

func TestAuthContextBridge_WithClaims(t *testing.T) {
	// Setup: Create a handler that captures the auth context
	var capturedAuthCtx *app.AuthContext
	handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		capturedAuthCtx = app.GetAuthContext(r.Context())
	})

	// Create claims and set them in context
	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
		Role: "admin",
	}
	ctx := markClaimsValidatedTestOnly(context.Background(), claims)
	ctx = setValidatedClaims(ctx)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	// Execute middleware
	AuthContextBridge(handler).ServeHTTP(w, req)

	// Verify auth context was populated from claims
	require.NotNil(t, capturedAuthCtx)
	assert.Equal(t, "user-123", capturedAuthCtx.SubjectID)
	assert.Equal(t, "admin", capturedAuthCtx.Role)
	assert.True(t, capturedAuthCtx.IsAdmin())
}

func TestAuthContextBridge_WithoutClaims(t *testing.T) {
	// Setup: Create a handler that captures the auth context
	var capturedAuthCtx *app.AuthContext
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedAuthCtx = app.GetAuthContext(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute middleware (no claims in context)
	AuthContextBridge(handler).ServeHTTP(w, req)

	// Verify handler was called but auth context is nil
	assert.True(t, handlerCalled)
	assert.Nil(t, capturedAuthCtx)
}

func TestAuthContextBridge_IgnoresUnvalidatedClaims(t *testing.T) {
	// Claims are present but not marked as validated by JWTAuth
	var capturedAuthCtx *app.AuthContext
	handlerCalled := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		capturedAuthCtx = app.GetAuthContext(r.Context())
	})

	claims := &ctxutil.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: "user-123",
		},
		Role: "admin",
	}
	ctx := ctxutil.SetClaims(context.Background(), claims) // no validated flag

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	AuthContextBridge(handler).ServeHTTP(w, req)

	assert.True(t, handlerCalled)
	assert.Nil(t, capturedAuthCtx, "auth context must not be set when claims are unvalidated")
}

func TestAuthContextBridge_PreservesSubject(t *testing.T) {
	tests := []struct {
		name            string
		subject         string
		role            string
		expectedSubject string
		expectedRole    string
	}{
		{
			name:            "admin user",
			subject:         "admin-user-id",
			role:            app.RoleAdmin,
			expectedSubject: "admin-user-id",
			expectedRole:    app.RoleAdmin,
		},
		{
			name:            "regular user",
			subject:         "regular-user-id",
			role:            app.RoleUser,
			expectedSubject: "regular-user-id",
			expectedRole:    app.RoleUser,
		},
		{
			name:            "empty role",
			subject:         "user-no-role",
			role:            "",
			expectedSubject: "user-no-role",
			expectedRole:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedAuthCtx *app.AuthContext
			handler := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				capturedAuthCtx = app.GetAuthContext(r.Context())
			})

			claims := &ctxutil.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Subject: tt.subject,
				},
				Role: tt.role,
			}
			ctx := markClaimsValidatedTestOnly(context.Background(), claims)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			AuthContextBridge(handler).ServeHTTP(w, req)

			require.NotNil(t, capturedAuthCtx)
			assert.Equal(t, tt.expectedSubject, capturedAuthCtx.SubjectID)
			assert.Equal(t, tt.expectedRole, capturedAuthCtx.Role)
		})
	}
}
