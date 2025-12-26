---
stepsCompleted: [1, 2, 3, 4]
status: 'approved'
inputDocuments:
  - "_bmad-output/prd.md"
  - "_bmad-output/architecture-decisions.md"
  - "_bmad-output/architecture.md"
workflowType: 'epics-and-stories'
project_name: 'golang-api-hexagonal'
user_name: 'Gan'
date: '2024-12-24'
---

# golang-api-hexagonal - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for golang-api-hexagonal, decomposing the requirements from the PRD and Architecture decisions into implementable stories.

---

## Requirements Inventory

### Functional Requirements (52 FRs)

#### Category 1: Correctness & Bug Fixes (5 FRs)
- **FR1:** System uses `AUDIT_REDACT_EMAIL` config correctly for PII redaction
- **FR2:** Database pool lifecycle manages reconnection properly
- **FR3:** Audit events include `ActorID` and `RequestID` metadata
- **FR4:** UUID v7 parsing errors handled gracefully (no 500)
- **FR5:** Schema migration (`schema_info`) is idempotent

#### Category 2: Security & Authentication (12 FRs)
- **FR6:** System validates JWT with configured algorithm whitelist only
- **FR7:** System requires `exp` claim and validates `iss`/`aud` claims
- **FR8:** System enforces JWT when `ENV=production` (no-auth guard)
- **FR9:** System rejects JWT secrets shorter than 32 bytes
- **FR10:** System protects `/metrics` from public access (internal port/auth)
- **FR11:** System extracts real IP only when `TRUST_PROXY=true`
- **FR12:** Users can only access their own data (IDOR prevention)
- **FR13:** System applies consistent authorization at application layer
- **FR41:** System never logs or returns JWT_SECRET / Authorization header / sensitive tokens
- **FR42:** System returns generic 500 errors (details in logs only)
- **FR43:** System validates/normalizes inbound `X-Request-ID` (length/charset)
- **FR44:** Auth checks are constant-time to prevent user enumeration

#### Category 3: Observability & Correlation (7 FRs)
- **FR14:** System injects `request_id` into all log entries
- **FR15:** System injects `trace_id` into all log entries
- **FR16:** Error responses include `request_id` extension (RFC7807)
- **FR17:** Audit events include `request_id` for correlation
- **FR18:** Metrics use route patterns (no `{id}` in labels for safe cardinality)
- **FR45:** Error responses include BOTH `request_id` and `trace_id` extensions
- **FR46:** `/metrics` content is audited (no sensitive labels/values)

#### Category 4: API Contract & Reliability (9 FRs)
- **FR19:** API rejects JSON with unknown fields
- **FR20:** API rejects JSON with trailing data
- **FR21:** Health endpoint (`/ready`) does not mutate runtime state
- **FR22:** System configures HTTP `ReadHeaderTimeout`
- **FR23:** System performs graceful shutdown with cleanup chain
- **FR24:** Server responds with `Location` header on 201 Created
- **FR47:** System produces OpenAPI 3.1 spec for all endpoints
- **FR48:** CI runs OpenAPI-based contract tests
- **FR51:** System sets `MaxHeaderBytes` and `ReadHeaderTimeout` via config

#### Category 5: Rate Limiting & Networking (2 FRs)
- **FR49:** Integration test verifies IP extraction logic (TRUST_PROXY scenarios)
- **FR50:** RealIP middleware is enabled only when `TRUST_PROXY=true`

#### Category 6: Developer Experience (7 FRs)
- **FR25:** Developer can setup local environment with `make setup`
- **FR26:** Developer can run all unit tests with `make test`
- **FR27:** Developer can run integration tests with test database
- **FR28:** Developer can generate code with `make generate`
- **FR29:** CI verifies generated code is up-to-date (no diff)
- **FR30:** CI runs lint, test, vuln scan, secret scan gates
- **FR52:** CI enforces wire/sqlc generate idempotency (`git diff --exit-code`)

#### Category 7: Governance & Documentation (5 FRs)
- **FR31:** Project has `SECURITY.md` with threat model
- **FR32:** Project has `CONTRIBUTING.md` guidelines
- **FR33:** Project has runbook for common operations
- **FR34:** Project has ADRs for key decisions
- **FR35:** README documents setup in ≤5 commands

#### Category 8: Data Layer & Infrastructure (5 FRs)
- **FR36:** Database pool is configurable via environment variables
- **FR37:** Email uniqueness is case-insensitive (CITEXT)
- **FR38:** sqlc generates typed queries for Users module
- **FR39:** sqlc generates typed queries for Audit module
- **FR40:** Wire generates dependency injection code

---

### Non-Functional Requirements (31 NFRs)

#### Performance (5 NFRs)
- **NFR1:** API p99 < 100ms (non-DB); p99 < 300ms (DB-bound) at baseline load
- **NFR2:** Health check `/health` responds < 10ms
- **NFR3:** CI full pipeline ≤ 10 minutes with parallel jobs
- **NFR20:** Service startup < 10s (without migration)
- **NFR21:** Memory baseline < 200MB

#### Security (6 NFRs)
- **NFR4:** TLS 1.2+ if service terminates TLS; document ingress + mTLS option
- **NFR5:** No secrets in repo; env + `*_FILE` pattern; audit secret source
- **NFR6:** No sensitive data in logs (PII redaction active)
- **NFR22:** JWT rotation without downtime (dual keys / JWKS)
- **NFR23:** `/metrics` restricted (internal port / network / auth)
- **NFR24:** `Request-ID` bounded (max length/charset)

#### Reliability (4 NFRs)
- **NFR7:** Graceful shutdown ≤ 30s
- **NFR8:** Health checks distinguish liveness vs readiness
- **NFR9:** DB outage no runtime mutation; readiness can fail
- **NFR25:** All dependencies have timeouts (no unbounded waits)

#### Maintainability (6 NFRs)
- **NFR10:** Domain + App coverage ≥ 80%
- **NFR11:** golangci-lint clean with gosec, revive, errorlint, bodyclose
- **NFR12:** All dependencies vuln-scanned
- **NFR13:** Generated code (wire/sqlc) reproducible and idempotent
- **NFR26:** No infra leakage: domain/app no import wire/sqlc/vault
- **NFR27:** ADR coverage for major decisions

#### Observability (5 NFRs)
- **NFR14:** request_id + trace_id correlation in logs + RFC7807 + audit
- **NFR15:** Structured logging with consistent field names
- **NFR16:** OTel semantic conventions for traces and metrics
- **NFR28:** Metrics cardinality budget (route patterns only)
- **NFR29:** Audit redaction policy tested (AUDIT_REDACT_EMAIL)

#### Integration & Delivery (5 NFRs)
- **NFR17:** OTel exporter configurable via env
- **NFR18:** Prometheus metrics endpoint available
- **NFR19:** OIDC/JWKS rotation supported (future)
- **NFR30:** `make ci` = GitHub Actions (local parity)
- **NFR31:** SBOM generated + govulncheck/OSV scan in CI

---

### Additional Requirements from Architecture

**Brownfield Upgrade Constraints:**
- Hexagonal Architecture preserved (layer violations forbidden)
- No breaking changes (versioned approach if needed)
- Existing modules maintained (Users + Audit)

**Key Technical Decisions:**
- DI Framework: Google Wire (compile-time)
- Secret Management: Platform injection + `*_FILE` pattern
- SQL Layer: sqlc (infra-only, domain isolated)
- JWT: Whitelist-based algorithm (HS256 default)
- Errors: RFC 7807 + request_id/trace_id extensions

**Implementation Patterns:**
- Tests: co-located `*_test.go`
- Mocks: co-located small, `internal/mocks/` large
- Event naming: dot.separated lowercase (`user.created`)
- Logging: structured JSON, never log secrets/Authorization

---

### FR Coverage Map

| FR | Epic | Description |
|----|------|-------------|
| FR1-5 | Epic 1 | Correctness & Bug Fixes |
| FR6-13 | Epic 2 | Security Baseline |
| FR41-44 | Epic 2 | Security (secrets, errors, auth timing) |
| FR11, FR49-50 | Epic 2 | Security (networking/rate limit) |
| FR14-18 | Epic 3 | Observability & Correlation |
| FR45-46 | Epic 3 | Observability (RFC7807, metrics audit) |
| FR19-24 | Epic 4 | API Contract & Reliability |
| FR47-48, FR51 | Epic 4 | API (OpenAPI, timeouts) |
| FR36-39 | Epic 5 | Data Layer (sqlc, pool, CITEXT) |
| FR25-30 | Epic 6 | Developer Experience |
| FR40, FR52 | Epic 6 | DX (Wire, generate check) |
| FR31-35 | Epic 7 | Governance & Documentation |

---

## Epic List

### Epic 1: Correctness & Bug Fixes
**Phase:** MVP/P0

Fix existing bugs and ensure system behaves correctly before adding features.

**User Outcome:** System operates reliably without unexpected errors or data issues.

**FRs covered:** FR1, FR2, FR3, FR4, FR5

---

### Epic 2: Security Baseline + Networking
**Phase:** MVP/P0-P1

Establish comprehensive security posture including JWT validation, access control, and network security.

**User Outcome:** System is secure against common attack vectors and properly handles authentication/authorization.

**FRs covered:** FR6, FR7, FR8, FR9, FR10, FR11, FR12, FR13, FR41, FR42, FR43, FR44, FR49, FR50

---

### Epic 3: Observability & Correlation
**Phase:** MVP/P1

Implement comprehensive request tracing and correlation across all system layers.

**User Outcome:** Operators can trace any request through logs, traces, and audit events using request_id.

**FRs covered:** FR14, FR15, FR16, FR17, FR18, FR45, FR46

---

### Epic 4: API Contract & Reliability
**Phase:** MVP/P1

Enforce strict API contracts and ensure system reliability under various conditions.

**User Outcome:** API behaves predictably with clear contracts, proper error responses, and graceful degradation.

**FRs covered:** FR19, FR20, FR21, FR22, FR23, FR24, FR47, FR48, FR51

---

### Epic 5: Data Layer Modernization
**Phase:** Growth/P2

Modernize data layer with sqlc, optimized pool configuration, and proper data handling.

**User Outcome:** Data operations are type-safe, efficient, and properly configured.

**FRs covered:** FR36, FR37, FR38, FR39

---

### Epic 6: Developer Experience & Build System
**Phase:** Growth/P2

Optimize developer workflow with proper tooling, CI gates, and code generation.

**User Outcome:** Developers can setup, develop, and deploy efficiently with confidence.

**FRs covered:** FR25, FR26, FR27, FR28, FR29, FR30, FR40, FR52

---

### Epic 7: Governance & Documentation
**Phase:** Vision/P3

Establish project governance with security documentation, contribution guidelines, and operational runbooks.

**User Outcome:** Project is well-documented, maintainable, and ready for team adoption.

**FRs covered:** FR31, FR32, FR33, FR34, FR35

---
---

## Epic Stories

---

## Epic 1: Correctness & Bug Fixes

### Story 1.1: Fix AUDIT_REDACT_EMAIL Config
**FR:** FR1

**As a** system operator,  
**I want** the system to correctly use AUDIT_REDACT_EMAIL config,  
**So that** PII is properly redacted in audit logs based on configured mode.

**Acceptance Criteria:**

**Given** `AUDIT_REDACT_EMAIL="full"`  
**When** an audit event containing an email is recorded  
**Then** the stored audit payload contains email as `***REDACTED***`

**Given** `AUDIT_REDACT_EMAIL="partial"`  
**When** an audit event containing an email is recorded  
**Then** the stored audit payload contains partially redacted email (e.g., `u***@domain.com`)

**And** config is read from env and wired into RedactorConfig (no hardcode)  
**And** unit tests cover both "full" and "partial" modes  
**And** integration test ensures main bootstrap uses `cfg.AuditRedactEmail`

---

### Story 1.2: Fix Database Pool Reconnection
**FR:** FR2

**As a** system operator,  
**I want** the database pool to handle reconnection properly,  
**So that** the system recovers gracefully from transient DB outages.

**Acceptance Criteria:**

**Given** the database connection is lost  
**When** a new request arrives  
**Then** the pool attempts reconnection transparently  
**And** failed connections do not cause panics or pool corruption  
**And** pool metrics reflect connection state accurately  
**And** unit/integration tests verify reconnection behavior

---

### Story 1.3: Add ActorID and RequestID to Audit Events
**FR:** FR3

**As a** security auditor,  
**I want** audit events to include ActorID and RequestID metadata,  
**So that** I can trace who performed actions and correlate with request logs.

**Acceptance Criteria:**

**Given** a user performs an auditable action  
**When** the audit event is recorded  
**Then** the event includes `actor_id` (user ID or "system")  
**And** the event includes `request_id` from request context  
**And** unit tests verify both fields are populated correctly

---

### Story 1.4: Handle UUID v7 Parsing Gracefully
**FR:** FR4

**As a** developer,  
**I want** UUID v7 parsing errors to be handled gracefully,  
**So that** invalid IDs return 400 Bad Request instead of 500.

**Acceptance Criteria:**

**Given** an API request with an invalid UUID format  
**When** the handler parses the UUID  
**Then** the system returns 400 Bad Request with RFC7807 error  
**And** the error includes descriptive message (not stack trace)  
**And** unit tests cover valid and invalid UUID scenarios

---

### Story 1.5: Make Schema Migration Idempotent
**FR:** FR5

**As a** DevOps engineer,  
**I want** schema migrations to be idempotent,  
**So that** running migrations multiple times doesn't cause errors.

**Acceptance Criteria:**

**Given** migrations have already been applied  
**When** `make migrate` is run again  
**Then** migrations complete successfully without errors  
**And** `schema_info` table tracks applied migrations  
**And** integration test verifies double-run is safe

---

## Epic 2: Security Baseline + Networking

### Story 2.1: Implement JWT Algorithm Whitelist
**FR:** FR6

**As a** security engineer,  
**I want** JWT validation to use a configured algorithm whitelist,  
**So that** algorithm confusion attacks (alg:none) are prevented.

**Acceptance Criteria:**

**Given** `JWT_ALGO=HS256` is configured  
**When** a JWT with `alg:RS256` is presented  
**Then** the token is rejected with 401 Unauthorized  
**And** only configured algorithms are accepted  
**And** unit tests cover whitelist enforcement

---

### Story 2.2: Enforce JWT Claims Validation
**FR:** FR7

**As a** security engineer,  
**I want** JWT validation to require exp and validate iss/aud claims,  
**So that** expired or misrouted tokens are rejected.

**Acceptance Criteria:**

**Given** a JWT without `exp` claim  
**When** the token is validated  
**Then** the request is rejected with 401 Unauthorized

**Given** JWT with non-matching `iss` or `aud`  
**When** the token is validated  
**Then** the request is rejected with 401 Unauthorized  
**And** clock skew is configurable via `JWT_CLOCK_SKEW`  
**And** unit tests cover all claim validation scenarios

---

### Story 2.3: Implement No-Auth Guard for Production
**FR:** FR8

**As a** platform engineer,  
**I want** JWT to be enforced when ENV=production,  
**So that** dev configs cannot accidentally run in production.

**Acceptance Criteria:**

**Given** `ENV=production` and `JWT_SECRET` is empty  
**When** the application starts  
**Then** startup fails with clear error message

**Given** `ENV=production` and valid `JWT_SECRET`  
**When** the application starts  
**Then** all protected endpoints require valid JWT  
**And** integration test verifies guard behavior

---

### Story 2.4: Enforce Minimum JWT Secret Length
**FR:** FR9

**As a** security engineer,  
**I want** JWT secrets shorter than 32 bytes to be rejected,  
**So that** weak secrets cannot be used in production.

**Acceptance Criteria:**

**Given** `JWT_SECRET` is 20 bytes  
**When** the application starts  
**Then** startup fails with "JWT_SECRET must be >= 32 bytes"  
**And** unit test verifies length validation

---

### Story 2.5: Protect /metrics Endpoint
**FR:** FR10

**As a** security engineer,  
**I want** /metrics to be inaccessible from public port,  
**So that** internal metrics are not exposed.

**Acceptance Criteria:**

**Given** production deployment  
**When** accessing /metrics on public port (8080)  
**Then** the request returns 404 or is not routed

**Given** internal port (8081) is used  
**When** accessing /metrics  
**Then** metrics are available  
**And** documentation describes /metrics protection strategy

---

### Story 2.6: Implement TRUST_PROXY-Aware IP Extraction
**FR:** FR11, FR49, FR50

**As a** security engineer,  
**I want** real IP extraction only when TRUST_PROXY=true,  
**So that** IP spoofing via X-Forwarded-For is prevented.

**Acceptance Criteria:**

**Given** `TRUST_PROXY=false`  
**When** request has X-Forwarded-For header  
**Then** the header is ignored, remote address is used

**Given** `TRUST_PROXY=true`  
**When** request has X-Forwarded-For header  
**Then** Chi RealIP middleware extracts real IP  
**And** integration test verifies both scenarios

---

### Story 2.7: Implement IDOR Prevention
**FR:** FR12

**As a** user,  
**I want** to only access my own data,  
**So that** other users' data is protected.

**Acceptance Criteria:**

**Given** authenticated user A  
**When** requesting user B's data via /users/{id}  
**Then** the request returns 403 Forbidden  
**And** integration tests verify IDOR prevention

---

### Story 2.8: Implement Application-Layer Authorization
**FR:** FR13

**As a** developer,  
**I want** consistent authorization at application layer,  
**So that** access control is centralized and auditable.

**Acceptance Criteria:**

**Given** a protected endpoint  
**When** authorization check fails  
**Then** consistent error format is returned  
**And** authorization is checked in use-case layer (not handler)  
**And** audit log records authorization decisions

---

### Story 2.9: Prevent Secret Logging
**FR:** FR41

**As a** security engineer,  
**I want** JWT_SECRET and Authorization headers never logged,  
**So that** secrets don't leak to log aggregators.

**Acceptance Criteria:**

**Given** any log statement in the codebase  
**When** it could log request headers or config  
**Then** JWT_SECRET and Authorization are redacted  
**And** code review checklist includes secret logging check  
**And** grep search confirms no secret logging

---

### Story 2.10: Implement Generic 500 Errors
**FR:** FR42

**As a** security engineer,  
**I want** 500 errors to return generic messages,  
**So that** stack traces and internal details don't leak.

**Acceptance Criteria:**

**Given** an internal error occurs  
**When** the error is returned to client  
**Then** message is generic ("Internal Server Error")  
**And** full details are logged with request_id  
**And** unit test verifies error sanitization

---

### Story 2.11: Validate X-Request-ID Header
**FR:** FR43

**As a** security engineer,  
**I want** X-Request-ID to be validated and normalized,  
**So that** log injection attacks are prevented.

**Acceptance Criteria:**

**Given** X-Request-ID with invalid characters or >64 chars  
**When** the request is processed  
**Then** a new valid request_id is generated  
**And** logging uses only validated request_id  
**And** unit tests cover validation rules

---

### Story 2.12: Implement Constant-Time Auth
**FR:** FR44

**As a** security engineer,  
**I want** auth checks to use constant-time comparison,  
**So that** user enumeration via timing attacks is prevented.

**Acceptance Criteria:**

**Given** invalid credentials  
**When** auth check fails  
**Then** response time is constant (not faster for unknown users)  
**And** crypto/subtle.ConstantTimeCompare is used  
**And** unit test verifies constant-time behavior

---

## Epic 3: Observability & Correlation

### Story 3.1: Inject request_id into All Logs
**FR:** FR14

**As a** SRE,  
**I want** request_id injected into all log entries,  
**So that** I can filter logs by request.

**Acceptance Criteria:**

**Given** any request to the API  
**When** logs are written  
**Then** all log entries include `request_id` field  
**And** request_id is propagated via context  
**And** unit tests verify injection

---

### Story 3.2: Inject trace_id into All Logs
**FR:** FR15

**As a** SRE,  
**I want** trace_id injected into all log entries,  
**So that** I can correlate logs with distributed traces.

**Acceptance Criteria:**

**Given** tracing is enabled  
**When** logs are written  
**Then** all log entries include `trace_id` field  
**And** trace_id comes from OTel span context  
**And** unit tests verify injection

---

### Story 3.3: Add request_id to RFC7807 Errors
**FR:** FR16

**As a** developer,  
**I want** RFC7807 errors to include request_id,  
**So that** I can correlate client errors with server logs.

**Acceptance Criteria:**

**Given** any API error response  
**When** the error is formatted  
**Then** the response includes `request_id` extension  
**And** unit tests verify RFC7807 + request_id

---

### Story 3.4: Add request_id to Audit Events
**FR:** FR17

**As a** security auditor,  
**I want** audit events to include request_id,  
**So that** I can correlate audit with request logs.

**Acceptance Criteria:**

**Given** an auditable action  
**When** the audit event is recorded  
**Then** the event includes `request_id` field  
**And** unit tests verify correlation

---

### Story 3.5: Use Route Patterns in Metrics
**FR:** FR18

**As a** SRE,  
**I want** metrics to use route patterns not actual IDs,  
**So that** cardinality remains bounded.

**Acceptance Criteria:**

**Given** request to `/users/123`  
**When** metrics are recorded  
**Then** route label is `/users/{id}` not `/users/123`  
**And** Prometheus metrics audit confirms no high-cardinality labels

---

### Story 3.6: Add trace_id to RFC7807 Errors
**FR:** FR45

**As a** developer,  
**I want** RFC7807 errors to include trace_id,  
**So that** I can find distributed traces from error responses.

**Acceptance Criteria:**

**Given** tracing is enabled  
**When** an API error is returned  
**Then** the response includes `trace_id` extension  
**And** both request_id and trace_id are present  
**And** unit tests verify both fields

---

### Story 3.7: Audit /metrics Content
**FR:** FR46

**As a** security engineer,  
**I want** /metrics content audited for sensitive data,  
**So that** no PII or secrets leak via metrics.

**Acceptance Criteria:**

**Given** /metrics endpoint  
**When** metrics are scraped  
**Then** no labels contain user IDs, emails, or secrets  
**And** audit checklist for metrics content exists  
**And** integration test scrapes and validates metrics

---

## Epic 4: API Contract & Reliability

### Story 4.1: Reject JSON with Unknown Fields
**FR:** FR19

**As a** API consumer,  
**I want** unknown JSON fields to be rejected,  
**So that** typos in requests are caught early.

**Acceptance Criteria:**

**Given** POST /users with unknown field `usernmae`  
**When** the request is processed  
**Then** 400 Bad Request is returned  
**And** error message indicates unknown field  
**And** unit tests verify strict decoding

---

### Story 4.2: Reject JSON with Trailing Data
**FR:** FR20

**As a** API consumer,  
**I want** trailing data in JSON to be rejected,  
**So that** malformed requests are caught.

**Acceptance Criteria:**

**Given** request body `{"name":"foo"}extra`  
**When** the request is processed  
**Then** 400 Bad Request is returned  
**And** json.Decoder.DisallowUnknownFields + More() check used  
**And** unit tests verify rejection

---

### Story 4.3: Non-Mutating Health Endpoint
**FR:** FR21

**As a** platform engineer,  
**I want** /ready to not mutate runtime state,  
**So that** health checks don't cause side effects.

**Acceptance Criteria:**

**Given** /ready endpoint  
**When** probed repeatedly  
**Then** no DB pool reset, no state mutation occurs  
**And** code review confirms no side effects  
**And** integration test verifies idempotency

---

### Story 4.4: Configure HTTP Timeouts
**FR:** FR22, FR51

**As a** SRE,  
**I want** HTTP ReadHeaderTimeout and MaxHeaderBytes configured,  
**So that** slowloris attacks are mitigated.

**Acceptance Criteria:**

**Given** environment variables for timeouts  
**When** HTTP server starts  
**Then** ReadHeaderTimeout is set (default 10s)  
**And** MaxHeaderBytes is set (default 1MB)  
**And** values are configurable via env  
**And** unit tests verify configuration

---

### Story 4.5: Implement Graceful Shutdown
**FR:** FR23

**As a** platform engineer,  
**I want** graceful shutdown with cleanup chain,  
**So that** in-flight requests complete before termination.

**Acceptance Criteria:**

**Given** SIGTERM signal  
**When** shutdown is initiated  
**Then** new requests are rejected  
**And** in-flight requests complete (up to 30s timeout)  
**And** DB connections are closed cleanly  
**And** integration test verifies graceful shutdown

---

### Story 4.6: Return Location Header on 201
**FR:** FR24

**As a** API consumer,  
**I want** 201 Created responses to include Location header,  
**So that** I can navigate to the created resource.

**Acceptance Criteria:**

**Given** POST /users creates a user  
**When** the response is returned  
**Then** status is 201 Created  
**And** Location header is `/api/v1/users/{id}`  
**And** unit tests verify header presence

---

### Story 4.7: Generate OpenAPI 3.1 Spec
**FR:** FR47

**As a** API consumer,  
**I want** OpenAPI 3.1 spec available,  
**So that** I can generate clients and understand the API.

**Acceptance Criteria:**

**Given** the codebase  
**When** `make openapi` is run  
**Then** openapi.yaml is generated  
**And** spec is valid OpenAPI 3.1  
**And** spec matches actual endpoints

---

### Story 4.8: OpenAPI Contract Tests in CI
**FR:** FR48

**As a** developer,  
**I want** CI to run OpenAPI-based contract tests,  
**So that** API changes are validated against spec.

**Acceptance Criteria:**

**Given** openapi.yaml exists  
**When** CI runs  
**Then** contract tests validate responses match spec  
**And** CI fails if spec is out of sync

---

## Epic 5: Data Layer Modernization

### Story 5.1: Configurable Database Pool
**FR:** FR36

**As a** platform engineer,  
**I want** database pool configurable via environment,  
**So that** I can tune performance per deployment.

**Acceptance Criteria:**

**Given** `DB_POOL_MAX_CONNS`, `DB_POOL_MIN_CONNS`, `DB_POOL_MAX_LIFETIME`  
**When** the application starts  
**Then** pool is configured with these values  
**And** defaults are sensible (max=25, min=5, lifetime=1h)  
**And** unit tests verify configuration

---

### Story 5.2: Case-Insensitive Email with CITEXT
**FR:** FR37

**As a** user,  
**I want** email uniqueness to be case-insensitive,  
**So that** User@Example.com and user@example.com are the same.

**Acceptance Criteria:**

**Given** email column uses CITEXT type  
**When** registering with different case emails  
**Then** uniqueness constraint is enforced case-insensitively  
**And** migration adds CITEXT extension  
**And** integration test verifies behavior

---

### Story 5.3: sqlc for Users Module
**FR:** FR38

**As a** developer,  
**I want** sqlc-generated typed queries for Users,  
**So that** SQL is type-safe and reviewed.

**Acceptance Criteria:**

**Given** sqlc.yaml configuration  
**When** `make generate` is run  
**Then** Users module queries are generated  
**And** queries are in infra layer only (not domain)  
**And** unit tests use generated code

---

### Story 5.4: sqlc for Audit Module
**FR:** FR39

**As a** developer,  
**I want** sqlc-generated typed queries for Audit,  
**So that** audit queries are type-safe.

**Acceptance Criteria:**

**Given** sqlc.yaml configuration  
**When** `make generate` is run  
**Then** Audit module queries are generated  
**And** queries are in infra layer only  
**And** unit tests use generated code

---

## Epic 6: Developer Experience & Build System

### Story 6.1: Implement make setup
**FR:** FR25

**As a** new developer,  
**I want** `make setup` to prepare local environment,  
**So that** I can start working quickly.

**Acceptance Criteria:**

**Given** fresh clone  
**When** `make setup` is run  
**Then** dependencies are installed  
**And** .env.local is created from .env.example  
**And** tools are installed (golangci-lint, sqlc, wire)

---

### Story 6.2: Implement make test
**FR:** FR26

**As a** developer,  
**I want** `make test` to run all unit tests,  
**So that** I can verify changes quickly.

**Acceptance Criteria:**

**Given** codebase with tests  
**When** `make test` is run  
**Then** all unit tests execute  
**And** coverage report is generated  
**And** exit code reflects test status

---

### Story 6.3: Integration Tests with Test Database
**FR:** FR27

**As a** developer,  
**I want** integration tests with test database,  
**So that** I can verify DB interactions.

**Acceptance Criteria:**

**Given** test database `*_test` exists  
**When** `make test-integration` is run  
**Then** integration tests execute against test DB  
**And** tests are isolated and clean up after

---

### Story 6.4: Implement make generate
**FR:** FR28

**As a** developer,  
**I want** `make generate` to run all code generators,  
**So that** generated code is up-to-date.

**Acceptance Criteria:**

**Given** sqlc.yaml and wire configuration  
**When** `make generate` is run  
**Then** sqlc generates query code  
**And** wire generates DI code  
**And** go generate runs for other generators

---

### Story 6.5: CI Generate Check
**FR:** FR29, FR52

**As a** developer,  
**I want** CI to verify generated code is up-to-date,  
**So that** forgotten regeneration is caught.

**Acceptance Criteria:**

**Given** PR with code changes  
**When** CI runs  
**Then** `make generate` is run  
**And** `git diff --exit-code` checks for changes  
**And** CI fails if generated code is stale

---

### Story 6.6: CI Gates for Quality
**FR:** FR30

**As a** developer,  
**I want** CI to run lint, test, vuln, secret scans,  
**So that** quality is enforced automatically.

**Acceptance Criteria:**

**Given** PR with code changes  
**When** CI runs  
**Then** golangci-lint runs with required linters  
**And** go test runs with coverage  
**And** govulncheck scans for vulnerabilities  
**And** secret scanner checks for leaked secrets  
**And** CI fails if any gate fails

---

### Story 6.7: Wire DI Integration
**FR:** FR40

**As a** developer,  
**I want** Wire to generate dependency injection code,  
**So that** DI is compile-time safe.

**Acceptance Criteria:**

**Given** wire.go provider sets  
**When** `make generate` is run  
**Then** wire_gen.go is generated  
**And** DI graph is compile-time validated  
**And** no runtime reflection for DI

---

## Epic 7: Governance & Documentation

### Story 7.1: Create SECURITY.md
**FR:** FR31

**As a** security researcher,  
**I want** SECURITY.md with threat model,  
**So that** I understand security posture and can report issues.

**Acceptance Criteria:**

**Given** the repository  
**When** SECURITY.md is viewed  
**Then** it includes threat model summary  
**And** security contact/reporting process is documented  
**And** security-related design decisions are listed

---

### Story 7.2: Create CONTRIBUTING.md
**FR:** FR32

**As a** contributor,  
**I want** CONTRIBUTING.md with guidelines,  
**So that** I know how to contribute properly.

**Acceptance Criteria:**

**Given** the repository  
**When** CONTRIBUTING.md is viewed  
**Then** it includes PR process  
**And** coding standards are documented  
**And** testing requirements are clear

---

### Story 7.3: Create Operational Runbook
**FR:** FR33

**As a** SRE,  
**I want** runbook for common operations,  
**So that** I can troubleshoot issues quickly.

**Acceptance Criteria:**

**Given** docs/runbook.md  
**When** viewed  
**Then** it includes DB connection issues troubleshooting  
**And** JWT/auth issues troubleshooting  
**And** rate limiting adjustment procedures  
**And** deployment rollback procedures

---

### Story 7.4: Create ADRs for Key Decisions
**FR:** FR34

**As a** architect,  
**I want** ADRs documenting key decisions,  
**So that** future developers understand rationale.

**Acceptance Criteria:**

**Given** docs/adr/ directory  
**When** ADRs are reviewed  
**Then** ADR exists for: Wire DI choice  
**And** ADR exists for: sqlc choice  
**And** ADR exists for: `*_FILE` secret pattern  
**And** ADR exists for: JWT auth model

---

### Story 7.5: Update README with Quick Start
**FR:** FR35

**As a** new developer,  
**I want** README with ≤5 commands to run,  
**So that** I can get started quickly.

**Acceptance Criteria:**

**Given** README.md  
**When** Quick Start section is followed  
**Then** server is running in ≤5 commands  
**And** commands are: clone, setup, migrate, run, test  
**And** prerequisites are clearly listed


