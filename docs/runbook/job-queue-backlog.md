# Runbook: Job Queue Backlog Alert

## Overview
This runbook addresses alerts for job processing issues and stalled workers.

## Alerts Covered
- `JobQueueBacklog` - Warning: Success rate < 90% for 10 minutes
- `JobProcessingStalled` - Warning: No jobs processed but recent history for 10 minutes

## Investigation Steps

### 1. Check Worker Status
```bash
# Check worker container
docker ps | grep worker

# Check worker logs
docker logs worker --tail 100
```

### 2. Check Redis (Queue Backend)
```bash
# Check Redis connection
docker exec -it redis redis-cli ping

# Check queue sizes
docker exec -it redis redis-cli LLEN asynq:{default}:pending
```

### 3. Check Job Metrics
```bash
curl -s http://localhost:8080/metrics | grep job_
```

### 4. Common Causes
| Cause | Symptoms | Resolution |
|-------|----------|------------|
| Worker crashed | No processing | Restart worker |
| Redis down | Connection errors | Check Redis |
| Slow jobs | Growing backlog | Scale workers |
| Retry storm | Many retries | Fix failing jobs |

## Resolution Actions

1. **If worker down**: Restart worker service
2. **If Redis issue**: Check Redis connectivity
3. **If backlog**: Scale up workers

## Asynq Commands
```bash
# List queues and their sizes
asynq queue ls

# Inspect pending tasks
asynq task ls --queue=default --state=pending
```

## Escalation
If unresolved after 15 minutes, escalate to on-call engineer.
