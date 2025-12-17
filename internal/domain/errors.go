package domain

import "errors"

// Sentinel errors for the domain layer.
// These errors are used by the application layer to identify specific error conditions
// and map them appropriately to transport-layer responses.
var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailAlreadyExists is returned when attempting to create a user with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidEmail is returned when the email format is invalid or empty.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidFirstName is returned when the first name is empty or invalid.
	ErrInvalidFirstName = errors.New("invalid first name")

	// ErrInvalidLastName is returned when the last name is empty or invalid.
	ErrInvalidLastName = errors.New("invalid last name")
)
