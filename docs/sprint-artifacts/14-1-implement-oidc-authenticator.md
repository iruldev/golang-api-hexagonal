# Story 14.1: Implement OIDC Authenticator

Status: Done

## Story

As a security engineer,
I want to authenticate users via OIDC (e.g., Auth0, Keycloak),
So that we use centralized identity management.

## Acceptance Criteria

1. **Given** `OIDCAuthenticator` configured with issuer URL
   **When** JWT from provider is received
   **Then** signature is validated using JWKS
   **And** claims are mapped to internal context

## Tasks / Subtasks

- [x] Add OIDC dependencies
  - [x] Add `github.com/coreos/go-oidc/v3`
  - [x] Add `golang.org/x/oauth2`
- [x] Update Configuration
  - [x] Add `OIDC` section to `internal/config/config.go` (IssuerURL, ClientID, Audience)
  - [x] Add validation for OIDC config (Fail fast if IssuerURL is missing when OIDC is enabled)
- [x] Implement OIDC Authenticator
  - [x] Create `internal/interface/http/middleware/oidc.go`
  - [x] Define `OIDCAuthenticator` struct
  - [x] Implement `Authenticate(r *http.Request) (Claims, error)` method
  - [x] Implement ID Token verification using `go-oidc` KeySet
  - [x] Map OIDC Claims (sub, groups/roles) to internal `middleware.Claims`
- [x] Helper for Claims Mapping
  - [x] Create flexible mapping logic (e.g., map `https://example.com/roles` to `Roles`)
- [x] Update Main wiring
  - [x] In `cmd/api/main.go`, initialize `OIDCAuthenticator` if configured
  - [x] Pass to `AuthMiddleware`
- [x] Tests
  - [x] Create unit tests for `OIDCAuthenticator` (mocking the verifier or using test keys)
  - [x] Verify claims mapping logic

## Dev Notes

### Architecture Integration
- **Interface**: Must implement `Authenticator` interface defined in `internal/interface/http/middleware/auth.go`.
- **Location**: Implementation should be in `internal/interface/http/middleware/oidc.go` (or `oidc_authenticator.go`).
- **Configuration**: Use `internal/config` to hold OIDC settings (`Issuer`, `ClientID`).
- **Context**: The `Authenticate` method returns `Claims`, which the `AuthMiddleware` then puts into `context`.

### Technical Specifics
- **Library**: Use `github.com/coreos/go-oidc/v3/oidc`.
- **Verifier**: Use `provider.Verifier(oidc.Config{ClientID: ...})`.
- **Token Extraction**: The `Authenticator` interface expects the implementation to extract the token (usually from `Authorization: Bearer ...`).
- **Claims Mapping**:
  - `sub` -> `UserID`
  - Custom claims (e.g., `realm_access.roles`, `scope`, or specific claim) -> `Roles`/`Permissions`.
  - Since standard OIDC doesn't strictly define role structure, provide a documented way (or simple default) to map a claim to roles.

### Common Pitfalls
- **Token Source**: Ensure you strip `Bearer ` prefix if parsing header manually.
- **Context Cancellation**: Ensure OIDC provider discovery uses context.
- **TLS**: In dev/docker, issuer URL might be http (allow `InsecureSkipSignatureCheck` or similar only in dev).

## Dev Agent Record

### Context Reference
- `docs/epics.md` (Story 14.1)
- `internal/interface/http/middleware/auth.go` (Interface definition)

### Agent Model Used
- standard-agent-model

### Completion Notes List
- Confirmed `Authenticator` interface exists.
- Confirmed library choice `go-oidc`.
- Implemented `OIDCAuthenticator` with `go-oidc` verifier.
- Added `OIDCConfig` to support `enabled`, `issuer_url`, `client_id`, `audience`.
- Mapped `sub` -> `UserID` and `realm_access.roles` -> `Roles`.
- Initialized in `cmd/server/main.go` and injected into `RouterDeps`.
- Added unit tests with `MockKeySet` to verify token validation logic without external provider.

### Review Follow-ups (AI)
- **High Severity Fix**: Implemented flexible claims mapping support via `RolesClaim` config. now supports dot-notation nested claims (e.g. `permissions.roles`) and space-separated string claims.
- **Medium Severity Fix**: Enabled manual audience validation in `OIDCAuthenticator` to support `Audience` config list, skipping `go-oidc` default check when multiple audiences are configured.
- **Low Severity Fix**: Cleaned up `OIDCAuthenticator` struct usage and fixed documentation/implementation mismatches.
- **Tests**: Expanded `oidc_test.go` to cover custom claims, whitespace-separated roles, and audience validation scenarios.

