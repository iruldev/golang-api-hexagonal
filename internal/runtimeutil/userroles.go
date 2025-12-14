// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/domain/auth"
)

// Sentinel errors for UserRoleProvider.
var (
	// ErrUserNotFound indicates the user was not found in the store.
	ErrUserNotFound = errors.New("userroles: user not found")
	// ErrInvalidUserID indicates the user ID is malformed (not a valid UUID).
	ErrInvalidUserID = errors.New("userroles: invalid user ID")
	// ErrInvalidRole indicates the role value is not valid.
	ErrInvalidRole = errors.New("userroles: invalid role")
)

// UserRoles represents a user's assigned roles with metadata.
type UserRoles struct {
	UserID    string    `json:"user_id"`
	Roles     []string  `json:"roles"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserRoleProvider defines the interface for user role management.
// Implement this interface for database-backed stores or external identity providers.
//
// Usage Example:
//
//	// Get user roles
//	roles, err := provider.GetUserRoles(ctx, "user-uuid")
//
//	// Set user roles (replace all)
//	err := provider.SetUserRoles(ctx, "user-uuid", []string{"admin", "user"})
//
//	// Add a single role
//	err := provider.AddUserRole(ctx, "user-uuid", "admin")
//
//	// Remove a single role
//	err := provider.RemoveUserRole(ctx, "user-uuid", "service")
type UserRoleProvider interface {
	// GetUserRoles returns the roles assigned to a user.
	// Returns ErrUserNotFound if the user doesn't exist in the store.
	GetUserRoles(ctx context.Context, userID string) (*UserRoles, error)

	// SetUserRoles replaces all roles for a user.
	// Returns ErrInvalidRole if any role is not valid.
	SetUserRoles(ctx context.Context, userID string, roles []string) (*UserRoles, error)

	// AddUserRole adds a single role to a user.
	// Idempotent: adding an existing role has no effect.
	// Returns ErrInvalidRole if the role is not valid.
	AddUserRole(ctx context.Context, userID string, role string) (*UserRoles, error)

	// RemoveUserRole removes a single role from a user.
	// Idempotent: removing a non-existent role has no effect.
	RemoveUserRole(ctx context.Context, userID string, role string) (*UserRoles, error)
}

// InMemoryUserRoleStore implements UserRoleProvider with in-memory storage.
// Thread-safe via sync.RWMutex. State changes are lost on restart.
//
// This implementation is suitable for development and testing.
// For production, use a persistent store (database, Redis).
type InMemoryUserRoleStore struct {
	mu    sync.RWMutex
	users map[string]*UserRoles
}

// InMemoryUserRoleOption is a functional option for configuring InMemoryUserRoleStore.
type InMemoryUserRoleOption func(*InMemoryUserRoleStore)

// NewInMemoryUserRoleStore creates a new InMemoryUserRoleStore.
//
// Example:
//
//	store := runtimeutil.NewInMemoryUserRoleStore(
//	    runtimeutil.WithInitialUserRoles(map[string][]string{
//	        "admin-uuid": {"admin", "user"},
//	        "user-uuid":  {"user"},
//	    }),
//	)
func NewInMemoryUserRoleStore(opts ...InMemoryUserRoleOption) *InMemoryUserRoleStore {
	s := &InMemoryUserRoleStore{
		users: make(map[string]*UserRoles),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// userRolesTimeNow is a variable for testing purposes.
var userRolesTimeNow = time.Now

// WithInitialUserRoles sets initial user roles when creating the store.
func WithInitialUserRoles(userRoles map[string][]string) InMemoryUserRoleOption {
	return func(s *InMemoryUserRoleStore) {
		now := userRolesTimeNow()
		for userID, roles := range userRoles {
			s.users[userID] = &UserRoles{
				UserID:    userID,
				Roles:     roles,
				UpdatedAt: now,
			}
		}
	}
}

// IsValidRole checks if the role is one of the defined standard roles.
func IsValidRole(role string) bool {
	return auth.Role(role).IsValid()
}

// GetUserRoles returns the roles assigned to a user.
func (s *InMemoryUserRoleStore) GetUserRoles(ctx context.Context, userID string) (*UserRoles, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	userRoles, exists := s.users[userID]
	if !exists {
		// Return empty roles for unknown users (auto-create on first write)
		return &UserRoles{
			UserID:    userID,
			Roles:     []string{},
			UpdatedAt: time.Time{},
		}, nil
	}

	// Return a copy to prevent external mutation (lock still held)
	rolesCopy := make([]string, len(userRoles.Roles))
	copy(rolesCopy, userRoles.Roles)

	return &UserRoles{
		UserID:    userRoles.UserID,
		Roles:     rolesCopy,
		UpdatedAt: userRoles.UpdatedAt,
	}, nil
}

// SetUserRoles replaces all roles for a user.
func (s *InMemoryUserRoleStore) SetUserRoles(ctx context.Context, userID string, roles []string) (*UserRoles, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate all roles
	for _, role := range roles {
		if !IsValidRole(role) {
			return nil, ErrInvalidRole
		}
	}

	// Deduplicate roles
	roleSet := make(map[string]bool)
	uniqueRoles := make([]string, 0, len(roles))
	for _, role := range roles {
		if !roleSet[role] {
			roleSet[role] = true
			uniqueRoles = append(uniqueRoles, role)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := userRolesTimeNow()
	s.users[userID] = &UserRoles{
		UserID:    userID,
		Roles:     uniqueRoles,
		UpdatedAt: now,
	}

	// Return a copy
	rolesCopy := make([]string, len(uniqueRoles))
	copy(rolesCopy, uniqueRoles)

	return &UserRoles{
		UserID:    userID,
		Roles:     rolesCopy,
		UpdatedAt: now,
	}, nil
}

// AddUserRole adds a single role to a user.
func (s *InMemoryUserRoleStore) AddUserRole(ctx context.Context, userID string, role string) (*UserRoles, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if !IsValidRole(role) {
		return nil, ErrInvalidRole
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := userRolesTimeNow()

	userRoles, exists := s.users[userID]
	if !exists {
		// Create new user with the role
		s.users[userID] = &UserRoles{
			UserID:    userID,
			Roles:     []string{role},
			UpdatedAt: now,
		}
		return &UserRoles{
			UserID:    userID,
			Roles:     []string{role},
			UpdatedAt: now,
		}, nil
	}

	// Check if role already exists (idempotent)
	for _, r := range userRoles.Roles {
		if r == role {
			// Role already exists, return current state
			rolesCopy := make([]string, len(userRoles.Roles))
			copy(rolesCopy, userRoles.Roles)
			return &UserRoles{
				UserID:    userRoles.UserID,
				Roles:     rolesCopy,
				UpdatedAt: userRoles.UpdatedAt,
			}, nil
		}
	}

	// Add the role
	userRoles.Roles = append(userRoles.Roles, role)
	userRoles.UpdatedAt = now

	// Return a copy
	rolesCopy := make([]string, len(userRoles.Roles))
	copy(rolesCopy, userRoles.Roles)

	return &UserRoles{
		UserID:    userID,
		Roles:     rolesCopy,
		UpdatedAt: now,
	}, nil
}

// RemoveUserRole removes a single role from a user.
func (s *InMemoryUserRoleStore) RemoveUserRole(ctx context.Context, userID string, role string) (*UserRoles, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Validate role before proceeding
	if !IsValidRole(role) {
		return nil, ErrInvalidRole
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := userRolesTimeNow()

	userRoles, exists := s.users[userID]
	if !exists {
		// User doesn't exist, return empty roles (idempotent)
		return &UserRoles{
			UserID:    userID,
			Roles:     []string{},
			UpdatedAt: now,
		}, nil
	}

	// Find and remove the role
	newRoles := make([]string, 0, len(userRoles.Roles))
	found := false
	for _, r := range userRoles.Roles {
		if r == role {
			found = true
			continue
		}
		newRoles = append(newRoles, r)
	}

	if !found {
		// Role not found, return current state (idempotent)
		rolesCopy := make([]string, len(userRoles.Roles))
		copy(rolesCopy, userRoles.Roles)
		return &UserRoles{
			UserID:    userRoles.UserID,
			Roles:     rolesCopy,
			UpdatedAt: userRoles.UpdatedAt,
		}, nil
	}

	// Update the roles
	userRoles.Roles = newRoles
	userRoles.UpdatedAt = now

	// Return a copy
	rolesCopy := make([]string, len(newRoles))
	copy(rolesCopy, newRoles)

	return &UserRoles{
		UserID:    userID,
		Roles:     rolesCopy,
		UpdatedAt: now,
	}, nil
}
