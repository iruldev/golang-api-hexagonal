// Package app provides application-layer types and utilities.
// This file contains authorization types and helpers for role-based access control.
// Authorization checks happen in app layer (use cases) per architecture.md.
package app

import (
	"context"
	"errors"
)

// Role constants for authorization.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// ErrNoAuthContext indicates that no authentication context was found.
var ErrNoAuthContext = errors.New("no authentication context")

// AuthContext represents the authenticated actor for authorization checks.
// It is used by use cases to verify permissions before executing business logic.
type AuthContext struct {
	SubjectID string // From claims.Subject (user ID)
	Role      string // Role for authorization
}

// authContextKey is the unexported type for the context key to prevent collisions.
type authContextKey struct{}

// SetAuthContext stores the auth context in the request context.
func SetAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey{}, authCtx)
}

// GetAuthContext retrieves the auth context from the request context.
// Returns nil if no auth context is present.
func GetAuthContext(ctx context.Context) *AuthContext {
	if authCtx, ok := ctx.Value(authContextKey{}).(*AuthContext); ok {
		return authCtx
	}
	return nil
}

// HasRole checks if the auth context has the specified role.
func (ac *AuthContext) HasRole(role string) bool {
	return ac != nil && ac.Role == role
}

// IsAdmin checks if the auth context has admin role.
func (ac *AuthContext) IsAdmin() bool {
	return ac.HasRole(RoleAdmin)
}

// IsUser checks if the auth context has user role.
func (ac *AuthContext) IsUser() bool {
	return ac.HasRole(RoleUser)
}
