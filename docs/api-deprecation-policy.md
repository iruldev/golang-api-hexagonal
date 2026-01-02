# API Deprecation Policy

This document defines the policy for deprecating API features, endpoints, and versions in the golang-api-hexagonal project. It ensures API consumers have adequate time and information to migrate away from deprecated features.

## Overview

When API features are deprecated, we follow a structured process to ensure consumers can plan and execute migrations safely. This policy applies to:

- **Endpoint deprecation** - Individual API endpoints being retired
- **Field deprecation** - Request/response fields being removed or changed
- **Parameter deprecation** - Query or path parameters being retired
- **Version deprecation** - Entire API versions being sunset

**Key Commitment:** We provide a **minimum 6-month notice period** before removing any stable API feature.

## Deprecation Types and Scopes

### Endpoint Deprecation

When an individual endpoint is deprecated:

| Scope | Example | Impact Level |
|-------|---------|--------------|
| Single endpoint removal | `DELETE /api/v1/legacy-feature` removed | High |
| Endpoint replacement | `POST /users` → `POST /members` | High |
| Endpoint consolidation | Multiple endpoints merged into one | Medium |

**Process:**
1. New endpoint (if replacement) is available before deprecation announcement
2. Deprecation headers added to old endpoint
3. Documentation updated with migration guide
4. Sunset date communicated

### Field/Parameter Deprecation

When request or response fields are deprecated:

| Scope | Example | Impact Level |
|-------|---------|--------------|
| Response field removal | `metadata` field removed from response | Medium |
| Request field removal | `legacy_flag` parameter no longer accepted | Medium |
| Field rename | `firstName` → `first_name` | High |
| Type change | `id: integer` → `id: string` | High |

**Process:**
1. New field available (if replacement) before deprecation
2. Field marked `deprecated: true` in OpenAPI spec
3. Documentation notes which field replaces it
4. During transition, both old and new fields may be present in responses

### Version Deprecation

When an entire API version is deprecated:

| Scope | Timeline | Notice Required |
|-------|----------|-----------------|
| Major version (v1 → v2) | 12-24 months overlap | 6 months minimum |
| Minor version updates | N/A | No deprecation (backward compatible) |

**Process:**
1. New version released with all needed functionality
2. Old version marked deprecated with sunset date
3. Migration guide published
4. Extended support period for critical security fixes only

### Deprecation Severity Levels

| Severity | Description | Min Notice Period | Consumer Action |
|----------|-------------|-------------------|-----------------|
| **Critical** | Breaking change, no workaround | 6 months | Immediate migration planning |
| **Major** | Breaking change, workaround available | 6 months | Plan migration within notice period |
| **Minor** | Non-breaking, behavioral change | 3 months | Update when convenient |

## Notice Period and Timeline

### Minimum Notice Periods

| Feature Type | Minimum Notice | Recommended Notice |
|--------------|----------------|-------------------|
| Stable endpoints | 6 months (180 days) | 12 months |
| Stable fields/parameters | 6 months (180 days) | 9 months |
| Beta endpoints | 30 days | 60 days |
| Alpha endpoints | 0 days (may change without notice) | 7 days courtesy |

### Deprecation Timeline Milestones

```
D-Day         D+30          D+60          D+150         D+173         D+180
  │             │             │              │             │             │
  ▼             ▼             ▼              ▼             ▼             ▼
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ Announce │ │ 30-Day   │ │ 60-Day   │ │ 30-Day   │ │ 7-Day    │ │ SUNSET   │
│ Headers  │ │ Reminder │ │ Reminder │ │ Warning  │ │ Final    │ │ Endpoint │
│ Added    │ │ Email    │ │ Email    │ │ Email    │ │ Warning  │ │ Removed  │
└──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘ └──────────┘
```

### Step-by-Step Deprecation Process

1. **D-Day (Announcement)**
   - Add `Deprecation` and `Sunset` headers to affected endpoints
   - Update OpenAPI spec with `deprecated: true`
   - Publish migration documentation
   - Send initial notification email to registered consumers
   - Update developer portal with deprecation notice

2. **D+30 (First Reminder)**
   - Send reminder email to consumers
   - Log usage analytics of deprecated feature

3. **D+60 (Second Reminder)**
   - Send second reminder email
   - Escalate to consumers with high usage

4. **D+150 (30-Day Warning)**
   - Send urgent warning email
   - Personal outreach to top consumers if needed

5. **D+173 (7-Day Final Warning)**
   - Send final warning email
   - Prepare removal changeset

6. **D+180 (Sunset)**
   - Remove endpoint/feature
   - Return 410 Gone or redirect if applicable
   - Archive migration documentation

## HTTP Headers

When a feature is deprecated, the following HTTP headers are added to all responses:

### Deprecation Header (IETF Draft)

Indicates the date when the deprecation was announced:

```http
Deprecation: Sat, 01 Jul 2026 00:00:00 GMT
```

**Format:** HTTP-date as defined in [RFC 7231 Section 7.1.1.1](https://tools.ietf.org/html/rfc7231#section-7.1.1.1)
**Spec:** [draft-ietf-httpapi-deprecation-header](https://datatracker.ietf.org/doc/draft-ietf-httpapi-deprecation-header/)

### Sunset Header (RFC 8594)

Indicates the date when the feature will be removed:

```http
Sunset: Sun, 01 Jan 2027 00:00:00 GMT
```

**Format:** HTTP-date as defined in [RFC 7231 Section 7.1.1.1](https://tools.ietf.org/html/rfc7231#section-7.1.1.1)

### Link Header (RFC 8288)

Provides a link to migration documentation:

```http
Link: <https://api.example.com/docs/migrations/v1-to-v2>; rel="deprecation"
```

**Format:** Web Linking as defined in [RFC 8288](https://tools.ietf.org/html/rfc8288)

### Complete Header Example

```http
HTTP/1.1 200 OK
Content-Type: application/json
Deprecation: Sat, 01 Jul 2026 00:00:00 GMT
Sunset: Sun, 01 Jan 2027 00:00:00 GMT
Link: <https://api.example.com/docs/migrations/legacy-users>; rel="deprecation"

{
  "id": "123",
  "name": "John Doe",
  "legacy_field": "This field is deprecated"
}
```

**Key Points:**
- The response body remains fully functional until sunset date
- Consumers should monitor for these headers
- All three headers appear together when deprecation is active

## Communication Process

### Channels

| Channel | Timing | Audience |
|---------|--------|----------|
| **HTTP Headers** | Immediate (every request) | All API consumers |
| **OpenAPI Spec** | Immediate (spec update) | Developers, tooling |
| **Developer Portal** | Immediate | All visitors |
| **Email Notifications** | Scheduled (D-Day, +30, +60, +150, +173) | Registered consumers |
| **GitHub Release Notes** | On release | Repository watchers |
| **Changelog** | On release | All users |

### Email Notification Schedule

| Email | Timing | Subject Template |
|-------|--------|------------------|
| **Initial** | D-Day | "[Action Required] API Deprecation Notice: {feature}" |
| **30-Day Reminder** | D+30 | "[Reminder] API Feature {feature} deprecated - 150 days remaining" |
| **60-Day Reminder** | D+60 | "[Reminder] API Feature {feature} deprecated - 120 days remaining" |
| **30-Day Warning** | D+150 | "[Urgent] API Feature {feature} sunsetting in 30 days" |
| **Final Warning** | D+173 | "[Final Warning] API Feature {feature} sunsetting in 7 days" |

### OpenAPI Specification Markup

Deprecated endpoints are marked in the OpenAPI spec:

```yaml
paths:
  /api/v1/legacy-users:
    get:
      deprecated: true
      summary: Get legacy users (DEPRECATED)
      description: |
        ⚠️ **DEPRECATED**: This endpoint will be removed on 2027-01-01.
        
        **Migration:** Use `/api/v2/users` instead.
        See [Migration Guide](./docs/migrations/legacy-users.md).
      responses:
        '200':
          description: Successful response
          headers:
            Deprecation:
              $ref: '#/components/headers/Deprecation'
            Sunset:
              $ref: '#/components/headers/Sunset'
            Link:
              $ref: '#/components/headers/Link'
```

## Consumer Responsibilities

### Detecting Deprecated Features

**1. Check Response Headers:**
```bash
curl -I https://api.example.com/api/v1/legacy-endpoint
# Look for: Deprecation, Sunset, Link headers
```

**2. Monitor OpenAPI Specification:**
- Check for `deprecated: true` on endpoints
- Review spec diffs in release notes

**3. Subscribe to Notifications:**
- Register for email notifications via developer portal
- Watch repository for release announcements

**4. Use Automated Monitoring:**
```javascript
// Example: Client-side deprecation detection
response.headers.get('Deprecation') && 
  console.warn('API endpoint is deprecated:', response.headers.get('Sunset'));
```

### Migration Path Expectations

When migrating from deprecated features:

1. **Review Migration Guide** - Check documentation for step-by-step instructions
2. **Test in Staging** - Validate new endpoints before production switch
3. **Parallel Usage** - Run old and new simultaneously during transition if possible
4. **Monitor Errors** - Watch for unexpected behavior after migration
5. **Complete Early** - Don't wait until sunset date

### Backward Compatibility During Deprecation

**Guarantees during deprecation period:**
- Deprecated endpoints continue to function normally
- Response format remains unchanged
- Error codes and behavior stay consistent
- Performance SLAs are maintained (though may be reduced priority)

**Exceptions:**
- Security vulnerabilities may require immediate changes
- Critical bugs may be fixed with shorter notice
- Rate limits may be applied more aggressively

### Exception Handling for Deprecated Endpoints

After sunset, deprecated endpoints return:

```http
HTTP/1.1 410 Gone
Content-Type: application/problem+json

{
  "type": "https://api.example.com/problems/endpoint-removed",
  "title": "Endpoint Removed",
  "status": 410,
  "detail": "This endpoint was deprecated on 2026-07-01 and removed on 2027-01-01. Please use /api/v2/users instead.",
  "code": "SYS-010",
  "instance": "/api/v1/legacy-users"
}
```

## Maintainer Checklist

When deprecating an API feature, maintainers must complete:

### Pre-Deprecation

- [ ] Identify replacement feature/endpoint (if applicable)
- [ ] Implement and release replacement
- [ ] Create migration documentation
- [ ] Determine sunset date (minimum 6 months from announcement)
- [ ] Prepare email notification templates

### Announcement (D-Day)

- [ ] Add `Deprecation`, `Sunset`, and `Link` headers to endpoint
- [ ] Update OpenAPI spec with `deprecated: true`
- [ ] Add deprecation notice to endpoint description
- [ ] Publish migration guide to documentation
- [ ] Send initial deprecation email to registered consumers
- [ ] Update developer portal with deprecation banner
- [ ] Add entry to CHANGELOG.md
- [ ] Create GitHub release note

### During Deprecation Period

- [ ] Send 30-day reminder email (D+30)
- [ ] Send 60-day reminder email (D+60)
- [ ] Monitor usage analytics for the deprecated feature
- [ ] Send 30-day warning email (D+150)
- [ ] Personal outreach to high-usage consumers (D+150)
- [ ] Send final 7-day warning email (D+173)

### Sunset (D+180)

- [ ] Remove endpoint from codebase
- [ ] Return 410 Gone for removed endpoints
- [ ] Remove deprecated fields from responses
- [ ] Update OpenAPI spec to remove endpoint
- [ ] Archive migration documentation
- [ ] Send sunset confirmation email
- [ ] Update CHANGELOG.md

## Extensions and Exceptions

### Extension Requests

Consumers may request a sunset extension if:
- Significant technical challenges prevent migration
- Business-critical integration requires more time
- Force majeure circumstances apply

**Process:**
1. Submit extension request via GitHub issue
2. Include: Consumer ID, affected integration, reason, proposed new date
3. Review by API maintainers within 5 business days
4. Extensions granted on case-by-case basis

### Emergency Removal

In exceptional cases, features may be removed with shortened notice:

| Reason | Minimum Notice | Authority |
|--------|----------------|-----------|
| Security vulnerability | 24-48 hours | Security team |
| Legal/compliance requirement | 7 days | Legal team |
| Critical infrastructure issue | 48 hours | Engineering leadership |

Emergency removals will be communicated via all available channels immediately.

## Related Documents

- [API Versioning Strategy](./api-versioning.md) - Version lifecycle and migration guidance
- [OpenAPI Specification](./openapi.yaml) - Machine-readable API definition
- [Changelog](../CHANGELOG.md) - Version history and deprecation notices
- [RFC 8594: Sunset Header](https://tools.ietf.org/html/rfc8594) - Sunset HTTP header specification
- [RFC 8288: Web Linking](https://tools.ietf.org/html/rfc8288) - Link header specification

## Contact

For questions about deprecation policy or migration assistance:

- **Issues**: [GitHub Issues](https://github.com/iruldev/golang-api-hexagonal/issues)
- **Repository**: https://github.com/iruldev/golang-api-hexagonal
