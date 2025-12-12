# Story 8.6: Add Prometheus to docker-compose

Status: done

## Story

As a developer,
I want Prometheus running in local development,
So that I can develop and test metrics.

## Acceptance Criteria

### AC1: Prometheus in docker-compose
**Given** `docker-compose.yaml` is updated
**When** I run `make dev`
**Then** Prometheus starts on port 9090
**And** Prometheus scrapes application `/metrics`
**And** basic scrape config is provided

---

## Tasks / Subtasks

- [x] **Task 1: Create Prometheus config file** (AC: #1)
  - [x] Create `deploy/prometheus/prometheus.yml`
  - [x] Configure scrape interval (15s)
  - [x] Add job for application `/metrics` endpoint
  - [x] Use `host.docker.internal` for macOS/Windows Docker

- [x] **Task 2: Add Prometheus to docker-compose** (AC: #1)
  - [x] Add Prometheus service (prom/prometheus:v2.54.1)
  - [x] Mount config file
  - [x] Expose port 9090
  - [x] Add to app-network
  - [x] Add health check
  - [x] Add extra_hosts for Linux compatibility

- [x] **Task 3: Update help/documentation** (AC: #1)
  - [x] Prometheus URL included in docker-compose comments
  - [x] Documentation in story file verification section

---

## Dev Notes

### Architecture Placement

```
deploy/
└── prometheus/
    └── prometheus.yml    # Prometheus scrape configuration
```

---

### Prometheus Configuration Pattern

```yaml
# deploy/prometheus/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'golang-api'
    static_configs:
      # host.docker.internal works on macOS/Windows Docker
      # For Linux, use host networking or container IP
      - targets: ['host.docker.internal:8080']
    metrics_path: '/metrics'
    scheme: 'http'
```

---

### docker-compose Prometheus Service

```yaml
prometheus:
  image: prom/prometheus:v2.54.1
  container_name: golang-api-prometheus
  restart: unless-stopped
  ports:
    - "9090:9090"
  volumes:
    - ./deploy/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    - prometheus_data:/prometheus
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--storage.tsdb.path=/prometheus'
    - '--web.console.libraries=/etc/prometheus/console_libraries'
    - '--web.console.templates=/etc/prometheus/consoles'
    - '--web.enable-lifecycle'
  healthcheck:
    test: ["CMD", "wget", "-q", "--spider", "http://localhost:9090/-/healthy"]
    interval: 10s
    timeout: 5s
    retries: 3
  networks:
    - app-network
  extra_hosts:
    - "host.docker.internal:host-gateway"
```

---

### Existing docker-compose Services

- `postgres` - Port 5432
- `redis` - Port 6379
- `jaeger` - Port 16686 (UI), 4317 (OTLP)

**Adding:** `prometheus` - Port 9090

---

### Verification

After implementation:
1. Run `docker compose up -d`
2. Access http://localhost:9090/targets
3. Verify `golang-api` job shows "UP"
4. Query `up{job="golang-api"}` - should return 1

---

### File List

**Create:**
- `deploy/prometheus/prometheus.yml`

**Modify:**
- `docker-compose.yaml` - Add Prometheus service + volume

---

## Dev Agent Record

### Agent Model Used
{{agent_model_name_version}}

### Completion Notes
- Prometheus needs app running to scrape - manual verification required
- Use `host.docker.internal` for Docker Desktop (macOS/Windows)
- For Linux, may need `extra_hosts` or host network mode
