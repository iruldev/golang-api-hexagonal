# Runbook: Job Failure Rate Alert

## Overview
This runbook addresses alerts for elevated job failure rates.

## Alerts Covered
- `JobFailureRate` - Warning: Job failures > 10% for 5 minutes
- `JobFailureRateCritical` - Critical: Job failures > 25% for 2 minutes

## Investigation Steps

### 1. Check Worker Logs
```bash
# Check for errors in worker logs
docker logs worker 2>&1 | grep -i error | tail -50
```

### 2. Identify Failing Tasks
```bash
# Check failed jobs in Asynq
asynq task ls --queue=default --state=failed

# View specific failed task
asynq task inspect <task_id>
```

### 3. Check Job Metrics
```bash
curl -s http://localhost:8080/metrics | grep 'job_processed_total{.*status="failed"}'
```

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Validation errors | SkipRetry in logs | Fix payload |
| Dependency failure | Connection errors | Check dependencies |
| Timeout | Context deadline | Increase timeout |
| Bug in handler | Panic/error | Fix and deploy |

## Resolution Actions

1. **If validation error**: Fix the enqueuing code
2. **If dependency failure**: Check DB/Redis/external services
3. **If timeout**: Increase job timeout or optimize processing
4. **If code bug**: Fix handler code and deploy

## Retry Management
```bash
# Re-run failed jobs
asynq task run <task_id>

# Archive old failed jobs
asynq task archive --queue=default --state=failed
```

## Escalation
- **Warning**: Investigate within 30 minutes
- **Critical**: Immediate escalation to on-call engineer
