# Project Context: Backend Service Golang Boilerplate

> Critical rules for AI agents implementing code in this project.

## Technology Stack

| Tech | Version | Package |
|------|---------|---------|
| Go | 1.24.x | - |
| Router | v5 | go-chi/chi/v5 |
| Database | v5 | jackc/pgx/v5 |
| Query | latest | sqlc-dev/sqlc |
| Logger | latest | go.uber.org/zap |
| Config | v2 | knadh/koanf/v2 |
| Tracing | latest | go.opentelemetry.io/otel |
| Testing | latest | stretchr/testify |

---

## Go-Specific Rules

### Naming
- Struct fields: `PascalCase`
- JSON tags: `snake_case`
- Files: `lower_snake_case.go`
- Packages: `lowercase`, singular

### Imports
- Group: stdlib → external → internal
- No dot imports
- Avoid import aliases unless conflict

### Error Handling
- Always wrap: `fmt.Errorf("context: %w", err)`
- Sentinel errors in `errors.go` per domain
- AppError for HTTP responses

---

## Framework Rules (chi + pgx)

### HTTP Handlers
```go
func (h *Handler) Method(w http.ResponseWriter, r *http.Request)
ctx := r.Context() // Always from request
```

### Database
- Use sqlc for queries
- TxManager for transactions
- Never expose *sql.Tx to usecase

---

## Testing Rules

### Style
- Table-driven + `t.Run` + AAA
- Use testify (require/assert)
- `t.Parallel()` when safe

### Structure
- Co-located: `handler_test.go`
- Naming: `Test<Thing>_<Behavior>`

---

## Code Quality

### Linting
- Must pass golangci-lint
- Cyclomatic complexity ≤ 15
- No circular imports

### Context
- `ctx context.Context` as FIRST param
- Never `context.Background()` mid-chain

---

## Critical Don'ts

❌ Create `common/utils/helpers` packages
❌ Use `zap.L()` global logger
❌ Import infra from usecase
❌ Import interface from infra
❌ Log same error at every layer
❌ Hardcode secrets
❌ Expose internal errors to clients

---

## Layer Boundaries

```
cmd/ → app only
domain → stdlib, runtimeutil only
usecase → domain only
infra → domain, observability
interface → usecase, domain, httpx
```

---

## Response Format

```json
{"success": bool, "data": {}, "error": {"code": "ERR_*", "message": ""}}
```

---

## Reference

Full details: `docs/architecture.md`
