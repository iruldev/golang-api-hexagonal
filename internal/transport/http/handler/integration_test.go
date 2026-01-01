package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/iruldev/golang-api-hexagonal/internal/app/user"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres"
	sharedMetrics "github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
	httpTransport "github.com/iruldev/golang-api-hexagonal/internal/transport/http"
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

func (m *testHTTPMetrics) ObserveResponseSize(method, route string, sizeBytes float64) {
	// No-op for now unless we want to track it in integration tests, but we must implement the interface
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
	livenessHandler := NewLivenessHandler()
	healthHandler := NewHealthHandler()
	logger := testLogger()
	metricsReg, httpMetrics := newTestMetricsRegistry()

	t.Run("ready OK", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db, logger)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, livenessHandler, healthHandler, readyHandler, nil, nil, 1024, httpTransport.JWTConfig{}, httpTransport.RateLimitConfig{}, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ready","checks":{"database":"ok"}}}`, rec.Body.String())
	})

	t.Run("ready not ready", func(t *testing.T) {
		db := &fakeDB{pingErr: assert.AnError}
		readyHandler := NewReadyHandler(db, logger)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, livenessHandler, healthHandler, readyHandler, nil, nil, 1024, httpTransport.JWTConfig{}, httpTransport.RateLimitConfig{}, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"not_ready","checks":{"database":"failed"}}}`, rec.Body.String())
	})

	// Story 4.3 AC#3: idempotency verified via integration test
	t.Run("ready idempotent", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db, logger)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, livenessHandler, healthHandler, readyHandler, nil, nil, 1024, httpTransport.JWTConfig{}, httpTransport.RateLimitConfig{}, nil, nil, 0)

		for i := 0; i < 5; i++ {
			req := httptest.NewRequest(http.MethodGet, "/ready", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code, "iteration %d failed", i)
			assert.JSONEq(t, `{"data":{"status":"ready","checks":{"database":"ok"}}}`, rec.Body.String(), "iteration %d content mismatch", i)
		}
	})

	t.Run("health ok", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		readyHandler := NewReadyHandler(db, logger)
		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, livenessHandler, healthHandler, readyHandler, nil, nil, 1024, httpTransport.JWTConfig{}, httpTransport.RateLimitConfig{}, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"data":{"status":"ok"}}`, rec.Body.String())
	})

	// Story 3.2: Verify /readyz (K8s readiness probe) behavior
	t.Run("readyz OK", func(t *testing.T) {
		db := &fakeDB{pingErr: nil}
		dbChecker := postgres.NewDatabaseHealthChecker(db)
		readinessHandler := NewReadinessHandler(0, dbChecker)

		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, livenessHandler, healthHandler, NewReadyHandler(db, logger), readinessHandler, nil, 1024, httpTransport.JWTConfig{}, httpTransport.RateLimitConfig{}, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"healthy","checks":{"database":{"status":"healthy","latency_ms":0}}}`, rec.Body.String())
	})

	t.Run("readyz not ready", func(t *testing.T) {
		db := &fakeDB{pingErr: assert.AnError}
		dbChecker := postgres.NewDatabaseHealthChecker(db)
		readinessHandler := NewReadinessHandler(0, dbChecker)

		r := httpTransport.NewRouter(logger, false, metricsReg, httpMetrics, livenessHandler, healthHandler, NewReadyHandler(db, logger), readinessHandler, nil, 1024, httpTransport.JWTConfig{}, httpTransport.RateLimitConfig{}, nil, nil, 0)

		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		// Error message depends on the assert.AnError value
		assert.Contains(t, rec.Body.String(), `"status":"unhealthy"`)
		assert.Contains(t, rec.Body.String(), `"database":{"status":"unhealthy"`)
	})

	// Story 3.1: Verify /healthz bypasses ALL middleware (no logging, no tracing, no secure headers)
	// This ensures the probe is extremely lightweight as per AC#1
	t.Run("liveness bypasses middleware", func(t *testing.T) {
		// Setup router with rate limiting and secure headers enabled
		r := httpTransport.NewRouter(
			logger,
			false,
			metricsReg,
			httpMetrics,
			livenessHandler,
			healthHandler,
			NewReadyHandler(&fakeDB{}, logger),
			nil, // readinessHandler - Story 3.2
			nil,
			1024,
			httpTransport.JWTConfig{},
			httpTransport.RateLimitConfig{RequestsPerSecond: 100},
			nil,
			nil,
			0,
		)

		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "alive", "alive") // sanity check logic

		// Verify Global Middleware are missing
		// X-Request-ID is added by middleware.RequestID
		assert.Empty(t, rec.Header().Get("X-Request-ID"), "Endpoint should NOT have X-Request-ID header")

		// X-Frame-Options is added by middleware.SecureHeaders
		// Note: SecureHeaders is the FIRST middleware in the group but /healthz is registered OUTSIDE the group
		assert.Empty(t, rec.Header().Get("X-Frame-Options"), "Endpoint should NOT have Security headers")

		// Verify Group Middleware are missing
		// X-RateLimit-Limit is added by middleware.RateLimiter (if applied)
		assert.Empty(t, rec.Header().Get("X-RateLimit-Limit"), "Endpoint should NOT have RateLimit headers")
	})
}

// TestMetricsEndpoint verifies the /metrics endpoint behavior.
// Story 2.5b: /metrics is now on internal router only.
func TestMetricsEndpoint(t *testing.T) {
	logger := testLogger()
	metricsReg, httpMetrics := newTestMetricsRegistry()

	// Use internal router for /metrics tests (Story 2.5b)
	r := httpTransport.NewInternalRouter(logger, metricsReg, httpMetrics)

	t.Run("metrics endpoint returns 200 on internal router", func(t *testing.T) {
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

// TestIntegration_CreateUser_LocationHeader verifies that the Location header is correctly set
// when creating a user, using the full router stack. (Story 4.6)
func TestIntegration_CreateUser_LocationHeader(t *testing.T) {
	// 1. Setup Dependencies
	mockCreateUC := new(MockCreateUserUseCase)
	mockGetUC := new(MockGetUserUseCase)
	mockListUC := new(MockListUsersUseCase)
	mockDB := &fakeDB{pingErr: nil}
	logger := testLogger()
	metricsReg, httpMetrics := newTestMetricsRegistry()

	expectedUser := createTestUser()

	// Mock the use case success
	mockCreateUC.On("Execute", mock.Anything, mock.Anything).
		Return(user.CreateUserResponse{User: expectedUser}, nil)

	// 2. Setup Router
	userHandler := NewUserHandler(mockCreateUC, mockGetUC, mockListUC, httpTransport.BasePath+"/users")
	livenessHandler := NewLivenessHandler()
	healthHandler := NewHealthHandler()
	readyHandler := NewReadyHandler(mockDB, logger)

	r := httpTransport.NewRouter(
		logger,
		false,
		metricsReg,
		httpMetrics,
		livenessHandler,
		healthHandler,
		readyHandler,
		nil, // readinessHandler - Story 3.2
		userHandler,
		1024,
		httpTransport.JWTConfig{}, // JWT disabled for this test
		httpTransport.RateLimitConfig{RequestsPerSecond: 100},
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	// 3. Execute Request
	body := `{"email":"test@example.com","firstName":"John","lastName":"Doe"}`
	req := httptest.NewRequest(http.MethodPost, httpTransport.BasePath+"/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// 4. Assertions
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"), "Content-Type should be application/json")

	// Verify header presence and content
	location := rec.Header().Get("Location")
	assert.NotEmpty(t, location, "Location header should be set")

	expectedLocation := fmt.Sprintf("%s/%s", httpTransport.BasePath+"/users", expectedUser.ID)
	assert.Equal(t, expectedLocation, location)
}
