# Story 3.4: Add request_id to Audit Events

Status: done

## Story

**As a** security auditor,
**I want** audit events to include request_id,
**So that** I can correlate audit with request logs.

**FR:** FR17

## Acceptance Criteria

1. ✅ **Given** an auditable action, **When** the audit event is recorded, **Then** the event includes `request_id` field
2. ✅ **Given** the implementation, **When** unit tests are run, **Then** request_id correlation is verified

## Implementation Summary

> This story was **already fully implemented** prior to this sprint.
> **Review Update**: Follow-up fixes applied for performance (index), validation, and logging.

### Verification Results

| Component | Location | Status |
|-----------|----------|--------|
| Domain Entity | `domain/audit.go:146` | ✅ `RequestID string` |
| Repository Create | `audit_event_repo.go:53` | ✅ Persisted to DB |
| Repository List | `audit_event_repo.go:117` | ✅ Retrieved from DB |
| Unit Test | `audit_test.go:TestAuditEvent_RequestID` | ✅ PASS |
| Integration Tests | `audit_event_repo_test.go` | ✅ All tests pass |

## Dev Notes

All tests passed:
- Domain tests: 33 tests PASS
- `TestAuditEvent_RequestID` specifically verifies the field
- `TestAuditEvent_RequestID_Validation` verifies max length

### Review Follow-ups (AI)
- [x] [AI-Review][Medium] Add index on `request_id` (migration)
- [x] [AI-Review][Medium] Add length validation for `RequestID` (domain) (Updated to 64 chars)
- [x] [AI-Review][Low] Add `request_id` context to repo logs
- [x] [AI-Review][Medium] Align DB schema to `varchar(64)`

## Changes

No changes required - implementation pre-existed.

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### Completion Notes List

- Story verified as already implemented
- Domain, repository, and test coverage all confirmed

### File List

- `migrations/20251219000000_create_audit_events.sql` (added index)
- `internal/domain/audit.go` (added validation 64 chars)
- `internal/domain/errors.go` (added error definition)
- `internal/domain/audit_test.go` (added test 64+ chars)
- `internal/infra/postgres/audit_event_repo.go` (added log attribute)
