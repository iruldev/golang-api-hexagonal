# Story 8.6: Create Guide for Adding New Adapters

Status: Done

## Story

As a **developer**,
I want **guide for adding new infrastructure adapters**,
so that **I can integrate external services correctly**.

## Acceptance Criteria

1. **Given** I want to add a new adapter (e.g., Redis cache, email service), **When** I read `docs/guides/adding-adapter.md`, **Then** I see:
   - Where adapters live (`internal/infra/`)
   - Interface definition in domain or app layer
   - Implementation in infra layer
   - Configuration via environment variables
   - Testing strategy (mocks, testcontainers)
   - Wiring in main.go

*Covers: FR69*

## Tasks / Subtasks

- [x] Task 1: Create `docs/guides/adding-adapter.md` Document Structure (AC: #1)
  - [x] 1.1 Create document with clear section headers
  - [x] 1.2 Add table of contents for easy navigation
  - [x] 1.3 Add overview explaining adapters in hexagonal architecture

- [x] Task 2: Document Adapter Types and Location (AC: #1)
  - [x] 2.1 Explain the difference between driving and driven adapters
  - [x] 2.2 Document `internal/infra/` directory structure
  - [x] 2.3 List existing adapters as examples:
    - `internal/infra/postgres/` - PostgreSQL repository adapters
    - `internal/infra/config/` - Configuration adapter
    - `internal/infra/observability/` - Logging, tracing, metrics adapters
  - [x] 2.4 Provide guidance on creating new adapter directories (e.g., `internal/infra/redis/`, `internal/infra/email/`)

- [x] Task 3: Document Interface Definition Pattern (AC: #1)
  - [x] 3.1 Explain port (interface) definition in domain layer for data-related adapters
  - [x] 3.2 Explain service interface definition in app layer for auxiliary services
  - [x] 3.3 Show example interface with `context.Context` parameter
  - [x] 3.4 Reference `internal/domain/user.go` (UserRepository interface) as example
  - [x] 3.5 Reference `internal/domain/querier.go` (Querier, TxManager interfaces) as example

- [x] Task 4: Document Adapter Implementation (AC: #1)
  - [x] 4.1 Create adapter struct implementing the interface
  - [x] 4.2 Document constructor pattern `NewXAdapter(deps...) *XAdapter`
  - [x] 4.3 Document error wrapping pattern with `op` string
  - [x] 4.4 Document compile-time interface check `var _ domain.XInterface = (*XAdapter)(nil)`
  - [x] 4.5 Reference `internal/infra/postgres/user_repo.go` implementation as example
  - [x] 4.6 Reference `internal/infra/observability/logger.go` as non-repository adapter example

- [x] Task 5: Document Configuration via Environment Variables (AC: #1)
  - [x] 5.1 Show how to add new configuration fields to `internal/infra/config/config.go`
  - [x] 5.2 Document envconfig struct tags (`envconfig`, `required`, `default`)
  - [x] 5.3 Document validation in `Config.Validate()` method
  - [x] 5.4 Document `.env.example` update requirements
  - [x] 5.5 Reference `internal/infra/config/config.go` as example
  - [x] 5.6 Show examples for different adapter types:
    - Redis: `REDIS_URL`, `REDIS_PASSWORD`, `REDIS_DB`
    - Email: `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`

- [x] Task 6: Document Testing Strategy (AC: #1)
  - [x] 6.1 Document mock creation for unit tests
  - [x] 6.2 Document testify mock generation or manual mock implementation
  - [x] 6.3 Document testcontainers-go for integration tests
  - [x] 6.4 Provide example test file structure
  - [x] 6.5 Reference `internal/infra/postgres/user_repo_test.go` as integration test example

- [x] Task 7: Document Wiring in main.go (AC: #1)
  - [x] 7.1 Document dependency injection pattern
  - [x] 7.2 Show initialization order (config → adapters → use cases → handlers)
  - [x] 7.3 Document graceful shutdown for adapters with resources
  - [x] 7.4 Reference `cmd/api/main.go` as example

- [x] Task 8: Add Example Adapter Templates (AC: #1)
  - [x] 8.1 Create Redis cache adapter template (conceptual)
  - [x] 8.2 Create email service adapter template (conceptual)
  - [x] 8.3 Include step-by-step walkthrough for each example

- [x] Task 9: Add Quick Reference Section
  - [x] 9.1 Create adapter checklist table (files to create/modify)
  - [x] 9.2 Add import rules summary for infra layer
  - [x] 9.3 Add common patterns and pitfalls

- [x] Task 10: Review and Verify (AC: #1)
  - [x] 10.1 Ensure all examples reference existing codebase files
  - [x] 10.2 Verify file paths are accurate
  - [x] 10.3 Ensure document is scannable with clear headers
  - [x] 10.4 Add GitHub-style alerts (TIP, IMPORTANT, WARNING, CAUTION)

## Dependencies & Blockers

- **Depends on:** Epic 4 (Reference Implementation - Users Module) - Completed
  - Story 4.2: Implement User PostgreSQL Repository ✅
- **Uses:** Existing PostgreSQL adapter as reference pattern in `internal/infra/postgres/`
- **Uses:** Existing config adapter in `internal/infra/config/`
- **Uses:** Existing observability adapters in `internal/infra/observability/`
- **Related to:** Story 8.5 (Adding New Modules) - provides complementary guidance

## Assumptions & Open Questions

- Assumes PostgreSQL repository is the canonical reference for data adapters
- Assumes Logger/Tracer are the canonical reference for non-repository adapters
- Target audience: developers who want to integrate external services
- Document should be self-contained but reference existing code as examples
- Assumes reader understands hexagonal architecture concepts (references Story 8.2)

## Definition of Done

- [x] `docs/guides/adding-adapter.md` created with all required sections
- [x] Each section includes code examples from existing adapters
- [x] File paths and commands are accurate and verified
- [x] Internal links to reference files work correctly
- [x] Document follows GitHub-style alerts for tips/warnings
- [x] Configuration section shows proper envconfig patterns
- [x] Testing section covers both mocking and testcontainers approaches
- [x] Quick reference checklist included

## Non-Functional Requirements

- Documentation should be scannable with clear headers
- Include actual code snippets from existing adapters
- Add tips/notes for common pitfalls
- Keep document practical and action-oriented
- Use GitHub-style alerts (NOTE, TIP, WARNING, CAUTION, IMPORTANT)
- Include copy-paste ready code templates

## Testing & Verification

### Manual Verification Steps

1. **File Paths:** Verify all referenced file paths exist in codebase
2. **Code Snippets:** Verify code examples match actual implementation
3. **References:** Ensure all links to adapter files are correct

### Example Verification Commands

```bash
# Verify referenced adapter files exist
ls internal/infra/postgres/user_repo.go
ls internal/infra/postgres/querier.go
ls internal/infra/postgres/tx_manager.go
ls internal/infra/config/config.go
ls internal/infra/observability/logger.go
ls internal/infra/observability/tracer.go
ls internal/infra/observability/metrics.go

# Verify domain interface files
ls internal/domain/user.go
ls internal/domain/querier.go

# Verify main.go wiring
ls cmd/api/main.go
```

## Dev Notes

### Existing Adapter Reference Files

| Layer | File | Purpose |
|-------|------|---------|
| Domain | `internal/domain/user.go` | Repository interface (port) definition |
| Domain | `internal/domain/querier.go` | Querier, TxManager interfaces |
| Infra | `internal/infra/postgres/user_repo.go` | PostgreSQL repository adapter |
| Infra | `internal/infra/postgres/querier.go` | Querier implementation |
| Infra | `internal/infra/postgres/tx_manager.go` | TxManager implementation |
| Infra | `internal/infra/postgres/pool.go` | PostgreSQL connection pool |
| Infra | `internal/infra/config/config.go` | Configuration adapter |
| Infra | `internal/infra/observability/logger.go` | Logging adapter |
| Infra | `internal/infra/observability/tracer.go` | Tracing adapter |
| Infra | `internal/infra/observability/metrics.go` | Metrics adapter |

### Layer Import Rules (Critical for Infra)

| Layer | Can Import | CANNOT Import |
|-------|------------|---------------|
| **Domain** | `$gostd` only | `slog`, `uuid`, `pgx`, `otel`, ANY external |
| **App** | `$gostd`, `internal/domain` | `slog`, `otel`, `uuid`, `net/http`, `pgx`, `transport`, `infra` |
| **Transport** | `domain`, `app`, `chi`, `uuid`, `stdlib` | `pgx`, `internal/infra` |
| **Infra** | `domain`, `pgx`, `slog`, `otel`, everything | `app`, `transport` |

### Key Patterns to Document

1. **Port vs Adapter:**
   - Port = Interface defined in domain (or app for auxiliary services)
   - Adapter = Implementation in infra layer

2. **Dependency Injection:**
   - Adapters receive dependencies via constructor
   - No global state
   - Facilitates testing with mocks

3. **Error Handling in Infra:**
   - Wrap errors with `op` string: `fmt.Errorf("%s: %w", op, err)`
   - Map infrastructure-specific errors to domain errors where appropriate

4. **Resource Management:**
   - Implement proper cleanup in adapters with resources (DB pools, connections)
   - Support graceful shutdown in main.go

### Conceptual Adapter Examples

**Redis Cache Adapter:**
```
internal/infra/redis/
├── cache.go          # Cache interface implementation
├── cache_test.go     # Integration tests with testcontainers
└── client.go         # Redis client initialization
```

**Email Service Adapter:**
```
internal/infra/email/
├── sender.go         # EmailSender interface implementation
├── sender_test.go    # Unit tests with mock SMTP
└── smtp.go           # SMTP client configuration
```

### References

- [Source: docs/epics.md#Story 8.6] Lines 1765-1784
- [Source: docs/architecture.md#Hexagonal Architecture Developer Guide] Lines 58-105
- [Source: docs/architecture.md#Four Layers and Responsibilities] Lines 108-462
- [Source: docs/project-context.md#Critical Layer Rules] Lines 35-89
- [Source: docs/project-context.md#Error Handling] Lines 186-212
- [Source: internal/domain/user.go] Repository interface example
- [Source: internal/domain/querier.go] Querier/TxManager interface example
- [Source: internal/infra/postgres/user_repo.go] Repository adapter implementation
- [Source: internal/infra/config/config.go] Configuration adapter example
- [Source: internal/infra/observability/logger.go] Non-repository adapter example
- [Source: FR69] Documentation includes step-by-step guide for adding new adapters

### Epic 8 Context

Epic 8 implements Documentation & Developer Guides:
- **8.1:** README Quick Start ✅ (done)
- **8.2:** Architecture and Layer Responsibilities ✅ (done)
- **8.3:** Local Development Workflow ✅ (done)
- **8.4:** Observability Configuration ✅ (done)
- **8.5:** Guide for Adding New Modules ✅ (done)
- **8.6 (this story):** Guide for Adding New Adapters ✅ (Ready for Review)

### Previous Story Learnings (8.5)

From Story 8.5 implementation:
- Use GitHub-style alerts (NOTE, TIP, WARNING, CAUTION, IMPORTANT) throughout
- Include copy-paste ready commands and configurations
- Verify all documented file paths are accurate
- Use tables for quick reference
- Reference existing code files as examples
- Each step should be actionable and concrete

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-23)

Files analyzed:
- `docs/epics.md` - Story 8.6 acceptance criteria (lines 1765-1784)
- `internal/infra/postgres/user_repo.go` - Repository adapter example (156 lines)
- `internal/infra/postgres/querier.go` - Querier implementation
- `internal/infra/postgres/tx_manager.go` - TxManager implementation
- `internal/infra/config/config.go` - Configuration adapter (155 lines)
- `internal/infra/observability/logger.go` - Logging adapter (65 lines)
- `internal/infra/observability/tracer.go` - Tracing adapter
- `internal/infra/observability/metrics.go` - Metrics adapter
- `internal/domain/user.go` - Repository interface definition
- `internal/domain/querier.go` - Querier/TxManager interfaces
- `docs/project-context.md` - Layer rules and conventions
- `docs/architecture.md` - Hexagonal architecture guide
- `docs/sprint-artifacts/8-5-create-guide-for-adding-new-modules.md` - Previous story

### Agent Model Used

Google Gemini (Antigravity)

### Debug Log References

N/A

### Completion Notes List

- ✅ Created comprehensive `docs/guides/adding-adapter.md` (approximately 700 lines)
- ✅ Document covers all required sections per AC #1:
  - Where adapters live (`internal/infra/`)
  - Interface definition in domain layer (ports)
  - Implementation in infra layer (adapters)
  - Configuration via environment variables (envconfig patterns)
  - Testing strategy (mocks and testcontainers-go)
  - Wiring in main.go with initialization order and graceful shutdown
- ✅ Included conceptual Redis cache adapter example with complete code
- ✅ Included conceptual email service (SMTP) adapter example with complete code
- ✅ Added quick reference checklist table
- ✅ Added common pitfalls and warnings
- ✅ Used GitHub-style alerts throughout (TIP, IMPORTANT, WARNING, CAUTION)
- ✅ All referenced file paths verified to exist in codebase
- ✅ Document follows scannable structure with table of contents

### File List

| Action | File |
|--------|------|
| Created | `docs/guides/adding-adapter.md` |
| Modified | `docs/sprint-artifacts/sprint-status.yaml` |
| Modified | `docs/sprint-artifacts/8-6-create-guide-for-adding-new-adapters.md` |

### Change Log

| Date | Change |
|------|--------|
| 2025-12-23 | Story 8.6 drafted by create-story workflow |
| 2025-12-23 | Implemented Story 8.6: Created comprehensive guide for adding new adapters |
