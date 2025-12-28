# Story 3.5: Doc Template for New Test Files

Status: done

## Story

As a **developer**,
I want test file templates,
so that I write tests consistently.

## Acceptance Criteria

1. **AC1:** `docs/templates/unit_test.go.tmpl` example
2. **AC2:** `docs/templates/integration_test.go.tmpl` example
3. **AC3:** Templates include proper imports, TestMain, naming

## Tasks / Subtasks

- [x] Task 1: Create unit test template (AC: #1, #3)
  - [x] Create `docs/templates/unit_test.go.tmpl`
  - [x] Include proper package declaration
  - [x] Include standard test imports
  - [x] Include TestMain with goleak
  - [x] Include example table-driven test
- [x] Task 2: Create integration test template (AC: #2, #3)
  - [x] Create `docs/templates/integration_test.go.tmpl`
  - [x] Include `//go:build integration` tag
  - [x] Include testcontainers imports
  - [x] Include TestMain with goleak + cleanup
  - [x] Include example database test
- [x] Task 3: Add template usage documentation
  - [x] Document template usage in testing-guide.md or adoption-guide.md
  - [x] Include copy instructions

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **FR-28:** Test file templates

### Template Requirements

| Element | Unit Test | Integration Test |
|---------|-----------|------------------|
| Build tag | No | `//go:build integration` |
| Package | Same as source | Same as source |
| Imports | testify, goleak | + testcontainers |
| TestMain | goleak.VerifyTestMain | + t.Cleanup for containers |
| Example | Table-driven | Database test |

### Unit Test Template Structure

```go
//go:build !integration

package {{.Package}}

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/goleak"
)

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}

func Test{{.FunctionName}}(t *testing.T) {
    tests := []struct {
        name    string
        input   any
        want    any
        wantErr bool
    }{
        {
            name:  "success case",
            input: /* ... */,
            want:  /* ... */,
        },
        {
            name:    "error case",
            input:   /* ... */,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            // Act
            // Assert
        })
    }
}
```

### Integration Test Template Structure

```go
//go:build integration

package {{.Package}}

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "go.uber.org/goleak"

    "your-module/internal/testutil/containers"
)

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}

func Test{{.FunctionName}}_Integration(t *testing.T) {
    // Setup container
    pool := containers.NewPostgres(t)
    containers.Migrate(t, pool, "../../migrations")

    tests := []struct {
        name    string
        setup   func()
        wantErr bool
    }{
        {
            name:  "success with database",
            setup: func() { /* seed data */ },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Truncate between tests
            containers.Truncate(t, pool, "users")

            if tt.setup != nil {
                tt.setup()
            }

            // Arrange
            // Act
            // Assert
        })
    }
}
```

### Naming Conventions

| Type | Pattern | Example |
|------|---------|---------|
| Unit test file | `*_test.go` | `user_service_test.go` |
| Integration file | `*_test.go` + build tag | `user_repo_test.go` |
| Test function | `Test*` | `TestCreateUser` |
| Integration suffix | `*_Integration` | `TestCreateUser_Integration` |
| Table test | `tests := []struct{}` | Standard pattern |
| Subtests | `t.Run(tt.name, ...)` | Named subtests |

### Existing Patterns in Codebase

| File | Pattern Used |
|------|--------------|
| `internal/domain/errors/errors_test.go` | Table-driven |
| `internal/transport/http/contract/error_test.go` | Table-driven with httptest |
| `internal/infra/postgres/user_repo_test.go` | Integration with containers |

### References

- [Source: _bmad-output/architecture.md#Testing Patterns]
- [Source: _bmad-output/epics.md#Story 3.5]
- [Source: _bmad-output/prd.md#FR28]
- [Existing: docs/testing-guide.md]
- [Existing: docs/adoption-guide.md]

## Dev Agent Record

### Agent Model Used

### Agent Model Used

Antigravity (Adversarial Reviewer)

### Debug Log References

- Verified templates: `docs/templates/unit_test.go.tmpl`, `docs/templates/integration_test.go.tmpl`
- Verified documentation: `docs/testing-guide.md`

### Completion Notes List

- Verified all valid ACs are met.
- Validated template syntax and "go:build" tags.
- Confirmed `goleak` usage in templates.

### File List

_Files created/modified during implementation:_
- [x] `docs/templates/unit_test.go.tmpl` (new)
- [x] `docs/templates/integration_test.go.tmpl` (new)
- [x] `docs/testing-guide.md` (modified)
