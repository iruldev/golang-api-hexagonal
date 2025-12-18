// Package ctxutil provides context utility functions for storing and retrieving
// request-scoped values such as JWT claims.
package ctxutil

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// claimsKey is the unexported type for the context key to prevent collisions.
type claimsKey struct{}

// Claims represents JWT claims extracted from the token.
// It embeds jwt.RegisteredClaims for standard fields (iss, sub, aud, exp, nbf, iat, jti).
type Claims struct {
	jwt.RegisteredClaims
	// Role for authorization (e.g., "admin", "user")
	Role string `json:"role,omitempty"`
	// Custom claims can be added here as the application evolves.
}

// SetClaims stores the claims in the context.
func SetClaims(ctx context.Context, claims *Claims) context.Context {
	if claims != nil {
		claims.Role = strings.ToLower(strings.TrimSpace(claims.Role))
	}
	return context.WithValue(ctx, claimsKey{}, claims)
}

// GetClaims retrieves the claims from the context.
// Returns nil if no claims are present.
func GetClaims(ctx context.Context) *Claims {
	if claims, ok := ctx.Value(claimsKey{}).(*Claims); ok {
		return claims
	}
	return nil
}
