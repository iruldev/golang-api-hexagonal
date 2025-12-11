# Story 3.1: Setup Chi Router with Versioned API

Status: done

## Story

As a developer,
I want versioned API endpoints under `/api/v1/`,
So that I can evolve the API without breaking clients.

## Acceptance Criteria

### AC1: Health endpoint responds with 200
**Given** the HTTP server is running
**When** I request `GET /api/v1/health`
**Then** response status is 200
**And** route is mounted under versioned prefix

---

## Tasks / Subtasks

- [x] **Task 1: Add chi dependency** (AC: #1)
  - [x] Run `go get github.com/go-chi/chi/v5` ✅ v5.2.3
  - [x] Verify chi is added to go.mod ✅

- [x] **Task 2: Create router package** (AC: #1)
  - [x] Create `internal/interface/http/router.go` ✅
  - [x] Initialize chi router with `/api/v1` prefix ✅
  - [x] Implement `NewRouter(cfg *config.Config) chi.Router` ✅

- [x] **Task 3: Create health handler** (AC: #1)
  - [x] Create `internal/interface/http/handlers/health.go` ✅
  - [x] Implement `HealthHandler` returning 200 OK ✅
  - [x] Return simple JSON: `{"status": "ok"}` ✅

- [x] **Task 4: Integrate router in main.go** (AC: #1)
  - [x] Import router package in main.go ✅
  - [x] Replace `Handler: nil` with `Handler: router` ✅
  - [x] Verify server uses the new router ✅

- [x] **Task 5: Create router tests** (AC: #1)
  - [x] Create `internal/interface/http/router_test.go` ✅
  - [x] Test `GET /api/v1/health` returns 200 ✅
  - [x] Use `httptest` for testing ✅
  - [x] Added bonus tests (method not allowed, 404, version prefix) ✅

- [x] **Task 6: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass ✅ (100% router coverage)
  - [x] Run `make lint` - 0 issues ✅

---

## Dev Notes

### chi Router Setup Pattern

```go
// internal/interface/http/router.go
package http

import (
    "github.com/go-chi/chi/v5"
    "github.com/iruldev/golang-api-hexagonal/internal/config"
    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/handlers"
)

// NewRouter creates a new chi router with versioned API routes.
func NewRouter(cfg *config.Config) chi.Router {
    r := chi.NewRouter()

    // API v1 routes
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/health", handlers.HealthHandler)
    })

    return r
}
```

### Health Handler Pattern

```go
// internal/interface/http/handlers/health.go
package handlers

import (
    "encoding/json"
    "net/http"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
    Status string `json:"status"`
}

// HealthHandler returns the health status of the service.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}
```

### main.go Integration

```go
// cmd/server/main.go changes:
import (
    httpx "github.com/iruldev/golang-api-hexagonal/internal/interface/http"
)

// Replace Handler: nil with:
server := &http.Server{
    Addr:    ":" + port,
    Handler: httpx.NewRouter(cfg),
}
```

### Architecture Compliance

**Layer:** `internal/interface/http` (allowed: chi, stdlib, internal packages)
**Pattern:** Clean Architecture - interface layer talks to domain via use cases
**Dependency Direction:** HTTP → UseCase → Domain (not implemented yet in this story)

### Testing Strategy

```go
// internal/interface/http/router_test.go
package http_test

import (
    "net/http"
    "net/http/httptest"
    "testing"

    httpx "github.com/iruldev/golang-api-hexagonal/internal/interface/http"
    "github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
    router := httpx.NewRouter(nil)
    
    req := httptest.NewRequest("GET", "/api/v1/health", nil)
    rec := httptest.NewRecorder()
    
    router.ServeHTTP(rec, req)
    
    assert.Equal(t, http.StatusOK, rec.Code)
    assert.Contains(t, rec.Body.String(), "ok")
}
```

### Previous Epic Learnings Applied

From **Epic 2 Retrospective:**
- ✅ Integrate with main.go in first story (not wait until later)
- ✅ Keep test coverage high from start
- ✅ Add new config fields to .env.example if needed

### Dependencies

**New:** `github.com/go-chi/chi/v5`
**Existing:** Uses `internal/config` for Config struct

### File Structure After Implementation

```
internal/interface/http/
├── doc.go           # Package documentation (exists)
├── router.go        # Chi router setup (NEW)
├── router_test.go   # Router tests (NEW)
├── handlers/
│   └── health.go    # Health handler (NEW)
└── httpx/           # Existing subpackage
```

### References

- [Source: docs/epics.md#Story-3.1]
- [chi documentation](https://github.com/go-chi/chi)
- [Epic 2 Retrospective - Action Items](file:///docs/sprint-artifacts/epic-2-retrospective.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
First story in Epic 3: HTTP API Core.
Establishes the foundation for all HTTP handlers.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11
- Implementation completed: 2025-12-11
  - Added chi/v5 v5.2.3 dependency
  - Created router.go with versioned API routes
  - Created handlers/health.go with HealthHandler
  - Integrated router in main.go (replaced nil Handler)
  - Created 4 router tests with 100% coverage
  - Fixed errcheck lint issue on json.Encode
  - Coverage: 100% (http package), Lint: 0 issues

### File List

Files created:
- `internal/interface/http/router.go` - Chi router setup (21 lines)
- `internal/interface/http/router_test.go` - Router tests (64 lines)
- `internal/interface/http/handlers/health.go` - Health handler (23 lines)

Files modified:
- `go.mod` - Added chi/v5 dependency
- `go.sum` - Updated with chi dependencies
- `cmd/server/main.go` - Integrated NewRouter(cfg) (55 lines)
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
