package domain

import (
	"context"
	"strings"
	"time"
)

// User represents a minimal domain entity used to validate unit testing patterns.
// This is a sample entity created for Story 3.2 to demonstrate test infrastructure.
type User struct {
	ID        ID
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if the User entity has valid required fields.
// Returns a domain error if validation fails.
func (u User) Validate() error {
	if strings.TrimSpace(u.Email) == "" {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(u.Name) == "" {
		return ErrInvalidUserName
	}

	return nil
}

// UserRepository defines the interface for user persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
type UserRepository interface {
	// Create stores a new user and returns the created user with ID.
	Create(ctx context.Context, user User) (User, error)

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, id ID) (User, error)

	// GetByEmail retrieves a user by their email address.
	GetByEmail(ctx context.Context, email string) (User, error)

	// Update modifies an existing user.
	Update(ctx context.Context, user User) error

	// Delete removes a user by their ID.
	Delete(ctx context.Context, id ID) error
}
