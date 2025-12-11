# Story 4.3: Configure sqlc for Type-Safe Queries

Status: done

## Story

As a developer,
I want to write type-safe SQL queries using sqlc,
So that SQL errors are caught at compile time.

## Acceptance Criteria

### AC1: sqlc generates Go code from configuration
**Given** `sqlc.yaml` configuration exists
**When** I run `make gen`
**Then** Go code is generated from SQL queries
**And** generated code has typed parameters and returns

### AC2: Queries are converted to Go functions
**Given** `db/queries/*.sql` contains a query
**When** I run `make gen`
**Then** corresponding Go function is generated

---

## Tasks / Subtasks

- [x] **Task 1: Install sqlc** (AC: #1, #2)
  - [x] Install sqlc via `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
  - [x] Verify with `sqlc version`

- [x] **Task 2: Create sqlc.yaml configuration** (AC: #1)
  - [x] Create `sqlc.yaml` at project root
  - [x] Configure for PostgreSQL with pgx
  - [x] Set output to `internal/infra/postgres/sqlc`

- [x] **Task 3: Create directory structure** (AC: #2)
  - [x] Create `db/queries/` for SQL query files
  - [x] Create `db/schema/` for schema files

- [x] **Task 4: Create example schema and query** (AC: #1, #2)
  - [x] Create example schema file
  - [x] Create example query file
  - [x] Verify generated Go code has typed params

- [x] **Task 5: Add Makefile target** (AC: #1)
  - [x] Add `make gen` target for sqlc generate
  - [x] Add `make sqlc-check` for validation

- [x] **Task 6: Generate and verify code** (AC: #1, #2)
  - [x] Run `make gen`
  - [x] Verify generated Go files
  - [x] Run `make test` - all pass

---

## Dev Notes

### sqlc.yaml Configuration

```yaml
# sqlc.yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries"
    schema: "db/schema"
    gen:
      go:
        package: "sqlc"
        out: "internal/infra/postgres/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: true
```

### Example Schema (db/schema/001_users.sql)

```sql
-- db/schema/001_users.sql
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Example Query (db/queries/users.sql)

```sql
-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (email, name)
VALUES ($1, $2)
RETURNING *;
```

### Makefile Target

```makefile
# Makefile
.PHONY: gen sqlc-check

gen: ## Generate sqlc code
	sqlc generate

sqlc-check: ## Validate sqlc configuration
	sqlc compile
```

### Directory Structure

```
project/
├── sqlc.yaml
├── db/
│   ├── schema/
│   │   └── 001_users.sql
│   └── queries/
│       └── users.sql
└── internal/
    └── infra/
        └── postgres/
            └── sqlc/  # Generated code
                ├── db.go
                ├── models.go
                └── users.sql.go
```

### Architecture Compliance

**Layer:** `internal/infra/postgres/sqlc`
**Pattern:** Generated data access layer
**Benefit:** Type-safe SQL, compile-time error checking

### References

- [Source: docs/epics.md#Story-4.3]
- [sqlc documentation](https://docs.sqlc.dev)
- [Story 4.1 - PostgreSQL Connection](file:///docs/sprint-artifacts/4-1-setup-postgresql-connection-with-pgx.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Third story in Epic 4: Database & Persistence.
Sets up sqlc for type-safe query generation.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `sqlc.yaml` - sqlc configuration
- `db/schema/001_users.sql` - Example schema
- `db/queries/users.sql` - Example queries
- `internal/infra/postgres/sqlc/` - Generated code (by sqlc)

Files to modify:
- `Makefile` - Add gen target
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
