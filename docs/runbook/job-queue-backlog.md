# Runbook: Job Queue Backlog

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `JobQueueBacklog`, `JobProcessingStalled` |
| **Severity** | warning |
| **Component** | worker |
| **Last Updated** | 2025-12-14 |
| **Author** | Platform Team |

---

## Overview

This runbook addresses alerts for job processing issues and queue backlog.

**What triggers this alert:**
- **JobQueueBacklog (Warning):** Job success rate < 90% for 10 minutes
- **JobProcessingStalled (Warning):** No jobs processed for 10 minutes but recent history exists

**Business impact:**
- Async operations delayed or not completing
- Users waiting for background tasks (emails, notifications, etc.)
- Data processing delays

> **Deployment Context:** Commands in this runbook assume a Docker/Kubernetes deployment. For local development (where the worker runs via `make worker` on the host), replace Docker commands with direct process inspection (e.g., check application logs in the terminal).

---

## Quick Response Checklist

- [ ] **Acknowledge** - Acknowledge the alert
- [ ] **Assess** - Check worker status and queue depth
- [ ] **Investigate** - Follow diagnosis steps below
- [ ] **Mitigate** - Restart workers or clear stuck jobs
- [ ] **Communicate** - Update team if user-facing async operations affected
- [ ] **Document** - Record findings and actions
- [ ] **Resolve** - Confirm job processing resumes

---

## Symptoms

### Observable Indicators
- Background tasks not completing
- Queue depth growing over time
- Worker containers not processing jobs

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `job_processed_total{status="success"}` rate | > 90% | < 90% |
| Job processing rate | > 0 jobs/min | 0 for 10 min |
| Queue depth | Stable or decreasing | Growing |

### Related Alerts
- `JobFailureRate` - May indicate root cause
- `HighLatency` - If jobs are slow to process

---

## Diagnosis

### Step 1: Check Worker Status

**Goal:** Verify worker containers are running

```bash
# Check worker container
docker ps | grep worker

# Check worker logs for errors
docker logs worker --tail 100 2>&1 | grep -i "error\|panic\|fatal"

# For Kubernetes
kubectl get pods -l app=worker
kubectl logs deployment/worker --tail=100
```

**Expected outcome:** Worker should be running without errors

### Step 2: Check Redis (Queue Backend)

**Goal:** Verify Redis connectivity and queue state

```bash
# Check Redis connection
docker exec -it redis redis-cli ping

# Check queue sizes (asynq uses lists)
docker exec -it redis redis-cli LLEN asynq:{default}:pending
docker exec -it redis redis-cli LLEN asynq:{default}:scheduled
docker exec -it redis redis-cli LLEN asynq:{default}:retry

# Check Redis memory
docker exec -it redis redis-cli INFO memory | grep used_memory_human
```

### Step 3: Check Job Metrics

**Goal:** Identify processing rates and failure patterns

```bash
# Check job metrics
curl -s http://localhost:8080/metrics | grep job_

# Check processed job counts
curl -s http://localhost:8080/metrics | grep 'job_processed_total'
```

### Step 4: Inspect Failed Jobs

**Goal:** Understand why jobs are failing

```bash
# Using asynq CLI (if available)
asynq task ls --queue=default --state=pending
asynq task ls --queue=default --state=retry
asynq task ls --queue=default --state=archived

# Check specific failed task
asynq task inspect <task_id>
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Worker crashed | No processing | Container not running | Restart worker |
| Redis down | Connection errors | Redis ping fails | Check Redis |
| Slow jobs | Growing backlog | Long job duration | Scale workers |
| Retry storm | Many retries | High retry queue | Fix failing jobs |
| Resource exhaustion | OOM/CPU high | Check resources | Increase limits |
| Deadlock | Jobs stuck | No progress | Restart worker |

---

## Remediation

### Immediate Actions

1. **Restart worker service:**
   ```bash
   docker restart worker
   
   # For Kubernetes
   kubectl rollout restart deployment/worker
   ```

2. **Check and restart Redis if needed:**
   ```bash
   docker restart redis
   
   # Verify Redis is accepting connections
   docker exec -it redis redis-cli ping
   ```

3. **Scale up workers:**
   ```bash
   # Kubernetes
   kubectl scale deployment/worker --replicas=3
   
   # Docker Compose - edit docker-compose.yaml and restart
   ```

4. **Clear stuck retry jobs (last resort):**
   ```bash
   # Archive all retry jobs
   asynq task archive --queue=default --state=retry
   
   # Or delete archived jobs older than 24h
   asynq task delete --queue=default --state=archived
   ```

### Post-Incident Actions

1. **Root Cause Analysis:** Identify why jobs stopped processing
2. **Prevention:** Add worker health checks, improve monitoring
3. **Monitoring:** Set up queue depth alerting

---

## Asynq CLI Reference

```bash
# List all queues
asynq queue ls

# Get queue info
asynq queue inspect default

# List tasks by state
asynq task ls --queue=default --state=pending
asynq task ls --queue=default --state=scheduled
asynq task ls --queue=default --state=retry
asynq task ls --queue=default --state=archived
asynq task ls --queue=default --state=active

# Inspect specific task
asynq task inspect <task_id>

# Re-run a failed task
asynq task run <task_id>

# Archive failed tasks
asynq task archive --queue=default --state=failed

# Delete old archived tasks
asynq task delete --queue=default --state=archived
```

---

## Escalation

### Escalation Timeline

| Severity | Initial Response | Escalation If Unresolved |
|----------|------------------|--------------------------|
| **Warning** | Investigate within 30 minutes | Escalate after 1 hour |

### Escalation Path

1. **First Responder:** On-call engineer
2. **Level 1 Escalation:** Team lead / Senior engineer
3. **Level 2 Escalation:** Engineering manager

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | PagerDuty / #oncall | First response |
| Team Lead | @team-lead | After 1 hour |

---

## References

- [Asynq Documentation](https://github.com/hibiken/asynq)
- [Redis Commands](https://redis.io/commands/)
- [Job Failure Rate Runbook](job-failure-rate.md)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| 2025-12-14 | Platform Team | Enhanced with standardized template |
