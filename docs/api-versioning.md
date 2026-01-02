# API Versioning Strategy

This document describes the API versioning strategy for the golang-api-hexagonal project, ensuring API consumers understand how versions evolve and how to handle changes.

## Overview

This API uses **URL-based versioning** with the version embedded in the URL path. This approach provides clear, explicit version identification and simplifies routing at the infrastructure level.

**Current Version:** v1 (Stable)

**Base URL Pattern:**
```
/api/v{major}/resource
```

**Examples:**
```
GET /api/v1/users
GET /api/v2/users   # Future version
POST /api/v1/users
GET /api/v1/users/{id}
```

## Versioning Scheme

### URL-Based Versioning

We use URL path versioning where the major version number is embedded in the URL:

| Version | URL Prefix | Status |
|---------|------------|--------|
| v1 | `/api/v1/` | Stable |
| v2 | `/api/v2/` | (Future) |

**Rationale for URL-Based Versioning:**
- **Explicit and visible** - Version is immediately apparent in every request
- **Easy routing** - Load balancers and API gateways can route based on URL path
- **Cache-friendly** - URLs are unique per version, simplifying caching
- **Browser-friendly** - Can be tested directly in browser address bar
- **Industry standard** - Used by major APIs (Twitter, GitHub, Stripe)

### Unversioned Endpoints

System operational endpoints exist at the root level and are **not** versioned. These purely infrastructure-level probes do not change behavior based on API versions:

- `/healthz`
- `/readyz`
- `/startupz`

### Semantic Versioning Principles

While the URL contains only the major version, internal versioning follows [Semantic Versioning 2.0.0](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

| Component | Increment When |
|-----------|---------------|
| **MAJOR** | Incompatible API changes (requires new URL prefix) |
| **MINOR** | Backward-compatible functionality additions |
| **PATCH** | Backward-compatible bug fixes |

**Example:** Version `1.3.2` means:
- Major version 1 (URL: `/api/v1/`)
- Minor version 3 (third feature release)
- Patch version 2 (second bug fix)

The full version is available in the API documentation and can be queried via the OpenAPI specification.

## Version Lifecycle

Each API version progresses through defined lifecycle stages:

```
┌─────────┐     ┌────────┐     ┌─────────┐     ┌────────────┐     ┌─────────┐
│  Alpha  │ ──▶ │  Beta  │ ──▶ │ Stable  │ ──▶ │ Deprecated │ ──▶ │ Sunset  │
└─────────┘     └────────┘     └─────────┘     └────────────┘     └─────────┘
```

### Stage Definitions

| Stage | Description | SLA | Notice Period |
|-------|-------------|-----|---------------|
| **Alpha** | Early development. API may change without notice. | None | None |
| **Beta** | Feature complete. Breaking changes with 30-day notice. | Best effort | 30 days |
| **Stable** | Production-ready. Full deprecation policy applies. | Full SLA | Per deprecation policy |
| **Deprecated** | Still functional but scheduled for removal. | Reduced SLA | 6 months minimum |
| **Sunset** | Version removed from service. | None | N/A |

### Stage Indicators

The current lifecycle stage is communicated via:

1. **OpenAPI Specification** - Each endpoint marked with stability status
2. **API Documentation** - Version support matrix maintained
3. **Response Headers** - Deprecation notices when applicable

## Breaking Changes

### What Constitutes a Breaking Change

The following changes are considered **breaking** and require a new major version:

| Change Type | Example | Impact |
|-------------|---------|--------|
| **Removing an endpoint** | DELETE `/api/v1/legacy-endpoint` | Clients using removed endpoint will fail |
| **Removing a response field** | Remove `metadata` from response | Clients parsing removed field will break |
| **Changing field data type** | `id`: Integer → String | Type parsing will fail |
| **Adding required request field** | New required `tenant_id` field | Existing clients won't send new field |
| **Changing field name** | `firstName` → `first_name` | Field mapping will break |
| **Changing URL structure** | `/users/{id}` → `/users/{uuid}` | Existing URL patterns fail |
| **Changing HTTP method** | POST → PUT | Client HTTP method calls fail |
| **Changing authentication scheme** | Bearer → OAuth2 | Auth flow incompatibility |
| **Reducing rate limits significantly** | 1000/min → 100/min | Clients may get throttled |
| **Changing error response structure** | JSON → Problem Details | Error parsing breaks |

### What Constitutes a Non-Breaking Change

The following changes are **non-breaking** and can be released in minor versions:

| Change Type | Example | Why Safe |
|-------------|---------|----------|
| **Adding new endpoint** | POST `/api/v1/users/bulk` | Existing clients unaffected |
| **Adding optional request field** | New optional `nickname` field | Clients can omit it |
| **Adding response field** | New `avatarUrl` in response | Clients ignore unknown fields |
| **Relaxing validation** | Email: required → optional | Existing requests still valid |
| **Expanding enum values** | Status: add "PENDING" | Existing values unchanged |
| **Increasing rate limits** | 100/min → 500/min | More permissive is safe |
| **Adding new response codes** | New 207 Multi-Status | Existing codes unchanged |
| **Improving error messages** | Better `detail` text | Structure unchanged |
| **Adding new query parameters** | New `?fields=` filter | Optional, backward compatible |
| **Performance improvements** | Faster response times | Transparent to clients |

## Migration Guide

### Migrating Between Major Versions

When a new major version is released, follow this migration process:

#### Step 1: Review Release Notes
Check the changelog for:
- Breaking changes list
- Mapping from old to new endpoints/fields
- New features available

#### Step 2: Update API Client
```diff
- const API_BASE = 'https://api.example.com/api/v1';
+ const API_BASE = 'https://api.example.com/api/v2';
```

#### Step 3: Update Request/Response Handling
Adapt to any changed field names or structures:
```diff
// Example: Field name change
- const name = response.firstName;
+ const name = response.first_name;
```

#### Step 4: Test in Staging
- Run integration tests against new version
- Verify all workflows function correctly
- Check error handling for new error formats

#### Step 5: Deploy Gradually
- Use feature flags if possible
- Monitor error rates during rollout
- Keep fallback to previous version ready

### Parallel Version Support

During migration periods, both versions remain available:
- New features only in new version
- Bug fixes backported to stable versions
- Security fixes applied to all supported versions

## Deprecation Process

### How Deprecation is Communicated

When an endpoint or version is deprecated, we communicate through:

#### 1. HTTP Headers (RFC 8594)

```http
Deprecation: Sat, 01 Jul 2026 00:00:00 GMT
Sunset: Sun, 01 Jan 2027 00:00:00 GMT
Link: <https://api.example.com/docs/migration-v2>; rel="deprecation"
```

| Header | Description | Format |
|--------|-------------|--------|
| `Deprecation` | Date when deprecation was announced | HTTP-date (RFC 7231) |
| `Sunset` | Date when endpoint will be removed | HTTP-date (RFC 7231) |
| `Link` | URL to migration documentation | RFC 8288 Web Linking |

#### 2. OpenAPI Specification

Deprecated endpoints are marked in the OpenAPI spec:

```yaml
paths:
  /api/v1/legacy-endpoint:
    get:
      deprecated: true
      description: |
        **⚠️ DEPRECATED**: This endpoint will be removed on 2027-01-01.
        Use `/api/v2/new-endpoint` instead.
```

#### 3. Email Notifications

Registered API consumers receive:
- **Initial Notice**: When deprecation is announced
- **30-Day Reminder**: One month before sunset
- **7-Day Warning**: Final warning before removal

#### 4. Developer Portal

- Deprecation banner on affected endpoint documentation
- Migration guides with code examples
- FAQ for common migration questions

### Deprecation Timeline

| Milestone | Timing | Action |
|-----------|--------|--------|
| **Announcement** | D-Day | Deprecation published, headers added |
| **30-Day Notice** | D+30 | Reminder email sent |
| **60-Day Notice** | D+60 | Reminder email sent |
| **Final Warning** | D+173 (7 days before sunset) | Final warning email |
| **Sunset** | D+180 (6 months) | Endpoint removed |

**Minimum Deprecation Period:** 6 months (180 days)

## Version Support Matrix

### Currently Supported Versions

| Version | Status | Release Date | Deprecation Date | Sunset Date |
|---------|--------|--------------|------------------|-------------|
| v1 | **Stable** | 2026-01-01 | - | - |

### Version Support Policy

- **Stable versions**: Supported for minimum 24 months after release
- **Deprecated versions**: Remain functional for 6 months after deprecation
- **Security patches**: Applied to all supported versions within 72 hours
- **Bug fixes**: Applied to current stable and previous stable version

### Planned Versions

| Version | Planned Status | Target Date | Key Features |
|---------|----------------|-------------|--------------|
| v2 | Planning | TBD | Enhanced pagination, GraphQL support |

## Additional Resources

- [API Deprecation Policy](./api-deprecation-policy.md) - Full deprecation policy details
- [OpenAPI Specification](./openapi.yaml) - Machine-readable API definition
- [Changelog](../CHANGELOG.md) - Version history and release notes
- [RFC 8594: Sunset Header](https://tools.ietf.org/html/rfc8594) - Sunset HTTP header specification

## Contact

For questions about API versioning or migration assistance:
- **Issues**: [GitHub Issues](https://github.com/iruldev/golang-api-hexagonal/issues)
- **Repository**: https://github.com/iruldev/golang-api-hexagonal
