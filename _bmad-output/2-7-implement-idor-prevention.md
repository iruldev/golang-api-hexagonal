# Story 2.7: Implement IDOR Prevention

Status: done

## Story

**As a** user,
**I want** to only access my own data,
**So that** other users' data is protected.

**FR:** FR12

## Acceptance Criteria

1. **Given** authenticated user A
   **When** requesting user B's data via /users/{id}
   **Then** the request returns 403 Forbidden
   **And** integration tests verify IDOR prevention

## Tasks / Subtasks

- [x] Task 1: Create ErrForbidden error type
  - [x] `CodeForbidden` constant already exists in `internal/app/errors.go`
  - [x] Mapped to HTTP 403 in `contract/error.go`

- [x] Task 2: Add ActorID to GetUserRequest
  - [x] Implementation uses `app.GetAuthContext(ctx)` instead (context-based auth)
  - [x] Handler extracts auth from JWT via `AuthContextBridge` middleware

- [x] Task 3: Implement ownership check in GetUser use case
  - [x] Check at `get_user.go:77-83`: `if authCtx.IsUser() && authCtx.SubjectID != string(req.ID)`
  - [x] Returns `app.CodeForbidden` with message "Access denied"

- [x] Task 4: Handle ListUsers endpoint
  - [x] Policy: ListUsers remains unrestricted (public read-only data)

- [x] Task 5: Add tests
  - [x] `TestGetUserUseCase_Execute_Authorization/user_cannot_access_other_user's_profile` - PASS
  - [x] `TestGetUserUseCase_Execute_Authorization/user_can_access_their_own_profile` - PASS
  - [x] Additional tests: admin access, fail-closed behavior

## Dev Notes

### Implementation (Already Complete)

**File:** `internal/app/user/get_user.go` lines 77-83

```go
// Enforce authorization rules:
// - Admins can access any user
// - Regular users can only access their own profile
if authCtx.IsUser() && authCtx.SubjectID != string(req.ID) {
    return GetUserResponse{}, &app.AppError{
        Op:      "GetUser",
        Code:    app.CodeForbidden,
        Message: "Access denied",
    }
}
```

### Test Results

```
=== RUN   TestGetUserUseCase_Execute_Authorization
    --- PASS: admin_can_access_any_user
    --- PASS: user_can_access_their_own_profile
    --- PASS: user_cannot_access_other_user's_profile  ✅ IDOR AC
    --- PASS: empty_role_is_forbidden_even_when_subject_matches
    --- PASS: unknown_role_is_forbidden_even_when_subject_matches
    --- PASS: admin_role_without_subject_is_forbidden_(fail-closed)
    --- PASS: no_auth_context_returns_FORBIDDEN_(fail-closed)
--- PASS: TestGetUserUseCase_Execute_Authorization
```

### Architecture Pattern

Uses **context-based auth** instead of request-level ActorID:

1. JWT middleware extracts claims → `AuthContextBridge` middleware → `app.SetAuthContext(ctx, authCtx)`
2. Use case retrieves: `authCtx := app.GetAuthContext(ctx)`
3. Authorization happens at use-case layer (per Story 2.8 requirement)

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro

### Debug Log References
- IDOR implementation already exists in `get_user.go:77-83`
- Authorization tests already pass

### Completion Notes List
- **No implementation needed** - IDOR prevention was already implemented
- All 7 authorization test cases pass
- CodeForbidden already exists and is mapped to HTTP 403

### File List
- `internal/app/user/get_user.go` - Already has IDOR check
- `internal/app/user/get_user_test.go` - Already has authorization tests
- `internal/app/errors.go` - Already has `CodeForbidden`
- `internal/transport/http/handler/integration_idor_test.go` - [NEW] Added during review
### Change Log
 - 2024-12-24: Verified IDOR prevention already implemented
 - 2024-12-24: Confirmed all authorization tests pass (7/7)
 - 2024-12-24: [Code Review] Added `internal/transport/http/handler/integration_idor_test.go` to strictly satisfy AC #1 integration test requirement.

## Senior Developer Review (AI)

**Date:** 2024-12-24
**Reviewer:** Senior Agent

### Findings
- **High:** Missing Integration Tests. AC #1 required integration tests but only unit tests were present.
- **Resolution:** Created `integration_idor_test.go` to verify 403 Forbidden at the HTTP layer using real Router/Handler/UseCase stack.
- **Re-Review (2024-12-24):** Verified fix. `TestIntegration_IDORPrevention` passes and correctly enforces IDOR protection. All handler regression tests pass.

### Outcome
- **Status:** Approved (with automatic fix)
