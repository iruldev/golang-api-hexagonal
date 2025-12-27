---
stepsCompleted: [1, 2, 3]
inputDocuments:
  - "_bmad-output/prd.md"
  - "_bmad-output/architecture.md"
---

# golang-api-hexagonal - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for **golang-api-hexagonal**, decomposing the requirements from the PRD and Architecture into implementable stories focused on **Testing Architecture + Reliability + Error Model** improvements.

**Project Type:** Brownfield  
**Timeline:** 3 Sprints

---

## Requirements Inventory

### Functional Requirements (32 FRs)

#### Test Infrastructure Setup (5 FRs)
- **FR1:** Developer dapat run `make bootstrap` untuk setup semua tools (mockgen, sqlc, goose)
- **FR2:** Developer dapat run `make test-unit` untuk execute unit tests
- **FR3:** Developer dapat run `make test-integration` untuk execute integration tests dengan testcontainers
- **FR4:** Developer dapat run `make test-race` untuk execute tests dengan race detector
- **FR5:** System dapat generate mocks via `mockgen` ke lokasi standar `internal/testutil/mocks/`

#### Test Organization (5 FRs)
- **FR6:** Developer dapat menemukan unit tests per package di `*_test.go` files
- **FR7:** Developer dapat menemukan integration tests via `//go:build integration` tag
- **FR8:** Developer dapat mengakses shared fixtures/helpers di `internal/testutil/`
- **FR9:** Developer dapat run specific test subset via `go test -run TestX/...`
- **FR10:** System memastikan test files tidak melebihi 500 LOC atau 15 test cases

#### CI Quality Gates (5 FRs)
- **FR11:** CI dapat run `go test -shuffle=on` untuk detect hidden coupling
- **FR12:** CI dapat run goleak verification di TestMain untuk detect goroutine leaks
- **FR13:** CI dapat run race detection (selective/nightly)
- **FR14:** CI dapat report test failures dengan readable diffs (go-cmp)
- **FR15:** CI dapat complete pipeline dalam ≤15 menit

#### Integration Testing (4 FRs)
- **FR16:** Integration tests dapat spin up PostgreSQL container via testcontainers-go
- **FR17:** Integration tests dapat wait for container ready dengan proper wait strategies
- **FR18:** Integration tests dapat cleanup containers setelah test completion
- **FR19:** Integration tests dapat isolate database state per test

#### Error Model (5 FRs)
- **FR20:** Domain layer dapat return domain-specific errors dengan error codes
- **FR21:** App layer dapat transform domain errors ke app errors
- **FR22:** Transport layer dapat return RFC 7807 responses dengan trace_id
- **FR23:** Error responses dapat include granular error codes untuk debugging
- **FR24:** System dapat propagate context dan trace_id across layers

#### Documentation & Adoption (4 FRs)
- **FR25:** Developer dapat baca "Testing Quickstart" guide (≤1 halaman)
- **FR26:** Developer dapat baca "Adoption Guide" untuk copy patterns ke service lain
- **FR27:** System menyediakan copy-paste kit (`internal/testutil/` + Makefile + CI)
- **FR28:** Documentation menyediakan template untuk new test files

#### Observability & Reliability (4 FRs)
- **FR29:** System dapat test graceful shutdown behavior
- **FR30:** System dapat test context cancellation propagation
- **FR31:** System dapat verify timeout configurations
- **FR32:** System dapat detect deadlock/race conditions via testing

---

### Non-Functional Requirements (20 NFRs)

#### Performance (4 NFRs)
- **NFR1:** Unit tests selesai dalam ≤5 menit di CI runner standar
- **NFR2:** Integration tests selesai dalam ≤15 menit
- **NFR3:** Total CI pipeline selesai dalam ≤15 menit
- **NFR4:** Testcontainer startup time ≤30 detik per container

#### Reliability (4 NFRs)
- **NFR5:** Flaky test rate = 0 (tidak perlu rerun manual untuk green)
- **NFR6:** `go test -shuffle=on` pass rate = 100%
- **NFR7:** Goroutine leak detection = 0 leaks terdeteksi
- **NFR8:** Race condition detection pass consistently

#### Maintainability (4 NFRs)
- **NFR9:** Test file size ≤500 LOC atau ≤15 test cases per file
- **NFR10:** All generated mocks in single location (`internal/testutil/mocks/`)
- **NFR11:** No dead/placeholder files in test directories
- **NFR12:** Test patterns consistent across all packages

#### Testability (4 NFRs)
- **NFR13:** New test dapat ditambahkan ≤5 menit oleh developer familiar
- **NFR14:** Test failure messages include readable diffs (go-cmp)
- **NFR15:** Integration tests reproducible tanpa external dependencies
- **NFR16:** Concurrency tests deterministic (no time.Sleep)

#### Developer Experience (4 NFRs)
- **NFR17:** Onboarding developer baru ≤30 menit untuk paham test structure
- **NFR18:** Documentation up-to-date dengan actual patterns
- **NFR19:** Error messages include trace_id untuk debugging
- **NFR20:** Adoption ke service lain ≤1 hari

---

### Additional Requirements (from Architecture)

**Technical Requirements:**
- Brownfield project - preserve existing hexagonal architecture
- 10 ADRs locked (5 from research + 5 from session)
- 11 Implementation Patterns must be followed
- 4 CI Enforcement Scripts defined

**Dependencies to Add:**
- `github.com/google/go-cmp` - Test equality
- `go.uber.org/mock` - Mock generation
- `go.uber.org/goleak` - Goroutine leak detection
- `github.com/testcontainers/testcontainers-go` - Container-based tests

**Starter Template:** N/A (brownfield, extending existing codebase)

---

### FR Coverage Map

{{requirements_coverage_map}}

---

### Implementation Notes (Refinements)

> **These refinements clarify ambiguity for epic/story creation:**

1. **FR10/NFR9 (Jumbo Test Gate):** Enforcement is **touched-files/hotspots only** in Phase 1 (not repo-wide) to avoid noisy CI.

2. **FR13/NFR8 (Race Policy):** Race detection is **nightly/selective** (not PR gate). Selective packages defined in `scripts/race_packages.txt` or via `make test-race-selective`.

3. **FR16-19 (Integration Helpers):** Golden path helpers required:
   - `containers.NewPostgres(t)` + `containers.Migrate(t, db)`
   - Isolation default: tx+rollback; fallback: truncate/reset helper

4. **FR1 (Bootstrap):** Toolchain pinning via `tools.go` + `go install @version` (reproducible), not manual install.

5. **FR20-23 (Error Model):** **Backward compatible** RFC7807 envelope - new fields allowed, no remove/rename. Error codes are **stable** once defined.

6. **Anti-pattern: Scattered Mocks:** Detection via **MockGen marker** ("Code generated by MockGen") + path check, not just `*_mock.go` pattern.

---

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| FR1-FR10 | Epic 1 | Test Infrastructure + Organization |
| FR11-FR15 | Epic 2 | CI Quality Gates |
| FR16-FR19 | Epic 2 | Integration Testing |
| FR20-FR24 | Epic 3 | Error Model |
| FR25 | Epic 1 | Testing Quickstart (moved from Epic 3) |
| FR26-FR28 | Epic 3 | Documentation & Adoption |
| FR29-FR32 | Epic 2 | Observability & Reliability (moved from Epic 3) |

---

## Epic List

### Epic 1: Testing Infrastructure Foundation + Onboarding

**Goal:** Developer dapat bootstrap & run unit tests dengan struktur benar, langsung ngerti "where to put what".

**User Outcome:** "Aku bisa `make bootstrap` lalu `make test-unit` dan semuanya works first time. Quickstart guide bikin aku paham struktur."

**FRs covered:** FR1–FR10 + FR25 (11 FRs)

**Stories:**
- 1.1: Create `internal/testutil/` structure (assert, containers, fixtures, mocks)
- 1.2: Setup mock generation with uber-go/mock + centralized mocks
- 1.3: Implement shared TestMain helper (goleak)
- 1.4: Create Makefile test targets (test-unit, shuffle, mocks, gencheck)
- 1.5: Add tools.go for reproducible toolchain
- 1.6: Write Testing Quickstart guide (≤1 halaman)

---

### Epic 2: Integration Testing & CI Determinism + Reliability Tests

**Goal:** CI green = beneran reliable; integration reproducible via testcontainers; reliability behaviors teruji.

**User Outcome:** "CI hijau = benar-benar green, tidak flaky, integration tests jalan di container, dan graceful shutdown sudah teruji."

**FRs covered:** FR11–FR19 + FR29–FR32 (13 FRs)

**Stories:**
- 2.1: Setup testcontainers-go for PostgreSQL
- 2.2: Implement container helpers (containers subpackage: NewPostgres, Migrate, WithTx)
- 2.3: Add PR gates: shuffle + goleak + gencheck
- 2.4: Add race detection policy (nightly/selective)
- 2.5: Update GitHub Actions workflows (PR + nightly)
- 2.6: Implement graceful shutdown test
- 2.7: Implement context cancellation propagation test
- 2.8: Implement timeout configs verification test

---

### Epic 3: Error Model Hardening + Adoption Kit (incl. Go 1.25 enabler)

**Goal:** Error envelope konsisten dari domain ke transport + adoptable ke repo lain ≤1 hari.

**User Outcome:** "Error responses bagus untuk debugging dengan trace_id, dan aku bisa copy patterns ini ke service lain dalam 1 hari."

**FRs covered:** FR20–FR28 (8 FRs) + Go 1.25 upgrade

---

# Story Details

## Epic 1: Testing Infrastructure Foundation + Onboarding

### Story 1.1: Create testutil Structure

**As a** developer, **I want** a well-organized `internal/testutil/` directory, **So that** I know where to find and add test helpers.

**ACs:**
- `internal/testutil/` exists with subpackages: `assert/`, `containers/`, `fixtures/`, `mocks/`
- Each subpackage has a placeholder file with package doc
- `testutil.go` contains common helpers (context, timeout)

**FRs:** FR8

---

### Story 1.2: Mock Generation with uber-go/mock

**As a** developer, **I want** centralized mock generation, **So that** all mocks are consistent and in one place.

**ACs:**
- `go.uber.org/mock` added to dependencies
- `//go:generate` directives in port interfaces
- All mocks generated to `internal/testutil/mocks/`
- `make mocks` target generates all mocks
- Mocks contain `Code generated by MockGen` marker

**FRs:** FR5

---

### Story 1.3: Shared TestMain Helper (goleak)

**As a** developer, **I want** automatic goroutine leak detection, **So that** I catch leaks early.

**ACs:**
- `go.uber.org/goleak` added to dependencies
- `testutil.RunWithGoleak(m)` helper created
- Integration test packages use `TestMain` with goleak
- goleak runs with `IgnoreCurrent()` for known background goroutines

**FRs:** FR12, NFR7

---

### Story 1.4: Makefile Test Targets

**As a** developer, **I want** simple make commands for testing, **So that** I don't need to remember go test flags.

**ACs:**
- `make test-unit` runs unit tests with coverage
- `make test-shuffle` runs with `-shuffle=on`
- `make gencheck` verifies generated files are up-to-date
- `make test` runs unit + shuffle
- Coverage report generated to `coverage.out`

**FRs:** FR2, FR11

---

### Story 1.5: tools.go for Reproducible Toolchain

**As a** developer, **I want** pinned tool versions, **So that** everyone uses same versions.

**ACs:**
- `tools/tools.go` with blank imports for: mockgen, sqlc, goose, golangci-lint
- `make bootstrap` runs `go install` for all tools
- Tool versions pinned via go.mod
- README documents `make bootstrap` as first step

**FRs:** FR1

---

### Story 1.6: Testing Quickstart Guide

**As a** new developer, **I want** a one-page testing guide, **So that** I understand the test structure in <30 minutes.

**ACs:**
- `docs/testing-quickstart.md` exists (≤1 page)
- Covers: directory structure, make targets, naming conventions
- Includes copy-paste examples for unit and integration tests
- Links to detailed docs for advanced topics

**FRs:** FR25, NFR17

---

## Epic 2: Integration Testing & CI Determinism + Reliability Tests

### Story 2.1: testcontainers-go for PostgreSQL

**As a** developer, **I want** containerized database tests, **So that** tests are reproducible without external setup.

**ACs:**
- `github.com/testcontainers/testcontainers-go` added
- `containers.NewPostgres(t)` returns working pool
- Container auto-cleanup via `t.Cleanup()`
- Container starts in ≤30 seconds

**FRs:** FR16, NFR4

---

### Story 2.2: Container Helpers Package

**As a** developer, **I want** golden-path container helpers, **So that** all integration tests use consistent patterns.

**ACs:**
- `containers.Migrate(t, pool)` applies goose migrations
- `containers.WithTx(t, pool, fn)` provides tx+rollback isolation
- `containers.Truncate(t, pool, tables...)` for truncate fallback
- Helpers documented in testutil/containers/README.md

**FRs:** FR17, FR18, FR19

---

### Story 2.3: PR Gates (shuffle + goleak + gencheck)

**As a** CI system, **I want** quality gates on every PR, **So that** hidden coupling and leaks are caught early.

**ACs:**
- GitHub Actions runs `make test-shuffle` on PRs
- GitHub Actions runs `make gencheck` on PRs
- goleak verification happens via TestMain
- Failed gates block PR merge

**FRs:** FR11, FR12, FR14, NFR6

---

### Story 2.4: Race Detection Policy (nightly/selective)

**As a** CI system, **I want** race detection without slowing PRs, **So that** races are caught without blocking velocity.

**ACs:**
- `make test-race-selective` runs race on packages in `scripts/race_packages.txt`
- Nightly workflow runs full race detection
- `scripts/race_packages.txt` lists high-risk packages
- Race failures notify team via workflow alert

**FRs:** FR4, FR13, NFR8

---

### Story 2.5: GitHub Actions Workflows (PR + Nightly)

**As a** maintainer, **I want** comprehensive CI workflows, **So that** quality is enforced automatically.

**ACs:**
- `ci.yml` runs on PRs: lint, test-unit, test-shuffle, gencheck
- `nightly.yml` runs daily: test-race, test-integration
- Total PR pipeline ≤15 minutes
- Integration tests use testcontainers

**FRs:** FR15, NFR1, NFR2, NFR3

---

### Story 2.6: Graceful Shutdown Test

**As a** SRE, **I want** tested graceful shutdown, **So that** I trust the app handles SIGTERM properly.

**ACs:**
- Integration test sends SIGTERM to app
- Test verifies in-flight requests complete
- Test verifies DB connections are closed
- Test times out after configurable duration

**FRs:** FR29

---

### Story 2.7: Context Cancellation Propagation Test

**As a** developer, **I want** tested context cancellation, **So that** I trust cancellation propagates correctly.

**ACs:**
- Test creates request with cancellable context
- Test cancels context mid-request
- Test verifies downstream operations are cancelled
- Test verifies no goroutine leaks after cancel

**FRs:** FR30, FR24

---

### Story 2.8: Timeout Configs Verification Test

**As a** developer, **I want** tested timeout configs, **So that** I trust timeouts are applied correctly.

**ACs:**
- Test verifies HTTP server timeout is applied
- Test verifies DB query timeout is applied
- Test verifies graceful shutdown timeout is applied
- Timeouts are configurable via env

**FRs:** FR31

---

## Epic 3: Error Model Hardening + Adoption Kit

### Story 3.1: Domain Error Types + Stable Codes

**As a** developer, **I want** well-defined domain errors, **So that** error handling is consistent.

**ACs:**
- `internal/domain/errors/` package with error types
- Error codes format: `ERR_{DOMAIN}_{CODE}` (e.g., `ERR_USER_NOT_FOUND`)
- `errors.Is` / `errors.As` work correctly
- Error codes are documented and stable (no breaking changes)

**FRs:** FR20, FR23

---

### Story 3.2: App→Transport Error Mapping

**As a** developer, **I want** consistent error mapping, **So that** domain errors become proper HTTP responses.

**ACs:**
- Mapping table: domain error → HTTP status
- Unit tests cover all error mappings
- Unknown errors map to 500
- Mapping preserves error code in response

**FRs:** FR21

---

### Story 3.3: RFC 7807 Response with trace_id

**As a** API consumer, **I want** standardized error responses, **So that** I can debug issues easily.

**ACs:**
- Error responses follow RFC 7807 format
- Response includes: type, title, status, detail, trace_id
- trace_id propagated from OpenTelemetry
- Error codes included as `code` extension field

**FRs:** FR22, FR24, NFR19

---

### Story 3.4: Adoption Guide + Copy-Paste Kit

**As a** adopting team lead, **I want** a complete adoption guide, **So that** I can bring these patterns to my service.

**ACs:**
- `docs/adoption-guide.md` with checklist
- Copy-paste kit: `internal/testutil/`, Makefile snippets, CI workflow
- Step-by-step migration guide for brownfield
- Expected adoption time: ≤1 day

**FRs:** FR26, FR27, NFR20

---

### Story 3.5: Doc Template for New Test Files

**As a** developer, **I want** test file templates, **So that** I write tests consistently.

**ACs:**
- `docs/templates/unit_test.go.tmpl` example
- `docs/templates/integration_test.go.tmpl` example
- Templates include proper imports, TestMain, naming

**FRs:** FR28

---

### Story 3.6: Go 1.25 Upgrade + Targeted synctest

**As a** developer, **I want** Go 1.25 with synctest, **So that** time-based tests are deterministic.

**ACs:**
- go.mod updated to Go 1.25 (when released)
- 1-2 existing flaky time-based tests refactored to use `testing/synctest`
- synctest usage documented in testing guide
- Fallback: if Go 1.25 not released, document as future work

**FRs:** FR32, NFR16

---

# Summary

| Epic | Stories | FRs Covered |
|------|---------|-------------|
| Epic 1 | 1.1–1.6 (6) | FR1-10, FR25 |
| Epic 2 | 2.1–2.8 (8) | FR11-19, FR29-32 |
| Epic 3 | 3.1–3.6 (6) | FR20-28 |
| **Total** | **20** | **32** |



