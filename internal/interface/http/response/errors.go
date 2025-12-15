package response

// Standard error codes for API responses.
// These codes are used in the error.code field of API responses.
// Format: UPPER_SNAKE_CASE without ERR_ prefix.
const (
	// CodeBadRequest indicates a malformed request (HTTP 400).
	CodeBadRequest = "BAD_REQUEST"

	// CodeUnauthorized indicates missing or invalid authentication (HTTP 401).
	CodeUnauthorized = "UNAUTHORIZED"

	// CodeForbidden indicates insufficient permissions (HTTP 403).
	CodeForbidden = "FORBIDDEN"

	// CodeNotFound indicates the requested resource was not found (HTTP 404).
	CodeNotFound = "NOT_FOUND"

	// CodeConflict indicates a conflict with current state (HTTP 409).
	CodeConflict = "CONFLICT"

	// CodeValidation indicates validation errors in request data (HTTP 422).
	CodeValidation = "VALIDATION_FAILED"

	// CodeInternalServer indicates an internal server error (HTTP 500).
	CodeInternalServer = "INTERNAL_ERROR"

	// CodeTimeout indicates a gateway timeout (HTTP 504).
	CodeTimeout = "TIMEOUT"

	// CodeServiceUnavailable indicates service is unavailable (HTTP 503).
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// Deprecated: Use Code* constants instead. These will be removed in a future version.
const (
	ErrBadRequest         = "ERR_BAD_REQUEST"
	ErrUnauthorized       = "ERR_UNAUTHORIZED"
	ErrForbidden          = "ERR_FORBIDDEN"
	ErrNotFound           = "ERR_NOT_FOUND"
	ErrConflict           = "ERR_CONFLICT"
	ErrValidation         = "ERR_VALIDATION"
	ErrInternalServer     = "ERR_INTERNAL_SERVER"
	ErrTimeout            = "ERR_TIMEOUT"
	ErrServiceUnavailable = "ERR_SERVICE_UNAVAILABLE"
)
