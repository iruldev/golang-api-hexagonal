# Story 1.1: Create testutil Structure

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a **developer**,
I want a well-organized `internal/testutil/` directory,
so that I know where to find and add test helpers.

## Acceptance Criteria

1. **AC1:** `internal/testutil/` exists with subpackages: `assert/`, `containers/`, `fixtures/`, `mocks/`
2. **AC2:** Each subpackage has a placeholder file with package doc
3. **AC3:** `testutil.go` contains common helpers (context, timeout)

## Tasks / Subtasks

- [x] Task 1: Create directory structure (AC: #1)
  - [x] Create `internal/testutil/`
  - [x] Create `internal/testutil/assert/`
  - [x] Create `internal/testutil/containers/`
  - [x] Create `internal/testutil/fixtures/`
  - [x] Create `internal/testutil/mocks/`
- [x] Task 2: Create placeholder files with package docs (AC: #2)
  - [x] Create `internal/testutil/assert/assert.go` with package doc
  - [x] Create `internal/testutil/containers/containers.go` with package doc
  - [x] Create `internal/testutil/fixtures/fixtures.go` with package doc
  - [x] Create `internal/testutil/mocks/doc.go` with package doc
- [x] Task 3: Create main testutil.go helper (AC: #3)
  - [x] Create `internal/testutil/testutil.go`
  - [x] Add `TestContext(t)` helper for context with timeout
  - [x] Add `RunWithGoleak(m *testing.M)` stub (to be completed in Story 1.3)

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-001:** Test Directory Structure = Structured subpackages under `internal/testutil/`
- **Pattern 1:** Test File Naming follows established conventions
- **Pattern 9:** Integration Layout requires `testutil/containers` helpers

### Project Structure Notes

```
internal/
└── testutil/
    ├── testutil.go         # Main helpers (context, timeout)
    ├── assert/
    │   └── assert.go       # go-cmp helpers (Story 1.2+ will add Diff, ErrorIs)
    ├── containers/
    │   └── containers.go   # testcontainers helpers (Story 2.1 will implement)
    ├── fixtures/
    │   └── fixtures.go     # Builders/factories (future stories)
    └── mocks/
        └── doc.go          # Generated mocks location (Story 1.2)
```

### Technical Requirements

- Go module: `github.com/iruldev/golang-api-hexagonal`
- Package naming: follow existing conventions in `internal/`
- Package docs: each file starts with `// Package <name> provides...`

**Package Doc Examples:**

```go
// assert/assert.go
// Package assert provides test assertion helpers using go-cmp.
package assert

// containers/containers.go  
// Package containers provides testcontainers helpers for integration tests.
package containers

// fixtures/fixtures.go
// Package fixtures provides test data builders and factories.
package fixtures

// mocks/doc.go
// Package mocks contains generated mock implementations for testing.
package mocks
```

**TestContext Implementation:**

```go
// testutil.go
package testutil

import (
    "context"
    "testing"
    "time"
)

// TestContext returns a context with test timeout (30s default).
// Automatically cancelled when test completes.
func TestContext(t testing.TB) context.Context {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    t.Cleanup(cancel)
    return ctx
}

// RunWithGoleak is a stub for goleak integration (Story 1.3).
// For now, just run tests normally.
func RunWithGoleak(m *testing.M) int {
    // TODO: Story 1.3 will add goleak.VerifyTestMain(m)
    return m.Run()
}
```

### Testing Standards

- No tests required for this story (structure/scaffolding only)
- Verify: `go build ./internal/testutil/...` succeeds

### References

- [Source: _bmad-output/architecture.md#Project Structure Blueprint]
- [Source: _bmad-output/architecture.md#AD-001 Test Directory Structure]
- [Source: _bmad-output/epics.md#Story 1.1]
- [Source: _bmad-output/prd.md#FR8]

## Dev Agent Record

### Agent Model Used

Antigravity (Gemini-based)

### Debug Log References

- `go build ./internal/testutil/...` - ✅ SUCCESS (no errors)

### Completion Notes List

- Created testutil structure with 4 subpackages per AD-001
- Added TestContext with 30s default + TestContextWithTimeout variant
- Added RunWithGoleak stub (Story 1.3 will implement goleak)
- All files have proper package documentation
- [Fix] Refactored testutil.go: Unexported default timeout and improved documentation to prevent drift (Review 1.1)
- [Fix] Staged missing subpackages `assert`, `containers`, `fixtures`, `mocks` (Review 1.1 Rerun)

### File List

_Files created during implementation:_
- [x] `internal/testutil/testutil.go`
- [x] `internal/testutil/assert/assert.go`
- [x] `internal/testutil/containers/containers.go`
- [x] `internal/testutil/fixtures/fixtures.go`
- [x] `internal/testutil/mocks/doc.go`

