package errors

import "errors"

// DomainError represents a domain-level error with a public code.
// It implements the standard error interface and supports error chaining.
//
// The Code field contains the public error code (UPPER_SNAKE_CASE) that is
// exposed in API responses for consistent client error handling.
//
// The Hint field is optional and provides additional guidance for API clients.
// WARNING: Never include sensitive information or internal error details in Hint.
type DomainError struct {
	// Code is the public error code in UPPER_SNAKE_CASE format.
	// Example: "NOT_FOUND", "VALIDATION_ERROR"
	Code string

	// Message is the human-readable error message.
	// This is returned to clients in the error.message field.
	Message string

	// Hint provides optional additional guidance for API clients.
	// WARNING: Never include sensitive info or internal error details here.
	Hint string

	// cause is the underlying error for error chaining.
	// This is unexported to prevent external modification.
	cause error
}

// Error implements the error interface.
// Returns the human-readable error message.
func (e *DomainError) Error() string {
	return e.Message
}

// Unwrap returns the underlying cause of the error for error chaining.
// This enables errors.Is() and errors.As() to traverse the error chain.
func (e *DomainError) Unwrap() error {
	return e.cause
}

// NewDomain creates a new DomainError with the given code and message.
// The code should be one of the constants from codes.go.
//
// Example:
//
//	err := errors.NewDomain(codes.CodeNotFound, "note not found")
func NewDomain(code, message string) *DomainError {
	if !IsValidCode(code) {
		panic("invalid domain error code: " + code)
	}
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// NewDomainWithHint creates a new DomainError with a hint for API clients.
// The hint provides additional guidance on how to resolve the error.
//
// WARNING: Never include sensitive information or internal error details in the hint.
// Hints should only contain user-facing guidance.
//
// Example:
//
//	err := errors.NewDomainWithHint(
//	    codes.CodeValidationError,
//	    "invalid email format",
//	    "email must be a valid email address like user@example.com",
//	)
func NewDomainWithHint(code, message, hint string) *DomainError {
	if !IsValidCode(code) {
		panic("invalid domain error code: " + code)
	}
	return &DomainError{
		Code:    code,
		Message: message,
		Hint:    hint,
	}
}

// NewDomainWithCause creates a new DomainError that wraps an underlying error.
// This enables error chaining with errors.Is() and errors.As().
//
// Example:
//
//	if err := db.Query(...); err != nil {
//	    return errors.NewDomainWithCause(codes.CodeNotFound, "note not found", err)
//	}
func NewDomainWithCause(code, message string, cause error) *DomainError {
	if !IsValidCode(code) {
		panic("invalid domain error code: " + code)
	}
	return &DomainError{
		Code:    code,
		Message: message,
		cause:   cause,
	}
}

// Is implements the interface for errors.Is() comparison.
// It returns true if the target is a *DomainError with the same Code.
// This allows type-safe error comparisons using errors.Is().
//
// Example:
//
//	if errors.Is(err, &errors.DomainError{Code: codes.CodeNotFound}) {
//	    // handle not found error
//	}
func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		// If target has a Code, compare codes
		if t.Code != "" {
			return e.Code == t.Code
		}
		// If target has no Code, just check type match
		return true
	}
	return false
}

// IsDomainError checks if the error is or wraps a DomainError.
// If true, returns the DomainError. If false, returns nil.
//
// Example:
//
//	if domainErr := errors.IsDomainError(err); domainErr != nil {
//	    log.Printf("Domain error: %s (code: %s)", domainErr.Message, domainErr.Code)
//	}
func IsDomainError(err error) *DomainError {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr
	}
	return nil
}
