---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - docs/prd.md
  - docs/architecture-decisions.md
  - project_context.md
---

# Go Golden Template - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for **Backend Service Golang Boilerplate** (Go Golden Template), decomposing the requirements from the PRD and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements (43 FRs)

| ID | Category | Requirement |
|----|----------|-------------|
| FR1 | Quality Gates | Developer dapat menjalankan `make lint` dan mendapatkan hasil golangci-lint v2 |
| FR2 | Quality Gates | CI pipeline blocks PR jika lint violations terdeteksi |
| FR3 | Quality Gates | depguard rules mencegah import yang melanggar layer boundaries |
| FR4 | Quality Gates | Developer dapat melihat lint violations dengan link ke rule + contoh fix |
| FR5 | Quality Gates | Coverage report per package/layer dengan baseline file |
| FR6 | Developer Experience | Developer dapat menjalankan `make up` untuk start semua services |
| FR7 | Developer Experience | Developer dapat menjalankan `make verify` untuk lint + unit tests |
| FR8 | Developer Experience | Developer dapat menjalankan `make reset` untuk clean slate |
| FR9 | Developer Experience | `make hooks` installs pre-commit hooks |
| FR10 | Developer Experience | `bplat generate` scaffolds new domain entity |
| FR11 | Developer Experience | First-time setup dalam ≤4 jam |
| FR12 | API Standards | REST endpoints menggunakan path-based versioning `/api/v1` |
| FR13 | API Standards | gRPC menggunakan package name `service.v1.NoteService` |
| FR14 | API Standards | Response menggunakan Envelope{data, error, meta} |
| FR15 | API Standards | meta.trace_id mandatory di semua responses |
| FR16 | API Standards | error.code menggunakan public UPPER_SNAKE codes |
| FR17 | Authentication | JWT middleware untuk token validation |
| FR18 | Authentication | RBAC middleware dengan RequireRole() |
| FR19 | Authentication | API Key support dengan rotation dan revocation |
| FR20 | Authentication | Multiple active API keys dengan expiry |
| FR21 | Authentication | Auth context unified interface |
| FR22 | Rate Limiting | Template default ON rate limiting |
| FR23 | Rate Limiting | Per-route opt-in configuration |
| FR24 | Rate Limiting | Redis-backed rate limiter |
| FR25 | Rate Limiting | Health checks exempt dari rate limiting |
| FR26 | Context Propagation | Semua IO functions menerima context.Context first |
| FR27 | Context Propagation | Context propagation enforced via linter |
| FR28 | Context Propagation | Wrapper pattern untuk DB, Redis, HTTP |
| FR29 | Context Propagation | Default timeout di wrappers |
| FR30 | Documentation | OpenAPI spec auto-generated/maintained |
| FR31 | Documentation | Proto files dengan documentation |
| FR32 | Documentation | Golden path cookbook |
| FR33 | Documentation | API documentation dari code annotations |
| FR34 | Observability | Structured JSON logging dengan Zap |
| FR35 | Observability | Prometheus metrics |
| FR36 | Observability | OpenTelemetry tracing |
| FR37 | Observability | Audit logging untuk sensitive operations |
| FR38 | Observability | Trace correlation across services |
| FR39 | Observability | Runbook templates untuk incidents |
| FR40 | Security | System fail-fast jika config required missing |
| FR41 | Security | CI pipeline secret scanning |
| FR42 | Policy | Policy pack sebagai single source of truth |
| FR43 | Migration | Compatibility mode untuk response/route |

### Non-Functional Requirements (22 NFRs)

| ID | Category | Requirement |
|----|----------|-------------|
| NFR-P1 | Performance | CI full pipeline p50 ≤8min, p95 ≤15min |
| NFR-P2 | Performance | Quick checks ≤5min |
| NFR-P3 | Performance | make verify (lint+unit) ≤3min |
| NFR-P4 | Performance | Integration tests ≤10-15min |
| NFR-P5 | Performance | make lint ≤60sec |
| NFR-S1 | Security | 0 Critical vulnerabilities |
| NFR-S2 | Security | High vulns require waiver + expiry |
| NFR-S3 | Security | Secrets via env/secret manager only |
| NFR-R1 | Reliability | Test flake rate <1% |
| NFR-R2 | Reliability | CI pass rate >95% |
| NFR-R3 | Reliability | Flaky tests quarantined |
| NFR-R4 | Reliability | Test reproducibility ≥99% |
| NFR-M1 | Maintainability | Coverage ≥80% for domain/usecase |
| NFR-M2 | Maintainability | Coverage baseline file |
| NFR-M3 | Maintainability | Cyclomatic complexity ≤15 (gocyclo) |
| NFR-M4 | Maintainability | Duplication detection (dupl) |
| NFR-M5 | Maintainability | Public API docs coverage |
| NFR-DX1 | Dev Experience | Setup success rate ≥95% |
| NFR-DX2 | Dev Experience | TTFP ≤4 jam |
| NFR-DX3 | Dev Experience | make up ≤2min |
| NFR-DX4 | Dev Experience | Pre-commit hooks available |
| NFR-DX5 | Dev Experience | First-time failure rate <5% |

### Additional Requirements (from Architecture)

- Brownfield upgrade (existing hexagonal architecture)
- policy/ directory as single source of truth
- Spec-first OpenAPI with ogen/oapi-codegen
- Context propagation via linter + wrapper pattern
- Hybrid error registry (central + domain-specific)
- Hybrid test organization (unit collocated, integration separate)
- New tooling: golangci-lint v2, depguard, buf, openapi-diff, sqlc-verify, gitleaks, govulncheck

### FR Coverage Map

_To be populated after epic design_

## Epic List

_To be generated in Step 2_

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| FR1-FR5 | Epic 1 | Quality gates, lint, coverage |
| FR6-FR9 | Epic 1 | Make commands, dev workflow |
| FR10-FR11 | Epic 5 | bplat generate, TTFP |
| FR12-FR16 | Epic 2 | API versioning, Envelope, trace_id |
| FR17-FR21 | Epic 3 | JWT, RBAC, API Keys |
| FR22-FR25 | Epic 3 | Rate limiting |
| FR26-FR29 | Epic 2 | Context propagation |
| FR30-FR33 | Epic 5 | OpenAPI, proto docs, cookbook |
| FR34-FR39 | Epic 4 | Logging, metrics, tracing, audit |
| FR40 | Epic 1 | Fail-fast config |
| FR41 | Epic 5 | Secret scanning |
| FR42 | Epic 1 | Policy pack |
| FR43 | Epic 6 | Compatibility mode |

---

## Epic 1: Foundation & Quality Gates

**Goal:** Developer dapat setup project dan jalankan quality checks secara konsisten.

**FRs covered:** FR1-FR9, FR40, FR42
**Phase:** MVP

---

## Epic 2: API Standards & Response Contract

**Goal:** Semua API responses mengikuti Envelope contract yang konsisten dan traceable.

**FRs covered:** FR12-FR16, FR26-FR29
**Phase:** MVP

---

## Epic 3: Authentication & Authorization

**Goal:** Developer dapat implement secure auth dengan JWT + RBAC + API Keys.

**FRs covered:** FR17-FR25
**Phase:** MVP

---

## Epic 4: Observability & Tracing

**Goal:** Developer dapat trace requests end-to-end dan monitor service health.

**FRs covered:** FR34-FR39
**Phase:** Growth

---

## Epic 5: Documentation & Developer Experience

**Goal:** Developer baru dapat productive dalam ≤4 jam.

**FRs covered:** FR10-FR11, FR30-FR33, FR41
**Phase:** Growth

---

## Epic 6: Migration & Contract Stability

**Goal:** Team dapat migrate existing services tanpa breaking changes.

**FRs covered:** FR43
**Phase:** Vision

---

# Epic Stories

## Epic 1: Foundation & Quality Gates

### Story 1.1: Setup Policy Pack Directory

As a developer,
I want a centralized policy directory with lint configuration,
So that all quality checks use a single source of truth.

**Acceptance Criteria:**

**Given** a fresh clone of the repository
**When** I run `make lint`
**Then** golangci-lint v2 loads configuration from `policy/golangci.yml`
**And** the lint output shows enabled linters from the policy file

---

### Story 1.2: Configure depguard Layer Boundaries

As a developer,
I want depguard rules enforcing hexagonal layer boundaries,
So that I cannot accidentally import infra from usecase.

**Acceptance Criteria:**

**Given** code in `internal/usecase/` that imports `internal/infra/`
**When** I run `make lint`
**Then** depguard reports a boundary violation error
**And** the error message explains which layer rule was violated

---

### Story 1.3: Implement Makefile Developer Commands

As a developer,
I want `make up`, `make verify`, and `make reset` commands,
So that I can manage development workflow with one command.

**Acceptance Criteria:**

**Given** docker-compose is installed
**When** I run `make up`
**Then** all services (postgres, redis) start within 2 minutes
**And** running `make verify` completes lint + unit tests within 3 minutes
**And** running `make reset` stops containers and removes volumes

---

### Story 1.4: Add Pre-commit Hook Support

As a developer,
I want `make hooks` to install pre-commit hooks,
So that I catch lint issues before pushing.

**Acceptance Criteria:**

**Given** I have run `make hooks`
**When** I commit code with lint violations
**Then** the pre-commit hook runs and blocks the commit
**And** I see which violations need fixing

---

### Story 1.5: Setup CI Pipeline with Quality Gates

As a developer,
I want CI to block PRs with lint or test failures,
So that code quality is enforced automatically.

**Acceptance Criteria:**

**Given** a PR with golangci-lint violations
**When** CI pipeline runs
**Then** the PR is blocked with clear failure message
**And** the full pipeline completes within 15 minutes (p95)
**And** quick checks complete within 5 minutes

---

### Story 1.6: Implement Fail-Fast Config Validation

As a developer,
I want the server to fail immediately if required config is missing,
So that I don't discover config issues at runtime.

**Acceptance Criteria:**

**Given** a required environment variable is missing
**When** the server starts
**Then** it exits immediately with error code 1
**And** the error message identifies missing config without leaking secrets

---

## Epic 2: API Standards & Response Contract

### Story 2.1: Implement Response Envelope Package

As a developer,
I want a standard Envelope{data, error, meta} response package,
So that all API responses are consistent.

**Acceptance Criteria:**

**Given** any HTTP handler returning a response
**When** I use the response package
**Then** the JSON output follows {data, error, meta} structure
**And** meta.trace_id is automatically populated from context

---

### Story 2.2: Create Central Error Code Registry

As a developer,
I want typed domain errors with public codes,
So that errors are consistent and API clients can handle them.

**Acceptance Criteria:**

**Given** `internal/domain/errors/codes.go` with central registry
**When** I create a domain error with `errors.NewDomain("NOT_FOUND", msg)`
**Then** the error implements the standard error interface
**And** the code is UPPER_SNAKE format

---

### Story 2.3: Implement Context Wrapper Package

As a developer,
I want wrapper functions for DB and HTTP with mandatory context,
So that context propagation is consistent.

**Acceptance Criteria:**

**Given** `internal/infra/wrapper/` package
**When** I call `wrapper.Query(ctx, pool, query, args...)`
**Then** context is propagated correctly
**And** default timeout is applied if context has no deadline

---

### Story 2.4: Add HTTP Error Mapping Middleware

As a developer,
I want domain errors automatically mapped to HTTP status codes,
So that I don't repeat error handling logic.

**Acceptance Criteria:**

**Given** a handler returns `errors.ErrNotFound`
**When** the response middleware processes it
**Then** HTTP 404 is returned with Envelope error format
**And** error.code is "NOT_FOUND"

---

## Epic 3: Authentication & Authorization

### Story 3.1: Implement JWT Middleware

As a developer,
I want JWT validation middleware,
So that I can secure routes that require authentication.

**Acceptance Criteria:**

**Given** a request with valid JWT in Authorization header
**When** the middleware processes it
**Then** the request continues with user context populated
**And** invalid/expired tokens return 401 Unauthorized

---

### Story 3.2: Implement RBAC Middleware

As a developer,
I want RequireRole() middleware,
So that I can restrict routes to specific roles.

**Acceptance Criteria:**

**Given** a route protected with `RequireRole("admin")`
**When** a user without admin role accesses it
**Then** 403 Forbidden is returned
**And** the error includes required role information

---

### Story 3.3: Implement API Key Store

As a developer,
I want rotatable API keys with expiry,
So that service-to-service auth is secure and manageable.

**Acceptance Criteria:**

**Given** an API key with expiry date
**When** the API key is past expiry
**Then** requests are rejected with 401
**And** multiple active keys can exist per service

---

### Story 3.4: Implement Rate Limiting Middleware

As a developer,
I want Redis-backed rate limiting,
So that API endpoints are protected from abuse.

**Acceptance Criteria:**

**Given** a route with rate limit of 100 req/min
**When** requests exceed the limit
**Then** HTTP 429 is returned
**And** health check endpoints are exempt

---

## Epic 4: Observability & Tracing

### Story 4.1: Structured Logging with Zap

As a developer,
I want structured JSON logging with consistent field names,
So that logs are queryable in our logging platform.

**Acceptance Criteria:**

**Given** a log statement with trace context
**When** the log is output
**Then** it includes `trace_id`, `request_id` fields
**And** field names match policy/log-fields conventions

---

### Story 4.2: OpenTelemetry Tracing Integration

As a developer,
I want OTel tracing for HTTP and database calls,
So that I can trace requests end-to-end.

**Acceptance Criteria:**

**Given** an HTTP request
**When** it makes database queries
**Then** spans are created for HTTP handler and DB calls
**And** trace_id is propagated to child spans

---

### Story 4.3: Prometheus Metrics Endpoint

As a developer,
I want standard Prometheus metrics,
So that we can monitor service health.

**Acceptance Criteria:**

**Given** `/metrics` endpoint
**When** scraped by Prometheus
**Then** it returns Go runtime metrics and custom counters
**And** HTTP request latency histograms are available

---

### Story 4.4: Audit Logging for Sensitive Operations

As a developer,
I want audit logs for sensitive operations,
So that we have an audit trail for compliance.

**Acceptance Criteria:**

**Given** a sensitive operation (create/update/delete)
**When** it completes
**Then** an audit log entry is created
**And** sensitive fields are masked in the log

---

## Epic 5: Documentation & Developer Experience

### Story 5.1: Setup Spec-First OpenAPI

As a developer,
I want OpenAPI spec as source of truth,
So that API contract is explicit and diffable.

**Acceptance Criteria:**

**Given** `api/openapi.yaml` spec file
**When** I run `make openapi-gen`
**Then** server stubs are generated using ogen/oapi-codegen
**And** CI runs `make openapi-diff` to detect breaking changes

---

### Story 5.2: Implement bplat Generate Command

As a developer,
I want `bplat generate entity` CLI,
So that I can scaffold new domain entities quickly.

**Acceptance Criteria:**

**Given** command `bplat generate entity user`
**When** executed
**Then** it creates entity, errors, repository in `internal/domain/user/`
**And** creates usecase in `internal/usecase/user/`

---

### Story 5.3: Create Golden Path Examples

As a developer,
I want `examples/goldenpath/` with working examples,
So that I can learn patterns by example.

**Acceptance Criteria:**

**Given** the examples directory
**When** I read `examples/goldenpath/note_crud/`
**Then** it shows complete CRUD implementation following all patterns
**And** examples are kept in sync with main codebase via CI

---

### Story 5.4: Add Secret Scanning to CI

As a developer,
I want gitleaks scanning in CI,
So that we catch accidentally committed secrets.

**Acceptance Criteria:**

**Given** a PR with hardcoded secret
**When** CI pipeline runs
**Then** gitleaks blocks the PR
**And** the error shows which file/line contains the secret

---

## Epic 6: Migration & Contract Stability

### Story 6.1: Implement Contract Tests for Envelope

As a developer,
I want contract tests for response envelope,
So that we catch breaking changes before merge.

**Acceptance Criteria:**

**Given** contract test fixtures
**When** envelope structure changes
**Then** contract tests fail
**And** the failure explains what broke

---

### Story 6.2: Add Proto Breaking Change Check

As a developer,
I want buf breaking check in CI,
So that gRPC contract changes are controlled.

**Acceptance Criteria:**

**Given** a PR that removes a proto field
**When** CI runs `make proto-check`
**Then** buf reports the breaking change
**And** the PR is blocked

---

### Story 6.3: Create Migration Guide

As a developer,
I want a migration guide document,
So that existing services can adopt the golden template.

**Acceptance Criteria:**

**Given** the migration guide at `docs/migration.md`
**When** I follow the steps
**Then** I can incrementally adopt golden template patterns
**And** compatibility mode allows gradual migration
