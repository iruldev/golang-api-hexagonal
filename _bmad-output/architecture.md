---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8, 9]
inputDocuments:
  - "_bmad-output/prd.md"
  - "_bmad-output/technical-go-testing-research-2025-12-27.md"
  - "_bmad-output/brainstorming-session-2025-12-27.md"
  - "docs/index.md"
  - "docs/architecture.md"
workflowType: 'architecture'
project_name: 'golang-api-hexagonal'
user_name: 'Chat'
date: '2025-12-27'
projectType: 'brownfield'
---

# Architecture Decision Document

**Project:** golang-api-hexagonal  
**Author:** Chat  
**Date:** 2025-12-27  
**Project Type:** Brownfield - Testing Architecture + Reliability + Error Model Improvement

---

## Project Context Analysis

### Requirements Overview

**Functional Requirements (32 FRs)** organized into 7 capability areas:
1. Test Infrastructure Setup (5 FRs) - make targets, mock generation
2. Test Organization (5 FRs) - structure, build tags, testutil
3. CI Quality Gates (5 FRs) - shuffle, goleak, race detection
4. Integration Testing (4 FRs) - testcontainers, isolation
5. Error Model (5 FRs) - domain→app→transport mapping, trace_id
6. Documentation & Adoption (4 FRs) - guides, copy-paste kit
7. Observability & Reliability (4 FRs) - shutdown, context, timeouts

**Non-Functional Requirements (20 NFRs)** in 5 categories:
- **Performance:** ≤5 min unit, ≤15 min integration, ≤15 min CI total
- **Reliability:** 0 flaky tests, 100% shuffle pass, 0 goroutine leaks
- **Maintainability:** ≤500 LOC/file, centralized mocks, consistent patterns
- **Testability:** deterministic, readable diffs, no time.Sleep
- **Developer Experience:** ≤30 min onboarding, ≤1 day adoption

### Scale & Complexity

| Indicator | Value |
|-----------|-------|
| Project Complexity | Medium |
| Primary Domain | Developer Tools / API Backend |
| Technical Domain | Go, Testing, CI/CD |
| Cross-cutting Concerns | Error handling, Observability, CI gates |

### Technical Constraints & Dependencies

**Existing Stack (must preserve):**
- Go 1.24.11 (→ 1.25 upgrade planned in Sprint 3)
- Chi v5.2.3, pgx v5.7.6, Uber Fx
- OpenTelemetry, Prometheus, slog
- Hexagonal Architecture with layer enforcement

**New Dependencies (from Research ADRs):**
- `github.com/google/go-cmp` - Test equality
- `go.uber.org/mock` - Mock generation
- `go.uber.org/goleak` - Goroutine leak detection
- `github.com/testcontainers/testcontainers-go` - Container-based tests

**CI Constraints:**
- GitHub Actions with Docker access
- ≤15 min total pipeline target
- Parallelization: unit parallel, integration sequential

### Cross-Cutting Concerns

1. **Error Propagation** - Domain → App → Transport mapping
2. **Context/Timeout** - Proper propagation and cancellation
3. **Test Organization** - Centralized testutil across packages
4. **Mock Strategy** - Single location, generated mocks
5. **CI Integration** - Makefile targets + GitHub Actions

### Existing Architecture Decisions (from Research)

| ADR | Decision | Rationale |
|-----|----------|-----------|
| ADR-001 | go-cmp for equality | Readable diffs, better than reflect.DeepEqual |
| ADR-002 | uber-go/mock + centralized | Maintained fork, single location |
| ADR-003 | testcontainers + wait strategies | Reproducible, no external deps |
| ADR-004 | synctest for concurrency | Go 1.25, deterministic timer tests |
| ADR-005 | shuffle/race/goleak gates | Quality gates for stability |

---

## Starter Template Evaluation

### Decision: N/A (Brownfield Project)

This is a **brownfield project** - extending an existing codebase, not creating a new project. No starter template evaluation needed.

### Existing Technology Stack (Confirmed)

| Category | Technology | Version | Status |
|----------|------------|---------|--------|
| Language | Go | 1.24.11 → 1.25 | ✅ Preserve |
| Router | Chi | v5.2.3 | ✅ Preserve |
| Database | pgx | v5.7.6 | ✅ Preserve |
| DI | Uber Fx | v1.23.0 | ✅ Preserve |
| Observability | OpenTelemetry | v1.33.0 | ✅ Preserve |
| Metrics | Prometheus | current | ✅ Preserve |
| Logging | slog | stdlib | ✅ Preserve |
| Linting | golangci-lint | v1.64.2 | ✅ Preserve |
| Migrations | goose | v3 | ✅ Preserve |
| SQL Gen | sqlc | v1.28.0 | ✅ Preserve |

### New Dependencies to Add

| Package | Purpose | Action |
|---------|---------|--------|
| `github.com/google/go-cmp` | Test equality with readable diffs | Add |
| `go.uber.org/mock` | Mock generation (maintained fork) | Add |
| `go.uber.org/goleak` | Goroutine leak detection | Add |
| `github.com/testcontainers/testcontainers-go` | Container-based integration tests | Add |

### Initialization Command

```bash
# No project initialization needed
# Changes extend existing structure in internal/testutil/ and CI/Makefile
```

### Key Architectural Decisions Already Made

- ✅ Hexagonal Architecture pattern (domain/app/infra layers)
- ✅ Layer enforcement via golangci-lint depguard
- ✅ DI pattern via Uber Fx
- ✅ Error handling pattern (domain → app → transport)
- ✅ Observability via OpenTelemetry + Prometheus

---

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Already Locked from Research):**
- ADR-001 to ADR-005 (testing infrastructure choices)
- Existing hexagonal architecture preservation

**Important Decisions (Locked in this Session):**
- Test directory structure
- Mock generation strategy
- Build tags convention
- TestMain pattern
- Timeout configuration

### Decision 1: Test Directory Structure — Structured Subpackages

**Decision:** Organized subpackages under `internal/testutil/`

```
internal/testutil/
  assert/        # cmp helpers + options
  containers/    # testcontainers helpers (postgres, migrations)
  fixtures/      # builders/factories
  mocks/         # generated mocks (single location)
  testutil.go    # common helpers (random, context, http test server)
```

**Rationale:** Single source of truth, scalable, Go-idiomatic separate imports.

**Usage:**
```go
import (
  "yourmod/internal/testutil/fixtures"
  "yourmod/internal/testutil/mocks"
  "yourmod/internal/testutil/assert"
)
```

### Decision 2: Mock Generation Strategy — Both (go:generate + Makefile)

**Decision:** go:generate near interfaces + Makefile aggregator

**go:generate Pattern:**
```go
//go:generate mockgen -source=../ports/user_repo.go -destination=../../testutil/mocks/user_repo_mock.go -package=mocks
```

**Makefile Aggregator:**
```make
.PHONY: mocks
mocks:
	go generate ./...

.PHONY: gencheck
gencheck:
	git diff --exit-code || (echo "Generated files out of date. Run: make mocks" && exit 1)
```

**Tools Version Pin (tools.go):**
```go
//go:build tools
package tools

import (
  _ "go.uber.org/mock/mockgen"
)
```

**Rationale:** Source-of-truth near interface (maintainable), single command for CI (reproducible).

### Decision 3: Build Tags Convention — `integration`

**Decision:** Standard `integration` tag for integration tests.

**Test File Header:**
```go
//go:build integration
// +build integration
```

**Makefile Targets:**
```make
test-unit:
	go test -short -timeout $(UNIT_TIMEOUT) ./...

test-integration:
	go test -tags=integration -timeout $(INTEGRATION_TIMEOUT) ./...
```

**Rationale:** Clear, standard, future-proof (can add `e2e` in Phase 2).

### Decision 4: TestMain Pattern — Shared Helper

**Decision:** Centralized goleak helper in testutil.

**Helper Implementation:**
```go
// internal/testutil/testmain.go
package testutil

import (
  "os"
  "testing"
  "go.uber.org/goleak"
)

func RunWithGoleak(m *testing.M, opts ...goleak.Option) {
  goleak.VerifyTestMain(m, opts...)
  os.Exit(m.Run())
}
```

**Usage in Test Packages:**
```go
func TestMain(m *testing.M) {
  testutil.RunWithGoleak(m)
}
```

**Rationale:** Consistent enforcement, anti-copy-paste, easy to extend.

### Decision 5: Timeout Configuration — Configurable with Defaults

**Decision:** Environment-configurable with sensible defaults.

**Defaults:**
- Unit tests: 30s (fail fast)
- Integration tests: 5m (container startup time)

**Makefile Configuration:**
```make
UNIT_TIMEOUT ?= 30s
INTEGRATION_TIMEOUT ?= 5m

ifdef TEST_TIMEOUT
UNIT_TIMEOUT := $(TEST_TIMEOUT)
INTEGRATION_TIMEOUT := $(TEST_TIMEOUT)
endif
```

**Rationale:** Matches NFR speed targets, flexible for debugging/CI tuning.

### Decision Impact Analysis

**Implementation Sequence:**
1. Create `internal/testutil/` subpackage structure
2. Add tools.go with mockgen
3. Implement shared TestMain helper
4. Create Makefile targets with timeouts
5. Migrate existing tests to new patterns

**Cross-Component Dependencies:**
- Mocks depend on testutil/mocks location
- Integration tests depend on containers subpackage
- CI depends on Makefile targets
- All tests depend on TestMain pattern

---

## Implementation Patterns & Consistency Rules

### Pattern Summary

| # | Pattern | Purpose |
|---|---------|---------|
| 1 | Test File Naming | Distinguish unit vs integration |
| 2 | Mock Naming | Consistent generated mocks |
| 3 | Test Function Naming | Clear intent |
| 4 | Fixture Naming | Builder/factory patterns |
| 5 | Error Pattern | Domain error consistency |
| 6 | Test Package Policy | Whitebox vs blackbox |
| 7 | t.Parallel() Policy | Safe parallelization |
| 8 | Assertion Style | Consistent assertions |
| 9 | Integration Layout | Container isolation |
| 10 | Deterministic Time | Anti time.Sleep |
| 11 | Generated Code | Marker-based detection |

### Pattern 1: Test File Naming

- `*_test.go` - unit tests (default, no build tag)
- `*_integration_test.go` - integration tests with `//go:build integration`
- `*_mock.go` - generated mocks (in `internal/testutil/mocks/`)

### Pattern 2: Mock Naming Convention

- **File:** `{interface_name}_mock.go` (e.g., `user_repo_mock.go`)
- **Type:** `Mock{Interface}` (e.g., `MockUserRepo`)
- **Location:** `internal/testutil/mocks/` ONLY
- **Marker:** Must contain `Code generated by MockGen`

### Pattern 3: Test Function Naming

- **Unit:** `Test{Function}_{Scenario}` (e.g., `TestCreateUser_InvalidEmail`)
- **Integration:** `TestIntegration_{Feature}_{Scenario}`
- **Table-driven:** `Test{Function}` with `t.Run()` subtests

### Pattern 4: Fixture Naming

- **Builders:** `New{Entity}Builder()` → returns builder with fluent API
- **Factories:** `Make{Entity}()` → returns entity with defaults

### Pattern 5: Error Pattern

- **Domain errors:** `Err{Domain}{Reason}` (e.g., `ErrUserNotFound`)
- **Error codes:** `ERR_{DOMAIN}_{CODE}` (e.g., `ERR_USER_NOT_FOUND`)
- **Checking:** Always use `errors.Is/errors.As`, never string compare

### Pattern 6: Test Package Policy

- **Default:** Same package (`package foo`) for unit tests - access to unexported
- **Blackbox (`package foo_test`)** ONLY for:
  - Transport/HTTP boundary tests
  - Contract tests
  - Public API surface validation

### Pattern 7: t.Parallel() Policy

- **Allowed:** `t.Parallel()` at test case level if no global state
- **Forbidden:**
  - Tests that mutate env vars / global singletons
  - Tests that share DB/container without isolation
- **Rule:** All env manipulation via `t.Setenv()` (Go 1.17+)

### Pattern 8: Assertion & Style Consistency

- **Default:** stdlib `testing` + go-cmp via `testutil/assert`
- **Helpers provided:**
  - `assert.Diff(t, want, got, opts...)`
  - `assert.ErrorIs(t, err, target)`
- **Never:** String comparison for errors

### Pattern 9: Integration Test Layout

- All integration tests MUST use `testutil/containers` helpers
- Migration flow: `containers.NewPostgres(t)` → `containers.Migrate(t, db)`
- **Isolation rule:**
  - Default: tx + rollback per test
  - Fallback: truncate/reset helper for specific tests

### Pattern 10: Deterministic Time & Waiting

- **Wait patterns:** `testutil.WaitUntil(t, timeout, interval, conditionFunc)`
- **time.Sleep:** BANNED by default
  - Exception: `//nolint:test-sleep` comment (rare, requires justification)
- **Future:** `synctest` in Sprint 3 for timer/concurrency tests

### Pattern 11: Generated Code Conventions

- All generated code must contain marker header
- Mock marker: `Code generated by MockGen`
- sqlc marker: `Code generated by sqlc`
- Detection: By marker + location, not file name pattern

### Anti-Patterns (Enforced)

| Anti-Pattern | Rule | Detection |
|--------------|------|-----------|
| `time.Sleep` in tests | ❌ Ban unless `//nolint:test-sleep` | grep + marker check |
| Hardcoded DSN | ❌ All DSN from container helper | grep for `postgresql://` |
| Scattered mocks | ❌ MockGen marker outside `testutil/mocks/` | marker + path check |
| Jumbo tests | ❌ >500 LOC or >15 cases (touched files) | wc -l + git diff |
| Placeholder tests | ❌ Empty body or `t.Skip()` without reason | AST check |
| String error compare | ❌ Must use `errors.Is/As` | grep `.Error() ==` |

### Enforcement Scripts (CI-Ready)

**1. Check Mock Location:**
```bash
#!/bin/bash
# Find MockGen markers outside testutil/mocks/
grep -r "Code generated by MockGen" --include="*.go" | \
  grep -v "internal/testutil/mocks/" && exit 1 || exit 0
```

**2. Check Integration Tag:**
```bash
#!/bin/bash
# Integration test files must have build tag
for f in $(find . -name "*_integration_test.go"); do
  grep -q "//go:build integration" "$f" || (echo "Missing tag: $f" && exit 1)
done
```

**3. Check time.Sleep:**
```bash
#!/bin/bash
# Find time.Sleep without nolint marker
for f in $(find . -name "*_test.go"); do
  if grep -q "time.Sleep(" "$f" && ! grep -q "nolint:test-sleep" "$f"; then
    echo "Banned time.Sleep in: $f" && exit 1
  fi
done
```

**4. Check Test Size (touched files only):**
```bash
#!/bin/bash
# Check only changed test files
for f in $(git diff --name-only origin/main... | grep "_test.go$"); do
  lines=$(wc -l < "$f")
  [ "$lines" -gt 500 ] && echo "Too large: $f ($lines lines)" && exit 1
done
```

### Enforcement Guidelines

**All AI Agents MUST:**
1. Follow naming patterns exactly as defined
2. Place mocks only in `internal/testutil/mocks/`
3. Use `testutil/containers` for integration tests
4. Never use `time.Sleep` without justification
5. Check patterns before PR submission

**Pattern Enforcement:**
- Pre-commit: lint scripts in Makefile
- CI: enforcement scripts run on every PR
- Review: patterns documented in architecture.md

---

## Project Structure Blueprint

### Target Folder Structure: `internal/testutil/`

```
internal/testutil/
├── assert/
│   ├── assert.go         # Diff, ErrorIs, ErrorAs helpers
│   └── options.go        # go-cmp options (ignore fields, etc.)
├── containers/
│   ├── postgres.go       # NewPostgres(t) → testcontainer
│   ├── migrate.go        # Migrate(t, db) → apply goose
│   └── helpers.go        # WaitUntil, cleanup helpers
├── fixtures/
│   ├── user_builder.go   # NewUserBuilder() fluent API
│   └── factories.go      # MakeUser(), MakeAuditEvent()
├── mocks/
│   ├── user_repo_mock.go # Generated by mockgen
│   └── audit_repo_mock.go
└── testutil.go           # RunWithGoleak, context helpers
```

### Example: Unit Test File

```go
// internal/domain/user/service_test.go
package user

import (
    "context"
    "testing"
    
    "yourmod/internal/testutil/mocks"
    "go.uber.org/mock/gomock"
)

func TestCreateUser_InvalidEmail(t *testing.T) {
    ctrl := gomock.NewController(t)
    mockRepo := mocks.NewMockUserRepo(ctrl)
    svc := NewService(mockRepo)
    
    _, err := svc.Create(context.Background(), User{Email: "invalid"})
    
    if !errors.Is(err, ErrUserInvalidEmail) {
        t.Errorf("expected ErrUserInvalidEmail, got %v", err)
    }
}
```

### Example: Integration Test File

```go
// internal/infra/postgres/user_repo_integration_test.go
//go:build integration
// +build integration

package postgres

import (
    "testing"
    
    "yourmod/internal/testutil/containers"
    "yourmod/internal/testutil/fixtures"
)

func TestMain(m *testing.M) {
    testutil.RunWithGoleak(m)
}

func TestIntegration_UserRepo_Create(t *testing.T) {
    db := containers.NewPostgres(t)
    containers.Migrate(t, db)
    
    repo := NewUserRepo(db)
    user := fixtures.NewUserBuilder().WithEmail("test@example.com").Build()
    
    created, err := repo.Create(context.Background(), user)
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }
    
    assert.Diff(t, user.Email, created.Email)
}
```

### Example: Mock Generation

```go
// internal/domain/user/ports/mocks_gen.go
package ports

//go:generate mockgen -source=user_repo.go -destination=../../testutil/mocks/user_repo_mock.go -package=mocks
```

---

## Architecture Summary

### Document Statistics

| Metric | Count |
|--------|-------|
| Total ADRs | 10 (5 research + 5 session) |
| Implementation Patterns | 11 |
| Anti-Patterns | 6 |
| Enforcement Scripts | 4 |
| Example Files | 3 |

### Key Decisions

| Category | Decision |
|----------|----------|
| Test Framework | stdlib + go-cmp + uber-go/mock |
| Integration Tests | testcontainers-go |
| Quality Gates | shuffle, race, goleak |
| Mock Location | Centralized `internal/testutil/mocks/` |
| Test Organization | Subpackages under `internal/testutil/` |
| Timeout Strategy | Configurable with defaults |

### Implementation Readiness

This architecture document is ready for downstream workflows:
- ✅ PRD alignment verified (32 FRs, 20 NFRs covered)
- ✅ Technology stack confirmed
- ✅ Patterns comprehensive enough for AI agent consistency
- ✅ Enforcement scripts defined for CI gates
- ✅ Examples provided for copy-paste implementation

### Next Steps

1. **Create Epics & Stories** - `/bmad-bmm-workflows-create-epics-and-stories`
2. **Sprint Planning** - Break stories into sprints
3. **Implementation** - Execute stories with architecture guidance

---

_Document generated: 2025-12-27_  
_Workflow: create-architecture (fast-track completed)_





