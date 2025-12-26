Status: done

## Story

As a **platform engineer**,
I want integration tests and documentation for /metrics protection,
so that the security guarantee is verified and discoverable.

**Dependencies:** Story 2.5c (Dual Server)

## Acceptance Criteria

1. **Given** integration test suite
   **When** /metrics accessed on public port
   **Then** returns 404

2. **Given** integration test suite
   **When** /metrics accessed on internal port
   **Then** returns 200 with metrics

3. **And** README documents /metrics protection strategy
4. **And** .env.example has INTERNAL_PORT

## Tasks / Subtasks

- [x] Task 1: Update Integration Tests
  - [x] Test public router /metrics → 404 (`TestNewRouter_MetricsNotExposed`)
  - [x] Test internal router /metrics → 200 (`TestNewInternalRouter_MetricsAvailable`)
  - [x] Update existing metrics tests (`TestMetricsEndpoint` uses internal router)

- [x] Task 2: Update README
  - [x] Add "Internal Port" section
  - [x] Document INTERNAL_PORT env var
  - [x] Describe /metrics protection strategy
  - [x] Add Kubernetes/Docker access examples

- [x] Task 3: Update .env.example
  - [x] Add INTERNAL_PORT=8081 (already done in 2.5a)

- [x] Fix: Code Review Findings
  - [x] Add metrics middleware to internal router (`router.go`)
  - [x] Add `smoke_test.go` to story artifacts
  - [x] Commit implementation files (`main.go`, `config.go`)

## Dev Notes

### Test Coverage (from Story 2.5b)

| Test | Location | Assertion |
|------|----------|-----------|
| `TestNewRouter_MetricsNotExposed` | `router_test.go` | Public /metrics → 404 |
| `TestNewInternalRouter_MetricsAvailable` | `router_test.go` | Internal /metrics → 200 |
| `TestMetricsEndpoint` | `integration_test.go` | Full metrics verification |
| `TestMain_Smoke` | `cmd/api/smoke_test.go` | End-to-end dual server verification |

### README Section Added

Added "🔐 Internal Port (Metrics Protection)" section with:
- Two-server architecture table
- Why separate ports explanation
- Access examples (local, Docker, Kubernetes)
- Configuration table

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro

### Debug Log References
- Tests verified: `TestNewRouter_MetricsNotExposed`, `TestNewInternalRouter_MetricsAvailable`
- Regression: 16 packages ALL PASS

### Completion Notes List
- Task 1: Tests already implemented in Story 2.5b - verified passing
- Task 2: Added comprehensive "Internal Port (Metrics Protection)" section to README
- Task 3: `.env.example` already had INTERNAL_PORT from Story 2.5a
- Code Review Fixes: Applied metrics middleware to internal router, added smoke test

### File List
- `README.md` - MODIFIED (Added Internal Port section)
- `internal/transport/http/router.go` - MODIFIED (Added metrics middleware to internal router)
- `internal/transport/http/router_test.go` - Verified (tests from 2.5b)
- `internal/transport/http/handler/integration_test.go` - Verified (tests from 2.5b)
- `.env.example` - Verified (INTERNAL_PORT from 2.5a)
- `cmd/api/smoke_test.go` - NEW (Smoke test for dual servers)
- `cmd/api/main.go` - MODIFIED (Dual server wiring)
- `internal/infra/config/config.go` - MODIFIED (Internal port config)
- `internal/infra/config/config_test.go` - MODIFIED (Internal port config tests)

### Change Log
- 2024-12-24: Verified integration tests from Story 2.5b cover all acceptance criteria
- 2024-12-24: Added "Internal Port (Metrics Protection)" section to README
- 2024-12-24: Confirmed .env.example already documents INTERNAL_PORT
- 2024-12-24: Fixed code review issues (metrics middleware, untracked files)
