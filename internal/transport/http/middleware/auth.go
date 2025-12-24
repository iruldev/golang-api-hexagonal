// Package middleware provides HTTP middleware for the transport layer.
// This file implements JWT authentication middleware with deterministic time handling.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// validatedClaimsKey marks that claims have been validated by JWTAuth.
type validatedClaimsKey struct{}

func setValidatedClaims(ctx context.Context) context.Context {
	return context.WithValue(ctx, validatedClaimsKey{}, true)
}

func isClaimsValidated(ctx context.Context) bool {
	val, ok := ctx.Value(validatedClaimsKey{}).(bool)
	return ok && val
}

// AllowedAlgorithm is the only allowed JWT signing method (HS256).
const AllowedAlgorithm = "HS256"

// NormalizeRole lowercases and trims a role string for consistent comparisons.
func NormalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

// JWTAuthConfig holds configuration for JWTAuth middleware.
type JWTAuthConfig struct {
	// Secret is the key used for JWT signature validation (HS256).
	Secret []byte
	// Logger for authentication events.
	Logger *slog.Logger
	// Now provides the current time for token validation.
	Now func() time.Time
	// Issuer is the expected issuer claim. If non-empty, tokens must match.
	Issuer string
	// Audience is the expected audience claim. If non-empty, tokens must match.
	Audience string
	// ClockSkew is the tolerance for expired tokens.
	ClockSkew time.Duration
}

// Note: test-only helpers for marking claims validated live in _test.go files.
// Production must rely on JWTAuth + AuthContextBridge to set validated claims.

// JWTAuth returns middleware that validates JWT tokens from the Authorization header.
//
// Security considerations:
//   - Only HS256 algorithm is accepted (prevents algorithm confusion attacks)
//   - Expiration (exp) claim is required
//   - Issuer (iss) and Audience (aud) are validated if configured
//   - Clock skew tolerance is configurable
//   - No error details are exposed in responses (prevents enumeration/timing attacks)
//   - Token must be in Authorization: Bearer <token> format
func JWTAuth(cfg JWTAuthConfig) func(http.Handler) http.Handler {
	// Build parser options
	parserOptions := []jwt.ParserOption{
		jwt.WithValidMethods([]string{AllowedAlgorithm}),
		jwt.WithExpirationRequired(), // AC #1: Reject tokens without exp
		jwt.WithTimeFunc(cfg.Now),
	}

	// Add issuer validation if configured
	if cfg.Issuer != "" {
		parserOptions = append(parserOptions, jwt.WithIssuer(cfg.Issuer))
	}

	// Add audience validation if configured
	if cfg.Audience != "" {
		parserOptions = append(parserOptions, jwt.WithAudience(cfg.Audience))
	}

	// Add clock skew tolerance if configured
	if cfg.ClockSkew > 0 {
		parserOptions = append(parserOptions, jwt.WithLeeway(cfg.ClockSkew))
	}

	parser := jwt.NewParser(parserOptions...)
	logger := cfg.Logger

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header (AC #1)
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.DebugContext(r.Context(), "auth failed: missing specific authentication header")
				writeUnauthorized(w, r)
				return
			}

			// Validate Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				logger.WarnContext(r.Context(), "auth failed: invalid header format", "header", authHeader)
				writeUnauthorized(w, r)
				return
			}

			tokenString := parts[1]

			// Parse and validate token
			claims := &ctxutil.Claims{}
			token, err := parser.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (interface{}, error) {
				return cfg.Secret, nil
			})

			if err != nil || !token.Valid {
				// Return 401 for any validation failure (malformed, wrong signature, expired, wrong iss/aud)
				// Do NOT expose the specific reason for failure (security requirement)
				logger.WarnContext(r.Context(), "auth failed: invalid token", "error", err)
				writeUnauthorized(w, r)
				return
			}

			// Store claims in context for downstream handlers
			if strings.TrimSpace(claims.Subject) == "" {
				logger.WarnContext(r.Context(), "auth failed: empty subject in claims")
				writeUnauthorized(w, r)
				return
			}

			normalizedRole := NormalizeRole(claims.Role)
			claims.Role = normalizedRole

			// Only JWTAuth should mark claims validated in production.
			ctx := ctxutil.SetClaims(r.Context(), claims)
			ctx = setValidatedClaims(ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeUnauthorized writes an RFC 7807 error response for authentication failures.
// It intentionally provides no detail about why authentication failed (security requirement).
func writeUnauthorized(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	contract.WriteProblemJSON(w, r, &app.AppError{
		Op:      "JWTAuth",
		Code:    app.CodeUnauthorized,
		Message: "Unauthorized",
	})
}
