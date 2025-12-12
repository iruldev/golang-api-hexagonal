# Story 7.7: Create Note Integration Test

Status: done

## Story

As a developer,
I want an example integration test,
So that I can understand E2E testing patterns.

## Acceptance Criteria

### AC1: Integration test implemented
**Given** `internal/interface/http/note/handler_integration_test.go` exists
**When** I run `make test`
**Then** integration test hits real HTTP endpoints
**And** test uses test database
**And** cleanup happens after test

---

## Tasks / Subtasks

- [x] **Task 1: Create test setup** (AC: #1)
  - [x] Setup test server with router
  - [x] Configure testRepository (in-memory)
  - [x] Create httptest.NewServer

- [x] **Task 2: Implement integration tests** (AC: #1)
  - [x] Test Create note (POST)
  - [x] Test Get note (GET /:id)
  - [x] Test List notes (GET)
  - [x] Test Update note (PUT /:id)
  - [x] Test Delete note (DELETE /:id)
  - [x] Verify deleted (cleanup)

- [x] **Task 3: Verify implementation** (AC: #1)
  - [x] Run `go test -tags=integration` - all pass
  - [x] Run `golangci-lint` - 0 issues

---

## Dev Notes

### Integration Test Pattern

```go
// internal/interface/http/note/handler_integration_test.go
func TestNoteHandler_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    // Setup test server
    router := chi.NewRouter()
    // ... wire up dependencies

    srv := httptest.NewServer(router)
    defer srv.Close()

    // Test cases
    t.Run("create note", func(t *testing.T) {
        resp, err := http.Post(srv.URL+"/api/v1/notes", ...)
        // assertions
    })
}
```

### File List

Files to create:
- `internal/interface/http/note/handler_integration_test.go`
