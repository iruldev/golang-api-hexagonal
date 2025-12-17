package domain

import "errors"

// Sentinel errors for the domain layer.
// These errors are used by the application layer to identify specific error conditions
// and map them appropriately to transport-layer responses.
var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailExists is returned when attempting to create a user with an existing email.
	ErrEmailExists = errors.New("email already exists")

	// ErrInvalidEmail is returned when the email format is invalid or empty.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidUserName is returned when the user name is empty or invalid.
	ErrInvalidUserName = errors.New("invalid user name")
)
