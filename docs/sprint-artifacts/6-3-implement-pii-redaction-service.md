# Story 6.3: Implement PII Redaction Service

Status: Done

## Story

As a **developer**,
I want **automatic PII redaction in audit payloads**,
so that **sensitive data is never stored in audit logs**.

## Acceptance Criteria

1. **Given** audit payload contains PII fields (password, token, secret, authorization, creditCard, credit_card, ssn, email), **When** redaction service processes the payload, **Then** those fields are fully redacted (replaced with `"[REDACTED]"`).

2. **Given** configuration `AUDIT_REDACT_EMAIL=partial`, **When** email is redacted, **Then** email shows partial mask: `ab***@domain.com` (first 2 chars + domain).

3. **Given** `AUDIT_REDACT_EMAIL=full` (default), **When** email is redacted, **Then** email is replaced with `"[REDACTED]"`.

4. **And** redaction happens BEFORE conversion to []byte JSON, **And** original payload is NOT stored anywhere.

*Covers: FR37*

## Source of Truth (Important)

- The canonical requirements are in `docs/epics.md` under "Story 6.3".
- App layer patterns established in `internal/app/user/` use cases.
- Error patterns established in `internal/app/errors.go`.
- Configuration patterns established in `internal/infra/config/config.go`.
- If any snippet conflicts with `architecture.md`, **follow architecture.md**.

## Tasks / Subtasks

- [x] Task 1: Add configuration for email redaction mode (AC: #2, #3)
  - [x] 1.1 Add `AuditRedactEmail` field to `internal/infra/config/config.go` with `envconfig:"AUDIT_REDACT_EMAIL" default:"full"`
  - [x] 1.2 Validate `AuditRedactEmail` is either "full" or "partial"
  - [x] 1.3 Update `.env.example` with `AUDIT_REDACT_EMAIL=full` and comment explaining options

- [x] Task 2: Create redaction service interface in domain layer (AC: #1)
  - [x] 2.1 Create `internal/domain/redactor.go` (or add to existing domain file)
  - [x] 2.2 Define `RedactorConfig` struct with `EmailMode string` field (stdlib only!)
  - [x] 2.3 Define `Redactor` interface with method: `RedactMap(data map[string]any) map[string]any`
  - [x] 2.4 Domain layer must use stdlib only - NO encoding/json import

- [x] Task 3: Implement PII redaction logic (AC: #1, #2, #3, #4)
  - [x] 3.1 Create `internal/shared/redact/redactor.go` package
  - [x] 3.2 Implement `PIIRedactor` struct with `NewPIIRedactor(cfg domain.RedactorConfig) *PIIRedactor`
  - [x] 3.3 Implement `RedactMap(data map[string]any) map[string]any` method
  - [x] 3.4 Define PII field patterns (case-insensitive matching):
    - `password`, `token`, `secret`, `authorization`, `creditCard`, `credit_card`, `ssn`
    - `email` (with configurable mode)
  - [x] 3.5 For `email` with mode="partial": mask to `ab***@domain.com` pattern
  - [x] 3.6 For `email` with mode="full": replace with `"[REDACTED]"`
  - [x] 3.7 Process nested maps and slices recursively
  - [x] 3.8 NEVER modify original map - create new copy with redacted values

- [x] Task 4: Create JSON-to-bytes helper function (AC: #4)
  - [x] 4.1 Add helper function `RedactAndMarshal(redactor Redactor, data any) ([]byte, error)` in shared/redact
  - [x] 4.2 Convert input to map[string]any (handle struct, map, or JSON bytes)
  - [x] 4.3 Apply redaction
  - [x] 4.4 Marshal to JSON bytes for AuditEvent.Payload

- [x] Task 5: Write unit tests (AC: all)
  - [x] 5.1 Create `internal/shared/redact/redactor_test.go`
  - [x] 5.2 Test full redaction of standard PII fields (password, token, secret, etc.)
  - [x] 5.3 Test email full redaction mode
  - [x] 5.4 Test email partial redaction mode (first 2 chars + domain)
  - [x] 5.5 Test case-insensitive field matching (Password, PASSWORD, etc.)
  - [x] 5.6 Test nested object redaction
  - [x] 5.7 Test array/slice redaction
  - [x] 5.8 Test that original map is NOT modified (immutability)
  - [x] 5.9 Test edge cases: nil input, empty map, non-string values
  - [x] 5.10 Achieve ≥80% coverage for new code

- [x] Task 6: Verify layer compliance and integration (AC: implicit)
  - [x] 6.1 Run `make lint` to verify depguard rules pass (0 issues)
  - [x] 6.2 Run `make test` to ensure all unit tests pass
  - [x] 6.3 Run `make ci` (or `ALLOW_DIRTY=1 make ci`) for full CI check

## Dependencies & Blockers

- **Hard dependency:** Story 6.2 (Audit Event PostgreSQL Repository) - **DONE**
- Story 6.4 (Audit Event Service) will consume this redaction service
- Uses config pattern from `internal/infra/config/config.go`

## Assumptions & Open Questions

- Redaction service is placed in `internal/shared/redact/` as a utility (not app layer use case since it has no business logic state)
- Email partial masking shows first 2 characters only (e.g., `jo***@example.com`)
- If email has fewer than 2 characters before @, show what's available (e.g., `j***@d.com`)
- Redaction service should handle deeply nested structures (maps inside maps inside arrays)
- PII field matching is case-insensitive (`Password`, `PASSWORD`, `password` all match)

## Definition of Done

- Configuration option `AUDIT_REDACT_EMAIL` added with validation
- `Redactor` interface defined in domain layer (stdlib only)
- `PIIRedactor` implements `Redactor` interface in shared layer
- All PII fields are properly redacted
- Email redaction respects configuration (full/partial)
- Unit tests pass with ≥80% coverage
- `make lint` passes (layer compliance verified)
- `make ci` passes

## Non-Functional Requirements

- Domain layer: stdlib ONLY (no encoding/json, no external packages)
- Shared layer: can import stdlib, domain (not app, transport, infra)
- Redaction creates new map - NEVER modifies original
- Performance: O(n) where n = number of fields in payload

## Testing & Coverage

- Unit tests with table-driven test style
- Test all PII field types individually
- Test nested structures (maps, slices)
- Test email redaction modes
- Test case-insensitive matching
- Test immutability (original not modified)
- Aim for coverage ≥80% for new redaction code

## Dev Notes

### ⚠️ CRITICAL: Layer Rules

The redaction service will be placed in `internal/shared/` since it's a utility without business logic state:

```
✅ internal/shared/redact/
   - Can import: stdlib, domain
   - Cannot import: app, transport, infra
   
✅ internal/domain/redactor.go
   - STDLIB ONLY (no encoding/json!)
   - Interface definition only
```

### Existing Code Context

**From Story 6.1 & 6.2 (Domain + Repository - DONE):**
| File | Description |
|------|-------------|
| `internal/domain/audit.go` | AuditEvent entity with `Payload []byte` field |
| `internal/infra/postgres/audit_event_repo.go` | Repository stores `Payload` as JSONB |

**Reference Configuration Pattern:**
| File | Description |
|------|-------------|
| `internal/infra/config/config.go` | Existing config with envconfig tags |
| `.env.example` | Environment variable documentation |

**This story CREATES:**
| File | Description |
|------|-------------|
| `internal/domain/redactor.go` | Redactor interface (stdlib only) |
| `internal/shared/redact/redactor.go` | PIIRedactor implementation |
| `internal/shared/redact/redactor_test.go` | Unit tests |

**This story MODIFIES:**
| File | Description |
|------|-------------|
| `internal/infra/config/config.go` | Add `AuditRedactEmail` field |
| `.env.example` | Add AUDIT_REDACT_EMAIL documentation |
| `docs/sprint-artifacts/sprint-status.yaml` | Sprint tracking status updates |

### PII Field Patterns

Fields to redact (case-insensitive):
- `password` → `"[REDACTED]"`  
- `token` → `"[REDACTED]"`
- `secret` → `"[REDACTED]"`
- `authorization` → `"[REDACTED]"`
- `creditCard` / `credit_card` → `"[REDACTED]"`
- `ssn` → `"[REDACTED]"`
- `email` → depends on config mode

### Email Redaction Examples

**Full mode (default):**
```
"email": "john.doe@example.com" → "email": "[REDACTED]"
```

**Partial mode:**
```
"email": "john.doe@example.com" → "email": "jo***@example.com"
"email": "a@x.com" → "email": "a***@x.com"
```

### Implementation Pattern

```go
// internal/domain/redactor.go
package domain

// RedactorConfig holds configuration for PII redaction.
type RedactorConfig struct {
    EmailMode string // "full" or "partial"
}

// Redactor defines the interface for PII redaction.
type Redactor interface {
    // RedactMap processes a map and returns a new map with PII fields redacted.
    // Original map is NOT modified.
    RedactMap(data map[string]any) map[string]any
}
```

```go
// internal/shared/redact/redactor.go
package redact

import (
    "strings"
    
    "github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// PIIRedactor implements domain.Redactor for PII redaction.
type PIIRedactor struct {
    emailMode string
}

// NewPIIRedactor creates a new PIIRedactor with the given configuration.
func NewPIIRedactor(cfg domain.RedactorConfig) *PIIRedactor {
    return &PIIRedactor{emailMode: cfg.EmailMode}
}

// RedactMap processes a map and returns a new map with PII fields redacted.
func (r *PIIRedactor) RedactMap(data map[string]any) map[string]any {
    if data == nil {
        return nil
    }
    result := make(map[string]any, len(data))
    for k, v := range data {
        result[k] = r.redactValue(k, v)
    }
    return result
}

func (r *PIIRedactor) redactValue(key string, value any) any {
    lowerKey := strings.ToLower(key)
    
    // Check if this key is a PII field
    if r.isPIIField(lowerKey) {
        return r.redactPIIValue(lowerKey, value)
    }
    
    // Recursively handle nested structures
    switch v := value.(type) {
    case map[string]any:
        return r.RedactMap(v)
    case []any:
        return r.redactSlice(v)
    default:
        return v
    }
}

// isPIIField checks if a field name matches known PII patterns.
func (r *PIIRedactor) isPIIField(lowerKey string) bool {
    switch lowerKey {
    case "password", "token", "secret", "authorization", 
         "creditcard", "credit_card", "ssn", "email":
        return true
    }
    return false
}

// Compile-time interface check
var _ domain.Redactor = (*PIIRedactor)(nil)
```

### Config Addition Pattern

```go
// In internal/infra/config/config.go
type Config struct {
    // ... existing fields ...
    
    // Audit
    // AuditRedactEmail controls how email addresses are redacted in audit logs.
    // Options: "full" (default, replaces with [REDACTED]) or "partial" (shows first 2 chars + domain).
    AuditRedactEmail string `envconfig:"AUDIT_REDACT_EMAIL" default:"full"`
}

// In Validate():
switch c.AuditRedactEmail {
case "full", "partial":
default:
    return fmt.Errorf("invalid AUDIT_REDACT_EMAIL: must be 'full' or 'partial'")
}
```

### Verification Commands

```bash
# Run all unit tests
make test

# Run lint to verify layer compliance
make lint

# Run full local CI
ALLOW_DIRTY=1 make ci

# Check coverage for shared layer
go test -cover ./internal/shared/...
```

### References

- [Source: docs/epics.md#Story 6.3] - Acceptance criteria and FR coverage
- [Source: docs/architecture.md#Implementation Patterns] - Layer rules
- [Source: docs/project-context.md] - Critical layer rules and conventions
- [Source: internal/infra/config/config.go] - Configuration patterns
- [Source: internal/domain/audit.go] - AuditEvent entity with Payload field

### Learnings from Previous Stories

**From Story 6.1 & 6.2 (Audit Event Domain + Repository):**
1. Domain layer uses stdlib only - interface definitions only
2. `Payload` field is `[]byte` - pre-redacted JSON
3. Redaction happens BEFORE creating AuditEvent (in app layer or service)
4. Use compile-time interface checks

**From Story 5.x (Middleware Stories):**
1. Config additions need validation in `Validate()` method
2. Update `.env.example` with new variables
3. Unit tests should cover all branches

### Security Considerations

1. **Never store original:** Original payload with PII must never be stored
2. **Deep copy:** Always create new map, never modify original
3. **Recursive redaction:** Handle deeply nested structures
4. **Case-insensitive:** Match PII fields regardless of case
5. **Fail-safe:** If uncertain whether a field is PII, err on side of redaction

### Epic 6 Context

Epic 6 implements the Audit Trail System for compliance requirements:
- **6.1 (DONE):** Domain model (entity + repository interface)
- **6.2 (DONE):** PostgreSQL repository implementation
- **6.3 (this story):** PII redaction service
- **6.4:** Audit event service (app layer) - uses redactor
- **6.5:** Extensible event types

This story provides the redaction utility that 6.4 (service) will use before persisting audit events.

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-19)

- `docs/epics.md` - Story 6.3 acceptance criteria
- `docs/architecture.md` - Layer rules and patterns
- `docs/project-context.md` - Layer constraints and conventions
- `docs/sprint-artifacts/6-2-implement-audit-event-postgresql-repository.md` - Previous story
- `internal/infra/config/config.go` - Configuration patterns
- `internal/app/errors.go` - App layer error patterns
- `internal/domain/audit.go` - AuditEvent entity

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

N/A

### Completion Notes List

- ✅ Implemented PII redaction service with 94.8% test coverage
- ✅ Created `Redactor` interface in domain layer (stdlib only)
- ✅ Created `PIIRedactor` implementation in `internal/shared/redact/`
- ✅ 27 comprehensive unit tests covering all acceptance criteria
- ✅ Email redaction supports both "full" and "partial" modes
- ✅ Recursive redaction handles nested maps/slices
- ✅ Immutability ensured - original map never modified
- ✅ `RedactAndMarshal` helper supports map, struct, and []byte inputs
- ✅ Fixed pre-existing config test issue (empty DATABASE_URL validation)
- ✅ [Review-Fix] Implemented recursion depth limit (100) to prevent stack overflow
- ✅ [Review-Fix] Added performance TODO for struct serialization optimization
- ✅ [Review-Fix-2] Recursion limit now returns empty structures (not unredacted data)
- ✅ [Review-Fix-2] Config AuditRedactEmail normalizes case before validation
- ✅ [Review-Fix-2] Added test for unmarshalable struct in RedactAndMarshal
- ✅ [Review-Fix-2] Added test to verify PII does not leak at deep nesting
- ✅ [Review-Fix-3] Enhanced PII matching to catch substrings (e.g., "access_token")
- ✅ [Review-Fix-3] Moved PII field patterns to dedicated constants (maintainability)
- ✅ [Code-Review] Verified robustness with dedicated test suite (robustness_test.go)


### Change Log

- 2025-12-19: Implemented full PII redaction service with configuration, domain interface, shared implementation, and comprehensive tests
- 2025-12-19: [AI-Review] Fixed recursion safety issue and added performance optimization TODO
- 2025-12-19: [Polish] Fixed all low-priority review findings (constants, robustness, docs)
- 2025-12-19: [AI-Review-2] Fixed 2 MEDIUM + 3 LOW issues: (M1) recursion returns empty map for security, (M2) config normalizes case, (L1-L3) improved test coverage
- 2025-12-19: [Code-Review] Added `robustness_test.go` to verify PII matching edge cases and false positives

### File List

**New Files:**
- `internal/domain/redactor.go`
- `internal/shared/redact/redactor.go`
- `internal/shared/redact/redactor_test.go`
- `internal/shared/redact/robustness_test.go`

**Modified Files:**
- `internal/infra/config/config.go` (added AuditRedactEmail field + validation + case normalization)
- `internal/infra/config/config_test.go` (added AuditRedactEmail tests, fixed empty DATABASE_URL test)
- `.env.example` (added AUDIT_REDACT_EMAIL documentation)
- `docs/sprint-artifacts/sprint-status.yaml` (status updates)

