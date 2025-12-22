# Adding Audit Events to New Modules

This guide explains how to add audit event recording to new modules in the golang-api-hexagonal project.

## Overview

The audit system records business operations for compliance and debugging. Events are:
- Stored in the `audit_events` PostgreSQL table
- Automatically PII-redacted before storage
- Recorded atomically within the same database transaction as the business operation

## Quick Start

### 1. Define Event Type Constant

Add your event type constant in `internal/domain/audit.go`:

```go
const EventOrderCreated = "order.created"
```

**Naming Convention:** `entity.action` (lowercase)
- entity: Singular noun (e.g., `user`, `order`, `payment`)
- action: Past-tense verb (e.g., `created`, `updated`, `deleted`)

### 2. Add AuditService Dependency

Update your use case struct:

```go
type CreateOrderUseCase struct {
    orderRepo    domain.OrderRepository
    auditService *audit.AuditService  // Add this
    txManager    domain.TxManager
    db           domain.Querier
}

func NewCreateOrderUseCase(
    orderRepo domain.OrderRepository,
    auditService *audit.AuditService,
    txManager domain.TxManager,
    db domain.Querier,
) *CreateOrderUseCase {
    return &CreateOrderUseCase{
        orderRepo:    orderRepo,
        auditService: auditService,
        txManager:    txManager,
        db:           db,
    }
}
```

### 3. Update Request Struct

Add `RequestID` and `ActorID` fields to carry context:

```go
type CreateOrderRequest struct {
    // ... business fields ...
    
    // RequestID correlates with the HTTP request
    RequestID string
    // ActorID identifies who performed the action
    ActorID domain.ID
}
```

### 4. Record Audit Event

Call `auditService.Record()` within your transaction:

```go
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (CreateOrderResponse, error) {
    var order *domain.Order
    
    err := uc.txManager.WithTx(ctx, func(tx domain.Querier) error {
        // Business logic
        order = &domain.Order{
            ID:     uc.idGen.NewID(),
            // ... other fields ...
        }
        
        if err := uc.orderRepo.Create(ctx, tx, order); err != nil {
            return err
        }
        
        // Record audit event (same transaction)
        auditInput := audit.AuditEventInput{
            EventType:  domain.EventOrderCreated,
            ActorID:    req.ActorID,
            EntityType: "order",
            EntityID:   order.ID,
            Payload:    order,  // Automatically PII-redacted
            RequestID:  req.RequestID,
        }
        
        return uc.auditService.Record(ctx, tx, auditInput)
    })
    
    if err != nil {
        return CreateOrderResponse{}, err
    }
    
    return CreateOrderResponse{Order: *order}, nil
}
```

### 5. Update HTTP Handler

Extract context values in your handler:

```go
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var body CreateOrderBody
    // ... parse body ...
    
    req := order.CreateOrderRequest{
        // ... map body fields ...
        RequestID: middleware.GetRequestID(r.Context()),
        ActorID:   getActorIDFromJWT(r.Context()),
    }
    
    resp, err := h.createOrderUC.Execute(r.Context(), req)
    // ... handle response ...
}
```

### 6. Wire Dependencies in main.go

```go
// In cmd/api/main.go
auditService := audit.NewAuditService(auditRepo, redactor, idGen)
createOrderUC := order.NewCreateOrderUseCase(orderRepo, auditService, txManager, pool)
```

## Standard Actions

| Action | When to Use |
|--------|-------------|
| `created` | Entity created |
| `updated` | Entity modified |
| `deleted` | Entity removed |
| `enabled` | Entity activated |
| `disabled` | Entity deactivated |
| `approved` | Entity approved |
| `rejected` | Entity rejected |
| `submitted` | Entity submitted for processing |
| `completed` | Process completed |
| `canceled` | Action canceled |

## Testing

Mock `AuditService` dependencies in unit tests. Since `AuditService` is a concrete struct, you cannot mock it directly. Instead, create a real `AuditService` with mock dependencies:

```go
// Test double for AuditEventRepository
type mockAuditRepo struct {
    events []*domain.AuditEvent
}

func (m *mockAuditRepo) Create(ctx context.Context, q domain.Querier, event *domain.AuditEvent) error {
    m.events = append(m.events, event)
    return nil
}

func (m *mockAuditRepo) ListByEntityID(ctx context.Context, q domain.Querier, entityType string, entityID domain.ID, params domain.ListParams) ([]domain.AuditEvent, int, error) {
    return nil, 0, nil
}

// Helper to create service with mocks
func newMockAuditService() (*audit.AuditService, *mockAuditRepo) {
    repo := &mockAuditRepo{}
    // Use real/simple implementations for other deps or mock them if needed
    redactor := &mockRedactor{} 
    idGen := &mockIDGenerator{}
    
    return audit.NewAuditService(repo, redactor, idGen), repo
}
```

Test that audit is recorded by checking the mock repository:

```go
func TestCreateOrder_RecordsAuditEvent(t *testing.T) {
    mockAuditService, mockRepo := newMockAuditService()
    
    // Inject mockAuditService into your use case
    useCase := NewCreateOrderUseCase(..., mockAuditService, ...)
    
    // ... execute use case ...
    
    // FAST ASSERTION: Verify event was recorded in repo
    assert.NotEmpty(t, mockRepo.events)
    event := mockRepo.events[0]
    
    assert.Equal(t, domain.EventOrderCreated, event.EventType)
    assert.Equal(t, "order", event.EntityType)
}
```

## Checklist

Before completing your audit integration:

- [ ] Event type constant defined in `internal/domain/audit.go`
- [ ] AuditService injected as use case dependency
- [ ] Request struct includes `RequestID` and `ActorID` fields
- [ ] `auditService.Record()` called within transaction
- [ ] Handler extracts and passes `RequestID` from context
- [ ] Handler extracts and passes `ActorID` from JWT claims
- [ ] Unit tests verify `Record()` is called with correct input
- [ ] Integration works end-to-end (event appears in `audit_events` table)
- [ ] `make lint` passes
- [ ] `make test` passes

## Reference Implementation

See these files for a complete working example:

- **Domain Constants:** `internal/domain/audit.go` - Event type definitions
- **Use Case:** `internal/app/user/create_user.go` - Full integration pattern
- **Tests:** `internal/app/user/create_user_test.go` - Audit mocking
- **Handler:** `internal/transport/http/handler/user_handler.go` - Context extraction
- **Service:** `internal/app/audit/service.go` - AuditService implementation

## Architecture Notes

### Why RequestID/ActorID in Request Struct?

The app layer cannot import the transport layer (hexagonal architecture). Context values from HTTP (request ID, JWT claims) must be extracted by the handler and passed via the request struct.

### Why Record Within Transaction?

Audit events must be recorded atomically with business operations. If the business operation rolls back, the audit event should also roll back. This ensures data consistency.

### PII Redaction

The `Payload` field is automatically redacted before storage. Fields with struct tags like `json:"password"` or containing sensitive data are masked. See `internal/shared/redact/redactor.go` for details.
