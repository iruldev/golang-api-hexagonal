# Story 5.2: Case-Insensitive Email with CITEXT

Status: done

## Story

**As a** user,
**I want** email uniqueness to be case-insensitive,
**So that** User@Example.com and user@example.com are the same.

**FR:** FR37

## Acceptance Criteria

1. ✅ **Given** email column uses CITEXT type, **When** registering with different case emails, **Then** uniqueness enforced case-insensitively
2. ✅ **Given** migration, **When** applied, **Then** CITEXT extension is created
3. ✅ **Given** implementation, **When** integration test runs, **Then** behavior is verified

## Implementation Summary

### Task 1: Create migration with CITEXT ✅
- Created `migrations/20251226084756_add_citext_email.sql`
- Enables citext extension
- ALTER COLUMN email TYPE CITEXT

### Task 2: Integration tests ✅
- `TestCaseInsensitiveEmail_CITEXT` - verifies case-insensitive uniqueness
- `TestCITEXTExtensionExists` - verifies extension installed

## Changes

| File | Change |
|------|--------|
| `migrations/20251226084756_add_citext_email.sql` | NEW |
| `internal/infra/postgres/citext_integration_test.go` | NEW |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `migrations/20251226084756_add_citext_email.sql` - NEW
- `internal/infra/postgres/citext_integration_test.go` - NEW

### Senior Developer Review (AI)

- [x] Verified Files: `migrations/20251226084756_add_citext_email.sql`, `internal/infra/postgres/citext_integration_test.go`
- [x] Outcome: Approved with fixes
- [x] Fixes Applied:
    - Added untracked files to git
    - Refactored `citext_integration_test.go` to use safety guards and avoid magic strings (using `23505` code check)
