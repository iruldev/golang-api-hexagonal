# Runbook: Job Failure Rate

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `JobFailureRate`, `JobFailureRateCritical` |
| **Severity** | warning / critical |
| **Component** | worker |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses alerts for elevated job failure rates in the async worker.

**What triggers this alert:**
- **JobFailureRate (Warning):** Job failures > 10% for 5 minutes
- **JobFailureRateCritical (Critical):** Job failures > 25% for 2 minutes

**Business impact:**
- Background operations failing (emails not sent, data not processed)
- User-initiated async operations not completing
- Potential data consistency issues

> **Deployment Context:** Commands in this runbook assume a Docker/Kubernetes deployment. For local development (where the app runs via `make dev` / `make worker` on the host), replace Docker commands with direct process inspection (e.g., check application logs in the terminal).

---

## Quick Response Checklist

- [ ] **Acknowledge** - Acknowledge the alert (immediate for critical)
- [ ] **Assess** - Check failure rate and affected job types
- [ ] **Investigate** - Identify failing tasks and error messages
- [ ] **Mitigate** - Fix underlying issue or disable failing jobs
- [ ] **Communicate** - Update #incidents if user-facing
- [ ] **Document** - Record error patterns found
- [ ] **Resolve** - Confirm failure rate returns to normal

---

## Symptoms

### Observable Indicators
- Jobs moving to retry queue repeatedly
- Error logs with job processing failures
- Archived/dead-letter queue growing

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `job_processed_total{status="failed"}` / total | < 5% | > 10% (warning), > 25% (critical) |
| Retry queue depth | Low | Growing |

### Related Alerts
- `JobQueueBacklog` - Failures cause backlog
- `DBConnectionExhausted` - If jobs fail due to DB
- `HighErrorRate` - If HTTP calls in jobs fail

---

## Diagnosis

### Step 1: Check Worker Logs

**Goal:** Find error messages for failing jobs

```bash
# Check for errors in worker logs
docker logs worker 2>&1 | grep -i "error\|failed\|panic" | tail -50

# For Kubernetes
kubectl logs deployment/worker --tail=100 | grep -i "error\|failed"
```

**Look for:** Error patterns, stack traces, connection failures

### Step 2: Identify Failing Task Types

**Goal:** Determine which job types are failing

```bash
# Check failed jobs by type
asynq task ls --queue=default --state=failed

# Check retry queue
asynq task ls --queue=default --state=retry

# Get details on a specific failed task
asynq task inspect <task_id>
```

### Step 3: Check Job Metrics

**Goal:** See failure patterns by job type

```bash
# Check job metrics breakdown
curl -s http://localhost:8080/metrics | grep 'job_processed_total{.*status="failed"}'

# Check job duration (slow jobs may timeout)
curl -s http://localhost:8080/metrics | grep job_duration
```

### Step 4: Check Dependencies

**Goal:** Verify services that jobs depend on

```bash
# Check database connectivity
docker exec -it postgres pg_isready

# Check Redis
docker exec -it redis redis-cli ping

# Check external services (if jobs call external APIs)
curl -s https://api.external-service.com/health
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Validation errors | SkipRetry in logs | Check task payload | Fix enqueuing code |
| Dependency failure | Connection errors | Check deps | Fix dependency |
| Timeout | Context deadline | Check job duration | Increase timeout |
| Bug in handler | Panic/error | Stack trace in logs | Fix and deploy |
| Bad data | Specific records fail | Check task payload | Fix data |
| Rate limiting | 429 errors | External API logs | Add backoff |

---

## Remediation

### Immediate Actions

1. **Fix underlying dependency:**
   ```bash
   # If database issue
   docker restart postgres
   
   # If Redis issue
   docker restart redis
   ```

2. **Re-run specific failed jobs:**
   ```bash
   # Re-run a single failed job
   asynq task run <task_id>
   
   # Re-run all jobs in retry queue
   asynq task runall --queue=default --state=retry
   ```

3. **Archive failed jobs (if unfixable):**
   ```bash
   # Archive all failed jobs (moves to archive/dead-letter)
   asynq task archive --queue=default --state=failed
   ```

4. **Deploy fix if code bug:**
   ```bash
   # Build and deploy worker with fix
   make build-worker
   docker-compose up -d worker
   ```

### Post-Incident Actions

1. **Root Cause Analysis:** Document which jobs failed and why
2. **Prevention:** Add better validation, improve error handling
3. **Monitoring:** Add alerting for specific error types

---

## Job Retry Management

### Asynq Retry Configuration

```go
// Jobs are configured with retry options
asynq.ProcessIn(time.Minute),    // Delay before first attempt
asynq.MaxRetry(3),               // Maximum retry attempts
asynq.Timeout(5 * time.Minute),  // Processing timeout
```

### Managing Retries

```bash
# View retry queue
asynq task ls --queue=default --state=retry

# Clear retry queue (all will be archived)
asynq task archive --queue=default --state=retry

# View archived (dead letter) jobs
asynq task ls --queue=default --state=archived

# Delete old archived jobs
asynq task delete --queue=default --state=archived
```

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
| Team Lead | @team-lead | After 15 min (critical) / 1 hour (warning) |

---

## References

- [Asynq Error Handling](https://github.com/hibiken/asynq/wiki/Error-Handling)
- [Job Queue Backlog Runbook](job-queue-backlog.md)
- [High Error Rate Runbook](high-error-rate.md)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |
