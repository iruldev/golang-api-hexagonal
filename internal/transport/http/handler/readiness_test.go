package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Story 3.2: Readiness Probe Unit Tests (AC: #1, #2, #3)
// =============================================================================

// mockDependencyChecker is a test mock for DependencyChecker interface.
type mockDependencyChecker struct {
	name    string
	status  string
	latency time.Duration
	err     error
}

func (m *mockDependencyChecker) Name() string {
	return m.name
}

func (m *mockDependencyChecker) CheckHealth(ctx context.Context) (string, time.Duration, error) {
	// Simulate the latency
	if m.latency > 0 {
		select {
		case <-time.After(m.latency):
		case <-ctx.Done():
			return "unhealthy", m.latency, ctx.Err()
		}
	}
	return m.status, m.latency, m.err
}

// TestReadinessHandler_AllHealthy verifies 200 OK when all dependencies are healthy (AC1).
func TestReadinessHandler_AllHealthy(t *testing.T) {
	dbChecker := &mockDependencyChecker{
		name:    "database",
		status:  "healthy",
		latency: 5 * time.Millisecond,
		err:     nil,
	}

	handler := NewReadinessHandler(2*time.Second, dbChecker)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// AC1: 200 OK when all healthy
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", resp.Status)
	assert.Len(t, resp.Checks, 1)
	assert.Equal(t, "healthy", resp.Checks["database"].Status)
	assert.GreaterOrEqual(t, resp.Checks["database"].LatencyMs, int64(0))
	assert.Empty(t, resp.Checks["database"].Error)
}

// TestReadinessHandler_DatabaseUnhealthy verifies 503 when database is unavailable (AC2).
func TestReadinessHandler_DatabaseUnhealthy(t *testing.T) {
	dbChecker := &mockDependencyChecker{
		name:    "database",
		status:  "unhealthy",
		latency: 10 * time.Millisecond,
		err:     errors.New("connection refused"),
	}

	handler := NewReadinessHandler(2*time.Second, dbChecker)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// AC2: 503 Service Unavailable when database unhealthy
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "unhealthy", resp.Status)
	assert.Equal(t, "unhealthy", resp.Checks["database"].Status)
	assert.Equal(t, "connection refused", resp.Checks["database"].Error)
}

// TestReadinessHandler_LatencyIncluded verifies latency is measured and included (AC3).
func TestReadinessHandler_LatencyIncluded(t *testing.T) {
	expectedLatency := 50 * time.Millisecond
	dbChecker := &mockDependencyChecker{
		name:    "database",
		status:  "healthy",
		latency: expectedLatency,
		err:     nil,
	}

	handler := NewReadinessHandler(2*time.Second, dbChecker)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	// AC3: Latency is included in response
	assert.GreaterOrEqual(t, resp.Checks["database"].LatencyMs, expectedLatency.Milliseconds())
}

// TestReadinessHandler_Timeout verifies timeout handling (AC3).
func TestReadinessHandler_Timeout(t *testing.T) {
	// Create a checker that takes longer than the timeout
	slowChecker := &mockDependencyChecker{
		name:    "database",
		status:  "healthy",
		latency: 500 * time.Millisecond, // Longer than timeout
		err:     nil,
	}

	// Use a short timeout
	handler := NewReadinessHandler(50*time.Millisecond, slowChecker)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// AC3: When timeout exceeded, should return unhealthy
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "unhealthy", resp.Status)
	assert.Contains(t, resp.Checks["database"].Error, "context deadline exceeded")
}

// TestReadinessHandler_MultipleDependencies verifies multiple dependency checks.
func TestReadinessHandler_MultipleDependencies(t *testing.T) {
	dbChecker := &mockDependencyChecker{
		name:    "database",
		status:  "healthy",
		latency: 5 * time.Millisecond,
		err:     nil,
	}
	cacheChecker := &mockDependencyChecker{
		name:    "cache",
		status:  "healthy",
		latency: 2 * time.Millisecond,
		err:     nil,
	}

	handler := NewReadinessHandler(2*time.Second, dbChecker, cacheChecker)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", resp.Status)
	assert.Len(t, resp.Checks, 2)
	assert.Equal(t, "healthy", resp.Checks["database"].Status)
	assert.Equal(t, "healthy", resp.Checks["cache"].Status)
}

// TestReadinessHandler_OneUnhealthyFailsAll verifies any unhealthy dependency fails overall.
func TestReadinessHandler_OneUnhealthyFailsAll(t *testing.T) {
	dbChecker := &mockDependencyChecker{
		name:    "database",
		status:  "healthy",
		latency: 5 * time.Millisecond,
		err:     nil,
	}
	cacheChecker := &mockDependencyChecker{
		name:    "cache",
		status:  "unhealthy",
		latency: 2 * time.Millisecond,
		err:     errors.New("cache connection failed"),
	}

	handler := NewReadinessHandler(2*time.Second, dbChecker, cacheChecker)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// One unhealthy should make overall unhealthy
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "unhealthy", resp.Status)
	assert.Equal(t, "healthy", resp.Checks["database"].Status)
	assert.Equal(t, "unhealthy", resp.Checks["cache"].Status)
}

// TestReadinessHandler_NoDependencies verifies behavior with no checkers.
func TestReadinessHandler_NoDependencies(t *testing.T) {
	handler := NewReadinessHandler(2 * time.Second)

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// No dependencies means healthy (nothing to fail)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp ReadinessResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)

	assert.Equal(t, "healthy", resp.Status)
	assert.Empty(t, resp.Checks)
}

// TestReadinessHandler_DefaultTimeout verifies default timeout is applied.
func TestReadinessHandler_DefaultTimeout(t *testing.T) {
	handler := NewReadinessHandler(0) // 0 should use default

	// Verify handler was created with default timeout
	assert.Equal(t, DefaultCheckTimeout, handler.timeout)
}

// BenchmarkReadinessHandler_Healthy validates performance with healthy dependencies.
func BenchmarkReadinessHandler_Healthy(b *testing.B) {
	dbChecker := &mockDependencyChecker{
		name:    "database",
		status:  "healthy",
		latency: 0, // No artificial delay
		err:     nil,
	}

	handler := NewReadinessHandler(2*time.Second, dbChecker)
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			b.Fatalf("Expected 200 OK, got %d", rec.Code)
		}
	}
}
