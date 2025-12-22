# Story 7.5: Create Migration Helper Commands

Status: Ready for Review

## Story

As a **developer**,
I want **make commands for managing migrations**,
so that **I can create and manage migrations easily**.

## Acceptance Criteria

1. **Given** I want to create a new migration, **When** I run `make migrate-create name=add_orders_table`, **Then** new migration file is created: `migrations/YYYYMMDDHHMMSS_add_orders_table.sql`, **And** file contains template with `-- +goose Up` and `-- +goose Down` sections.

2. **Given** I want to check migration status, **When** I run `make migrate-status`, **Then** I see list of applied and pending migrations.

*Covers: FR52-54*

## Implementation Status: ✅ Already Implemented

**Important:** These commands are already implemented in the Makefile from earlier Epic 1 work. This story documents the existing implementation and confirms it meets acceptance criteria.

### Existing Implementation in Makefile

```makefile
## migrate-status: Show migration status
.PHONY: migrate-status
migrate-status: _check-goose _check-db-url
	goose -dir migrations postgres "$(DATABASE_URL)" status

## migrate-create: Create a new migration (usage: make migrate-create name=description)
.PHONY: migrate-create
migrate-create: _check-goose
	@if [ -z "$(name)" ]; then echo "Usage: make migrate-create name=description"; exit 1; fi
	goose -dir migrations create "$(name)" sql
```

## Tasks / Subtasks

- [x] Task 1: Verify `migrate-create` command exists (AC: #1)
  - [x] 1.1 Command exists at Makefile lines 319-323
  - [x] 1.2 Validates `name` parameter is provided
  - [x] 1.3 Uses goose to create migration file in correct format
  - [x] 1.4 goose creates file with correct timestamp naming `YYYYMMDDHHMMSS_name.sql`
  - [x] 1.5 goose creates file with `-- +goose Up` and `-- +goose Down` sections

- [x] Task 2: Verify `migrate-status` command exists (AC: #2)
  - [x] 2.1 Command exists at Makefile lines 314-317
  - [x] 2.2 Uses goose status to show applied/pending status
  - [x] 2.3 Requires DATABASE_URL environment variable

- [x] Task 3: Verify help documentation
  - [x] 3.1 Commands are documented in Makefile with `##` comments
  - [x] 3.2 Commands appear in `make help` output

## Dependencies & Blockers

- **Depends on:** Story 1.4 (Implement Database Migration System) - Already completed
- **Uses:** goose v3 for migration management (installed via `make setup`)
- **Uses:** Existing helper functions `_check-goose` and `_check-db-url`

## Assumptions & Open Questions

- Commands were implemented during Epic 1 Story 1.4
- Parameter uses lowercase `name=` (not `NAME=` as in epics.md)
- Both commands follow established Makefile patterns
- goose automatically creates files with correct timestamp format

## Definition of Done

- [x] `make migrate-create name=xxx` creates new migration file with correct naming
- [x] Created file contains `-- +goose Up` and `-- +goose Down` sections
- [x] `make migrate-status` shows applied/pending migrations
- [x] Commands are documented in help output
- [x] Commands validate prerequisites (goose installed, DATABASE_URL set)

## Non-Functional Requirements

- Commands should execute quickly (<1 second)
- Clear error messages when prerequisites missing
- Consistent output format with other Makefile targets

## Testing & Verification

### Manual Verification Steps

1. **Test migrate-create:**
   ```bash
   # From project root
   make migrate-create name=test_story_7_5
   # Should create: migrations/YYYYMMDDHHMMSS_test_story_7_5.sql
   # Verify file contains:
   #   -- +goose Up
   #   -- +goose Down
   ```

2. **Test migrate-status:**
   ```bash
   # Ensure infrastructure is running
   make infra-up
   export DATABASE_URL="postgres://postgres:postgres@localhost:5432/golang_api_hexagonal?sslmode=disable"
   make migrate-status
   # Should show list of migrations with applied/pending status
   ```

3. **Test error handling:**
   ```bash
   # Without name parameter
   make migrate-create
   # Should show: Usage: make migrate-create name=description

   # Without DATABASE_URL
   unset DATABASE_URL
   make migrate-status
   # Should show error about missing DATABASE_URL
   ```

## Dev Notes

### Implementation Details

The migration helper commands leverage goose's built-in functionality:

- **migrate-create:** Uses `goose create` which automatically:
  - Generates timestamp prefix (`YYYYMMDDHHMMSS`)
  - Creates file in `migrations/` directory
  - Adds `-- +goose Up` and `-- +goose Down` sections
  - Uses `.sql` extension for SQL-based migrations

- **migrate-status:** Uses `goose status` which shows:
  - Version number
  - Applied/Pending status
  - Migration file name
  - Timestamp when applied (if applicable)

### Makefile Helper Functions

Both commands use shared helper functions:

```makefile
_check-goose:
	@which goose > /dev/null || (echo "❌ goose not found. Run 'make setup' first." && exit 1)

_check-db-url:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL is not set."; \
		# ... helpful instructions ...
		exit 1; \
	fi
```

### Parameter Case Sensitivity

The epics.md shows `NAME=add_orders_table` but implementation uses lowercase `name=description`. This is consistent with Makefile conventions and the command works correctly.

### Related Makefile Targets

| Target | Requires DATABASE_URL | Description |
|--------|----------------------|-------------|
| `migrate-up` | Yes | Apply all pending migrations |
| `migrate-down` | Yes | Rollback last migration |
| `migrate-status` | Yes | Show migration status |
| `migrate-create` | No | Create new migration file |
| `migrate-validate` | No | Validate migration syntax |

### References

- [Source: Makefile#migrate-create] Lines 319-323
- [Source: Makefile#migrate-status] Lines 314-317
- [Source: Makefile#_check-goose] Lines 285-286
- [Source: Makefile#_check-db-url] Lines 288-298
- [Source: docs/epics.md#Story 7.5] - Acceptance criteria
- [Source: docs/sprint-artifacts/7-4-implement-ci-migration-verification.md] - Story format reference

### Epic 7 Context

Epic 7 implements the CI/CD Pipeline for automated quality verification:
- **7.1 (done):** GitHub Actions workflow setup
- **7.2 (done):** CI lint and test steps with coverage enforcement
- **7.3 (done):** CI build and security scan (govulncheck)
- **7.4 (done):** CI migration verification with PostgreSQL service
- **7.5 (this story):** Migration helper commands - **Already implemented**

This story documents existing functionality that was implemented during Epic 1.

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-22)

- `docs/epics.md` - Story 7.5 acceptance criteria
- `Makefile` - Migration commands (lines 314-323)
- `docs/sprint-artifacts/7-4-implement-ci-migration-verification.md` - Story format reference
- `docs/project-context.md` - Project conventions

### Agent Model Used

Claude (Anthropic)

### Debug Log References

N/A

### Completion Notes List

- Story 7.5 requirements discovered to be already implemented in Makefile
- `migrate-create` at lines 319-323 with goose integration
- `migrate-status` at lines 314-317 with goose status
- Both commands follow established Makefile patterns
- Story file documents existing implementation for record

### Change Log

- 2025-12-22: Story file created (ready-for-dev)
- 2025-12-22: Verification completed - all acceptance criteria confirmed (Ready for Review)

### File List

**Files analyzed (no changes needed):**
- `Makefile` - Contains existing `migrate-create` and `migrate-status` implementations
