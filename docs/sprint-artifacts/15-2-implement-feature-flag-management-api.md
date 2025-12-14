# Story 15.2: Implement Feature Flag Management API

Status: Done

## Story

As a product manager,
I want to toggle feature flags via API,
so that features can be enabled/disabled without deployment.

## Acceptance Criteria

1. **Given** `GET /admin/features` endpoint
   **When** called by admin
   **Then** list of all configured feature flags is returned
   **And** each flag includes name, current state, and description

2. **Given** `GET /admin/features/{flag}` endpoint
   **When** called by admin with valid flag name
   **Then** flag details are returned (name, enabled, description, updated_at)
   **And** 404 is returned for unknown flags

3. **Given** `POST /admin/features/{flag}/enable` endpoint
   **When** called by admin
   **Then** flag state is updated to enabled
   **And** change is effective immediately for subsequent requests
   **And** 201 Created response with updated flag state is returned

4. **Given** `POST /admin/features/{flag}/disable` endpoint
   **When** called by admin
   **Then** flag state is updated to disabled
   **And** 201 Created response with updated flag state is returned

5. **Given** feature flag state changes
   **When** any enable/disable operation is performed
   **Then** audit log is emitted with actor, action, and timestamp

6. **Given** all feature flag admin endpoints
   **When** accessed without admin role
   **Then** 403 Forbidden is returned (existing RBAC from Story 15.1)

## Tasks / Subtasks

- [x] Create Persistent Feature Flag Provider
  - [x] Define `AdminFeatureFlagProvider` interface extending `FeatureFlagProvider` with write methods
  - [x] Add `SetEnabled(ctx, flag, enabled) error` method
  - [x] Add `List(ctx) ([]FeatureFlagState, error)` method
  - [x] Add `Get(ctx, flag) (FeatureFlagState, error)` method
  - [x] Create `FeatureFlagState` struct with Name, Enabled, Description, UpdatedAt fields
- [x] Implement In-Memory State Store
  - [x] Create `InMemoryFeatureFlagStore` implementing `AdminFeatureFlagProvider`
  - [x] Initialize from environment variables (fall back to existing EnvProvider behavior)
  - [x] Support dynamic state updates (in-memory only initially)
  - [x] Add thread-safety with `sync.RWMutex`
- [x] Create Feature Flag Admin Handler
  - [x] Create `internal/interface/http/admin/features.go`
  - [x] Implement `GET /features` - ListFlagsHandler
  - [x] Implement `GET /features/{flag}` - GetFlagHandler
  - [x] Implement `POST /features/{flag}/enable` - EnableFlagHandler
  - [x] Implement `POST /features/{flag}/disable` - DisableFlagHandler
  - [x] Inject `AdminFeatureFlagProvider` via dependency
- [x] Register Routes in Admin Router
  - [x] Add routes to `RegisterAdminRoutes(r)` in `routes_admin.go`
  - [x] Pass feature flag provider via RouterDeps
- [x] Add Audit Logging for State Changes
  - [x] Use `observability.LogAudit` to log enable/disable actions
  - [x] Include actor (from claims), action, flag name, new state, timestamp
- [x] Write Unit Tests
  - [x] Test ListFlagsHandler returns all flags
  - [x] Test GetFlagHandler returns single flag details
  - [x] Test GetFlagHandler returns 404 for unknown flag
  - [x] Test EnableFlagHandler updates state to enabled
  - [x] Test DisableFlagHandler updates state to disabled
  - [x] Test concurrent access thread safety
  - [x] Test 403 when non-admin accesses endpoints (covered by existing admin tests)
- [x] Documentation
  - [x] Add Feature Flag API section to AGENTS.md
  - [x] Document Admin Feature Flag Management pattern

## Dev Agent Record

### Context Reference

- `docs/epics.md` - Requirements source (Epic 15)
- `internal/runtimeutil/featureflags.go` - Existing feature flag interface
- `internal/interface/http/admin/handler.go` - Existing admin handler pattern
- `internal/interface/http/routes_admin.go` - Admin route registration
- `internal/observability/audit.go` - Audit logger

### Agent Model Used

Gemini 2.5 Pro

### Debug Log References

### Completion Notes List

- Created `AdminFeatureFlagProvider` interface extending `FeatureFlagProvider` with SetEnabled, List, Get methods
- Created `FeatureFlagState` struct with Name, Enabled, Description, UpdatedAt fields
- Implemented `InMemoryFeatureFlagStore` with thread-safety via sync.RWMutex
- Created `FeaturesHandler` with ListFlags, GetFlag, EnableFlag, DisableFlag handlers
- Added audit logging for enable/disable actions using `observability.LogAudit`
- Created `AdminDeps` struct for dependency injection in admin routes
- Updated `RouterDeps` with `AdminFeatureFlagProvider` field
- Added 15 unit tests for `InMemoryFeatureFlagStore` (including race condition testing)
- Added 8 unit tests for `FeaturesHandler`
- Updated AGENTS.md with Admin Feature Flag Management API section
- All 30+ test packages pass with no regressions

### File List

**New Files:**
- `internal/interface/http/admin/features.go` - FeaturesHandler with 4 endpoints
- `internal/interface/http/admin/features_test.go` - 10 handler tests (including RBAC)

**Modified Files:**
- `internal/runtimeutil/featureflags.go` - Added AdminFeatureFlagProvider, FeatureFlagState, InMemoryFeatureFlagStore with env fallback
- `internal/runtimeutil/featureflags_test.go` - Added 17 tests for InMemoryFeatureFlagStore including Get() env fallback
- `internal/interface/http/routes_admin.go` - Added AdminDeps struct, feature flag routes
- `internal/interface/http/router.go` - Added AdminFeatureFlagProvider to RouterDeps
- `AGENTS.md` - Added Admin Feature Flag Management API section

### Change Log

| Date | Change |
|------|--------|
| 2025-12-14 | Implemented Feature Flag Management API (Story 15.2) |
| 2025-12-14 | [Code Review] Fixed Get() to fall back to env provider for consistency with IsEnabled() |
| 2025-12-14 | [Code Review] Added RBAC 403 test for all feature flag endpoints |


