---
stepsCompleted: [1, 2, 3, 4, 5]
status: 'approved'
inputDocuments:

  - "_bmad-output/prd.md"
  - "_bmad-output/architecture.md"
  - "_bmad-output/index.md"
  - "_bmad-output/api-contracts.md"
  - "_bmad-output/data-models.md"
  - "_bmad-output/source-tree-analysis.md"
  - "_bmad-output/project-overview.md"
  - "_bmad-output/development-guide.md"
  - "_bmad-output/research/technical-production-boilerplate-research-2024-12-24.md"
hasProjectContext: false
workflowType: 'architecture'
project_name: 'golang-api-hexagonal'
user_name: 'Gan'
date: '2024-12-24'
---

# Architecture Decision Document - golang-api-hexagonal

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

---

## Brownfield Context

This is a **brownfield upgrade** of an existing Hexagonal Architecture codebase. The existing architecture documentation (`_bmad-output/architecture.md`) provides the baseline, and this workflow focuses on **decisions to upgrade it to production-grade**.

---

## Project Context Analysis

### Requirements Overview

**Functional Requirements (52 FRs across 8 areas):**
1. Correctness & Bug Fixes (5 FRs) - Config, pool, audit, UUID, migration
2. Security & Authentication (12 FRs) - JWT, no-auth guard, IDOR, secrets
3. Observability & Correlation (7 FRs) - request_id, trace_id, RFC7807
4. API Contract & Reliability (9 FRs) - JSON strict, timeouts, shutdown
5. Rate Limiting & Networking (2 FRs) - TRUST_PROXY, IP extraction
6. Developer Experience (7 FRs) - make commands, CI, tests
7. Governance & Documentation (5 FRs) - SECURITY.md, ADRs, runbook
8. Data Layer & Infrastructure (5 FRs) - Wire, sqlc, pool config

**Non-Functional Requirements (31 NFRs across 6 categories):**
- Performance: p99 < 100ms (non-DB), < 300ms (DB), startup < 10s
- Security: TLS 1.2+, no secrets in repo, JWT rotation, /metrics restricted
- Reliability: Graceful shutdown ≤ 30s, all deps have timeouts
- Maintainability: 80% coverage, golangci-lint clean, no infra leakage
- Observability: request_id + trace_id correlation, cardinality budget
- Integration: OTel configurable, Prometheus, SBOM + govulncheck

### Scale & Complexity

| Indicator | Value |
|-----------|-------|
| Primary Domain | API Backend (Hexagonal Architecture) |
| Complexity Level | Medium-High |
| Project Type | Brownfield Upgrade |
| Phased Roadmap | MVP → Growth → Vision |
| Estimated Components | ~15-20 architectural components |

### Technical Constraints & Dependencies

1. **Hexagonal Architecture preserved** - Layer violations forbidden
2. **No breaking changes** - Versioned approach (v2) if needed
3. **No multi-tenancy** - Design note only
4. **Existing modules** - Users + Audit (no new business modules)
5. **Key Decisions Fixed:**
   - DI: Google Wire (compile-time)
   - Secrets: Platform injection + `*_FILE` pattern
   - SQL: sqlc (infra-only, domain isolated)

### Cross-Cutting Concerns Identified

| Concern | Impact |
|---------|--------|
| **Authentication & Authorization** | JWT validation, no-auth guard, IDOR prevention |
| **Observability** | request_id + trace_id across logs, traces, audit |
| **Error Handling** | RFC7807, generic 500, no PII in errors |
| **Security** | Rate limiting, TRUST_PROXY, constant-time auth |
| **Pool/Connection Management** | Timeouts, graceful shutdown, reconnection |
| **Configuration** | ENV-based, `*_FILE` pattern, validation |

---

## Starter Template Evaluation

> **SKIPPED** - This is a brownfield upgrade of an existing codebase.
> 
> The existing Go + Hexagonal Architecture codebase serves as our foundation.
> No starter template evaluation needed.

**Existing Foundation:**
- Language: Go 1.21+
- Architecture: Hexagonal (Ports & Adapters)
- Router: Chi v5
- Database: PostgreSQL with pgx
- DI: Google Wire (to be migrated)
- SQL: sqlc (to be added)

---

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
1. DI Framework: Google Wire ✅
2. Secret Management: Platform injection + `*_FILE` ✅
3. SQL Layer: sqlc (infra-only) ✅
4. JWT Algorithm: Whitelist-based (HS256 default, RS256 optional)
5. No-auth Guard: ENV=production enforces JWT

**Important Decisions (Shape Architecture):**
1. /metrics protection: Internal port (8081) or network policy
2. Health check semantics: Liveness vs Readiness separation
3. Graceful shutdown: Signal → context cancel → cleanup chain
4. Rate limiting: IP-based with TRUST_PROXY awareness

**Deferred Decisions (Post-MVP):**
1. OpenAPI spec generation (P3)
2. JWKS rotation support (future)
3. mTLS internal communication (future)

### Data Architecture

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Database | PostgreSQL 14+ | Existing, CITEXT support |
| Driver | pgx/v5 | Existing, pool management |
| Pool Config | ENV-based (`DB_POOL_*`) | NFR requirement |
| Migration | Existing schema + schema_info | Idempotent, auditable |
| Email Uniqueness | CITEXT column type | Case-insensitive matching |
| Query Generation | sqlc (infra layer only) | Type-safe, domain isolated |

### Authentication & Security

| Decision | Choice | Rationale |
|----------|--------|-----------|
| JWT Library | golang-jwt/jwt/v5 | Existing, well-maintained |
| Algorithm | Whitelist (JWT_ALGO env) | Prevent alg:none attacks |
| Claims Validation | Require exp, validate iss/aud | Standard JWT security |
| Key Length | Min 32 bytes enforced | HMAC security requirement |
| No-Auth Guard | ENV=production → JWT required | Prevent dev config in prod |
| Auth Failures | Constant-time comparison | Prevent timing attacks |

### API & Communication Patterns

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Error Format | RFC 7807 + extensions | Industry standard, request_id/trace_id |
| JSON Parsing | Strict (reject unknown fields) | Contract enforcement |
| Rate Limiting | IP-based, TRUST_PROXY aware | Existing + fix spoof prevention |
| CORS | Configurable via env | Flexibility |
| Request Validation | X-Request-ID bounded (64 chars, UUID chars) | Log injection prevention |

### Infrastructure & Deployment

| Decision | Choice | Rationale |
|----------|--------|-----------|
| HTTP Server | ReadHeaderTimeout + MaxHeaderBytes | Slowloris prevention |
| Graceful Shutdown | 30s timeout, cleanup chain | NFR requirement |
| Metrics Port | Internal (8081) or network policy | Security isolation |
| Health Probes | Separate /health (liveness) and /ready (readiness) | K8s semantics |
| OTel Exporter | Configurable via env (OTLP default) | Platform agnostic |
| CI Pipeline | Lint + Test + Vuln + Secret + Generate check | NFR requirement |

### Decision Impact Analysis

**Implementation Sequence:**
1. Bug fixes (P0) - Config, pool, metadata
2. Security baseline (P0/P1) - JWT, no-auth guard
3. API contract (P1) - JSON strict, RFC7807
4. Data layer (P2) - Wire, sqlc
5. Governance (P3) - Docs, ADRs

**Cross-Component Dependencies:**
- JWT validation → all protected endpoints
- request_id → logging, errors, audit, traces
- Graceful shutdown → all connections (DB, HTTP)
- Wire DI → all service instantiation

---

## Implementation Patterns & Consistency Rules

### Naming Patterns

| Category | Convention | Example |
|----------|------------|---------|
| Database tables | snake_case, plural | `users`, `audit_events` |
| Database columns | snake_case | `created_at`, `user_id` |
| Go files | snake_case.go | `user_repository.go` |
| Go packages | lowercase | `domain`, `app`, `infra` |
| Go functions | CamelCase | `CreateUser`, `GetByID` |
| API routes | /api/v1/{resource} | `/api/v1/users` |
| JSON fields | snake_case | `user_id`, `created_at` |

### Structure Patterns

**Tests:**
- Co-located `*_test.go` in same directory as source
- Test files follow same package naming

**Mocks:**
- Small interfaces: co-located `*_mock.go`
- Large/central interfaces: `internal/mocks/`
- **Rule:** Mocks MUST NOT leak into domain/app packages

### API Format Patterns

**Success Responses:**
- Direct JSON objects (no global `{"data":...}` wrapper)
- Rationale: Avoid breaking changes to existing clients

**Error Responses (RFC 7807):**
```json
{
  "type": "about:blank",
  "title": "Bad Request",
  "status": 400,
  "detail": "Field validation failed",
  "instance": "/api/v1/users",
  "request_id": "abc-123-xyz",
  "trace_id": "1234567890abcdef"
}
```
- Always include `request_id`
- Include `trace_id` when tracing enabled
- **Rule:** Only additive changes allowed, fields MUST remain stable

### Event & Logging Patterns

**Audit Event Naming:**
- Dot-separated lowercase: `user.created`, `user.updated`, `user.accessed`
- Consistent namespace for easy filtering

**Logging Rules:**
- Structured JSON format (keep existing)
- **NEVER log:** Authorization header, JWT_SECRET, tokens
- Apply redaction for PII fields (email when AUDIT_REDACT_EMAIL=true)

### Error Handling Patterns

**Domain/App Errors:**
- Typed domain errors with error codes
- Mapped in transport layer to RFC 7807

**Generic 500 Errors:**
- Client receives: generic message only
- Logs receive: full details (sanitized, no PII)

**Request-ID Handling:**
- Accept `X-Request-ID` from client
- Validate: max 64 chars, UUID-safe charset
- Invalid/missing → generate new UUID

### Security/Networking Patterns

**TRUST_PROXY:**
- `TRUST_PROXY=true` → enable Chi RealIP middleware
- `TRUST_PROXY=false` → ignore X-Forwarded-For headers completely

**/metrics Protection:**
- Never expose publicly in production
- Options: internal port (8081) OR network policy OR auth

### Enforcement Guidelines

**All AI Agents MUST:**
1. Follow naming conventions exactly as specified
2. Never log secrets or Authorization headers
3. Return RFC 7807 errors with request_id
4. Keep mocks out of domain/app packages
5. Respect TRUST_PROXY setting for IP extraction

**Anti-Patterns (Avoid):**
- ❌ Wrapper responses like `{"data": ..., "error": ...}`
- ❌ Logging full request headers
- ❌ Trusting X-Forwarded-For when TRUST_PROXY=false
- ❌ Exposing /metrics on public port

