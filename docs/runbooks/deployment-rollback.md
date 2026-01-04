# Runbook: Deployment Rollback

**Severity:** 🟡 Medium (Procedural)
**Last Updated:** 2026-01-04
**Owner:** DevOps Team
**Related ADRs:** [ADR-001: Hexagonal Architecture](../adr/ADR-001-hexagonal-architecture.md)

---

## Overview

This runbook provides step-by-step instructions for rolling back a deployment when a new release causes issues. This is a **procedural** runbook used when a deployment has introduced bugs, performance regressions, or breaking changes.

**Use Case:** Failed deployment recovery, hotfix reversion, emergency rollback.

---

## Prerequisites

- **Access Required:**
  - [ ] Kubernetes cluster access (`kubectl`)
  - [ ] GitHub Actions access (for CI/CD)
  - [ ] Docker registry access (for image verification)

- **Tools Required:**
  - [ ] `kubectl` CLI
  - [ ] `git` CLI
  - [ ] GitHub CLI (`gh`) - optional

- **Knowledge Required:**
  - Kubernetes deployment management
  - Git version control
  - CI/CD pipeline basics

---

## Symptoms

### When to Use This Runbook

| Trigger | Description |
|---------|-------------|
| 🚨 High error rate after deployment | `http_requests_total{status="5xx"}` spike |
| 📉 Performance degradation | Response latency increased significantly |
| 🐛 Critical bug discovered | User-impacting bug in new release |
| ❌ Health checks failing | New deployment not passing health probes |
| ⚠️ Feature rollback | Business decision to revert a feature |

### Key Metrics to Evaluate

- Error rate before vs after deployment
- Response latency p50/p95/p99
- Health check status
- User-reported issues

---

## Investigation Steps

### Step 1: Confirm Deployment is the Cause

```bash
# Check when last deployment occurred
kubectl rollout history deployment/api-server

# Compare current vs previous revision
kubectl rollout history deployment/api-server --revision=<current>
kubectl rollout history deployment/api-server --revision=<previous>

# Check deployment events
kubectl describe deployment api-server | grep -A20 "Events:"
```

### Step 2: Identify Current State

```bash
# Check current pods
kubectl get pods -l app=api-server -o wide

# Check current image version
kubectl get deployment api-server -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check pod logs for errors
kubectl logs -l app=api-server --tail=100 | grep -iE "(error|panic|fatal)"
```

### Step 3: Compare Metrics Before/After

Review dashboards to compare:
- Error rate timeline
- Latency percentiles
- Resource usage
- Health check results

**Decision Point:** If deployment is confirmed as the cause, proceed with rollback.

---

## Resolution Steps (Rollback Procedure)

### Method 1: Kubernetes Rollback (Fastest)

> ⚡ **Time to Execute:** < 2 minutes

```bash
# View rollout history
kubectl rollout history deployment/api-server

# Rollback to previous revision
kubectl rollout undo deployment/api-server

# Verify rollback is progressing
kubectl rollout status deployment/api-server

# Confirm rollback complete
kubectl get deployment api-server -o jsonpath='{.spec.template.spec.containers[0].image}'
```

### Method 2: Rollback to Specific Revision

```bash
# View available revisions
kubectl rollout history deployment/api-server

# Rollback to specific revision (e.g., revision 5)
kubectl rollout undo deployment/api-server --to-revision=5

# Monitor rollback
kubectl rollout status deployment/api-server
```

### Method 3: Manual Image Update

```bash
# Set specific image version manually
kubectl set image deployment/api-server \
  api-server=registry.example.com/api-server:v1.2.3

# Monitor deployment
kubectl rollout status deployment/api-server
```

### Method 4: Git Revert + Redeploy

> 🔧 **Use when:** Need permanent code reversion

```bash
# Revert the problematic commit
git revert <problematic-commit-hash>
git push origin main

# Trigger CI/CD pipeline
# (automatic if configured, or manually trigger)
gh workflow run ci.yml
```

---

## Post-Rollback Actions

### Immediate Verification

```bash
# Verify pods are running with previous version
kubectl get pods -l app=api-server -o wide
kubectl get deployment api-server -o jsonpath='{.spec.template.spec.containers[0].image}'

# Check health endpoints
curl -s http://${API_HOST}:8080/healthz
curl -s http://${API_HOST}:8080/readyz | jq .

# Verify error rate normalized
kubectl logs -l app=api-server --tail=50 | grep -iE "(error|panic)" || echo "No errors found"
```

### Communication

1. **Notify team:** Post in #incidents channel
2. **Update status page:** If public-facing
3. **Create incident ticket:** Document the rollback

### Root Cause Investigation

After stabilization:
1. Create bug report for the issue
2. Add regression test for the bug
3. Fix the issue in a new branch
4. Deploy fix with proper testing

---

## Rollback Prevention

### Before Next Deployment

- [ ] Verify staging tests passed
- [ ] Check canary metrics (if using canary deployment)
- [ ] Review deployment diff
- [ ] Confirm rollback procedure documented
- [ ] Ensure previous revision is available

---

## Verification

### Confirm Rollback Success

- [ ] Previous image version is running
- [ ] All pods are in Running state
- [ ] Health checks passing (`/healthz`, `/readyz`)
- [ ] Error rate returned to pre-deployment levels
- [ ] No user-reported issues with functionality

### Metrics to Monitor

| Metric | Expected Value | Dashboard |
|--------|---------------|-----------|
| `http_requests_total{status="5xx"}` | Back to baseline | API Dashboard |
| `http_request_duration_seconds_p99` | Back to baseline | API Dashboard |
| Pod ready status | All pods ready | K8s Dashboard |
| Deployment revision | Previous version | K8s Dashboard |

### Observation Period

⏱️ **Recommended:** Monitor for 15 minutes to ensure stability

---

## Escalation

| Level | Contact | Criteria | Response Time |
|-------|---------|----------|---------------|
| L1 | On-call Engineer | Standard rollback | Immediate |
| L2 | Senior DevOps | Rollback failing | 10 min |
| L3 | Team Lead | Data integrity concerns | 5 min |

### Escalation Triggers

- ⚠️ Escalate to L2 if: Rollback stuck, pods not starting
- 🚨 Escalate to L3 if: Rollback causes data issues, multiple services affected

---

## Related Runbooks

- [Database Connection Failure](./database-connection-failure.md) - If deployment affected DB
- [Circuit Breaker Open](./circuit-breaker-open.md) - If deployment triggered circuit breakers
- [Memory/CPU Spike](./memory-cpu-spike.md) - If deployment caused resource issues

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| 2026-01-04 | DevOps Team | Initial creation |
