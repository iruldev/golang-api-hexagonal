package domain

import (
	"context"
	"strings"
	"time"
)

// User represents a domain entity for user data.
// This entity follows hexagonal architecture principles - no external dependencies.
type User struct {
	ID        ID
	Email     string
	FirstName string
	LastName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Validate checks if the User entity has valid required fields.
// Returns a domain error if validation fails.
// Validation order: Email first, then FirstName, then LastName.
func (u User) Validate() error {
	if strings.TrimSpace(u.Email) == "" {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(u.FirstName) == "" {
		return ErrInvalidFirstName
	}

	if strings.TrimSpace(u.LastName) == "" {
		return ErrInvalidLastName
	}

	return nil
}

// UserRepository defines the interface for user persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
// All methods accept a Querier to support both connection pool and transaction usage.
type UserRepository interface {
	// Create stores a new user.
	Create(ctx context.Context, q Querier, user *User) error

	// GetByID retrieves a user by their ID.
	GetByID(ctx context.Context, q Querier, id ID) (*User, error)

	// List retrieves users with pagination.
	// Returns the slice of users, total count of matching users, and any error.
	List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)
}
