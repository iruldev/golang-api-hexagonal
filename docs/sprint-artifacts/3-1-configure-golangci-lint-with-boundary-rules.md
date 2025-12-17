# Story 3.1: Configure golangci-lint with Boundary Rules

Status: done

## Story

As a **developer**,
I want **linting configured with hexagonal boundary rules**,
So that **import violations are detected locally before push**.

## Acceptance Criteria

1. **Given** I have golangci-lint installed
   **When** I run `make lint`
   **Then** golangci-lint runs with `.golangci.yml` configuration
   **And** `make lint` exits with code 0 when no violations

2. **Given** code in `internal/domain/` imports external package (e.g., `github.com/google/uuid`)
   **When** I run `make lint`
   **Then** depguard rule fails with exit code non-zero
   **And** error message includes:
   - Rule name (e.g., "domain-layer")
   - Forbidden import path
   - File path and line number

3. **Given** code in `internal/app/` imports `net/http` or `github.com/jackc/pgx`
   **When** I run `make lint`
   **Then** depguard rule fails with clear error identifying the boundary violation

4. **Given** code in `internal/transport/` imports `github.com/iruldev/golang-api-hexagonal/internal/infra/observability`
   **When** I run `make lint`
   **Then** depguard rule fails with clear error identifying the boundary violation

5. **Given** code in `internal/infra/` imports `github.com/iruldev/golang-api-hexagonal/internal/app`
   **When** I run `make lint`
   **Then** depguard rule fails with clear error identifying the boundary violation

*Covers: FR45, FR47*

## Tasks / Subtasks

- [x] Task 1: Create `.golangci.yml` configuration file (AC: #1)
  - [x] Configure default linters (depguard, govet, errcheck, staticcheck, ineffassign, unused)
  - [x] Configure timeout and concurrency settings
  - [x] Set output format for clear error messages

- [x] Task 2: Configure depguard for domain layer boundary (AC: #2)
  - [x] Define "domain-layer" rule applying to `internal/domain/`
  - [x] Allow stdlib only (use depguard `$gostd` variable; do not enumerate stdlib packages)
  - [x] Block external packages and stdlib logging (`log/slog`) in domain
  - [x] Verify error message includes rule name and forbidden import path

- [x] Task 3: Configure depguard for app layer boundary (AC: #3)
  - [x] Define "app-layer" rule applying to `internal/app/`
  - [x] Allow domain imports: `github.com/iruldev/golang-api-hexagonal/internal/domain...`
  - [x] Allow stdlib EXCEPT HTTP packages (deny `net/http` and related)
  - [x] Block external/runtime concerns: `github.com/jackc/pgx/v5...`, `go.opentelemetry.io...`, `github.com/google/uuid`, `log/slog`, `internal/transport...`, `internal/infra...`

- [x] Task 4: Configure depguard for transport layer boundary (AC: #4)
  - [x] Define "transport-layer" rule applying to `internal/transport/`
  - [x] Allow: domain, app, shared, stdlib, router/mux (`github.com/go-chi/chi/v5`), UUID (`github.com/google/uuid`), OTEL (`go.opentelemetry.io/otel...`)
  - [x] Block: DB driver (`github.com/jackc/pgx/v5...`), direct `internal/infra...` imports

- [x] Task 5: Configure depguard for infra layer boundary (AC: #5)
  - [x] Define "infra-layer" rule applying to `internal/infra/`
  - [x] Allow: domain, infra dependencies (pgx/otel/prometheus/etc), external packages as needed
  - [x] Block: `internal/app...`, `internal/transport...`

- [x] Task 6: Verify `make lint` target works (AC: #1, #2, #3, #4, #5)
  - [x] Run `make lint` successfully when no violations
  - [x] Test with intentional violation in domain layer
  - [x] Test with intentional violation in app layer
  - [x] Test with intentional violation in transport layer
  - [x] Test with intentional violation in infra layer
  - [x] Remove test violations after verification

- [x] Task 7: Update setup documentation if needed (N/A)
  - [x] No README/update required (lint already available via `make lint` and config lives in `.golangci.yml`)

## Dev Notes

### Architecture Context

This story enforces the hexagonal architecture boundaries defined in `docs/project-context.md`:

```
Domain Layer (internal/domain/):
  ✅ ALLOWED: stdlib only (errors, context, time, strings, fmt)
  ❌ FORBIDDEN: slog, uuid, pgx, chi, otel, ANY external package

App Layer (internal/app/):
  ✅ ALLOWED: domain imports only
  ❌ FORBIDDEN: net/http, pgx, slog, otel, uuid, transport, infra

Transport Layer (internal/transport/http/):
  ✅ ALLOWED: domain, app, chi, uuid, stdlib
  ❌ FORBIDDEN: pgx, direct infra/observability imports

Infra Layer (internal/infra/):
  ✅ ALLOWED: domain, pgx, slog, otel, uuid, external packages
  ❌ FORBIDDEN: app, transport
```

### depguard Configuration Reference

This repo uses `golangci-lint` v2.x (currently installed locally as `v2.1.6`). The depguard configuration uses the `linters.settings.depguard` structure (v2 config schema).

```yaml
linters:
  settings:
    depguard:
      rules:
        domain-layer:
          # Use `strict` + `$gostd` to enforce "stdlib only".
          list-mode: strict
          files:
            - "internal/domain/**"
          allow:
            - $gostd
```

For other layers, use allow-lists + targeted deny entries (e.g., deny `net/http` for `internal/app/**` while still allowing other stdlib packages).

### Import Path Mapping (Use in depguard allow/deny)

Use real Go import paths (depguard matches imports, not nicknames):

- `slog` → `log/slog`
- `uuid` → `github.com/google/uuid`
- `pgx` → `github.com/jackc/pgx/v5` (and subpackages like `.../pgxpool`)
- `chi` → `github.com/go-chi/chi/v5` (and `.../middleware`)
- `otel` → `go.opentelemetry.io/otel` (and subpackages)

### Shared Interface Pattern (from Epic 2 Retrospective)

Epic 2 introduced `internal/shared/` for cross-layer communication to avoid import violations:
- `internal/shared/metrics/http_metrics.go` - allows transport to use metrics without importing infra

This pattern should be allowed in depguard rules for transport layer.

### File Locations

| Action | Path |
|--------|------|
| Create | `.golangci.yml` (project root) |
| Verify | `Makefile` already has `lint` target |
| Reference | `docs/project-context.md` for layer rules |
| Reference | `docs/architecture.md` for complete boundary specs |

### Testing Strategy

1. **Positive test**: Run `make lint` - should pass with exit 0
2. **Negative test**: Temporarily add invalid import, verify lint fails
3. **Error message test**: Verify output includes rule name and location

### Technology Specifics

- **golangci-lint version**: v2.x (tested with v2.1.6 in this repo)
- **depguard**: Built into golangci-lint; supports rule-based `allow`/`deny` with variables like `$gostd`
- **Configuration format**: YAML in `.golangci.yml`

### Learnings from Epic 2

From Epic 2 Retrospective action items:
1. "Research depguard configuration for boundary rules" - this story addresses this
2. "Document shared interface pattern in architecture docs" - relevant for transport layer rules

### References

- [Source: docs/project-context.md#Critical-Layer-Rules]
- [Source: docs/architecture.md#Architectural-Boundaries]
- [Source: docs/epics.md#Story-3.1]
- [Source: docs/sprint-artifacts/epic-2-retro-2025-12-17.md#Epic-3-Preparation]

## Dev Agent Record

### Context Reference

- `docs/project-context.md` - Layer import rules
- `docs/architecture.md` - Complete architecture decisions
- `docs/sprint-artifacts/epic-2-retro-2025-12-17.md` - Previous learnings

### Agent Model Used

GPT-5 (Codex CLI)

### Debug Log References

### Completion Notes List

### File List

- `.golangci.yml` (create)
- `Makefile` (verify only; change only if needed for lint config discovery)
- `README.md` (optional; only if you decide to document lint setup)
- `internal/transport/http/handler/integration_test.go` (update tests to avoid transport → infra imports)
- `internal/transport/http/middleware/metrics_test.go` (update tests to avoid transport → infra imports)
- `docs/project-context.md` (clarify transport allowed dependencies)
- `docs/sprint-artifacts/sprint-status.yaml` (sync story status)
