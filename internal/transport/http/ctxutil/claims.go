// Package ctxutil provides context utility functions for storing and retrieving
// request-scoped values such as JWT claims.
package ctxutil

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
)

// claimsKey is the unexported type for the context key to prevent collisions.
type claimsKey struct{}

// Claims represents JWT claims extracted from the token.
// It embeds jwt.RegisteredClaims for standard fields (iss, sub, aud, exp, nbf, iat, jti).
type Claims struct {
	jwt.RegisteredClaims
	// Custom claims can be added here as the application evolves.
	// Example: UserID string `json:"userId,omitempty"`
}

// SetClaims stores the claims in the context.
func SetClaims(ctx context.Context, claims *Claims) context.Context {
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
