package http

import (
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
	)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, stdhttp.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}
