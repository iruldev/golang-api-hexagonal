# Story 4.2: Implement User PostgreSQL Repository

Status: done

## Story

As a **developer**,
I want **UserRepository implementation for PostgreSQL**,
so that **I can persist and retrieve users from the database**.

## Acceptance Criteria

1. **Given** the infra layer, **When** I view `internal/infra/postgres/user_repo.go`, **Then** `UserRepo` struct implements `domain.UserRepository`

2. **Given** migration file `migrations/YYYYMMDDHHMMSS_create_users.sql`, **When** migration is applied, **Then** table `users` is created with columns:
   - `id` (uuid, primary key)
   - `email` (varchar, unique)
   - `first_name` (varchar)
   - `last_name` (varchar)
   - `created_at` (timestamptz)
   - `updated_at` (timestamptz)

3. **Given** Create is called with valid user, **When** the operation succeeds, **Then** user is inserted into `users` table **And** `domain.ID` is parsed to `uuid.UUID` at repository boundary

4. **Given** GetByID is called with non-existent ID, **When** query returns no rows, **Then** `domain.ErrUserNotFound` is returned (wrapped: `"userRepo.GetByID: %w"`)

5. **Given** Create is called with duplicate email, **When** unique constraint violation occurs, **Then** `domain.ErrEmailAlreadyExists` is returned

6. **Given** List is called, **When** query executes, **Then** results are ordered by `created_at DESC, id DESC` **And** total count is returned for pagination

*Covers: FR11 (partial)*

## Source of Truth (Important)

- The canonical requirements for this story are in `docs/epics.md` under “Story 4.2”.
- The canonical domain contracts (types + method signatures) are in:
  - `internal/domain/user.go` (User, UserRepository)
  - `internal/domain/querier.go` (Querier)
  - `internal/domain/tx.go` (TxManager)
- If any snippet in this story conflicts with Story 4.1 / domain files, follow the domain files.

## Tasks / Subtasks

- [x] Task 1: Create Users Table Migration (AC: #2)
  - [x] 1.1 Create migration file `migrations/YYYYMMDDHHMMSS_create_users.sql` (use current timestamp)
  - [x] 1.2 Define `-- +goose Up` section with `CREATE TABLE users`
  - [x] 1.3 Add unique index `uniq_users_email` on email column
  - [x] 1.4 Add index `idx_users_created_at` for sorting
  - [x] 1.5 Define `-- +goose Down` section with `DROP TABLE users`
  - [x] 1.6 Verify migration applies: `make migrate-up`

- [x] Task 2: Implement Querier Adapter (AC: #1, #3)
  - [x] 2.1 Create `internal/infra/postgres/querier.go` with `PoolQuerier` wrapper
  - [x] 2.2 Implement `Exec`, `Query`, `QueryRow` methods that return pgx results as `any` (domain stays stdlib-only)
  - [x] 2.3 Add `NewPoolQuerier(pool *pgxpool.Pool) domain.Querier` constructor (call it with `postgresPool.Pool()` from `internal/infra/postgres/pool.go`)

- [x] Task 3: Implement TxManager (AC: #1)
  - [x] 3.1 Create `internal/infra/postgres/tx_manager.go` with `TxManager` struct
  - [x] 3.2 Implement `WithTx(ctx context.Context, fn func(tx domain.Querier) error) error`
  - [x] 3.3 Handle commit on nil error, rollback on error
  - [x] 3.4 Create `TxQuerier` wrapper for transactions implementing domain.Querier

- [x] Task 4: Implement UserRepo Create Method (AC: #1, #3, #5)
  - [x] 4.1 Create `internal/infra/postgres/user_repo.go` with `UserRepo` struct
  - [x] 4.2 Add `NewUserRepo() *UserRepo` constructor
  - [x] 4.3 Implement `Create(ctx context.Context, q domain.Querier, user *domain.User) error`
  - [x] 4.4 Parse `domain.ID` to `uuid.UUID` using `github.com/google/uuid`
  - [x] 4.5 Execute INSERT with parameterized query
  - [x] 4.6 Detect unique constraint violation on email → return `domain.ErrEmailAlreadyExists`
  - [x] 4.7 Wrap all errors with `op` string: `fmt.Errorf("%s: %w", op, err)`

- [x] Task 5: Implement UserRepo GetByID Method (AC: #1, #4)
  - [x] 5.1 Implement `GetByID(ctx context.Context, q domain.Querier, id domain.ID) (*domain.User, error)`
  - [x] 5.2 Parse `domain.ID` to `uuid.UUID`
  - [x] 5.3 Execute SELECT with parameterized query
  - [x] 5.4 On `pgx.ErrNoRows` → return `domain.ErrUserNotFound` (wrapped with op)
  - [x] 5.5 Scan result and convert back to domain.User

- [x] Task 6: Implement UserRepo List Method (AC: #1, #6)
  - [x] 6.1 Implement `List(ctx context.Context, q domain.Querier, params domain.ListParams) ([]domain.User, int, error)`
  - [x] 6.2 Execute COUNT query for total count
  - [x] 6.3 Execute SELECT with ORDER BY `created_at DESC, id DESC`
  - [x] 6.4 Apply LIMIT and OFFSET from `params.Limit()` and `params.Offset()`
  - [x] 6.5 Return `(users, totalCount, nil)` tuple

- [x] Task 7: Write Integration Tests (AC: all)
  - [x] 7.1 Create `internal/infra/postgres/user_repo_test.go` with build tag `//go:build integration`
  - [x] 7.2 Reuse local Docker Compose PostgreSQL (`make infra-up`) via `DATABASE_URL` (preferred: no new deps)
  - [x] 7.3 Apply migrations using goose in test setup (against a dedicated test database)
  - [x] 7.4 Test Create success case
  - [x] 7.5 Test Create duplicate email returns ErrEmailAlreadyExists
  - [x] 7.6 Test GetByID success case
  - [x] 7.7 Test GetByID not found returns ErrUserNotFound
  - [x] 7.8 Test List with pagination returns correct count and order
  - [x] 7.9 Test transaction rollback via TxManager.WithTx

- [x] Task 8: Verify Layer Compliance
  - [x] 8.1 Run `make lint` to verify no depguard violations
  - [x] 8.2 Confirm infra layer only imports domain (not app/transport)
  - [x] 8.3 Verify all error wrapping uses `op` pattern

## Dev Notes

### ⚠️ CRITICAL: Existing Code Context

**From Story 4.1 (Complete):**
| File | Description |
|------|-------------|
| `internal/domain/user.go` | User entity with `FirstName`/`LastName`, `UserRepository` interface with `Querier` param |
| `internal/domain/querier.go` | `Querier` interface with stdlib-only types |
| `internal/domain/tx.go` | `TxManager` interface |
| `internal/domain/pagination.go` | `ListParams` with `Offset()`/`Limit()` methods |
| `internal/domain/errors.go` | `ErrUserNotFound`, `ErrEmailAlreadyExists` sentinel errors |
| `internal/infra/postgres/pool.go` | `Pool` wrapper with pgxpool.Pool |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/infra/postgres/user_repo.go` | UserRepository implementation |
| `internal/infra/postgres/querier.go` | PoolQuerier/TxQuerier adapters |
| `internal/infra/postgres/tx_manager.go` | TxManager implementation |
| `internal/infra/postgres/user_repo_test.go` | Integration tests |
| `migrations/YYYYMMDDHHMMSS_create_users.sql` | Users table migration |

### Architecture Constraints (depguard enforced)

```
✅ ALLOWED in infra layer: domain, pgx, slog, otel, uuid, external packages
❌ FORBIDDEN: app, transport
```

**Infra Layer Rules:**
- Implement domain interfaces
- Accept `Querier` parameter (works with pool or tx)
- Wrap errors with `op` string: `fmt.Errorf("%s: %w", op, err)`
- Convert `domain.ID` ↔ `uuid.UUID` at repository boundary

### Pool Wrapper Note (Existing)

This repo already has `internal/infra/postgres/pool.go` which wraps `*pgxpool.Pool`.
Use `postgresPool.Pool()` when you need the underlying pool for adapters/transactions.

### UserRepository Interface (from domain)

```go
// internal/domain/user.go
type UserRepository interface {
    Create(ctx context.Context, q Querier, user *User) error
    GetByID(ctx context.Context, q Querier, id ID) (*User, error)
    List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)
}
```

### Querier Adapter Pattern

The `domain.Querier` interface uses `any` return types to keep domain stdlib-only.
Infra layer adapters should return pgx results as `any`, and repo code can type-assert to
small local interfaces (preferred over asserting concrete pgx types everywhere).

```go
// internal/infra/postgres/querier.go
package postgres

import (
    "context"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

type rowScanner interface {
    Scan(dest ...any) error
}

type rowsScanner interface {
    Close()
    Err() error
    Next() bool
    Scan(dest ...any) error
}

// PoolQuerier wraps pgxpool.Pool to implement domain.Querier
type PoolQuerier struct {
    pool *pgxpool.Pool
}

func NewPoolQuerier(pool *pgxpool.Pool) domain.Querier {
    return &PoolQuerier{pool: pool}
}

func (q *PoolQuerier) Exec(ctx context.Context, sql string, args ...any) (any, error) {
    return q.pool.Exec(ctx, sql, args...)
}

func (q *PoolQuerier) Query(ctx context.Context, sql string, args ...any) (any, error) {
    return q.pool.Query(ctx, sql, args...)
}

func (q *PoolQuerier) QueryRow(ctx context.Context, sql string, args ...any) any {
    return q.pool.QueryRow(ctx, sql, args...)
}

// TxQuerier wraps pgx.Tx to implement domain.Querier
type TxQuerier struct {
    tx pgx.Tx
}

func NewTxQuerier(tx pgx.Tx) domain.Querier {
    return &TxQuerier{tx: tx}
}

// ... similar Exec/Query/QueryRow methods
```

### TxManager Implementation Pattern

```go
// internal/infra/postgres/tx_manager.go
package postgres

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
    
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

type TxManager struct {
    pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) domain.TxManager {
    return &TxManager{pool: pool}
}

func (m *TxManager) WithTx(ctx context.Context, fn func(tx domain.Querier) error) (err error) {
    const op = "TxManager.WithTx"
    
    tx, err := m.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("%s: begin: %w", op, err)
    }
    
    defer func() {
        if p := recover(); p != nil {
            _ = tx.Rollback(ctx)
            panic(p)
        }
        if err != nil {
            _ = tx.Rollback(ctx)
            return
        }
        if commitErr := tx.Commit(ctx); commitErr != nil {
            err = fmt.Errorf("%s: commit: %w", op, commitErr)
        }
    }()
    
    return fn(NewTxQuerier(tx))
}
```

### Error Mapping Pattern

```go
// internal/infra/postgres/user_repo.go
import (
    "errors"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgconn"
)

const (
    pgUniqueViolation = "23505" // PostgreSQL unique_violation error code
)

func (r *UserRepo) Create(ctx context.Context, q domain.Querier, user *domain.User) error {
    const op = "userRepo.Create"
    
    // Parse domain.ID to uuid.UUID
    id, err := uuid.Parse(string(user.ID))
    if err != nil {
        return fmt.Errorf("%s: parse ID: %w", op, err)
    }
    
    _, err = q.Exec(ctx, `
        INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, id, user.Email, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt)
    
    if err != nil {
        var pgErr *pgconn.PgError
        if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
            return fmt.Errorf("%s: %w", op, domain.ErrEmailAlreadyExists)
        }
        return fmt.Errorf("%s: %w", op, err)
    }
    return nil
}

func (r *UserRepo) GetByID(ctx context.Context, q domain.Querier, id domain.ID) (*domain.User, error) {
    const op = "userRepo.GetByID"
    
    uid, err := uuid.Parse(string(id))
    if err != nil {
        return nil, fmt.Errorf("%s: parse ID: %w", op, err)
    }
    
    row := q.QueryRow(ctx, `
        SELECT id, email, first_name, last_name, created_at, updated_at
        FROM users WHERE id = $1
    `, uid)
    
    // Type assert to a minimal Scan interface (keeps code resilient to adapter changes)
    pgxRow, ok := row.(rowScanner)
    if !ok {
        return nil, fmt.Errorf("%s: invalid querier type", op)
    }
    
    var user domain.User
    var dbID uuid.UUID
    err = pgxRow.Scan(&dbID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
    }
    if err != nil {
        return nil, fmt.Errorf("%s: %w", op, err)
    }
    
    user.ID = domain.ID(dbID.String())
    return &user, nil
}
```

### Database Migration Pattern

```sql
-- migrations/YYYYMMDDHHMMSS_create_users.sql
-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX uniq_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- +goose Down
DROP TABLE IF EXISTS users;
```

**Note:** NO DEFAULT on `id` column - UUID v7 generated by app layer.

### Integration Test Pattern

```go
//go:build integration

package postgres_test

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jackc/pgx/v5/pgxpool"
    
    // ... imports
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
    ctx := context.Background()

    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        t.Skip("DATABASE_URL not set (run `make infra-up` and set DATABASE_URL to a dedicated test DB)")
    }

    pool, err := pgxpool.New(ctx, databaseURL)
    require.NoError(t, err)
    
    // Run migrations
    // ... goose migrate
    
    cleanup := func() {
        pool.Close()
    }
    
    return pool, cleanup
}

func TestUserRepo_Create(t *testing.T) {
    pool, cleanup := setupTestDB(t)
    defer cleanup()
    
    repo := NewUserRepo()
    querier := NewPoolQuerier(pool)
    
    user := &domain.User{
        ID:        "019400a0-1234-7def-8888-123456789abc",
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    err := repo.Create(context.Background(), querier, user)
    assert.NoError(t, err)
    
    // Verify stored
    found, err := repo.GetByID(context.Background(), querier, user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Email, found.Email)
}
```

**Optional alternative:** If you prefer fully self-contained tests, you can use `testcontainers-go`, but that adds new dependencies and should be a deliberate choice for this repo.

### Verification Commands

```bash
# Start local PostgreSQL (Docker Compose)
make infra-up

# Create a dedicated test database (one-time)
docker exec -i golang-api-hexagonal-db psql -U postgres -d postgres -c "CREATE DATABASE golang_api_hexagonal_test;" || true

# Run migration
make migrate-up

# Run integration tests (requires Docker + DATABASE_URL pointed at a dedicated test DB)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/golang_api_hexagonal_test?sslmode=disable \
  go test -tags=integration ./internal/infra/postgres/... -v

# Safety: integration tests refuse to run on non `_test` DB by default.
# Override only if you KNOW what you're doing:
# ALLOW_NON_TEST_DATABASE=true DATABASE_URL=... go test -tags=integration ./internal/infra/postgres/... -v

# Run lint to verify layer compliance
make lint

# Run full CI
make ci
```

### References

- [Source: docs/epics.md#Story 4.2] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Repository Pattern] - Querier abstraction design
- [Source: docs/architecture.md#Transaction Handling] - TxManager pattern
- [Source: docs/project-context.md#Infra Layer] - Layer rules and allowed imports
- [Source: docs/sprint-artifacts/4-1-implement-user-domain-entity.md] - Previous story patterns

### Learnings from Story 4.1

**From Story 4.1 Completion Notes:**
- Domain layer uses `type ID string` - parse to uuid.UUID at infra boundary
- `UserRepository` interface accepts `Querier` parameter for pool/tx support
- `ListParams` has `Offset()` and `Limit()` helper methods
- Error sentinel names: `ErrUserNotFound`, `ErrEmailAlreadyExists`
- All production code must be stdlib-only in domain layer

**From Epic 3 Retrospective:**
- depguard rules configured for production vs test files
- Integration tests can reuse Docker Compose PostgreSQL (preferred); testcontainers is optional
- Use `op` string pattern for all error wrapping
- Run `make ci` before marking story complete

## Dev Agent Record

### Context Reference

<!-- Path(s) to story context XML will be added here by context workflow -->

### Agent Model Used

Gemini Antigravity

### Debug Log References

No debug issues encountered.

### Completion Notes List

- **2025-12-17**: Implemented complete UserRepository for PostgreSQL
  - Created `20251217000000_create_users.sql` migration with users table, unique email index, and created_at index
  - Implemented `PoolQuerier` and `TxQuerier` adapters in `querier.go` to bridge domain.Querier interface with pgx types
  - Implemented `TxManager` in `tx_manager.go` with proper commit/rollback/panic handling
  - Implemented `UserRepo` in `user_repo.go` with Create, GetByID, and List methods
  - All methods use `op` string pattern for error wrapping
  - `domain.ID` ↔ `uuid.UUID` conversion happens at repository boundary
  - Unique constraint violations on email return `domain.ErrEmailAlreadyExists`
  - Not found queries return `domain.ErrUserNotFound`
  - List orders by `created_at DESC, id DESC` and returns total count for pagination
- **Tests**: All 8 integration tests pass (Create success/duplicate, GetByID success/not found, List pagination/ordering, TxManager rollback/commit)
- **Lint**: 0 issues, depguard layer boundaries enforced
- **Dependencies added**: `github.com/pressly/goose/v3` for integration test migrations
- **2025-12-17 (Review fixes)**:
  - Tightened unique violation mapping to only treat `uniq_users_email` as `ErrEmailAlreadyExists`
  - Hardened integration tests with DB safety guard and robust migrations path + dialect
  - Added tie-breaker ordering test for `created_at DESC, id DESC`
- **2025-12-17 (Low review tweaks)**:
  - Dokumentasi integration test guard + override env diperjelas
  - Added note about `List()` count-query tradeoff for future scalability

### File List

| Action | Path |
|--------|------|
| NEW | `migrations/20251217000000_create_users.sql` |
| NEW | `internal/infra/postgres/querier.go` |
| NEW | `internal/infra/postgres/tx_manager.go` |
| NEW | `internal/infra/postgres/user_repo.go` |
| NEW | `internal/infra/postgres/user_repo_test.go` |
| MODIFIED | `docs/sprint-artifacts/sprint-status.yaml` |
| MODIFIED | `docs/sprint-artifacts/4-2-implement-user-postgresql-repository.md` |
| MODIFIED | `go.mod` |
| MODIFIED | `go.sum` |

### Change Log

- **2025-12-17**: Story 4.2 implementation complete - UserRepo with Create/GetByID/List, Querier adapters, TxManager, and full integration test suite
- **2025-12-17**: Senior code review performed - fixed error mapping, hardened integration tests, and synced sprint status

## Senior Developer Review (AI)

Reviewer: Chat  
Date: 2025-12-17

### Summary

- Outcome: Changes Requested (High/Medium ditemukan)
- High issues:
  - Unique violation mapping terlalu generik (berisiko salah map untuk constraint selain email)
  - Integration test berpotensi destruktif jika `DATABASE_URL` bukan DB test
- Medium issues:
  - Goose dialect tidak diset eksplisit
  - `migrationsDir` rapuh (relative path)
  - Test pagination pakai `time.Sleep` (lambat/flaky)
  - Tie-breaker `id DESC` belum dites saat `created_at` sama

### Fixes Applied

- Repo: mapping `unique_violation (23505)` hanya dipetakan ke `ErrEmailAlreadyExists` bila constraint = `uniq_users_email`.
- Tests:
  - Tambah guard DB aman: hanya jalan bila nama DB berakhiran `_test` (kecuali override `ALLOW_NON_TEST_DATABASE=true`).
  - Set `goose` dialect ke `postgres`.
  - Hitung `migrationsDir` via lokasi file test (runtime) agar robust.
  - Hapus `time.Sleep` dari test pagination.
  - Tambah test tie-breaker `created_at DESC, id DESC` saat `created_at` sama.

## Senior Developer Review (AI) - Rerun (Final)

Reviewer: Chat  
Date: 2025-12-17

### Summary

- Outcome: Approved (semua High/Medium dibereskan)
- Verification:
  - `go test ./...` PASS (2025-12-17)
  - `go test -tags=integration ./internal/infra/postgres -run TestUserRepo_List_OrderByIDDescWhenCreatedAtEqual` PASS (2025-12-17)

### Low Improvements (Applied)

- Dokumentasi integrasi test diperjelas: default guard hanya menerima DB suffix `_test` + cara override `ALLOW_NON_TEST_DATABASE=true`.
- Catatan performa ditambahkan: `List()` memakai `COUNT(*)` + page query (cukup untuk fase awal; pertimbangkan keyset pagination bila tabel besar).

### Notes

- Working tree saat review bisa berisi file `??` (untracked) karena belum `git add`. Setelah verifikasi, pastikan menambahkan file baru agar history jelas.
