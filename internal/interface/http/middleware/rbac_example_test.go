package middleware_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/auth"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
)

// rbacMockAuthenticator is a simple authenticator for RBAC examples
type rbacMockAuthenticator struct {
	claims middleware.Claims
}

func (m *rbacMockAuthenticator) Authenticate(r *http.Request) (middleware.Claims, error) {
	return m.claims, nil
}

// ExampleRequireRole demonstrates using RequireRole middleware with a chi router.
// This example shows how to protect an admin-only endpoint.
func ExampleRequireRole() {
	// Create a mock authenticator with admin role
	mockAuth := &rbacMockAuthenticator{
		claims: middleware.Claims{
			UserID: "admin-user-123",
			Roles:  []string{string(auth.RoleAdmin)},
		},
	}

	// Create chi router
	r := chi.NewRouter()

	// Protected admin-only route
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(mockAuth))
		r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
		r.Delete("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "User deleted")
		})
	})

	// Test the endpoint
	req := httptest.NewRequest(http.MethodDelete, "/users/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	// Output: Status: 200
}

// ExampleRequireRole_multipleRoles demonstrates allowing multiple roles.
// This example shows how to allow either admin or service roles.
func ExampleRequireRole_multipleRoles() {
	// Create a mock authenticator with service role
	mockAuth := &rbacMockAuthenticator{
		claims: middleware.Claims{
			UserID: "service-account",
			Roles:  []string{string(auth.RoleService)},
		},
	}

	r := chi.NewRouter()

	// Allow either admin or service roles
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(mockAuth))
		r.Use(middleware.RequireRole(string(auth.RoleAdmin), string(auth.RoleService)))
		r.Get("/internal/metrics", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Metrics data")
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/internal/metrics", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	// Output: Status: 200
}

// ExampleRequirePermission demonstrates requiring ALL specified permissions.
// This example shows how to require both read and write permissions (AND logic).
func ExampleRequirePermission() {
	// Create a mock authenticator with multiple permissions
	mockAuth := &rbacMockAuthenticator{
		claims: middleware.Claims{
			UserID:      "editor-user",
			Permissions: []string{string(auth.PermNoteRead), string(auth.PermNoteUpdate)},
		},
	}

	r := chi.NewRouter()

	// Require BOTH read and update permissions (AND logic)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(mockAuth))
		r.Use(middleware.RequirePermission(
			string(auth.PermNoteRead),
			string(auth.PermNoteUpdate),
		))
		r.Put("/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Note updated")
		})
	})

	req := httptest.NewRequest(http.MethodPut, "/notes/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	// Output: Status: 200
}

// ExampleRequireAnyPermission demonstrates requiring ANY of the specified permissions.
// This example shows how to allow users with either update or delete permissions (OR logic).
func ExampleRequireAnyPermission() {
	// Create a mock authenticator with only delete permission
	mockAuth := &rbacMockAuthenticator{
		claims: middleware.Claims{
			UserID:      "moderator-user",
			Permissions: []string{string(auth.PermNoteDelete)},
		},
	}

	r := chi.NewRouter()

	// Require ANY of update or delete permissions (OR logic)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(mockAuth))
		r.Use(middleware.RequireAnyPermission(
			string(auth.PermNoteUpdate),
			string(auth.PermNoteDelete),
		))
		r.Patch("/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "Note modified")
		})
	})

	req := httptest.NewRequest(http.MethodPatch, "/notes/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	// Output: Status: 200
}

// ExampleRequireRole_combinedMiddleware demonstrates a complete auth + RBAC chain.
// This example shows the recommended pattern for protected endpoints.
func ExampleRequireRole_combinedMiddleware() {
	// Create authenticator with user having admin role and multiple permissions
	mockAuth := &rbacMockAuthenticator{
		claims: middleware.Claims{
			UserID:      "super-admin",
			Roles:       []string{string(auth.RoleAdmin), string(auth.RoleUser)},
			Permissions: []string{string(auth.PermNoteCreate), string(auth.PermNoteDelete)},
			Metadata:    map[string]string{"department": "engineering"},
		},
	}

	r := chi.NewRouter()

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	// Protected API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Apply auth middleware to all API routes
		r.Use(middleware.AuthMiddleware(mockAuth))

		// User-accessible routes (any authenticated user)
		r.Get("/notes", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "List notes")
		})

		// Admin-only routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole(string(auth.RoleAdmin)))
			r.Delete("/notes/{id}", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, "Note deleted by admin")
			})
		})
	})

	// Test admin endpoint
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/notes/456", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	fmt.Printf("Status: %d\n", rec.Code)
	// Output: Status: 200
}
