# Story 6.4: Implement Audit Event Service

Status: ready-for-dev

## Story

As a **developer**,
I want **a service to record and query audit events synchronously**,
so that **business operations are tracked for compliance**.

## Acceptance Criteria

1. **Given** the app layer, **When** I view `internal/app/audit/`, **Then** `AuditService` exists with methods:
   - `Record(ctx context.Context, q Querier, event AuditEventInput) error`
   - `ListByEntity(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)`

2. **Given** `Record` is called within a transaction, **When** event is processed, **Then**:
   - PII redaction is applied to payload via `domain.Redactor`
   - `requestID` is extracted from context
   - `ActorID` is extracted from auth claims (empty string if no auth/unauthenticated)
   - Event is persisted via `AuditEventRepository` in the SAME transaction (passed Querier)

3. **Given** audit insert fails, **When** business transaction attempts to commit, **Then**:
   - Transaction is rolled back
   - API returns 500 with Code="INTERNAL_ERROR"

4. **Given** user creates a new user via API, **When** `CreateUserUseCase` completes successfully, **Then** audit event with type="user.created" is recorded in same transaction.

*Covers: FR35, FR38*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 6.4".
- App layer patterns established in `internal/app/user/` use cases.
- Error patterns established in `internal/app/errors.go`.
- Domain interfaces in `internal/domain/audit.go` and `internal/domain/redactor.go`.
- PII redaction implementation in `internal/shared/redact/redactor.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [ ] Task 1: Create AuditService in app layer (AC: #1, #2)
  - [ ] 1.1 Create `internal/app/audit/service.go`
  - [ ] 1.2 Define `AuditEventInput` struct with fields for creating audit events
  - [ ] 1.3 Implement `AuditService` struct with dependencies: `AuditEventRepository`, `Redactor`, `IDGenerator`
  - [ ] 1.4 Implement `Record(ctx, q, input) error` method:
    - Extract requestID from context (use helper function)
    - Extract actorID from auth context (empty if not authenticated)
    - Apply PII redaction via `redact.RedactAndMarshal()`
    - Generate new ID via IDGenerator
    - Create `domain.AuditEvent` and validate it
    - Call `repository.Create(ctx, q, event)`
  - [ ] 1.5 Implement `ListByEntity(ctx, q, entityType, entityID, params)` method (delegates to repository)

- [ ] Task 2: Create context helpers for request/actor extraction (AC: #2)
  - [ ] 2.1 Create `internal/app/audit/context.go` (or add to service.go)
  - [ ] 2.2 Define `ContextKey` type and constants for request ID and actor ID
  - [ ] 2.3 Implement `GetRequestIDFromContext(ctx) string` helper
  - [ ] 2.4 Implement `GetActorIDFromContext(ctx) domain.ID` helper
  - [ ] 2.5 NOTE: These may need to integrate with existing middleware context utilities

- [ ] Task 3: Integrate AuditService with CreateUserUseCase (AC: #4)
  - [ ] 3.1 Add `AuditService` dependency to `CreateUserUseCase`
  - [ ] 3.2 Modify `NewCreateUserUseCase` constructor to accept `AuditService`
  - [ ] 3.3 Update `Execute` to record audit event after user creation (same transaction)
  - [ ] 3.4 Update handler/DI wiring to pass AuditService

- [ ] Task 4: Write unit tests for AuditService (AC: all)
  - [ ] 4.1 Create `internal/app/audit/service_test.go`
  - [ ] 4.2 Test `Record` with valid input - verify redaction called, event created
  - [ ] 4.3 Test `Record` with repository error - verify error propagation
  - [ ] 4.4 Test `Record` extracts requestID and actorID from context
  - [ ] 4.5 Test `ListByEntity` delegates correctly to repository
  - [ ] 4.6 Mock `domain.AuditEventRepository`, `domain.Redactor`, `domain.IDGenerator`
  - [ ] 4.7 Achieve ≥80% coverage for new code

- [ ] Task 5: Update CreateUserUseCase tests (AC: #3, #4)
  - [ ] 5.1 Update existing tests to mock AuditService
  - [ ] 5.2 Add test verifying audit event is recorded on successful user creation
  - [ ] 5.3 Add test verifying error propagation when audit fails

- [ ] Task 6: Verify layer compliance and integration (AC: implicit)
  - [ ] 6.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [ ] 6.2 Run `make test` to ensure all unit tests pass
  - [ ] 6.3 Run `make ci` (or `ALLOW_DIRTY=1 make ci`) for full CI check

## Dependencies & Blockers

- **Hard dependency:** Story 6.3 (PII Redaction Service) - **DONE**
- **Hard dependency:** Story 6.2 (Audit Event PostgreSQL Repository) - **DONE**
- **Hard dependency:** Story 6.1 (Audit Event Domain Model) - **DONE**
- Story 6.5 will document how to extend for new modules

## Assumptions & Open Questions

- `AuditService` is placed in `internal/app/audit/` following the user use case pattern
- Request ID extraction uses existing middleware utilities (check `internal/transport/http/middleware/`)
- Actor ID extraction uses existing auth context (check JWT middleware implementation)
- If context helpers already exist, reuse them; otherwise create in audit package
- The `AuditEventInput` struct provides a simpler API than constructing `domain.AuditEvent` directly

## Definition of Done

- `AuditService` created in `internal/app/audit/` with `Record` and `ListByEntity` methods
- Context helpers extract requestID and actorID from context
- PII redaction applied via shared redaction service
- `CreateUserUseCase` integrated with audit recording
- Unit tests pass with ≥80% coverage
- `make lint` passes (layer compliance verified)
- `make ci` passes

## Non-Functional Requirements

- App layer: domain imports only (no net/http, pgx, slog, uuid, transport, infra)
- Use `domain.Redactor` interface (not concrete `PIIRedactor`)
- Use `domain.IDGenerator` for ID generation
- Use `domain.Querier` for database operations (transaction support)
- Performance: O(1) for Record (single DB insert)

## Testing & Coverage

- Unit tests with table-driven test style
- Mock all dependencies using interfaces
- Test successful record flow
- Test error propagation
- Test context value extraction
- Aim for coverage ≥80% for new audit service code

## Dev Notes

### ⚠️ CRITICAL: Layer Rules

The audit service MUST follow app layer rules:

```
✅ internal/app/audit/
   - Can import: domain, shared/redact (via domain.Redactor interface)
   - Cannot import: net/http, pgx, slog, uuid, transport, infra

✅ Dependency Injection:
   - Receive domain.Redactor interface (not concrete *PIIRedactor)
   - Receive domain.AuditEventRepository interface
   - Receive domain.IDGenerator interface
```

### Existing Code Context

**From Story 6.1, 6.2, 6.3 (DONE):**
| File | Description |
|------|-------------|
| `internal/domain/audit.go` | AuditEvent entity with `Payload []byte`, Validate() method, event type constants |
| `internal/domain/redactor.go` | Redactor interface, RedactorConfig, EmailMode constants |
| `internal/infra/postgres/audit_event_repo.go` | Repository implementation with Create, ListByEntityID |
| `internal/shared/redact/redactor.go` | PIIRedactor implementation, RedactAndMarshal helper |

**Reference App Layer Pattern (CreateUserUseCase):**
| File | Description |
|------|-------------|
| `internal/app/user/create_user.go` | Use case pattern with Request/Response structs |
| `internal/app/errors.go` | AppError type with Code field |
| `internal/domain/id_generator.go` | IDGenerator interface |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/app/audit/service.go` | AuditService with Record, ListByEntity |
| `internal/app/audit/context.go` | Context helpers (if needed) |
| `internal/app/audit/service_test.go` | Unit tests |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/app/user/create_user.go` | Add AuditService dependency, record audit event |
| `internal/app/user/create_user_test.go` | Update tests to mock AuditService |
| `cmd/api/main.go` (or DI setup) | Wire AuditService dependencies |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### AuditEventInput Design

Instead of requiring caller to construct full `domain.AuditEvent`, provide simpler input:

```go
// internal/app/audit/service.go
package audit

import (
    "context"
    "time"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
)

// AuditEventInput represents the input for recording an audit event.
// The service handles ID generation, timestamp, requestID extraction, and PII redaction.
type AuditEventInput struct {
    EventType  string    // "user.created" - use domain constants
    ActorID    domain.ID // Who performed action (empty for system/unauthenticated)
    EntityType string    // "user"
    EntityID   domain.ID // The affected entity's ID
    Payload    any       // Will be redacted and marshaled to JSON
}

// AuditService provides audit event recording and querying capabilities.
type AuditService struct {
    repo    domain.AuditEventRepository
    redactor domain.Redactor
    idGen   domain.IDGenerator
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
// PII fields in the payload are automatically redacted.
// RequestID is extracted from context.
func (s *AuditService) Record(ctx context.Context, q domain.Querier, input AuditEventInput) error {
    op := "AuditService.Record"
    
    // Redact and marshal payload
    payload, err := redact.RedactAndMarshal(s.redactor, input.Payload)
    if err != nil {
        return &app.AppError{
            Op:      op,
            Code:    app.CodeInternalError,
            Message: "Failed to redact payload",
            Err:     err,
        }
    }
    
    // Extract requestID from context (implement helper)
    requestID := GetRequestIDFromContext(ctx)
    
    // Create domain event
    event := &domain.AuditEvent{
        ID:         s.idGen.NewID(),
        EventType:  input.EventType,
        ActorID:    input.ActorID,
        EntityType: input.EntityType,
        EntityID:   input.EntityID,
        Payload:    payload,
        Timestamp:  time.Now().UTC(),
        RequestID:  requestID,
    }
    
    // Validate event
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
func (s *AuditService) ListByEntity(
    ctx context.Context,
    q domain.Querier,
    entityType string,
    entityID domain.ID,
    params domain.ListParams,
) ([]domain.AuditEvent, int, error) {
    return s.repo.ListByEntityID(ctx, q, entityType, entityID, params)
}
```

### Context Helper Pattern

Check existing middleware for request ID context. Likely in `internal/transport/http/middleware/`. If a helper already exists, import and use it. Otherwise, create minimal helpers:

```go
// internal/app/audit/context.go
package audit

import "context"

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
    // RequestIDKey is the context key for request ID.
    // This should match the key used by request ID middleware.
    RequestIDKey ContextKey = "requestId"
)

// GetRequestIDFromContext extracts the request ID from context.
// Returns empty string if not present.
func GetRequestIDFromContext(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return ""
}
```

**NOTE:** Verify the actual context key used by existing middleware before implementing!

### Integration with CreateUserUseCase

```go
// internal/app/user/create_user.go (modified)
type CreateUserUseCase struct {
    userRepo     domain.UserRepository
    auditService *audit.AuditService  // ADD THIS
    idGen        domain.IDGenerator
    db           domain.Querier
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
    // ... existing user creation logic ...
    
    // After successful user creation, record audit event
    auditInput := audit.AuditEventInput{
        EventType:  domain.EventUserCreated,
        ActorID:    GetActorIDFromContext(ctx), // From JWT claims
        EntityType: "user",
        EntityID:   user.ID,
        Payload:    user, // Will be redacted by service
    }
    
    if err := uc.auditService.Record(ctx, uc.db, auditInput); err != nil {
        return CreateUserResponse{}, err
    }
    
    return CreateUserResponse{User: *user}, nil
}
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
ALLOW_DIRTY=1 make ci

# Check coverage for app layer
go test -cover ./internal/app/...
```

### References

- [Source: docs/epics.md#Story 6.4] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Implementation Patterns] - Layer rules
- [Source: docs/project-context.md] - Critical layer rules and conventions
- [Source: internal/domain/audit.go] - AuditEvent entity and repository interface
- [Source: internal/domain/redactor.go] - Redactor interface
- [Source: internal/shared/redact/redactor.go] - RedactAndMarshal helper function
- [Source: internal/app/user/create_user.go] - Use case pattern reference
- [Source: internal/app/errors.go] - AppError pattern

### Learnings from Previous Stories

**From Story 6.3 (PII Redaction Service):**
1. `RedactAndMarshal()` accepts any input (map, struct, []byte) and returns `[]byte`
2. Use `domain.Redactor` interface, not concrete `*PIIRedactor`
3. Redaction creates new data - never modifies original

**From Story 6.1 & 6.2 (Domain + Repository):**
1. `AuditEvent.Payload` is `[]byte` - pre-redacted JSON
2. Domain entity has `Validate()` method - call before persisting
3. Repository accepts `Querier` for transaction support
4. Use domain event type constants: `domain.EventUserCreated`

**From User Module (Epic 4):**
1. Use case receives dependencies via constructor
2. Return `*app.AppError` with appropriate `Code`
3. Use `domain.IDGenerator` for ID generation
4. Extract context values for cross-cutting concerns

### Security Considerations

1. **PII always redacted:** Use redactor before storing payload
2. **Transaction safety:** Audit failure rolls back entire transaction
3. **No sensitive data leakage:** AppError messages are generic
4. **Actor tracking:** Extract authenticated user ID from JWT claims

### Epic 6 Context

Epic 6 implements the Audit Trail System for compliance requirements:
- **6.1 (DONE):** Domain model (entity + repository interface)
- **6.2 (DONE):** PostgreSQL repository implementation
- **6.3 (DONE):** PII redaction service
- **6.4 (this story):** Audit event service (app layer)
- **6.5:** Extensible event types (documentation)

This story provides the app-layer service that orchestrates audit event recording with PII redaction, integrating with the existing domain and infrastructure layers.

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-19)

- `docs/epics.md` - Story 6.4 acceptance criteria
- `docs/architecture.md` - Layer rules and patterns
- `docs/project-context.md` - Layer constraints and conventions
- `docs/sprint-artifacts/6-3-implement-pii-redaction-service.md` - Previous story
- `internal/domain/audit.go` - AuditEvent entity and repository interface
- `internal/domain/redactor.go` - Redactor interface
- `internal/shared/redact/redactor.go` - PIIRedactor implementation
- `internal/app/user/create_user.go` - Use case pattern reference
- `internal/app/errors.go` - AppError pattern

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

(To be filled by dev agent)

### Change Log

(To be filled by dev agent)

### File List

(To be filled by dev agent)
