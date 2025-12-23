# Story 8.5: Create Guide for Adding New Modules

Status: Done

## Story

As a **developer**,
I want **step-by-step guide for adding new modules**,
so that **I can extend the application correctly**.

## Acceptance Criteria

1. **Given** I want to add a new module (e.g., "orders"), **When** I read `docs/guides/adding-module.md`, **Then** I see step-by-step instructions:
   1. Create domain entity and repository interface (`internal/domain/order.go`)
   2. Create migration (`make migrate-create name=create_orders`)
   3. Implement repository (`internal/infra/postgres/order_repo.go`)
   4. Create use cases (`internal/app/order/`)
   5. Create DTOs (`internal/transport/http/contract/order.go`)
   6. Create handlers (`internal/transport/http/handler/order.go`)
   7. Wire routes in router
   8. Add audit events
   9. Write tests

2. **And** each step references existing Users module as example

*Covers: FR68*

## Tasks / Subtasks

- [x] Task 1: Create `docs/guides/adding-module.md` Document Structure (AC: #1)
  - [x] 1.1 Create document with clear section headers
  - [x] 1.2 Add table of contents for easy navigation
  - [x] 1.3 Add overview explaining the purpose and approach

- [x] Task 2: Document Step 1 - Domain Entity and Repository Interface (AC: #1, #2)
  - [x] 2.1 Explain domain entity creation with `type ID string` (NOT uuid.UUID)
  - [x] 2.2 Show domain entity struct with validation method
  - [x] 2.3 Define repository interface in domain layer (accepts `Querier` parameter)
  - [x] 2.4 Document domain errors (sentinel errors pattern)
  - [x] 2.5 Reference `internal/domain/user.go` as example
  - [x] 2.6 Emphasize stdlib-only imports in domain layer

- [x] Task 3: Document Step 2 - Database Migration (AC: #1, #2)
  - [x] 3.1 Show `make migrate-create name=create_orders` command
  - [x] 3.2 Explain goose migration file format with `-- +goose Up` and `-- +goose Down`
  - [x] 3.3 Document naming conventions: snake_case, plural table names
  - [x] 3.4 Document UUID PRIMARY KEY without DEFAULT (app provides UUID v7)
  - [x] 3.5 Show index naming conventions (`idx_*`, `uniq_*`)
  - [x] 3.6 Reference `migrations/20251217000000_create_users.sql` as example

- [x] Task 4: Document Step 3 - Repository Implementation (AC: #1, #2)
  - [x] 4.1 Create repository struct implementing domain interface
  - [x] 4.2 Document `domain.ID ↔ uuid.UUID` conversion at infra boundary
  - [x] 4.3 Document error wrapping pattern with `op` string
  - [x] 4.4 Show `Querier` parameter usage for both pool and tx support
  - [x] 4.5 Document compile-time interface check `var _ domain.XRepository = (*XRepo)(nil)`
  - [x] 4.6 Reference `internal/infra/postgres/user_repo.go` as example

- [x] Task 5: Document Step 4 - Use Cases (AC: #1, #2)
  - [x] 5.1 Create use case package directory structure (`internal/app/order/`)
  - [x] 5.2 Document request/response struct patterns
  - [x] 5.3 Show `TxManager.WithTx()` usage for transactions
  - [x] 5.4 Document `AppError` creation with proper `Code`
  - [x] 5.5 Document dependency injection via constructor
  - [x] 5.6 Explain NO logging in app layer (tracing context only)
  - [x] 5.7 Reference `internal/app/user/create_user.go` as example

- [x] Task 6: Document Step 5 - DTOs/Contracts (AC: #1, #2)
  - [x] 6.1 Create request struct with validation tags (`validate:"required,..."`)
  - [x] 6.2 Create response struct with JSON tags (camelCase)
  - [x] 6.3 Document `ToXResponse()` mapper functions
  - [x] 6.4 Document pagination response pattern (if applicable)
  - [x] 6.5 Reference `internal/transport/http/contract/user.go` as example

- [x] Task 7: Document Step 6 - HTTP Handlers (AC: #1, #2)
  - [x] 7.1 Create handler struct with use case interfaces
  - [x] 7.2 Document UUID v7 generation at transport boundary
  - [x] 7.3 Document UUID parsing from path params with validation
  - [x] 7.4 Document error mapping to RFC 7807 Problem Details
  - [x] 7.5 Show `contract.WriteProblemJSON()` usage
  - [x] 7.6 Reference `internal/transport/http/handler/user.go` as example

- [x] Task 8: Document Step 7 - Wire Routes in Router (AC: #1, #2)
  - [x] 8.1 Define routes interface in router package
  - [x] 8.2 Add handler to `NewRouter()` function signature
  - [x] 8.3 Register routes in `/api/v1` route group
  - [x] 8.4 Document middleware application (JWT, rate limiting, etc.)
  - [x] 8.5 Reference `internal/transport/http/router.go` as example

- [x] Task 9: Document Step 8 - Add Audit Events (AC: #1, #2)
  - [x] 9.1 Create new event type constant in domain layer
  - [x] 9.2 Record audit events within transaction in use case
  - [x] 9.3 Document `AuditEventInput` struct usage
  - [x] 9.4 Reference `docs/guides/adding-audit-events.md` for details
  - [x] 9.5 Reference existing audit integration in `create_user.go`

- [x] Task 10: Document Step 9 - Write Tests (AC: #1, #2)
  - [x] 10.1 Document unit tests for domain entities (100% coverage target)
  - [x] 10.2 Document unit tests for use cases (90% coverage target)
  - [x] 10.3 Document handler tests with testify + httptest
  - [x] 10.4 Document integration tests for repositories with testcontainers
  - [x] 10.5 Document table-driven test pattern
  - [x] 10.6 Reference test files: `*_test.go` co-located with source

- [x] Task 11: Add Quick Reference Section
  - [x] 11.1 Create file checklist table (all files to create/modify)
  - [x] 11.2 Add common commands summary
  - [x] 11.3 Add import rules table quick reference

- [x] Task 12: Review and Verify (AC: #1-2)
  - [x] 12.1 Ensure all steps reference existing Users module
  - [x] 12.2 Verify file paths are accurate
  - [x] 12.3 Ensure document is scannable with clear headers
  - [x] 12.4 Add GitHub-style alerts (TIP, IMPORTANT, WARNING, CAUTION)

## Dependencies & Blockers

- **Depends on:** Epic 4 (Reference Implementation - Users Module) - Completed
  - Story 4.1: Implement User Domain Entity ✅
  - Story 4.2: Implement User PostgreSQL Repository ✅
  - Story 4.3: Implement User Use Cases ✅
  - Story 4.5: Implement Transport Contracts/DTOs ✅
  - Story 4.6: Implement User HTTP Handlers ✅
- **Uses:** Existing Users module as reference pattern in `internal/`
- **Uses:** Existing audit events guide in `docs/guides/adding-audit-events.md`

## Assumptions & Open Questions

- Assumes Users module is the canonical reference implementation
- Target audience: developers who want to add new business modules
- Document should be self-contained but reference existing code as examples
- Assumes reader has Go familiarity and understands hexagonal architecture concepts

## Definition of Done

- [x] `docs/guides/adding-module.md` created with all 9 steps
- [x] Each step includes code examples from Users module
- [x] File paths and commands are accurate and verified
- [x] Internal links to reference files work correctly
- [x] Document follows GitHub-style alerts for tips/warnings
- [x] Tests section references project's test design system
- [x] Quick reference checklist included

## Non-Functional Requirements

- Documentation should be scannable with clear headers
- Include actual code snippets from existing Users module
- Add tips/notes for common pitfalls
- Keep document practical and action-oriented
- Use GitHub-style alerts (NOTE, TIP, WARNING, CAUTION, IMPORTANT)
- Include copy-paste ready code templates

## Testing & Verification

### Manual Verification Steps

1. **File Paths:** Verify all referenced file paths exist in codebase
2. **Code Snippets:** Verify code examples match actual implementation
3. **Commands:** Verify `make migrate-create name=...` command works
4. **References:** Ensure all links to Users module files are correct

### Example Verification Commands

```bash
# Verify migration command
make migrate-create name=test_module

# Clean up test migration
rm migrations/*test_module*.sql

# Verify referenced files exist
ls internal/domain/user.go
ls internal/infra/postgres/user_repo.go
ls internal/app/user/create_user.go
ls internal/transport/http/contract/user.go
ls internal/transport/http/handler/user.go
ls internal/transport/http/router.go
```

## Dev Notes

### Users Module Reference Files

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `internal/domain/user.go` | Entity + Repository interface |
| Domain | `internal/domain/errors.go` | Domain errors |
| Infra | `internal/infra/postgres/user_repo.go` | Repository implementation |
| App | `internal/app/user/create_user.go` | Create use case |
| App | `internal/app/user/get_user.go` | Get use case |
| App | `internal/app/user/list_users.go` | List use case |
| Transport | `internal/transport/http/contract/user.go` | DTOs |
| Transport | `internal/transport/http/handler/user.go` | HTTP handlers |
| Transport | `internal/transport/http/router.go` | Route registration |
| Migration | `migrations/20251217000000_create_users.sql` | Database schema |

### Layer Import Rules (Critical)

| Layer | Can Import | CANNOT Import |
|-------|------------|---------------|
| **Domain** | `$gostd` only | `slog`, `uuid`, `pgx`, `otel`, ANY external |
| **App** | `$gostd`, `internal/domain` | `slog`, `otel`, `uuid`, `net/http`, `pgx`, `transport`, `infra` |
| **Transport** | `domain`, `app`, `chi`, `uuid`, `stdlib` | `pgx`, `internal/infra` |
| **Infra** | `domain`, `pgx`, `slog`, `otel`, everything | `app`, `transport` |

### Key Patterns to Document

1. **UUID v7 Handling:**
   - Domain: `type ID string` (NOT `uuid.UUID`)
   - Transport: Generate UUID v7 with `uuid.NewV7()`
   - Infra: Parse `domain.ID` to `uuid.UUID` at boundary

2. **Repository Pattern:**
   - Interface in domain layer
   - Accepts `Querier` parameter (works with pool or tx)
   - Implementation in infra layer

3. **Error Handling:**
   - Domain: Sentinel errors (`var ErrXNotFound = errors.New(...)`)
   - Infra: Wrap with op string (`fmt.Errorf("%s: %w", op, err)`)
   - App: Create `AppError` with `Code`
   - Transport: Map to HTTP status + RFC 7807

4. **Transaction Handling:**
   - Use `TxManager.WithTx()` in use case
   - Pass `tx` (Querier) to all repository calls within transaction

### Recommended Document Structure

```markdown
# Guide: Adding a New Module

## Overview
Brief intro explaining what you'll build

## Step 1: Create Domain Entity & Repository Interface
### 1.1 Define the Entity
### 1.2 Add Validation
### 1.3 Define Repository Interface
### 1.4 Add Domain Errors

## Step 2: Create Database Migration
### 2.1 Generate Migration File
### 2.2 Write Schema

## Step 3: Implement Repository
### 3.1 Create Repository Struct
### 3.2 Implement Interface Methods
### 3.3 Add Compile-Time Check

## Step 4: Create Use Cases
### 4.1 Create Use Case Package
### 4.2 Implement Create Use Case
### 4.3 Implement Get Use Case
### 4.4 Implement List Use Case

## Step 5: Create DTOs/Contracts
### 5.1 Request Structs
### 5.2 Response Structs
### 5.3 Mapper Functions

## Step 6: Create HTTP Handlers
### 6.1 Handler Struct
### 6.2 Implement Handler Methods

## Step 7: Wire Routes
### 7.1 Define Routes Interface
### 7.2 Update Router

## Step 8: Add Audit Events
### 8.1 Define Event Type
### 8.2 Record Events

## Step 9: Write Tests
### 9.1 Domain Tests
### 9.2 Use Case Tests
### 9.3 Handler Tests
### 9.4 Repository Integration Tests

## Quick Reference Checklist
```

### References

- [Source: docs/epics.md#Story 8.5] Lines 1738-1762
- [Source: docs/architecture.md#Hexagonal Architecture Developer Guide] Lines 58-105
- [Source: docs/project-context.md#Critical Layer Rules] Lines 35-89
- [Source: docs/project-context.md#UUID v7 Handling] Lines 92-124
- [Source: docs/project-context.md#Naming Conventions] Lines 128-183
- [Source: docs/project-context.md#Error Handling] Lines 186-212
- [Source: internal/domain/user.go] Domain entity example
- [Source: internal/infra/postgres/user_repo.go] Repository implementation example
- [Source: internal/app/user/create_user.go] Use case example
- [Source: internal/transport/http/contract/user.go] DTO example
- [Source: internal/transport/http/handler/user.go] Handler example
- [Source: internal/transport/http/router.go] Router wiring example
- [Source: migrations/20251217000000_create_users.sql] Migration example
- [Source: FR68] Documentation includes step-by-step guide for adding new modules

### Epic 8 Context

Epic 8 implements Documentation & Developer Guides:
- **8.1:** README Quick Start ✅ (done)
- **8.2:** Architecture and Layer Responsibilities ✅ (done)
- **8.3:** Local Development Workflow ✅ (done)
- **8.4:** Observability Configuration ✅ (done)
- **8.5 (this story):** Guide for Adding New Modules ← current
- **8.6:** Guide for Adding New Adapters (backlog)

### Previous Story Learnings (8.4)

From Story 8.4 implementation:
- Use GitHub-style alerts (NOTE, TIP, WARNING, CAUTION, IMPORTANT) throughout
- Include copy-paste ready commands and configurations
- Verify all documented commands work before completing
- Use tables for quick reference
- Reference existing code files as examples

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-23)

Files analyzed:
- `docs/epics.md` - Story 8.5 acceptance criteria (lines 1738-1762)
- `internal/domain/user.go` - Domain entity example (53 lines)
- `internal/infra/postgres/user_repo.go` - Repository implementation (156 lines)
- `internal/app/user/create_user.go` - Use case example (137 lines)
- `internal/transport/http/contract/user.go` - DTO example (83 lines)
- `internal/transport/http/handler/user.go` - Handler example (173 lines)
- `internal/transport/http/router.go` - Router wiring (121 lines)
- `migrations/20251217000000_create_users.sql` - Migration example (20 lines)
- `docs/project-context.md` - Layer rules and conventions
- `docs/architecture.md` - Hexagonal architecture guide
- `docs/sprint-artifacts/8-4-document-observability-configuration.md` - Previous story

### Agent Model Used

Google Gemini (Antigravity)

### Debug Log References

N/A

### Completion Notes List

✅ **Implementation completed successfully (2025-12-23)**

1. Created comprehensive `docs/guides/adding-module.md` with all 9 steps
2. Added table of contents for easy navigation
3. Each step includes code examples from Users module reference files
4. Used GitHub-style alerts (CAUTION, IMPORTANT, TIP, WARNING, NOTE) throughout
5. Added Quick Reference Checklist section with:
   - Files to create/modify table
   - Import rules quick reference table
   - Common commands summary
6. Verified all referenced file paths exist in codebase
7. Tested `make migrate-create name=test_module` command - works correctly
8. Cleaned up test migration file after verification

### File List

**Created:**
- `docs/guides/adding-module.md` - Comprehensive step-by-step guide for adding new modules

**Referenced (verified existing):**
- `internal/domain/user.go`
- `internal/domain/errors.go`
- `internal/infra/postgres/user_repo.go`
- `internal/app/user/create_user.go`
- `internal/transport/http/contract/user.go`
- `internal/transport/http/handler/user.go`
- `internal/transport/http/router.go`
- `migrations/20251217000000_create_users.sql`
- `docs/guides/adding-audit-events.md`

### Change Log

| Date | Change |
|------|--------|
| 2025-12-23 | Story 8.5 drafted by create-story workflow |
| 2025-12-23 | Created `docs/guides/adding-module.md` with all 9 steps, code examples, and quick reference |
| 2025-12-23 | Verified all file paths and migration command |
| 2025-12-23 | Story completed - Ready for Review |
