package domain

import "errors"

// Sentinel errors for the domain layer.
// These errors are used by the application layer to identify specific error conditions
// and map them appropriately to transport-layer responses.
var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")

	// ErrEmailAlreadyExists is returned when attempting to create a user with an existing email.
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrInvalidEmail is returned when the email format is invalid or empty.
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidFirstName is returned when the first name is empty or invalid.
	ErrInvalidFirstName = errors.New("invalid first name")

	// ErrInvalidLastName is returned when the last name is empty or invalid.
	ErrInvalidLastName = errors.New("invalid last name")

	// Audit event errors

	// ErrAuditEventNotFound is returned when an audit event cannot be found.
	ErrAuditEventNotFound = errors.New("audit event not found")

	// ErrInvalidEventType is returned when the event type format is invalid or empty.
	ErrInvalidEventType = errors.New("invalid event type format")

	// ErrInvalidEntityType is returned when the entity type is empty or invalid.
	ErrInvalidEntityType = errors.New("invalid entity type")

	// ErrInvalidEntityID is returned when the entity ID is empty or invalid.
	ErrInvalidEntityID = errors.New("invalid entity ID")

	// ErrInvalidID is returned when the audit event ID is empty or invalid.
	ErrInvalidID = errors.New("invalid audit event ID")

	// ErrInvalidTimestamp is returned when the audit event timestamp is zero.
	ErrInvalidTimestamp = errors.New("invalid timestamp")

	// ErrInvalidPayload is returned when the audit event payload is nil.
	ErrInvalidPayload = errors.New("invalid payload")
)
