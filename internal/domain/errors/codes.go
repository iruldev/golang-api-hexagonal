// Package errors provides the central error code registry and domain error types.
//
// # Error Code Naming Convention
//
// All public error codes follow UPPER_SNAKE_CASE format without prefix:
//   - ✅ NOT_FOUND (correct)
//   - ❌ ERR_NOT_FOUND (incorrect - no ERR_ prefix)
//   - ❌ NotFound (incorrect - use UPPER_SNAKE_CASE)
//
// These codes are used in API responses for consistent client error handling.
// Each code maps to a specific error type that can be programmatically handled.
//
// # Usage
//
//	err := errors.NewDomain(codes.CodeNotFound, "note not found")
//	// or with hint
//	err := errors.NewDomainWithHint(codes.CodeValidation, "invalid email", "check email format")
package errors

// Central error code constants for the domain layer.
// These codes are used in API responses for consistent error handling.
// Format: UPPER_SNAKE_CASE without ERR_ prefix.
const (
	// CodeNotFound indicates a requested resource was not found.
	CodeNotFound = "NOT_FOUND"

	// CodeValidationError indicates validation errors in request data.
	CodeValidationError = "VALIDATION_ERROR"

	// CodeUnauthorized indicates missing or invalid authentication.
	CodeUnauthorized = "UNAUTHORIZED"

	// CodeForbidden indicates insufficient permissions.
	CodeForbidden = "FORBIDDEN"

	// CodeConflict indicates a conflict with current state.
	CodeConflict = "CONFLICT"

	// CodeInternalError indicates an internal server error.
	CodeInternalError = "INTERNAL_ERROR"

	// CodeTimeout indicates an operation timed out.
	CodeTimeout = "TIMEOUT"

	// CodeRateLimitExceeded indicates rate limit has been exceeded.
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"

	// CodeBadRequest indicates a malformed request.
	CodeBadRequest = "BAD_REQUEST"
)

// allCodes is a registry of all valid error codes.
var allCodes = map[string]struct{}{
	CodeNotFound:          {},
	CodeValidationError:   {},
	CodeUnauthorized:      {},
	CodeForbidden:         {},
	CodeConflict:          {},
	CodeInternalError:     {},
	CodeTimeout:           {},
	CodeRateLimitExceeded: {},
	CodeBadRequest:        {},
}

// IsValidCode checks if the provided code is a valid registered error code.
func IsValidCode(code string) bool {
	_, ok := allCodes[code]
	return ok
}
