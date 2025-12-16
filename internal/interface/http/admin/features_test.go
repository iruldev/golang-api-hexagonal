package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

func TestFeaturesHandler_ListFlags(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore(
		runtimeutil.WithInitialFlags(map[string]bool{
			"flag_a": true,
			"flag_b": false,
		}),
	)
	handler := NewFeaturesHandler(store, nil)

	req := httptest.NewRequest(http.MethodGet, "/admin/features", nil)
	rec := httptest.NewRecorder()

	handler.ListFlags(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Success bool                           `json:"success"`
		Data    []runtimeutil.FeatureFlagState `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Len(t, resp.Data, 2)
}

func TestFeaturesHandler_GetFlag(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore(
		runtimeutil.WithInitialFlags(map[string]bool{
			"my_flag": true,
		}),
	)
	handler := NewFeaturesHandler(store, nil)

	// Create router with URL params
	r := chi.NewRouter()
	r.Get("/features/{flag}", handler.GetFlag)

	req := httptest.NewRequest(http.MethodGet, "/features/my_flag", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Success bool                         `json:"success"`
		Data    runtimeutil.FeatureFlagState `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "my_flag", resp.Data.Name)
	assert.True(t, resp.Data.Enabled)
}

func TestFeaturesHandler_GetFlag_FallsBackToEnv(t *testing.T) {
	// With the fix for Get() env fallback, nonexistent flags return synthesized state
	// (disabled by default from env) rather than 404
	store := runtimeutil.NewInMemoryFeatureFlagStore()
	handler := NewFeaturesHandler(store, nil)

	r := chi.NewRouter()
	r.Get("/features/{flag}", handler.GetFlag)

	req := httptest.NewRequest(http.MethodGet, "/features/nonexistent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Returns 200 with synthesized state from env (default: disabled)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Success bool                         `json:"success"`
		Data    runtimeutil.FeatureFlagState `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "nonexistent", resp.Data.Name)
	assert.False(t, resp.Data.Enabled) // Default from env
	assert.Equal(t, "(from environment)", resp.Data.Description)
}

func TestFeaturesHandler_EnableFlag(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore(
		runtimeutil.WithInitialFlags(map[string]bool{
			"my_flag": false,
		}),
	)
	handler := NewFeaturesHandler(store, nil)

	r := chi.NewRouter()
	r.Post("/features/{flag}/enable", handler.EnableFlag)

	// Add claims to context for audit logging
	req := httptest.NewRequest(http.MethodPost, "/features/my_flag/enable", nil)
	ctx := ctxutil.NewClaimsContext(req.Context(), ctxutil.Claims{UserID: "admin-user"})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Flag    string `json:"flag"`
			Enabled bool   `json:"enabled"`
		} `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "my_flag", resp.Data.Flag)
	assert.True(t, resp.Data.Enabled)

	// Verify flag is now enabled
	enabled, _ := store.IsEnabled(context.Background(), "my_flag")
	assert.True(t, enabled)
}

func TestFeaturesHandler_DisableFlag(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore(
		runtimeutil.WithInitialFlags(map[string]bool{
			"my_flag": true,
		}),
	)
	handler := NewFeaturesHandler(store, nil)

	r := chi.NewRouter()
	r.Post("/features/{flag}/disable", handler.DisableFlag)

	req := httptest.NewRequest(http.MethodPost, "/features/my_flag/disable", nil)
	ctx := ctxutil.NewClaimsContext(req.Context(), ctxutil.Claims{UserID: "admin-user"})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Flag    string `json:"flag"`
			Enabled bool   `json:"enabled"`
		} `json:"data"`
	}
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "my_flag", resp.Data.Flag)
	assert.False(t, resp.Data.Enabled)

	// Verify flag is now disabled
	enabled, _ := store.IsEnabled(context.Background(), "my_flag")
	assert.False(t, enabled)
}

func TestFeaturesHandler_EnableFlag_AutoRegisters(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore()
	handler := NewFeaturesHandler(store, nil)

	r := chi.NewRouter()
	r.Post("/features/{flag}/enable", handler.EnableFlag)

	req := httptest.NewRequest(http.MethodPost, "/features/new_flag/enable", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify flag is now enabled
	enabled, _ := store.IsEnabled(context.Background(), "new_flag")
	assert.True(t, enabled)
}

func TestFeaturesHandler_InvalidFlagName(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore()
	handler := NewFeaturesHandler(store, nil)

	r := chi.NewRouter()
	r.Get("/features/{flag}", handler.GetFlag)

	// Test with invalid flag name
	req := httptest.NewRequest(http.MethodGet, "/features/invalid@flag", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFeaturesHandler_NoClaims_DefaultsToUnknown(t *testing.T) {
	store := runtimeutil.NewInMemoryFeatureFlagStore(
		runtimeutil.WithInitialFlags(map[string]bool{
			"my_flag": false,
		}),
	)
	handler := NewFeaturesHandler(store, nil)

	r := chi.NewRouter()
	r.Post("/features/{flag}/enable", handler.EnableFlag)

	// No claims in context
	req := httptest.NewRequest(http.MethodPost, "/features/my_flag/enable", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	// Should still work, with actor defaulting to "unknown"
	assert.Equal(t, http.StatusCreated, rec.Code)
}

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

func TestFeaturesRoutes_NonAdminUser_Returns403(t *testing.T) {
	// This test validates that feature flag endpoints are protected by RBAC.
	// User is authenticated but does NOT have admin role.
	store := runtimeutil.NewInMemoryFeatureFlagStore(
		runtimeutil.WithInitialFlags(map[string]bool{
			"test_flag": true,
		}),
	)
	handler := NewFeaturesHandler(store, nil)

	// Create mock authenticator returning non-admin user
	auth := &mockAuthenticator{
		claims: &ctxutil.Claims{
			UserID: "user-123",
			Roles:  []string{"user"}, // No "admin" role
		},
	}

	// Create router with full middleware stack (auth + RBAC)
	r := chi.NewRouter()
	r.Route("/admin", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(auth, observability.NewNopLoggerInterface(), false))
		r.Use(middleware.RequireRole(observability.NewNopLoggerInterface(), "admin"))
		r.Get("/features", handler.ListFlags)
		r.Get("/features/{flag}", handler.GetFlag)
		r.Post("/features/{flag}/enable", handler.EnableFlag)
		r.Post("/features/{flag}/disable", handler.DisableFlag)
	})

	// Test all endpoints return 403
	testCases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/admin/features"},
		{http.MethodGet, "/admin/features/test_flag"},
		{http.MethodPost, "/admin/features/test_flag/enable"},
		{http.MethodPost, "/admin/features/test_flag/disable"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusForbidden, rec.Code, "Expected 403 for %s %s", tc.method, tc.path)
		})
	}
}
