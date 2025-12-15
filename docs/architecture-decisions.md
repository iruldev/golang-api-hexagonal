---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - docs/prd.md
  - docs/index.md
  - docs/architecture.md
  - docs/analysis/brainstorming-session-2025-12-15.md
  - docs/analysis/research/technical-go-golden-template-2025-12-15.md
workflowType: 'architecture'
lastStep: 1
project_name: 'backend service golang boilerplate'
user_name: 'Gan'
date: '2025-12-15'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Reference: Existing Architecture

> [!NOTE]
> This project already has comprehensive architecture documentation at `docs/architecture.md` from the document-project workflow.
> This Architecture Decision Document focuses on **new decisions and upgrades** for the golden template initiative.

---

## Project Context Analysis

### Requirements Overview

**Functional Requirements:** 43 FRs across 11 capability areas

| Area | Count | Key Capabilities |
|------|-------|-----------------|
| Quality Gates & CI/CD | 5 | golangci-lint, depguard, coverage enforcement |
| Developer Experience | 6 | make up/verify/reset, bplat generate, hooks |
| API Standards | 5 | Envelope response, versioning, trace_id |
| Authentication | 5 | JWT, RBAC, API Keys |
| Rate Limiting | 4 | Per-route config, Redis-backed |
| Context Propagation | 4 | Mandatory context, trace correlation |
| Documentation | 4 | OpenAPI, proto docs, golden path |
| Observability | 6 | Zap, Prometheus, OpenTelemetry |
| Security | 2 | Secret scanning, fail-fast config |
| Policy Enforcement | 1 | Policy pack as single source of truth |
| Migration Readiness | 1 | Compatibility mode, contract tests |

**Non-Functional Requirements:** 22 NFRs across 5 categories

| Category | Key Targets |
|----------|-------------|
| Performance | CI p50 ≤8min, make up ≤2min, lint ≤60sec |
| Security | 0 Critical vulns, secrets via env only |
| Reliability | CI flake <1%, pass rate >95% |
| Maintainability | Coverage ≥80%, complexity ≤15 |
| Developer Experience | Setup success ≥95%, TTFP ≤4 jam |

### Scale & Complexity

| Dimension | Assessment |
|-----------|------------|
| **Project Complexity** | Medium |
| **Primary Domain** | Platform/Infrastructure Tooling |
| **Technical Type** | API Backend Monolith Multi-Entrypoint |
| **Architectural Components** | ~15-20 |
| **Integration Complexity** | Low (internal tooling) |
| **Compliance** | Internal standards only |

### Technical Constraints & Dependencies

**Locked (Existing Codebase):**
- Go 1.24.x
- chi v5 router
- PostgreSQL + pgx/sqlc
- Redis + Asynq
- Hexagonal architecture pattern

**New Tooling (To Add):**
- golangci-lint v2 (strict mode)
- depguard (boundary enforcement)
- gitleaks (secret scanning)
- govulncheck (dependency scanning)

**Infrastructure:**
- Docker/docker-compose for local dev
- CI/CD pipeline (GitHub Actions)

### Cross-Cutting Concerns

| Concern | Impact | Components Affected |
|---------|--------|-------------------|
| **Context Propagation** | Mandatory | All layers |
| **Tracing** | All IO | HTTP, Worker, DB, Redis |
| **Logging** | Structured JSON | All components |
| **Error Handling** | Typed + Envelope | Interface → Domain |
| **Configuration** | koanf + validation | All entrypoints |
| **Testing** | ≥80% coverage | domain, usecase |

---

## Technology Stack Evaluation (Brownfield)

### Locked Technology Stack (Existing - No Change)

| Component | Technology | Version | Status |
|-----------|------------|---------|--------|
| Language | Go | 1.24.x | ✅ Locked |
| Router | chi | v5 | ✅ Locked |
| Database | PostgreSQL + pgx | latest | ✅ Locked |
| DB Queries | sqlc | latest | ✅ Locked |
| Cache/Queue | Redis + Asynq | latest | ✅ Locked |
| Kafka | Sarama | latest | ✅ Locked |
| RabbitMQ | amqp091-go | latest | ✅ Locked |
| GraphQL | gqlgen | latest | ✅ Locked |
| gRPC | grpc-go | latest | ✅ Locked |
| Config | koanf | v2 | ✅ Locked |
| Logging | Zap | latest | ✅ Locked |
| Tracing | OpenTelemetry | latest | ✅ Locked |
| Metrics | Prometheus | latest | ✅ Locked |

### New Tooling Additions (Golden Template)

| Tool | Purpose | Priority |
|------|---------|----------|
| golangci-lint v2 | Static analysis, strict mode | 🎯 MVP |
| depguard | Layer boundary enforcement | 🎯 MVP |
| gitleaks | Secret scanning pre-commit/CI | �� MVP |
| govulncheck | Dependency vulnerabilities | 🎯 MVP |
| gocyclo | Complexity tracking | 📈 Growth |
| dupl | Duplication detection | 📈 Growth |
| buf | Proto breaking change guard | 🆕 Add |
| openapi-diff | OpenAPI spec diff on PR | 🆕 Add |
| sqlc-verify | Schema/query mismatch check | 🆕 Add |
| cyclonedx-gomod | SBOM generation | ✨ Optional |

### CI/CD Pipeline Targets

| Target | Tools | Phase |
|--------|-------|-------|
| `make lint` | golangci-lint, depguard | MVP |
| `make test` | go test, coverage | MVP |
| `make verify` | lint + test + sqlc-verify | MVP |
| `make security` | gitleaks, govulncheck | MVP |
| `make proto-check` | buf breaking | Growth |
| `make openapi-diff` | openapi-diff | Growth |
| `make sbom` | cyclonedx-gomod | Vision |

### Architectural Patterns (Enhancement)

| Pattern | Current | Golden Template Enhancement |
|---------|---------|---------------------------|
| Hexagonal | ✅ Established | + depguard enforcement |
| Repository | ✅ Established | + context propagation check |
| Domain Errors | ✅ Basic | → typed errors + codes |
| DI | ✅ Constructor | No change |
| Multi-entrypoint | ✅ cmd/* | Standardize Makefile |

---

## Core Architectural Decisions

### Decision Summary

| # | Decision | Choice | Rationale |
|---|----------|--------|-----------|
| 1 | Error Code Registry | **Hybrid** | Central registry for public/stable codes, domain-specific for internal |
| 2 | golangci-lint Config | **Policy Pack Directory** | Single source of truth (lint + depguard + policies) |
| 3 | Context Propagation | **Both (Linter + Wrapper)** | Defense in depth from MVP |
| 4 | OpenAPI Generation | **Spec-first (ogen/oapi-codegen)** | Stable contract for diff/breaking check |
| 5 | Test Organization | **Hybrid** | Unit collocated, integration separate with tag |

---

### Decision 1: Error Code Registry Strategy

**Choice:** Hybrid

**Implementation:**
- Central registry: `internal/domain/errors/codes.go` for public/stable error codes
- Domain-specific: Each domain has own errors with convention `DOMAIN_ERROR_NAME`
- Public codes exposed in API responses, internal codes for logging only

**Example:**
```go
// internal/domain/errors/codes.go (public registry)
const (
    ErrCodeNotFound       = "NOT_FOUND"
    ErrCodeValidation     = "VALIDATION_ERROR"
    ErrCodeUnauthorized   = "UNAUTHORIZED"
    ErrCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
)

// internal/domain/note/errors.go (domain-specific)
var ErrNoteNotFound = errors.NewWithCode(codes.ErrCodeNotFound, "note not found")
```

---

### Decision 2: golangci-lint Configuration Location

**Choice:** Policy Pack Directory

**Structure:**
```
policy/
├── golangci.yml          # Main lint config
├── depguard.yml          # Layer boundary rules
├── error-codes.yml       # Error code registry
├── log-fields.yml        # Approved log field names
└── README.md             # Policy documentation
```

**CI Integration:**
- `make lint` reads from `policy/golangci.yml`
- All tools reference policy/ as single source of truth

---

### Decision 3: Context Propagation Enforcement

**Choice:** Both (Linter + Wrapper) from MVP

**Linter Configuration:**
```yaml
# policy/golangci.yml
linters:
  enable:
    - contextcheck
```

**Wrapper Pattern for IO:**
```go
// internal/infra/wrapper/db.go
func Query(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) (pgx.Rows, error) {
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    return pool.Query(ctx, query, args...)
}

// internal/infra/wrapper/http.go
func DoRequest(ctx context.Context, client *http.Client, req *http.Request) (*http.Response, error) {
    req = req.WithContext(ctx)
    // Default timeout if not set
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
        req = req.WithContext(ctx)
    }
    return client.Do(req)
}
```

---

### Decision 4: OpenAPI Generation Approach

**Choice:** Spec-first (ogen/oapi-codegen)

**Workflow:**
1. Write/modify `api/openapi.yaml` (source of truth)
2. Run `make openapi-gen` to generate server stubs
3. CI: `make openapi-diff` checks breaking changes on PR

**Tooling:**
- Generator: `ogen` or `oapi-codegen`
- Diff tool: `openapi-diff`
- Breaking check: Part of CI pipeline

**Benefits:**
- Contract is explicit and versioned
- Breaking changes detected before merge
- Documentation always in sync

---

### Decision 5: Test Organization

**Choice:** Hybrid

**Structure:**
```
internal/
├── usecase/
│   └── note/
│       ├── usecase.go
│       └── usecase_test.go      # Unit tests (collocated)
│
tests/
├── integration/
│   ├── note_api_test.go         # Integration tests
│   └── testutil/                # Shared test utilities
└── e2e/                         # End-to-end tests
```

**Conventions:**
- Unit tests: Same package, `*_test.go`
- Integration tests: `tests/integration/`, build tag `integration`
- Run: `make test` (unit), `make test-integration` (with tag)

---

### Decision Impact Analysis

**Implementation Sequence:**
1. Policy pack directory setup (Decision 2)
2. Error code registry (Decision 1)
3. Context wrappers (Decision 3)
4. OpenAPI spec migration (Decision 4)
5. Test reorganization (Decision 5)

**Cross-Component Dependencies:**
- Policy pack affects all CI/CD and local tooling
- Error codes affect Interface and Domain layers
- Context wrappers affect Infrastructure layer
- OpenAPI affects Interface layer
- Tests affect all layers

---

## Implementation Patterns & Consistency Rules

### Pattern Categories

**Conflict Prevention:** These patterns ensure all AI agents write compatible, consistent code.

### Naming Patterns (Established)

| Category | Convention | Example |
|----------|------------|---------|
| Go files | snake_case | `note_handler.go` |
| Packages | lowercase | `note`, `postgres` |
| Types | PascalCase | `NoteHandler`, `CreateNoteRequest` |
| Functions | PascalCase (exported) | `NewHandler()`, `Create()` |
| Variables | camelCase | `noteID`, `userRepo` |
| Constants | PascalCase/UPPER_SNAKE | `ErrNotFound`, `DEFAULT_TIMEOUT` |
| DB Tables | snake_case plural | `notes`, `api_keys` |
| DB Columns | snake_case | `created_at`, `user_id` |
| JSON fields | snake_case | `trace_id`, `error_code` |

### API Patterns (Established + Enhanced)

| Category | Convention | Example |
|----------|------------|---------|
| REST endpoints | Plural nouns, lowercase | `/api/v1/notes` |
| Route params | `{id}` style (chi) | `/api/v1/notes/{id}` |
| Query params | snake_case | `?page_size=10` |
| Headers | Title-Case | `X-Request-ID` |
| Response | Envelope snake_case | `{data, error, meta}` |
| Error codes | UPPER_SNAKE | `NOTE_NOT_FOUND` |

### Structure Patterns (Golden Template)

```
internal/
├── domain/{entity}/          # Entity, Errors, Repository interface
├── usecase/{entity}/         # Business logic
├── interface/
│   ├── http/{entity}/        # HTTP handlers, DTOs
│   ├── grpc/{entity}/        # gRPC handlers
│   └── graphql/              # GraphQL resolvers
└── infra/
    ├── postgres/             # DB implementations
    ├── redis/                # Cache implementations
    └── wrapper/              # Context wrappers (NEW)

policy/                       # Single source of truth (NEW)
├── golangci.yml
├── depguard.yml
└── error-codes.yml

api/                          # Specs (NEW)
├── openapi.yaml              # Source of truth for REST
└── proto/                    # gRPC definitions

tests/
├── integration/              # Integration tests with tag
└── e2e/                      # End-to-end tests
```

### Error Handling Pattern

```go
// Domain layer: typed errors with public codes
var ErrNoteNotFound = errors.NewDomain("NOTE_NOT_FOUND", "note not found")

// Interface layer: map to Envelope
func mapError(err error) response.Error {
    var domainErr *errors.DomainError
    if errors.As(err, &domainErr) {
        return response.Error{
            Code:    domainErr.Code,
            Message: domainErr.Message,
            Hint:    domainErr.Hint,
        }
    }
    return response.InternalError()
}
```

### Context Propagation Pattern

```go
// MANDATORY: All IO operations receive context first
func (r *NoteRepo) GetByID(ctx context.Context, id string) (*Note, error)

// Use wrappers for consistent timeout
result, err := wrapper.Query(ctx, pool, query, args...)
resp, err := wrapper.DoHTTP(ctx, client, req)
```

### Logging Pattern

```go
// Structured fields with consistent naming
logger.Info("note created",
    zap.String("note_id", note.ID),
    zap.String("trace_id", traceID),
    zap.Duration("duration", elapsed),
)
```

### Enforcement Guidelines

**All AI Agents MUST:**
1. Follow naming conventions from existing codebase
2. Use context as first parameter for all IO
3. Map domain errors to Envelope format
4. Write unit tests collocated, integration in `tests/`
5. Reference `policy/` for lint and depguard rules

**Linter Enforcement:**
- golangci-lint enforces code style
- depguard enforces layer boundaries
- contextcheck enforces context propagation

---

## Project Structure & Boundaries

### Complete Project Directory Structure

```
golang-api-hexagonal/
├── .github/workflows/          # CI pipelines (ci, lint, security)
├── api/
│   ├── openapi.yaml            # REST spec (source of truth)
│   └── proto/v1/               # gRPC definitions
├── cmd/
│   ├── server/                 # HTTP/gRPC/GraphQL
│   ├── worker/                 # Asynq background jobs
│   ├── scheduler/              # Cron jobs
│   └── bplat/                  # CLI tooling
├── db/
│   ├── migrations/             # SQL migrations
│   └── queries/                # sqlc queries
├── internal/
│   ├── config/                 # koanf configs
│   ├── domain/
│   │   ├── {entity}/           # Entity, Errors, Repository interface
│   │   └── errors/codes.go     # Central error registry
│   ├── usecase/{entity}/       # Business logic
│   ├── interface/
│   │   ├── http/
│   │   │   ├── middleware/     # Auth, RBAC, Security, RateLimit
│   │   │   ├── response/       # Envelope, error mapping
│   │   │   └── {entity}/       # Handlers, DTOs
│   │   ├── grpc/
│   │   └── graphql/
│   ├── infra/
│   │   ├── postgres/           # sqlc implementations
│   │   ├── redis/              # Cache, Asynq
│   │   └── wrapper/            # Context wrappers (NEW)
│   ├── observability/          # Zap, OTel, Prometheus, Audit
│   ├── runtimeutil/            # Feature flags
│   └── worker/                 # Asynq handlers
├── policy/                     # Single source of truth (NEW)
│   ├── golangci.yml
│   ├── depguard.yml
│   └── error-codes.yml
├── tests/
│   ├── integration/            # build tag: integration
│   └── e2e/
├── examples/goldenpath/        # Living reference
├── docs/
├── Makefile
├── docker-compose.yaml
├── .env.example
└── .pre-commit-config.yaml
```

### Architectural Boundaries

| Boundary | From | To | Allowed | Enforcement |
|----------|------|-----|---------|-------------|
| Domain → Nothing | `internal/domain/` | - | No deps | depguard |
| Usecase → Domain | `internal/usecase/` | `internal/domain/` | ✅ | depguard |
| Usecase → Infra | `internal/usecase/` | `internal/infra/` | ❌ | depguard |
| Interface → Usecase | `internal/interface/` | `internal/usecase/` | ✅ | depguard |
| Interface → Infra | `internal/interface/` | `internal/infra/` | ❌ | depguard |
| Infra → Domain | `internal/infra/` | `internal/domain/` | ✅ | depguard |

### FR to Structure Mapping

| FR Category | Location |
|-------------|----------|
| Quality Gates | `policy/`, `.github/workflows/` |
| Developer Experience | `Makefile`, `cmd/bplat/` |
| API Standards | `api/openapi.yaml`, `internal/interface/http/response/` |
| Authentication | `internal/interface/http/middleware/` |
| Rate Limiting | `internal/interface/http/middleware/` |
| Context Propagation | `internal/infra/wrapper/` |
| Documentation | `docs/`, `api/` |
| Observability | `internal/observability/` |

### Integration Points

| Integration | Location |
|-------------|----------|
| PostgreSQL (pgx+sqlc) | `internal/infra/postgres/` |
| Redis + Asynq | `internal/infra/redis/`, `internal/worker/` |
| HTTP Clients | `internal/infra/wrapper/http.go` |
| gRPC | `internal/interface/grpc/` |
| Kafka/RabbitMQ | `internal/infra/kafka/`, `internal/infra/rabbitmq/` |

---

## Architecture Validation Results

### Coherence Validation ✅

| Check | Status |
|-------|--------|
| Decision Compatibility | ✅ All tooling compatible with Go ecosystem |
| Pattern Consistency | ✅ Naming/structure aligned with hexagonal |
| Structure Alignment | ✅ Directories support all decisions |
| Version Compatibility | ✅ Go 1.24+, golangci-lint v2 |

### Requirements Coverage ✅

**Functional Requirements (43 FRs):** All covered by architectural decisions
**Non-Functional Requirements (22 NFRs):** All addressable with defined patterns

| FR Category | Architectural Support |
|-------------|----------------------|
| Quality Gates | `policy/`, CI workflows |
| Developer Experience | `Makefile`, `cmd/bplat/` |
| API Standards | `api/openapi.yaml`, `response/` |
| Auth & RBAC | `middleware/` |
| Context Propagation | `infra/wrapper/`, linter |
| Observability | `observability/` |
| Security | `policy/`, CI security jobs |

### Implementation Readiness ✅

| Aspect | Status |
|--------|--------|
| Decision Completeness | ✅ 5 core decisions documented |
| Pattern Completeness | ✅ Naming, API, structure, errors, context |
| Structure Completeness | ✅ Full directory tree with mappings |
| Enforcement Mechanisms | ✅ depguard, golangci-lint |

### Architecture Completeness Checklist

- [x] Project context analyzed
- [x] Technology stack locked + tooling additions
- [x] 5 Core architectural decisions documented
- [x] Implementation patterns defined
- [x] Complete project structure with boundaries
- [x] FR/NFR coverage verified
- [x] Enforcement mechanisms defined

### Architecture Readiness Assessment

**Overall Status:** READY FOR IMPLEMENTATION 🚀
**Confidence Level:** HIGH

**Key Strengths:**
- Existing hexagonal architecture provides solid foundation
- Policy pack as single source of truth
- Spec-first OpenAPI enables contract stability
- Context propagation enforced via linter + wrapper

**First Implementation Priority:**
1. Setup `policy/` directory with golangci-lint v2 config
2. Configure depguard for layer boundaries
3. Create `internal/infra/wrapper/` for context propagation
4. Migrate to spec-first OpenAPI

---

## Architecture Completion Summary

### Workflow Completion

- **Architecture Decision Workflow:** COMPLETED ✅
- **Total Steps Completed:** 8
- **Date Completed:** 2025-12-15
- **Document Location:** `docs/architecture-decisions.md`

### Final Architecture Deliverables

| Deliverable | Status |
|-------------|--------|
| Core Decisions | 5 decisions documented |
| Implementation Patterns | 6 pattern categories |
| Project Structure | Complete with boundaries |
| FR/NFR Coverage | 43 FRs + 22 NFRs mapped |
| Validation | All checks passed |

### Implementation Handoff

**First Implementation Priority:**
1. Setup `policy/` directory with golangci-lint v2 config
2. Configure depguard rules in `policy/depguard.yml`
3. Create `internal/infra/wrapper/` for context propagation
4. Setup spec-first OpenAPI with ogen/oapi-codegen
5. Create `internal/domain/errors/codes.go` registry

**AI Agent Guidelines:**
- Follow all architectural decisions exactly as documented
- Use implementation patterns consistently
- Respect project structure and layer boundaries
- Reference this document for all architectural questions

---

**Architecture Status:** READY FOR IMPLEMENTATION ✅

*Generated by BMad Method create-architecture workflow on 2025-12-15*
