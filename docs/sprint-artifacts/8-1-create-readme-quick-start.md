# Story 8.1: Create README Quick Start

Status: Ready for Review

## Story

As a **developer**,
I want **a README with quick start instructions**,
so that **I can get the service running without external help**.

## Acceptance Criteria

1. **Given** I clone the repository, **When** I read `README.md`, **Then** I see clear quick start section with commands:
   1. Prerequisites (Go 1.24+, Docker)
   2. `make setup`
   3. `make infra-up`
   4. `make migrate-up`
   5. `make run`
   6. Verify endpoints:
      - `curl http://localhost:8080/health` → 200
      - `curl http://localhost:8080/ready` → 200 (DB connected)
      - `curl http://localhost:8080/metrics` → Prometheus format

2. **And** optional "First API Call" section:
   - `POST /api/v1/users` with sample body
   - `GET /api/v1/users/{id}`
   - `GET /api/v1/users`

3. **And** all commands work without modification

4. **And** estimated time is mentioned (< 15 minutes)

*Covers: FR64*

## Current State Analysis

The README.md already exists with substantial quick start content. This story requires **enhancement** rather than creation from scratch.

### Existing Coverage (✅)

- Prerequisites (Go 1.23+ - needs update to 1.24+, Docker)
- `make setup` command
- `make infra-up` command  
- `make migrate-up` command
- `make run` command
- Health endpoint verification (`curl http://localhost:8080/health`)
- Ready endpoint verification (`curl http://localhost:8080/ready`)

### Gaps to Address (⚠️)

1. **Go version:** States 1.23+ but should be 1.24+ per architecture.md
2. **Metrics endpoint:** Not shown in quick start verification
3. **First API Call section:** Users API examples not documented
4. **Estimated time:** Not mentioned in README

## Tasks / Subtasks

- [x] Task 1: Update Prerequisites (AC: #1)
  - [x] 1.1 Change Go version from 1.23+ to 1.24+
  - [x] 1.2 Verify alignment with `go.mod` toolchain version (go 1.24.11)

- [x] Task 2: Enhance Quick Start verification (AC: #1.6)
  - [x] 2.1 Add `/metrics` endpoint to verification steps
  - [x] 2.2 Show expected output format for each endpoint

- [x] Task 3: Add First API Call section (AC: #2)
  - [x] 3.1 Add POST /api/v1/users example with sample JSON body
  - [x] 3.2 Add GET /api/v1/users/{id} example
  - [x] 3.3 Add GET /api/v1/users example (list)
  - [x] 3.4 Show expected response format

- [x] Task 4: Add estimated time (AC: #4)
  - [x] 4.1 Add "Setup time: ~15 minutes" to quick start section

- [x] Task 5: Verify commands work (AC: #3)
  - [x] 5.1 Verified quick start sequence commands are copy-paste ready
  - [x] 5.2 Verified curl commands match actual API response format

## Dependencies & Blockers

- **Depends on:** Epic 4 (Users Module) - Completed
- **Depends on:** Epic 1 (Project Foundation) - Completed
- **Uses:** Existing README.md structure

## Assumptions & Open Questions

- Users module is fully functional with create/get/list endpoints
- JWT authentication may be required for users endpoints (need to check)
- Sample request body should match DTOs in `internal/transport/http/contract/user.go`

## Definition of Done

- [x] README.md prerequisites shows Go 1.24+
- [x] Quick start includes metrics endpoint verification
- [x] First API Call section with working curl examples
- [x] Estimated time mentioned (~15 minutes)
- [x] All commands tested and verified working
- [x] Documentation reviewed for accuracy

## Non-Functional Requirements

- Documentation should be concise and scannable
- Commands should be copy-paste ready
- Examples should use realistic but anonymized data
- Follow markdown best practices

## Testing & Verification

### Manual Verification Steps

1. **Fresh clone test:**
   ```bash
   git clone https://github.com/iruldev/golang-api-hexagonal.git /tmp/test-clone
   cd /tmp/test-clone
   # Follow README quick start exactly
   # Time the process - should be < 15 minutes
   ```

2. **Health endpoints:**
   ```bash
   curl http://localhost:8080/health
   # Expected: {"data":{"status":"ok"}}

   curl http://localhost:8080/ready
   # Expected: {"data":{"status":"ready","checks":{"database":"ok"}}}

   curl http://localhost:8080/metrics
   # Expected: Prometheus format text output
   ```

3. **Users API (if no auth required):**
   ```bash
   # Create user
   curl -X POST http://localhost:8080/api/v1/users \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","firstName":"John","lastName":"Doe"}'
   
   # Get user
   curl http://localhost:8080/api/v1/users/{id}
   
   # List users
   curl http://localhost:8080/api/v1/users
   ```

## Dev Notes

### README Structure Reference

Current README.md sections:
1. Quick Start (🚀)
2. Requirements (📋)
3. Make Targets (🛠️)
4. Architecture (🏗️)
5. Configuration (🔧)
6. API Endpoints (📡)
7. Database Migrations (🗄️)
8. Testing (🧪)
9. Project Structure (📁)
10. License (📝)

### Users Module Context

From Epic 4 implementation:
- Domain: `internal/domain/user.go`
- Use cases: `internal/app/user/`
- Handlers: `internal/transport/http/handler/user.go`
- DTOs: `internal/transport/http/contract/user.go`

### JWT Authentication Note

Check if users endpoints require JWT authentication. If so, First API Call section needs to:
1. Explain token requirement
2. Provide example or mock token for testing
3. Or suggest running in development mode without auth

### References

- [Source: docs/epics.md#Story 8.1] Lines 1631-1661
- [Source: README.md] Current implementation
- [Source: docs/project-context.md] Technology stack (Go 1.24+)
- [Source: docs/architecture.md] API design standards

### Epic 8 Context

Epic 8 implements Documentation & Developer Guides:
- **8.1 (this story):** README Quick Start
- **8.2:** Architecture and Layer Responsibilities
- **8.3:** Local Development Workflow
- **8.4:** Observability Configuration
- **8.5:** Guide for Adding New Modules
- **8.6:** Guide for Adding New Adapters

## Dev Agent Record

### Context Reference

Story context created by: create-story workflow (2025-12-22)

- `docs/epics.md` - Story 8.1 acceptance criteria
- `README.md` - Current documentation
- `docs/project-context.md` - Technology versions
- `docs/architecture.md` - API patterns

### Agent Model Used

Claude (Anthropic) - Gemini Antigravity Agent

### Debug Log References

N/A

### Completion Notes List

- ✅ Updated Go version from 1.23+ to 1.24+ (aligned with go.mod toolchain 1.24.11)
- ✅ Added estimated time (⏱️ ~15 minutes) prominently at start of Quick Start
- ✅ Added /metrics endpoint to verification steps with expected output
- ✅ Added expected output format for /health and /ready endpoints
- ✅ Created "First API Call" section with Users API examples (POST, GET by ID, GET list)
- ✅ All curl commands are copy-paste ready with proper JSON format
- ✅ Response format matches actual API (email, firstName, lastName camelCase)
- Note: JWT authentication is optional and disabled by default in dev mode

### Change Log

- 2025-12-22: Story file created (ready-for-dev)
- 2025-12-23: Implementation completed (Ready for Review)
  - Updated README.md with all acceptance criteria
  - Go version updated to 1.24+
  - Added metrics endpoint verification
  - Added First API Call section
  - Added estimated time

### File List

**Files modified:**
- `README.md` - Enhanced Quick Start section per acceptance criteria
