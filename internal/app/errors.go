// Package app provides application-layer types and utilities.
// It contains error types used by use cases to communicate failures
// with machine-readable codes for transport layer mapping.
package app

import (
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// Error codes re-exported from domain layer for app-layer use.
// This maintains clean layer abstraction while using domain as single source of truth.
//
// Format: ERR_{DOMAIN}_{CODE} (stable, documented in domain/errors/codes.go)
//
// IMPORTANT: These codes are STABLE and must not be changed once published.
// They are used in API responses and client integrations.
const (
	// User domain codes
	CodeUserNotFound = string(domainerrors.ErrCodeUserNotFound) // "ERR_USER_NOT_FOUND"
	CodeEmailExists  = string(domainerrors.ErrCodeEmailExists)  // "ERR_USER_EMAIL_EXISTS"

	// General codes
	CodeValidationError   = string(domainerrors.ErrCodeValidation)   // "ERR_VALIDATION"
	CodeUnauthorized      = string(domainerrors.ErrCodeUnauthorized) // "ERR_UNAUTHORIZED"
	CodeForbidden         = string(domainerrors.ErrCodeForbidden)    // "ERR_FORBIDDEN"
	CodeInternalError     = string(domainerrors.ErrCodeInternal)     // "ERR_INTERNAL"
	CodeRequestTooLarge   = "ERR_REQUEST_TOO_LARGE"                  // App-specific (no domain equivalent)
	CodeRateLimitExceeded = "ERR_RATE_LIMIT_EXCEEDED"                // App-specific (no domain equivalent)
)

// AppError represents an application-layer error with machine-readable code.
// It wraps domain errors and provides context for transport layer mapping.
type AppError struct {
	Op      string // operation name: "GetUser", "CreateUser"
	Code    string // machine-readable: "ERR_USER_NOT_FOUND"
	Message string // human-readable message
	Err     error  // wrapped error
}

// Error returns the error string representation.
func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Op + ": " + e.Message + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Message
}

// Unwrap returns the wrapped error for errors.Is and errors.As support.
func (e *AppError) Unwrap() error {
	return e.Err
}
