package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/stretchr/testify/assert"
)

// mockAuthenticator is a test double for the Authenticator interface.
type mockAuthenticator struct {
	claims *ctxutil.Claims
	err    error
}

func (m *mockAuthenticator) Authenticate(r *http.Request) (ctxutil.Claims, error) {
	if m.err != nil {
		return ctxutil.Claims{}, m.err
	}
	return *m.claims, nil
}

func TestRouter_AdminRoutes_WithAuthenticator(t *testing.T) {
	tests := []struct {
		name         string
		auth         *mockAuthenticator
		wantStatus   int
		wantContains string
	}{
		{
			name: "no_token_returns_401",
			auth: &mockAuthenticator{
				err: middleware.ErrUnauthenticated,
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_token_no_admin_role_returns_403",
			auth: &mockAuthenticator{
				claims: &ctxutil.Claims{
					UserID: "user-123",
					Roles:  []string{"user"},
				},
			},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "valid_token_with_admin_role_returns_200",
			auth: &mockAuthenticator{
				claims: &ctxutil.Claims{
					UserID: "admin-123",
					Roles:  []string{"admin"},
				},
			},
			wantStatus:   http.StatusOK,
			wantContains: `"admin_access":true`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create real router with authenticator
			cfg := &config.Config{
				App: config.AppConfig{Env: "test"},
			}
			deps := RouterDeps{
				Config:        cfg,
				Authenticator: tt.auth,
			}
			router := NewRouter(deps)

			req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
			rr := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantContains != "" {
				assert.Contains(t, rr.Body.String(), tt.wantContains)
			}
		})
	}
}

func TestRouter_AdminRoutes_NoAuthenticator_NotMounted(t *testing.T) {
	// Arrange: Create router WITHOUT authenticator
	cfg := &config.Config{
		App: config.AppConfig{Env: "test"},
	}
	deps := RouterDeps{
		Config:        cfg,
		Authenticator: nil, // No authenticator
	}
	router := NewRouter(deps)

	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	rr := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rr, req)

	// Assert: Should be 404 because admin routes are not mounted
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestRouter_SecurityHeaders(t *testing.T) {
	// Setup minimal dependencies
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
	}
	deps := RouterDeps{
		Config: cfg,
	}

	router := NewRouter(deps)

	// Create a request to a known endpoint (e.g., /healthz)
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify security headers are present
	assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rec.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", rec.Header().Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'", rec.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "max-age=63072000; includeSubDomains", rec.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", rec.Header().Get("Permissions-Policy"))
}
