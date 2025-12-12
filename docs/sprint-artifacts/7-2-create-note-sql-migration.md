# Story 7.2: Create Note SQL Migration

Status: done

## Story

As a developer,
I want an example migration for Note table,
So that I can understand migration patterns.

## Acceptance Criteria

### AC1: Migration files exist
**Given** `db/migrations/YYYYMMDD_create_notes.up.sql` exists
**When** I run `make migrate-up`
**Then** notes table is created with proper columns
**And** down migration drops the table

---

## Tasks / Subtasks

- [x] **Task 1: Create up migration** (AC: #1)
  - [x] Create notes table with id, title, content, created_at, updated_at
  - [x] Add proper column types (UUID, TEXT, TIMESTAMPTZ)

- [x] **Task 2: Create down migration** (AC: #1)
  - [x] Drop notes table

- [x] **Task 3: Verify migration** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Up Migration

```sql
-- db/migrations/YYYYMMDDHHMMSS_create_notes.up.sql
CREATE TABLE IF NOT EXISTS notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    content TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notes_created_at ON notes(created_at DESC);
```

### Down Migration

```sql
-- db/migrations/YYYYMMDDHHMMSS_create_notes.down.sql
DROP TABLE IF EXISTS notes;
```

### File List

Files to create:
- `db/migrations/YYYYMMDDHHMMSS_create_notes.up.sql`
- `db/migrations/YYYYMMDDHHMMSS_create_notes.down.sql`
