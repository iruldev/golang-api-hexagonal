# Panduan Observability

Panduan ini menjelaskan cara mengkonfigurasi dan menggunakan fitur observability di golang-api-hexagonal, termasuk logging, tracing, dan metrics.

## Daftar Isi

- [Overview](#overview)
- [Logging](#logging)
- [Tracing (OpenTelemetry)](#tracing-opentelemetry)
- [Metrics](#metrics)
- [Request Correlation](#request-correlation)
- [Custom Metrics](#custom-metrics)
- [Observability Stack Lokal (Opsional)](#observability-stack-lokal-opsional)

---

## Overview

Golang API Hexagonal menggunakan tiga pilar observability:

| Komponen | Teknologi | Tujuan |
|----------|-----------|--------|
| **Logging** | Go `slog` dengan JSON handler | Structured logging dengan format JSON |
| **Tracing** | OpenTelemetry + OTLP gRPC | Distributed tracing dengan W3C Trace Context |
| **Metrics** | Prometheus client library | Metrics endpoint untuk monitoring |

Semua request memiliki correlation IDs (request ID, trace ID, span ID) yang memungkinkan pelacakan end-to-end dari log ke trace.

---

## Logging

### Konfigurasi

| Environment Variable | Default | Deskripsi |
|---------------------|---------|-----------|
| `LOG_LEVEL` | `info` | Level logging: `debug`, `info`, `warn`, `error` |
| `SERVICE_NAME` | `golang-api-hexagonal` | Nama service yang muncul di setiap log |
| `ENV` | `development` | Environment: `development`, `staging`, `production` |

> [!TIP]
> Gunakan `LOG_LEVEL=debug` untuk troubleshooting di development. Gunakan `LOG_LEVEL=warn` atau `LOG_LEVEL=error` di production untuk mengurangi noise.

### Format Log

Semua log dalam format JSON dengan field standar:

| Field | Deskripsi |
|-------|-----------|
| `time` | Timestamp dalam format RFC3339 |
| `level` | Level log: DEBUG, INFO, WARN, ERROR |
| `msg` | Pesan log |
| `service` | Nama service (dari `SERVICE_NAME`) |
| `env` | Environment (dari `ENV`) |

**Field tambahan per request:**

| Field | Deskripsi |
|-------|-----------|
| `requestId` | Unique ID untuk setiap HTTP request |
| `traceId` | OpenTelemetry trace ID (jika tracing diaktifkan) |
| `spanId` | OpenTelemetry span ID (jika tracing diaktifkan) |

### Contoh Output Log

```json
{
  "time": "2025-12-23T09:30:00.123456789+07:00",
  "level": "INFO",
  "msg": "user created",
  "service": "golang-api-hexagonal",
  "env": "development",
  "requestId": "abc123-def456-789",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "00f067aa0ba902b7"
}
```

### Log Levels

| Level | Kapan Digunakan |
|-------|-----------------|
| `debug` | Detail untuk debugging, tidak untuk production |
| `info` | Alur normal aplikasi (startup, request handling) |
| `warn` | Situasi tidak terduga yang bisa ditangani |
| `error` | Error yang mempengaruhi operasi |

**Perilaku Filtering:**

- `LOG_LEVEL=debug` → Menampilkan semua log (debug, info, warn, error)
- `LOG_LEVEL=info` → Menampilkan info, warn, error
- `LOG_LEVEL=warn` → Menampilkan warn, error
- `LOG_LEVEL=error` → Hanya menampilkan error

---

## Tracing (OpenTelemetry)

### Konfigurasi

| Environment Variable | Default | Deskripsi |
|---------------------|---------|-----------|
| `OTEL_ENABLED` | `false` | Aktifkan OpenTelemetry tracing |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `localhost:4317` | Endpoint OTLP gRPC (Jaeger, Tempo, dll) |
| `OTEL_EXPORTER_OTLP_INSECURE` | `false` | Gunakan plaintext (tanpa TLS) untuk local dev |

> [!IMPORTANT]
> Tracing dinonaktifkan secara default. Set `OTEL_ENABLED=true` untuk mengaktifkan.

### Contoh Konfigurasi

```bash
# Development dengan Jaeger lokal
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_EXPORTER_OTLP_INSECURE=true
```

### W3C Trace Context

Tracer menggunakan W3C Trace Context untuk propagasi trace antar service:

- **Header:** `traceparent`
- **Format:** `{version}-{trace-id}-{parent-id}-{trace-flags}`
- **Contoh:** `00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01`

Ketika request masuk dengan header `traceparent`, trace ID dari upstream service akan digunakan untuk mempertahankan kontinuitas trace.

### Span Attributes

Setiap HTTP span otomatis memiliki atribut berikut:

| Attribute | Deskripsi |
|-----------|-----------|
| `http.method` | HTTP method (GET, POST, dll) |
| `http.route` | Route pattern (contoh: `/api/v1/users/:id`) |
| `http.status_code` | HTTP status code respons |
| `service.name` | Nama service |
| `deployment.environment` | Environment (development, production, dll) |

### Graceful Shutdown

Tracer harus di-shutdown dengan benar untuk memastikan semua span terkirim:

```go
tp, err := observability.InitTracer(ctx, cfg)
if err != nil {
    log.Fatal(err)
}
defer tp.Shutdown(ctx) // Pastikan dipanggil saat shutdown
```

---

## Metrics

### Konfigurasi

Metrics endpoint tersedia di `/metrics` dalam format Prometheus. Tidak ada konfigurasi khusus yang diperlukan - endpoint aktif secara default.

```bash
# Akses metrics endpoint
curl http://localhost:8080/metrics
```

### Built-in HTTP Metrics

| Metric | Type | Labels | Deskripsi |
|--------|------|--------|-----------|
| `http_requests_total` | Counter | `method`, `route`, `status` | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | `method`, `route` | Durasi request dalam detik |

**Contoh Output:**

```prometheus
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",route="/api/v1/users",status="200"} 42
http_requests_total{method="POST",route="/api/v1/users",status="201"} 15

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",route="/api/v1/users",le="0.005"} 10
http_request_duration_seconds_bucket{method="GET",route="/api/v1/users",le="0.01"} 25
http_request_duration_seconds_bucket{method="GET",route="/api/v1/users",le="+Inf"} 42
http_request_duration_seconds_sum{method="GET",route="/api/v1/users"} 0.523
http_request_duration_seconds_count{method="GET",route="/api/v1/users"} 42
```

### Go Runtime Metrics

Otomatis tersedia dari Go collector:

| Metric Pattern | Deskripsi |
|----------------|-----------|
| `go_goroutines` | Jumlah goroutines aktif |
| `go_memstats_alloc_bytes` | Bytes yang dialokasikan dan masih digunakan |
| `go_memstats_heap_alloc_bytes` | Heap bytes yang dialokasikan |
| `process_cpu_seconds_total` | Total CPU seconds yang digunakan |

### Prometheus Scraping

Tambahkan job ke konfigurasi Prometheus untuk scrape metrics:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'golang-api-hexagonal'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: /metrics
    scrape_interval: 15s
```

---

## Request Correlation

### Request ID

Setiap request memiliki unique request ID yang:

1. **Dibuat otomatis** jika tidak ada header `X-Request-ID`
2. **Diteruskan** dari header `X-Request-ID` jika ada
3. **Ditambahkan** ke semua log dalam request lifecycle
4. **Dikembalikan** dalam response header `X-Request-ID`

### Trace Correlation

Ketika OpenTelemetry diaktifkan, setiap log entry memiliki:

| Field | Source |
|-------|--------|
| `requestId` | Request ID middleware |
| `traceId` | OpenTelemetry trace context |
| `spanId` | Current span ID |

### Contoh Log dengan Korelasi Lengkap

```json
{
  "time": "2025-12-23T09:30:00.123456789+07:00",
  "level": "INFO",
  "msg": "user created",
  "service": "golang-api-hexagonal",
  "env": "development",
  "requestId": "abc123-def456-789",
  "traceId": "4bf92f3577b34da6a3ce929d0e0e4736",
  "spanId": "00f067aa0ba902b7"
}
```

### Melacak Request End-to-End

1. **Dari Log:** Gunakan `requestId` atau `traceId` untuk filter log
   ```bash
   # Filter log berdasarkan request ID
   grep "abc123-def456-789" application.log
   
   # Atau gunakan jq untuk JSON logs
   cat application.log | jq 'select(.requestId == "abc123-def456-789")'
   ```

2. **Dari Trace:** Buka Jaeger UI dan cari dengan trace ID
   ```
   http://localhost:16686/trace/{traceId}
   ```

3. **Menghubungkan Log dan Trace:** Gunakan `traceId` dari log untuk menemukan trace di Jaeger

---

## Custom Metrics

### Membuat Custom Metrics

Package `observability` menyediakan factory functions untuk membuat custom metrics:

#### Counter

Counter adalah metric yang hanya bisa naik (atau reset ke nol saat restart). Gunakan untuk menghitung jumlah event.

```go
import "github.com/iruldev/golang-api-hexagonal/internal/infra/observability"

// Dengan error handling
counter, err := observability.NewCounter(registry, 
    "myapp_orders_total", 
    "Total number of orders processed",
    []string{"status"})
if err != nil {
    return err
}

// Gunakan counter
counter.WithLabelValues("completed").Inc()
counter.WithLabelValues("cancelled").Add(5)
```

#### Histogram

Histogram untuk mengukur distribusi nilai (durasi, ukuran, dll).

```go
// Dengan custom buckets
histogram, err := observability.NewHistogram(registry,
    "myapp_request_size_bytes",
    "Size of requests in bytes",
    []string{"endpoint"},
    []float64{100, 500, 1000, 5000, 10000}) // buckets

// Dengan default buckets (nil)
histogram, err := observability.NewHistogram(registry,
    "myapp_processing_duration_seconds",
    "Processing duration",
    []string{"operation"},
    nil) // uses prometheus.DefBuckets

histogram.WithLabelValues("/api/orders").Observe(1024)
```

#### Gauge

Gauge untuk nilai yang bisa naik atau turun (koneksi aktif, queue size, dll).

```go
gauge, err := observability.NewGauge(registry,
    "myapp_active_connections",
    "Number of active connections",
    []string{"pool"})

gauge.WithLabelValues("postgres").Set(10)
gauge.WithLabelValues("postgres").Inc()
gauge.WithLabelValues("postgres").Dec()
```

### Must* Variants

Untuk inisialisasi di main/startup, gunakan `Must*` variants yang panic on error:

```go
// Panic jika registrasi gagal - cocok untuk startup
counter := observability.MustNewCounter(registry, 
    "myapp_users_total", 
    "Total users",
    []string{"status"})
```

> [!CAUTION]
> Gunakan `Must*` variants hanya di initialization paths. Jangan gunakan di request handlers karena bisa menyebabkan panic saat runtime.

### HTTPMetrics Interface

Untuk middleware, gunakan interface `HTTPMetrics` dari package `internal/shared/metrics`:

```go
type HTTPMetrics interface {
    IncRequest(method, route, status string)
    ObserveRequestDuration(method, route string, seconds float64)
}
```

---

## Observability Stack Lokal (Opsional)

Untuk development lokal, Anda bisa menjalankan observability stack dengan Docker Compose.

### docker-compose.observability.yml

```yaml
version: '3.8'

services:
  jaeger:
    image: jaegertracing/all-in-one:1.53
    ports:
      - "16686:16686"  # Jaeger UI
      - "4317:4317"    # OTLP gRPC
      - "4318:4318"    # OTLP HTTP
    environment:
      - COLLECTOR_OTLP_ENABLED=true

  prometheus:
    image: prom/prometheus:v2.48.0
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    extra_hosts:
      - "host.docker.internal:host-gateway"

  grafana:
    image: grafana/grafana:10.2.2
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana-data:/var/lib/grafana

volumes:
  grafana-data:
```

### prometheus.yml (untuk container di atas)

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'golang-api-hexagonal'
    static_configs:
      - targets: ['host.docker.internal:8080']
    metrics_path: /metrics
```

### Menjalankan Observability Stack

```bash
# Jalankan stack
docker-compose -f docker-compose.observability.yml up -d

# Konfigurasi aplikasi untuk mengirim trace ke Jaeger
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_EXPORTER_OTLP_INSECURE=true

# Jalankan aplikasi
make run
```

### Mengakses UI

| Service | URL | Deskripsi |
|---------|-----|-----------|
| Jaeger UI | http://localhost:16686 | Melihat dan query traces |
| Prometheus | http://localhost:9090 | Query metrics dan alerts |
| Grafana | http://localhost:3000 | Dashboard (admin/admin) |

> [!NOTE]
> Stack ini opsional dan hanya untuk development. Di production, gunakan observability platform yang sesuai (Datadog, New Relic, Grafana Cloud, dll).

---

## Referensi Cepat

### Environment Variables

```bash
# Logging
LOG_LEVEL=info                           # debug, info, warn, error
SERVICE_NAME=golang-api-hexagonal
ENV=development

# Tracing
OTEL_ENABLED=false
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_EXPORTER_OTLP_INSECURE=false
```

### Verifikasi

```bash
# Cek metrics endpoint
curl http://localhost:8080/metrics | head -50

# Cek environment variables
grep -E "OTEL|LOG_LEVEL" .env.example

# Test dengan request ID custom
curl -H "X-Request-ID: my-test-id" http://localhost:8080/health
```

### File Implementasi

| File | Deskripsi |
|------|-----------|
| `internal/infra/observability/logger.go` | Structured JSON logger (slog) |
| `internal/infra/observability/tracer.go` | OpenTelemetry tracer |
| `internal/infra/observability/metrics.go` | Prometheus metrics registry |
| `internal/shared/metrics/http_metrics.go` | HTTPMetrics interface |
