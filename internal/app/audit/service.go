// Package audit provides use cases for audit event operations.
//
// This package follows hexagonal architecture principles - it depends only on the domain layer
// and defines ports (interfaces) that infrastructure adapters must implement.
//
// # Overview
//
// The audit package provides the AuditService which handles:
//   - Recording audit events with automatic PII redaction
//   - Querying audit events by entity
//
// # Adding Audit Events to a New Module
//
// Follow these steps to integrate audit events into your module:
//
// 1. DEFINE EVENT TYPE CONSTANT in internal/domain/audit.go:
//
//	const EventOrderCreated = "order.created"
//
// 2. ADD DEPENDENCY to your use case struct:
//
//	type CreateOrderUseCase struct {
//	    orderRepo    domain.OrderRepository
//	    auditService *audit.AuditService  // Add this
//	    txManager    domain.TxManager
//	    db           domain.Querier
//	}
//
// 3. UPDATE REQUEST STRUCT to carry context values:
//
//	type CreateOrderRequest struct {
//	    // ... order fields ...
//	    RequestID string    // Transport layer extracts this
//	    ActorID   domain.ID // Transport layer extracts from JWT
//	}
//
// 4. RECORD AUDIT EVENT within your transaction:
//
//	func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) error {
//	    return uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
//	        // ... create order logic ...
//
//	        auditInput := audit.AuditEventInput{
//	            EventType:  domain.EventOrderCreated,
//	            ActorID:    req.ActorID,
//	            EntityType: "order",
//	            EntityID:   order.ID,
//	            Payload:    order,  // Automatically PII-redacted
//	            RequestID:  req.RequestID,
//	        }
//	        return uc.auditService.Record(ctx, tx, auditInput)
//	    })
//	}
//
// 5. UPDATE HANDLER to populate context values:
//
//	func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
//	    req := order.CreateOrderRequest{
//	        // ... order fields from body ...
//	        RequestID: ctxutil.GetRequestID(r.Context()),
//	        ActorID:   getActorIDFromContext(r.Context()),
//	    }
//	    resp, err := h.useCase.Execute(r.Context(), req)
//	    // ...
//	}
//
// 6. ADD TESTS verifying audit recording (see internal/app/user/create_user_test.go)
//
// # Developer Checklist
//
// Before considering your audit integration complete:
//
//   - [ ] Event type constant defined in domain layer
//   - [ ] AuditService injected as dependency in use case
//   - [ ] Request struct includes RequestID and ActorID fields
//   - [ ] auditService.Record() called within transaction
//   - [ ] Handler extracts and passes RequestID and ActorID
//   - [ ] Unit tests mock AuditService and verify Record() is called
//   - [ ] `make lint` passes (no architecture violations)
//   - [ ] `make test` passes (all tests green)
//
// # Canonical Reference
//
// See internal/app/user/create_user.go for a complete working example.
package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// AuditEventInput represents the input data for recording an audit event.
// This struct provides a simpler API than constructing domain.AuditEvent directly.
// The transport layer extracts context values (requestID, actorID) and passes them here.
type AuditEventInput struct {
	// EventType describes what happened, in "entity.action" format.
	// Use domain constants like domain.EventUserCreated.
	EventType string

	// ActorID identifies who performed the action.
	// Empty for system-initiated or unauthenticated operations.
	ActorID domain.ID

	// EntityType identifies the type of entity affected (e.g., "user", "order").
	EntityType string

	// EntityID identifies which specific entity was affected.
	EntityID domain.ID

	// Payload is the event data that will be redacted and marshaled to JSON.
	// Can be any JSON-serializable type (struct, map, slice).
	Payload any

	// RequestID correlates this event with the originating HTTP request.
	// Extracted from context by transport layer, passed here.
	RequestID string
}

// AuditService provides audit event recording and querying capabilities.
// It orchestrates PII redaction and delegates persistence to the repository.
type AuditService struct {
	repo     domain.AuditEventRepository
	redactor domain.Redactor
	idGen    domain.IDGenerator
}

// NewAuditService creates a new AuditService instance.
func NewAuditService(
	repo domain.AuditEventRepository,
	redactor domain.Redactor,
	idGen domain.IDGenerator,
) *AuditService {
	return &AuditService{
		repo:     repo,
		redactor: redactor,
		idGen:    idGen,
	}
}

// Record persists an audit event within the provided transaction.
// PII fields in the payload are automatically redacted before storage.
// RequestID and ActorID come from the input struct (passed by transport layer).
func (s *AuditService) Record(ctx context.Context, q domain.Querier, input AuditEventInput) error {
	op := "AuditService.Record"

	// Redact payload using injected redactor
	redactedData := s.redactor.Redact(input.Payload)
	payload, err := json.Marshal(redactedData)
	if err != nil {
		return &app.AppError{
			Op:      op,
			Code:    app.CodeInternalError,
			Message: "Failed to marshal payload",
			Err:     err,
		}
	}

	// Create domain event (requestID and actorID come from input, not context)
	event := &domain.AuditEvent{
		ID:         s.idGen.NewID(),
		EventType:  input.EventType,
		ActorID:    input.ActorID,
		EntityType: input.EntityType,
		EntityID:   input.EntityID,
		Payload:    payload,
		Timestamp:  time.Now().UTC(),
		RequestID:  input.RequestID,
	}

	// Validate event using domain rules
	if err := event.Validate(); err != nil {
		return &app.AppError{
			Op:      op,
			Code:    app.CodeValidationError,
			Message: "Invalid audit event",
			Err:     err,
		}
	}

	// Persist via repository
	if err := s.repo.Create(ctx, q, event); err != nil {
		return &app.AppError{
			Op:      op,
			Code:    app.CodeInternalError,
			Message: "Failed to record audit event",
			Err:     err,
		}
	}

	return nil
}

// ListByEntity retrieves audit events for a specific entity.
// Results are ordered by timestamp DESC (newest first).
func (s *AuditService) ListByEntity(
	ctx context.Context,
	q domain.Querier,
	entityType string,
	entityID domain.ID,
	params domain.ListParams,
) ([]domain.AuditEvent, int, error) {
	return s.repo.ListByEntityID(ctx, q, entityType, entityID, params)
}
