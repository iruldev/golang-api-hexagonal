// Package domain contains core business logic and domain types.
package domain

import "errors"

// Domain error types for mapping to HTTP status codes (Story 3.8).
// Use errors.Is() to compare: errors.Is(err, ErrNotFound)
var (
	// ErrNotFound indicates a resource was not found.
	// Maps to HTTP 404 Not Found.
	ErrNotFound = errors.New("resource not found")

	// ErrValidation indicates invalid input data.
	// Maps to HTTP 400 Bad Request.
	ErrValidation = errors.New("validation error")

	// ErrUnauthorized indicates missing or invalid credentials.
	// Maps to HTTP 401 Unauthorized.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden indicates insufficient permissions.
	// Maps to HTTP 403 Forbidden.
	ErrForbidden = errors.New("forbidden")

	// ErrConflict indicates a conflict with current state.
	// Maps to HTTP 409 Conflict.
	ErrConflict = errors.New("conflict")

	// ErrInternal indicates an internal server error.
	// Maps to HTTP 500 Internal Server Error.
	ErrInternal = errors.New("internal error")
)

// WrapError wraps an error with a domain error type.
// This allows errors.Is() to match the domain error.
//
// Example:
//
//	return domain.WrapError(domain.ErrNotFound, "user not found")
func WrapError(domainErr error, message string) error {
	return &wrappedError{
		domainErr: domainErr,
		message:   message,
	}
}

type wrappedError struct {
	domainErr error
	message   string
}

func (e *wrappedError) Error() string {
	return e.message
}

func (e *wrappedError) Unwrap() error {
	return e.domainErr
}
