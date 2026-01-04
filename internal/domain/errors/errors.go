// Package errors provides domain-level error types with stable error codes.
//
// This package defines structured error types that can be used throughout the
// application to provide consistent error handling and stable error codes
// for API responses.
//
// # Error Codes Stability
//
// Error codes defined in this package are STABLE and should not be changed
// once published. Adding new codes is allowed, but existing codes must not
// be modified or removed to maintain backward compatibility.
//
// # Usage
//
//	// Check for specific error
//	if errors.Is(err, errors.ErrUserNotFound) {
//	    // handle user not found
//	}
//
//	// Get error details
//	var domainErr *errors.DomainError
//	if errors.As(err, &domainErr) {
//	    log.Printf("Error code: %s", domainErr.Code)
//	}
//
//	// Create error with context
//	err := errors.NewUserNotFound("user-123")
package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents a stable, documented error code.
// Error codes follow the format: ERR_{DOMAIN}_{CODE}.
type ErrorCode string

// String returns the string representation of the error code.
func (c ErrorCode) String() string {
	return string(c)
}

// DomainError represents a domain-layer error with a stable code.
// It implements the error interface and supports error wrapping.
type DomainError struct {
	// Code is the stable error code for this error type.
	Code ErrorCode

	// Message is a human-readable description of the error.
	Message string

	// Err is the underlying error that caused this error, if any.
	Err error
}

// Error returns the error message with code prefix.
func (e *DomainError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for error chain traversal.
func (e *DomainError) Unwrap() error {
	return e.Err
}

// Is implements errors.Is matching by comparing error codes.
// Two DomainErrors are considered equal if they have the same Code.
func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// New creates a new DomainError with the given code and message.
func New(code ErrorCode, message string) error {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// Wrap creates a new DomainError that wraps an underlying error.
func Wrap(code ErrorCode, message string, err error) error {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
