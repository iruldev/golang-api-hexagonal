// Package ctxutil provides cross-cutting context utilities for extracting
// request-scoped information. This package can be imported from any layer
// in the hexagonal architecture.
//
// NOTE: The context setting (NewContext, NewRequestIDContext) should be done
// in the interface layer (middleware), but reading can happen anywhere.
package ctxutil

import "context"

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

const (
	claimsKey    contextKey = "auth_claims"
	requestIDKey contextKey = "request_id"
)

// Sentinel errors for context operations.
var (
	// ErrNoClaimsInContext indicates claims were not found in context.
	ErrNoClaimsInContext = claimsNotFoundError{}
)

type claimsNotFoundError struct{}

func (claimsNotFoundError) Error() string {
	return "no claims in context"
}

// NewClaimsContext returns a new context with the given claims.
func NewClaimsContext(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext extracts claims from context.
// Returns ErrNoClaimsInContext if claims are not present.
func ClaimsFromContext(ctx context.Context) (Claims, error) {
	claims, ok := ctx.Value(claimsKey).(Claims)
	if !ok {
		return Claims{}, ErrNoClaimsInContext
	}
	return claims, nil
}

// NewRequestIDContext returns a new context with the given request ID.
func NewRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext retrieves the request ID from context.
// Returns empty string if no request ID is set.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}
