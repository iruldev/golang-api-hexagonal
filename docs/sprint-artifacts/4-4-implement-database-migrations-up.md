# Story 4.4: Implement Database Migrations (Up)

Status: done

## Story

As a developer,
I want to run database migrations via `make migrate-up`,
So that schema changes are applied consistently.

## Acceptance Criteria

### AC1: Migrations applied in order
**Given** migrations exist in `db/migrations/`
**When** I run `make migrate-up`
**Then** pending migrations are applied in order
**And** migration version is tracked

### AC2: No re-apply of existing migrations
**Given** migrations are already applied
**When** I run `make migrate-up`
**Then** no migrations are re-applied
**And** output shows "no change"

---

## Tasks / Subtasks

- [x] **Task 1: Install golang-migrate** (AC: #1, #2)
  - [x] Install migrate CLI via `go install`
  - [x] Verify with `migrate -version`

- [x] **Task 2: Create migrations directory** (AC: #1)
  - [x] Create `db/migrations/` directory
  - [x] Create first migration (users table)

- [x] **Task 3: Add Makefile migrate-up target** (AC: #1, #2)
  - [x] Add `make migrate-up` target
  - [x] Use DATABASE_URL from .env
  - [x] Handle "no change" case

- [x] **Task 4: Create migration helper** (AC: #1)
  - [x] Create `make migrate-create` target for new migrations
  - [x] Use timestamp naming convention

- [x] **Task 5: Test migration flow** (AC: #1, #2)
  - [x] Run `make migrate-up` with fresh DB
  - [x] Run again to verify "no change"
  - [x] Verify schema in database

- [x] **Task 6: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### golang-migrate Installation

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Migration File Format

```
db/migrations/
├── 000001_create_users_table.up.sql
└── 000001_create_users_table.down.sql
```

### Example Up Migration (000001_create_users_table.up.sql)

```sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

### Makefile Targets

```makefile
# Database URL (from .env)
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

migrate-up: ## Run database migrations
	migrate -database "$(DATABASE_URL)" -path db/migrations up

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	migrate create -ext sql -dir db/migrations -seq $(NAME)
```

### Architecture Compliance

**Layer:** `db/migrations/`
**Pattern:** SQL migrations with version tracking
**Benefit:** Reproducible schema changes across environments

### References

- [Source: docs/epics.md#Story-4.4]
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Story 4.3 - sqlc Configuration](file:///docs/sprint-artifacts/4-3-configure-sqlc-for-type-safe-queries.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Fourth story in Epic 4: Database & Persistence.
Sets up database migration infrastructure.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `db/migrations/000001_create_users_table.up.sql` - Users table migration
- `db/migrations/000001_create_users_table.down.sql` - Rollback migration

Files to modify:
- `Makefile` - Add migrate-up and migrate-create targets
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
