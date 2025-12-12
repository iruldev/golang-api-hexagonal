# Story 8.7: Create Grafana Dashboard Template

Status: done

## Story

As a SRE,
I want a Grafana dashboard template for the service,
So that I can visualize metrics immediately.

## Acceptance Criteria

### AC1: Grafana Dashboard Exists
**Given** `deploy/grafana/dashboards/service.json` exists
**When** I import to Grafana
**Then** dashboard shows HTTP golden signals (latency, traffic, errors)
**And** dashboard shows job processing rate and duration

> [!NOTE]
> DB latency/connection pool and job queue depth metrics require additional instrumentation (future story).

---

## Tasks / Subtasks

- [x] **Task 1: Add Grafana to docker-compose** (AC: #1)
  - [x] Add Grafana service (port 3000)
  - [x] Mount dashboards directory
  - [x] Configure Prometheus datasource auto-provisioning
  - [x] Add `grafana_data` volume to `volumes:` section
  - [x] Add health check for Grafana service

- [x] **Task 2: Create Grafana provisioning files** (AC: #1)
  - [x] Create `deploy/grafana/provisioning/datasources/prometheus.yml`
  - [x] Create `deploy/grafana/provisioning/dashboards/dashboards.yml`

- [x] **Task 3: Create service dashboard** (AC: #1)
  - [x] Create `deploy/grafana/dashboards/service.json`
  - [x] Add HTTP traffic panel (requests/sec by status)
  - [x] Add HTTP latency panel (p50/p95/p99)
  - [x] Add HTTP error rate panel
  - [x] Add job processing panel (jobs/sec by status)
  - [x] Add job duration panel (p50/p95)

---

## Dev Notes

### Architecture Placement

```
deploy/
├── grafana/
│   ├── dashboards/
│   │   └── service.json           # Main Grafana dashboard
│   └── provisioning/
│       ├── datasources/
│       │   └── prometheus.yml     # Prometheus datasource config
│       └── dashboards/
│           └── dashboards.yml     # Dashboard auto-provisioning
└── prometheus/
    └── prometheus.yml
```

---

### Existing Metrics to Visualize

From `internal/observability/metrics.go`:

**HTTP Metrics:**
- `http_requests_total{method, path, status}` - Counter
- `http_request_duration_seconds{method, path}` - Histogram

**Job Metrics:**
- `job_processed_total{task_type, queue, status}` - Counter
- `job_duration_seconds{task_type, queue}` - Histogram

---

### Grafana Datasource Provisioning

```yaml
# deploy/grafana/provisioning/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
```

---

### Grafana Dashboard Provisioning

```yaml
# deploy/grafana/provisioning/dashboards/dashboards.yml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /var/lib/grafana/dashboards
```

---

### docker-compose Grafana Service

```yaml
grafana:
  image: grafana/grafana:11.4.0
  container_name: golang-api-grafana
  restart: unless-stopped
  ports:
    - "3000:3000"
  environment:
    - GF_SECURITY_ADMIN_USER=admin
    - GF_SECURITY_ADMIN_PASSWORD=admin
    - GF_USERS_ALLOW_SIGN_UP=false
  volumes:
    - ./deploy/grafana/provisioning:/etc/grafana/provisioning:ro
    - ./deploy/grafana/dashboards:/var/lib/grafana/dashboards:ro
    - grafana_data:/var/lib/grafana
  healthcheck:
    test: ["CMD-SHELL", "wget --spider -S http://localhost:3000/api/health 2>&1 | grep '200 OK' || exit 1"]
    interval: 10s
    timeout: 5s
    retries: 3
  networks:
    - app-network
  depends_on:
    - prometheus
```

> [!WARNING]
> Default credentials (admin/admin) are for **development only**.
> For production, set `GF_SECURITY_ADMIN_PASSWORD` via secrets or use provisioning.

---

### Dashboard Panel PromQL Queries

**HTTP Request Rate:**
```promql
sum(rate(http_requests_total[5m])) by (status)
```

**HTTP Latency p95:**
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

**HTTP Error Rate:**
```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) * 100
```

**Job Processing Rate:**
```promql
sum(rate(job_processed_total[5m])) by (status)
```

**Job Duration p95:**
```promql
histogram_quantile(0.95, rate(job_duration_seconds_bucket[5m]))
```

---

### Dashboard JSON Skeleton

Use this skeletal template for `service.json`:

```json
{
  "annotations": { "list": [] },
  "editable": false,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 0,
  "id": null,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "reqps" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
      "id": 1,
      "options": { "legend": { "displayMode": "list" }, "tooltip": { "mode": "single" } },
      "targets": [
        {
          "expr": "sum(rate(http_requests_total[5m])) by (status)",
          "legendFormat": "{{status}}",
          "refId": "A"
        }
      ],
      "title": "HTTP Request Rate",
      "type": "timeseries"
    },
    {
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "s" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 0 },
      "id": 2,
      "options": { "legend": { "displayMode": "list" }, "tooltip": { "mode": "single" } },
      "targets": [
        {
          "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket[5m]))",
          "legendFormat": "p50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
          "legendFormat": "p95",
          "refId": "B"
        },
        {
          "expr": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))",
          "legendFormat": "p99",
          "refId": "C"
        }
      ],
      "title": "HTTP Latency",
      "type": "timeseries"
    },
    {
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "percent" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
      "id": 3,
      "options": { "legend": { "displayMode": "list" }, "tooltip": { "mode": "single" } },
      "targets": [
        {
          "expr": "sum(rate(http_requests_total{status=~\"5..\"}[5m])) / sum(rate(http_requests_total[5m])) * 100",
          "legendFormat": "Error Rate %",
          "refId": "A"
        }
      ],
      "title": "HTTP Error Rate",
      "type": "timeseries"
    },
    {
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "ops" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
      "id": 4,
      "options": { "legend": { "displayMode": "list" }, "tooltip": { "mode": "single" } },
      "targets": [
        {
          "expr": "sum(rate(job_processed_total[5m])) by (status)",
          "legendFormat": "{{status}}",
          "refId": "A"
        }
      ],
      "title": "Job Processing Rate",
      "type": "timeseries"
    },
    {
      "datasource": { "type": "prometheus", "uid": "prometheus" },
      "fieldConfig": {
        "defaults": { "color": { "mode": "palette-classic" }, "unit": "s" },
        "overrides": []
      },
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 16 },
      "id": 5,
      "options": { "legend": { "displayMode": "list" }, "tooltip": { "mode": "single" } },
      "targets": [
        {
          "expr": "histogram_quantile(0.50, rate(job_duration_seconds_bucket[5m]))",
          "legendFormat": "p50",
          "refId": "A"
        },
        {
          "expr": "histogram_quantile(0.95, rate(job_duration_seconds_bucket[5m]))",
          "legendFormat": "p95",
          "refId": "B"
        }
      ],
      "title": "Job Duration",
      "type": "timeseries"
    }
  ],
  "refresh": "30s",
  "schemaVersion": 39,
  "tags": ["golang-api", "service"],
  "templating": { "list": [] },
  "time": { "from": "now-1h", "to": "now" },
  "timepicker": {},
  "timezone": "browser",
  "title": "Golang API Service",
  "uid": "golang-api-service",
  "version": 1,
  "weekStart": ""
}
```

> [!TIP]
> Dashboard supports template variables for dynamic filtering. Add variables in Grafana UI after import if needed.

---

### Verification

1. Run `docker compose up -d`
2. Access http://localhost:3000 (admin/admin)
3. Navigate to Dashboards -> Golang API Service
4. Verify panels render (may need running app for data)

---

### File List

**Create:**
- `deploy/grafana/provisioning/datasources/prometheus.yml`
- `deploy/grafana/provisioning/dashboards/dashboards.yml`
- `deploy/grafana/dashboards/service.json`
- `deploy/grafana/README.md`

**Modify:**
- `docker-compose.yaml` - Add Grafana service + volume
- `docs/sprint-artifacts/8-7-create-grafana-dashboard-template.md` - Story file
- `docs/sprint-artifacts/sprint-status.yaml` - Status tracking

---

## Dev Agent Record

### Agent Model Used
Gemini 2.5 Pro (implementation), Claude 4.5 Sonnet (code review)

### Completion Notes
- Implemented Grafana service in docker-compose.yaml with health check
- Created Prometheus datasource auto-provisioning with explicit UID
- Created dashboard provisioning for auto-loading dashboards
- Built service.json dashboard with 5 panels:
  - HTTP Request Rate (by status code)
  - HTTP Latency (p50/p95/p99 percentiles)
  - HTTP Error Rate (5xx percentage)
  - Job Processing Rate (by status)
  - Job Duration (p50/p95 percentiles)
- All PromQL queries verified against existing metrics in `internal/observability/metrics.go`
- Docker-compose validation passed
- JSON syntax validation passed
- Added README.md with usage documentation and production security guidance

### Code Review Fixes (2025-12-12)
- Fixed datasource UID coordination between prometheus.yml and service.json
- Added `deploy/grafana/README.md` with usage documentation
- Updated AC1 to reflect actually available metrics (DB/queue metrics not yet instrumented)
- Updated File List to include all changed files
- Added production credentials guidance in README

### Change Log
- 2025-12-12: Implemented Story 8.7 - Added Grafana dashboard template with observability panels
- 2025-12-12: Code review fixes - Added README, fixed datasource UID, updated documentation
