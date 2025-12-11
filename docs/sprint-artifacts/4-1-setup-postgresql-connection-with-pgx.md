# Story 4.1: Setup PostgreSQL Connection with pgx

Status: done

## Story

As a developer,
I want the system to connect to PostgreSQL with connection pooling,
So that database connections are efficiently managed.

## Acceptance Criteria

### AC1: Connection pool established with config
**Given** valid `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
**When** the application starts
**Then** connection pool is established
**And** `DB_MAX_OPEN_CONNS`, `DB_MAX_IDLE_CONNS` are respected

---

## Tasks / Subtasks

- [x] **Task 1: Add database config** (AC: #1)
  - [x] Add database fields to `internal/config/config.go`
  - [x] Add DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
  - [x] Add DB_MAX_OPEN_CONNS (default: 25)
  - [x] Add DB_MAX_IDLE_CONNS (default: 5)
  - [x] Add DB_CONN_MAX_LIFETIME (default: 5m)

- [x] **Task 2: Add pgx dependency** (AC: #1)
  - [x] Run `go get github.com/jackc/pgx/v5`
  - [x] Run `go get github.com/jackc/pgx/v5/pgxpool`
  - [x] Update go.mod and go.sum

- [x] **Task 3: Create database connection** (AC: #1)
  - [x] Create `internal/infra/postgres/postgres.go`
  - [x] Implement `NewPool(cfg *config.Config) (*pgxpool.Pool, error)`
  - [x] Configure connection pool from config
  - [x] Implement `Close()` for graceful shutdown

- [x] **Task 4: Update .env.example** (AC: #1)
  - [x] Add database environment variables
  - [x] Document each variable

- [x] **Task 5: Integrate with main.go** (AC: #1)
  - [x] Initialize database pool on startup
  - [x] Handle connection errors
  - [x] Add graceful shutdown of pool

- [x] **Task 6: Create connection tests** (AC: #1)
  - [x] Create `internal/infra/postgres/postgres_test.go`
  - [x] Test: Pool creation with valid config
  - [x] Test: Pool respects max connections
  - [x] Test: Pool close works correctly

- [x] **Task 7: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Config Extension

```go
// internal/config/config.go
type Config struct {
    // ... existing fields
    DB DBConfig `mapstructure:",squash"`
}

type DBConfig struct {
    Host            string        `mapstructure:"DB_HOST" validate:"required"`
    Port            int           `mapstructure:"DB_PORT" validate:"required,min=1,max=65535"`
    User            string        `mapstructure:"DB_USER" validate:"required"`
    Password        string        `mapstructure:"DB_PASSWORD" validate:"required"`
    Name            string        `mapstructure:"DB_NAME" validate:"required"`
    MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS" default:"25"`
    MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS" default:"5"`
    ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME" default:"5m"`
}

func (c *DBConfig) DSN() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        c.User, c.Password, c.Host, c.Port, c.Name)
}
```

### PostgreSQL Pool

```go
// internal/infra/postgres/postgres.go
package postgres

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/iruldev/golang-api-hexagonal/internal/config"
)

// NewPool creates a new PostgreSQL connection pool.
func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
    poolConfig, err := pgxpool.ParseConfig(cfg.DB.DSN())
    if err != nil {
        return nil, fmt.Errorf("parse pool config: %w", err)
    }

    poolConfig.MaxConns = int32(cfg.DB.MaxOpenConns)
    poolConfig.MinConns = int32(cfg.DB.MaxIdleConns)
    poolConfig.MaxConnLifetime = cfg.DB.ConnMaxLifetime

    pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    // Test connection
    if err := pool.Ping(ctx); err != nil {
        pool.Close()
        return nil, fmt.Errorf("ping database: %w", err)
    }

    return pool, nil
}
```

### Main Integration

```go
// cmd/server/main.go
func main() {
    // ... existing code

    // Initialize database
    pool, err := postgres.NewPool(context.Background(), cfg)
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }
    defer pool.Close()

    logger.Info("Database connected",
        zap.String("host", cfg.DB.Host),
        zap.Int("max_conns", cfg.DB.MaxOpenConns),
    )

    // ... rest of server setup
}
```

### Environment Variables

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hexagonal
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

### Architecture Compliance

**Layer:** `internal/infra/postgres`
**Pattern:** Infrastructure adapter for database
**Benefit:** Database driver can be swapped without affecting domain

### References

- [Source: docs/epics.md#Story-4.1]
- [pgx documentation](https://github.com/jackc/pgx)
- [Epic 3 - HTTP API Core](file:///docs/sprint-artifacts/epic-3-retrospective.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
First story in Epic 4: Database & Persistence.
Establishes PostgreSQL connection infrastructure.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/infra/postgres/postgres.go` - Connection pool
- `internal/infra/postgres/postgres_test.go` - Pool tests

Files to modify:
- `internal/config/config.go` - Add DBConfig struct
- `.env.example` - Add database variables
- `cmd/server/main.go` - Initialize database pool
- `go.mod` - Add pgx dependency
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
