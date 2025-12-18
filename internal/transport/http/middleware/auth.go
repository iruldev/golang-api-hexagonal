// Package middleware provides HTTP middleware for the transport layer.
// This file implements JWT authentication middleware with deterministic time handling.
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

// JWTAuth returns middleware that validates JWT tokens from the Authorization header.
// The now function is injected for deterministic time testing (AC #3).
//
// Security considerations:
//   - Only HS256 algorithm is accepted (AC #7, prevents algorithm confusion attacks)
//   - No error details are exposed in responses (prevents enumeration/timing attacks)
//   - Token must be in Authorization: Bearer <token> format
func JWTAuth(secret []byte, now func() time.Time) func(http.Handler) http.Handler {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithTimeFunc(now), // Inject time function for exp/nbf validation (AC #3)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header (AC #1)
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, r)
				return
			}

			// Validate Bearer token format
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeUnauthorized(w, r)
				return
			}

			tokenString := parts[1]

			// Parse and validate token (AC #2, #4, #7)
			claims := &ctxutil.Claims{}
			token, err := parser.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (interface{}, error) {
				return secret, nil
			})

			if err != nil || !token.Valid {
				// AC #2, #4: Return 401 for any validation failure (malformed, wrong signature, expired)
				// Do NOT expose the specific reason for failure (security requirement)
				writeUnauthorized(w, r)
				return
			}

			// Store claims in context for downstream handlers (AC #5)
			ctx := ctxutil.SetClaims(r.Context(), claims)
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
