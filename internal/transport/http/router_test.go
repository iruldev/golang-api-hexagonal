package http

import (
	"context"
	"log/slog"
	stdhttp "net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// --- Mocks ---

type MockHTTPMetrics struct {
	mock.Mock
}

func (m *MockHTTPMetrics) IncRequest(method, route, status string) {
	m.Called(method, route, status)
}

func (m *MockHTTPMetrics) ObserveRequestDuration(method, route string, seconds float64) {
	m.Called(method, route, seconds)
}

func (m *MockHTTPMetrics) ObserveResponseSize(method, route string, sizeBytes float64) {
	m.Called(method, route, sizeBytes)
}

type MockUserRoutes struct {
	mock.Mock
}

func (m *MockUserRoutes) CreateUser(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	m.Called(w, r)
	w.WriteHeader(stdhttp.StatusCreated)
}

func (m *MockUserRoutes) GetUser(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	m.Called(w, r)
	w.WriteHeader(stdhttp.StatusOK)
}

func (m *MockUserRoutes) ListUsers(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	m.Called(w, r)
	w.WriteHeader(stdhttp.StatusOK)
}

// --- Tests ---

func TestNewRouter_JWTEnabled(t *testing.T) {
	// Setup dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	mockUserHandler := new(MockUserRoutes)
	// Expect handlers NOT to be called if auth fails
	// But if auth succeeds (or is disabled), they will be called.

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtSecret := []byte("this-is-a-32-byte-secret-key!!@@")
	jwtConfig := JWTConfig{
		Enabled:   true,
		Secret:    jwtSecret,
		Issuer:    "test-issuer",
		Audience:  "test-audience",
		ClockSkew: time.Minute,
	}

	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 100,
		TrustProxy:        false,
	}

	router := NewRouter(
		logger,
		false, // tracing disabled
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		mockUserHandler,
		1024,
		jwtConfig,
		rateLimitConfig,
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	// Case 1: JWT Enabled, No Token provided -> Should be 401
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, stdhttp.StatusUnauthorized, w.Code, "Expected 401 when JWT is enabled and no token provided")
	mockUserHandler.AssertNotCalled(t, "ListUsers")
}

func TestNewRouter_JWTDisabled(t *testing.T) {
	// Setup dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	mockUserHandler := new(MockUserRoutes)
	mockUserHandler.On("ListUsers", mock.Anything, mock.Anything).Return()

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtConfig := JWTConfig{
		Enabled: false, // Disabled!
	}

	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 100,
		TrustProxy:        false,
	}

	router := NewRouter(
		logger,
		false,
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		mockUserHandler,
		1024,
		jwtConfig,
		rateLimitConfig,
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	// Case 2: JWT Disabled, No Token provided -> Should be 200 (Mock returns 200)
	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, stdhttp.StatusOK, w.Code, "Expected 200 when JWT is disabled and no token provided")
	mockUserHandler.AssertCalled(t, "ListUsers", mock.Anything, mock.Anything)
}

func TestNewRouter_HealthCheck_NoAuth(t *testing.T) {
	// Setup dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	mockUserHandler := new(MockUserRoutes)

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte("OK"))
	})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtConfig := JWTConfig{
		Enabled: true, // Enabled for API, but shouldn't affect /health
		Secret:  []byte("secret"),
	}

	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 100,
	}

	router := NewRouter(
		logger,
		false,
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		mockUserHandler,
		1024,
		jwtConfig,
		rateLimitConfig,
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, stdhttp.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

// =============================================================================
// Story 2.5b: Internal Router Tests
// =============================================================================

// TestNewRouter_MetricsNotExposed tests that /metrics returns 404 on public router
func TestNewRouter_MetricsNotExposed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtConfig := JWTConfig{Enabled: false}
	rateLimitConfig := RateLimitConfig{RequestsPerSecond: 100}

	router := NewRouter(
		logger,
		false,
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		nil, // no user handler
		1024,
		jwtConfig,
		rateLimitConfig,
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	// /metrics should return 404 on public router
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, stdhttp.StatusNotFound, w.Code, "/metrics should not be exposed on public router")
}

// TestNewInternalRouter_MetricsAvailable tests that /metrics returns 200 on internal router
func TestNewInternalRouter_MetricsAvailable(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	router := NewInternalRouter(logger, metricsReg, mockMetrics)

	// /metrics should return 200 on internal router
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, stdhttp.StatusOK, w.Code, "/metrics should be available on internal router")
	// Note: Empty registry returns empty metrics, which is valid
}

// =============================================================================
// Story 2.6: TRUST_PROXY-Aware RealIP Tests
// =============================================================================

// TestNewRouter_TrustProxyFalse_IgnoresXFF tests that X-Forwarded-For is ignored when TrustProxy=false
func TestNewRouter_TrustProxyFalse_IgnoresXFF(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		// Write the RemoteAddr to response body to verify it wasn't modified by RealIP
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte(r.RemoteAddr))
	})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtConfig := JWTConfig{Enabled: false}
	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 100,
		TrustProxy:        false, // RealIP middleware should NOT be applied
	}

	router := NewRouter(
		logger,
		false,
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		nil,
		1024,
		jwtConfig,
		rateLimitConfig,
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	// Request with X-Forwarded-For header - should be IGNORED
	req := httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, stdhttp.StatusOK, w.Code)
	// RemoteAddr should NOT be modified (RealIP middleware not applied)
	assert.Contains(t, w.Body.String(), "192.168.1.100", "RemoteAddr should be original, not X-Forwarded-For")
	assert.NotContains(t, w.Body.String(), "203.0.113.50", "X-Forwarded-For should be ignored when TrustProxy=false")
}

// TestNewRouter_TrustProxyTrue_UsesXFF tests that X-Forwarded-For is used when TrustProxy=true
func TestNewRouter_TrustProxyTrue_UsesXFF(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		// Write the RemoteAddr to response body to verify it WAS modified by RealIP
		w.WriteHeader(stdhttp.StatusOK)
		w.Write([]byte(r.RemoteAddr))
	})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtConfig := JWTConfig{Enabled: false}
	rateLimitConfig := RateLimitConfig{
		RequestsPerSecond: 100,
		TrustProxy:        true, // RealIP middleware SHOULD be applied
	}

	router := NewRouter(
		logger,
		false,
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		nil,
		1024,
		jwtConfig,
		rateLimitConfig,
		nil, // shutdownCoord - not tested here
		nil, // idempotencyStore - not tested here
		0,   // idempotencyTTL
	)

	// Request with X-Forwarded-For header - should be USED
	req := httptest.NewRequest("GET", "/health", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.50")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, stdhttp.StatusOK, w.Code)
	// RemoteAddr SHOULD be modified by RealIP middleware
	assert.Contains(t, w.Body.String(), "203.0.113.50", "RemoteAddr should be from X-Forwarded-For when TrustProxy=true")
}

// =============================================================================
// Story 2.4: Idempotency Middleware Integration Tests
// =============================================================================

// MockIdempotencyStore is a mock implementation of IdempotencyStore for router tests.
type MockIdempotencyStore struct {
	mock.Mock
}

func (m *MockIdempotencyStore) Get(ctx context.Context, key string) (*middleware.IdempotencyRecord, error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*middleware.IdempotencyRecord), args.Error(1)
}

func (m *MockIdempotencyStore) Store(ctx context.Context, record *middleware.IdempotencyRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0) // Fix: Match method signature (args.Error(0))
}

func TestNewRouter_IdempotencyIntegration(t *testing.T) {
	// Setup dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	metricsReg := prometheus.NewRegistry()
	mockMetrics := new(MockHTTPMetrics)
	mockMetrics.On("IncRequest", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveRequestDuration", mock.Anything, mock.Anything, mock.Anything).Return()
	mockMetrics.On("ObserveResponseSize", mock.Anything, mock.Anything, mock.Anything).Return()

	mockUserHandler := new(MockUserRoutes)
	mockUserHandler.On("CreateUser", mock.Anything, mock.Anything).Return()

	mockIdempotencyStore := new(MockIdempotencyStore)

	healthHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})
	readyHandler := stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {})

	jwtConfig := JWTConfig{Enabled: false}
	rateLimitConfig := RateLimitConfig{RequestsPerSecond: 100}

	router := NewRouter(
		logger,
		false,
		metricsReg,
		mockMetrics,
		healthHandler,
		readyHandler,
		mockUserHandler,
		1024,
		jwtConfig,
		rateLimitConfig,
		nil,
		mockIdempotencyStore, // Provide store to enable middleware
		0,                    // idempotencyTTL
	)

	// Case 1: POST request with Idempotency-Key
	// Should call store.Get and store.Store
	t.Run("POST with Idempotency-Key triggers middleware", func(t *testing.T) {
		key := "550e8400-e29b-41d4-a716-446655440000"

		// Expect Get call - return nil (not found)
		mockIdempotencyStore.On("Get", mock.Anything, key).Return(nil, nil).Once()

		// Expect Store call
		mockIdempotencyStore.On("Store", mock.Anything, mock.MatchedBy(func(record *middleware.IdempotencyRecord) bool {
			return record.Key == key
		})).Return(nil).Once()

		req := httptest.NewRequest("POST", "/api/v1/users", nil)
		req.Header.Set("Idempotency-Key", key) // Use string literal to avoid import cycle for constant
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, stdhttp.StatusCreated, w.Code)
		// "stored" status header indicates middleware ran and stored response
		assert.Equal(t, "stored", w.Header().Get("Idempotency-Status"))

		mockIdempotencyStore.AssertExpectations(t)
	})

	// Case 2: POST request WITHOUT Idempotency-Key
	// Should NOT call store methods
	t.Run("POST without Idempotency-Key bypasses middleware", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/users", nil)
		// No Idempotency-Key header

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, stdhttp.StatusCreated, w.Code)
		assert.Empty(t, w.Header().Get("Idempotency-Status"))

		// Assert Store wasn't called (implied by previous expectations being "Once" and consumed)
	})
}
