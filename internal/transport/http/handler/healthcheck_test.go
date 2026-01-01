package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/heptiolabs/healthcheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Story 3.4: Health Check Library Integration - Unit Tests
// =============================================================================

// TestNewHealthCheckRegistrySimple verifies simple registry creation.
func TestNewHealthCheckRegistrySimple(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	assert.NotNil(t, registry)
	assert.NotNil(t, registry.handler)
}

// TestHealthCheckRegistry_AddLivenessCheck verifies check registration.
func TestHealthCheckRegistry_AddLivenessCheck(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	// Add a passing liveness check
	registry.AddLivenessCheck("test-check", func() error {
		return nil
	})

	// Verify the handler works after adding check
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()

	registry.LiveHandler()(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHealthCheckRegistry_AddReadinessCheck verifies readiness check registration.
func TestHealthCheckRegistry_AddReadinessCheck(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	// Add a passing readiness check
	registry.AddReadinessCheck("database", func() error {
		return nil
	})

	// Verify the handler works after adding check
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	registry.ReadyHandler()(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestHealthCheckRegistry_LiveHandler_AllPass verifies 200 OK when all checks pass.
func TestHealthCheckRegistry_LiveHandler_AllPass(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	registry.AddLivenessCheck("check1", func() error { return nil })
	registry.AddLivenessCheck("check2", func() error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()

	registry.LiveHandler()(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
	// Library returns {} for success
	assert.Equal(t, "{}\n", rec.Body.String())
}

// TestHealthCheckRegistry_LiveHandler_OneFails verifies 503 when one check fails.
func TestHealthCheckRegistry_LiveHandler_OneFails(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	registry.AddLivenessCheck("passing", func() error { return nil })
	registry.AddLivenessCheck("failing", func() error { return errors.New("check failed") })

	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	rec := httptest.NewRecorder()

	registry.LiveHandler()(rec, req)

	// Should return 503 when any check fails (AC: #2)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// TestHealthCheckRegistry_ReadyHandler_AllPass verifies 200 OK for readiness.
func TestHealthCheckRegistry_ReadyHandler_AllPass(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	registry.AddReadinessCheck("database", func() error { return nil })
	registry.AddReadinessCheck("cache", func() error { return nil })

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	registry.ReadyHandler()(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	// Library returns {} for success
	assert.Equal(t, "{}\n", rec.Body.String())
}

// TestHealthCheckRegistry_ReadyHandler_OneFails verifies 503 when readiness fails.
func TestHealthCheckRegistry_ReadyHandler_OneFails(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	registry.AddReadinessCheck("database", func() error { return nil })
	registry.AddReadinessCheck("external-api", func() error { return errors.New("connection refused") })

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	registry.ReadyHandler()(rec, req)

	// Should return 503 when any check fails (AC: #2)
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// TestHealthCheckRegistry_Handler verifies underlying handler access.
func TestHealthCheckRegistry_Handler(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	handler := registry.Handler()
	assert.NotNil(t, handler)
	// Handler is verified by interface satisfaction at compile time
}

// TestHealthCheckRegistry_LivenessIncludedInReadiness verifies liveness checks are in readiness.
func TestHealthCheckRegistry_LivenessIncludedInReadiness(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	// Add a failing liveness check
	registry.AddLivenessCheck("critical", func() error { return errors.New("liveness failed") })

	// Readiness should also fail because liveness checks are included
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	registry.ReadyHandler()(rec, req)

	// Per library design, every liveness check is also included as readiness
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

// TestHealthCheckRegistry_EmptyChecks verifies behavior with no checks registered.
func TestHealthCheckRegistry_EmptyChecks(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	t.Run("liveness empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/live", nil)
		rec := httptest.NewRecorder()

		registry.LiveHandler()(rec, req)

		// With no checks, should return 200 OK
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("readiness empty", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/ready", nil)
		rec := httptest.NewRecorder()

		registry.ReadyHandler()(rec, req)

		// With no checks, should return 200 OK
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

// TestHealthCheckRegistry_TimeoutCheck verifies timeout wrapper integration.
func TestHealthCheckRegistry_TimeoutCheck(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	// Add a check that takes too long, wrapped with timeout
	slowCheck := func() error {
		time.Sleep(500 * time.Millisecond)
		return nil
	}

	// Wrap with 50ms timeout
	registry.AddReadinessCheck("slow", healthcheck.Timeout(slowCheck, 50*time.Millisecond))

	start := time.Now()

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	registry.ReadyHandler()(rec, req)

	elapsed := time.Since(start)

	// Should fail due to timeout
	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	// Should complete quickly due to timeout
	require.Less(t, elapsed, 200*time.Millisecond, "Timeout should trigger within 50ms")
}

// TestHealthCheckRegistry_MultipleChecks verifies multiple checks work together.
func TestHealthCheckRegistry_MultipleChecks(t *testing.T) {
	registry := NewHealthCheckRegistrySimple()

	// Add multiple readiness checks
	checkCount := 0
	for i := 0; i < 5; i++ {
		registry.AddReadinessCheck("check-"+string(rune('a'+i)), func() error {
			checkCount++
			return nil
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()

	registry.ReadyHandler()(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 5, checkCount, "All 5 checks should have been executed")
}
