# Story 3.3: RFC 7807 Response with trace_id

Status: done

## Story

As a **API consumer**,
I want standardized error responses,
so that I can debug issues easily.

## Acceptance Criteria

1. **AC1:** Error responses follow RFC 7807 format
2. **AC2:** Response includes: type, title, status, detail, trace_id
3. **AC3:** trace_id propagated from OpenTelemetry
4. **AC4:** Error codes included as `code` extension field

## Tasks / Subtasks

- [x] Task 1: Verify RFC 7807 compliance (AC: #1, #2)
  - [x] Review `ProblemDetail` struct fields
  - [x] Verify all required fields present
  - [x] Add any missing RFC 7807 fields if needed
- [x] Task 2: Verify trace_id propagation (AC: #3)
  - [x] Confirm `TraceID` populated from OpenTelemetry
  - [x] Test trace_id appears in error responses
  - [x] Verify `ctxutil.GetTraceID()` integration
- [x] Task 3: Verify error code inclusion (AC: #4)
  - [x] Confirm `Code` field in responses
  - [x] Test domain error codes appear correctly
  - [x] Verify code format matches `ERR_DOMAIN_CODE`

## Dev Notes

### Architecture Compliance

Per `architecture.md` ADRs and patterns:
- **FR-22:** RFC 7807 error responses
- **FR-24:** Trace ID in errors
- **NFR-19:** Observability

### Current State - ALREADY IMPLEMENTED

existing `internal/transport/http/contract/error.go`:

```go
// ProblemDetail represents an RFC 7807 Problem Details response.
type ProblemDetail struct {
    Type             string            `json:"type"`              // RFC 7807 required
    Title            string            `json:"title"`             // RFC 7807 required
    Status           int               `json:"status"`            // RFC 7807 required
    Detail           string            `json:"detail"`            // RFC 7807 optional
    Instance         string            `json:"instance"`          // RFC 7807 optional
    Code             string            `json:"code"`              // Extension field (AC4)
    RequestID        string            `json:"request_id,omitempty"`
    TraceID          string            `json:"trace_id,omitempty"` // AC3
    ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}
```

All ACs appear to be ALREADY SATISFIED by existing implementation!

### RFC 7807 Compliance Check

| Field | Required | Status |
|-------|----------|--------|
| `type` | Yes | ✅ Present |
| `title` | Yes | ✅ Present |
| `status` | Yes | ✅ Present |
| `detail` | No | ✅ Present |
| `instance` | No | ✅ Present |

### Extension Fields

| Field | Purpose | Status |
|-------|---------|--------|
| `code` | Stable error code (AC4) | ✅ Present |
| `trace_id` | OTel trace (AC3) | ✅ Present |
| `request_id` | Request correlation | ✅ Present |
| `validation_errors` | Field-level validation | ✅ Present |

### trace_id Propagation

```go
func populateIDs(ctx context.Context, problem *ProblemDetail) {
    problem.RequestID = ctxutil.GetRequestID(ctx)
    if traceID := ctxutil.GetTraceID(ctx); traceID != "" && traceID != ctxutil.EmptyTraceID {
        problem.TraceID = traceID
    }
}
```

### Story 3.3 Focus

Since implementation exists, Story 3.3 should:
1. **Verify** existing implementation meets all ACs
2. **Add tests** to prove compliance
3. **Document** RFC 7807 compliance

### Testing Verification

Tests should verify:
- Response Content-Type is `application/problem+json`
- All required RFC 7807 fields present
- trace_id populated when OTel context available
- Code field matches domain error code

### References

- [RFC 7807](https://tools.ietf.org/html/rfc7807)
- [Source: _bmad-output/architecture.md#Error Handling]
- [Source: _bmad-output/epics.md#Story 3.3]
- [Source: _bmad-output/prd.md#FR22, FR24, NFR19]
- [Existing: internal/transport/http/contract/error.go]

## Dev Agent Record

### Agent Model Used

_To be filled by dev agent_

### Debug Log References

_To be filled during implementation_

### Completion Notes List

_To be filled after implementation_

### File List

 _Files created/modified during implementation:_
 - [x] `internal/transport/http/contract/error.go` (verified)
 - [x] `internal/transport/http/contract/error_test.go` (verified)
 - [x] `internal/transport/http/ctxutil/trace.go` (verified)
 - [x] `_bmad-output/sprint-status.yaml` (sync status)

 ## Senior Developer Review (AI)

 _Reviewer: @bmad-bmm-workflows-code-review on 2025-12-28_

 ### Findings
 - **[HIGH] Fixed**: Tasks were marked [ ] but implementation was complete. Marked all tasks as done.
 - **[HIGH] Fixed**: Story file was untracked in git. Added to tracking.
 - **[MEDIUM] Noted**: Implementation files were pre-existing (committed in previous stories). Verified they fully satisfy AC1-AC4.
 - **[Success]**: Comprehensive tests in `error_test.go` cover RFC 7807 structure, error codes, and trace ID propagation.

 ### Outcome
 **Approved** with automated fixes applied.
