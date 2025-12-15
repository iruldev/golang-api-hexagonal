package middleware

import (
	"errors"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/ctxutil"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/request"
	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/response"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// Sentinel errors for authentication failures.
var (
	// ErrUnauthenticated indicates authentication failed (invalid credentials).
	ErrUnauthenticated = errors.New("unauthenticated")

	// ErrTokenExpired indicates the token has expired.
	ErrTokenExpired = errors.New("token expired")

	// ErrTokenInvalid indicates the token format or signature is invalid.
	ErrTokenInvalid = errors.New("token invalid")
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
//	func (a *JWTAuthenticator) Authenticate(r *http.Request) (ctxutil.Claims, error) {
//	    token := r.Header.Get("Authorization")
//	    // Validate token and extract claims...
//	    return claims, nil
//	}
type Authenticator interface {
	// Authenticate validates credentials from the request and returns claims.
	// Returns ErrUnauthenticated if authentication fails.
	// Returns ErrTokenExpired if token is valid but expired.
	// Returns ErrTokenInvalid if token format/signature is invalid.
	Authenticate(r *http.Request) (ctxutil.Claims, error)
}

// AuthMiddleware creates authentication middleware using the provided Authenticator.
// Authenticated claims are stored in the request context.
// Error responses use the Envelope format with trace_id from Story 2.1.
//
// Usage:
//
//	router.Group(func(r chi.Router) {
//	    r.Use(middleware.AuthMiddleware(jwtAuth, logger, false))
//	    r.Get("/api/v1/notes", noteHandler.List)
//	})
func AuthMiddleware(auth Authenticator, logger observability.Logger, trustProxyHeaders bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if auth == nil {
			panic("AuthMiddleware: authenticator cannot be nil")
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := auth.Authenticate(r)
			if err != nil {
				var domainErr error
				// Map sentinel errors to DomainErrors for consistent handling
				switch {
				case errors.Is(err, ErrTokenExpired):
					logger.Warn("authentication failed: token expired",
						observability.Err(err),
						observability.String("ip", request.GetRealIP(r, trustProxyHeaders)),
						observability.String("trace_id", ctxutil.RequestIDFromContext(r.Context())),
					)
					domainErr = domainerrors.NewDomain(domainerrors.CodeTokenExpired, "Token has expired")
				case errors.Is(err, ErrTokenInvalid):
					logger.Warn("authentication failed: token invalid",
						observability.Err(err),
						observability.String("ip", request.GetRealIP(r, trustProxyHeaders)),
						observability.String("trace_id", ctxutil.RequestIDFromContext(r.Context())),
					)
					domainErr = domainerrors.NewDomain(domainerrors.CodeTokenInvalid, "Invalid token")
				case errors.Is(err, ErrUnauthenticated):
					// Don't log "unauthenticated" as warning if it's just missing header (common),
					// but DO log if it's malformed or other unauthenticated reason if needed.
					// For now, let's keep it Info or Debug, or just leave it.
					// The review said "Missing Security Logs: Failed authentication attempts... malformed tokens".
					// Unauthenticated usually means missing header. Let's log it as Info to not spam.
					// But review explicitly asked for logs. Let's use Warn for Invalid/Expired, and Info for Unauthenticated.
					logger.Info("authentication missing or failed",
						observability.String("ip", request.GetRealIP(r, trustProxyHeaders)),
						observability.String("trace_id", ctxutil.RequestIDFromContext(r.Context())),
					)
					domainErr = domainerrors.NewDomain(domainerrors.CodeUnauthorized, "Authentication required")
				default:
					// Log unexpected errors for observability using structured logger
					logger.Error("unexpected authentication error",
						observability.Err(err),
						observability.String("path", r.URL.Path),
						observability.String("method", r.Method),
						observability.String("ip", request.GetRealIP(r, trustProxyHeaders)),
						observability.String("trace_id", ctxutil.RequestIDFromContext(r.Context())),
					)
					domainErr = domainerrors.NewDomainWithCause(domainerrors.CodeUnauthorized, "Authentication required", err)
				}

				// Delegate to response package for consistent mapping and formatting
				response.HandleErrorCtx(w, r.Context(), domainErr)
				return
			}

			// Store claims in context using ctxutil
			ctx := ctxutil.NewClaimsContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
