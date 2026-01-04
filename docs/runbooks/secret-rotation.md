# Runbook: Secret Rotation

**Severity:** 🟡 Medium (Procedural)
**Last Updated:** 2026-01-04
**Owner:** DevOps Team
**Related ADRs:** [ADR-001: Hexagonal Architecture](../adr/ADR-001-hexagonal-architecture.md)

---

## Overview

This runbook provides step-by-step instructions for rotating secrets and credentials. Secret rotation should be performed on a regular schedule or immediately in response to a security incident.

**Use Cases:**
- Scheduled credential rotation (e.g., quarterly)
- Compromised credential response
- Employee offboarding
- Security compliance requirements

---

## Prerequisites

- **Access Required:**
  - [ ] Kubernetes cluster access (`kubectl`)
  - [ ] Secret management system (Vault, AWS Secrets Manager, etc.)
  - [ ] Database admin access (for DB credentials)
  - [ ] Cloud provider console access

- **Tools Required:**
  - [ ] `kubectl` CLI
  - [ ] Secret management CLI (if applicable)
  - [ ] Password generator (`openssl`, `pwgen`)

- **Knowledge Required:**
  - Kubernetes secrets management
  - Application configuration for secrets
  - Zero-downtime rotation strategies

---

## Symptoms / Triggers

### When to Rotate Secrets

| Trigger | Urgency | Description |
|---------|---------|-------------|
| 🗓️ Scheduled rotation | Low | Quarterly or per security policy |
| 🚨 Credential exposed | **Critical** | Secret found in logs, repo, or public |
| 👤 Employee departure | Medium | Revoke access, rotate shared secrets |
| 🔒 Security audit | Medium | Compliance requirement |
| ⚠️ Unusual access patterns | High | Suspected unauthorized access |

---

## Pre-Rotation Checklist

Before rotating secrets, verify:

- [ ] Identify all services using the secret
- [ ] Plan rotation window (consider traffic patterns)
- [ ] Notify team of planned rotation
- [ ] Ensure rollback plan is ready
- [ ] Test rotation in staging first

---

## Secret Types and Rotation Procedures

### 1. Database Credentials

#### Step 1: Generate New Password

```bash
# Generate strong password
openssl rand -base64 32

# Or use a password manager to generate
```

#### Step 2: Update Database User

```bash
# Connect to PostgreSQL
PGPASSWORD=${OLD_PASSWORD} psql -h ${DB_HOST} -U postgres

# Update password
ALTER USER ${DB_USER} WITH PASSWORD '${NEW_PASSWORD}';

# Verify new password works
PGPASSWORD=${NEW_PASSWORD} psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -c "SELECT 1;"
```

#### Step 3: Update Kubernetes Secret

```bash
# Create new secret
kubectl create secret generic db-credentials \
  --from-literal=password=${NEW_PASSWORD} \
  --dry-run=client -o yaml | kubectl apply -f -

# Or update existing secret
kubectl patch secret db-credentials \
  -p='{"stringData":{"password":"'${NEW_PASSWORD}'"}}'
```

#### Step 4: Restart Application

```bash
# Rolling restart to pick up new secret
kubectl rollout restart deployment/api-server

# Monitor rollout
kubectl rollout status deployment/api-server
```

---

### 2. JWT Signing Key

#### Step 1: Generate New Key

```bash
# Generate new HS256 secret (256 bits minimum)
openssl rand -base64 32
```

#### Step 2: Update Kubernetes Secret

```bash
kubectl create secret generic jwt-secret \
  --from-literal=JWT_SECRET=${NEW_SECRET} \
  --dry-run=client -o yaml | kubectl apply -f -
```

#### Step 3: Restart Application

```bash
kubectl rollout restart deployment/api-server
```

> ⚠️ **Note:** Rotating JWT secret will invalidate all existing tokens. Users will need to re-authenticate.

---

### 3. API Keys (External Services)

#### Step 1: Generate New Key in External Service

- Log into the external service dashboard
- Generate new API key
- Keep old key active during transition

#### Step 2: Update Kubernetes Secret

```bash
kubectl create secret generic external-api-keys \
  --from-literal=API_KEY=${NEW_API_KEY} \
  --dry-run=client -o yaml | kubectl apply -f -
```

#### Step 3: Restart and Verify

```bash
kubectl rollout restart deployment/api-server

# Verify external API connectivity
kubectl logs -l app=api-server --tail=50 | grep -i "external"
```

#### Step 4: Revoke Old Key

- Return to external service dashboard
- Revoke/delete the old API key

---

### 4. Environment-Based Secrets (.env files)

#### For Local Development

```bash
# Update .env.local file
sed -i 's/OLD_SECRET/NEW_SECRET/g' .env.local

# Restart local services
docker-compose down && docker-compose up -d
```

#### For Production (if using .env.docker)

> ⚠️ **Important:** Production should use Kubernetes secrets, not .env files

---

## Zero-Downtime Rotation Strategy

For critical secrets that require zero downtime:

### Dual-Secret Support

1. **Add new secret alongside old:**
   ```bash
   kubectl patch secret api-secrets \
     -p='{"stringData":{"NEW_KEY":"'${NEW_VALUE}'","OLD_KEY":"'${OLD_VALUE}'"}}'
   ```

2. **Update application to accept both** (if supported)

3. **Rolling restart:**
   ```bash
   kubectl rollout restart deployment/api-server
   ```

4. **Verify new secret is working**

5. **Remove old secret:**
   ```bash
   kubectl patch secret api-secrets \
     --type=json -p='[{"op": "remove", "path": "/data/OLD_KEY"}]'
   ```

---

## Rollback Procedure

### If New Secret Causes Issues

1. **Revert Kubernetes secret:**
   ```bash
   kubectl create secret generic db-credentials \
     --from-literal=password=${OLD_PASSWORD} \
     --dry-run=client -o yaml | kubectl apply -f -
   ```

2. **Restart application:**
   ```bash
   kubectl rollout restart deployment/api-server
   ```

3. **Revert database password (if changed):**
   ```bash
   # Connect as admin and reset password
   ALTER USER ${DB_USER} WITH PASSWORD '${OLD_PASSWORD}';
   ```

### Verify Rollback

- [ ] Application connecting successfully
- [ ] Health checks passing
- [ ] No authentication errors in logs

---

## Verification

### Confirm Rotation Success

- [ ] Application starts without secret-related errors
- [ ] All health checks passing
- [ ] External service integrations working
- [ ] User authentication working (for JWT rotation)
- [ ] Database queries executing successfully

### Logs to Check

```bash
# Check for secret-related errors
kubectl logs -l app=api-server --tail=100 | grep -iE "(secret|auth|credential|password|key)"

# Should see no errors related to authentication/authorization
```

### Metrics to Monitor

| Metric | Expected Value | Dashboard |
|--------|---------------|-----------|
| Authentication errors | 0 (or baseline) | Auth Dashboard |
| Database connection errors | 0 | DB Dashboard |
| External API errors | 0 | Integration Dashboard |
| Pod restart count | 1 (planned restart) | K8s Dashboard |

### Observation Period

⏱️ **Recommended:** Monitor for 30 minutes after rotation

---

## Escalation

| Level | Contact | Criteria | Response Time |
|-------|---------|----------|---------------|
| L1 | On-call Engineer | Planned rotation | Immediate |
| L2 | Security Team | Credential exposure | 15 min |
| L3 | CISO / Management | Major security breach | 5 min |

### Escalation Triggers

- ⚠️ Escalate to L2 if: Rotation failing, secret exposure suspected
- 🚨 Escalate to L3 if: Confirmed credential compromise, data breach potential

---

## Related Runbooks

- [Database Connection Failure](./database-connection-failure.md) - If DB credential rotation fails
- [Deployment Rollback](./deployment-rollback.md) - If secret rotation requires deployment rollback

---

## Security Best Practices

1. **Never log secrets** - Even in debug mode
2. **Use strong secrets** - Minimum 256-bit entropy for cryptographic keys
3. **Rotate regularly** - Quarterly at minimum, immediately if exposed
4. **Limit access** - Principle of least privilege
5. **Audit access** - Log all secret access attempts
6. **Automate rotation** - Use tools like Vault for automatic rotation

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | DevOps Team | Initial creation |
