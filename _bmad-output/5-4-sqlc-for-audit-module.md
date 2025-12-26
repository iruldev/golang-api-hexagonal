# Story 5.4: sqlc for Audit Module

Status: done

## Story

**As a** developer,
**I want** sqlc-generated typed queries for Audit,
**So that** audit queries are type-safe.

**FR:** FR39

## Acceptance Criteria

1. ✅ **Given** sqlc.yaml configuration, **When** `make generate` is run, **Then** Audit module queries are generated
2. ✅ **Given** generated code, **When** inspected, **Then** queries are in infra layer only
3. ✅ **Given** generated code, **When** tests run, **Then** unit tests use generated code (integration tested)

## Implementation Summary

### Task 1: Create audit queries ✅
- Created `queries/audit.sql` with 5 queries
- Ran `make generate`

### Task 2: Generated files ✅
- `sqlcgen/audit.sql.go` - NEW
- Updated `models.go` (includes AuditEvent)

### Task 3: Integrate with Audit Repo ✅
- Updated `audit_event_repo.go` to use `sqlcgen` generated code
- Refactored `Create` and `ListByEntityID` to use type-safe queries

## Queries Added

| Query | Type | Description |
|-------|------|-------------|
| CreateAuditEvent | exec | Insert audit event |
| ListAuditEventsByEntity | many | List by entity type/id |
| CountAuditEventsByEntity | one | Count by entity |
| GetAuditEventByID | one | Get single event |
| ListAuditEventsByRequestID | many | List by request_id |

## Changes

| File | Change |
|------|--------|
| `queries/audit.sql` | NEW |
| `internal/infra/postgres/sqlcgen/audit.sql.go` | GENERATED |
| `internal/infra/postgres/sqlcgen/models.go` | UPDATED |
| `internal/infra/postgres/sqlcgen/db.go` | UPDATED |
| `internal/infra/postgres/sqlcgen/users.sql.go` | UPDATED |
| `internal/infra/postgres/audit_event_repo.go` | MODIFIED |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro (Code Reviewer)

### File List

- `queries/audit.sql`
- `internal/infra/postgres/sqlcgen/audit.sql.go`
- `internal/infra/postgres/sqlcgen/models.go`
- `internal/infra/postgres/sqlcgen/db.go`
- `internal/infra/postgres/sqlcgen/users.sql.go`
- `internal/infra/postgres/audit_event_repo.go`
