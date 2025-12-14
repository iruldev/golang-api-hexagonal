# Story 12.4: Add GraphQL Playground

Status: done

## Story

As a developer,
I want a GraphQL Playground interface,
So that I can interactively explore the schema and test queries.

## Acceptance Criteria

1. **Given** the application is running in development mode (`APP_ENV=development` or `APP_ENV=local`)
   **When** I navigate to `/playground`
   **Then** the GraphQL Playground UI is loaded
   **And** it is connected to the GraphQL endpoint (`/query`)

2. **Given** the application is running in production mode (`APP_ENV=production` or `APP_ENV=staging`)
   **When** I navigate to `/playground`
   **Then** the endpoint returns 404 (disabled for security)

3. **Given** the playground is active
   **When** I execute a query or mutation
   **Then** it successfully interacts with the backend

## Tasks / Subtasks

- [x] Task 1: Add Playground Handler Dependency
  - [x] Ensure `github.com/99designs/gqlgen/graphql/playground` is available (usually part of gqlgen lib)

- [x] Task 2: Register Playground Handler
  - [x] Modify `cmd/server/main.go` to initialize playground handler
  - [x] Mount handler at `/playground`
  - [x] Ensure it points to `/query` endpoint

- [x] Task 3: Implement Environment Restriction
  - [x] Check `cfg.App.IsDevelopment()` before registering the route
  - [x] Only register if `APP_ENV` is `development` or `local`
  - [x] Log whether playground is enabled or disabled at startup

- [x] Task 4: Documentation
  - [x] Update `README.md` with playground URL info (both dev and local modes)
  - [x] Update `AGENTS.md` if necessary (minor)

- [x] Task 5: Testing (Added by Code Review)
  - [x] Add integration tests for environment restriction (playground_test.go)
  - [x] Test playground enabled in development and local modes
  - [x] Test playground returns 404 in staging and production modes
  - [x] Test playground connectivity to /query endpoint

## Dev Notes

### Architecture Compliance

- **Location:** The playground handler is an HTTP interface concern. It should be wired in `cmd/server/main.go` alongside the GraphQL handler.
- **Environment Safety:** CRITICAL. Playground provides introspection and easy traversal of the graph. It MUST NOT be exposed in production. Use `cfg.App.IsDevelopment()` helper to gate it.

### Routing

- Currently, `/query` is registered directly on the router in `cmd/server/main.go`.
- Register `/playground` in the same block, wrapped in `if cfg.App.IsDevelopment() { ... }`.

### Library

- Use `playground.Handler` from `github.com/99designs/gqlgen/graphql/playground`.
- `playground.Handler("GraphQL playground", "/query")`

### Best Practices (Code Review Notes)

- Use `config.EnvDevelopment`, `config.EnvLocal`, etc. constants instead of hardcoded strings.
- Use `cfg.App.IsDevelopment()` helper method for consistent environment checks.
- Tests verify behavior for all known environments (development, local, staging, production).

### References

- [Source: docs/epics.md#Story-12.4]
- [GQLGen Playground Docs](https://gqlgen.com/getting-started/)

## Dev Agent Record

### Context Reference

<!-- Story context created by create-story workflow -->

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- **Task 1:** Verified `gqlgen v0.17.84` is already a dependency - playground package available at `github.com/99designs/gqlgen/graphql/playground`.
- **Task 2:** Added playground import and handler registration in `cmd/server/main.go`. Handler registered at `/playground` pointing to `/query`.
- **Task 3:** Implemented environment restriction - playground only enabled when `cfg.App.IsDevelopment()` returns true. Logs startup status for both enabled and disabled cases.
- **Task 4:** Updated `README.md` with GraphQL Playground section under Quick Start (mentions both dev and local). Updated `AGENTS.md` with GraphQL Playground documentation under GraphQL Server Patterns section.
- **Task 5 (Code Review):** Added integration tests in `playground_test.go`. Tests verify playground is enabled in dev/local and returns 404 in staging/production. Added `IsDevelopment()` helper test.

### File List

- `cmd/server/main.go` (modified) - Added playground import and handler registration, refactored to use `IsDevelopment()` helper
- `internal/config/config.go` (modified) - Added environment constants and `IsDevelopment()` helper method
- `internal/interface/graphql/playground_test.go` (new) - Integration tests for environment restriction
- `README.md` (modified) - Added GraphQL Playground section mentioning both dev and local modes
- `AGENTS.md` (modified) - Added GraphQL Playground documentation section

### Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Implemented GraphQL Playground with environment restriction (Story 12.4) |
| 2025-12-14 | Code Review: Added environment constants, IsDevelopment() helper, integration tests, updated docs |

