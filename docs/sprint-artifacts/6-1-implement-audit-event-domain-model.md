Status: done

## Story

As a **developer**,
I want **audit event entity and interfaces in domain layer**,
so that **I have a clear contract for audit logging**.

## Acceptance Criteria

1. **Given** the domain layer, **When** I view `internal/domain/audit.go`, **Then** `AuditEvent` entity exists with fields:
   - `ID` (type ID)
   - `EventType` (string, pattern: "entity.action", e.g., "user.created")
   - `ActorID` (ID, nullable for system/unauthenticated events)
   - `EntityType` (string, e.g., "user")
   - `EntityID` (ID, what was affected)
   - `Payload` ([]byte, JSON, already redacted)
   - `Timestamp` (time.Time)
   - `RequestID` (string, for correlation)

2. **And** `AuditEventRepository` interface is defined:
   - `Create(ctx context.Context, q Querier, event *AuditEvent) error`
   - `ListByEntityID(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)`

3. **And** domain layer has NO external imports (stdlib only)

*Covers: FR35 (partial)*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 6.1".
- Domain patterns established in `internal/domain/user.go` (entity pattern, repository interface).
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Create AuditEvent entity (AC: #1)
  - [x] 1.1 Create `internal/domain/audit.go`
  - [x] 1.2 Define `AuditEvent` struct with all required fields
  - [x] 1.3 Use `domain.ID` for ID, ActorID, EntityID (NOT uuid.UUID)
  - [x] 1.4 Use `time.Time` for Timestamp
  - [x] 1.5 Use `[]byte` for Payload (JSON bytes)
  - [x] 1.6 Add struct comments explaining each field's purpose

- [x] Task 2: Define AuditEventRepository interface (AC: #2)
  - [x] 2.1 Add `AuditEventRepository` interface to `internal/domain/audit.go`
  - [x] 2.2 Define `Create(ctx context.Context, q Querier, event *AuditEvent) error`
  - [x] 2.3 Define `ListByEntityID(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)`
  - [x] 2.4 Add method comments documenting behavior

- [x] Task 3: Add domain sentinel errors for audit (AC: #1)
  - [x] 3.1 Add `ErrAuditEventNotFound = errors.New("audit event not found")` to `internal/domain/errors.go`
  - [x] 3.2 Add any other audit-related domain errors if needed

- [x] Task 4: Add event type constants (optional, recommended)
  - [x] 4.1 Define standard event type constants: `const EventUserCreated = "user.created"` etc.
  - [x] 4.2 Add documentation comment about "entity.action" naming pattern

- [x] Task 5: Write unit tests (AC: all)
  - [x] 5.1 Create `internal/domain/audit_test.go`
  - [x] 5.2 Test AuditEvent struct can be instantiated with all fields
  - [x] 5.3 Test any validation method if created
  - [x] 5.4 Achieve ≥80% coverage for new code

- [x] Task 6: Verify layer compliance (AC: #3)
  - [x] 6.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 6.2 Run `make test` to ensure all tests pass
  - [x] 6.3 Confirm only stdlib imports in `internal/domain/audit.go`

## Dependencies & Blockers

- No blockers - this is the first story of Epic 6
- Uses existing domain patterns from Epic 4 (User entity, repository interfaces)
- Depends on existing `domain.ID`, `domain.Querier`, `domain.ListParams`

## Assumptions & Open Questions

- `ActorID` is optional (nullable) for system-initiated or unauthenticated events
- `Payload` is pre-redacted JSON bytes (redaction handled by app layer, Story 6.3)
- `RequestID` is a string (not domain.ID) since it comes from transport layer
- Event type uses dot notation: "entity.action" (e.g., "user.created", "user.updated")

## Definition of Done

- `AuditEvent` struct created in `internal/domain/audit.go`
- `AuditEventRepository` interface defined in same file
- All fields match acceptance criteria exactly
- Only stdlib imports in domain layer (depguard passes)
- Sentinel errors added for audit events
- Unit tests pass with ≥80% coverage
- Lint passes (layer compliance verified)

## Non-Functional Requirements

- No external dependencies in domain layer (stdlib only)
- Code follows Go idioms and project conventions
- Comments document all exported types and methods

## Testing & Coverage

- Unit tests for AuditEvent struct instantiation
- Unit tests for any validation methods
- Verify no external imports (lint check)
- Aim for coverage ≥80% for new domain code

## Dev Notes

### ⚠️ CRITICAL: Domain Layer Purity

The domain layer must have **NO external imports**. This is enforced by depguard in CI.

```
✅ ALLOWED: stdlib only (errors, context, time, strings, fmt)
❌ FORBIDDEN: slog, uuid, pgx, chi, otel, ANY external package
```

### Existing Code Context

**From Existing Domain Layer:**
| File | Description |
|------|-------------|
| `internal/domain/user.go` | Reference pattern for entity + repository interface |
| `internal/domain/errors.go` | Domain sentinel errors pattern |
| `internal/domain/id.go` | `type ID string` - use this for IDs |
| `internal/domain/querier.go` | Querier interface for repository methods |
| `internal/domain/pagination.go` | ListParams for pagination |
| `internal/domain/tx.go` | TxManager interface for transactions |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/domain/audit.go` | AuditEvent entity + AuditEventRepository interface |
| `internal/domain/audit_test.go` | Unit tests for audit domain |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/domain/errors.go` | Add audit-related sentinel errors |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### Architecture Constraints (depguard enforced)

```
Domain layer imports allowed:
- "context"
- "errors"
- "time"
- "strings"
- "fmt"

Domain layer imports FORBIDDEN:
- github.com/google/uuid     ❌
- log/slog                   ❌
- github.com/jackc/pgx       ❌
- ANY third-party package    ❌
```

### AuditEvent Entity Pattern

Follow the existing `User` entity pattern from `internal/domain/user.go`:

```go
// internal/domain/audit.go
package domain

import (
    "time"
)

// AuditEvent represents an audit trail entry for tracking business operations.
// This entity follows hexagonal architecture principles - no external dependencies.
//
// EventType follows the pattern "entity.action", for example:
//   - "user.created"
//   - "user.updated"
//   - "user.deleted"
//   - "order.created"
//
// The Payload field contains pre-redacted JSON data. PII redaction is handled
// by the app layer BEFORE creating the AuditEvent.
type AuditEvent struct {
    // ID is the unique identifier for this audit event.
    ID ID

    // EventType describes what happened, in "entity.action" format.
    EventType string

    // ActorID identifies who performed the action.
    // Empty for system-initiated or unauthenticated operations.
    ActorID ID

    // EntityType identifies the type of entity affected (e.g., "user", "order").
    EntityType string

    // EntityID identifies which specific entity was affected.
    EntityID ID

    // Payload contains JSON-encoded event data with PII already redacted.
    Payload []byte

    // Timestamp records when the event occurred.
    Timestamp time.Time

    // RequestID correlates this event with the originating HTTP request.
    RequestID string
}
```

### AuditEventRepository Interface Pattern

Follow the existing `UserRepository` interface pattern:

```go
// AuditEventRepository defines the interface for audit event persistence operations.
// This interface is defined in the domain layer and implemented by the infrastructure layer.
// All methods accept a Querier to support both connection pool and transaction usage.
type AuditEventRepository interface {
    // Create stores a new audit event.
    Create(ctx context.Context, q Querier, event *AuditEvent) error

    // ListByEntityID retrieves audit events for a specific entity.
    // Results are ordered by timestamp DESC (newest first).
    // Returns the slice of events, total count of matching events, and any error.
    ListByEntityID(ctx context.Context, q Querier, entityType string, entityID ID, params ListParams) ([]AuditEvent, int, error)
}
```

### Sentinel Errors Pattern

Add to `internal/domain/errors.go`:

```go
// ErrAuditEventNotFound is returned when an audit event cannot be found.
ErrAuditEventNotFound = errors.New("audit event not found")

// ErrInvalidEventType is returned when the event type format is invalid.
ErrInvalidEventType = errors.New("invalid event type format")
```

### Event Type Constants (Recommended)

Consider adding standard event type constants for documentation and consistency:

```go
// Standard event types for the Users module.
// New modules should follow the "entity.action" naming pattern.
const (
    // EventUserCreated is recorded when a new user is created.
    EventUserCreated = "user.created"
    
    // EventUserUpdated is recorded when a user is updated.
    EventUserUpdated = "user.updated"
    
    // EventUserDeleted is recorded when a user is deleted.
    EventUserDeleted = "user.deleted"
)
```

### Validation Method (Optional)

Consider adding a validation method similar to User.Validate():

```go
// Validate checks if the AuditEvent has required fields.
func (e AuditEvent) Validate() error {
    if e.EventType == "" {
        return ErrInvalidEventType
    }
    if e.EntityType == "" {
        return errors.New("entity type is required")
    }
    if e.EntityID.IsEmpty() {
        return errors.New("entity ID is required")
    }
    return nil
}
```

### Unit Test Pattern

Follow existing test patterns from `internal/domain/user_test.go`:

```go
// internal/domain/audit_test.go
package domain_test

import (
    "testing"
    "time"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/stretchr/testify/assert"
)

func TestAuditEvent_Fields(t *testing.T) {
    event := domain.AuditEvent{
        ID:         domain.ID("test-id"),
        EventType:  "user.created",
        ActorID:    domain.ID("actor-id"),
        EntityType: "user",
        EntityID:   domain.ID("entity-id"),
        Payload:    []byte(`{"email":"[REDACTED]"}`),
        Timestamp:  time.Now(),
        RequestID:  "req-123",
    }

    assert.Equal(t, domain.ID("test-id"), event.ID)
    assert.Equal(t, "user.created", event.EventType)
    assert.Equal(t, domain.ID("actor-id"), event.ActorID)
    // ... test all fields
}

func TestAuditEvent_Validate(t *testing.T) {
    tests := []struct {
        name    string
        event   domain.AuditEvent
        wantErr bool
    }{
        {
            name: "valid event",
            event: domain.AuditEvent{
                EventType:  "user.created",
                EntityType: "user",
                EntityID:   domain.ID("123"),
            },
            wantErr: false,
        },
        {
            name: "missing event type",
            event: domain.AuditEvent{
                EntityType: "user",
                EntityID:   domain.ID("123"),
            },
            wantErr: true,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.event.Validate()
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
make ci

# Check specific domain tests
go test -v ./internal/domain/...

# Check coverage
go test -cover ./internal/domain/...
```

### References

- [Source: docs/epics.md#Story 6.1] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Communication Patterns] - Audit event structure
- [Source: docs/project-context.md#Domain Layer] - Layer constraints
- [Source: internal/domain/user.go] - Entity and repository interface patterns
- [Source: internal/domain/errors.go] - Sentinel error patterns

### Learnings from Previous Epics

**Critical Patterns to Follow:**
1. **Domain Purity:** ONLY stdlib imports - no uuid, slog, or external packages
2. **ID Type:** Use `domain.ID` (string type alias) for all identifiers
3. **Repository Pattern:** Interface accepts `Querier` to work with pool or tx
4. **Sentinel Errors:** Use `var Err... = errors.New(...)` pattern
5. **Comments:** Document all exported types and methods

**From Epic 4 (Users Module):**
- Entity structs contain data fields, minimal logic
- Repository interfaces define persistence contracts
- Use `ListParams` for pagination parameters
- Return `(slice, totalCount, error)` for list operations

**From Epic 5 (Security):**
- Testing patterns with table-driven tests
- Coverage targets of ≥80%

### Security Considerations

1. **Payload is pre-redacted:** The domain layer assumes Payload already has PII removed
2. **ActorID nullable:** Empty ActorID must be handled properly (system/anon events)
3. **No logging in domain:** Absolutely no logging or tracing in domain layer

### Epic 6 Context

Epic 6 implements the Audit Trail System for compliance requirements:
- **6.1 (this story):** Domain model (entity + repository interface)
- **6.2:** PostgreSQL repository implementation
- **6.3:** PII redaction service
- **6.4:** Audit event service (app layer)
- **6.5:** Extensible event types

This story establishes the foundation that all subsequent stories build upon.

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-18)

- `docs/epics.md` - Story 6.1 acceptance criteria
- `docs/architecture.md` - Audit event patterns, repository pattern
- `docs/project-context.md` - Domain layer conventions, layer rules
- `docs/sprint-artifacts/5-5-implement-rate-limiting-middleware.md` - Previous story format
- `internal/domain/user.go` - Entity pattern reference
- `internal/domain/errors.go` - Sentinel error pattern

### Agent Model Used

Gemini 2.5

### Debug Log References

N/A

### Completion Notes List

- ✅ Created `internal/domain/audit.go` with AuditEvent entity following User entity patterns
- ✅ Defined AuditEventRepository interface with Create and ListByEntityID methods
- ✅ Added event type constants (EventUserCreated, EventUserUpdated, EventUserDeleted) with "entity.action" pattern
- ✅ Added Validate() method for AuditEvent to check required fields
- ✅ Added sentinel errors: ErrAuditEventNotFound, ErrInvalidEventType, ErrInvalidEntityType, ErrInvalidEntityID
- ✅ Created comprehensive unit tests with 100% coverage for domain layer
- ✅ All tests pass (make test successful)
- ✅ Lint passes with 0 issues (depguard layer compliance verified)
- ✅ All CI checks pass (ALLOW_DIRTY=1 make ci)
- ✅ Only stdlib imports in domain layer (context, strings, time)

### Change Log

- 2025-12-18: Implemented audit event domain model (Story 6.1)
- 2025-12-19: Reviewed and fixed validation gaps for ID, Timestamp, and Payload (Code Review)

### File List

**New Files:**
- `internal/domain/audit.go` - AuditEvent entity, AuditEventRepository interface, event type constants
- `internal/domain/audit_test.go` - Comprehensive unit tests (100% coverage)

**Modified Files:**
- `internal/domain/errors.go` - Added audit-related sentinel errors
- `docs/sprint-artifacts/sprint-status.yaml` - Updated status to in-progress, then review, then done

## Senior Developer Review (AI)

**Review Date:** 2025-12-19
**Reviewer:** Chat (AI)

### Findings
- 🔴 **CRITICAL**: `Validate()` originally allowed `ID` and `Timestamp` to be empty.
- 🟡 **MEDIUM**: `Validate()` allowed `Payload` to be nil.
- 🟢 **LOW**: `EventUser*` constants introduce domain coupling (accepted for now).

### Actions Taken
- [x] Updated `AuditEvent.Validate()` to enforce `ID`, `Timestamp`, and `Payload` presence.
- [x] Added `ErrInvalidID`, `ErrInvalidTimestamp`, `ErrInvalidPayload` sentinel errors.
- [x] Updated `internal/domain/audit_test.go` to cover new validation rules.
- [x] Verified all tests pass.

**Outcome:** Approved (fixes verified on rerun).
