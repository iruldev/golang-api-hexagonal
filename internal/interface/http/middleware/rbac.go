// Package middleware provides HTTP middleware components for the application.
// This file contains RBAC (Role-Based Access Control) middleware for authorization.
//
// Security Note: Authorization failures return 403 Forbidden without revealing
// which specific role or permission was missing. For audit logging of authorization
// failures, wrap these middlewares with application-level logging or use the HTTP
// logging middleware which captures all 403 responses.
package middleware

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// RequireRole creates middleware that checks if the user has one of the required roles.
// The middleware extracts claims from context (set by AuthMiddleware) and verifies
// that the user has at least one of the specified roles.
//
// Returns 403 Forbidden with ERR_FORBIDDEN code if:
//   - No claims are found in context (user not authenticated)
//   - User does not have any of the required roles
//
// Usage:
//
//	r.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Use(middleware.RequireRole("admin", "service"))
//	    r.Delete("/users/{id}", deleteUserHandler)
//	})
//
// The middleware uses OR logic: user needs at least ONE of the specified roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims from context (set by AuthMiddleware)
			claims, err := FromContext(r.Context())
			if err != nil {
				// No claims in context means user is not authenticated
				// Return 403 Forbidden (not 401) as this is an authorization check
				response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Access denied")
				return
			}

			// Check if user has any of the required roles (OR logic)
			for _, required := range roles {
				if claims.HasRole(required) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// User doesn't have any of the required roles
			// Security: Don't reveal which specific role was missing
			response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient role")
		})
	}
}

// RequirePermission creates middleware that checks if the user has ALL required permissions.
// The middleware extracts claims from context and verifies that the user has every
// specified permission (AND logic).
//
// Returns 403 Forbidden with ERR_FORBIDDEN code if:
//   - No claims are found in context (user not authenticated)
//   - User is missing any of the required permissions
//
// Usage:
//
//	r.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Use(middleware.RequirePermission("note:create", "note:read"))
//	    r.Post("/notes", createNoteHandler)
//	})
//
// The middleware uses AND logic: user needs ALL of the specified permissions.
func RequirePermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims from context (set by AuthMiddleware)
			claims, err := FromContext(r.Context())
			if err != nil {
				// No claims in context means user is not authenticated
				response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Access denied")
				return
			}

			// Check if user has ALL required permissions (AND logic)
			for _, required := range perms {
				if !claims.HasPermission(required) {
					// User is missing at least one required permission
					// Security: Don't reveal which specific permission was missing
					response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient permission")
					return
				}
			}

			// User has all required permissions
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission creates middleware that checks if the user has ANY of the required permissions.
// The middleware extracts claims from context and verifies that the user has at least
// one of the specified permissions (OR logic).
//
// Returns 403 Forbidden with ERR_FORBIDDEN code if:
//   - No claims are found in context (user not authenticated)
//   - User does not have any of the required permissions
//
// Usage:
//
//	r.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Use(middleware.RequireAnyPermission("note:update", "note:delete"))
//	    r.Patch("/notes/{id}", modifyNoteHandler)
//	})
//
// The middleware uses OR logic: user needs at least ONE of the specified permissions.
func RequireAnyPermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract claims from context (set by AuthMiddleware)
			claims, err := FromContext(r.Context())
			if err != nil {
				// No claims in context means user is not authenticated
				response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Access denied")
				return
			}

			// Check if user has ANY of the required permissions (OR logic)
			for _, required := range perms {
				if claims.HasPermission(required) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// User doesn't have any of the required permissions
			// Security: Don't reveal which specific permission was missing
			response.Error(w, http.StatusForbidden, "ERR_FORBIDDEN", "Insufficient permission")
		})
	}
}
