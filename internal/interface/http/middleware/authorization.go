package middleware

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// RequireRole returns middleware that restricts access to users with at least one of the required roles.
// Uses structured logging and central error codes for consistency with AuthMiddleware.
func RequireRole(roles []string, logger observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := ctxutil.RequestIDFromContext(r.Context())
			claims, err := ctxutil.ClaimsFromContext(r.Context())
			if err != nil {
				// ctxutil.Claims missing means AuthMiddleware was skipped or configured incorrectly
				// This is a server error (misconfiguration), not a client error
				logger.Error("middleware: RequireRole failed, no claims in context",
					observability.Err(err),
					observability.String("trace_id", traceID),
					observability.String("path", r.URL.Path),
					observability.String("method", r.Method),
				)
				domainErr := domainerrors.NewDomain(domainerrors.CodeInternalError, "Internal Server Error")
				response.HandleErrorCtx(w, r.Context(), domainErr)
				return
			}

			// Check if user has any of the required roles
			for _, role := range roles {
				if claims.HasRole(role) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Forbidden - user authenticated but lacks required role
			domainErr := domainerrors.NewDomain(domainerrors.CodeForbidden, "Insufficient role")
			response.HandleErrorCtx(w, r.Context(), domainErr)
		})
	}
}

// RequirePermission returns middleware that restricts access to users with ALL of the required permissions.
// Uses structured logging and central error codes for consistency with AuthMiddleware.
func RequirePermission(perms []string, logger observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := ctxutil.RequestIDFromContext(r.Context())
			claims, err := ctxutil.ClaimsFromContext(r.Context())
			if err != nil {
				// ctxutil.Claims missing means AuthMiddleware was skipped or configured incorrectly
				logger.Error("middleware: RequirePermission failed, no claims in context",
					observability.Err(err),
					observability.String("trace_id", traceID),
					observability.String("path", r.URL.Path),
					observability.String("method", r.Method),
				)
				domainErr := domainerrors.NewDomain(domainerrors.CodeInternalError, "Internal Server Error")
				response.HandleErrorCtx(w, r.Context(), domainErr)
				return
			}

			// Check if user has ALL of the required permissions
			for _, perm := range perms {
				if !claims.HasPermission(perm) {
					domainErr := domainerrors.NewDomain(domainerrors.CodeForbidden, "Insufficient permission")
					response.HandleErrorCtx(w, r.Context(), domainErr)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission returns middleware that restricts access to users with at least one of the required permissions.
// Uses structured logging and central error codes for consistency with AuthMiddleware.
func RequireAnyPermission(perms []string, logger observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := ctxutil.RequestIDFromContext(r.Context())
			claims, err := ctxutil.ClaimsFromContext(r.Context())
			if err != nil {
				// ctxutil.Claims missing means AuthMiddleware was skipped or configured incorrectly
				logger.Error("middleware: RequireAnyPermission failed, no claims in context",
					observability.Err(err),
					observability.String("trace_id", traceID),
					observability.String("path", r.URL.Path),
					observability.String("method", r.Method),
				)
				domainErr := domainerrors.NewDomain(domainerrors.CodeInternalError, "Internal Server Error")
				response.HandleErrorCtx(w, r.Context(), domainErr)
				return
			}

			// Check if user has ANY of the required permissions
			for _, perm := range perms {
				if claims.HasPermission(perm) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Forbidden - user authenticated but lacks required permission
			domainErr := domainerrors.NewDomain(domainerrors.CodeForbidden, "Insufficient permission")
			response.HandleErrorCtx(w, r.Context(), domainErr)
		})
	}
}
