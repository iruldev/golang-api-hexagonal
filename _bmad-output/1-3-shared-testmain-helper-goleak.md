# Story 1.3: Shared TestMain Helper (goleak)

Status: done

## Story

As a **developer**,
I want automatic goroutine leak detection,
so that I catch leaks early.

## Acceptance Criteria

1. **AC1:** `go.uber.org/goleak` added to dependencies
2. **AC2:** `testutil.RunWithGoleak(m)` helper implemented (currently a stub from Story 1.1)
3. **AC3:** Integration test packages use `TestMain` with goleak
4. **AC4:** goleak runs with `IgnoreCurrent()` for known background goroutines

## Tasks / Subtasks

- [ ] Task 1: Add goleak dependency (AC: #1)
  - [ ] Run `go get go.uber.org/goleak@latest`
  - [ ] Verify import works in testutil
- [ ] Task 2: Implement RunWithGoleak helper (AC: #2)
  - [ ] Update `internal/testutil/testutil.go`
  - [ ] Replace stub with actual goleak.VerifyTestMain
  - [ ] Add IgnoreCurrent() option
- [ ] Task 3: Add TestMain to integration test packages (AC: #3)
  - [ ] Add TestMain to `internal/infra/postgres/` tests
  - [ ] Add TestMain to any other integration test files
- [ ] Task 4: Configure known goroutine ignores (AC: #4)
  - [ ] Add ignores for pgx pool goroutines
  - [ ] Add ignores for otel/observability goroutines if needed
  - [ ] Document ignore patterns

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-003:** goleak with IgnoreCurrent at package TestMain
- **Pattern 3:** TestMain convention for test packages

### RunWithGoleak Implementation

```go
// testutil/testutil.go
import "go.uber.org/goleak"

// RunWithGoleak runs tests with goroutine leak detection.
// Use this in TestMain for packages with integration tests.
func RunWithGoleak(m *testing.M) int {
    // Ignore known background goroutines
    opts := []goleak.Option{
        goleak.IgnoreCurrent(),
        // Add specific ignores for database pools, observability, etc.
    }
    return goleak.VerifyTestMain(m, opts...)
}
```

### TestMain Pattern

```go
// In integration test file (e.g., postgres/user_repo_test.go)
package postgres

import (
    "os"
    "testing"

    "github.com/iruldev/golang-api-hexagonal/internal/testutil"
)

func TestMain(m *testing.M) {
    os.Exit(testutil.RunWithGoleak(m))
}
```

### Known Background Goroutines to Ignore

| Package | Goroutine Pattern | Reason |
|---------|-------------------|--------|
| pgx | `(*Pool).backgroundHealthCheck` | Connection pool health check |
| otel | `go.opentelemetry.io/*` | Telemetry exporters |

### Testing Standards

- Run `go test ./internal/infra/postgres/...` to verify goleak works
- Should pass with no leak errors
- If leaks detected, add to ignore list or fix the leak

### Previous Story Learnings (Story 1.1, 1.2)

- `testutil.RunWithGoleak(m)` stub already exists in testutil.go
- Created in Story 1.1, now needs full implementation
- Mocks available for testing (Story 1.2)

### References

- [Source: _bmad-output/architecture.md#AD-003 goleak Integration]
- [Source: _bmad-output/epics.md#Story 1.3]
- [Source: _bmad-output/prd.md#FR12, NFR7]
- [uber-go/goleak docs](https://pkg.go.dev/go.uber.org/goleak)

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List

_Files created/modified during implementation:_
- [x] `go.mod` (add go.uber.org/goleak)
- [x] `internal/testutil/testutil.go` (implement RunWithGoleak)
- [x] `internal/infra/postgres/*_test.go` (add TestMain)

## Senior Developer Review (AI)

_Reviewer: Antigravity on 2025-12-27_

### Critical Findings (Fixed)
- **AC4 Violation:** `goleak.IgnoreCurrent()` was missing. Fixed by refactoring `RunWithGoleak` to use `goleak.VerifyTestMain` along with the ignore options.
- **Implementation Deviation:** Refactored `RunWithGoleak` to use the standard `VerifyTestMain` pattern instead of manual `Find`.

### Medium Findings (Fixed)
- **Git Tracking:** Added `internal/infra/postgres/main_test.go` to git.

### Outcome
- **Validation:** `go test ./internal/infra/postgres/...` passed.
- **Status:** Approved (Fixed).
