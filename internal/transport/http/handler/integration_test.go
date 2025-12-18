package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	sharedMetrics "github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
	"github.com/stretchr/testify/assert"
)

// fakeDB implements DatabaseChecker for integration-style router tests.
type fakeDB struct {
	pingErr error
}

func (f *fakeDB) Ping(ctx context.Context) error {
	return f.pingErr
}

// testLogger returns a discarding logger for use in tests.
func testLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.Discard, nil))
}

type testHTTPMetrics struct {
	requests  *prometheus.CounterVec
	durations *prometheus.HistogramVec
}

func (m *testHTTPMetrics) IncRequest(method, route, status string) {
	m.requests.WithLabelValues(method, route, status).Inc()
}

func (m *testHTTPMetrics) ObserveRequestDuration(method, route string, seconds float64) {
	m.durations.WithLabelValues(method, route).Observe(seconds)
}

func newTestMetricsRegistry() (*prometheus.Registry, sharedMetrics.HTTPMetrics) {
	reg := prometheus.NewRegistry()

	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "route", "status"},
	)

	durations := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(requests)
	reg.MustRegister(durations)

	return reg, &testHTTPMetrics{
		requests:  requests,
		durations: durations,
	}
}

// TestIntegrationRoutes covers /health and /ready through the router with DB ok/fail.
func TestIntegrationRoutes(t *testing.T) {
	healthHandler := NewHealthHandler()
	logger := testLogger()
	metricsReg, httpMetrics := newTestMetricsRegistry()

	t.Run("ready OK", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler, nil)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ready","checks":{"database":"ok"}}}`, rec.Body.String())
	})

	t.Run("ready not ready", func(t *testing.T) {
		db := &fakeDB{pingErr: assert.AnError}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler, nil)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"not_ready","checks":{"database":"failed"}}}`, rec.Body.String())
	})

	t.Run("health ok", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler, nil)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ok"}}`, rec.Body.String())
	})
}

// TestMetricsEndpoint verifies the /metrics endpoint behavior.
func TestMetricsEndpoint(t *testing.T) {
	healthHandler := NewHealthHandler()
	logger := testLogger()
	metricsReg, httpMetrics := newTestMetricsRegistry()
	db := &fakeDB{pingErr: nil}
	readyHandler := NewReadyHandler(db)

	r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler, nil)

	t.Run("metrics endpoint returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("metrics content-type contains text/plain", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		contentType := rec.Header().Get("Content-Type")
		assert.Contains(t, contentType, "text/plain")
	})

	t.Run("metrics contains Go runtime metrics", func(t *testing.T) {
		// First make some requests to generate metrics
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
		}

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		body := rec.Body.String()
		// Prometheus exposition format check
		assert.Contains(t, body, "# HELP")
		assert.Contains(t, body, "# TYPE")
		// Go runtime metrics (from collectors.NewGoCollector)
		assert.Contains(t, body, "go_goroutines")
	})

	t.Run("custom metrics created via factory appear at /metrics", func(t *testing.T) {
		custom := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "custom_events_total",
				Help: "Custom events",
			},
			[]string{},
		)
		metricsReg.MustRegister(custom)
		custom.WithLabelValues().Inc()

		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		body := rec.Body.String()
		assert.Contains(t, body, "custom_events_total")
		assert.Contains(t, body, "# HELP custom_events_total")
	})
}
