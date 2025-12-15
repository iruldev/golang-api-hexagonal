// Package domain contains core business logic and domain types.
package domain

import (
	"errors"

	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// Domain errors using the central error code registry.
// Use these for new code instead of the deprecated sentinel errors below.
var (
	// NewErrNotFound creates a NOT_FOUND domain error.
	NewErrNotFound = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeNotFound, msg)
	}

	// NewErrValidation creates a VALIDATION_ERROR domain error.
	NewErrValidation = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeValidationError, msg)
	}

	// NewErrUnauthorized creates an UNAUTHORIZED domain error.
	NewErrUnauthorized = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeUnauthorized, msg)
	}

	// NewErrForbidden creates a FORBIDDEN domain error.
	NewErrForbidden = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeForbidden, msg)
	}

	// NewErrConflict creates a CONFLICT domain error.
	NewErrConflict = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeConflict, msg)
	}

	// NewErrInternal creates an INTERNAL_ERROR domain error.
	NewErrInternal = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeInternalError, msg)
	}

	// NewErrTimeout creates a TIMEOUT domain error.
	NewErrTimeout = func(msg string) *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeTimeout, msg)
	}
)

// Deprecated: Use NewErr* functions or domainerrors.NewDomain() instead.
// These sentinel errors are kept for backward compatibility.
// They will be removed in a future version.
//
// Domain error types for mapping to HTTP status codes.
// Use errors.Is() to compare: errors.Is(err, ErrNotFound)
var (
	// Deprecated: Use NewErrNotFound("message") instead.
	// ErrNotFound indicates a resource was not found.
	// Maps to HTTP 404 Not Found.
	ErrNotFound = errors.New("resource not found")

	// Deprecated: Use NewErrValidation("message") instead.
	// ErrValidation indicates invalid input data.
	// Maps to HTTP 400 Bad Request.
	ErrValidation = errors.New("validation error")

	// Deprecated: Use NewErrUnauthorized("message") instead.
	// ErrUnauthorized indicates missing or invalid credentials.
	// Maps to HTTP 401 Unauthorized.
	ErrUnauthorized = errors.New("unauthorized")

	// Deprecated: Use NewErrForbidden("message") instead.
	// ErrForbidden indicates insufficient permissions.
	// Maps to HTTP 403 Forbidden.
	ErrForbidden = errors.New("forbidden")

	// Deprecated: Use NewErrConflict("message") instead.
	// ErrConflict indicates a conflict with current state.
	// Maps to HTTP 409 Conflict.
	ErrConflict = errors.New("conflict")

	// Deprecated: Use NewErrInternal("message") instead.
	// ErrInternal indicates an internal server error.
	// Maps to HTTP 500 Internal Server Error.
	ErrInternal = errors.New("internal error")

	// Deprecated: Use NewErrTimeout("message") instead.
	// ErrTimeout indicates an operation timed out.
	// Maps to HTTP 504 Gateway Timeout.
	ErrTimeout = errors.New("operation timed out")
)

// Deprecated: Use domainerrors.NewDomainWithCause() instead.
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
