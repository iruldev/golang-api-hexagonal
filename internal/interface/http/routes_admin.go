// Package http provides HTTP server and routing functionality.
package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/admin"
)

// RegisterAdminRoutes registers all Admin API routes under the /admin prefix.
//
// # Admin Route Pattern (Story 15.1)
//
// Admin routes are mounted at root level (/admin) NOT under /api/v1 to clearly
// separate administrative endpoints from the versioned API.
//
// # Security Requirements
//
// Admin routes MUST have both:
// 1. AuthMiddleware - validates user identity (returns 401 if missing)
// 2. RequireRole("admin") - validates admin access (returns 403 if unauthorized)
//
// These middleware are applied in router.go BEFORE this function is called.
//
// # Adding New Admin Endpoints
//
// 1. Create handler in internal/interface/http/admin/
// 2. Add route registration here using r.Method("/path", admin.YourHandler)
// 3. Authentication and RBAC are applied automatically via middleware
//
// # Example Usage
//
//	func RegisterAdminRoutes(r chi.Router) {
//	    r.Get("/health", admin.HealthHandler)           // GET /admin/health
//	    r.Get("/users", admin.ListUsersHandler)         // GET /admin/users
//	    r.Post("/users/{id}/ban", admin.BanUserHandler) // POST /admin/users/{id}/ban
//	}
func RegisterAdminRoutes(r chi.Router) {
	// Admin health check - validates admin access is working (Story 15.1)
	r.Get("/health", admin.HealthHandler)

	// -------------------------------------------------------------------------
	// Add new admin routes below this line
	// -------------------------------------------------------------------------

}
