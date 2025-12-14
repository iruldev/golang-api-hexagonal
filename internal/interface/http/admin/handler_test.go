package admin_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/admin"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
)

// mockAuthenticator is a test double for the Authenticator interface.
type mockAuthenticator struct {
	claims *middleware.Claims
	err    error
}

func (m *mockAuthenticator) Authenticate(r *http.Request) (middleware.Claims, error) {
	if m.err != nil {
		return middleware.Claims{}, m.err
	}
	return *m.claims, nil
}

// setupTestRouter creates a router with admin routes protected by auth and RBAC middleware.
func setupTestRouter(authenticator middleware.Authenticator) chi.Router {
	r := chi.NewRouter()

	// Apply middleware in correct order: Auth before RBAC
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authenticator))
		r.Use(middleware.RequireRole("admin"))
		r.Get("/health", admin.HealthHandler)
	})

	return r
}

func TestAdminRoutes_NoToken_Returns401(t *testing.T) {
	// Arrange: Authenticator returns unauthenticated error
	auth := &mockAuthenticator{
		err: middleware.ErrUnauthenticated,
	}
	router := setupTestRouter(auth)

	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	rr := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAdminRoutes_ValidTokenNoAdminRole_Returns403(t *testing.T) {
	// Arrange: User is authenticated but has no admin role
	auth := &mockAuthenticator{
		claims: &middleware.Claims{
			UserID: "user-123",
			Roles:  []string{"user"},
		},
	}
	router := setupTestRouter(auth)

	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	rr := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestAdminRoutes_ValidTokenWithAdminRole_Returns200(t *testing.T) {
	// Arrange: User is authenticated with admin role
	auth := &mockAuthenticator{
		claims: &middleware.Claims{
			UserID: "admin-123",
			Roles:  []string{"admin"},
		},
	}
	router := setupTestRouter(auth)

	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	rr := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"admin_access":true`)
}

func TestAdminRoutes_MiddlewareOrder_AuthBeforeRBAC(t *testing.T) {
	// This test validates that when claims retrieval fails (due to no auth),
	// we get 401, not 500 (which would indicate RBAC running before auth)

	// Arrange: Authenticator returns token invalid error
	auth := &mockAuthenticator{
		err: middleware.ErrTokenInvalid,
	}
	router := setupTestRouter(auth)

	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	rr := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rr, req)

	// Assert: Should be 401 (auth problem), not 500 (RBAC misconfiguration)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAdminHealthHandler(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/admin/health", nil)
	rr := httptest.NewRecorder()

	// Act
	admin.HealthHandler(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	body := rr.Body.String()
	require.Contains(t, body, `"success":true`)
	require.Contains(t, body, `"status":"ok"`)
	require.Contains(t, body, `"admin_access":true`)
}
