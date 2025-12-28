# Story 2.7: Context Cancellation Propagation Test

Status: done

## Story

As a **developer**,
I want tested context cancellation,
so that I trust cancellation propagates correctly.

## Acceptance Criteria

1. **AC1:** Test creates request with cancellable context
2. **AC2:** Test cancels context mid-request
3. **AC3:** Test verifies downstream operations are cancelled
4. **AC4:** Test verifies no goroutine leaks after cancel

## Tasks / Subtasks

- [x] Task 1: Create context cancellation test file (AC: #1, #2)
  - [x] Create `internal/transport/http/context_cancel_test.go`
  - [x] Create request with cancellable context
  - [x] Cancel context mid-request
  - [x] Verify request is cancelled
- [x] Task 2: Verify downstream cancellation (AC: #3)
  - [x] Create test that passes context to service layer
  - [x] Verify service layer receives cancellation
  - [x] Verify DB operations cancelled via context
- [x] Task 3: Verify no goroutine leaks (AC: #4)
  - [x] Use goleak to verify no leaks after cancel
  - [x] Ensure all goroutines clean up properly

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-008:** Context propagation
- **FR-30:** Context cancellation
- **FR-24:** Goroutine leak prevention

### Test Implementation Approach

```go
//go:build integration

func TestContextCancellation(t *testing.T) {
    // Create a handler that respects context
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        select {
        case <-ctx.Done():
            // Context cancelled - clean up
            return
        case <-time.After(5 * time.Second):
            w.WriteHeader(http.StatusOK)
        }
    })

    srv := &http.Server{Handler: handler}
    ln, err := net.Listen("tcp", "127.0.0.1:0")
    require.NoError(t, err)

    go srv.Serve(ln)
    defer srv.Shutdown(context.Background())

    // Create request with cancellable context
    ctx, cancel := context.WithCancel(context.Background())
    req, err := http.NewRequestWithContext(ctx, "GET", 
        fmt.Sprintf("http://%s/", ln.Addr()), nil)
    require.NoError(t, err)

    // Start request in goroutine
    errCh := make(chan error, 1)
    go func() {
        _, err := http.DefaultClient.Do(req)
        errCh <- err
    }()

    // Wait briefly then cancel
    time.Sleep(100 * time.Millisecond)
    cancel()

    // Verify request was cancelled
    select {
    case err := <-errCh:
        assert.ErrorIs(t, err, context.Canceled)
    case <-time.After(2 * time.Second):
        t.Fatal("request did not cancel in time")
    }
}
```

### Downstream Cancellation Test

```go
func TestContextPropagationToService(t *testing.T) {
    pool := containers.NewPostgres(t)
    containers.MigrateWithPath(t, pool, "../../migrations")

    // Create cancellable context
    ctx, cancel := context.WithCancel(context.Background())

    // Start a slow query
    errCh := make(chan error, 1)
    go func() {
        _, err := pool.Exec(ctx, "SELECT pg_sleep(10)")
        errCh <- err
    }()

    // Cancel mid-query
    time.Sleep(100 * time.Millisecond)
    cancel()

    // Verify query was cancelled
    select {
    case err := <-errCh:
        assert.Error(t, err, "query should be cancelled")
    case <-time.After(2 * time.Second):
        t.Fatal("query did not cancel in time")
    }
}
```

### Goroutine Leak Verification

Use `goleak.VerifyNone(t)` at end of test:

```go
func TestNoGoroutineLeaksOnCancel(t *testing.T) {
    defer goleak.VerifyNone(t)

    // ... cancel test ...
}
```

### Testing Standards

- Use `-tags=integration` for integration tests
- Use testcontainers for DB tests
- Use goleak for goroutine verification

### Previous Story Learnings (Story 2.6)

- Use channels for deterministic synchronization (no time.Sleep waits)
- Use MigrateWithPath for relative paths from test directory
- Document known issues for future work

### References

- [Source: _bmad-output/architecture.md#AD-008 Context Propagation]
- [Source: _bmad-output/epics.md#Story 2.7]
- [Source: _bmad-output/prd.md#FR30, FR24]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

- Implemented context cancellation tests in `internal/transport/http/context_cancel_test.go`
- Implemented DB propagation tests in `internal/transport/http/context_db_test.go`
- Used `goleak` to verify no goroutine leaks
- Refactored `context_cancel_test.go` to use channel synchronization instead of `time.Sleep` for determinism

### File List

_Files created/modified during implementation:_
- [x] `internal/transport/http/context_cancel_test.go` (new)
- [x] `internal/transport/http/context_db_test.go` (new)
