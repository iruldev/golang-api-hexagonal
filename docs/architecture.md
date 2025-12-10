---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - 'docs/prd.md'
  - 'docs/analysis/product-brief-backend-service-golang-boilerplate-2025-12-10.md'
  - 'docs/analysis/research/technical-golang-enterprise-boilerplate-research-2025-12-10.md'
  - 'docs/analysis/brainstorming-session-2025-12-10.md'
documentCounts:
  prd: 1
  research: 1
  briefs: 1
  brainstorming: 1
  ux: 0
  epics: 0
workflowType: 'architecture'
lastStep: 8
status: 'complete'
completedAt: '2025-12-10'
project_name: 'backend service golang boilerplate'
user_name: 'Gan'
date: '2025-12-10'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

---

## Project Context Analysis

### Project Overview

**Name:** Backend Service Golang Boilerplate – Enterprise Golden Template
**Type:** Developer Tool / Backend Service Template
**Language:** Go 1.24.x

**Core Objectives:**
- Setup < 30 minutes (clone → running)
- First feature ≤ 2 days
- 70-80% adoption in 12 months
- Observability baseline automatic

---

### Problems Solved

| Problem | Solution |
|---------|----------|
| Repetitive setup | One-command template |
| Inconsistent patterns | Enforced hexagonal structure |
| Missing prod basics | Three Pillars default |
| Slow onboarding | AGENTS.md + example module |

---

### Architectural Style

**Style:** Hexagonal / Clean Architecture

```
interface (HTTP, jobs)
    ↓
usecase (application logic)
    ↓
domain (entities, rules)
    ↓
infra (DB, external)
```

**Philosophy:**
- **Three Pillars:** Berjalan – Diamati – Dipercaya
- **Opinionated but not locking:** Clear stack + hook interfaces
- **AI-native:** AGENTS.md as explicit contract

---

### Technical Constraints

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Runtime | Go 1.24.x | Modern, stable |
| Router | chi | Lightweight, idiomatic |
| Database | PostgreSQL + pgx + sqlc | Type-safe SQL |
| Logger | zap | Structured, fast |
| Config | koanf | Flexible, typed |
| Tracing | OpenTelemetry | Industry standard |
| Metrics | Prometheus | Cloud-native |

**Out of Scope v1:** gRPC, GraphQL, multi-DB, full auth/RBAC, feature flags

---

### Cross-Cutting Concerns

| Concern | Architectural Implication |
|---------|--------------------------|
| Logging | Consistent pattern, trace/request ID correlation |
| Tracing | Context flows HTTP → usecase → infra via OTEL |
| Errors | AppError model, consistent mapping to HTTP |
| Config | Centralized, accessible without globals |
| Testing | Layer-specific patterns (unit/integration) |

---

### Stakeholders

| Persona | Focus |
|---------|-------|
| Andi (Engineer) | Fast setup, domain logic |
| Rudi (Tech Lead) | Standards, architecture |
| Maya (SRE) | Observability, reliability |
| Dina (Junior) | Onboarding, learning |
| AI Assistant | AGENTS.md compliance |

---

## Technology Stack (Confirmed)

*Note: This boilerplate IS the starter template. Stack decisions from PRD confirmed.*

| Component | Package | Rationale |
|-----------|---------|-----------|
| **Go** | 1.24.x | Modern, stable |
| **Router** | go-chi/chi/v5 | Lightweight, idiomatic |
| **Database** | jackc/pgx/v5 | Fast, pure Go |
| **Query** | sqlc-dev/sqlc | Type-safe SQL |
| **Logger** | go.uber.org/zap | Structured, fast |
| **Config** | knadh/koanf/v2 | Flexible, typed |
| **Tracing** | go.opentelemetry.io/otel | Industry standard |
| **Testing** | stretchr/testify | Assertion library |
| **Linting** | golangci-lint | Curated rules |

---

## Core Architectural Decisions

### Data Architecture

#### Migration Tool
**Decision:** golang-migrate
- Folder: `db/migrations`
- Naming: `YYYYMMDDHHMMSS_description.{up,down}.sql`

#### Connection Pool
| Setting | Env Var | Default |
|---------|---------|---------|
| Max open | `DB_MAX_OPEN_CONNS` | 20 |
| Max idle | `DB_MAX_IDLE_CONNS` | 5 |
| Lifetime | `DB_CONN_MAX_LIFETIME` | 30m |

#### Transaction Pattern
**TxManager interface:** Keeps usecase layer SQL-agnostic

---

### Security Baseline

#### Middleware Order (outer → inner)
1. Recovery → 2. Request ID → 3. OTEL → 4. Logging → 5. Auth hook → 6. Handler

#### Validation
- **HTTP:** go-playground/validator
- **Domain:** Business rules → `ErrInvalidInput`

#### Error Sanitization
- Response: Generic messages
- Logs: Full details, no secrets

---

### API Patterns

#### Response Envelope
```json
{"success": bool, "data": {}, "error": {"code": "ERR_*", "message": ""}}
```

#### Error Codes
| Code | HTTP |
|------|------|
| ERR_BAD_REQUEST | 400 |
| ERR_UNAUTHORIZED | 401 |
| ERR_FORBIDDEN | 403 |
| ERR_RESOURCE_NOT_FOUND | 404 |
| ERR_CONFLICT | 409 |
| ERR_INTERNAL | 500 |

#### Pagination
Limit/offset: `page`, `page_size` (default 20, max 100)

---

### Infrastructure & DX

#### Docker-Compose
| Service | Purpose |
|---------|---------|
| postgres | Database |
| jaeger | Tracing (optional) |

#### Makefile
| Target | Action |
|--------|--------|
| dev | docker-compose + go run |
| test | go test ./... |
| lint | golangci-lint |
| migrate-up/down | Migration |
| gen | sqlc generate |

#### Environment Variables
| Prefix | Variables |
|--------|-----------|
| APP_ | NAME, ENV, HTTP_PORT |
| DB_ | DSN, MAX_*_CONNS |
| OTEL_ | ENDPOINT, SERVICE_NAME |
| LOG_ | LEVEL, FORMAT |

---

## Implementation Patterns & Consistency Rules

### Naming Patterns

#### Structs & JSON
- Go fields: `PascalCase`
- JSON tags: `snake_case`
- Optional: add `omitempty`

#### Files
- Pattern: `lower_snake_case.go`
- Form: **singular** (`handler.go`, `entity.go`)
- Tests: co-located (`handler_test.go`)

#### Packages
- **lowercase**, no underscore, singular
- Domain: `internal/domain/note`
- **BANNED:** `common`, `utils`, `helpers`

---

### Testing Patterns

**Style:** Table-driven + `t.Run` + AAA

```go
func TestCreateNote(t *testing.T) {
    tests := []struct{...}{}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange → Act → Assert
        })
    }
}
```

- Use `testify` (require/assert)
- `t.Parallel()` when safe
- Naming: `Test<Thing>_<Behavior>`

---

### Error Handling

**AppError:** `Code`, `Message`, `Err` (wrapped)

**Wrapping:** `fmt.Errorf("ctx: %w", err)`

**Sentinels:** in `errors.go` per domain

---

### Logging

**Context-based:** `log.FromContext(ctx)`

**HTTP Fields:** trace_id, request_id, method, path, status, duration

**Levels:** Debug (internal), Info (state change), Warn (recoverable), Error (failure)

---

### Context & Signatures

`ctx context.Context` as **first param**

Never `context.Background()` mid-chain

---

### AI Guardrails (AGENTS.md)

**MUST:**
- JSON snake_case
- ctx first param
- Error wrapping %w
- Response envelope
- Table-driven tests

**MUST NOT:**
- `common/utils/helpers`
- `context.Background()` mid-chain
- Multi-layer error logging

---

## Project Structure & Boundaries

### High-level Layout

```
golang-api-hexagonal/
├── cmd/app/main.go              # Entry point
├── internal/
│   ├── app/app.go               # Composition root
│   ├── config/                  # Config loading
│   ├── domain/note/             # Entity, errors, repo interface
│   ├── usecase/note/            # Business logic + tests
│   ├── infra/postgres/note/     # Repository impl
│   ├── interface/http/
│   │   ├── note/handler.go      # HTTP handlers
│   │   ├── httpx/               # Response helpers
│   │   ├── middleware/          # Logging, tracing, recovery
│   │   ├── router.go
│   │   └── routes.go
│   ├── observability/           # Logger, tracer, metrics
│   └── runtimeutil/             # Clock, ID gen
├── db/migrations/               # SQL migrations
├── db/queries/                  # sqlc queries
├── sqlc.yaml
├── docker-compose.yaml
├── Makefile
├── .env.example
├── .golangci.yml
├── README.md, ARCHITECTURE.md, AGENTS.md
```

---

### Layer Boundaries

| Layer | Can Import | Cannot Import |
|-------|------------|---------------|
| cmd/ | app only | everything else |
| domain | stdlib, runtimeutil | usecase, interface, infra |
| usecase | domain | infra, interface/http |
| infra | domain, observability | interface, usecase |
| interface/http | usecase, domain, httpx | infra directly |

---

### Module Pattern: note

| Layer | Location |
|-------|----------|
| Entity | internal/domain/note/ |
| Usecase | internal/usecase/note/ |
| Postgres | internal/infra/postgres/note/ |
| HTTP | internal/interface/http/note/ |

*All new modules follow this pattern.*

---

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility:** All technologies work together
- Go 1.24.x + chi/v5 + pgx/v5 + sqlc = modern, type-safe
- zap + OTEL = unified observability
- koanf = flexible deployment config

**Pattern Consistency:** Uniform across layers
**Structure Alignment:** Hexagonal boundaries enforced

---

### Requirements Coverage ✅

| Source | Status |
|--------|--------|
| 56 FRs | Mapped to architectural components |
| 31 NFRs | Performance, reliability, observability |
| Three Pillars | Berjalan ✅ Diamati ✅ Dipercaya ✅ |

---

### Implementation Readiness ✅

- [x] Technology versions verified
- [x] Patterns comprehensive
- [x] Structure complete
- [x] Boundaries clear
- [x] AI guardrails ready

---

### Readiness Assessment

**Status:** ✅ READY FOR IMPLEMENTATION

**Confidence:** HIGH

**Strengths:**
- Clear hexagonal architecture
- AI-native design
- Three Pillars embedded
- Comprehensive patterns

**Future:** Diagrams, ADR history
