// Package observability provides logging, tracing, and metrics utilities.
//
// # Logging
//
// Use NewLogger to create a structured JSON logger:
//
//	logger := observability.NewLogger(cfg)
//	logger.Info("user created", "userId", id)
//
// # Tracing
//
// Use InitTracer to initialize OpenTelemetry tracing:
//
//	tp, err := observability.InitTracer(ctx, cfg)
//	defer tp.Shutdown(ctx)
//
// # Metrics
//
// Use NewMetricsRegistry to create base metrics registry:
//
//	registry, httpMetrics := observability.NewMetricsRegistry()
//
// Create custom metrics with factory functions:
//
//	counter := observability.MustNewCounter(registry, "users_total", "Total users", []string{"status"})
//	counter.WithLabelValues("active").Inc()
//
//	histogram := observability.MustNewHistogram(registry, "request_size", "Request size", []string{}, nil)
//	histogram.WithLabelValues().Observe(1024)
//
//	gauge := observability.MustNewGauge(registry, "connections", "Active connections", []string{"pool"})
//	gauge.WithLabelValues("db").Set(5)
package observability

import (
	"errors"
	"fmt"
	"strings"

	"slices"

	dto "github.com/prometheus/client_model/go"

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

// NewCounter creates and registers a counter metric with the given registry.
//
// A counter is a cumulative metric that represents a single monotonically increasing value.
// Counters can only increase (or be reset to zero on restart).
// Use counters to track the number of requests, completed tasks, or errors.
//
// Example:
//
//	counter := observability.NewCounter(registry, "users_created_total",
//	    "Total number of users created", []string{"source"})
//	counter.WithLabelValues("api").Inc()
//	counter.WithLabelValues("api").Add(5)
func NewCounter(registry *prometheus.Registry, name, help string, labels []string) (*prometheus.CounterVec, error) {
	if registry == nil {
		return nil, errors.New("observability.NewCounter: registry is nil")
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	return registerCounter(registry, name, counter)
}

// NewHistogram creates and registers a histogram metric with the given registry.
//
// A histogram samples observations (usually things like request durations or response sizes)
// and counts them in configurable buckets. It also provides a sum of all observed values.
// When buckets is nil, prometheus.DefBuckets is used.
//
// Example:
//
//	histogram := observability.NewHistogram(registry, "request_payload_size_bytes",
//	    "Size of request payloads in bytes", []string{"endpoint"},
//	    []float64{100, 500, 1000, 5000, 10000})
//	histogram.WithLabelValues("/api/users").Observe(1024)
func NewHistogram(registry *prometheus.Registry, name, help string, labels []string, buckets []float64) (*prometheus.HistogramVec, error) {
	if registry == nil {
		return nil, errors.New("observability.NewHistogram: registry is nil")
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		},
		labels,
	)
	return registerHistogram(registry, name, histogram)
}

// NewGauge creates and registers a gauge metric with the given registry.
//
// A gauge is a metric that represents a single numerical value that can arbitrarily go up and down.
// Use gauges to track current values like temperatures, concurrent requests, or queue sizes.
//
// Example:
//
//	gauge := observability.NewGauge(registry, "active_connections",
//	    "Number of active connections", []string{"pool"})
//	gauge.WithLabelValues("postgres").Set(10)
//	gauge.WithLabelValues("postgres").Inc()
//	gauge.WithLabelValues("postgres").Dec()
func NewGauge(registry *prometheus.Registry, name, help string, labels []string) (*prometheus.GaugeVec, error) {
	if registry == nil {
		return nil, errors.New("observability.NewGauge: registry is nil")
	}

	gauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
		labels,
	)
	return registerGauge(registry, name, gauge)
}

// MustNewCounter wraps NewCounter and panics on error. Intended for initialization paths
// where a fatal misconfiguration should stop startup immediately.
func MustNewCounter(registry *prometheus.Registry, name, help string, labels []string) *prometheus.CounterVec {
	counter, err := NewCounter(registry, name, help, labels)
	if err != nil {
		panic(err)
	}
	return counter
}

// MustNewHistogram wraps NewHistogram and panics on error. Intended for initialization paths.
func MustNewHistogram(registry *prometheus.Registry, name, help string, labels []string, buckets []float64) *prometheus.HistogramVec {
	histogram, err := NewHistogram(registry, name, help, labels, buckets)
	if err != nil {
		panic(err)
	}
	return histogram
}

// MustNewGauge wraps NewGauge and panics on error. Intended for initialization paths.
func MustNewGauge(registry *prometheus.Registry, name, help string, labels []string) *prometheus.GaugeVec {
	gauge, err := NewGauge(registry, name, help, labels)
	if err != nil {
		panic(err)
	}
	return gauge
}

func registerCounter(registry *prometheus.Registry, name string, counter *prometheus.CounterVec) (*prometheus.CounterVec, error) {
	if err := registry.Register(counter); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			existing, ok := are.ExistingCollector.(*prometheus.CounterVec)
			if !ok {
				return nil, fmt.Errorf("observability.NewCounter: existing collector for %s has unexpected type", name)
			}
			if !descriptorsMatch(counter, existing) {
				return nil, fmt.Errorf("observability.NewCounter: descriptor mismatch for %s; ensure labels/help align", name)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("observability.NewCounter: register %s: %w", name, err)
	}
	return counter, nil
}

func registerHistogram(registry *prometheus.Registry, name string, histogram *prometheus.HistogramVec) (*prometheus.HistogramVec, error) {
	if err := registry.Register(histogram); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			existing, ok := are.ExistingCollector.(*prometheus.HistogramVec)
			if !ok {
				return nil, fmt.Errorf("observability.NewHistogram: existing collector for %s has unexpected type", name)
			}
			if !descriptorsMatch(histogram, existing) {
				return nil, fmt.Errorf("observability.NewHistogram: descriptor mismatch for %s; ensure labels/help/buckets align", name)
			}
			if !bucketsMatch(histogram, existing) {
				return nil, fmt.Errorf("observability.NewHistogram: bucket mismatch for %s; ensure buckets align", name)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("observability.NewHistogram: register %s: %w", name, err)
	}
	return histogram, nil
}

func registerGauge(registry *prometheus.Registry, name string, gauge *prometheus.GaugeVec) (*prometheus.GaugeVec, error) {
	if err := registry.Register(gauge); err != nil {
		if are, ok := err.(prometheus.AlreadyRegisteredError); ok {
			existing, ok := are.ExistingCollector.(*prometheus.GaugeVec)
			if !ok {
				return nil, fmt.Errorf("observability.NewGauge: existing collector for %s has unexpected type", name)
			}
			if !descriptorsMatch(gauge, existing) {
				return nil, fmt.Errorf("observability.NewGauge: descriptor mismatch for %s; ensure labels/help align", name)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("observability.NewGauge: register %s: %w", name, err)
	}
	return gauge, nil
}

// descriptorsMatch compares the first descriptor string of both collectors to ensure label schema/help align.
func descriptorsMatch(a prometheus.Collector, b prometheus.Collector) bool {
	da := descriptorString(a)
	db := descriptorString(b)
	return da == db
}

func descriptorString(c prometheus.Collector) string {
	ch := make(chan *prometheus.Desc, 1)
	c.Describe(ch)
	close(ch)
	var descs []string
	for d := range ch {
		descs = append(descs, d.String())
	}
	return strings.Join(descs, ";")
}

// bucketsMatch compares the configured buckets by inspecting collected metrics.
func bucketsMatch(a *prometheus.HistogramVec, b *prometheus.HistogramVec) bool {
	return slices.Equal(bucketSignature(a), bucketSignature(b))
}

func bucketSignature(h *prometheus.HistogramVec) []float64 {
	metricCh := make(chan prometheus.Metric, 1)
	h.Collect(metricCh)
	close(metricCh)

	for m := range metricCh {
		dtoM := &dto.Metric{}
		if err := m.Write(dtoM); err != nil {
			return nil
		}
		if hist := dtoM.GetHistogram(); hist != nil {
			bounds := make([]float64, 0, len(hist.Bucket))
			for _, b := range hist.Bucket {
				bounds = append(bounds, b.GetUpperBound())
			}
			return bounds
		}
	}
	return nil
}
