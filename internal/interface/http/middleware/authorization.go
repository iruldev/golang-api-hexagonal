package middleware

import (
	"log"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// RequireRole returns middleware that restricts access to users with at least one of the required roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := FromContext(r.Context())
			if err != nil {
				// Claims missing means AuthMiddleware was skipped or configured incorrectly
				// This is a server error (misconfiguration), not a client error
				log.Printf("middleware: RequireRole failed, no claims in context: %v", err)
				response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER", "Internal Server Error")
				return
			}

			// Check if user has any of the required roles
			for _, role := range roles {
				if claims.HasRole(role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Unauthorized
			response.Error(w, http.StatusForbidden, "ERR_INSUFFICIENT_ROLE", "Insufficient role")
		})
	}
}

// RequirePermission returns middleware that restricts access to users with ALL of the required permissions.
func RequirePermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := FromContext(r.Context())
			if err != nil {
				// Claims missing means AuthMiddleware was skipped or configured incorrectly
				log.Printf("middleware: RequirePermission failed, no claims in context: %v", err)
				response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER", "Internal Server Error")
				return
			}

			// Check if user has ALL of the required permissions
			for _, perm := range perms {
				if !claims.HasPermission(perm) {
					response.Error(w, http.StatusForbidden, "ERR_INSUFFICIENT_PERMISSION", "Insufficient permission")
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that restricts access to users with at least one of the required permissions.
func RequireAnyPermission(perms ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := FromContext(r.Context())
			if err != nil {
				// Claims missing means AuthMiddleware was skipped or configured incorrectly
				log.Printf("middleware: RequireAnyPermission failed, no claims in context: %v", err)
				response.Error(w, http.StatusInternalServerError, "ERR_INTERNAL_SERVER", "Internal Server Error")
				return
			}

			// Check if user has ANY of the required permissions
			for _, perm := range perms {
				if claims.HasPermission(perm) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Unauthorized
			response.Error(w, http.StatusForbidden, "ERR_INSUFFICIENT_PERMISSION", "Insufficient permission")
		})
	}
}
