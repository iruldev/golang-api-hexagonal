package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
)

// Sentinel errors for authentication failures.
var (
	// ErrUnauthenticated indicates authentication failed (invalid credentials).
	ErrUnauthenticated = errors.New("unauthenticated")

	// ErrTokenExpired indicates the token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenInvalid indicates the token format or signature is invalid.
	ErrTokenInvalid = errors.New("token invalid")

	// ErrNoClaimsInContext indicates claims were not found in context.
	// Deprecated: Use ctxutil.ErrNoClaimsInContext instead.
	ErrNoClaimsInContext = ctxutil.ErrNoClaimsInContext

	// ErrForbidden indicates the user lacks permission for the requested resource.
	// This is returned when authorization (not authentication) fails.
	ErrForbidden = errors.New("forbidden")

	// ErrInsufficientRole indicates the user does not have the required role.
	ErrInsufficientRole = errors.New("insufficient role")

	// ErrInsufficientPermission indicates the user does not have the required permission.
	ErrInsufficientPermission = errors.New("insufficient permission")
)

// Authenticator defines the interface for authentication providers.
// Implementations may use JWT, API keys, sessions, or other mechanisms.
//
// Example implementations:
//   - JWTAuthenticator: Validates JWT tokens from Authorization header
//   - APIKeyAuthenticator: Validates API keys from X-API-Key header
//   - SessionAuthenticator: Validates session cookies
//
// Usage:
//
//	type JWTAuthenticator struct {
//	    secretKey []byte
//	}
//
//	func (a *JWTAuthenticator) Authenticate(r *http.Request) (Claims, error) {
//	    token := r.Header.Get("Authorization")
//	    // Validate token and extract claims...
//	    return claims, nil
//	}
type Authenticator interface {
	// Authenticate validates credentials from the request and returns claims.
	// Returns ErrUnauthenticated if authentication fails.
	// Returns ErrTokenExpired if token is valid but expired.
	// Returns ErrTokenInvalid if token format/signature is invalid.
	Authenticate(r *http.Request) (Claims, error)
}

// Claims is an alias for ctxutil.Claims for backwards compatibility.
// New code should use ctxutil.Claims directly.
type Claims = ctxutil.Claims

// NewContext returns a new context with the given claims.
// Deprecated: Use ctxutil.NewClaimsContext instead.
func NewContext(ctx context.Context, claims Claims) context.Context {
	return ctxutil.NewClaimsContext(ctx, claims)
}

// FromContext extracts claims from context.
// Returns ErrNoClaimsInContext if claims are not present.
// Deprecated: Use ctxutil.ClaimsFromContext instead.
func FromContext(ctx context.Context) (Claims, error) {
	return ctxutil.ClaimsFromContext(ctx)
}

// AuthMiddleware creates authentication middleware using the provided Authenticator.
// Authenticated claims are stored in the request context.
//
// Usage:
//
//	router.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth))
//	    r.Get("/api/v1/notes", noteHandler.List)
//	})
func AuthMiddleware(auth Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := auth.Authenticate(r)
			if err != nil {
				// Map error to response with appropriate error code using project's response package
				var errCode, errMsg string
				switch {
				case errors.Is(err, ErrTokenExpired):
					errCode = "ERR_TOKEN_EXPIRED"
					errMsg = "Token has expired"
				case errors.Is(err, ErrTokenInvalid):
					errCode = "ERR_TOKEN_INVALID"
					errMsg = "Invalid token"
				default:
					errCode = "ERR_UNAUTHORIZED"
					errMsg = "Authentication required"
				}

				response.Error(w, http.StatusUnauthorized, errCode, errMsg)
				return
			}

			// Store claims in context using ctxutil
			ctx := ctxutil.NewClaimsContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
