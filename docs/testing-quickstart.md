# Testing Quickstart Guide

> ⏱️ **Read time: ~5 minutes** | Get started testing in golang-api-hexagonal

## Directory Structure

```
internal/
├── testutil/           # Shared test helpers
│   ├── testutil.go     # TestContext, RunWithGoleak
│   └── mocks/          # Generated mocks (make mocks)
├── domain/*_test.go    # Unit tests (pure logic)
├── app/*_test.go       # Use case tests (with mocks)
└── infra/postgres/     # Integration tests (requires DB)
```

## Make Targets

| Command | Description |
|---------|-------------|
| `make test` | Run all tests with coverage + shuffle |
| `make test-unit` | Run unit tests only |
| `make test-shuffle` | Run with `-shuffle=on` |
| `make test-integration` | Run integration tests (requires DB) |
| `make coverage` | Check 80% threshold |

## Naming Conventions

- **Files:** `*_test.go` in same package
- **Functions:** `TestUnit_Scenario` or `TestUnit_Scenario_Expected`
- **Table-driven:** Use `tests` or `cases` variable

## Example: Unit Test

```go
func TestUser_Validate_EmptyEmail(t *testing.T) {
    user := domain.User{Email: ""}
    err := user.Validate()
    assert.ErrorIs(t, err, domain.ErrInvalidEmail)
}
```

## Example: Integration Test with TestMain

```go
package postgres

import (
    "testing"
    "github.com/iruldev/golang-api-hexagonal/internal/testutil"
)

func TestMain(m *testing.M) {
    testutil.RunWithGoleak(m)
}

func TestUserRepo_Create(t *testing.T) {
    ctx := testutil.TestContext(t)
    // ... test with real DB
}
```

## Further Reading

- [TestUtil Package](../internal/testutil/) - Shared helpers
- [Testing Guide](./testing-guide.md) - Detailed analysis
- [Architecture](../_bmad-output/architecture.md) - Testing patterns
