// Package http provides HTTP server and routing functionality.
package http

import (
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/admin"
	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// AdminDeps holds dependencies for admin routes.
type AdminDeps struct {
	// FeatureFlagProvider for feature flag management (Story 15.2)
	FeatureFlagProvider runtimeutil.AdminFeatureFlagProvider
	// UserRoleProvider for user role management (Story 15.3)
	UserRoleProvider runtimeutil.UserRoleProvider
	// Logger for audit logging
	Logger *zap.Logger
}

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
//	func RegisterAdminRoutes(r chi.Router, deps AdminDeps) {
//	    r.Get("/health", admin.HealthHandler)           // GET /admin/health
//	    r.Get("/users", admin.ListUsersHandler)         // GET /admin/users
//	    r.Post("/users/{id}/ban", admin.BanUserHandler) // POST /admin/users/{id}/ban
//	}
func RegisterAdminRoutes(r chi.Router, deps AdminDeps) {
	// Admin health check - validates admin access is working (Story 15.1)
	r.Get("/health", admin.HealthHandler)

	// -------------------------------------------------------------------------
	// Feature Flag Management (Story 15.2)
	// -------------------------------------------------------------------------
	if deps.FeatureFlagProvider != nil {
		featuresHandler := admin.NewFeaturesHandler(deps.FeatureFlagProvider, deps.Logger)
		r.Get("/features", featuresHandler.ListFlags)
		r.Get("/features/{flag}", featuresHandler.GetFlag)
		r.Post("/features/{flag}/enable", featuresHandler.EnableFlag)
		r.Post("/features/{flag}/disable", featuresHandler.DisableFlag)
	}

	// -------------------------------------------------------------------------
	// User Role Management (Story 15.3)
	// -------------------------------------------------------------------------
	if deps.UserRoleProvider != nil {
		rolesHandler := admin.NewRolesHandler(deps.UserRoleProvider, deps.Logger)
		r.Get("/users/{id}/roles", rolesHandler.GetUserRoles)
		r.Post("/users/{id}/roles", rolesHandler.SetUserRoles)
		r.Post("/users/{id}/roles/add", rolesHandler.AddUserRole)
		r.Post("/users/{id}/roles/remove", rolesHandler.RemoveUserRole)
	}

	// -------------------------------------------------------------------------
	// Add new admin routes below this line
	// -------------------------------------------------------------------------

}
