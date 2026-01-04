package domain

import "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"

// Sentinel errors for the domain layer.
// These errors are used by the application layer to identify specific error conditions
// and map them appropriately to transport-layer responses.
// Package errors aliases
// Use aliases to maintain backward compatibility while switching to the new structured error types.
var (
	// User domain errors.
	ErrUserNotFound       = errors.ErrUserNotFound
	ErrEmailAlreadyExists = errors.ErrEmailExists
	ErrInvalidEmail       = errors.ErrInvalidEmail
	ErrInvalidFirstName   = errors.ErrInvalidFirstName
	ErrInvalidLastName    = errors.ErrInvalidLastName

	// Audit domain errors.
	ErrAuditEventNotFound = errors.ErrAuditNotFound
	ErrInvalidEventType   = errors.ErrInvalidEventType
	ErrInvalidEntityType  = errors.ErrInvalidEntityType
	ErrInvalidEntityID    = errors.ErrInvalidEntityID
	ErrInvalidID          = errors.ErrInvalidID
	ErrInvalidTimestamp   = errors.ErrInvalidTimestamp
	ErrInvalidPayload     = errors.ErrInvalidPayload
	ErrInvalidRequestID   = errors.ErrInvalidRequestID
)
