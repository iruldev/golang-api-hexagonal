// Package middleware provides HTTP middleware for cross-cutting concerns.
package middleware

import (
	"fmt"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

const (
	logKeyUserID        = "user_id"
	logKeyPath          = "path"
	logKeyMethod        = "method"
	logKeyRequiredRoles = "required_roles"
	logKeyRequiredPerms = "required_perms"
)

// RequireRole returns middleware that restricts access to users with the specific role.
func RequireRole(logger observability.Logger, role string) func(http.Handler) http.Handler {
	return RequireAnyRole(logger, role)
}

// RequireAnyRole returns middleware that restricts access to users with at least one of the required roles.
// Note: logger is first argument to allow variadic roles.
func RequireAnyRole(logger observability.Logger, roles ...string) func(http.Handler) http.Handler {
	return enforceConstraint(
		logger,
		func(claims ctxutil.Claims) bool {
			for _, role := range roles {
				if claims.HasRole(role) {
					return true
				}
			}
			return false
		},
		func() []observability.Field {
			return []observability.Field{
				observability.Any(logKeyRequiredRoles, roles),
			}
		},
		domainerrors.CodeInsufficientRole,
		"Insufficient role",
		"RequireAnyRole",
	)
}

// RequirePermission returns middleware that restricts access to users with the specific permission.
func RequirePermission(logger observability.Logger, perm string) func(http.Handler) http.Handler {
	return RequireAnyPermission(logger, perm)
}

// RequireAnyPermission returns middleware that restricts access to users with at least one of the required permissions.
func RequireAnyPermission(logger observability.Logger, perms ...string) func(http.Handler) http.Handler {
	return enforceConstraint(
		logger,
		func(claims ctxutil.Claims) bool {
			for _, perm := range perms {
				if claims.HasPermission(perm) {
					return true
				}
			}
			return false
		},
		func() []observability.Field {
			return []observability.Field{
				observability.Any(logKeyRequiredPerms, perms),
			}
		},
		domainerrors.CodeInsufficientPermission,
		"Insufficient permission",
		"RequireAnyPermission",
	)
}

// enforceConstraint is a helper to reduce duplication in RBAC middleware.
func enforceConstraint(
	logger observability.Logger,
	check func(claims ctxutil.Claims) bool,
	logFields func() []observability.Field,
	errorCode string,
	errorMessage string,
	middlewareName string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := ctxutil.RequestIDFromContext(r.Context())
			claims, err := ctxutil.ClaimsFromContext(r.Context())
			if err != nil {
				logger.Error(fmt.Sprintf("middleware: %s failed, no claims in context", middlewareName),
					observability.Err(err),
					observability.String(observability.LogKeyTraceID, traceID),
					observability.String(logKeyPath, r.URL.Path),
					observability.String(logKeyMethod, r.Method),
				)
				domainErr := domainerrors.NewDomain(domainerrors.CodeInternalError, "Internal Server Error")
				response.HandleErrorCtx(w, r.Context(), domainErr)
				return
			}

			if check(claims) {
				next.ServeHTTP(w, r)
				return
			}

			// Log access denied with context-specific fields
			fields := []observability.Field{
				observability.String(observability.LogKeyTraceID, traceID),
				observability.String(logKeyUserID, claims.UserID),
				observability.String(logKeyPath, r.URL.Path),
				observability.String(logKeyMethod, r.Method),
			}
			fields = append(fields, logFields()...)

			logger.Warn(fmt.Sprintf("middleware: %s access denied", middlewareName), fields...)

			domainErr := domainerrors.NewDomain(errorCode, errorMessage)
			response.HandleErrorCtx(w, r.Context(), domainErr)
		})
	}
}
