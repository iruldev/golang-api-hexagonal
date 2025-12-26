# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

Only the latest major version receives security updates. We recommend always running the most recent release.

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

1. **GitHub Private Disclosure** (Preferred): Use [GitHub's private vulnerability reporting](https://github.com/iruldev/golang-api-hexagonal/security/advisories/new)
2. **Email**: security@yourdomain.com (replace with actual security contact)

### What to Include

- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Any suggested fixes (optional)

### Response Timeline

| Action | Timeline |
|--------|----------|
| Acknowledgment | 48 hours |
| Initial assessment | 1 week |
| Fix released | Depends on severity |

### Severity Classification

- **Critical**: Remote code execution, authentication bypass → Patch within 24-48 hours
- **High**: Data exposure, privilege escalation → Patch within 1 week
- **Medium**: Information disclosure, DoS → Patch in next scheduled release
- **Low**: Minor issues → Addressed in future releases

## Threat Model Summary

### Authentication & Authorization

| Control | Implementation |
|---------|----------------|
| Authentication | JWT with HS256 signing |
| Token Validation | Issuer, audience, expiry verified |
| Clock Skew | Configurable tolerance (default: 30s) |
| Secret Management | Environment variables with `*_FILE` pattern support |

### Input Validation

| Control | Implementation |
|---------|----------------|
| JSON Parsing | Strict mode - unknown fields rejected |
| Trailing Data | Rejected to prevent injection |
| Request Size | Configurable limit (default: 1MB) |
| Content-Type | Validated before processing |

### Rate Limiting

| Control | Implementation |
|---------|----------------|
| Strategy | Per-IP rate limiting |
| Default | 100 requests/second |
| Trust Proxy | Configurable (disabled by default) |
| Headers | `X-RateLimit-*` headers in responses |

### Data Protection

| Control | Implementation |
|---------|----------------|
| PII Redaction | Audit logs redact email, sensitive fields |
| Error Responses | RFC 7807 format, no stack traces |
| Database | TLS connections supported |
| Secrets | Never logged, `*_FILE` pattern for container secrets |

## Security Design Decisions

### 1. JWT Over Session Cookies

**Decision**: Stateless JWT authentication instead of server-side sessions.

**Rationale**:
- Enables horizontal scaling without session replication
- Reduces database load (no session table queries)
- Works naturally with microservices architecture

**Trade-offs**:
- Tokens cannot be individually revoked before expiry
- Token size larger than session cookies

### 2. `*_FILE` Pattern for Secrets

**Decision**: Support reading secrets from files via `JWT_SECRET_FILE`, `DATABASE_URL_FILE`, etc.

**Rationale**:
- Compatible with Kubernetes secrets mounted as files
- Compatible with Docker secrets
- Avoids exposing secrets in `docker inspect` or `/proc/*/environ`

**Example**:
```bash
# Instead of:
export JWT_SECRET="supersecret"

# Use:
echo "supersecret" > /run/secrets/jwt_secret
export JWT_SECRET_FILE="/run/secrets/jwt_secret"
```

### 3. RFC 7807 Problem Details

**Decision**: All error responses use RFC 7807 format.

**Rationale**:
- Standardized error format for API consumers
- Never exposes internal stack traces
- Includes trace IDs for debugging without security risk

**Example Response**:
```json
{
  "type": "https://example.com/problems/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "Email format is invalid",
  "instance": "/api/v1/users",
  "trace_id": "abc123"
}
```

### 4. TLS Termination at Load Balancer

**Decision**: Application runs HTTP; TLS handled by infrastructure.

**Rationale**:
- Simplifies certificate management
- Better performance (hardware TLS offloading)
- Standard cloud-native pattern

**Configuration**:
- Set `TRUST_PROXY=true` when behind a trusted proxy
- Use `X-Forwarded-For` / `X-Real-IP` for real client IPs

### 5. Application-Layer Rate Limiting

**Decision**: Rate limiting in application, not just infrastructure.

**Rationale**:
- Defense in depth (infrastructure + application)
- Per-tenant/per-user limits possible
- Consistent behavior across environments

## Security Headers

The application sets the following security headers on all responses:

| Header | Value | Purpose |
|--------|-------|---------|
| `Content-Security-Policy` | `default-src 'none'` | Prevent XSS |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-XSS-Protection` | `0` | Disable legacy XSS filter |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer info |
| `Cache-Control` | `no-store` | Prevent sensitive data caching |

When deployed behind HTTPS (recommended):
- `Strict-Transport-Security: max-age=31536000; includeSubDomains`

## Security Scanning

### Automated Checks

The CI pipeline includes:

- **govulncheck**: Go vulnerability database scanning
- **gitleaks**: Secret detection in code/history
- **golangci-lint**: Static analysis including security linters

### Running Locally

```bash
# Vulnerability scanning
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Secret detection
docker run -v $(pwd):/path gitleaks/gitleaks:latest detect --source="/path"

# Linting (includes gosec)
make lint
```

## Dependency Management

- Dependencies pinned to specific versions in `go.mod`
- Dependabot enabled for automated security updates
- All dependencies scanned with `govulncheck` in CI

## Secure Development Practices

1. **No secrets in code**: Use environment variables or `*_FILE` pattern
2. **Input validation**: All user input validated before processing
3. **Output encoding**: JSON encoding prevents injection
4. **Least privilege**: Database user has minimal required permissions
5. **Audit logging**: Security-relevant events logged with PII redaction

---

*Last updated: 2025-12-26*
