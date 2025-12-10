# Story 1.3: Setup Environment Configuration

Status: done

## Story

As a developer,
I want to configure the service via environment variables,
So that I can adapt settings without code changes.

## Acceptance Criteria

### AC1: .env.example contains all required variables ✅
**Given** `.env.example` exists with all required variables
**When** I copy `.env.example` to `.env`
**Then** I have a complete reference of all configuration options

### AC2: APP_ENV variable documented for log format control ✅
**Given** `APP_ENV` variable is documented in `.env.example`
**When** a developer reads `.env.example`
**Then** they understand `APP_ENV=development` produces human-readable logs
**Note:** Actual log format implementation is in Story 5.x (Observability)

---

## Tasks / Subtasks

- [x] **Task 1: Complete .env.example** (AC: #1, #2)
  - [x] Add APP_ section (APP_NAME, APP_ENV, APP_PORT) with comments
  - [x] Add DB_ section (HOST, PORT, USER, PASSWORD, NAME, SSL_MODE)
  - [x] Add connection pool vars (DB_MAX_OPEN_CONNS, DB_MAX_IDLE_CONNS, DB_CONN_MAX_LIFETIME)
  - [x] Add OTEL_ section (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME)
  - [x] Add LOG_ section (LOG_LEVEL)
  - [x] Document APP_ENV values (development, staging, production)

- [x] **Task 2: Verify documentation quality** (AC: #1, #2)
  - [x] Each section has header comment
  - [x] Each variable has inline description
  - [x] Default values shown where applicable
  - [x] Matches architecture.md env var specifications

---

## Dev Notes

### Environment Variable Reference (from architecture.md)

| Prefix | Variables |
|--------|-----------|
| APP_ | NAME, ENV, HTTP_PORT |
| DB_ | HOST, PORT, USER, PASSWORD, NAME, SSL_MODE, MAX_OPEN_CONNS, MAX_IDLE_CONNS, CONN_MAX_LIFETIME |
| OTEL_ | EXPORTER_OTLP_ENDPOINT, SERVICE_NAME |
| LOG_ | LEVEL |

### Connection Pool Defaults (architecture.md)

| Setting | Env Var | Default |
|---------|---------|---------|
| Max open | DB_MAX_OPEN_CONNS | 20 |
| Max idle | DB_MAX_IDLE_CONNS | 5 |
| Lifetime | DB_CONN_MAX_LIFETIME | 30m |

### APP_ENV Values

```bash
# APP_ENV values:
# - development: Human-readable logs (console format)
# - staging: JSON logs, reduced verbosity
# - production: JSON logs, structured for aggregation
```

### Expected .env.example Output Format

```bash
# ===========================================
# Application Configuration
# ===========================================
APP_NAME=golang-api-hexagonal
APP_ENV=development  # development | staging | production
APP_PORT=8080

# ===========================================
# Database Configuration (PostgreSQL)
# ===========================================
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=app
DB_SSL_MODE=disable  # disable | require | verify-full

# Connection Pool
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=30m

# ===========================================
# OpenTelemetry Configuration
# ===========================================
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
OTEL_SERVICE_NAME=golang-api-hexagonal

# ===========================================
# Logging Configuration
# ===========================================
LOG_LEVEL=info  # debug | info | warn | error
```

### Out of Scope (Deferred to Epic 2)

- ❌ Config package (`internal/config/`) → Story 2.1-2.5
- ❌ Koanf loader implementation → Story 2.1
- ❌ Config validation → Story 2.3
- ❌ Typed Config struct → Story 2.4

**Rationale:** Epic 2 is dedicated to "Configuration & Environment" with 5 stories covering all config aspects. Story 1.3 only prepares the `.env.example` reference file.

### References

- [Source: docs/architecture.md#Environment-Variables]
- [Source: docs/architecture.md#Connection-Pool]
- [Source: docs/epics.md#Story-1.3]
- [Source: docs/epics.md#Epic-2] - Full config implementation

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Re-scoped by validate-create-story adversarial review (2025-12-11).

### Agent Model Used

dev-story workflow execution.

### Debug Log References

None.

### Completion Notes List

- Story created: 2025-12-11
- Validation applied: 2025-12-11 (MINIMAL scope)
- Implementation completed: 2025-12-11
- Code review fixes applied: 2025-12-11
  - Renamed APP_PORT → APP_HTTP_PORT (architecture compliance)
  - Added LOG_FORMAT variable (per architecture.md LOG_ spec)
  - Improved DB_SSL_MODE documentation
  - Fixed comment style consistency

### File List

Files modified:
- `.env.example` - Complete environment configuration reference (70 lines)
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking

Files NOT created (deferred to Epic 2):
- `internal/config/` → Story 2.1-2.5
