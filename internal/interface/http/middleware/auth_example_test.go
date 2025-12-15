package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// ExampleAuthenticator demonstrates how to implement the Authenticator interface.
// This example shows a simple API key authenticator.
func ExampleAuthenticator() {
	// Define a simple API key authenticator
	type APIKeyAuth struct {
		validKeys map[string]ctxutil.Claims
	}

	apiKeyAuth := &APIKeyAuth{
		validKeys: map[string]ctxutil.Claims{
			"secret-key-123": {
				UserID:      "service-account-1",
				Roles:       []string{"service"},
				Permissions: []string{"read", "write"},
			},
		},
	}

	// The Authenticate method would be implemented like this:
	_ = func(r *http.Request) (ctxutil.Claims, error) {
		key := r.Header.Get("X-API-Key")
		if key == "" {
			return ctxutil.Claims{}, middleware.ErrUnauthenticated
		}
		claims, ok := apiKeyAuth.validKeys[key]
		if !ok {
			return ctxutil.Claims{}, middleware.ErrTokenInvalid
		}
		return claims, nil
	}

	fmt.Println("API key authenticator implemented")
	// Output: API key authenticator implemented
}

// ExampleAuthMiddleware demonstrates how to use AuthMiddleware with a router.
func ExampleAuthMiddleware() {
	// Create a mock authenticator for demonstration
	mockAuth := &mockAuthenticator{
		claims: ctxutil.Claims{
			UserID:      "user-123",
			Roles:       []string{"admin", "user"},
			Permissions: []string{"read", "write", "delete"},
		},
	}

	// Create protected handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := ctxutil.ClaimsFromContext(r.Context())
		fmt.Printf("Authenticated user: %s\n", claims.UserID)
	})

	// Wrap with auth middleware
	protected := middleware.AuthMiddleware(mockAuth, observability.NewNopLoggerInterface(), false)(handler)

	// Test the protected endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)

	// Output: Authenticated user: user-123
}

// ExampleClaimsFromContext demonstrates extracting claims from context in a handler.
func ExampleClaimsFromContext() {
	// In a real handler, you would get context from the request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := ctxutil.ClaimsFromContext(r.Context())
		if err != nil {
			// This shouldn't happen if AuthMiddleware is applied
			http.Error(w, "Internal error", http.StatusInternalServerError)
			return
		}

		// Use claims for authorization
		if claims.HasRole("admin") {
			fmt.Println("User is admin")
		}

		if claims.HasPermission("delete") {
			fmt.Println("User can delete")
		}

		fmt.Printf("User ID: %s\n", claims.UserID)
	})

	// Simulate authenticated request
	mockAuth := &mockAuthenticator{
		claims: ctxutil.Claims{
			UserID:      "user-456",
			Roles:       []string{"admin"},
			Permissions: []string{"delete"},
		},
	}

	protected := middleware.AuthMiddleware(mockAuth, observability.NewNopLoggerInterface(), false)(handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	protected.ServeHTTP(rec, req)

	// Output:
	// User is admin
	// User can delete
	// User ID: user-456
}

// ExampleClaims_HasRole demonstrates checking user roles.
func ExampleClaims_HasRole() {
	claims := ctxutil.Claims{
		UserID: "user-789",
		Roles:  []string{"admin", "editor", "viewer"},
	}

	fmt.Printf("Is admin: %v\n", claims.HasRole("admin"))
	fmt.Printf("Is superuser: %v\n", claims.HasRole("superuser"))

	// Output:
	// Is admin: true
	// Is superuser: false
}

// ExampleClaims_HasPermission demonstrates checking user permissions.
func ExampleClaims_HasPermission() {
	claims := ctxutil.Claims{
		UserID:      "user-789",
		Permissions: []string{"notes:read", "notes:write", "notes:delete"},
	}

	fmt.Printf("Can read notes: %v\n", claims.HasPermission("notes:read"))
	fmt.Printf("Can manage users: %v\n", claims.HasPermission("users:manage"))

	// Output:
	// Can read notes: true
	// Can manage users: false
}

// mockAuthenticator is a test helper that implements Authenticator.
type mockAuthenticator struct {
	claims ctxutil.Claims
	err    error
}

func (m *mockAuthenticator) Authenticate(r *http.Request) (ctxutil.Claims, error) {
	return m.claims, m.err
}
