package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// mockUserRoleProvider implements UserRoleProvider for testing.
type mockUserRoleProvider struct {
	getUserRolesFunc   func(ctx context.Context, userID string) (*runtimeutil.UserRoles, error)
	setUserRolesFunc   func(ctx context.Context, userID string, roles []string) (*runtimeutil.UserRoles, error)
	addUserRoleFunc    func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error)
	removeUserRoleFunc func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error)
}

func (m *mockUserRoleProvider) GetUserRoles(ctx context.Context, userID string) (*runtimeutil.UserRoles, error) {
	if m.getUserRolesFunc != nil {
		return m.getUserRolesFunc(ctx, userID)
	}
	return &runtimeutil.UserRoles{
		UserID:    userID,
		Roles:     []string{},
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserRoleProvider) SetUserRoles(ctx context.Context, userID string, roles []string) (*runtimeutil.UserRoles, error) {
	if m.setUserRolesFunc != nil {
		return m.setUserRolesFunc(ctx, userID, roles)
	}
	return &runtimeutil.UserRoles{
		UserID:    userID,
		Roles:     roles,
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserRoleProvider) AddUserRole(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
	if m.addUserRoleFunc != nil {
		return m.addUserRoleFunc(ctx, userID, role)
	}
	return &runtimeutil.UserRoles{
		UserID:    userID,
		Roles:     []string{role},
		UpdatedAt: time.Now(),
	}, nil
}

func (m *mockUserRoleProvider) RemoveUserRole(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
	if m.removeUserRoleFunc != nil {
		return m.removeUserRoleFunc(ctx, userID, role)
	}
	return &runtimeutil.UserRoles{
		UserID:    userID,
		Roles:     []string{},
		UpdatedAt: time.Now(),
	}, nil
}

// setupRolesRouter creates a chi router with the roles handler for testing.
func setupRolesRouter(provider runtimeutil.UserRoleProvider) *chi.Mux {
	r := chi.NewRouter()
	handler := NewRolesHandler(provider, nil)

	r.Route("/admin/users/{id}/roles", func(r chi.Router) {
		r.Get("/", handler.GetUserRoles)
		r.Post("/", handler.SetUserRoles)
		r.Post("/add", handler.AddUserRole)
		r.Post("/remove", handler.RemoveUserRole)
	})

	return r
}

// withAdminClaims adds admin claims to the request context.
func withAdminClaims(r *http.Request) *http.Request {
	claims := ctxutil.Claims{
		UserID: "admin-user-id",
		Roles:  []string{"admin"},
	}
	ctx := ctxutil.NewClaimsContext(r.Context(), claims)
	return r.WithContext(ctx)
}

func TestRolesHandler_GetUserRoles(t *testing.T) {
	t.Parallel()

	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("returns user roles", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			getUserRolesFunc: func(ctx context.Context, userID string) (*runtimeutil.UserRoles, error) {
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     []string{"admin", "user"},
					UpdatedAt: time.Date(2025, 12, 14, 23, 0, 0, 0, time.UTC),
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		req := httptest.NewRequest(http.MethodGet, "/admin/users/"+validUUID+"/roles/", nil)
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, validUUID, data["user_id"])
		roles := data["roles"].([]interface{})
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, "admin")
		assert.Contains(t, roles, "user")
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		req := httptest.NewRequest(http.MethodGet, "/admin/users/invalid-uuid/roles/", nil)
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp["success"].(bool))
		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "ERR_BAD_REQUEST", errData["code"])
		assert.Equal(t, "Invalid user ID format", errData["message"])
	})

	t.Run("returns empty roles for unknown user", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			getUserRolesFunc: func(ctx context.Context, userID string) (*runtimeutil.UserRoles, error) {
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     []string{},
					UpdatedAt: time.Time{},
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		req := httptest.NewRequest(http.MethodGet, "/admin/users/"+validUUID+"/roles/", nil)
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		data := resp["data"].(map[string]interface{})
		roles := data["roles"].([]interface{})
		assert.Empty(t, roles)
	})
}

func TestRolesHandler_SetUserRoles(t *testing.T) {
	t.Parallel()

	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("sets user roles", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			setUserRolesFunc: func(ctx context.Context, userID string, roles []string) (*runtimeutil.UserRoles, error) {
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     roles,
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		body := `{"roles": ["admin", "user"]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		roles := data["roles"].([]interface{})
		assert.Len(t, roles, 2)
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		body := `{"roles": ["admin"]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/not-a-uuid/roles/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 400 for invalid request body", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		body := `{invalid json}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "Invalid request body", errData["message"])
	})

	t.Run("returns 400 for invalid role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			setUserRolesFunc: func(ctx context.Context, userID string, roles []string) (*runtimeutil.UserRoles, error) {
				return nil, runtimeutil.ErrInvalidRole
			},
		}
		router := setupRolesRouter(provider)

		body := `{"roles": ["invalid_role"]}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "Invalid role name", errData["message"])
	})
}

func TestRolesHandler_AddUserRole(t *testing.T) {
	t.Parallel()

	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("adds role to user", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			addUserRoleFunc: func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     []string{"user", role},
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/add", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		roles := data["roles"].([]interface{})
		assert.Contains(t, roles, "admin")
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/invalid/roles/add", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 400 for empty role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		body := `{"role": ""}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/add", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "Role is required", errData["message"])
	})

	t.Run("returns 400 for invalid role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			addUserRoleFunc: func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
				return nil, runtimeutil.ErrInvalidRole
			},
		}
		router := setupRolesRouter(provider)

		body := `{"role": "superuser"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/add", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("is idempotent for existing role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			addUserRoleFunc: func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
				// Already has admin, returns same state
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     []string{"admin", "user"},
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/add", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRolesHandler_RemoveUserRole(t *testing.T) {
	t.Parallel()

	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	t.Run("removes role from user", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			removeUserRoleFunc: func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     []string{"user"},
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/remove", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.True(t, resp["success"].(bool))
		data := resp["data"].(map[string]interface{})
		roles := data["roles"].([]interface{})
		assert.NotContains(t, roles, "admin")
		assert.Contains(t, roles, "user")
	})

	t.Run("returns 400 for invalid UUID", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/bad-id/roles/remove", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 400 for empty role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{}
		router := setupRolesRouter(provider)

		body := `{"role": ""}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/remove", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "Role is required", errData["message"])
	})

	t.Run("returns 400 for invalid role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			removeUserRoleFunc: func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
				return nil, runtimeutil.ErrInvalidRole
			},
		}
		router := setupRolesRouter(provider)

		body := `{"role": "superuser"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/remove", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusBadRequest, rr.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)

		errData := resp["error"].(map[string]interface{})
		assert.Equal(t, "Invalid role name", errData["message"])
	})

	t.Run("is idempotent for non-existent role", func(t *testing.T) {
		// Arrange
		provider := &mockUserRoleProvider{
			removeUserRoleFunc: func(ctx context.Context, userID string, role string) (*runtimeutil.UserRoles, error) {
				// Role doesn't exist, returns same state
				return &runtimeutil.UserRoles{
					UserID:    userID,
					Roles:     []string{"user"},
					UpdatedAt: time.Now(),
				}, nil
			},
		}
		router := setupRolesRouter(provider)

		body := `{"role": "admin"}`
		req := httptest.NewRequest(http.MethodPost, "/admin/users/"+validUUID+"/roles/remove", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req = withAdminClaims(req)
		rr := httptest.NewRecorder()

		// Act
		router.ServeHTTP(rr, req)

		// Assert
		require.Equal(t, http.StatusOK, rr.Code)
	})
}
