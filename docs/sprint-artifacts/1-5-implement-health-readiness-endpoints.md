# Story 1.5: Implement Health & Readiness Endpoints

Status: done

## Story

**As a** developer,
**I want** health and readiness endpoints,
**So that** I can verify the service is running and ready to accept traffic.

## Acceptance Criteria

1. **Given** the service is running
   **When** I call `GET /health` (liveness probe)
   **Then** I receive HTTP 200
   **And** response body is `{"data":{"status":"ok"}}`
   **And** this endpoint does NOT check database connectivity

2. **Given** the service is running and database is connected
   **When** I call `GET /ready` (readiness probe)
   **Then** I receive HTTP 200
   **And** response body is `{"data":{"status":"ready","checks":{"database":"ok"}}}`

3. **Given** the service is running but database is NOT connected
   **When** I call `GET /ready`
   **Then** I receive HTTP 503
   **And** response body is `{"data":{"status":"not_ready","checks":{"database":"failed"}}}`

## Tasks / Subtasks

- [x] Task 1: Setup HTTP server and router (AC: #1, #2, #3)
  - [x] Add chi router dependency (v5.2.3)
  - [x] Create HTTP server in `internal/transport/http/`
  - [x] Implement graceful shutdown with SIGTERM handling

- [x] Task 2: Implement /health endpoint (AC: #1)
  - [x] Create `internal/transport/http/handler/health.go`
  - [x] Implement `GET /health` returning `{"data":{"status":"ok"}}`
  - [x] Ensure no database check is performed

- [x] Task 3: Implement database connection (AC: #2, #3)
  - [x] Add pgx pool dependency (v5.7.6)
  - [x] Create `internal/infra/postgres/pool.go` with connection pool
  - [x] Implement Ping() method for health checks

- [x] Task 4: Implement /ready endpoint (AC: #2, #3)
  - [x] Create `internal/transport/http/handler/ready.go`
  - [x] Implement database connectivity check using Ping()
  - [x] Return HTTP 200 with `{"data":{"status":"ready","checks":{"database":"ok"}}}` on success
  - [x] Return HTTP 503 with `{"data":{"status":"not_ready","checks":{"database":"failed"}}}` on failure

- [x] Task 5: Wire up main.go (AC: #1, #2, #3)
  - [x] Initialize database pool with config.DatabaseURL
  - [x] Create router with /health and /ready endpoints
  - [x] Start HTTP server on config.Port

- [x] Task 6: Write tests (AC: #1, #2, #3)
  - [x] Unit tests for health handler
  - [x] Unit tests for ready handler with mock database
  - [x] Integration test verifying endpoints (router-level with DB ok/fail)

## Dev Notes

### Response Format [Source: docs/project-context.md]

All responses use the envelope format:
```json
{"data": {...}}
```

### Router Setup Pattern

```go
// internal/transport/http/router.go
package http

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

func NewRouter(healthHandler, readyHandler http.Handler) chi.Router {
    r := chi.NewRouter()
    r.Use(middleware.Recoverer)
    r.Use(middleware.RequestID)
    
    r.Get("/health", healthHandler.ServeHTTP)
    r.Get("/ready", readyHandler.ServeHTTP)
    
    return r
}
```

### Health Handler Pattern

```go
// internal/transport/http/handler/health.go
package handler

import (
    "encoding/json"
    "net/http"
)

type HealthResponse struct {
    Data struct {
        Status string `json:"status"`
    } `json:"data"`
}

func NewHealthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        resp := HealthResponse{}
        resp.Data.Status = "ok"
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(resp)
    }
}
```

### Database Pool Pattern [Source: docs/architecture.md]

```go
// internal/infra/postgres/pool.go
package postgres

import (
    "context"
    "fmt"
    
    "github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
    pool *pgxpool.Pool
}

func NewPool(ctx context.Context, databaseURL string) (*Pool, error) {
    pool, err := pgxpool.New(ctx, databaseURL)
    if err != nil {
        return nil, fmt.Errorf("postgres.NewPool: %w", err)
    }
    return &Pool{pool: pool}, nil
}

func (p *Pool) Ping(ctx context.Context) error {
    return p.pool.Ping(ctx)
}

func (p *Pool) Close() {
    p.pool.Close()
}
```

### Previous Story Learnings [Source: Story 1.1-1.4]

- Configuration loads DATABASE_URL from environment
- Docker Compose provides PostgreSQL at localhost:5432
- Use `kelseyhightower/envconfig` for config
- Migrations create schema_info table

## Technical Requirements

- **Router:** github.com/go-chi/chi/v5 v5.2.3
- **Database driver:** github.com/jackc/pgx/v5 v5.7.6
- **HTTP port:** From config.Port (default 8080)

## Project Context Reference

Full project context available at: [docs/project-context.md](../project-context.md)

Critical rules:
- /health is a LIVENESS probe - no external dependencies
- /ready is a READINESS probe - checks database connectivity
- Use envelope format for responses: `{"data": {...}}`

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-16)

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

- `go mod tidy` - SUCCESS (added chi v5.2.3, pgx v5.7.6)
- `go build ./...` - SUCCESS
- `go test -v ./...` - SUCCESS (14 tests pass)

### Completion Notes List

- [x] chi router added and server running with graceful shutdown
- [x] /health returns 200 with correct format (tested)
- [x] /ready returns 200 when DB connected (tested)
- [x] /ready returns 503 when DB disconnected (tested)
- [x] All 14 tests pass

### File List

Files created/modified:
- `internal/transport/http/router.go` (NEW)
- `internal/transport/http/handler/health.go` (NEW)
- `internal/transport/http/handler/health_test.go` (NEW)
- `internal/transport/http/handler/ready.go` (NEW)
- `internal/transport/http/handler/ready_test.go` (NEW)
- `internal/transport/http/handler/integration_test.go` (NEW)
- `internal/infra/postgres/pool.go` (NEW)
- `cmd/api/main.go` (MODIFIED)
- `go.mod` (MODIFIED - added chi, pgx)
- `go.sum` (MODIFIED)
- `docs/sprint-artifacts/sprint-status.yaml` (MODIFIED - status sync)
- Removed `.keep` placeholders in `internal/infra/postgres/` and `internal/transport/http/handler/`
- This story file (updated by review)

### Change Log

- 2025-12-16: Story 1.5 implemented - HTTP server with chi router, /health and /ready endpoints, pgx database pool, graceful shutdown, and unit tests
- 2025-12-16: Review fixes - startup allows service to run without DB (ready reports not_ready), file list updated, integration test still pending
