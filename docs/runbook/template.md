# Runbook Template

> **Use this template to create new runbooks for alerts.**
> Copy this file, rename it to match your alert, and fill in the sections.

## Metadata

| Field | Value |
|-------|-------|
| **Alert Name(s)** | `AlertName`, `AlertNameCritical` (if applicable) |
| **Severity** | warning / critical |
| **Component** | http / database / worker |
| **Last Updated** | YYYY-MM-DD |
| **Author** | [Your Name/Team] |

---

## Overview

Brief description of what this alert monitors and why it matters.

**What triggers this alert:**
- Condition that causes the alert to fire

**Business impact:**
- How this affects users or the system

---

## Quick Response Checklist

Use this checklist for rapid incident response:

- [ ] **Acknowledge** - Acknowledge the alert in your monitoring system
- [ ] **Assess** - Determine current impact and scope
- [ ] **Investigate** - Follow Investigation Steps below
- [ ] **Mitigate** - Apply immediate remediation
- [ ] **Communicate** - Update stakeholders if needed
- [ ] **Document** - Record findings and actions taken
- [ ] **Resolve** - Confirm resolution and close alert

---

## Symptoms

### Observable Indicators
- System behavior indicating the issue
- User-reported problems

### Metrics to Check
| Metric | Normal Range | Alert Threshold |
|--------|--------------|-----------------|
| `metric_name` | < X | > Y |

### Related Alerts
- Other alerts that may fire together
- Upstream/downstream dependencies

---

## Diagnosis

### Step 1: Initial Assessment

**Goal:** Understand the scope and impact

```bash
# Quick health check commands
# Example: Check service status
```

**Expected outcome:** Describe what healthy vs unhealthy looks like

### Step 2: Check Dependencies

**Goal:** Verify dependent services

```bash
# Commands to check dependencies
```

### Step 3: Analyze Logs

**Goal:** Find root cause indicators

```bash
# Log analysis commands
# Example: grep for errors
```

### Step 4: Check Metrics

**Goal:** Identify trends and correlations

```bash
# Prometheus/Grafana queries
```

---

## Common Causes

| Cause | Symptoms | Diagnosis | Resolution |
|-------|----------|-----------|------------|
| Cause 1 | What you see | How to confirm | Quick fix |
| Cause 2 | What you see | How to confirm | Quick fix |
| Cause 3 | What you see | How to confirm | Quick fix |

---

## Remediation

### Immediate Actions

Actions to take right now to mitigate impact:

1. **Action 1:** Description of first mitigation step
   ```bash
   # Commands if applicable
   ```

2. **Action 2:** Description of second mitigation step

3. **Action 3:** If above fails, try this

### Post-Incident Actions

After the immediate issue is resolved:

1. **Root Cause Analysis:** Document what caused the issue
2. **Prevention:** Implement changes to prevent recurrence
3. **Monitoring:** Add/improve alerting if needed

### Rollback Procedures

If changes made during remediation need to be reverted:

```bash
# Rollback commands
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
3. **Level 2 Escalation:** Engineering manager / Architect
4. **Executive Escalation:** CTO (major incidents only)

### Escalation Contacts

| Role | Contact | When to Contact |
|------|---------|-----------------|
| On-call Engineer | [PagerDuty/Slack channel] | First response |
| Team Lead | [Contact info] | After 15 min (critical) / 1 hour (warning) |
| DBA (if DB issue) | [Contact info] | Database-related issues |

### Communication Templates

**Slack Update Template:**
```
🚨 Alert: [Alert Name]
Status: Investigating / Mitigating / Resolved
Impact: [Brief impact description]
ETA: [Estimated resolution time]
Lead: [Your name]
```

---

## References

- [Grafana Dashboard](link-to-dashboard)
- [Architecture Documentation](docs/architecture.md)
- [Related Runbook 1](docs/runbook/related.md)
- [External Service Documentation](link)

---

## Revision History

| Date | Author | Change |
|------|--------|--------|
| YYYY-MM-DD | [Name] | Initial version |
