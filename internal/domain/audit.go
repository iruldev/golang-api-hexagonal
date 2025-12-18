package domain

import (
	"context"
	"strings"
	"time"
)

// Standard event types for the Users module.
// New modules should follow the "entity.action" naming pattern.
// Examples: "user.created", "order.placed", "payment.completed".
const (
	// EventUserCreated is recorded when a new user is created.
	EventUserCreated = "user.created"

	// EventUserUpdated is recorded when a user is updated.
	EventUserUpdated = "user.updated"

	// EventUserDeleted is recorded when a user is deleted.
	EventUserDeleted = "user.deleted"
)

// AuditEvent represents an audit trail entry for tracking business operations.
// This entity follows hexagonal architecture principles - no external dependencies.
//
// EventType follows the pattern "entity.action", for example:
//   - "user.created"
//   - "user.updated"
//   - "user.deleted"
//   - "order.placed"
//
// The Payload field contains pre-redacted JSON data. PII redaction is handled
// by the app layer BEFORE creating the AuditEvent.
type AuditEvent struct {
	// ID is the unique identifier for this audit event.
	ID ID

	// EventType describes what happened, in "entity.action" format.
	// Examples: "user.created", "user.updated", "order.placed".
	EventType string

	// ActorID identifies who performed the action.
	// Empty for system-initiated or unauthenticated operations.
	ActorID ID

	// EntityType identifies the type of entity affected (e.g., "user", "order").
	EntityType string

	// EntityID identifies which specific entity was affected.
	EntityID ID

	// Payload contains JSON-encoded event data with PII already redacted.
	// This is stored as raw bytes to avoid importing encoding/json.
	Payload []byte

	// Timestamp records when the event occurred.
	Timestamp time.Time

	// RequestID correlates this event with the originating HTTP request.
	// This is a string (not domain.ID) since it comes from transport layer.
	RequestID string
}

// Validate checks if the AuditEvent has required fields.
// Returns a domain error if validation fails.
// ActorID is optional (can be empty for system/unauthenticated events).
// Validation order: EventType first, then EntityType, then EntityID.
func (e AuditEvent) Validate() error {
	if e.ID.IsEmpty() {
		return ErrInvalidID
	}

	if strings.TrimSpace(e.EventType) == "" {
		return ErrInvalidEventType
	}

	if strings.TrimSpace(e.EntityType) == "" {
		return ErrInvalidEntityType
	}

	if e.EntityID.IsEmpty() {
		return ErrInvalidEntityID
	}

	if e.Payload == nil {
		return ErrInvalidPayload
	}

	if e.Timestamp.IsZero() {
		return ErrInvalidTimestamp
	}

	return nil
}

// AuditEventRepository defines the interface for audit event persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
// All methods accept a Querier to support both connection pool and transaction usage.
type AuditEventRepository interface {
	// Create stores a new audit event.
	// Returns an error if the event cannot be persisted.
	Create(ctx context.Context, q Querier, event *AuditEvent) error

	// ListByEntityID retrieves audit events for a specific entity.
	// Results are ordered by timestamp DESC (newest first).
	// Returns the slice of events, total count of matching events, and any error.
	ListByEntityID(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)
}
