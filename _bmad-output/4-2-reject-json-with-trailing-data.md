# Story 4.2: Reject JSON with Trailing Data

Status: done

## Story

**As a** API consumer,
**I want** trailing data in JSON to be rejected,
**So that** malformed requests are caught.

**FR:** FR20

## Acceptance Criteria

1. ✅ **Given** request body `{"name":"foo"}extra`, **When** processed, **Then** 400 Bad Request returned
2. ✅ **Given** implementation, **When** DecodeJSONStrict used, **Then** `decoder.More()` check added
3. ✅ **Given** implementation, **When** unit tests run, **Then** trailing data rejection verified

## Implementation Summary

### Task 1: Extend DecodeJSONStrict ✅
- Added `JSONDecodeErrorKindTrailingData` constant
- Added `dec.More()` check after successful decode
- Returns descriptive error for trailing data

### Task 2: Add unit tests ✅
- `TestDecodeJSONStrict_RejectsTrailingText` - `{"name":"foo"}extra` rejected
- `TestDecodeJSONStrict_RejectsTrailingJSON` - `{"name":"foo"}{"other":"json"}` rejected
- `TestDecodeJSONStrict_AllowsTrailingWhitespace` - `{"name":"foo"}   ` allowed

## Changes

| File | Change |
|------|--------|
| `internal/transport/http/contract/json.go` | MODIFIED - Added More() check |
| `internal/transport/http/contract/json_test.go` | MODIFIED - Added 3 tests |
| `internal/transport/http/contract/validation.go` | MODIFIED - Improved error mapping |
| `internal/transport/http/handler/user_test.go` | MODIFIED - Added handler integration test |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `internal/transport/http/contract/json.go` - MODIFIED
- `internal/transport/http/contract/json_test.go` - MODIFIED
- `internal/transport/http/contract/validation.go` - MODIFIED
- `internal/transport/http/handler/user_test.go` - MODIFIED

## Senior Developer Review (AI)

### Findings
- **Medium:** AC1 verification claimed but missing explicit handler-level test. (FIXED)
- **Low:** `validation.go` mapped trailing data error to generic "invalid request body". (FIXED)

### Fixes Applied
- Added `TestUserHandler_CreateUser_TrailingData` to `user_test.go`.
- Updated `validation.go` to handle `JSONDecodeErrorKindTrailingData` with specific error message.

### Outcome
Approved. All issues resolved and verified.
