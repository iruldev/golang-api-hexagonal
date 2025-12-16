// Package observability provides observability components including logging, tracing, and metrics.
package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
)

// httpMetrics implements the shared HTTPMetrics interface using Prometheus collectors.
type httpMetrics struct {
	requests  *prometheus.CounterVec
	durations *prometheus.HistogramVec
}

func (m *httpMetrics) IncRequest(method, route, status string) {
	m.requests.WithLabelValues(method, route, status).Inc()
}

func (m *httpMetrics) ObserveRequestDuration(method, route string, seconds float64) {
	m.durations.WithLabelValues(method, route).Observe(seconds)
}

// Reset is used in tests to clear collectors.
func (m *httpMetrics) Reset() {
	m.requests.Reset()
	m.durations.Reset()
}

// NewMetricsRegistry creates a new Prometheus registry with Go runtime collectors
// and HTTP metrics already registered. It returns the registry along with a recorder
// that satisfies the shared HTTPMetrics interface for use by transport middleware.
func NewMetricsRegistry() (*prometheus.Registry, metrics.HTTPMetrics) {
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

	// Go runtime metrics (go_goroutines, go_memstats_*, etc.)
	reg.MustRegister(collectors.NewGoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	// HTTP metrics
	reg.MustRegister(requests)
	reg.MustRegister(durations)

	return reg, &httpMetrics{
		requests:  requests,
		durations: durations,
	}
}
