# Story 6.4: Implement Audit Event Service

Status: Done

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
   - `requestID` is extracted from context *(Note: by transport layer, passed via input struct)*
   - `ActorID` is extracted from auth claims *(Note: by transport layer, passed via input struct)*
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

- [x] Task 1: Create AuditService in app layer (AC: #1, #2)
  - [x] 1.1 Create `internal/app/audit/service.go`
  - [x] 1.2 Define `AuditEventInput` struct with fields for creating audit events
  - [x] 1.3 Implement `AuditService` struct with dependencies: `AuditEventRepository`, `Redactor`, `IDGenerator`
  - [x] 1.4 Implement `Record(ctx, q, input) error` method:
    - Use requestID and actorID from input struct (passed by transport layer)
    - Apply PII redaction via `s.redactor.Redact()` + json.Marshal()
    - Generate new ID via IDGenerator
    - Create `domain.AuditEvent` and validate it
    - Call `repository.Create(ctx, q, event)`
  - [x] 1.5 Implement `ListByEntity(ctx, q, entityType, entityID, params)` method (delegates to repository)

- [x] Task 2: Pass context values through input struct (AC: #2)
  - [x] 2.1 Add `RequestID string` field to `AuditEventInput` struct
  - [x] 2.2 ActorID already exists in input - caller provides it
  - [x] 2.3 NOTE: Transport layer extracts from context, passes to app layer
  - [x] 2.4 NOTE: App layer does NOT import transport - proper hexagonal architecture

- [x] Task 3: Integrate AuditService with CreateUserUseCase (AC: #4)
  - [x] 3.1 Add `AuditService` dependency to `CreateUserUseCase`
  - [x] 3.2 Modify `NewCreateUserUseCase` constructor to accept `AuditService`
  - [x] 3.3 Update `Execute` to record audit event after user creation (same transaction)
  - [x] 3.4 Update handler/DI wiring to pass AuditService

- [x] Task 4: Write unit tests for AuditService (AC: all)
  - [x] 4.1 Create `internal/app/audit/service_test.go`
  - [x] 4.2 Test `Record` with valid input - verify redaction called, event created
  - [x] 4.3 Test `Record` with repository error - verify error propagation
  - [x] 4.4 Test `Record` uses requestID and actorID from input struct correctly
  - [x] 4.5 Test `ListByEntity` delegates correctly to repository
  - [x] 4.6 Mock `domain.AuditEventRepository`, `domain.Redactor`, `domain.IDGenerator`
  - [x] 4.7 Achieve ≥80% coverage for new code

- [x] Task 5: Update CreateUserUseCase tests (AC: #3, #4)
  - [x] 5.1 Update existing tests to mock AuditService
  - [x] 5.2 Add test verifying audit event is recorded on successful user creation
  - [x] 5.3 Add test verifying error propagation when audit fails

- [x] Task 6: Verify layer compliance and integration (AC: implicit)
  - [x] 6.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 6.2 Run `make test` to ensure all unit tests pass
  - [x] 6.3 Run `make ci` (or `ALLOW_DIRTY=1 make ci`) for full CI check

## Dependencies & Blockers

- **Hard dependency:** Story 6.3 (PII Redaction Service) - **DONE**
- **Hard dependency:** Story 6.2 (Audit Event PostgreSQL Repository) - **DONE**
- **Hard dependency:** Story 6.1 (Audit Event Domain Model) - **DONE**
- Story 6.5 will document how to extend for new modules

## Assumptions & Open Questions

- `AuditService` is placed in `internal/app/audit/` following the user use case pattern
- Transport layer (handler) extracts requestID and actorID from context, passes via input structs
- App layer does NOT import transport utilities - proper hexagonal architecture
- The `AuditEventInput` struct provides a simpler API than constructing `domain.AuditEvent` directly

## Definition of Done

- `AuditService` created in `internal/app/audit/` with `Record` and `ListByEntity` methods
- RequestID and ActorID passed via `AuditEventInput` struct (proper hexagonal architecture)
- PII redaction applied via injected `domain.Redactor` interface
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
    "encoding/json"
    "time"

    "github.com/iruldev/golang-api-hexagonal/internal/app"
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

type AuditEventInput struct {
    EventType  string    // "user.created" - use domain constants
    ActorID    domain.ID // Who performed action (empty for system/unauthenticated)
    EntityType string    // "user"
    EntityID   domain.ID // The affected entity's ID
    Payload    any       // Will be redacted and marshaled to JSON
    RequestID  string    // From transport layer context extraction
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

### Context Value Extraction Pattern

Per hexagonal architecture, the **transport layer** extracts context values and passes them to the app layer. The app layer does NOT import transport utilities.

**In the HTTP handler (transport layer):**
```go
// internal/transport/http/handler/user.go
import (
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
    "github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Transport layer extracts context values
    requestID := middleware.GetRequestID(r.Context())
    var actorID domain.ID
    if claims := ctxutil.GetClaims(r.Context()); claims != nil {
        actorID = domain.ID(claims.Subject)
    }
    
    // Pass to use case via request struct
    req := user.CreateUserRequest{
        // ... user fields ...
        RequestID: requestID,
        ActorID:   actorID,
    }
    resp, err := h.createUserUseCase.Execute(r.Context(), req)
}
```

**NOTE:** App layer receives values via input structs - never imports transport layer!

### Integration with CreateUserUseCase

```go
// internal/app/user/create_user.go (modified)
// NOTE: App layer only imports domain and app - NO transport imports!
import (
    "github.com/iruldev/golang-api-hexagonal/internal/app/audit"
)

// Add RequestID and ActorID to request struct
type CreateUserRequest struct {
    ID        domain.ID
    FirstName string
    LastName  string
    Email     string
    RequestID string    // From transport layer
    ActorID   domain.ID // From transport layer (JWT claims)
}

type CreateUserUseCase struct {
    userRepo     domain.UserRepository
    auditService *audit.AuditService  // ADD THIS
    idGen        domain.IDGenerator
    db           domain.Querier
}

func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
    // ... existing user creation logic ...
    
    // After successful user creation, record audit event
    // RequestID and ActorID come from the request struct (passed by handler)
    auditInput := audit.AuditEventInput{
        EventType:  domain.EventUserCreated,
        ActorID:    req.ActorID,    // From request
        EntityType: "user",
        EntityID:   user.ID,
        Payload:    user,
        RequestID:  req.RequestID,  // From request
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

Anthropic Claude (Antigravity)

### Debug Log References

N/A

### Completion Notes List

- Implemented `AuditService` in `internal/app/audit/service.go` with `Record()` and `ListByEntity()` methods
- `AuditEventInput` struct provides simplified API for creating audit events
- PII redaction applied via `domain.Redactor` interface before persisting
- RequestID and ActorID passed via input struct (hexagonal architecture - app layer does not import transport)
- Integrated with `CreateUserUseCase` to record audit events on user creation
- Unit tests achieve 92.3% coverage for audit service code
- All acceptance criteria met: AC#1 (service exists), AC#2 (redaction + context values), AC#3 (error propagation), AC#4 (CreateUser integration)

### Change Log

- 2025-12-22: Implemented Story 6.4 - Audit Event Service
  - Created `internal/app/audit/service.go` with `AuditService`, `AuditEventInput`
  - Created `internal/app/audit/service_test.go` with 14 unit tests
  - Modified `internal/app/user/create_user.go` to integrate audit recording
  - Modified `internal/app/user/create_user_test.go` with audit mocks and new tests
  - Modified `cmd/api/main.go` to wire AuditService dependencies
 - 2025-12-22: Review Fix - Transactional Integrity
   - Updated `CreateUserUseCase` to use `domain.TxManager`
   - Ensured user creation and audit recording occur in same transaction
   - Updated `create_user_test.go` and `main.go` for new dependency

### File List

**New Files:**
- `internal/app/audit/service.go` - AuditService implementation
- `internal/app/audit/service_test.go` - Unit tests for AuditService

**Modified Files:**
- `internal/app/user/create_user.go` - Added AuditService dependency, RequestID/ActorID to request struct, audit event recording
- `internal/app/user/create_user_test.go` - Added mock audit service, new tests for audit recording
- `cmd/api/main.go` - Added audit service DI wiring
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status
- `docs/sprint-artifacts/6-4-implement-audit-event-service.md` - Updated tasks and status
