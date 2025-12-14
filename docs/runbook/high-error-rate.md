# Runbook: High Error Rate

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `HighErrorRate`, `HighErrorRateCritical` |
| **Severity** | warning / critical |
| **Component** | http |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses alerts for elevated HTTP error rates (5xx responses).

**What triggers this alert:**
- **HighErrorRate (Warning):** 5xx errors > 5% for 5 minutes
- **HighErrorRateCritical (Critical):** 5xx errors > 10% for 2 minutes

**Business impact:**
- Users experience failures when accessing the API
- Data operations may fail, causing user frustration
- SLA violations if sustained

> **Deployment Context:** Commands in this runbook assume a Docker/Kubernetes deployment. For local development (where the app runs via `make dev` on the host), replace Docker commands with direct process inspection (e.g., check application logs in the terminal, use `ps aux | grep server`).

---

## Quick Response Checklist

- [ ] **Acknowledge** - Acknowledge the alert in PagerDuty/Slack
- [ ] **Assess** - Check error rate percentage and affected endpoints
- [ ] **Investigate** - Follow diagnosis steps below
- [ ] **Mitigate** - Apply immediate remediation (rollback if needed)
- [ ] **Communicate** - Update #incidents channel if critical
- [ ] **Document** - Record findings in incident log
- [ ] **Resolve** - Confirm error rate returns to normal

---

## Symptoms

### Observable Indicators
- API requests returning 500, 502, 503, or other 5xx status codes
- Users reporting "Something went wrong" errors
- Increased latency alongside errors

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `http_requests_total{status=~"5.."}` / total | < 1% | > 5% (warning), > 10% (critical) |

### Related Alerts
- `HighLatency` - May fire together if errors cause retries
- `ServiceDown` - If error rate reaches 100%
- `DBConnectionExhausted` - If database is the cause

---

## Diagnosis

### Step 1: Identify Affected Endpoints

**Goal:** Understand scope - is it all endpoints or specific ones?

```bash
# Check Prometheus for error breakdown by path
curl -s 'http://localhost:9090/api/v1/query?query=sum(rate(http_requests_total{status=~"5.."}[5m]))by(path)' | jq
```

**Expected outcome:** List of endpoints with their error rates

### Step 2: Check Application Logs

**Goal:** Find the actual error messages

```bash
# View recent error logs
docker logs golang-api 2>&1 | grep -i error | tail -50

# For Kubernetes
kubectl logs deployment/golang-api --tail=100 | grep -i error
```

### Step 3: Check Dependencies

**Goal:** Verify dependent services are healthy

```bash
# Check database
docker exec -it postgres pg_isready

# Check Redis
docker exec -it redis redis-cli ping

# Check readiness endpoint
curl -s http://localhost:8080/readyz
```

### Step 4: Check Resource Usage

**Goal:** Identify resource exhaustion

```bash
# Check container resources
docker stats golang-api --no-stream

# Check for OOM events
dmesg | grep -i "out of memory"
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Database down | Connection timeout errors | `pg_isready` fails | Check `db-connection-exhausted.md` |
| Out of memory | OOM kills in logs | dmesg shows OOM | Increase memory limits |
| Panic in handler | Stack traces in logs | Search logs for "panic" | Fix bug, rollback if needed |
| External API failure | Timeout errors | Check external service status | Enable circuit breaker |
| Bad deployment | Errors started at deploy time | Check deployment timeline | Rollback deployment |

---

## Remediation

### Immediate Actions

1. **If recent deployment:** Consider immediate rollback
   ```bash
   # Kubernetes rollback
   kubectl rollout undo deployment/golang-api
   
   # Docker rollback
   docker-compose down && docker-compose up -d
   ```

2. **If database issue:** Check database runbook
   - See: [DB Connection Exhausted](db-connection-exhausted.md)

3. **If memory issue:** Restart with increased limits
   ```bash
   docker-compose restart golang-api
   ```

4. **If code bug:** Deploy hotfix or rollback

### Post-Incident Actions

1. **Root Cause Analysis:** Document what caused the errors
2. **Prevention:** Add tests, improve monitoring
3. **Monitoring:** Consider alerting on specific error types

---

## Escalation

### Escalation Timeline

| Severity | Initial Response | Escalation If Unresolved |
|----------|------------------|--------------------------|
| **Warning** | Investigate within 30 minutes | Escalate after 1 hour |
| **Critical** | Immediate investigation | Escalate after 15 minutes |

### Escalation Path

1. **First Responder:** On-call engineer
2. **Level 1 Escalation:** Team lead / Senior engineer
3. **Level 2 Escalation:** Engineering manager

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | PagerDuty / #oncall | First response |
| Team Lead | @team-lead | After 15 min (critical) |
| DBA | @dba-team | If database-related |

---

## References

- [Grafana HTTP Dashboard](http://localhost:3000/d/http)
- [DB Connection Exhausted Runbook](db-connection-exhausted.md)
- [Service Down Runbook](service-down.md)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |
