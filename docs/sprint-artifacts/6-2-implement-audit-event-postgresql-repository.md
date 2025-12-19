Status: done

## Story

As a **developer**,
I want **AuditEventRepository implementation for PostgreSQL**,
so that **audit events are persisted to the database**.

## Acceptance Criteria

1. **Given** migration file `migrations/YYYYMMDDHHMMSS_create_audit_events.sql`, **When** migration is applied, **Then** table `audit_events` is created with columns:
   - `id` (uuid, primary key)
   - `event_type` (varchar, indexed)
   - `actor_id` (uuid, nullable for system events)
   - `entity_type` (varchar, indexed)
   - `entity_id` (uuid, indexed)
   - `payload` (jsonb)
   - `timestamp` (timestamptz, indexed)
   - `request_id` (varchar)

2. **Given** Create is called with valid audit event, **When** the operation succeeds, **Then** event is inserted into `audit_events` table, **And** `domain.ID` is parsed to `uuid.UUID` at repository boundary, **And** payload []byte is stored as JSONB.

3. **Given** ListByEntityID is called, **When** query executes, **Then** results are filtered by entity_type and entity_id, **And** ordered by `timestamp DESC`, **And** total count is returned for pagination.

*Covers: FR36*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 6.2".
- Repository patterns established in `internal/infra/postgres/user_repo.go`.
- Domain entity defined in `internal/domain/audit.go` (Story 6.1).
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Create migration file (AC: #1)
  - [x] 1.1 Create `migrations/YYYYMMDDHHMMSS_create_audit_events.sql` (use current timestamp)
  - [x] 1.2 Define `audit_events` table with all required columns
  - [x] 1.3 Add `-- +goose Up` and `-- +goose Down` sections
  - [x] 1.4 Create index `idx_audit_events_event_type` on `event_type`
  - [x] 1.5 Create index `idx_audit_events_entity_type` on `entity_type`
  - [x] 1.6 Create composite index `idx_audit_events_entity` on `(entity_type, entity_id)`
  - [x] 1.7 Create index `idx_audit_events_timestamp` on `timestamp`
  - [x] 1.8 Run `make migrate-up` to verify migration applies cleanly

- [x] Task 2: Implement AuditEventRepo.Create (AC: #2)
  - [x] 2.1 Create `internal/infra/postgres/audit_event_repo.go`
  - [x] 2.2 Implement `Create(ctx, q, event) error` method
  - [x] 2.3 Parse `domain.ID` to `uuid.UUID` at repository boundary for `ID`, `ActorID`, `EntityID`
  - [x] 2.4 Handle nullable `ActorID` (convert empty ID to NULL)
  - [x] 2.5 Store `Payload` []byte as JSONB
  - [x] 2.6 Use `const op = "auditEventRepo.Create"` for error wrapping
  - [x] 2.7 Add constructor `NewAuditEventRepo() *AuditEventRepo`
  - [x] 2.8 Add compile-time interface check: `var _ domain.AuditEventRepository = (*AuditEventRepo)(nil)`

- [x] Task 3: Implement AuditEventRepo.ListByEntityID (AC: #3)
  - [x] 3.1 Implement `ListByEntityID(ctx, q, entityType, entityID, params) ([]AuditEvent, int, error)`
  - [x] 3.2 Filter by `entity_type` AND `entity_id`
  - [x] 3.3 Order by `timestamp DESC, id DESC` (deterministic ordering)
  - [x] 3.4 Get total count for pagination
  - [x] 3.5 Apply `LIMIT` and `OFFSET` from `params.Limit()` and `params.Offset()`
  - [x] 3.6 Convert `uuid.UUID` back to `domain.ID` for returned events
  - [x] 3.7 Handle nullable `ActorID` scan (use `*uuid.UUID` for scanning)

- [x] Task 4: Write integration tests (AC: all)
  - [x] 4.1 Create `internal/infra/postgres/audit_event_repo_test.go` with `//go:build integration`
  - [x] 4.2 Add test `TestAuditEventRepo_Create` for successful creation
  - [x] 4.3 Add test `TestAuditEventRepo_Create_WithNullActorID` for system events
  - [x] 4.4 Add test `TestAuditEventRepo_ListByEntityID_WithPagination`
  - [x] 4.5 Add test `TestAuditEventRepo_ListByEntityID_OrderByTimestampDesc`
  - [x] 4.6 Add test `TestAuditEventRepo_ListByEntityID_FiltersByEntityTypeAndID`
  - [x] 4.7 Update cleanup in `setupTestDB` to include `DELETE FROM audit_events`
  - [x] 4.8 Achieve ≥80% coverage for new code (integration tests cover all paths)

- [x] Task 5: Verify layer compliance (AC: implicit)
  - [x] 5.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 5.2 Run `make test` to ensure all unit tests pass
  - [x] 5.3 Run integration tests: `DATABASE_URL=... go test -tags=integration ./internal/infra/postgres/...`
  - [x] 5.4 Run `make ci` (or `ALLOW_DIRTY=1 make ci`) for full CI check

## Dependencies & Blockers

- **Hard dependency:** Story 6.1 (Audit Event Domain Model) - **DONE**
- Uses existing `domain.AuditEvent`, `domain.AuditEventRepository`, `domain.Querier`
- Uses existing `domain.ListParams` for pagination
- Depends on existing PostgreSQL setup and migration system

## Assumptions & Open Questions

- `ActorID` can be NULL in database (for system/unauthenticated events)
- `Payload` is stored as JSONB (allows PostgreSQL JSON queries if needed later)
- Migration should use `timestamptz` for timezone-aware timestamps
- No foreign key to `users` table for `actor_id` (actors may be deleted, audit preserved)

## Definition of Done

- Migration file created and applies successfully
- `AuditEventRepo` implements `domain.AuditEventRepository`
- Create and ListByEntityID work correctly
- Integration tests pass with ≥80% coverage
- `make lint` passes (layer compliance verified)
- `make ci` passes

## Non-Functional Requirements

- Infra layer can import: domain, pgx, uuid, slog, otel
- Infra layer CANNOT import: app, transport
- Error wrapping with `op` string pattern
- Convert `domain.ID` ↔ `uuid.UUID` at repository boundary

## Testing & Coverage

- Integration tests with real PostgreSQL (testcontainers or local)
- Test nullable ActorID handling
- Test pagination (multiple pages)
- Test ordering (timestamp DESC)
- Test filtering (by entity_type and entity_id)
- Aim for coverage ≥80% for new repository code

## Dev Notes

### ⚠️ CRITICAL: Repository Layer Rules

The infra/postgres layer must follow strict patterns:

```
✅ ALLOWED: domain, pgx, uuid, slog, otel, external packages
❌ FORBIDDEN: app, transport, net/http
```

### Existing Code Context

**From Story 6.1 (Domain Layer - DONE):**
| File | Description |
|------|-------------|
| `internal/domain/audit.go` | AuditEvent entity, AuditEventRepository interface |
| `internal/domain/errors.go` | ErrAuditEventNotFound, ErrInvalidEventType, etc. |

**Reference Implementation (User Repository):**
| File | Description |
|------|-------------|
| `internal/infra/postgres/user_repo.go` | Reference pattern for repository implementation |
| `internal/infra/postgres/user_repo_test.go` | Reference pattern for integration tests |
| `internal/infra/postgres/querier.go` | PoolQuerier, TxQuerier, rowScanner interfaces |

**This story CREATES:**
| File | Description |
|------|-------------|
| `migrations/YYYYMMDDHHMMSS_create_audit_events.sql` | Database migration |
| `internal/infra/postgres/audit_event_repo.go` | AuditEventRepository implementation |
| `internal/infra/postgres/audit_event_repo_test.go` | Integration tests |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### Migration File Pattern

Follow existing migration style from `migrations/20251217000000_create_users.sql`:

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE audit_events (
    id uuid PRIMARY KEY,
    event_type varchar(100) NOT NULL,
    actor_id uuid,  -- NULL for system/unauthenticated events
    entity_type varchar(50) NOT NULL,
    entity_id uuid NOT NULL,
    payload jsonb NOT NULL,
    timestamp timestamptz NOT NULL,
    request_id varchar(50)
);

CREATE INDEX idx_audit_events_event_type ON audit_events(event_type);
CREATE INDEX idx_audit_events_entity_type ON audit_events(entity_type);
CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_timestamp ON audit_events(timestamp DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_events;
-- +goose StatementEnd
```

### Repository Implementation Pattern

Follow the existing `UserRepo` pattern exactly:

```go
package postgres

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// AuditEventRepo implements domain.AuditEventRepository for PostgreSQL.
type AuditEventRepo struct{}

// NewAuditEventRepo creates a new AuditEventRepo instance.
func NewAuditEventRepo() *AuditEventRepo {
    return &AuditEventRepo{}
}

// Create stores a new audit event in the database.
func (r *AuditEventRepo) Create(ctx context.Context, q domain.Querier, event *domain.AuditEvent) error {
    const op = "auditEventRepo.Create"

    // Parse domain.ID to uuid.UUID at repository boundary
    id, err := uuid.Parse(string(event.ID))
    if err != nil {
        return fmt.Errorf("%s: parse ID: %w", op, err)
    }

    entityID, err := uuid.Parse(string(event.EntityID))
    if err != nil {
        return fmt.Errorf("%s: parse EntityID: %w", op, err)
    }

    // Handle nullable ActorID
    var actorID *uuid.UUID
    if !event.ActorID.IsEmpty() {
        parsed, err := uuid.Parse(string(event.ActorID))
        if err != nil {
            return fmt.Errorf("%s: parse ActorID: %w", op, err)
        }
        actorID = &parsed
    }

    _, err = q.Exec(ctx, `
        INSERT INTO audit_events (id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, id, event.EventType, actorID, event.EntityType, entityID, event.Payload, event.Timestamp, event.RequestID)

    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}

// ListByEntityID retrieves audit events for a specific entity.
// Results are ordered by timestamp DESC (newest first).
func (r *AuditEventRepo) ListByEntityID(ctx context.Context, q domain.Querier, entityType string, entityID domain.ID, params domain.ListParams) ([]domain.AuditEvent, int, error) {
    const op = "auditEventRepo.ListByEntityID"

    // Parse entityID to uuid.UUID
    eid, err := uuid.Parse(string(entityID))
    if err != nil {
        return nil, 0, fmt.Errorf("%s: parse entityID: %w", op, err)
    }

    // Get total count
    countRow := q.QueryRow(ctx, `
        SELECT COUNT(*) FROM audit_events 
        WHERE entity_type = $1 AND entity_id = $2
    `, entityType, eid)
    countScanner, ok := countRow.(rowScanner)
    if !ok {
        return nil, 0, fmt.Errorf("%s: invalid querier type for count", op)
    }

    var totalCount int
    if err := countScanner.Scan(&totalCount); err != nil {
        return nil, 0, fmt.Errorf("%s: count: %w", op, err)
    }

    // If no results, return early
    if totalCount == 0 {
        return []domain.AuditEvent{}, 0, nil
    }

    // Get paginated results
    rows, err := q.Query(ctx, `
        SELECT id, event_type, actor_id, entity_type, entity_id, payload, timestamp, request_id
        FROM audit_events
        WHERE entity_type = $1 AND entity_id = $2
        ORDER BY timestamp DESC, id DESC
        LIMIT $3 OFFSET $4
    `, entityType, eid, params.Limit(), params.Offset())
    if err != nil {
        return nil, 0, fmt.Errorf("%s: query: %w", op, err)
    }

    scanner, ok := rows.(rowsScanner)
    if !ok {
        return nil, 0, fmt.Errorf("%s: invalid querier type for rows", op)
    }
    defer scanner.Close()

    var events []domain.AuditEvent
    for scanner.Next() {
        var event domain.AuditEvent
        var dbID, dbActorID, dbEntityID uuid.UUID
        var actorIDPtr *uuid.UUID
        
        if err := scanner.Scan(&dbID, &event.EventType, &actorIDPtr, &event.EntityType, &dbEntityID, &event.Payload, &event.Timestamp, &event.RequestID); err != nil {
            return nil, 0, fmt.Errorf("%s: scan: %w", op, err)
        }
        
        event.ID = domain.ID(dbID.String())
        event.EntityID = domain.ID(dbEntityID.String())
        if actorIDPtr != nil {
            event.ActorID = domain.ID(actorIDPtr.String())
        }
        // ActorID remains empty if NULL in DB (zero value)
        
        events = append(events, event)
    }

    if err := scanner.Err(); err != nil {
        return nil, 0, fmt.Errorf("%s: rows: %w", op, err)
    }

    return events, totalCount, nil
}

// Ensure AuditEventRepo implements domain.AuditEventRepository at compile time.
var _ domain.AuditEventRepository = (*AuditEventRepo)(nil)
```

### Integration Test Pattern

Follow the existing `user_repo_test.go` pattern:

```go
//go:build integration

package postgres_test

import (
    "context"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/iruldev/golang-api-hexagonal/internal/domain"
    "github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
)

func TestAuditEventRepo_Create(t *testing.T) {
    pool, cleanup := setupTestDB(t)
    defer cleanup()

    ctx := context.Background()
    repo := postgres.NewAuditEventRepo()
    querier := postgres.NewPoolQuerier(pool)

    id, err := uuid.NewV7()
    require.NoError(t, err)

    entityID, err := uuid.NewV7()
    require.NoError(t, err)

    actorID, err := uuid.NewV7()
    require.NoError(t, err)

    now := time.Now().UTC().Truncate(time.Microsecond)
    event := &domain.AuditEvent{
        ID:         domain.ID(id.String()),
        EventType:  domain.EventUserCreated,
        ActorID:    domain.ID(actorID.String()),
        EntityType: "user",
        EntityID:   domain.ID(entityID.String()),
        Payload:    []byte(`{"email":"[REDACTED]"}`),
        Timestamp:  now,
        RequestID:  "req-123",
    }

    err = repo.Create(ctx, querier, event)
    assert.NoError(t, err)

    // Verify via ListByEntityID
    events, count, err := repo.ListByEntityID(ctx, querier, "user", event.EntityID, domain.ListParams{Page: 1, PageSize: 10})
    assert.NoError(t, err)
    assert.Equal(t, 1, count)
    require.Len(t, events, 1)
    assert.Equal(t, event.EventType, events[0].EventType)
}

func TestAuditEventRepo_Create_WithNullActorID(t *testing.T) {
    pool, cleanup := setupTestDB(t)
    defer cleanup()

    ctx := context.Background()
    repo := postgres.NewAuditEventRepo()
    querier := postgres.NewPoolQuerier(pool)

    id, _ := uuid.NewV7()
    entityID, _ := uuid.NewV7()

    now := time.Now().UTC().Truncate(time.Microsecond)
    event := &domain.AuditEvent{
        ID:         domain.ID(id.String()),
        EventType:  "system.scheduled_task",
        ActorID:    "",  // Empty = system event
        EntityType: "job",
        EntityID:   domain.ID(entityID.String()),
        Payload:    []byte(`{"task":"cleanup"}`),
        Timestamp:  now,
        RequestID:  "cron-456",
    }

    err := repo.Create(ctx, querier, event)
    assert.NoError(t, err)

    // Verify ActorID is empty when retrieved
    events, _, err := repo.ListByEntityID(ctx, querier, "job", event.EntityID, domain.ListParams{Page: 1, PageSize: 10})
    assert.NoError(t, err)
    require.Len(t, events, 1)
    assert.Empty(t, events[0].ActorID)
}
```

### Cleanup Test Setup

Update the cleanup function in `setupTestDB` to include audit_events:

```go
cleanup := func() {
    // Clean up test data - order matters for foreign keys
    _, _ = pool.Exec(ctx, "DELETE FROM audit_events")
    _, _ = pool.Exec(ctx, "DELETE FROM users")
    db.Close()
    pool.Close()
}
```

### Verification Commands

```bash
# Apply migrations
make migrate-up

# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run integration tests (requires DATABASE_URL pointing to test DB)
DATABASE_URL=postgres://user:pass@localhost:5432/golang_api_hex_test?sslmode=disable \
  go test -tags=integration -v ./internal/infra/postgres/...

# Run full local CI
ALLOW_DIRTY=1 make ci

# Check coverage for infra layer
go test -tags=integration -cover ./internal/infra/postgres/...
```

### References

- [Source: docs/epics.md#Story 6.2] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Data Architecture] - Repository pattern, Querier abstraction
- [Source: docs/project-context.md#Infra Layer] - Layer constraints
- [Source: internal/infra/postgres/user_repo.go] - Repository implementation pattern
- [Source: internal/infra/postgres/user_repo_test.go] - Integration test pattern
- [Source: internal/domain/audit.go] - AuditEvent entity and repository interface

### Learnings from Previous Stories

**From Story 4.2 (User PostgreSQL Repository):**
1. Parse `domain.ID` to `uuid.UUID` at repository boundary entrance
2. Convert `uuid.UUID` back to `domain.ID` when returning entities
3. Use `const op = "repoName.methodName"` for error wrapping
4. Use interface check: `var _ domain.Interface = (*Impl)(nil)`
5. Integration tests require `//go:build integration` tag
6. Clean up tables in reverse order of foreign key dependencies

**From Story 6.1 (Audit Event Domain Model):**
1. `ActorID` is optional (empty for system/unauthenticated events)
2. `Payload` is pre-redacted JSON bytes
3. `RequestID` is a plain string (not domain.ID)
4. Event types follow "entity.action" pattern

### Security Considerations

1. **Nullable ActorID:** Handle NULL properly in DB (not empty string)
2. **JSONB Payload:** Allows PostgreSQL JSON queries for compliance audits
3. **Request ID:** Links audit events to HTTP request logs
4. **No FK to users:** Audit events preserved even if actor is deleted

### Epic 6 Context

Epic 6 implements the Audit Trail System for compliance requirements:
- **6.1 (DONE):** Domain model (entity + repository interface)
- **6.2 (this story):** PostgreSQL repository implementation
- **6.3:** PII redaction service
- **6.4:** Audit event service (app layer)
- **6.5:** Extensible event types

This story provides the persistence layer that 6.4 (service) will use.

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-19)

- `docs/epics.md` - Story 6.2 acceptance criteria
- `docs/architecture.md` - Repository patterns, Querier abstraction
- `docs/project-context.md` - Layer constraints and conventions
- `docs/sprint-artifacts/6-1-implement-audit-event-domain-model.md` - Previous story
- `internal/infra/postgres/user_repo.go` - Repository implementation pattern
- `internal/infra/postgres/user_repo_test.go` - Integration test pattern
- `internal/domain/audit.go` - AuditEvent entity and interface

### Agent Model Used

Gemini 2.5

### Debug Log References

N/A

### Completion Notes List

- Created migration `20251219000000_create_audit_events.sql` with all required columns and indexes
- Implemented `AuditEventRepo` following `UserRepo` patterns exactly
- Nullable `ActorID` handled via `*uuid.UUID` for both insert and scan
- Added 6 comprehensive integration tests covering all acceptance criteria
- Updated `setupTestDB` cleanup to delete `audit_events` before `users`
- All verification passed: `make lint` (0 issues), `make test`, `ALLOW_DIRTY=1 make ci`

### Change Log

- 2025-12-19: Story 6.2 implemented - migration, repository, integration tests

### File List

- `migrations/20251219000000_create_audit_events.sql` [NEW]
- `internal/infra/postgres/audit_event_repo.go` [NEW]
- `internal/infra/postgres/audit_event_repo_test.go` [NEW]
- `internal/infra/postgres/user_repo_test.go` [MODIFIED - cleanup function]
- `docs/sprint-artifacts/sprint-status.yaml` [MODIFIED - status updates]

## Senior Developer Review (AI)

- **Date:** 2025-12-19
- **Reviewer:** Antigravity (AI)
- **Status:** Approved with Fixes

### Findings & Actions
1. **Inefficient Indexing (Medium):** Updated `migrations/20251219000000_create_audit_events.sql` to use composite index `(entity_type, entity_id, timestamp DESC)`.
2. **Coupled Test Cleanup (Low):** Added explanatory comment to `internal/infra/postgres/user_repo_test.go` clarifying why `audit_events` deletion is necessary.
3. **Missing Negative Tests (Low):** Added `TestAuditEventRepo_ListByEntityID_InvalidID` to verify invalid UUID handling.
4. **Test Isolation (Low):** Fixed `internal/infra/config/config_test.go` to explicitly unset environment variables, preventing false negatives when running in an environment with loaded variables.

### Verification
- `make migrate-down/up` passed.
- All tests passed (including new negative test and fixed config test).
- Lint check passed.
