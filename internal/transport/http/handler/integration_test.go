package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
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

// TestIntegrationRoutes covers /health and /ready through the router with DB ok/fail.
func TestIntegrationRoutes(t *testing.T) {
	healthHandler := NewHealthHandler()
	logger := testLogger()
	metricsReg, httpMetrics := observability.NewMetricsRegistry()

	t.Run("ready OK", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ready","checks":{"database":"ok"}}}`, rec.Body.String())
	})

	t.Run("ready not ready", func(t *testing.T) {
		db := &fakeDB{pingErr: assert.AnError}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"not_ready","checks":{"database":"failed"}}}`, rec.Body.String())
	})

	t.Run("health ok", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler)

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
	metricsReg, httpMetrics := observability.NewMetricsRegistry()
	db := &fakeDB{pingErr: nil}
	readyHandler := NewReadyHandler(db)

	r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, healthHandler, readyHandler)

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
}
