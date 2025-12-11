package response

// Standard error codes for API responses.
// These codes are used in the error.code field of ErrorResponse.
const (
	// ErrBadRequest indicates a malformed request (HTTP 400).
	ErrBadRequest = "ERR_BAD_REQUEST"

	// ErrUnauthorized indicates missing or invalid authentication (HTTP 401).
	ErrUnauthorized = "ERR_UNAUTHORIZED"

	// ErrForbidden indicates insufficient permissions (HTTP 403).
	ErrForbidden = "ERR_FORBIDDEN"

	// ErrNotFound indicates the requested resource was not found (HTTP 404).
	ErrNotFound = "ERR_NOT_FOUND"

	// ErrConflict indicates a conflict with current state (HTTP 409).
	ErrConflict = "ERR_CONFLICT"

	// ErrValidation indicates validation errors in request data (HTTP 422).
	ErrValidation = "ERR_VALIDATION"

	// ErrInternalServer indicates an internal server error (HTTP 500).
	ErrInternalServer = "ERR_INTERNAL_SERVER"

	// ErrTimeout indicates a gateway timeout (HTTP 504).
	ErrTimeout = "ERR_TIMEOUT"

	// ErrServiceUnavailable indicates service is unavailable (HTTP 503).
	ErrServiceUnavailable = "ERR_SERVICE_UNAVAILABLE"
)
