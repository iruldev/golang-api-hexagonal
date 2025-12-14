# Runbook Documentation

This directory contains incident response runbooks for the golang-api-hexagonal service.

## Quick Links

| Alert | Severity | Runbook |
|-------|----------|---------|
| HighErrorRate | warning | [high-error-rate.md](high-error-rate.md) |
| HighErrorRateCritical | critical | [high-error-rate.md](high-error-rate.md) |
| HighLatency | warning | [high-latency.md](high-latency.md) |
| HighLatencyCritical | critical | [high-latency.md](high-latency.md) |
| ServiceDown | critical | [service-down.md](service-down.md) |
| DBConnectionExhausted | warning | [db-connection-exhausted.md](db-connection-exhausted.md) |
| DBSlowQueries | warning | [db-slow-queries.md](db-slow-queries.md) |
| JobQueueBacklog | warning | [job-queue-backlog.md](job-queue-backlog.md) |
| JobProcessingStalled | warning | [job-queue-backlog.md](job-queue-backlog.md) |
| JobFailureRate | warning | [job-failure-rate.md](job-failure-rate.md) |
| JobFailureRateCritical | critical | [job-failure-rate.md](job-failure-rate.md) |

## Severity Mapping

| Severity | Response Time | Escalation | Examples |
|----------|---------------|------------|----------|
| **critical** | Immediate | After 5-15 min | ServiceDown, HighErrorRateCritical |
| **warning** | Within 30 min | After 1 hour | HighLatency, DBSlowQueries |

## Runbook Structure

Each runbook follows a standardized template with these sections:

1. **Metadata** - Alert names, severity, component
2. **Overview** - What triggers the alert and business impact
3. **Quick Response Checklist** - Rapid incident response steps
4. **Symptoms** - Observable indicators and metrics to check
5. **Diagnosis** - Step-by-step investigation procedures
6. **Common Causes** - Table of causes, symptoms, and resolutions
7. **Remediation** - Immediate actions and post-incident steps
8. **Escalation** - Timeline, path, and contacts

## Alert Categories

### HTTP Service Alerts
- [High Error Rate](high-error-rate.md) - 5xx error rate elevation
- [High Latency](high-latency.md) - Response time degradation
- [Service Down](service-down.md) - Complete service outage

### Database Alerts
- [DB Connection Exhausted](db-connection-exhausted.md) - Connection pool issues
- [DB Slow Queries](db-slow-queries.md) - Query performance problems

### Job Queue Alerts (Asynq)
- [Job Queue Backlog](job-queue-backlog.md) - Processing delays
- [Job Failure Rate](job-failure-rate.md) - Job execution failures

## Creating New Runbooks

### 1. Copy the Template

```bash
cp docs/runbook/template.md docs/runbook/your-alert-name.md
```

### 2. Fill in Required Sections

1. Update **Metadata** table with alert details
2. Write clear **Overview** explaining the alert
3. Define **Symptoms** and metrics to check
4. Document **Diagnosis** steps with actual commands
5. Fill **Common Causes** table from experience
6. Add **Remediation** steps (immediate and post-incident)
7. Update **Escalation** contacts and timeline

### 3. Link from alerts.yaml

```yaml
annotations:
  runbook_url: "docs/runbook/your-alert-name.md"
```

### 4. Update This Index

Add your new runbook to the Quick Links table above.

## Best Practices

### Writing Effective Runbooks

- **Be specific** - Include exact commands, not just descriptions
- **Test commands** - Verify all diagnostic commands work
- **Update regularly** - Keep runbooks current after incidents
- **Cross-reference** - Link related runbooks
- **Include examples** - Show expected vs problematic output

### During Incidents

1. **Start with Quick Response Checklist** - For rapid initial response
2. **Follow Diagnosis steps in order** - They're prioritized
3. **Document findings** - Update runbook if you learn something new
4. **Time-box investigation** - Escalate if not resolved within timeline

### After Incidents

1. **Root Cause Analysis** - Document what happened
2. **Update Runbook** - Add new learnings
3. **Improve Monitoring** - Add alerts if blind spots found

## Related Resources

- [Prometheus Alerting Rules](../../deploy/prometheus/alerts.yaml)
- [Grafana Dashboards](../../deploy/grafana/dashboards/)
- [Architecture Documentation](../architecture.md)

---

*Last Updated: 2025-12-14*
