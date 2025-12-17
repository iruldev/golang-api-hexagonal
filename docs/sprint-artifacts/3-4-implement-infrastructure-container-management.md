# Story 3.4: Infrastructure Container Management

Status: done

## Story

As a **developer**,
I want **to start/stop infrastructure containers independently**,
So that **I can manage local development resources**.

## Acceptance Criteria

1. **Given** I want to start infrastructure
   **When** I run `make infra-up`
   **Then** only infrastructure containers start (not the app)
   **And** containers run in detached mode
   **And** command waits for healthcheck to pass

2. **Given** infrastructure is running
   **When** I run `make infra-down`
   **Then** containers stop and are removed
   **And** data volume is preserved (not deleted)

3. **Given** I want fresh infrastructure
   **When** I run `make infra-reset`
   **Then** warning message is printed: "WARNING: removing volumes"
   **And** containers and volumes are removed
   **And** infrastructure can be started fresh with `make infra-up`

*Covers: FR51*

## Tasks / Subtasks

- [x] Task 1: Ensure `infra-up` meets AC #1
  - [x] Ensure it starts **only** PostgreSQL (use service-scoped compose command, not full `up` for all services)
  - [x] Ensure it runs in detached mode (`-d`)
  - [x] Ensure it waits for PostgreSQL healthcheck with timeout (default 60s via `INFRA_TIMEOUT`)
  - [x] Verify behavior via `make infra-status` and logs if needed

- [x] Task 2: Ensure `infra-down` meets AC #2
  - [x] Ensure PostgreSQL container stops and is removed (without removing volumes)
  - [x] Verify data volume `golang-api-hexagonal-pgdata` is preserved
  - [x] Test: restart with `make infra-up` and confirm data persists

- [x] Task 3: Ensure `infra-reset` meets AC #3
  - [x] Ensure warning message includes exactly: `WARNING: removing volumes`
  - [x] Ensure confirmation prompt before destructive action
  - [x] Ensure PostgreSQL container is removed and `golang-api-hexagonal-pgdata` volume is removed
  - [x] Test: after reset, `make infra-up` starts fresh infrastructure

- [x] Task 4: Add helper targets for developers (optional enhancement)
  - [x] Add `infra-logs` target if not present
  - [x] Add `infra-status` target if not present
  - [x] Update `make help` documentation if needed

- [x] Task 5: Manual verification and documentation
  - [x] Run full test scenario: up → verify → down → up → verify data persists → reset → up → verify fresh
  - [x] Update story status to `done`

## Dev Notes

### Current Implementation Analysis

The existing `Makefile` already has infrastructure targets implemented from Story 1.3:

**`make infra-up`:**
- ✅ Starts PostgreSQL in detached mode (service-scoped)
- ✅ Waits for healthcheck with configurable timeout (`INFRA_TIMEOUT`, default 60s)
- ✅ Displays connection details on success
- ✅ Safe if an app service is later added to `docker-compose.yaml` (it won’t be started by infra targets)

**`make infra-down`:**
- ✅ Stops and removes only the PostgreSQL container
- ✅ Preserves named volume `golang-api-hexagonal-pgdata`
- ✅ Clear messaging about data preservation

**`make infra-reset`:**
- ✅ Prints required warning substring: `WARNING: removing volumes`
- ✅ Uses confirmation prompt (or set `INFRA_CONFIRM=y` for non-interactive runs)
- ✅ Removes PostgreSQL container and `golang-api-hexagonal-pgdata` volume

**Helper targets already exist:**
- `make infra-logs`
- `make infra-status`

### Docker Compose Configuration

File: `docker-compose.yaml`
- PostgreSQL 15 Alpine image
- Named container: `golang-api-hexagonal-db`
- Named volume: `golang-api-hexagonal-pgdata`
- Healthcheck: `pg_isready` with 5s interval, 5 retries
- No application container (infrastructure only) ✅

### Verification Strategy

1. **Automated Checks:**
   - `make infra-up` → verify container running with `docker compose ps`
   - `make infra-down` → verify container stopped, volume exists
   - `make infra-reset` → verify volume removed

2. **Manual Verification:**
   - Insert test data → `make infra-down` → `make infra-up` → verify data persists
   - `make infra-reset` → `make infra-up` → verify data is gone

**Config knobs:**
- `INFRA_TIMEOUT` (default `60`) to tune healthcheck wait, e.g. `make infra-up INFRA_TIMEOUT=120`
- `INFRA_CONFIRM=y` to skip confirmation prompt for `infra-reset` in non-interactive runs

### Potential Enhancements (if AC not fully met)

1. **Timeout configurability:** Current 60s hardcoded - could be parameterized

### Previous Story Learnings (From Story 3.3)

- Use consistent emoji indicators (✅ ❌ 🐘 ⏳ 🛑 ⚠️)
- Clear step-by-step output for user feedback
- Fail-fast with clear error messages

### Technology Specifics

- **Docker Compose**: v2 syntax (`docker compose` not `docker-compose`)
- **PostgreSQL**: 15-alpine image
- **Healthcheck**: `pg_isready` command
- **Volume**: Named volume for data persistence

### File Locations

| Action | Path |
|--------|------|
| Verify | `Makefile` (Infrastructure targets section) |
| Verify | `docker-compose.yaml` (PostgreSQL service) |
| Reference | `docs/sprint-artifacts/1-3-setup-docker-compose-infrastructure.md` |

### References

- [Source: docs/epics.md#Story-3.4]
- [Source: docs/project-context.md#Makefile-Commands]
- [Source: docs/sprint-artifacts/1-3-setup-docker-compose-infrastructure.md]
- [Source: Makefile#Infrastructure]

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-17)

- `docs/project-context.md` - Makefile patterns and conventions
- `docs/epics.md` - Story 3.4 acceptance criteria
- `Makefile` - Current infrastructure targets
- `docker-compose.yaml` - PostgreSQL service definition
- `docs/sprint-artifacts/3-3-create-local-ci-pipeline.md` - Previous story learnings

### Agent Model Used

gpt-5 (Codex CLI)

### Debug Log References

No debug issues encountered.

### Completion Notes List

- **2025-12-17**: Infrastructure targets updated to satisfy AC and reduce drift risk
  - ✅ AC #1: `make infra-up` starts only PostgreSQL, uses detached mode, waits for healthcheck (default 60s timeout)
  - ✅ AC #2: `make infra-down` stops container and preserves `golang-api-hexagonal-pgdata` volume
  - ✅ AC #3: `make infra-reset` prints `WARNING: removing volumes`, prompts confirmation, and removes `golang-api-hexagonal-pgdata`
  - ✅ Helper targets `infra-logs` and `infra-status` present in Makefile and `make help`
  - ✅ Manual end-to-end scenario executed:
    - `make infra-up` → `make infra-status` (healthy)
    - Insert sentinel row into DB, then `make infra-down` → `make infra-up` → sentinel still present (data persists)
    - `make infra-reset` → `make infra-up` → sentinel table no longer exists (fresh infra)

### File List

| Action | Path |
|--------|------|
| Modified | `Makefile` (Infrastructure targets: `infra-up`, `infra-down`, `infra-reset`, `infra-logs`, `infra-status`) |
| Verified | `docker-compose.yaml` (PostgreSQL service) |
| Modified | `docs/sprint-artifacts/sprint-status.yaml` (story status synced to `done`) |
| Modified | `docs/sprint-artifacts/3-4-implement-infrastructure-container-management.md` (story updates + audit trail) |

### Change Log

- **2025-12-17**: Updated infrastructure targets for correctness + auditability
  - Scoped `infra-up/infra-down/infra-reset` to the `postgres` service (prevents accidental app startup)
  - Fixed health detection to avoid `unhealthy` false-positive matches
  - Made stop/reset behavior idempotent but fail-fast (no silent success when Docker/compose errors)
  - Added `INFRA_TIMEOUT` and `INFRA_CONFIRM` knobs for developer ergonomics and non-interactive runs
  - Synced `docs/sprint-artifacts/sprint-status.yaml` story state to `done`

---
