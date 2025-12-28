package errors

// Stable Error Codes
//
// IMPORTANT: Error codes in this file are STABLE and must not be changed
// once published. Adding new codes is allowed, but modifying or removing
// existing codes will break backward compatibility.
//
// Format: ERR_{DOMAIN}_{CODE}

// User domain error codes
const (
	// ErrCodeUserNotFound indicates that a requested user was not found.
	ErrCodeUserNotFound ErrorCode = "ERR_USER_NOT_FOUND"

	// ErrCodeEmailExists indicates that the email is already registered.
	ErrCodeEmailExists ErrorCode = "ERR_USER_EMAIL_EXISTS"

	// ErrCodeInvalidEmail indicates that the email format is invalid.
	ErrCodeInvalidEmail ErrorCode = "ERR_USER_INVALID_EMAIL"

	// ErrCodeInvalidFirstName indicates that the first name is invalid or empty.
	ErrCodeInvalidFirstName ErrorCode = "ERR_USER_INVALID_FIRST_NAME"

	// ErrCodeInvalidLastName indicates that the last name is invalid or empty.
	ErrCodeInvalidLastName ErrorCode = "ERR_USER_INVALID_LAST_NAME"
)

// Audit domain error codes
const (
	// ErrCodeAuditNotFound indicates that an audit event was not found.
	ErrCodeAuditNotFound ErrorCode = "ERR_AUDIT_NOT_FOUND"

	// ErrCodeInvalidEventType indicates that the event type is invalid.
	ErrCodeInvalidEventType ErrorCode = "ERR_AUDIT_INVALID_EVENT_TYPE"

	// ErrCodeInvalidEntityType indicates that the entity type is invalid.
	ErrCodeInvalidEntityType ErrorCode = "ERR_AUDIT_INVALID_ENTITY_TYPE"

	// ErrCodeInvalidEntityID indicates that the entity ID is invalid.
	ErrCodeInvalidEntityID ErrorCode = "ERR_AUDIT_INVALID_ENTITY_ID"

	// ErrCodeInvalidID indicates that the audit event ID is invalid.
	ErrCodeInvalidID ErrorCode = "ERR_AUDIT_INVALID_ID"

	// ErrCodeInvalidTimestamp indicates that the timestamp is invalid.
	ErrCodeInvalidTimestamp ErrorCode = "ERR_AUDIT_INVALID_TIMESTAMP"

	// ErrCodeInvalidPayload indicates that the payload is invalid.
	ErrCodeInvalidPayload ErrorCode = "ERR_AUDIT_INVALID_PAYLOAD"

	// ErrCodeInvalidRequestID indicates that the request ID is invalid.
	ErrCodeInvalidRequestID ErrorCode = "ERR_AUDIT_INVALID_REQUEST_ID"
)

// General error codes
const (
	// ErrCodeInternal indicates an internal server error.
	ErrCodeInternal ErrorCode = "ERR_INTERNAL"

	// ErrCodeValidation indicates a validation error.
	ErrCodeValidation ErrorCode = "ERR_VALIDATION"

	// ErrCodeNotFound indicates a generic not found error.
	ErrCodeNotFound ErrorCode = "ERR_NOT_FOUND"

	// ErrCodeConflict indicates a resource conflict.
	ErrCodeConflict ErrorCode = "ERR_CONFLICT"

	// ErrCodeUnauthorized indicates missing or invalid authentication.
	ErrCodeUnauthorized ErrorCode = "ERR_UNAUTHORIZED"

	// ErrCodeForbidden indicates insufficient permissions.
	ErrCodeForbidden ErrorCode = "ERR_FORBIDDEN"
)

// Sentinel errors for common error conditions.
// These can be used with errors.Is() for comparison.
var (
	// User errors
	// ErrUserNotFound indicates a user was not found.
	ErrUserNotFound = New(ErrCodeUserNotFound, "user not found")
	// ErrEmailExists indicates the email is already in use.
	ErrEmailExists = New(ErrCodeEmailExists, "email already exists")
	// ErrInvalidEmail indicates the email format is invalid.
	ErrInvalidEmail = New(ErrCodeInvalidEmail, "invalid email format")
	// ErrInvalidFirstName indicates the first name is invalid.
	ErrInvalidFirstName = New(ErrCodeInvalidFirstName, "invalid first name")
	// ErrInvalidLastName indicates the last name is invalid.
	ErrInvalidLastName = New(ErrCodeInvalidLastName, "invalid last name")

	// Audit errors
	// ErrAuditNotFound indicates an audit log entry was not found.
	ErrAuditNotFound = New(ErrCodeAuditNotFound, "audit event not found")
	// ErrInvalidEventType indicates the event type is unsupported.
	ErrInvalidEventType = New(ErrCodeInvalidEventType, "invalid event type")
	// ErrInvalidEntityType indicates the entity type is unsupported.
	ErrInvalidEntityType = New(ErrCodeInvalidEntityType, "invalid entity type")
	// ErrInvalidEntityID indicates the entity ID is invalid.
	ErrInvalidEntityID = New(ErrCodeInvalidEntityID, "invalid entity ID")
	// ErrInvalidID indicates the ID is invalid.
	ErrInvalidID = New(ErrCodeInvalidID, "invalid ID")
	// ErrInvalidTimestamp indicates the timestamp is invalid.
	ErrInvalidTimestamp = New(ErrCodeInvalidTimestamp, "invalid timestamp")
	// ErrInvalidPayload indicates the payload is invalid.
	ErrInvalidPayload = New(ErrCodeInvalidPayload, "invalid payload")
	// ErrInvalidRequestID indicates the request ID is invalid.
	ErrInvalidRequestID = New(ErrCodeInvalidRequestID, "invalid request ID")

	// General errors
	// ErrInternal indicates an internal server error.
	ErrInternal = New(ErrCodeInternal, "internal error")
	// ErrValidation indicates a general validation error.
	ErrValidation = New(ErrCodeValidation, "validation error")
	// ErrNotFound indicates a resource was not found.
	ErrNotFound = New(ErrCodeNotFound, "not found")
	// ErrConflict indicates a resource conflict.
	ErrConflict = New(ErrCodeConflict, "resource conflict")
	// ErrUnauthorized indicates the user is not authenticated.
	ErrUnauthorized = New(ErrCodeUnauthorized, "unauthorized")
	// ErrForbidden indicates the user is not authorized.
	ErrForbidden = New(ErrCodeForbidden, "forbidden")
)

// Convenience constructors for user errors

// NewUserNotFound creates a user not found error with context.
func NewUserNotFound(userID string) error {
	return &DomainError{
		Code:    ErrCodeUserNotFound,
		Message: "user " + userID + " not found",
	}
}

// NewEmailExists creates an email exists error with context.
func NewEmailExists(email string) error {
	return &DomainError{
		Code:    ErrCodeEmailExists,
		Message: "email " + email + " already exists",
	}
}

// NewValidationError creates a validation error with a specific message.
func NewValidationError(message string) error {
	return &DomainError{
		Code:    ErrCodeValidation,
		Message: message,
	}
}

// NewInternalError creates an internal error wrapping an underlying error.
func NewInternalError(message string, cause error) error {
	return &DomainError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     cause,
	}
}
