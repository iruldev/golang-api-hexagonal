# Adoption Guide

This guide helps teams adopt the production-grade testing patterns and error handling from this boilerplate into their existing Go services.

**Estimated time:** 4-6 hours for full integration

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start (30 min)](#quick-start-30-min)
3. [Full Integration](#full-integration)
4. [Brownfield Migration](#brownfield-migration)
5. [Copy-Paste Kit](#copy-paste-kit)
6. [Adoption Checklist](#adoption-checklist)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

Before starting, ensure you have:

- [ ] Go 1.21 or later
- [ ] Docker installed and running (for testcontainers)
- [ ] PostgreSQL knowledge (for database tests)
- [ ] Familiarity with your project's test structure

---

## Quick Start (30 min)

Get up and running with the core testing utilities in 30 minutes.

### Step 1: Copy Test Utilities (10 min)

Copy the `internal/testutil/` directory to your project:

```bash
# From this boilerplate
cp -r internal/testutil/ /your-project/internal/testutil/
```

Update the import paths in all copied files to match your module.

### Step 2: Add Makefile Targets (5 min)

Add these targets to your Makefile:

```makefile
# Run unit tests with goleak
.PHONY: test
test:
	go test -v -race -count=1 ./...

# Run tests with shuffle for determinism
.PHONY: test-shuffle
test-shuffle:
	go test -v -race -count=1 -shuffle=on ./...

# Run integration tests
.PHONY: test-integration
test-integration:
	go test -v -race -count=1 -tags=integration ./...

# Run all tests
.PHONY: test-all
test-all: test test-integration
```

### Step 3: Add TestMain with goleak (10 min)

Create `internal/testutil/testutil.go`:

```go
package testutil

import (
	"os"
	"testing"

	"go.uber.org/goleak"
)

// TestMain runs all tests with goleak for goroutine leak detection.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
```

In each package with tests, create `main_test.go`:

```go
package yourpackage

import (
	"os"
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
```

### Step 4: Verify Setup (5 min)

Run your tests:

```bash
make test
```

---

## Full Integration

Complete integration takes 4-6 hours and includes all patterns.

### Phase 1: Test Utilities (1 hour)

| Component | Time | Description |
|-----------|------|-------------|
| `testutil.go` | 15 min | goleak TestMain wrapper |
| `main_test.go` per package | 45 min | Add to all test packages |

### Phase 2: Container Helpers (1.5 hours)

| Component | Time | Description |
|-----------|------|-------------|
| `containers/postgres.go` | 30 min | PostgreSQL container startup |
| `containers/migrate.go` | 30 min | Database migration helpers |
| `containers/tx.go` | 15 min | Transaction test wrappers |
| `containers/truncate.go` | 15 min | Table truncation helpers |

### Phase 3: Error Handling (1 hour)

| Component | Time | Description |
|-----------|------|-------------|
| `domain/errors/errors.go` | 30 min | DomainError type |
| `domain/errors/codes.go` | 30 min | Stable error codes |

### Phase 4: Error Mapping (30 min)

| Component | Time | Description |
|-----------|------|-------------|
| RFC 7807 response handling | 30 min | HTTP error mapping |

### Phase 5: CI Workflows (1 hour)

| Component | Time | Description |
|-----------|------|-------------|
| PR CI workflow | 30 min | Quality gates on PRs |
| Nightly CI workflow | 30 min | Race detection, integration |

---

## Brownfield Migration

For projects with existing tests, follow this incremental approach.

### Assessment (15 min)

1. **Inventory current tests:**
   ```bash
   find . -name "*_test.go" | wc -l
   ```

2. **Check for TestMain usage:**
   ```bash
   grep -r "func TestMain" --include="*_test.go" .
   ```

3. **Identify database tests:**
   ```bash
   grep -r "sql\.Open\|pgx\." --include="*_test.go" .
   ```

### Incremental Adoption Strategy

**Phase 1: New tests only (Week 1)**
- Add testutil to new test files
- Use containers for new database tests
- Don't modify existing tests yet

**Phase 2: Low-risk migration (Week 2)**
- Add TestMain with goleak to packages with few tests
- Convert database tests one package at a time
- Run full test suite after each package

**Phase 3: Full migration (Week 3+)**
- Migrate remaining packages
- Update CI workflows
- Remove old test infrastructure

### Migration Steps for Existing Tests

1. **Add goleak TestMain:**
   ```go
   // Create main_test.go in each package
   func TestMain(m *testing.M) {
       goleak.VerifyTestMain(m)
   }
   ```

2. **Convert database tests:**
   ```go
   // Before (global database)
   func TestUserRepo(t *testing.T) {
       db := globalTestDB
       // ... test logic
   }

   // After (container-based)
   func TestUserRepo(t *testing.T) {
       pool := containers.NewPostgres(t)
       containers.Migrate(t, pool, "../../migrations")
       // ... test logic
   }
   ```

3. **Add error types:**
   ```go
   // Before (string errors)
   return errors.New("user not found")

   // After (domain errors)
   return domainerrors.ErrUserNotFound
   ```

### Rollback Strategy

If issues arise during migration:

1. **Immediate rollback:** Revert the last commit
2. **Partial rollback:** Remove TestMain from problematic packages
3. **Investigate:** Check for goroutine leaks using `-v` flag

---

## Copy-Paste Kit

Ready-to-use code snippets for each component.

### TestMain with goleak

```go
// internal/testutil/testutil.go
package testutil

import (
	"os"
	"testing"

	"go.uber.org/goleak"
)

// TestMain provides a package-level test main with goleak.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
```

### PostgreSQL Container

```go
// internal/testutil/containers/postgres.go
package containers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// NewPostgres creates a PostgreSQL container for testing.
func NewPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate postgres: %v", err)
		}
	})

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}
```

### Database Migration Helper

```go
// internal/testutil/containers/migrate.go
package containers

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Migrate runs database migrations.
func Migrate(t *testing.T, pool *pgxpool.Pool, migrationsPath string) {
	t.Helper()

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("failed to set dialect: %v", err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
}
```

### Transaction Test Wrapper

```go
// internal/testutil/containers/tx.go
package containers

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx executes fn within a transaction that is rolled back after the test.
func WithTx(t *testing.T, pool *pgxpool.Pool, fn func(tx pgx.Tx)) {
	t.Helper()
	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	fn(tx)
}
```

### Table Truncation Helper

```go
// internal/testutil/containers/truncate.go
package containers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Truncate truncates the specified tables with CASCADE.
func Truncate(t testing.TB, pool *pgxpool.Pool, tables ...string) {
	t.Helper()
	ctx := context.Background()

	if len(tables) == 0 {
		return
	}

	quotedTables := make([]string, len(tables))
	for i, table := range tables {
		quotedTables[i] = fmt.Sprintf("%q", table)
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE", strings.Join(quotedTables, ", "))
	if _, err := pool.Exec(ctx, query); err != nil {
		t.Fatalf("failed to truncate tables %v: %v", tables, err)
	}
}
```

### Domain Error Types

```go
// internal/domain/errors/errors.go
package errors

import "errors"

// ErrorCode is a stable error code type.
type ErrorCode string

// DomainError represents a domain-level error with stable code.
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// New creates a new DomainError.
func New(code ErrorCode, message string) error {
	return &DomainError{Code: code, Message: message}
}

// Wrap creates a new DomainError wrapping an existing error.
func Wrap(code ErrorCode, message string, err error) error {
	return &DomainError{Code: code, Message: message, Err: err}
}
```

### Makefile Snippets

```makefile
# Testing targets
.PHONY: test test-shuffle test-integration test-all gencheck

test:
	go test -v -race -count=1 ./...

test-shuffle:
	go test -v -race -count=1 -shuffle=on ./...

test-integration:
	go test -v -race -count=1 -tags=integration ./...

test-all: test test-integration

# Generation check
gencheck:
	go generate ./...
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Generated files are out of date"; \
		git diff; \
		exit 1; \
	fi
```

### GitHub Actions CI Workflow

```yaml
# .github/workflows/ci.yml
name: CI

on:
  pull_request:
    branches: [main]
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Unit Tests
        run: make test-shuffle

      - name: Generation Check
        run: make gencheck

  integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Integration Tests
        run: make test-integration
```

---

## Adoption Checklist

Use this checklist to track your adoption progress.

### Core Setup
- [ ] Go 1.21+ installed
- [ ] Docker running
- [ ] `go.uber.org/goleak` added to go.mod
- [ ] `github.com/testcontainers/testcontainers-go` added

### Test Utilities
- [ ] `internal/testutil/testutil.go` created
- [ ] TestMain with goleak in all test packages
- [ ] Tests pass with `go test -race ./...`

### Container Helpers
- [ ] `internal/testutil/containers/` directory created
- [ ] `postgres.go` - NewPostgres helper
- [ ] `migrate.go` - Migrate helper
- [ ] `tx.go` - WithTx helper
- [ ] Integration tests use containers

### Error Handling
- [ ] `internal/domain/errors/` created
- [ ] DomainError type implemented
- [ ] Stable error codes defined
- [ ] HTTP error mapping implemented

### Makefile
- [ ] `test` target added
- [ ] `test-shuffle` target added
- [ ] `test-integration` target added
- [ ] `gencheck` target added

### CI Workflows
- [ ] `.github/workflows/ci.yml` created
- [ ] PR checks enabled
- [ ] Race detection enabled
- [ ] Integration tests in CI

---

## Troubleshooting

### "goleak: leaked goroutine" Error

```
goleak: Leaked goroutine: goroutine 42 [running]:
```

**Solution:** Check for unclosed resources:
- Database connections not closed
- HTTP clients without timeout
- Channels not drained

### "Container failed to start"

```
failed to start postgres: Cannot connect to Docker daemon
```

**Solution:**
1. Ensure Docker is running
2. Check Docker permissions
3. For macOS: Restart Docker Desktop

### "Migration failed"

```
failed to migrate: SQLSTATE 42P01: relation "users" does not exist
```

**Solution:**
1. Check migration path
2. Verify migration files exist
3. Check migration file naming (e.g., `001_create_users.sql`)

### Tests Pass Locally, Fail in CI

**Common causes:**
1. **Race conditions:** Add `-race` flag locally
2. **Time-dependent tests:** Use channel synchronization
3. **Container startup:** Increase wait timeouts

---

## Next Steps

After completing adoption:

1. **Run full test suite:** `make test-all`
2. **Check CI:** Verify workflows pass
3. **Train team:** Share this guide
4. **Iterate:** Add more patterns as needed

For questions or issues, refer to:
- [Testing Guide](testing-guide.md)
- [Testing Quickstart](testing-quickstart.md)
- [Architecture](architecture.md)
