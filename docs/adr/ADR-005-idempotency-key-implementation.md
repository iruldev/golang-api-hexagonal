# ADR-005: Idempotency Key Implementation

**Status:** Accepted
**Date:** 2026-01-04

## Context

POST/PUT/PATCH requests can cause duplicate side effects when:

- **Network failures**: Client doesn't receive response, retries
- **Timeouts**: Request succeeds but client times out
- **User behavior**: Double-clicking submit buttons
- **Retry mechanisms**: Client-side retry logic

Without idempotency protection:
- Duplicate orders created
- Double payments processed
- Multiple emails sent

The API needed a mechanism for clients to safely retry mutating requests.

## Decision

We implement **Idempotency-Key header** support following Stripe's pattern:

**Client Request:**
```http
POST /api/v1/users
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
Content-Type: application/json

{"email": "user@example.com", "first_name": "John"}
```

**Middleware Behavior:**

1. **First request**: 
   - Execute handler
   - Store: `{key, request_hash, response, status_code, created_at, expires_at}`
   - Return response with `Idempotency-Status: stored`

2. **Duplicate request (same key + body)**:
   - Return cached response
   - Add header `Idempotency-Status: replayed`

3. **Conflict (same key, different body)**:
   - Return `409 Conflict` with RFC 7807 Problem (`VAL-100`)

**Storage Schema (PostgreSQL):**

```sql
CREATE TABLE idempotency_keys (
    key             TEXT PRIMARY KEY,
    request_hash    TEXT NOT NULL,
    status_code     INTEGER NOT NULL,
    response_headers JSONB NOT NULL,
    response_body   BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys(expires_at);
```

**Configuration:**
- TTL: 24 hours (configurable)
- Key format: UUID v4 required
- Cleanup: Scheduled job removes expired records

**Middleware Location:** `internal/transport/http/middleware/idempotency.go`

## Consequences

### Positive

- **Safe retries**: Clients can retry without side effects
- **User experience**: Accidental double-submits handled gracefully
- **Debugging**: Idempotency status in response headers

### Negative

- **Database overhead**: Additional table and queries
- **Storage growth**: Records accumulate until cleanup
- **Request body hashing**: Computational overhead per request

### Neutral

- Only applies to POST/PUT/PATCH endpoints
- Optional for clients (requests without key processed normally)

## Related ADRs

- [ADR-001: Hexagonal Architecture](./ADR-001-hexagonal-architecture.md) - Middleware in transport, storage in infra
- [ADR-004: RFC 7807 Error Handling](./ADR-004-rfc7807-error-handling.md) - Conflict responses use RFC 7807 format
