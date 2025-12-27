package domain

import (
	"context"
	"strings"
	"time"
)

// ============================================================================
// AUDIT EVENT TYPE NAMING CONVENTIONS
// ============================================================================
//
// Event types follow the pattern "entity.action" where:
//   - entity: Lowercase singular noun representing the domain entity (e.g., "user", "order", "payment")
//   - action: Lowercase past-tense verb describing what happened (e.g., "created", "updated", "deleted")
//
// Standard Actions (use these consistently across all entities):
//   - created   : A new entity was created
//   - updated   : An existing entity was modified
//   - deleted   : An entity was removed
//   - enabled   : An entity was activated/enabled
//   - disabled  : An entity was deactivated/disabled
//   - approved  : An entity was approved
//   - rejected  : An entity was rejected
//   - submitted : An entity was submitted for processing
//   - completed : A process or workflow was completed
//   - canceled  : A process or action was canceled
//
// Examples of valid event types:
//   - "user.created", "user.updated", "user.deleted"
//   - "order.placed", "order.shipped", "order.canceled"
//   - "payment.processed", "payment.refunded", "payment.failed"
//   - "session.started", "session.ended"
//   - "permission.granted", "permission.revoked"
//
// ============================================================================
// HOW TO ADD NEW AUDIT EVENT TYPES
// ============================================================================
//
// Step 1: Define a new constant in this file following the naming pattern:
//
//	const EventOrderCreated = "order.created"
//
// Step 2: Add the constant to your use case and call auditService.Record():
//
//	auditInput := audit.AuditEventInput{
//	    EventType:  domain.EventOrderCreated,
//	    ActorID:    req.ActorID,    // From request struct
//	    EntityType: "order",        // Matches the entity in event type
//	    EntityID:   order.ID,
//	    Payload:    order,          // Will be PII-redacted automatically
//	    RequestID:  req.RequestID,  // From request struct
//	}
//	err := uc.auditService.Record(ctx, tx, auditInput)
//
// Step 3: Ensure your use case request struct carries RequestID and ActorID:
//
//	type CreateOrderRequest struct {
//	    // ... order fields ...
//	    RequestID string    // Extracted by transport layer
//	    ActorID   domain.ID // Extracted by transport layer from JWT
//	}
//
// For complete integration reference, see internal/app/user/create_user.go.
// ============================================================================

// Standard event types for the Users module.
// New modules should follow the "entity.action" naming pattern.
// Examples: "user.created", "order.placed", "payment.completed".
const (
	// EventUserCreated is recorded when a new user is created.
	// Used in: CreateUserUseCase
	EventUserCreated = "user.created"

	// EventUserUpdated is recorded when a user is updated.
	// Used in: UpdateUserUseCase
	EventUserUpdated = "user.updated"

	// EventUserDeleted is recorded when a user is deleted.
	// Used in: DeleteUserUseCase
	EventUserDeleted = "user.deleted"
)

// ============================================================================
// EXAMPLE EVENT TYPES FOR OTHER MODULES (Templates for Extension)
// ============================================================================
//
// Orders module:
//   const EventOrderCreated  = "order.created"
//   const EventOrderUpdated  = "order.updated"
//   const EventOrderShipped  = "order.shipped"
//   const EventOrderCanceled = "order.canceled"
//
// Payments module:
//   const EventPaymentProcessed = "payment.processed"
//   const EventPaymentRefunded  = "payment.refunded"
//   const EventPaymentFailed    = "payment.failed"
//
// Sessions module:
//   const EventSessionStarted = "session.started"
//   const EventSessionEnded   = "session.ended"
//
// Permissions module:
//   const EventPermissionGranted = "permission.granted"
//   const EventPermissionRevoked = "permission.revoked"
// ============================================================================

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
// RequestID max length is 64.
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

	// Story 2.11 / 3.4: Max length is 64 to match X-Request-ID standard
	if len(e.RequestID) > 64 {
		return ErrInvalidRequestID
	}

	return nil
}

// AuditEventRepository defines the interface for audit event persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
// All methods accept a Querier to support both connection pool and transaction usage.
//
//go:generate mockgen -destination=../testutil/mocks/audit_event_repository_mock.go -package=mocks github.com/iruldev/golang-api-hexagonal/internal/domain AuditEventRepository
type AuditEventRepository interface {
	// Create stores a new audit event.
	// Returns an error if the event cannot be persisted.
	Create(ctx context.Context, q Querier, event *AuditEvent) error

	// ListByEntityID retrieves audit events for a specific entity.
	// Results are ordered by timestamp DESC (newest first).
	// Returns the slice of events, total count of matching events, and any error.
	ListByEntityID(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)
}
