# Story 2.2: Container Helpers Package

Status: done

## Story

As a **developer**,
I want golden-path container helpers,
so that all integration tests use consistent patterns.

## Acceptance Criteria

1. **AC1:** `containers.Migrate(t, pool)` applies goose migrations
2. **AC2:** `containers.WithTx(t, pool, fn)` provides tx+rollback isolation
3. **AC3:** `containers.Truncate(t, pool, tables...)` for truncate fallback
4. **AC4:** Helpers documented in testutil/containers/README.md

## Tasks / Subtasks

- [x] Task 1: Implement Migrate helper (AC: #1)
  - [x] Create `containers.Migrate(t, pool)` function
  - [x] Uses goose to apply migrations from `migrations/` folder
  - [x] Fatals on migration error
- [x] Task 2: Implement WithTx helper (AC: #2)
  - [x] Create `containers.WithTx(t, pool, fn)` function
  - [x] Starts transaction, calls fn, then rollbacks
  - [x] Provides test isolation without truncate overhead
- [x] Task 3: Implement Truncate helper (AC: #3)
  - [x] Create `containers.Truncate(t, pool, tables...)` function
  - [x] Truncates specified tables with CASCADE
  - [x] Falls back when transaction isolation isn't suitable
- [x] Task 4: Create README documentation (AC: #4)
  - [x] Create `internal/testutil/containers/README.md`
  - [x] Document all helpers with examples
  - [x] Include usage patterns and recommendations

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-004:** testcontainers-go helpers
- **Pattern 5:** Test isolation strategies

### Migrate Implementation

```go
// Migrate applies goose migrations to the database.
// Uses migrations from the project's migrations/ directory.
func Migrate(t testing.TB, pool *pgxpool.Pool) {
    t.Helper()
    
    db, err := sql.Open("pgx", pool.Config().ConnString())
    if err != nil {
        t.Fatalf("failed to open sql.DB: %v", err)
    }
    defer db.Close()
    
    if err := goose.Up(db, "migrations"); err != nil {
        t.Fatalf("goose up failed: %v", err)
    }
}
```

### WithTx Implementation

```go
// WithTx runs the function within a transaction that is rolled back after.
// Provides test isolation without truncate overhead.
func WithTx(t testing.TB, pool *pgxpool.Pool, fn func(tx pgx.Tx)) {
    t.Helper()
    ctx := context.Background()
    
    tx, err := pool.Begin(ctx)
    if err != nil {
        t.Fatalf("failed to begin tx: %v", err)
    }
    
    defer func() {
        if err := tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
            t.Errorf("failed to rollback: %v", err)
        }
    }()
    
    fn(tx)
}
```

### Truncate Implementation

```go
// Truncate truncates the specified tables with CASCADE.
// Use when transaction isolation isn't suitable.
func Truncate(t testing.TB, pool *pgxpool.Pool, tables ...string) {
    t.Helper()
    ctx := context.Background()
    
    for _, table := range tables {
        // Use quoted identifier to prevent SQL injection
        query := fmt.Sprintf("TRUNCATE TABLE %q CASCADE", table)
        if _, err := pool.Exec(ctx, query); err != nil {
            t.Fatalf("failed to truncate %s: %v", table, err)
        }
    }
}
```

### Testing Standards

- Run `go test ./internal/testutil/containers/...` to verify
- Tests require Docker for container startup
- Use `-short` flag to skip integration tests

### Previous Story Learnings (Story 2.1)

- NewPostgres implemented with t.Cleanup()
- Use t.Helper() in all test helpers
- Fatal on unrecoverable errors, Error on cleanup failures

### References

- [Source: _bmad-output/architecture.md#AD-004 testcontainers-go]
- [Source: _bmad-output/epics.md#Story 2.2]
- [Source: _bmad-output/prd.md#FR17, FR18, FR19]
- [goose migrations](https://github.com/pressly/goose)

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

- Implemented `Migrate`, `WithTx`, and `Truncate` helpers.
- `Truncate` optimized to use single SQL statement.
- Documented usage in `internal/testutil/containers/README.md`.

### File List

_Files created/modified during implementation:_
- [x] `internal/testutil/containers/migrate.go` (new)
- [x] `internal/testutil/containers/tx.go` (new)
- [x] `internal/testutil/containers/truncate.go` (new)
- [x] `internal/testutil/containers/README.md` (new)
- [x] `internal/testutil/containers/containers.go` (modified)
