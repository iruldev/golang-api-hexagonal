package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// errorResponse matches the response.Error output format
type errorResponse struct {
	Success bool `json:"success"`
	Error   struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		userRoles      []string
		requiredRoles  []string
		wantStatusCode int
		wantErrCode    string
	}{
		{
			name:           "user has required role",
			userRoles:      []string{"admin"},
			requiredRoles:  []string{"admin"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "user has one of required roles",
			userRoles:      []string{"service"},
			requiredRoles:  []string{"admin", "service"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "user has multiple roles including required",
			userRoles:      []string{"user", "service"},
			requiredRoles:  []string{"service"},
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "user lacks required role",
			userRoles:      []string{"user"},
			requiredRoles:  []string{"admin"},
			wantStatusCode: http.StatusForbidden,
			wantErrCode:    "ERR_FORBIDDEN",
		},
		{
			name:           "empty user roles",
			userRoles:      []string{},
			requiredRoles:  []string{"admin"},
			wantStatusCode: http.StatusForbidden,
			wantErrCode:    "ERR_FORBIDDEN",
		},
		{
			name:           "nil user roles",
			userRoles:      nil,
			requiredRoles:  []string{"admin"},
			wantStatusCode: http.StatusForbidden,
			wantErrCode:    "ERR_FORBIDDEN",
		},
		{
			name:           "user lacks all required roles",
			userRoles:      []string{"user"},
			requiredRoles:  []string{"admin", "service"},
			wantStatusCode: http.StatusForbidden,
			wantErrCode:    "ERR_FORBIDDEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create request with claims in context
			claims := Claims{UserID: "test-user", Roles: tt.userRoles}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(NewContext(req.Context(), claims))
			rec := httptest.NewRecorder()

			// Create chain: RequireRole -> handler
			handler := RequireRole(tt.requiredRoles...)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)

			// Act
			handler.ServeHTTP(rec, req)

			// Assert status code
			if rec.Code != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			// Assert error response if expected
			if tt.wantErrCode != "" {
				var resp errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if resp.Error.Code != tt.wantErrCode {
					t.Errorf("ErrorCode = %v, want %v", resp.Error.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestRequireRole_NoClaimsInContext(t *testing.T) {
	// Arrange: Create request WITHOUT claims in context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusForbidden {
		t.Errorf("StatusCode = %v, want %v", rec.Code, http.StatusForbidden)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Error.Code != "ERR_FORBIDDEN" {
		t.Errorf("ErrorCode = %v, want ERR_FORBIDDEN", resp.Error.Code)
	}
	if resp.Error.Message != "Access denied" {
		t.Errorf("ErrorMessage = %v, want 'Access denied'", resp.Error.Message)
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name            string
		userPermissions []string
		requiredPerms   []string
		wantStatusCode  int
		wantErrCode     string
	}{
		{
			name:            "user has required permission",
			userPermissions: []string{"note:create"},
			requiredPerms:   []string{"note:create"},
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "user has all required permissions (AND logic)",
			userPermissions: []string{"note:create", "note:read", "note:delete"},
			requiredPerms:   []string{"note:create", "note:read"},
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "user lacks one required permission (AND logic fails)",
			userPermissions: []string{"note:create"},
			requiredPerms:   []string{"note:create", "note:delete"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "user lacks required permission",
			userPermissions: []string{"note:read"},
			requiredPerms:   []string{"note:delete"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "empty user permissions",
			userPermissions: []string{},
			requiredPerms:   []string{"note:create"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "nil user permissions",
			userPermissions: nil,
			requiredPerms:   []string{"note:create"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "no permissions required passes",
			userPermissions: []string{},
			requiredPerms:   []string{},
			wantStatusCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create request with claims in context
			claims := Claims{UserID: "test-user", Permissions: tt.userPermissions}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(NewContext(req.Context(), claims))
			rec := httptest.NewRecorder()

			// Create chain: RequirePermission -> handler
			handler := RequirePermission(tt.requiredPerms...)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)

			// Act
			handler.ServeHTTP(rec, req)

			// Assert status code
			if rec.Code != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			// Assert error response if expected
			if tt.wantErrCode != "" {
				var resp errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if resp.Error.Code != tt.wantErrCode {
					t.Errorf("ErrorCode = %v, want %v", resp.Error.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestRequirePermission_NoClaimsInContext(t *testing.T) {
	// Arrange: Create request WITHOUT claims in context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler := RequirePermission("note:create")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusForbidden {
		t.Errorf("StatusCode = %v, want %v", rec.Code, http.StatusForbidden)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Error.Code != "ERR_FORBIDDEN" {
		t.Errorf("ErrorCode = %v, want ERR_FORBIDDEN", resp.Error.Code)
	}
}

func TestRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name            string
		userPermissions []string
		requiredPerms   []string
		wantStatusCode  int
		wantErrCode     string
	}{
		{
			name:            "user has required permission",
			userPermissions: []string{"note:create"},
			requiredPerms:   []string{"note:create"},
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "user has one of required permissions (OR logic)",
			userPermissions: []string{"note:delete"},
			requiredPerms:   []string{"note:create", "note:delete"},
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "user has multiple matching permissions",
			userPermissions: []string{"note:create", "note:delete"},
			requiredPerms:   []string{"note:create", "note:delete"},
			wantStatusCode:  http.StatusOK,
		},
		{
			name:            "user lacks all required permissions",
			userPermissions: []string{"note:read"},
			requiredPerms:   []string{"note:create", "note:delete"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "empty user permissions",
			userPermissions: []string{},
			requiredPerms:   []string{"note:create"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "nil user permissions",
			userPermissions: nil,
			requiredPerms:   []string{"note:create"},
			wantStatusCode:  http.StatusForbidden,
			wantErrCode:     "ERR_FORBIDDEN",
		},
		{
			name:            "no permissions required passes (edge case)",
			userPermissions: []string{},
			requiredPerms:   []string{},
			wantStatusCode:  http.StatusForbidden, // No permissions to match, so fails
			wantErrCode:     "ERR_FORBIDDEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange: Create request with claims in context
			claims := Claims{UserID: "test-user", Permissions: tt.userPermissions}
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(NewContext(req.Context(), claims))
			rec := httptest.NewRecorder()

			// Create chain: RequireAnyPermission -> handler
			handler := RequireAnyPermission(tt.requiredPerms...)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			)

			// Act
			handler.ServeHTTP(rec, req)

			// Assert status code
			if rec.Code != tt.wantStatusCode {
				t.Errorf("StatusCode = %v, want %v", rec.Code, tt.wantStatusCode)
			}

			// Assert error response if expected
			if tt.wantErrCode != "" {
				var resp errorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response: %v", err)
				}
				if resp.Error.Code != tt.wantErrCode {
					t.Errorf("ErrorCode = %v, want %v", resp.Error.Code, tt.wantErrCode)
				}
			}
		})
	}
}

func TestRequireAnyPermission_NoClaimsInContext(t *testing.T) {
	// Arrange: Create request WITHOUT claims in context
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler := RequireAnyPermission("note:create")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusForbidden {
		t.Errorf("StatusCode = %v, want %v", rec.Code, http.StatusForbidden)
	}

	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Error.Code != "ERR_FORBIDDEN" {
		t.Errorf("ErrorCode = %v, want ERR_FORBIDDEN", resp.Error.Code)
	}
}

func TestMiddlewareChain_AuthAndRBAC(t *testing.T) {
	// Test that RBAC middleware works correctly after AuthMiddleware
	// Simulates: AuthMiddleware sets claims -> RequireRole checks them

	// Create a mock authenticator that always succeeds
	mockAuth := &mockAuthenticator{
		claims: Claims{
			UserID: "test-user",
			Roles:  []string{"admin"},
		},
	}

	// Create the middleware chain
	handler := AuthMiddleware(mockAuth)(
		RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			}),
		),
	)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Errorf("StatusCode = %v, want %v", rec.Code, http.StatusOK)
	}
}

// Note: mockAuthenticator is defined in auth_test.go and shared across test files

func TestForbiddenResponseFormat(t *testing.T) {
	// Verify the 403 response format matches project pattern
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(NewContext(req.Context(), Claims{UserID: "test", Roles: []string{"user"}}))
	rec := httptest.NewRecorder()

	handler := RequireRole("admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	handler.ServeHTTP(rec, req)

	// Assert status code is 403
	if rec.Code != http.StatusForbidden {
		t.Errorf("StatusCode = %v, want %v", rec.Code, http.StatusForbidden)
	}

	// Assert response body structure
	var resp errorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	if resp.Success != false {
		t.Errorf("response.success = %v, want false", resp.Success)
	}
	if resp.Error.Code != "ERR_FORBIDDEN" {
		t.Errorf("response.error.code = %v, want ERR_FORBIDDEN", resp.Error.Code)
	}
	if resp.Error.Message == "" {
		t.Error("response.error.message should not be empty")
	}
}
