package middleware

import (
	"context"
	"errors"
	"net/http"

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
	ErrNoClaimsInContext = errors.New("no claims in context")

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

// Claims represents authenticated user information.
// This struct is stored in the request context after successful authentication.
type Claims struct {
	UserID      string            `json:"user_id"`
	Roles       []string          `json:"roles"`
	Permissions []string          `json:"permissions"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// HasRole checks if the claims include the specified role.
func (c Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the claims include the specified permission.
func (c Claims) HasPermission(perm string) bool {
	for _, p := range c.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// contextKey is an unexported type for context keys to prevent collisions.
type contextKey string

const claimsKey contextKey = "auth_claims"

// NewContext returns a new context with the given claims.
func NewContext(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// FromContext extracts claims from context.
// Returns ErrNoClaimsInContext if claims are not present.
func FromContext(ctx context.Context) (Claims, error) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return Claims{}, ErrNoClaimsInContext
	}
	return claims, nil
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

			// Store claims in context
			ctx := NewContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
