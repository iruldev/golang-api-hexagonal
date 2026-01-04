# Runbook Template

**Severity:** [Critical | High | Medium | Low]
**Last Updated:** YYYY-MM-DD
**Owner:** [DevOps Team]
**Related ADRs:** [Link to relevant ADRs]

---

## Overview

Brief description of the incident this runbook addresses. What triggers this runbook? What is the impact?

---

## Prerequisites

- **Access Required:**
  - [ ] Kubernetes cluster access (`kubectl`)
  - [ ] Database admin access
  - [ ] Monitoring dashboard access

- **Tools Required:**
  - [ ] `kubectl` CLI
  - [ ] `psql` or database client
  - [ ] Grafana/Prometheus access

- **Knowledge Required:**
  - Understanding of [relevant architecture component]
  - Familiarity with [relevant runbook dependencies]

---

## Symptoms

### 🔔 Alerts

| Alert Name | Condition | Dashboard Link |
|------------|-----------|----------------|
| [Alert Name] | [Threshold/Condition] | [Link to dashboard] |

### 📊 Metrics

- [Metric name] shows [abnormal pattern]
- [Counter/gauge] exceeds [threshold]

### 🐛 User Reports

- [Specific error messages users may report]
- [Degraded functionality description]

---

## Investigation Steps

### Step 1: Initial Assessment

```bash
# Command to run
[command here]
```

**Expected Output:** [Description of healthy output]
**Abnormal Output:** [Description indicates problem]

### Step 2: [Next Investigation Step]

```bash
# Command to run
[command here]
```

**What to look for:** [Specific indicators]

### Step 3: [Additional Steps as Needed]

Continue investigation based on findings...

---

## Resolution Steps

### Immediate Mitigation

> ⚡ **Time-Critical:** Complete within [X minutes]

1. [First immediate action]
   ```bash
   # Command
   ```

2. [Second immediate action]
   ```bash
   # Command
   ```

### Root Cause Fix

> 🔧 **Standard Resolution:** Complete after stabilization

1. [First resolution step]
2. [Second resolution step]
3. [Additional steps as needed]

### Post-Fix Validation

1. [Validation step 1]
2. [Validation step 2]

---

## Rollback Procedure

### When to Rollback

- [Condition 1 that triggers rollback]
- [Condition 2 that triggers rollback]

### Rollback Steps

1. [First rollback step]
   ```bash
   # Command
   ```

2. [Second rollback step]
   ```bash
   # Command
   ```

### Verify Rollback Success

- [ ] [Verification check 1]
- [ ] [Verification check 2]

---

## Verification

### Confirm Resolution

- [ ] [Metric/indicator] has returned to normal
- [ ] No errors in logs for [X minutes]
- [ ] User-facing functionality restored

### Metrics to Monitor

| Metric | Expected Value | Dashboard |
|--------|---------------|-----------|
| [Metric 1] | [Normal value] | [Link] |
| [Metric 2] | [Normal value] | [Link] |

### Observation Period

⏱️ **Recommended:** Monitor for [X minutes/hours] before closing incident

---

## Escalation

| Level | Contact | Criteria | Response Time |
|-------|---------|----------|---------------|
| L1 | On-call engineer | First response | Immediate |
| L2 | Senior DevOps | No resolution in 30 min | 15 min |
| L3 | Team Lead / Architect | Critical business impact | 10 min |

### Escalation Triggers

- ⚠️ Escalate to L2 if: [Condition]
- 🚨 Escalate to L3 if: [Condition]

---

## Related Runbooks

- [Link to related runbook 1]
- [Link to related runbook 2]

---

## Change Log

| Date | Author | Change |
|------|--------|--------|
| YYYY-MM-DD | [Name] | Initial creation |
