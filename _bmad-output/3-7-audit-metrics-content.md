# Story 3.7: Audit /metrics Content

Status: done

## Story

**As a** security engineer,
**I want** /metrics content audited for sensitive data,
**So that** no PII or secrets leak via metrics.

**FR:** FR46

## Acceptance Criteria

1. ✅ **Given** /metrics endpoint, **When** metrics are scraped, **Then** no labels contain user IDs, emails, or secrets
2. ✅ **Given** the audit, **When** review is complete, **Then** audit checklist exists
3. ✅ **Given** the implementation, **When** integration tests run, **Then** test scrapes and validates

## Implementation Summary

### Task 1: Audit Prometheus metrics ✅

**Metrics Audited:**
| Metric | Labels | Status |
|--------|--------|--------|
| `http_requests_total` | method, route, status | ✅ SAFE |
| `http_request_duration_seconds` | method, route | ✅ SAFE |
| `http_response_size_bytes` | method, route | ✅ SAFE |
| `go_*`, `process_*` | N/A | ✅ SAFE |

**Safeguards:**
- Route labels use patterns (`/users/{id}`) not actual IDs
- Methods are whitelisted (non-standard → `"OTHER"`)
- Unmatched routes use `"unmatched"` fallback

### Task 2: Create audit checklist ✅
- Created `docs/metrics-audit-checklist.md`
- Documents allowed/forbidden label patterns
- Includes verification commands

### Task 3: Add integration test ✅
- Created `internal/transport/http/handler/metrics_audit_test.go`
- Tests: No UUIDs, no emails, no JWTs, route placeholders, sanitized methods
- 9 test cases all PASS

## Changes

| File | Change |
|------|--------|
| `docs/metrics-audit-checklist.md` | NEW - Audit checklist |
| `internal/transport/http/handler/metrics_audit_test.go` | NEW - Integration tests |

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `docs/metrics-audit-checklist.md` - NEW
- `internal/transport/http/handler/metrics_audit_test.go` - NEW

## Senior Developer Review (AI)

- **Date:** 2025-12-25
- **Outcome:** Approved (with automatic fixes)

### Findings
1. **Medium Severity:** The integration test `metrics_audit_test.go` was manually injecting safe data instead of using the real `Metrics` middleware. This meant the test wasn't actually validating the protection logic (middleware stack) itself.
   - **Fix:** Refactored `metrics_audit_test.go` to construct a real `chi.Router` with `middleware.Metrics`, firing actual HTTP requests with PII/sensitive data to prove the middleware effectively sanitizes them.

### Conclusion
The delivered work meets all acceptance criteria. The audit checklist is comprehensive, and the refactored integration test now provides genuine security assurance by validating the full request pipeline.
