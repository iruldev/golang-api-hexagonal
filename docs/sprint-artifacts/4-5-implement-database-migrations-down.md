# Story 4.5: Implement Database Migrations (Down)

Status: done

## Story

As a developer,
I want to rollback migrations via `make migrate-down`,
So that I can undo schema changes.

## Acceptance Criteria

### AC1: Rollback last migration
**Given** migrations have been applied
**When** I run `make migrate-down`
**Then** last migration is rolled back
**And** migration version is updated

### AC2: Rollback N migrations
**Given** I run `make migrate-down N=2`
**When** command completes
**Then** 2 migrations are rolled back

---

## Tasks / Subtasks

- [x] **Task 1: Update migrate-down target** (AC: #1)
  - [x] Verify `make migrate-down` rolls back 1 migration
  - [x] Migration version updated after rollback

- [x] **Task 2: Add N parameter support** (AC: #2)
  - [x] Update `make migrate-down` to support `N=X` parameter
  - [x] Default to 1 if N not specified

- [x] **Task 3: Add migrate-down-all target** (AC: #1, #2)
  - [x] Create `make migrate-down-all` to drop all migrations
  - [x] Add WARNING message for safety

- [x] **Task 4: Update Makefile help** (AC: #1, #2)
  - [x] Add migrate-down usage with N parameter
  - [x] Document migrate-down-all

- [x] **Task 5: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Makefile Targets

```makefile
# Rollback N migrations (default: 1)
migrate-down: ## Rollback migrations (usage: make migrate-down N=2)
	migrate -database "$(DATABASE_URL)" -path db/migrations down $(or $(N),1)

# Rollback all migrations (dangerous!)
migrate-down-all: ## Rollback ALL migrations
	@echo "Warning: This will drop ALL migrations!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	migrate -database "$(DATABASE_URL)" -path db/migrations down -all
```

### Usage Examples

```bash
# Rollback last migration
make migrate-down

# Rollback 2 migrations
make migrate-down N=2

# Rollback all migrations (with confirmation)
make migrate-down-all
```

### Implementation Note

The existing `make migrate-down` already rolls back 1 migration.
This story adds the `N=X` parameter for multiple rollbacks.

### Architecture Compliance

**Layer:** `Makefile`
**Pattern:** CLI interface for database operations
**Benefit:** Consistent migration workflow

### References

- [Source: docs/epics.md#Story-4.5]
- [Story 4.4 - Migrations Up](file:///docs/sprint-artifacts/4-4-implement-database-migrations-up.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Fifth story in Epic 4: Database & Persistence.
Completes migration infrastructure with rollback support.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to modify:
- `Makefile` - Update migrate-down, add migrate-down-all
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
