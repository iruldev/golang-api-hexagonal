# Story 7.3: Create Note Repository with sqlc

Status: done

## Story

As a developer,
I want an example sqlc repository,
So that I can understand type-safe query patterns.

## Acceptance Criteria

### AC1: sqlc queries defined and generated
**Given** `db/queries/note.sql` exists
**When** I run `make gen`
**Then** `internal/infra/postgres/note/queries.go` is generated
**And** queries include: CreateNote, GetNote, ListNotes, UpdateNote, DeleteNote

---

## Tasks / Subtasks

- [x] **Task 1: Create sqlc queries** (AC: #1)
  - [x] CreateNote query
  - [x] GetNote query (by ID)
  - [x] ListNotes query (with pagination)
  - [x] UpdateNote query
  - [x] DeleteNote query

- [x] **Task 2: Run sqlc generate** (AC: #1)
  - [x] Run `sqlc generate`
  - [x] Verify queries.go is generated

- [x] **Task 3: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### sqlc Queries

```sql
-- db/queries/note.sql

-- name: CreateNote :one
INSERT INTO notes (id, title, content, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetNote :one
SELECT * FROM notes WHERE id = $1;

-- name: ListNotes :many
SELECT * FROM notes ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateNote :one
UPDATE notes SET title = $2, content = $3, updated_at = $4 WHERE id = $1
RETURNING *;

-- name: DeleteNote :exec
DELETE FROM notes WHERE id = $1;
```

### File List

Files to create:
- `db/queries/note.sql` - sqlc queries

Files generated:
- `internal/infra/postgres/note/queries.go` (by sqlc)
