# Story 8.5: Migrate Integration Tests to Testcontainers

Status: done

## Story

As a developer,
I want integration tests to use testcontainers,
So that tests are self-contained and CI-friendly.

## Acceptance Criteria

### AC1: Testcontainers Package Exists
**Given** `internal/testing/containers.go` exists
**When** I review the code
**Then** PostgreSQL container helper is available
**And** Redis container helper is available
**And** containers are properly terminated after use

### AC2: Integration Tests Use Containers
**Given** integration test runs
**When** I run `go test -tags=integration ./...`
**Then** PostgreSQL container starts automatically
**And** Redis container starts if needed
**And** containers are cleaned up after tests

---

## Tasks / Subtasks

- [x] **Task 1: Add testcontainers dependency** (AC: #1)
  - [x] Run `go get github.com/testcontainers/testcontainers-go`
  - [x] Run `go get github.com/testcontainers/testcontainers-go/modules/postgres`
  - [x] Run `go get github.com/testcontainers/testcontainers-go/modules/redis`

- [x] **Task 2: Create containers.go helper** (AC: #1)
  - [x] Create `internal/testing/containers.go`
  - [x] Implement `NewPostgresContainer()` returning DSN and cleanup func
  - [x] Implement `NewRedisContainer()` returning address and cleanup func
  - [x] Use `Container.Terminate()` for cleanup

- [x] **Task 3: Create test fixtures** (AC: #2)
  - [x] Create `internal/testing/fixtures.go`
  - [x] Implement test database setup/teardown with `SetupTestDatabase()`
  - [x] Implement migrations runner for test database

- [x] **Task 4: Create integration tests** (AC: #2)
  - [x] Create `internal/testing/containers_integration_test.go`
  - [x] Test PostgreSQL container starts and queries work
  - [x] Test Redis container starts and address is valid
  - [x] Verify notes table is created via migrations

- [x] **Task 5: Add Makefile target** (AC: #2)
  - [x] Add `make test-integration` command
  - [x] Add to .PHONY and help

---

## Dev Notes

### Architecture Placement

```
internal/
└── testing/
    ├── containers.go       # Testcontainers helpers
    ├── fixtures.go         # Test data fixtures
    └── testmain_test.go    # TestMain for package-level setup
```

---

### Testcontainers PostgreSQL Pattern

```go
// internal/testing/containers.go
package testing

import (
    "context"
    "fmt"
    "time"

    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
    "github.com/testcontainers/testcontainers-go/wait"
)

type PostgresContainer struct {
    Container testcontainers.Container
    DSN       string
}

func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15-alpine"),
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(60*time.Second),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("start postgres container: %w", err)
    }

    dsn, err := container.ConnectionString(ctx, "sslmode=disable")
    if err != nil {
        return nil, fmt.Errorf("get connection string: %w", err)
    }

    return &PostgresContainer{
        Container: container,
        DSN:       dsn,
    }, nil
}

func (c *PostgresContainer) Terminate(ctx context.Context) error {
    return c.Container.Terminate(ctx)
}
```

---

### Testcontainers Redis Pattern

```go
func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
    container, err := redis.RunContainer(ctx,
        testcontainers.WithImage("redis:7-alpine"),
        testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections")),
    )
    if err != nil {
        return nil, fmt.Errorf("start redis container: %w", err)
    }

    endpoint, err := container.Endpoint(ctx, "")
    if err != nil {
        return nil, fmt.Errorf("get endpoint: %w", err)
    }

    return &RedisContainer{
        Container: container,
        Addr:      endpoint,
    }, nil
}
```

---

### Integration Test Pattern with TestMain

```go
// internal/interface/http/note/handler_integration_test.go
//go:build integration

package note

import (
    "context"
    "os"
    "testing"

    testinghelpers "github.com/iruldev/golang-api-hexagonal/internal/testing"
)

var testDSN string

func TestMain(m *testing.M) {
    ctx := context.Background()

    // Start PostgreSQL container
    pgContainer, err := testinghelpers.NewPostgresContainer(ctx)
    if err != nil {
        panic(err)
    }
    testDSN = pgContainer.DSN

    // Run tests
    code := m.Run()

    // Cleanup
    _ = pgContainer.Terminate(ctx)

    os.Exit(code)
}
```

---

### Previous Story Learnings

From Story 8-4 Job Observability:
- Use QueueDefault constant for queue labels
- Document limitations in comments
- Tests should use actual implementation, not mocks

---

### Existing Integration Test

Current `handler_integration_test.go`:
- Uses `//go:build integration` tag
- Uses in-memory `testRepository` mock
- Tests CRUD operations end-to-end
- 241 lines, well-structured

**Migration needed:** Replace `testRepository` with real PostgreSQL repository.

---

### File List

**Create:**
- `internal/testing/containers.go`
- `internal/testing/fixtures.go`

**Modify:**
- `go.mod` - Add testcontainers dependencies
- `internal/interface/http/note/handler_integration_test.go` - Use real containers
- `Makefile` - Add test-integration target

---

## Dev Agent Record

### Agent Model Used
{{agent_model_name_version}}

### Completion Notes
- Testcontainers requires Docker to be running
- Container startup adds ~5-10 seconds per container
- Consider container reuse for faster test runs
