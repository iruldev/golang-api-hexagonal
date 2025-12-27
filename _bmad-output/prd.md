---
stepsCompleted: [1, 2, 3, 4, 7, 8, 9, 10, 11]
inputDocuments:
  - "_bmad-output/technical-go-testing-research-2025-12-27.md"
  - "_bmad-output/brainstorming-session-2025-12-27.md"
  - "docs/index.md"
  - "docs/architecture.md"
  - "docs/testing-guide.md"
workflowType: 'prd'
projectType: 'brownfield'
lastStep: 10
documentCounts:
  briefs: 0
  research: 1
  brainstorming: 1
  projectDocs: 3
classification:
  projectType: 'api_backend'
  domain: 'developer_tools'
  complexity: 'medium'
---

# Product Requirements Document - golang-api-hexagonal

**Author:** Chat  
**Date:** 2025-12-27  
**Project Type:** Brownfield (existing codebase improvement)  
**Version:** Phase 1 - Comprehensive v1

---

## Executive Summary

**golang-api-hexagonal** adalah production-ready Go API yang mengimplementasikan Hexagonal Architecture dengan comprehensive observability stack. Proyek ini sudah memiliki foundation yang solid: 49 test files, 80% coverage threshold, 8 CI quality gates, dan proper layer enforcement via golangci-lint.

Namun, untuk mencapai status **"International-Grade Boilerplate"** yang bisa menjadi referensi lintas tim, diperlukan improvement sistematis di tiga area: **Testing Architecture**, **Reliability Guardrails**, dan **Error Model Hardening**.

PRD ini mendefinisikan **Phase 1: Comprehensive v1** dengan scope:
1. **Testing Architecture Overhaul** - Struktur test yang konsisten, mocks terpusat, anti-flaky integration tests
2. **Reliability Gates** - Testcontainers, race detection, goroutine leak prevention
3. **Error Model Hardening** - Error codes granular, timeout/context audit, Go 1.25 adoption

**Target Timeline:** 3 sprints (~6 minggu)

**Phase 2 Roadmap:** Full "International-Grade Boilerplate vNext" (idempotency keys, feature flags, background jobs skeleton)

### What Makes This Special

Bukan sekadar "cleanup", tapi **menciptakan standar enforceable**:

1. **Golden Path Standard** - Pola test, mocking, integration, dan CI gates yang sama untuk semua Go service
2. **Enforceable Guardrails** - Tooling + CI yang mencegah regresi (bukan dokumentasi saja)
3. **Production-Grade Confidence** - Stabilitas, bukan hanya coverage (anti-flaky, deterministic, observable)
4. **Copy-Paste Ready Boilerplate** - Struktur + ADRs + conventions yang bisa diadopsi service lain dengan minim adaptasi

### Project Classification

| Attribute | Value |
|-----------|-------|
| **Technical Type** | api_backend (existing Go REST API) |
| **Domain** | developer_tools (internal boilerplate/reference) |
| **Complexity** | medium |
| **Project Context** | Brownfield - extending existing system |

---

## Success Criteria

### User Success (Developer Experience)

**"Testing di repo ini enak" when:**
- Satu command jelas: developer tahu persis `make test-unit` vs `make test-integration`
- Test mudah ditemukan: folder structure = intent (unit / integration / contract)
- Debug cepat: failure output jelas (go-cmp diff), subset test mudah (`-run` / subtests)

**Onboarding Targets:**

| Metric | Target |
|--------|--------|
| Setup & run unit tests | ≤ 10 menit |
| Understand test structure | ≤ 30 menit |
| "Aha moment" | Developer sees `internal/testutil/` as standardized toolbox |

### Business Success (Adoption)

**Phase 1 Target:**
- Minimal 1 tim/service lain adopts: `internal/testutil/` structure, CI gates, testcontainers

**Reusability Proof:**
- README "How to write tests here" + template
- Copy-paste kit: `internal/testutil/` + Makefile targets + sample integration test
- Compatibility proof: applied to 1 other service with minimal changes

### Technical Success (Quality & Stability)

**Stability (Priority 1):**

| Metric | Target |
|--------|--------|
| Flaky tests | 0 (no CI reruns needed) |
| `go test -shuffle=on` | 100% pass |
| Goroutine leaks (goleak) | 0 detected |

**Speed (DX & CI Health):**

| Metric | Target |
|--------|--------|
| `make test-unit` | ≤ 5 menit |
| `make test-integration` | ≤ 15 menit |
| Total CI pipeline | ≤ 15 menit |

**Coverage:**
- Maintain ≥80% for domain+app
- +2-5% for currently uncovered areas (e.g., `shared/metrics`)
- Stability > Coverage

### Measurable Outcomes (Guardrails)

| Guardrail | Target |
|-----------|--------|
| Test file size | ≤500 LOC or ≤15 test cases per file |
| Mock consolidation | 100% in `internal/testutil/mocks/` |
| Integration tests | 100% use testcontainers |
| `time.Sleep` in tests | Near 0 (use synctest/explicit waits) |
| Dead/placeholder files | 0 |

### Key Metrics Dashboard

**Top 5 Metrics (Priority Order):**
1. **Flaky rate** → Target: 0
2. **CI duration (core)** → Target: ≤15 menit
3. **Shuffle pass rate** → Target: 100%
4. **Goroutine leaks** → Target: 0
5. **Mock location violations** → Target: 0

---

## Product Scope

### MVP - Phase 1: Comprehensive v1 (3 Sprints)

**Sprint 1: Testing Architecture Overhaul**
- Split jumbo test files (≤500 LOC each)
- Create `internal/testutil/` structure (fixtures, mocks, containers)
- Consolidate all mocks to single location
- Add tests for `shared/metrics/`
- Delete dead/temporary files

**Sprint 2: Integration & Reliability Gates**
- Testcontainers for all DB integration tests
- Build tags (unit vs integration)
- Make targets: `test-unit`, `test-integration`, `test-race`
- CI gates: shuffle, goleak, race (selective)

**Sprint 3: Error Model & Go 1.25**
- Error codes + domain→app→transport mapping
- Timeout & context propagation audit
- Go 1.25 upgrade + synctest adoption
- Replace time.Sleep with deterministic tests

### Growth Features (Phase 2)

- Idempotency key pattern
- Feature flags interface
- Background jobs skeleton
- OpenAPI codegen exploration

### Vision (Future)

- Full "International-Grade Boilerplate vNext"
- Multi-team adoption as official Go service template
- Automated compliance/security scanning

---

## User Journeys

### Journey 1: Budi — New Developer Onboarding

**Persona:** Budi, backend developer baru join tim, belum familiar hexagonal Go.  
**Goal:** Run tests + paham struktur + add test case baru dengan benar.

**The Story:**
Day 1, Budi clone repo. Dia run `make bootstrap` yang install semua tools (mockgen, sqlc, goose) dalam satu command. Kemudian `make test-unit` - green dan cepat. Dia baca "Testing Quickstart" (1 halaman) yang jelaskan folder structure dan cara pakai `internal/testutil`. Dalam 30 menit, dia berhasil add test case baru untuk handler validation dan commit - sesuai standar.

**Success Criteria:**
- ≤10 menit: run `make test-unit` end-to-end
- ≤30 menit: paham "where to put what", commit test baru
- "Aha moment": sees `internal/testutil/` as single source of truth

---

### Journey 2: Dewi — Veteran Maintainer Adding Feature Tests

**Persona:** Dewi, 1 tahun maintain codebase, sering kehambat test structure lama.  
**Goal:** Add fitur + tests tanpa mikir lama, tanpa nambah debt baru.

**The Story:**
Sprint baru, Dewi implement new endpoint. Untuk unit tests, dia ikuti pattern table-driven yang sama. Butuh mock? Ambil dari `internal/testutil/mocks` - tidak nyasar. Butuh DB test? Tambah integration test dengan `//go:build integration` pakai testcontainers. `make test-unit`, `make test-integration`, `make lint` - semua green. PR kecil, mudah review.

**Success Criteria:**
- Langsung tahu tempat yang benar (mocks, fixtures, integration)
- Tidak ada scattered helpers/mocks baru
- PR tetap kecil dan readable

---

### Journey 3: Raka — Adopting Team Lead

**Persona:** Raka, lead backend team department lain, mau standarize testing.  
**Goal:** Adopt patterns ke service lain dengan effort minimal.

**The Story:**
Raka buka "Adoption Guide" (1-2 halaman). Copy `internal/testutil/`, Makefile targets, CI jobs. Minimal changes - cukup module name dan DSN. `make test-unit` green. Dalam 1 hari, service lain sudah adopt standard baru.

**Success Criteria:**
- ≤1 hari: adopt dan CI green
- Perubahan minimal (config only)
- Pattern terasa "copy-paste ready"

---

### Journey 4: CI Pipeline — Automated Quality Gate

**Persona:** Non-human - GitHub Actions.  
**Goal:** Deterministik, signal jelas, cepat.

**The Story:**
Setiap PR: lint → generated checks → `make test-unit` → shuffle → goleak → (nightly) race → integration tests dengan testcontainers. Tidak pernah perlu rerun manual. Core pipeline ≤15 menit.

**Success Criteria:**
- 0 flaky: no manual reruns
- CI core ≤15 menit
- Shuffle 100%, goleak 0, race pass

---

### Journey 5: Andi — DevEx/Platform Owner (Optional)

**Persona:** Andi, maintain CI templates dan standar repo lintas tim.  
**Goal:** Menjaga standar agar tidak regress.

**The Story:**
Andi maintain Makefile targets dan CI templates. Ada rule "mocks hanya satu tempat" yang auto-check di CI. Perubahan standar bisa rollout ke repo lain cepat karena ada version-controlled templates.

**Success Criteria:**
- Guardrails enforced via CI
- Rollout changes ke repo lain cepat

---

### Journey 6: Sari — SRE/On-call Observer (Optional)

**Persona:** Sari, SRE on-call, peduli reliability.  
**Goal:** Confidence bahwa changes tidak nambah production risk.

**The Story:**
Sari lihat test coverage untuk shutdown, context cancellation, timeout configs. Error model konsisten (trace_id, error_code) memudahkan debugging. Lebih sedikit "mystery bugs" karena leaks/races terdeteksi di CI.

**Success Criteria:**
- Fewer mystery bugs
- Faster debugging via error_code + trace_id

---

### Journey Requirements Summary

| Journey | Key Capabilities Required |
|---------|---------------------------|
| Budi (Onboarding) | `make bootstrap`, Testing Quickstart guide, `internal/testutil/` structure |
| Dewi (Veteran) | Centralized mocks, split test files, table-driven patterns |
| Raka (Adoption) | Adoption Guide, copy-paste kit, minimal config changes |
| CI Pipeline | shuffle, goleak, race gates, testcontainers, parallel jobs |
| Andi (DevEx) | CI templates, auto-check rules, version-controlled standards |
| Sari (SRE) | Timeout tests, error model consistency, trace_id propagation |

---

## API Backend Specific Requirements

### Project-Type Overview

**Type:** Go REST API dengan Hexagonal Architecture  
**Existing Foundation:** Chi router, pgx database, Uber Fx DI, OpenTelemetry observability  
**Improvement Focus:** Testing infrastructure, reliability guardrails, error model consistency

### Technical Architecture Considerations

**Testing Infrastructure:**
- **Test Organization:** Centralized at `internal/testutil/` (fixtures, mocks, containers)
- **Mock Generation:** `uber-go/mock` dengan `mockgen`, semua output ke single location
- **Integration Tests:** `testcontainers-go` dengan PostgreSQL module
- **Build Tags:** `//go:build integration` untuk separation
- **Equality Checks:** `go-cmp` untuk readable diffs

**Reliability & Quality Gates:**
- **Race Detection:** `go test -race` (selective/nightly)
- **Shuffle Testing:** `go test -shuffle=on` untuk detect hidden coupling
- **Leak Detection:** `go.uber.org/goleak` di TestMain
- **Concurrency Testing:** Go 1.25 `testing/synctest` untuk deterministic timer tests

**Error Model:**
- **Pattern:** Domain errors → App errors → Transport RFC 7807
- **Trace Propagation:** `trace_id` di semua error responses
- **Error Codes:** Granular codes per error type untuk debugging

### CI/CD Integration Requirements

**GitHub Actions Constraints:**
- Docker access required untuk testcontainers
- Parallelization: unit tests parallel, integration tests sequential
- Secrets: database credentials untuk integration tests (if needed)

**Pipeline Structure:**
```
lint → generated checks → test-unit → shuffle → goleak → (nightly) race → test-integration
```

**Target Durations:**
| Stage | Target |
|-------|--------|
| Unit tests | ≤5 menit |
| Integration tests | ≤15 menit |
| Total pipeline | ≤15 menit |

### Dependencies to Add

| Package | Purpose |
|---------|---------|
| `github.com/google/go-cmp` | Test equality with readable diffs |
| `go.uber.org/mock` | Mock generation (maintained fork) |
| `go.uber.org/goleak` | Goroutine leak detection |
| `github.com/testcontainers/testcontainers-go` | Container-based integration tests |

### Implementation Considerations

**Brownfield Constraints:**
- Maintain existing hexagonal architecture boundaries
- Preserve layer enforcement via golangci-lint depguard
- Keep existing API contracts stable
- Incremental migration (tidak break existing tests)

**Architecture Decision Records (from Research):**
- ADR-001: go-cmp untuk equality
- ADR-002: uber-go/mock + centralized mocks
- ADR-003: testcontainers dengan wait strategies
- ADR-004: synctest untuk concurrency (Go 1.25)
- ADR-005: shuffle/race/goleak gates

---

## Functional Requirements

> **CAPABILITY CONTRACT:** Setiap feature yang dibangun harus trace back ke salah satu requirement ini. Jika capability tidak terdaftar di sini, capability tersebut TIDAK akan ada di final product.

### Test Infrastructure Setup

- **FR1:** Developer dapat run `make bootstrap` untuk setup semua tools (mockgen, sqlc, goose)
- **FR2:** Developer dapat run `make test-unit` untuk execute unit tests
- **FR3:** Developer dapat run `make test-integration` untuk execute integration tests dengan testcontainers
- **FR4:** Developer dapat run `make test-race` untuk execute tests dengan race detector
- **FR5:** System dapat generate mocks via `mockgen` ke lokasi standar `internal/testutil/mocks/`

### Test Organization

- **FR6:** Developer dapat menemukan unit tests per package di `*_test.go` files
- **FR7:** Developer dapat menemukan integration tests via `//go:build integration` tag
- **FR8:** Developer dapat mengakses shared fixtures/helpers di `internal/testutil/`
- **FR9:** Developer dapat run specific test subset via `go test -run TestX/...`
- **FR10:** System memastikan test files tidak melebihi 500 LOC atau 15 test cases

### CI Quality Gates

- **FR11:** CI dapat run `go test -shuffle=on` untuk detect hidden coupling
- **FR12:** CI dapat run goleak verification di TestMain untuk detect goroutine leaks
- **FR13:** CI dapat run race detection (selective/nightly)
- **FR14:** CI dapat report test failures dengan readable diffs (go-cmp)
- **FR15:** CI dapat complete pipeline dalam ≤15 menit

### Integration Testing

- **FR16:** Integration tests dapat spin up PostgreSQL container via testcontainers-go
- **FR17:** Integration tests dapat wait for container ready dengan proper wait strategies
- **FR18:** Integration tests dapat cleanup containers setelah test completion
- **FR19:** Integration tests dapat isolate database state per test

### Error Model

- **FR20:** Domain layer dapat return domain-specific errors dengan error codes
- **FR21:** App layer dapat transform domain errors ke app errors
- **FR22:** Transport layer dapat return RFC 7807 responses dengan trace_id
- **FR23:** Error responses dapat include granular error codes untuk debugging
- **FR24:** System dapat propagate context dan trace_id across layers

### Documentation & Adoption

- **FR25:** Developer dapat baca "Testing Quickstart" guide (≤1 halaman)
- **FR26:** Developer dapat baca "Adoption Guide" untuk copy patterns ke service lain
- **FR27:** System menyediakan copy-paste kit (`internal/testutil/` + Makefile + CI)
- **FR28:** Documentation menyediakan template untuk new test files

### Observability & Reliability

- **FR29:** System dapat test graceful shutdown behavior
- **FR30:** System dapat test context cancellation propagation
- **FR31:** System dapat verify timeout configurations
- **FR32:** System dapat detect deadlock/race conditions via testing

---

## Non-Functional Requirements

### Performance

- **NFR1:** Unit tests (`make test-unit`) selesai dalam ≤5 menit di CI runner standar
- **NFR2:** Integration tests (`make test-integration`) selesai dalam ≤15 menit
- **NFR3:** Total CI pipeline (core) selesai dalam ≤15 menit
- **NFR4:** Testcontainer startup time ≤30 detik per container

### Reliability

- **NFR5:** Flaky test rate = 0 (tidak perlu rerun manual untuk green)
- **NFR6:** `go test -shuffle=on` pass rate = 100%
- **NFR7:** Goroutine leak detection = 0 leaks terdeteksi
- **NFR8:** Race condition detection pass consistently (no false negatives)

### Maintainability

- **NFR9:** Test file size ≤500 LOC atau ≤15 test cases per file
- **NFR10:** All generated mocks in single location (`internal/testutil/mocks/`)
- **NFR11:** No dead/placeholder files in test directories
- **NFR12:** Test patterns consistent across all packages

### Testability

- **NFR13:** New test dapat ditambahkan ≤5 menit oleh developer familiar
- **NFR14:** Test failure messages harus include readable diffs (go-cmp)
- **NFR15:** Integration tests harus reproducible tanpa external dependencies
- **NFR16:** Concurrency tests harus deterministic (no time.Sleep for synchronization)

### Developer Experience

- **NFR17:** Onboarding developer baru ≤30 menit untuk paham test structure
- **NFR18:** Documentation harus up-to-date dengan actual patterns
- **NFR19:** Error messages harus include trace_id untuk debugging
- **NFR20:** Adoption ke service lain ≤1 hari dengan minimal changes

---

<!-- PRD Complete -->





