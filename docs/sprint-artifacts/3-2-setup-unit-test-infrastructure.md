# Story 3.2: Setup Unit Test Infrastructure with Coverage Gate

Status: done

## Story

As a **developer**,
I want **to run unit tests with race detection and coverage enforcement**,
So that **I can verify code correctness and maintain quality standards**.

## Acceptance Criteria

1. **Given** I have test files in `internal/domain/` and `internal/app/`
   **When** I run `make test`
   **Then** tests execute with `-race` flag enabled
   **And** test output shows pass/fail status for each test
   **And** coverage profile is generated (`coverage.out` at repo root)

2. **Given** I want to check coverage threshold
   **When** I run `make coverage`
   **Then** coverage is calculated for `./internal/domain/...` and `./internal/app/...`
   **And** coverage percentage is displayed
   **And** `make coverage` fails (exit non-zero) if combined coverage < 80%
   **And** `make coverage` passes (exit 0) if combined coverage ≥ 80%

## Tasks / Subtasks

- [x] Task 1: Ensure `make test` runs with race + coverage profile (AC: #1)
  - [x] Update `Makefile` `test` target to include `-race -coverprofile=coverage.out -covermode=atomic`
  - [x] Keep `-v` so output shows pass/fail status per test
  - [x] Verify `coverage.out` is created at repo root after `make test`

- [x] Task 2: Create sample unit tests for domain layer (AC: #1, #2)
  - [x] Create `internal/domain/id.go` with a minimal `type ID string`
  - [x] Create `internal/domain/errors.go` with sentinel errors used by tests
  - [x] Create `internal/domain/user.go` with `User` entity + `Validate() error`
  - [x] Create `internal/domain/user_test.go` with table-driven unit tests (testify assertions)
  - [x] Verify domain tests compile and pass

- [x] Task 3: Create sample unit tests for app layer (AC: #1, #2)
  - [x] Create `internal/app/user/create_user.go` use case stub
  - [x] Create `internal/app/user/create_user_test.go` with unit tests
  - [x] Mock repository dependencies using interfaces
  - [x] Verify app tests compile and pass

- [x] Task 4: Implement `make coverage` target (AC: #2)
  - [x] Add `coverage` target to Makefile
  - [x] Run tests with `-coverprofile=coverage.out` for domain + app packages
  - [x] Parse coverage output using `go tool cover -func`
  - [x] Extract total coverage percentage
  - [x] Fail with non-zero exit if coverage < 80%
  - [x] Pass with exit 0 if coverage ≥ 80%
  - [x] Display coverage percentage in human-readable format

- [x] Task 5: Verify complete test infrastructure works (AC: #1, #2)
  - [x] Run `make test` - should pass with race detection
  - [x] Run `make coverage` - should calculate and enforce threshold
  - [x] Verify existing tests (12 files) still pass
  - [x] Test edge case: intentionally drop coverage and verify failure

- [x] Task 6: Update Makefile help text if needed (N/A)
  - [x] Ensure `make help` shows `coverage` target with description

## Dev Notes

### Architecture Context

This story focuses on **Epic 3: Local Quality Gates** which ensures developers can verify code quality locally before pushing. The coverage gate specifically targets the business logic layers:
- `internal/domain/` - Domain entities, interfaces, errors (100% target, enforced at 80%)
- `internal/app/` - Use cases, application logic (90% target, enforced at 80%)

From `docs/project-context.md`:
- **Coverage requirement:** domain + app ≥ 80%
- **Test style:** Table-driven with testify assertions
- **Test location:** Co-located (`*_test.go` next to implementation)

### Current State Analysis

**Existing Test Infrastructure:**
- 12 test files already exist (infra and transport layers)
- Makefile has `test` target with `-race` flag
- testify is already a dependency
- All tests follow table-driven patterns

**Domain and App Layers:**
- Currently only contain `.keep` files (placeholder)
- Need sample implementations for testing infrastructure validation
- Sample code should be minimal but representative

### Makefile Enhancement

Current `make test`:
```makefile
test:
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
```

Add coverage generation and threshold checking (no external deps like `bc`):
```makefile
## COVERAGE_THRESHOLD: Minimum combined coverage percentage for domain+app
COVERAGE_THRESHOLD ?= 80

## coverage: Check test coverage meets 80% threshold for domain+app
.PHONY: coverage
coverage:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic \
		./internal/domain/... \
		./internal/app/...
	@echo ""
	@echo "📈 Coverage report:"
	@go tool cover -func=coverage.out | tail -1
	@THRESHOLD="$(COVERAGE_THRESHOLD)"; \
	COVERAGE=$$(go tool cover -func=coverage.out | tail -1 | awk '{gsub(/%/,"",$$3); print $$3}'); \
	if awk 'BEGIN {exit !('"$$COVERAGE"' < '"$$THRESHOLD"')}'; then \
		echo ""; \
		echo "❌ Coverage $$COVERAGE% is below $$THRESHOLD% threshold"; \
		exit 1; \
	else \
		echo ""; \
		echo "✅ Coverage $$COVERAGE% meets $$THRESHOLD% threshold"; \
	fi
```

### Sample Domain Entity (Minimal)

Create representative domain code to validate test infrastructure:

```go
package domain

import (
    "errors"
)

// internal/domain/errors.go
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrEmailExists      = errors.New("email already exists")
    ErrInvalidEmail     = errors.New("invalid email format")
)
```

```go
// internal/domain/id.go
package domain

// ID is a minimal identifier type (kept stdlib-only for domain boundary).
type ID string
```

```go
package domain

import (
	"context"
	"strings"
	"time"
)

// User represents a minimal domain entity used to validate unit testing patterns.
type User struct {
	ID        ID
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (u User) Validate() error {
	if strings.TrimSpace(u.Email) == "" {
		return ErrInvalidEmail
	}
	if strings.TrimSpace(u.Name) == "" {
		return ErrInvalidUserName
	}
	return nil
}

type UserRepository interface {
	Create(ctx context.Context, user User) (User, error)
	GetByID(ctx context.Context, id ID) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id ID) error
}
```

### Sample Tests Pattern

```go
// internal/domain/user_test.go
package domain

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestUser_Validate(t *testing.T) {
    tests := []struct {
        name    string
        user    User
        wantErr error
	}{
        {
            name:    "valid user",
            user:    User{Name: "John Doe", Email: "test@example.com"},
            wantErr: nil,
        },
        {
            name:    "missing email",
            user:    User{Name: "John Doe", Email: ""},
            wantErr: ErrInvalidEmail,
        },
        {
            name:    "missing name",
            user:    User{Name: "", Email: "test@example.com"},
            wantErr: ErrInvalidUserName,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.user.Validate()
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Previous Story Learnings (From Story 3.1)

1. **depguard rules active** - Domain tests MUST use stdlib only + testify
2. **App layer restrictions** - App tests cannot import net/http, pgx, slog, uuid
3. **Mock via interfaces** - Use interface-based mocking for repositories
4. **Shared patterns** - `internal/shared/` exists for cross-layer contracts

### Technology Specifics

- **Go version**: 1.24 (from go.mod / toolchain)
- **Test framework**: `testing` stdlib + `github.com/stretchr/testify`
- **Coverage tool**: `go tool cover` (stdlib)
- **Race detector**: `go test -race` (stdlib)

### File Locations

| Action | Path |
|--------|------|
| Update | `Makefile` (add `coverage` target) |
| Update | `.golangci.yml` (allow testify in domain/app `*_test.go`) |
| Create | `internal/domain/user.go` |
| Create | `internal/domain/user_test.go` |
| Create | `internal/domain/errors.go` |
| Create | `internal/domain/id.go` |
| Create | `internal/app/user/create_user.go` |
| Create | `internal/app/user/create_user_test.go` |
| Verify | All 12 existing test files |
| Reference | `docs/project-context.md#Testing-Patterns` |
| Reference | `docs/architecture.md#Test-Location` |

### Testing Strategy

1. **Positive test**: Run `make test` → all tests pass with race detection
2. **Coverage pass**: Run `make coverage` → should pass ≥80%
3. **Coverage fail**: Run `COVERAGE_THRESHOLD=101 make coverage` → should fail (exit non-zero)
4. **Existing tests**: Verify 12 existing test files still pass

### Depguard Considerations

From Story 3.1, depguard rules are now active:
- **domain-layer**: `$gostd` only (production code), with a separate test rule that also allows `github.com/stretchr/testify` for `*_test.go`
- **app-layer**: `$gostd` + `internal/domain` (production code), with a separate test rule that also allows `github.com/stretchr/testify` for `*_test.go`

Implementation note: update `.golangci.yml` to exclude `*_test.go` from the production rules and add `domain-layer-test` / `app-layer-test` rules to permit testify in tests while keeping the same “no logging / no HTTP / no pgx” denies.

### References

- [Source: docs/epics.md#Story-3.2]
- [Source: docs/project-context.md#Testing-Patterns]
- [Source: docs/architecture.md#Test-Location]
- [Source: docs/sprint-artifacts/3-1-configure-golangci-lint-with-boundary-rules.md#Dev-Notes]
- [Source: docs/sprint-artifacts/epic-2-retro-2025-12-17.md#Key-Learnings]

## Dev Agent Record

### Context Reference

- `docs/project-context.md` - Testing patterns and coverage requirements
- `docs/architecture.md` - File structure and test location patterns
- `docs/sprint-artifacts/3-1-configure-golangci-lint-with-boundary-rules.md` - Depguard rules affecting tests
- `docs/sprint-artifacts/epic-2-retro-2025-12-17.md` - Previous learnings and patterns

### Agent Model Used

GPT-5 (Codex CLI)

### Debug Log References

- `make test -race` on macOS may print linker warnings (`malformed LC_DYSYMTAB`) but tests still pass.

### Completion Notes List

- ✅ Task 1: Updated `make test` to run with `-v -race` and generate `coverage.out` (`-coverprofile` + `-covermode=atomic`)
- ✅ Task 2: Created domain layer entities (`id.go`, `errors.go`, `user.go`) with comprehensive unit tests (100% coverage)
- ✅ Task 3: Created app layer use case (`create_user.go`) with mock repository-based unit tests (≥ 80% coverage)
- ✅ Task 4: Added `make coverage` target that enforces combined coverage for domain+app ≥ 80%
- ✅ Task 5: Verified `make test` and `make coverage` pass; verified failure path by running `COVERAGE_THRESHOLD=101 make coverage`
- ✅ Task 6: `make help` shows `coverage` target with description
- ✅ Updated `.golangci.yml` depguard rules to allow `github.com/stretchr/testify` for domain/app `*_test.go` while keeping production rules strict
- ✅ All existing tests pass with no regressions (`go test ./...`, `golangci-lint run ./...`)

### Implementation Plan

1. Verified Makefile already has correct `test` and `coverage` targets
2. Created minimal domain entities following stdlib-only constraint
3. Created app layer use case following hexagonal architecture (imports only domain layer)
4. Used interface-based mocking for repository dependencies in tests
5. Fixed depguard rule pattern to properly exclude `*_test.go` from strict domain-layer rule

### File List

**Created:**
- `internal/domain/id.go` - Minimal ID type with helper methods
- `internal/domain/id_test.go` - Unit tests for ID type (4 tests)
- `internal/domain/errors.go` - Sentinel errors for domain
- `internal/domain/user.go` - User entity with Validate() and UserRepository interface
- `internal/domain/user_test.go` - Unit tests for User.Validate() (7 tests)
- `internal/app/user/create_user.go` - CreateUserUseCase with Execute method
- `internal/app/user/create_user_test.go` - Unit tests with mock repository (7 tests)

**Modified:**
- `Makefile` - Updated `test` and added `coverage` (supports `COVERAGE_THRESHOLD` override for manual failure-path verification)
- `.golangci.yml` - Added depguard rules to allow testify in domain/app `*_test.go` and keep strict production rules
- `.gitignore` - Ignore `coverage.out`
- `docs/sprint-artifacts/sprint-status.yaml` - Mark story as `done`
- `docs/architecture.md` - Align examples with domain stdlib-only rules (no pgx/uuid types in domain-facing examples; Go 1.24+)
- `docs/project-context.md` - Update Go version to 1.24+ (matches `go.mod`)

### Change Log

- 2025-12-17: Implemented Story 3.2 - Setup Unit Test Infrastructure with Coverage Gate
- 2025-12-17: Code review fixes (context propagation, depguard test rules, improved test determinism, coverage threshold override)
- 2025-12-17: Code review rerun (final) approved after doc alignment + verification runs

## Senior Developer Review (AI)

**Date:** 2025-12-17  
**Outcome:** Changes Requested → Fixed (HIGH+MEDIUM resolved)

### Ringkasan Temuan dan Perbaikan

- ✅ AC#1 dan AC#2 tervalidasi via eksekusi `make test` + `make coverage` (coverage.out dibuat; gate 80% berjalan).
- ✅ Perbaikan kualitas: `context.Context` sekarang dipropagasikan dari use case ke repository interface.
- ✅ Perbaikan test: mock ID generator dibuat deterministik dan jalur error repository diuji.
- ✅ depguard: aturan dipisah untuk production vs `*_test.go` agar testify legal di test tanpa melonggarkan boundary production.

---

## Senior Developer Review (AI) - Rerun

**Date:** 2025-12-17  
**Outcome:** Approved (minor notes)

### Catatan

- ✅ Rerun `make test`, `make coverage`, dan `golangci-lint run ./...` masih lulus.
- 🟢 macOS `-race` masih bisa menampilkan linker warning `malformed LC_DYSYMTAB` (noise, tidak mempengaruhi pass/fail).

---

## Senior Developer Review (AI) - Rerun (Final)

**Date:** 2025-12-17  
**Outcome:** Approved

### Checklist Eksekusi

- ✅ `make test` (race + `coverage.out`)
- ✅ `make coverage` (domain+app gate, default 80%)
- ✅ Failure-path diverifikasi: `COVERAGE_THRESHOLD=101 make coverage` (exit non-zero)
- ✅ `golangci-lint run ./...` (0 issues)

### Catatan

- 🟢 macOS race: linker warning `malformed LC_DYSYMTAB` masih mungkin muncul saat `-race` (noise saja).
