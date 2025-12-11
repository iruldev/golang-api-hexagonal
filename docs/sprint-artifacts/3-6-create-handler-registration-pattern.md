# Story 3.6: Create Handler Registration Pattern

Status: done

## Story

As a developer,
I want documented patterns for adding endpoints,
So that I can add new handlers consistently.

## Acceptance Criteria

### AC1: Route registered under /api/v1/
**Given** `internal/interface/http/routes.go` exists
**When** I follow the pattern to add a new handler
**Then** the route is registered under `/api/v1/`

### AC2: Middleware chain applied automatically
**Given** a new handler is registered
**When** the handler is called
**Then** middleware chain is applied automatically (Recovery, RequestID, Otel, Logging)

---

## Tasks / Subtasks

- [x] **Task 1: Create routes.go file structure** (AC: #1)
  - [x] Create `internal/interface/http/routes.go`
  - [x] Define `RegisterRoutes(r chi.Router)` function
  - [x] Document the pattern with clear examples

- [x] **Task 2: Refactor router.go to use routes.go** (AC: #1, #2)
  - [x] Move health handler registration to routes.go
  - [x] Call `RegisterRoutes` from `NewRouter`
  - [x] Verify middleware chain still applies

- [x] **Task 3: Create example handler for pattern** (AC: #1, #2)
  - [x] Create `internal/interface/http/handlers/example.go`
  - [x] Implement simple GET `/api/v1/example` handler
  - [x] Show how to use request context (trace ID, etc.)

- [x] **Task 4: Document handler creation pattern** (AC: #1)
  - [x] Add doc comments in routes.go explaining the pattern
  - [x] Include step-by-step instructions for adding new handlers
  - [x] Reference middleware that will be applied

- [x] **Task 5: Create routes tests** (AC: #1, #2)
  - [x] Create `internal/interface/http/routes_test.go`
  - [x] Test: routes are registered under /api/v1
  - [x] Test: middleware chain is applied to handlers

- [x] **Task 6: Verify implementation** (AC: #1, #2)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### Routes.go Pattern

```go
// internal/interface/http/routes.go
package http

import (
    "github.com/go-chi/chi/v5"
    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/handlers"
)

// RegisterRoutes registers all API routes under the /api/v1 prefix.
// 
// Adding a New Handler:
// 1. Create handler function in internal/interface/http/handlers/
// 2. Add route registration here using r.Method("/path", handlers.YourHandler)
// 3. Middleware chain (Recovery, RequestID, Otel, Logging) is applied automatically
//
// Example:
//   r.Get("/users", handlers.ListUsers)
//   r.Post("/users", handlers.CreateUser)
//   r.Get("/users/{id}", handlers.GetUser)
func RegisterRoutes(r chi.Router) {
    // Health check (Story 3.1)
    r.Get("/health", handlers.HealthHandler)
    
    // Example handler (Story 3.6)
    r.Get("/example", handlers.ExampleHandler)
    
    // Add new routes below this line
    // r.Get("/your-endpoint", handlers.YourHandler)
}
```

### Example Handler Pattern

```go
// internal/interface/http/handlers/example.go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/iruldev/golang-api-hexagonal/internal/interface/http/middleware"
    "github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// ExampleHandler demonstrates the handler pattern with context usage.
// Shows how to access request ID and create child spans.
func ExampleHandler(w http.ResponseWriter, r *http.Request) {
    // Access request ID from middleware
    requestID := middleware.GetRequestID(r.Context())
    
    // Access trace ID from OTEL middleware
    traceID := observability.GetTraceID(r.Context())
    
    // Create child span for tracing
    ctx, span := observability.StartSpan(r.Context(), "example-operation")
    defer span.End()
    
    // Use ctx for downstream operations
    _ = ctx
    
    response := map[string]string{
        "message":    "Example handler working",
        "request_id": requestID,
        "trace_id":   traceID,
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}
```

### Router Integration

```go
// internal/interface/http/router.go (updated)
func NewRouter(cfg *config.Config) chi.Router {
    // ... (existing logger and tracer init)
    
    r := chi.NewRouter()

    // Global middleware (order matters!)
    r.Use(middleware.Recovery(logger))
    r.Use(middleware.RequestID)
    r.Use(middleware.Otel("api"))
    r.Use(middleware.Logging(logger))

    // API v1 routes - all handlers get middleware automatically
    r.Route("/api/v1", func(r chi.Router) {
        RegisterRoutes(r)  // Delegate to routes.go
    })

    return r
}
```

### Architecture Compliance

**Layer:** `internal/interface/http`
**Pattern:** Separation of routing from handler implementation
**Benefit:** Centralized route management, consistent middleware application

### Dependencies

**Existing:**
- `github.com/go-chi/chi/v5`
- All existing middleware (Recovery, RequestID, Otel, Logging)

### References

- [Source: docs/epics.md#Story-3.6]
- [Story 3.1 - Chi Router](file:///docs/sprint-artifacts/3-1-setup-chi-router-with-versioned-api.md)
- [Story 3.5 - OTEL](file:///docs/sprint-artifacts/3-5-add-opentelemetry-trace-propagation.md)

---

## Dev Agent Record

### Context Reference

Story context created by create-story workflow.
Sixth story in Epic 3: HTTP API Core.
Establishes handler registration pattern for consistent endpoint creation.

### Agent Model Used

To be filled by dev agent.

### Debug Log References

None yet.

### Completion Notes List

- Story created: 2025-12-11

### File List

Files to create:
- `internal/interface/http/routes.go` - Route registration
- `internal/interface/http/routes_test.go` - Routes tests
- `internal/interface/http/handlers/example.go` - Example handler

Files to modify:
- `internal/interface/http/router.go` - Use RegisterRoutes
- `docs/sprint-artifacts/sprint-status.yaml` - Story status tracking
