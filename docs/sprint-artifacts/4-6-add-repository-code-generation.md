# Story 4.6: Add Repository Code Generation

Status: done

## Story

As a developer,
I want to generate repository code via `make gen`,
So that boilerplate code is automated.

## Acceptance Criteria

### AC1: Query files generate to domain-specific paths
**Given** `db/queries/note.sql` contains queries
**When** I run `make gen`
**Then** `internal/infra/postgres/note/queries.go` is generated
**And** generated code compiles without errors

---

## Tasks / Subtasks

- [x] **Task 1: Update sqlc.yaml configuration** (AC: #1)
  - [x] Configure multiple query sources
  - [x] Map each query file to its output package

- [x] **Task 2: Create note domain queries** (AC: #1)
  - [x] Create `db/queries/note.sql` with CRUD queries
  - [x] Define note table schema in migrations

- [x] **Task 3: Regenerate sqlc code** (AC: #1)
  - [x] Run `make gen`
  - [x] Verify `internal/infra/postgres/note/` is created

- [x] **Task 4: Verify compilation** (AC: #1)
  - [x] Run `go build ./...` - no errors
  - [x] Run `make test` - all pass

- [x] **Task 5: Verify implementation** (AC: #1)
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### sqlc.yaml Multi-Output Configuration

```yaml
version: "2"
sql:
  # Users queries (existing)
  - engine: "postgresql"
    queries: "db/queries/users.sql"
    schema: "db/migrations"
    gen:
      go:
        package: "users"
        out: "internal/infra/postgres/users"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true

  # Note queries (new for this story)
  - engine: "postgresql"
    queries: "db/queries/note.sql"
    schema: "db/migrations"
    gen:
      go:
        package: "note"
        out: "internal/infra/postgres/note"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
```

### Note Migration (000002_create_notes_table.up.sql)

```sql
CREATE TABLE IF NOT EXISTS notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notes_user_id ON notes(user_id);
```

### Note Queries (db/queries/note.sql)

```sql
-- name: GetNote :one
SELECT * FROM notes WHERE id = $1;

-- name: ListNotesByUser :many
SELECT * FROM notes WHERE user_id = $1 ORDER BY created_at DESC;

-- name: CreateNote :one
INSERT INTO notes (user_id, title, content)
VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateNote :one
UPDATE notes SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1 RETURNING *;

-- name: DeleteNote :exec
DELETE FROM notes WHERE id = $1;
```

### Generated Structure

```
internal/infra/postgres/
├── sqlc/         # Current single output (will be removed)
├── users/        # Generated from users.sql
│   ├── db.go
│   ├── models.go
│   ├── querier.go
│   └── users.sql.go
└── note/         # Generated from note.sql (Story 4.6)
    ├── db.go
    ├── models.go
    ├── querier.go
    └── note.sql.go
```

### Architecture Compliance

**Layer:** `internal/infra/postgres/{domain}/`
**Pattern:** Domain-separated generated data access
**Benefit:** Clean separation, each domain gets its own package

### References

- [Source: docs/epics.md#Story-4.6]
- [Story 4.3 - sqlc Configuration](file:///docs/sprint-artifacts/4-3-configure-sqlc-for-type-safe-queries.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Sixth story in Epic 4: Database & Persistence.
Splits sqlc output per domain for clean architecture.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `db/queries/note.sql` - Note CRUD queries
- `db/migrations/000002_create_notes_table.up.sql` - Note schema
- `db/migrations/000002_create_notes_table.down.sql` - Rollback
- `internal/infra/postgres/note/` - Generated (by sqlc)
- `internal/infra/postgres/users/` - Restructured output

Files to modify:
- `sqlc.yaml` - Multi-output configuration
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

Files to remove:
- `internal/infra/postgres/sqlc/` - Replaced by domain-specific folders
