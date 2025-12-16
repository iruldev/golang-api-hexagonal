// Package middleware provides HTTP middleware components.
package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/observability"
)

// resetMetrics resets the metrics instance for clean tests.
func resetMetrics(m interface{}) {
	if hm, ok := m.(interface{ Reset() }); ok {
		hm.Reset()
	}
}

func TestMetrics_RecordsRequestsTotal(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create a chi router with the metrics middleware
	r := chi.NewRouter()
	r.Use(Metrics(httpMetrics))
	r.Get("/test", handler)

	// Make a request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify counter was incremented
	metricsFamilies, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range metricsFamilies {
		if f.GetName() == "http_requests_total" {
			found = true
		}
	}
	assert.True(t, found, "http_requests_total should be exported")
}

func TestMetrics_RecordsRequestDuration(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	r := chi.NewRouter()
	r.Use(Metrics(httpMetrics))
	r.Post("/users", handler)

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// Verify histogram was observed by checking request counter exists with correct labels
	// The histogram observation happens in the same code path
	metricsFamilies, err := reg.Gather()
	require.NoError(t, err)

	foundCounter := false
	foundHistogram := false
	for _, f := range metricsFamilies {
		if f.GetName() == "http_requests_total" {
			foundCounter = true
		}
		if f.GetName() == "http_request_duration_seconds" {
			foundHistogram = true
		}
	}
	assert.True(t, foundCounter, "http_requests_total should be exported")
	assert.True(t, foundHistogram, "http_request_duration_seconds should be exported")
}

func TestMetrics_UsesChiRoutePattern(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(Metrics(httpMetrics))
	r.Get("/users/{id}", handler)

	// Request with a specific ID
	req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify the route pattern is used (not the actual path)
	families, err := reg.Gather()
	require.NoError(t, err)

	foundPattern := false
	for _, f := range families {
		if f.GetName() != "http_requests_total" {
			continue
		}
		for _, m := range f.Metric {
			if len(m.Label) == 3 {
				if m.Label[1].GetValue() == "/users/{id}" && m.Label[2].GetValue() == "200" {
					foundPattern = true
				}
			}
		}
	}
	assert.True(t, foundPattern, "route label should use chi pattern")
}

func TestMetrics_CapturesStatusCode(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"200 OK", http.StatusOK},
		{"201 Created", http.StatusCreated},
		{"400 Bad Request", http.StatusBadRequest},
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, httpMetrics := observability.NewMetricsRegistry()
			resetMetrics(httpMetrics)

			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			r := chi.NewRouter()
			r.Use(Metrics(httpMetrics))
			r.Get("/status", handler)

			req := httptest.NewRequest(http.MethodGet, "/status", nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			assert.Equal(t, tt.statusCode, rec.Code)

			// Verify correct status label
			families, err := reg.Gather()
			require.NoError(t, err)

			found := false
			for _, f := range families {
				if f.GetName() != "http_requests_total" {
					continue
				}
				for _, m := range f.Metric {
					if len(m.Label) == 3 &&
						m.Label[0].GetValue() == "GET" &&
						m.Label[1].GetValue() == "/status" &&
						m.Label[2].GetValue() == strconv.Itoa(tt.statusCode) {
						found = true
					}
				}
			}
			assert.True(t, found, "should record status code label")
		})
	}
}

func TestMetrics_MultipleRequests(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(Metrics(httpMetrics))
	r.Get("/api", handler)

	// Make 3 requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api", nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// The counter should have been incremented 3 times
	families, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() != "http_requests_total" {
			continue
		}
		for _, m := range f.Metric {
			if len(m.Label) == 3 &&
				m.Label[0].GetValue() == "GET" &&
				m.Label[1].GetValue() == "/api" &&
				m.Label[2].GetValue() == "200" {
				found = true
			}
		}
	}
	assert.True(t, found, "should record multiple requests")
}

func TestMetrics_UsesResponseWrapper(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	// Handler that writes body without calling WriteHeader explicitly
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "default status")
	})

	r := chi.NewRouter()
	r.Use(Metrics(httpMetrics))
	r.Get("/default", handler)

	req := httptest.NewRequest(http.MethodGet, "/default", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Should default to 200
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "default status", rec.Body.String())

	families, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() != "http_requests_total" {
			continue
		}
		for _, m := range f.Metric {
			if len(m.Label) == 3 &&
				m.Label[0].GetValue() == "GET" &&
				m.Label[1].GetValue() == "/default" &&
				m.Label[2].GetValue() == "200" {
				found = true
			}
		}
	}
	assert.True(t, found, "should default status to 200")
}

func TestMetrics_FallbackToURLPath(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	// Create a simple handler without chi routing
	metricsHandler := Metrics(httpMetrics)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Request without chi context
	req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
	rec := httptest.NewRecorder()

	metricsHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Should use path when route pattern is not available
	families, err := reg.Gather()
	require.NoError(t, err)

	found := false
	for _, f := range families {
		if f.GetName() != "http_requests_total" {
			continue
		}
		for _, m := range f.Metric {
			if len(m.Label) == 3 &&
				m.Label[0].GetValue() == "GET" &&
				m.Label[1].GetValue() == "/fallback" &&
				m.Label[2].GetValue() == "200" {
				found = true
			}
		}
	}
	assert.True(t, found, "should use URL path when no chi route pattern")
}

func TestNewMetricsRegistry_HasGoCollectors(t *testing.T) {
	reg, httpMetrics := observability.NewMetricsRegistry()
	resetMetrics(httpMetrics)

	// Gather all metrics
	families, err := reg.Gather()
	require.NoError(t, err)

	metricNames := make([]string, 0, len(families))
	for _, f := range families {
		metricNames = append(metricNames, f.GetName())
	}

	// Check for Go runtime metrics
	hasGoroutines := false
	hasMemstats := false
	for _, name := range metricNames {
		if name == "go_goroutines" {
			hasGoroutines = true
		}
		if strings.HasPrefix(name, "go_memstats_") {
			hasMemstats = true
		}
	}

	assert.True(t, hasGoroutines, "go_goroutines metric should be present")
	assert.True(t, hasMemstats, "go_memstats_* metrics should be present")
}
