# Story 7.1: Create SECURITY.md

Status: done

## Story

**As a** security researcher,
**I want** SECURITY.md with threat model,
**So that** I understand security posture and can report issues.

**FR:** FR31

## Acceptance Criteria

1. ✅ **Given** the repository, **When** SECURITY.md is viewed, **Then** it includes threat model summary
2. ✅ **Given** SECURITY.md, **When** viewed, **Then** security contact/reporting process is documented
3. ✅ **Given** SECURITY.md, **When** viewed, **Then** security-related design decisions are listed


## Tasks

- [x] Create SECURITY.md
- [x] Document vulnerability reporting
- [x] Document threat model
- [x] Document security design decisions
- [x] Document security scanning
- [x] [AI-Review] Fix placeholder email in SECURITY.md
- [x] [AI-Review] Track SECURITY.md in git

## Senior Developer Review (AI)

- **Reviewer**: Gan
- **Date**: 2025-12-26
- **Outcome**: Approve (after fixes)
- **Notes**:
    - `SECURITY.md` created with good initial content.
    - Addressed placeholder email issue.
    - Ensured files are tracked in git.
    - Added missing Tasks section to story file.

## Implementation Summary

Created comprehensive `SECURITY.md` with:

### Vulnerability Reporting
- GitHub private disclosure (preferred)
- Response timeline (48h acknowledgment, 1 week assessment)
- Severity classification (Critical/High/Medium/Low)

### Threat Model Summary
- Authentication (JWT with HS256)
- Input validation (strict JSON, size limits)
- Rate limiting (per-IP, configurable)
- Data protection (PII redaction, RFC 7807 errors)

### Security Design Decisions
1. JWT over session cookies (stateless, horizontal scaling)
2. `*_FILE` pattern for secrets (Kubernetes/Docker friendly)
3. RFC 7807 Problem Details (no stack traces)
4. TLS termination at load balancer
5. Application-layer rate limiting

### Security Scanning
- govulncheck, gitleaks, golangci-lint in CI
- Local commands documented

## Dev Agent Record

### Agent Model Used

Gemini 2.5 Pro

### File List

- `SECURITY.md` - NEW
