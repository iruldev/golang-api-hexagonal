# Operational Runbooks

Quick reference for handling common operational scenarios. Each runbook provides step-by-step instructions for incident response, resolution, and verification.

---

## 📋 Quick Navigation

| Severity | Runbooks |
|----------|----------|
| 🔴 Critical | [Database Connection Failure](#-critical-immediate-response-required) |
| 🟠 High | [Circuit Breaker Open](#-high-response-within-15-minutes), [Memory/CPU Spike](#-high-response-within-15-minutes) |
| 🟡 Medium | [Deployment Rollback](#-medium-response-within-1-hour), [Secret Rotation](#-medium-response-within-1-hour) |

---

## By Severity

### 🔴 Critical (Immediate Response Required)

> **Response Time:** Immediate (< 5 minutes)

| Runbook | Description | Key Metrics |
|---------|-------------|-------------|
| [Database Connection Failure](./database-connection-failure.md) | Complete loss of database connectivity | `db_connection_errors`, health check status |

---

### 🟠 High (Response Within 15 Minutes)

> **Response Time:** Within 15 minutes

| Runbook | Description | Key Metrics |
|---------|-------------|-------------|
| [Circuit Breaker Open](./circuit-breaker-open.md) | Downstream service failures triggering circuit breaker | `circuit_breaker_state`, `RES-001` errors |
| [Memory/CPU Spike](./memory-cpu-spike.md) | Resource exhaustion affecting service performance | Pod CPU/memory usage, response latency |

---

### 🟡 Medium (Response Within 1 Hour)

> **Response Time:** Within 1 hour (procedural runbooks)

| Runbook | Description | Use Case |
|---------|-------------|----------|
| [Deployment Rollback](./deployment-rollback.md) | Procedure for rolling back a failed deployment | Failed deployment recovery |
| [Secret Rotation](./secret-rotation.md) | Procedure for rotating secrets and credentials | Scheduled or emergency rotation |

---

## 🔧 Creating New Runbooks

1. Copy the [template](./template.md)
2. Fill in all required sections (do not skip any)
3. Add to this index with appropriate severity
4. Cross-reference related runbooks and ADRs
5. Have the runbook reviewed by on-call team

### Required Sections

Every runbook MUST include:

- ✅ **Overview:** What the runbook addresses
- ✅ **Prerequisites:** Access and tools needed
- ✅ **Symptoms:** How to recognize the issue
- ✅ **Investigation Steps:** Diagnostic commands
- ✅ **Resolution Steps:** How to fix
- ✅ **Rollback Procedure:** How to undo changes
- ✅ **Verification:** Confirm issue is resolved
- ✅ **Escalation:** Who to contact and when

---

## 📚 Related Documentation

- [Architecture Decision Records](../adr/index.md) - Understand why things are built this way
- [Local Development Guide](../local-development.md) - Development environment setup
- [Error Codes Reference](../error-codes.md) - Error code taxonomy
- [Observability Guide](../observability.md) - Metrics and tracing

---

## 📞 Emergency Contacts

| Role | Contact Method | When to Contact |
|------|---------------|-----------------|
| On-call Engineer | PagerDuty | Any incident |
| DevOps Lead | Slack #incidents | L2 escalation |
| Engineering Manager | Phone/Slack | Critical business impact |

---

## 🔄 Runbook Maintenance

- **Review Schedule:** Quarterly
- **Last Review:** 2026-01-04
- **Next Review:** 2026-04-04
- **Owner:** DevOps Team
