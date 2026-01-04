# ADR-004: RFC 7807 Error Handling

**Status:** Accepted
**Date:** 2026-01-04

## Context

The API needed a consistent error response format for:

- **Developer experience**: Predictable structure for client implementations
- **Debugging**: Correlation IDs for tracing errors across systems
- **Machine readability**: Programmatic error handling by API consumers
- **Internationalization**: Type URIs enable localized error messages

Previous error responses were inconsistent:
- Different fields across endpoints
- Missing correlation IDs
- No standard for validation errors
- HTTP status codes without detailed context

## Decision

We implement **RFC 7807 Problem Details** for all API error responses using `github.com/moogar0880/problems`.

**Response Structure:**

```json
{
  "type": "https://api.example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "Request validation failed",
  "code": "VAL-001",
  "request_id": "req-123abc",
  "trace_id": "abc123def456",
  "errors": [
    {
      "field": "email",
      "message": "invalid email format",
      "code": "VAL-010"
    }
  ]
}
```

**Implementation:**

```go
type Problem struct {
    *problems.DefaultProblem
    Code             string           `json:"code,omitempty"`
    RequestID        string           `json:"request_id,omitempty"`
    TraceID          string           `json:"trace_id,omitempty"`
    Errors           []FieldError     `json:"errors,omitempty"`
    ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}
```

**Error Taxonomy:**

| Prefix | Category | HTTP Status |
|--------|----------|-------------|
| `AUTH-` | Authentication | 401 |
| `AUTHZ-` | Authorization | 403 |
| `VAL-` | Validation | 400 |
| `USR-` | User domain | 400/404 |
| `DB-` | Database | 500 |
| `SYS-` | System | 500 |
| `RATE-` | Rate limiting | 429 |
| `RES-` | Resilience | 503 |

**Response Headers:**
- `Content-Type: application/problem+json`

**Correlation:**
- `request_id`: From `X-Request-ID` header or generated
- `trace_id`: From OpenTelemetry trace context

## Consequences

### Positive

- **Consistency**: All errors follow same structure
- **Debugging**: Correlation IDs link logs to responses
- **Client handling**: Predictable field names for error parsing
- **Standards compliance**: Industry-standard format (IETF RFC)

### Negative

- **Migration effort**: Existing error responses needed refactoring
- **Response size**: More fields increase payload size
- **Complexity**: Multiple error types to maintain

### Neutral

- Requires error registry maintenance
- Type URIs need hosting/documentation

## Related ADRs

- [ADR-001: Hexagonal Architecture](./ADR-001-hexagonal-architecture.md) - Problem Details implemented in transport layer
- [ADR-005: Idempotency Key Implementation](./ADR-005-idempotency-key-implementation.md) - Returns RFC 7807 for idempotency conflicts
