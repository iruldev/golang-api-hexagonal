// Package admin provides HTTP handlers for administrative endpoints.
//
// Admin handlers are mounted under /admin prefix (not /api/v1) to clearly
// separate administrative functions from the versioned API.
//
// Security: All handlers in this package require:
// - Authentication (via AuthMiddleware)
// - Admin role (via RequireRole("admin"))
//
// These middleware are applied at the route group level in router.go.
package admin

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// HealthHandler provides admin-specific health information.
// This endpoint validates that admin access is working correctly.
//
// Response:
//
//	{
//	  "success": true,
//	  "data": {
//	    "status": "ok",
//	    "admin_access": true
//	  }
//	}
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"status":       "ok",
		"admin_access": true,
	}
	response.Success(w, data)
}
