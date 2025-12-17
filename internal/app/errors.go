// Package app provides application-layer types and utilities.
// It contains error types used by use cases to communicate failures
// with machine-readable codes for transport layer mapping.
package app

// Error codes for machine-readable error handling.
// These codes are used by the transport layer to map to HTTP status codes.
const (
	// CodeUserNotFound indicates that the requested user does not exist.
	CodeUserNotFound = "USER_NOT_FOUND"
	// CodeEmailExists indicates that a user with the given email already exists.
	CodeEmailExists = "EMAIL_EXISTS"
	// CodeValidationError indicates that input validation failed.
	CodeValidationError = "VALIDATION_ERROR"
	// CodeInternalError indicates an unexpected internal error.
	CodeInternalError = "INTERNAL_ERROR"
)

// AppError represents an application-layer error with machine-readable code.
// It wraps domain errors and provides context for transport layer mapping.
type AppError struct {
	Op      string // operation name: "GetUser", "CreateUser"
	Code    string // machine-readable: "USER_NOT_FOUND"
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
