# Story 6.5: Enable Extensible Audit Event Types

Status: Done

## Story

As a **developer**,
I want **to easily add new audit event types**,
so that **I can extend auditing for new modules**.

## Acceptance Criteria

1. **Given** a developer wants to add new module auditing, **When** they follow the audit pattern, **Then**:
   - They can define new event type constants: `const EventOrderCreated = "order.created"`
   - Call `auditService.Record()` from their use case within transaction
   - New events appear in `audit_events` table

2. **Given** the audit module, **When** I view code comments or package documentation, **Then**:
   - Example shows how to add audit events for new entity types
   - Pattern "entity.action" is documented

*Covers: FR39*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 6.5".
- Domain layer event constants in `internal/domain/audit.go`.
- App layer `AuditService` in `internal/app/audit/service.go`.
- CreateUserUseCase integration pattern in `internal/app/user/create_user.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Document event type naming conventions (AC: #2)
  - [x] 1.1 Add detailed GoDoc comments to `internal/domain/audit.go` explaining:
    - Event type naming pattern: "entity.action"
    - Standard action verbs: created, updated, deleted, etc.
    - How to add new event constants for new modules
  - [x] 1.2 Add examples in comments showing event type patterns for common scenarios

- [x] Task 2: Document AuditService usage pattern (AC: #2)
  - [x] 2.1 Enhance `internal/app/audit/service.go` package-level documentation with:
    - Step-by-step guide for adding audit events to a new module
    - Example showing complete integration pattern
    - Reference to CreateUserUseCase as canonical example
  - [x] 2.2 Add inline code examples in AuditEventInput struct documentation

- [x] Task 3: Create developer guide for extending audit (AC: #1, #2)
  - [x] 3.1 Create `docs/guides/adding-audit-events.md` with:
    - Overview of audit system architecture
    - Step-by-step instructions for adding audit to a new module
    - Complete code examples following the pattern
    - Common patterns: "entity.created", "entity.updated", "entity.deleted"
  - [x] 3.2 Include checklist for developers adding audit events:
    - Define event type constant in domain layer
    - Call auditService.Record() within transaction
    - Pass requestID/actorID via input struct
    - Test audit recording in use case tests

- [x] Task 4: Add example event types for extensibility (AC: #1)
  - [x] 4.1 Add placeholder event type constants in `internal/domain/audit.go`:
    - Add commented examples showing the pattern for other entities
    - Document that these serve as templates for new modules
  - [x] 4.2 Ensure documentation references these examples

- [x] Task 5: Verify documentation accuracy (AC: all)
  - [x] 5.1 Review all documentation against actual code implementation
  - [x] 5.2 Run `make lint` to ensure no errors introduced
  - [x] 5.3 Verify all code examples compile correctly

## Dependencies & Blockers

- **Hard dependency:** Story 6.4 (Audit Event Service) - **DONE**
- **Hard dependency:** Story 6.3 (PII Redaction Service) - **DONE**
- **Hard dependency:** Story 6.2 (Audit Event PostgreSQL Repository) - **DONE**
- **Hard dependency:** Story 6.1 (Audit Event Domain Model) - **DONE**

## Assumptions & Open Questions

- This is primarily a documentation story with minimal code changes
- The existing implementation (Stories 6.1-6.4) already supports extensibility
- Key documentation locations: domain layer comments, app layer comments, developer guide
- No new Go code features required - just enhanced documentation

## Definition of Done

- GoDoc comments in `internal/domain/audit.go` explain event type pattern
- GoDoc comments in `internal/app/audit/service.go` show usage pattern
- Developer guide `docs/guides/adding-audit-events.md` exists with complete examples
- All code examples in documentation are verified to compile
- `make lint` passes
- Pattern is clear enough that a developer can add auditing to a new module without asking questions

## Non-Functional Requirements

- Documentation must be clear and actionable
- Examples must be copy-paste ready
- All documentation must follow existing project conventions
- No breaking changes to existing code

## Testing & Coverage

- No new unit tests required (documentation story)
- Verify all code examples compile by visual inspection
- Manual review that documentation is clear and complete

## Dev Notes

### ⚠️ CRITICAL: This is a Documentation Story

This story focuses on **documentation quality**, not new functionality. The audit system is already fully implemented and extensible. The goal is to make it easy for developers to understand and extend.

### Existing Code Context

**From Stories 6.1-6.4 (DONE):**
| File | Description |
|------|-------------|
| `internal/domain/audit.go` | AuditEvent entity, event type constants, repository interface |
| `internal/app/audit/service.go` | AuditService with Record, ListByEntity methods |
| `internal/app/user/create_user.go` | Canonical example of audit integration in use case |
| `internal/infra/postgres/audit_event_repo.go` | Repository implementation |
| `internal/shared/redact/redactor.go` | PII redaction implementation |

**This story CREATES:**
| File | Description |
|------|-------------|
| `docs/guides/adding-audit-events.md` | Developer guide for extending audit system |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/domain/audit.go` | Enhanced GoDoc comments, example event types |
| `internal/app/audit/service.go` | Enhanced package and type documentation |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### Event Type Naming Pattern

The established pattern is `"entity.action"`:

```go
// internal/domain/audit.go

// Standard event types for the Users module.
// New modules should follow the "entity.action" naming pattern.
// Examples: "user.created", "order.placed", "payment.completed".
const (
    EventUserCreated = "user.created"
    EventUserUpdated = "user.updated"
    EventUserDeleted = "user.deleted"
)

// To add audit events for a new module (e.g., orders):
//
// 1. Define event type constants:
//    const (
//        EventOrderCreated  = "order.created"
//        EventOrderUpdated  = "order.updated"
//        EventOrderCanceled = "order.canceled"
//    )
//
// 2. In your use case, call auditService.Record():
//    auditInput := audit.AuditEventInput{
//        EventType:  domain.EventOrderCreated,
//        ActorID:    req.ActorID,
//        EntityType: "order",
//        EntityID:   order.ID,
//        Payload:    order,
//        RequestID:  req.RequestID,
//    }
//    if err := uc.auditService.Record(ctx, q, auditInput); err != nil {
//        return err
//    }
```

### Complete Integration Pattern (Reference)

From `CreateUserUseCase` in `internal/app/user/create_user.go`:

```go
func (uc *CreateUserUseCase) Execute(ctx context.Context, req CreateUserRequest) (CreateUserResponse, error) {
    // ... user creation logic ...
    
    // Record audit event within the same transaction
    auditInput := audit.AuditEventInput{
        EventType:  domain.EventUserCreated,
        ActorID:    req.ActorID,    // From transport layer
        EntityType: "user",
        EntityID:   user.ID,
        Payload:    user,           // Will be PII-redacted automatically
        RequestID:  req.RequestID,  // From transport layer
    }
    
    if err := uc.auditService.Record(ctx, q, auditInput); err != nil {
        return CreateUserResponse{}, err
    }
    
    return CreateUserResponse{User: *user}, nil
}
```

### Developer Checklist for Adding Audit Events

When adding audit events to a new module, follow this checklist:

1. **Define event type constant** in `internal/domain/audit.go`:
   ```go
   const EventOrderCreated = "order.created"
   ```

2. **Add AuditService dependency** to your use case:
   ```go
   type CreateOrderUseCase struct {
       orderRepo    domain.OrderRepository
       auditService *audit.AuditService
       txManager    domain.TxManager
       // ...
   }
   ```

3. **Update use case request struct** with RequestID and ActorID:
   ```go
   type CreateOrderRequest struct {
       // ... order fields ...
       RequestID string    // Populated by transport layer
       ActorID   domain.ID // Populated by transport layer
   }
   ```

4. **Record audit event** within the transaction:
   ```go
   auditInput := audit.AuditEventInput{
       EventType:  domain.EventOrderCreated,
       ActorID:    req.ActorID,
       EntityType: "order",
       EntityID:   order.ID,
       Payload:    order,
       RequestID:  req.RequestID,
   }
   if err := uc.auditService.Record(ctx, q, auditInput); err != nil {
       return err
   }
   ```

5. **Update handler** to extract context values:
   ```go
   req := order.CreateOrderRequest{
       // ... order fields ...
       RequestID: middleware.GetRequestID(r.Context()),
       ActorID:   getActorID(r.Context()),
   }
   ```

6. **Add unit tests** verifying audit recording

### Verification Commands

```bash
# Verify no lint errors
make lint

# Verify all tests still pass
make test

# Build to verify code compiles
go build ./...
```

### References

- [Source: docs/epics.md#Story 6.5] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Implementation Patterns] - Layer rules
- [Source: docs/project-context.md] - Layer constraints and conventions
- [Source: internal/domain/audit.go] - Existing event type constants
- [Source: internal/app/audit/service.go] - AuditService implementation
- [Source: internal/app/user/create_user.go] - Canonical integration example

### Learnings from Previous Stories

**From Story 6.4 (Audit Event Service):**
1. Transport layer extracts requestID/actorID from context, passes via input struct
2. App layer does NOT import transport - proper hexagonal architecture
3. AuditService.Record() handles PII redaction automatically
4. Use TxManager.WithTx() to ensure audit recording in same transaction

**From Stories 6.1-6.3:**
1. Event type follows "entity.action" pattern
2. EntityType is lowercase string matching entity name
3. Payload is automatically PII-redacted before storage
4. AuditEvent.Validate() is called by service before persisting

### Epic 6 Context

Epic 6 implements the Audit Trail System for compliance requirements:
- **6.1 (DONE):** Domain model (entity + repository interface)
- **6.2 (DONE):** PostgreSQL repository implementation
- **6.3 (DONE):** PII redaction service
- **6.4 (DONE):** Audit event service (app layer)
- **6.5 (this story):** Extensible event types (documentation)

This is the final story in Epic 6, completing the audit trail system with comprehensive developer documentation.

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-22)

- `docs/epics.md` - Story 6.5 acceptance criteria
- `docs/architecture.md` - Layer rules and patterns
- `docs/project-context.md` - Layer constraints and conventions
- `docs/sprint-artifacts/6-4-implement-audit-event-service.md` - Previous story
- `internal/domain/audit.go` - Event type constants, AuditEvent entity
- `internal/app/audit/service.go` - AuditService implementation
- `internal/app/user/create_user.go` - Integration pattern reference

### Agent Model Used

Antigravity AI (2025-12-22)

### Debug Log References

N/A

### Completion Notes List

- Enhanced `internal/domain/audit.go` with comprehensive GoDoc: naming conventions, step-by-step guide, example templates for orders/payments/sessions/permissions modules
- Enhanced `internal/app/audit/service.go` with package-level documentation: 6-step integration guide, developer checklist, canonical reference
- Created `docs/guides/adding-audit-events.md`: comprehensive developer guide with Quick Start, code examples, standard actions table, testing guidance, checklist
- Verified: `go build ./...` succeeded, `make lint` 0 issues, `make test` all tests pass with 90.3% coverage
- FR39 fully satisfied: developers can now easily add new audit event types for new modules

### Change Log

- 2025-12-22: Implemented Story 6.5 - Enhanced documentation for extensible audit event types

### File List

**New Files:**
- `docs/guides/adding-audit-events.md` - Developer guide for extending audit system

**Modified Files:**
- `internal/domain/audit.go` - Enhanced documentation
- `internal/app/audit/service.go` - Enhanced documentation
- `docs/sprint-artifacts/sprint-status.yaml` - Updated story status
- `docs/sprint-artifacts/6-5-enable-extensible-audit-event-types.md` - Updated tasks and status
