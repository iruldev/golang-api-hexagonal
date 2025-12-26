Status: done

## Story

**As a** API consumer,
**I want** unknown JSON fields to be rejected,
**So that** typos in requests are caught early.

**FR:** FR19

## Acceptance Criteria

1. ✅ **Given** POST /users with unknown field `usernmae`, **When** request processed, **Then** 400 Bad Request returned
2. ✅ **Given** error response, **When** RFC7807 format used, **Then** error indicates unknown field name
3. ✅ **Given** implementation, **When** unit tests run, **Then** strict decoding is verified

## Implementation Summary

### Task 1: Create strict JSON decoder helper ✅
- Created `contract/json.go` with `DecodeJSONStrict()`
- Uses `decoder.DisallowUnknownFields()`
- Returns `*JSONDecodeError` with classified error info

### Task 2: Update validation.go ✅
- Replaced `json.NewDecoder` with `DecodeJSONStrict`
- Returns field-specific error for unknown fields

### Task 3: Update user.go handler ✅
- Uses `contract.DecodeJSONStrict` for strict parsing
- Returns 400 with field name for unknown fields

### Task 4: RFC7807 error (AC: #2) ✅
- Unknown fields return `ValidationError` with field name
- Error message: `unknown field`

### Task 5: Unit tests (AC: #3) ✅
- `TestDecodeJSONStrict_ValidJSON` - valid input
- `TestDecodeJSONStrict_RejectsUnknownField` - unknown field rejection
- `TestDecodeJSONStrict_RejectsMultipleUnknownFields` - first unknown detected
- `TestDecodeJSONStrict_InvalidSyntax` - syntax errors
- `TestDecodeJSONStrict_TypeMismatch` - type mismatch handling
- `TestDecodeJSONStrict_EmptyBody` - EOF handling

## Changes

| File | Change |
|------|--------|
| `internal/transport/http/contract/json.go` | NEW - Strict JSON decoder |
| `internal/transport/http/contract/json_test.go` | NEW - 6 unit tests |
| `internal/transport/http/contract/validation.go` | MODIFIED - Use DecodeJSONStrict |
| `internal/transport/http/handler/user.go` | MODIFIED - Use DecodeJSONStrict |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/transport/http/contract/json.go` - NEW
- `internal/transport/http/contract/json_test.go` - NEW
- `internal/transport/http/contract/validation.go` - MODIFIED
- `internal/transport/http/handler/user.go` - MODIFIED

## Senior Developer Review (AI)

### Findings
- **High:** AC1 ("Given POST /users with unknown field... Then 400 Bad Request") was claimed but `user_test.go` did not test the handler response for unknown fields. (FIXED)
- **Medium:** `json.go` and `json_test.go` were untracked. (FIXED)
- **Low:** Code duplication in `user.go` manual error handling. (FIXED)

### Fixes Applied
- Refactored `CreateUser` in `internal/transport/http/handler/user.go` to use `contract.ValidateRequestBody`, simplifying code and removing duplication.
- Added `TestUserHandler_CreateUser_UnknownField` to `internal/transport/http/handler/user_test.go` to verify unknown field rejection at the handler level.
- Added untracked files to git.

### Outcome
Approved. All issues resolved and verified with tests.
