//go:build integration

package handler_test

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/middleware"
)

// TestMetricsAudit_NoSensitiveDataInLabels validates that /metrics output
// does not contain sensitive data such as UUIDs, emails, or secrets in labels.
// This is a security audit test per Story 3.7 (FR46).
//
// REFACTOR NOTE: This test now uses the REAL middleware stack to ensure
// that the safeguards in middleware.Metrics() are actually effective.
func TestMetricsAudit_NoSensitiveDataInLabels(t *testing.T) {
	// 1. Setup real dependencies
	registry, httpMetrics := observability.NewMetricsRegistry()

	// 2. Setup Router with Middleware
	r := chi.NewRouter()
	r.Use(middleware.Metrics(httpMetrics))

	// 3. Define routes that simulate sensitive data exposure scenarios
	r.Get("/api/v1/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/api/v1/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	// Route not registered in chi to test "unmatched" fallback
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	// 4. Fire requests that include sensitive data in the URL
	ts := httptest.NewServer(r)
	defer ts.Close()

	client := ts.Client()

	// Case A: UUID in path -> Should be collapsed to {id}
	client.Get(ts.URL + "/api/v1/users/550e8400-e29b-41d4-a716-446655440000")

	// Case B: Email in path (even though route defines {id}) -> Should be collapsed to {id}
	client.Get(ts.URL + "/api/v1/users/user@example.com")

	// Case C: Unmatched route with PII -> Should be "unmatched"
	client.Get(ts.URL + "/api/v1/users/unknown/extra/path/user@example.com")

	// Case D: Non-standard method -> Should be "OTHER"
	req, _ := http.NewRequest("PROPFIND", ts.URL+"/api/v1/users", nil)
	client.Do(req)

	// 5. Scrape /metrics using the registry we created
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRec := httptest.NewRecorder()
	handler.ServeHTTP(metricsRec, metricsReq)

	require.Equal(t, http.StatusOK, metricsRec.Code)
	metricsOutput := metricsRec.Body.String()
	require.NotEmpty(t, metricsOutput, "metrics output should not be empty")

	// 6. Audit checks using patterns
	t.Run("No UUID patterns in labels", func(t *testing.T) {
		uuidPattern := regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
		matches := uuidPattern.FindAllString(metricsOutput, -1)
		assert.Empty(t, matches, "UUIDs should not appear in metrics labels: %v", matches)
	})

	t.Run("No email patterns in labels", func(t *testing.T) {
		emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
		matches := emailPattern.FindAllString(metricsOutput, -1)
		assert.Empty(t, matches, "Emails should not appear in metrics labels: %v", matches)
	})

	t.Run("No actual user IDs in route labels", func(t *testing.T) {
		// Check for patterns like /users/123 or /users/abc-123
		// We expect: route="/api/v1/users/{id}"
		actualIDPattern := regexp.MustCompile(`route="[^"]*/(users|accounts|orders)/[a-zA-Z0-9-]{3,}"`)
		matches := actualIDPattern.FindAllString(metricsOutput, -1)

		// Filter out legitimate patterns with placeholders
		var badMatches []string
		for _, m := range matches {
			if !strings.Contains(m, "{id}") && !strings.Contains(m, "{userId}") {
				badMatches = append(badMatches, m)
			}
		}
		assert.Empty(t, badMatches, "Actual user IDs should not appear in route labels: %v", badMatches)
	})

	t.Run("Route labels use placeholders", func(t *testing.T) {
		// Verify that parameterized routes use placeholders despite request having real IDs
		assert.Contains(t, metricsOutput, `route="/api/v1/users/{id}"`,
			"Route labels should use {id} placeholder")
	})

	t.Run("Unmatched routes use safe fallback", func(t *testing.T) {
		// The 404 request to /api/v1/users/unknown/extra... should fall here
		assert.Contains(t, metricsOutput, `route="unmatched"`,
			"Unmatched routes should use 'unmatched' label")
	})

	t.Run("Methods are sanitized", func(t *testing.T) {
		// The PROPFIND request
		assert.Contains(t, metricsOutput, `method="OTHER"`,
			"Non-standard methods should be sanitized to 'OTHER'")
	})

	t.Run("No JWT tokens in labels", func(t *testing.T) {
		jwtPattern := regexp.MustCompile(`eyJ[A-Za-z0-9_-]{10,}`)
		matches := jwtPattern.FindAllString(metricsOutput, -1)
		assert.Empty(t, matches, "JWT tokens should not appear in metrics: %v", matches)
	})
}
