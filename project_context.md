# Project Context - Go Golden Template

> **Purpose:** Concise AI agent guide. Read before implementing any code.

## Critical Rules

### Layer Boundaries (enforced by depguard)
```
domain → (nothing)
usecase → domain only
interface → usecase, domain
infra → domain only
```

### Context Propagation
- **ALL IO functions MUST receive `ctx context.Context` as first parameter**
- Use `internal/infra/wrapper/` for DB, Redis, HTTP clients
- Linter `contextcheck` enforces this

### Error Handling
- Domain errors: `errors.NewDomain("CODE", "message")`
- Central registry: `internal/domain/errors/codes.go`
- Public codes: UPPER_SNAKE (e.g., `NOTE_NOT_FOUND`)
- Never expose stack traces in production

### API Response Format
```go
type Envelope struct {
    Data  any            `json:"data,omitempty"`
    Error *ErrorResponse `json:"error,omitempty"`
    Meta  *Meta          `json:"meta,omitempty"`
}
// meta.trace_id is MANDATORY
```

### Naming Conventions
| Element | Style | Example |
|---------|-------|---------|
| Files | snake_case | `note_handler.go` |
| Types | PascalCase | `NoteHandler` |
| JSON | snake_case | `created_at` |
| DB tables | snake_case plural | `notes` |
| Error codes | UPPER_SNAKE | `NOTE_NOT_FOUND` |

## Testing Rules

- Unit tests: collocated (`*_test.go`)
- Integration tests: `tests/integration/` with build tag
- Coverage: ≥80% for domain/usecase
- Run: `make verify` (lint+unit), `make verify-full` (all)

## Anti-Patterns (DO NOT)

- ❌ Import `infra` from `usecase`
- ❌ Skip context parameter on IO functions
- ❌ Use camelCase in JSON responses
- ❌ Hardcode secrets in code
- ❌ Return raw errors to clients

## Key Commands

```bash
make up        # Start all services
make verify    # Lint + unit tests
make reset     # Clean slate
make hooks     # Install pre-commit
```

## Reference Docs

- Architecture: `docs/architecture-decisions.md`
- Existing patterns: `docs/architecture.md`
- PRD: `docs/prd.md`
