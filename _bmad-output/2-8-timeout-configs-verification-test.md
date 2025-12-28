# Story 2.8: Timeout Configs Verification Test

Status: done

## Story

As a **developer**,
I want tested timeout configs,
so that I trust timeouts are applied correctly.

## Acceptance Criteria

1. **AC1:** Test verifies HTTP server timeout is applied
2. **AC2:** Test verifies DB query timeout is applied
3. **AC3:** Test verifies graceful shutdown timeout is applied
4. **AC4:** Timeouts are configurable via env

## Tasks / Subtasks

- [x] Task 1: Create timeout verification test file (AC: #1)
  - [x] Create `internal/transport/http/timeout_test.go`
  - [x] Test HTTP ReadTimeout
  - [x] Test HTTP WriteTimeout
  - [x] Verify timeouts trigger when exceeded
- [x] Task 2: Verify DB query timeout (AC: #2)
  - [x] Test context timeout on queries
  - [x] Verify long queries are cancelled
  - [x] Use testcontainers for real DB
- [x] Task 3: Verify shutdown timeout (AC: #3)
  - [x] Test graceful shutdown respects configured timeout
  - [x] Verify timeout actually triggers
  - [x] Build on Story 2.6 shutdown tests
- [x] Task 4: Test env configuration (AC: #4)
  - [x] Verify timeouts can be set via env vars
  - [x] Test default values when env not set
  - [x] Document expected env var names

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **AD-009:** Timeout configuration
- **FR-31:** Configurable timeouts

### HTTP Server Timeout Test

```go
// ... (omitted for brevity, implementation exists in code)
```

### DB Query Timeout Test

```go
// ... (omitted for brevity, implementation exists in code)
```

### Environment Configuration Test

```go
// ... (omitted for brevity, implementation exists in code)
```

### Expected Env Variables

| Env Var | Default | Description |
|---------|---------|-------------|
| `HTTP_READ_TIMEOUT` | 5s | HTTP server read timeout |
| `HTTP_WRITE_TIMEOUT` | 30s | HTTP server write timeout |
| `SHUTDOWN_TIMEOUT` | 30s | Graceful shutdown timeout |
| `DB_QUERY_TIMEOUT` | 5s | Default DB query timeout |

### Testing Standards

- Use `-tags=integration` for integration tests
- Use testcontainers for DB tests
- Use t.Setenv() for env var tests
- Build on patterns from Stories 2.6-2.7

### Previous Story Learnings (Story 2.6, 2.7)

- Use channels for deterministic synchronization
- Use waitForActiveQuery for DB test coordination
- Use goleak for goroutine leak detection
- Use t.Cleanup() for goleak verification with testcontainers

### References

- [Source: _bmad-output/architecture.md#AD-009 Timeout Configuration]
- [Source: _bmad-output/epics.md#Story 2.8]
- [Source: _bmad-output/prd.md#FR31]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

- Implemented standard timeout tests
- Added AC4 env verification
- Added DBQueryTimeout to config
- [Review Fix] Fixed race condition in shutdown tests
- [Review Fix] Added configuration validation

### File List

_Files created/modified during implementation:_
- [x] `internal/transport/http/timeout_test.go` (new)
- [x] `internal/transport/http/timeout_db_test.go` (new)
- [x] `internal/infra/config/config.go` (modified)

## Senior Developer Review (AI)

- [x] Story file loaded from `{{story_path}}`
- [x] Story Status verified as reviewable (review)
- [x] Epic and Story IDs resolved (2.8)
- [x] Acceptance Criteria cross-checked against implementation
- [x] File List reviewed and validated for completeness
- [x] Tests identified and mapped to ACs
- [x] Code quality review performed on changed files
- [x] Outcome decided (Approve)
- [x] Status updated according to settings
- [x] Sprint status synced
- [x] Story saved successfully

_Reviewer: BMad on 2025-12-28_
