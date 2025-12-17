package observability

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewCounter(t *testing.T) {
	t.Run("creates and registers counter with labels", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		counter, err := NewCounter(registry, "test_counter_total", "A test counter", []string{"source", "status"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if counter == nil {
			t.Fatal("expected counter to be created, got nil")
		}

		// Increment the counter with label values
		counter.WithLabelValues("api", "success").Inc()
		counter.WithLabelValues("api", "success").Add(5)
		counter.WithLabelValues("cli", "error").Inc()

		// Verify counter appears in registry output
		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range gathered {
			if mf.GetName() == "test_counter_total" {
				found = true
				if len(mf.GetMetric()) != 2 {
					t.Errorf("expected 2 metric variants, got %d", len(mf.GetMetric()))
				}
				break
			}
		}

		if !found {
			t.Error("counter not found in registry")
		}
	})

	t.Run("counter increments correctly", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		counter, err := NewCounter(registry, "users_created_total", "Total users created", []string{"source"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		counter.WithLabelValues("api").Inc()
		counter.WithLabelValues("api").Inc()
		counter.WithLabelValues("api").Inc()

		expected := `
# HELP users_created_total Total users created
# TYPE users_created_total counter
users_created_total{source="api"} 3
`
		if err := testutil.GatherAndCompare(registry, strings.NewReader(expected), "users_created_total"); err != nil {
			t.Errorf("unexpected metric value: %v", err)
		}
	})

	t.Run("counter with Add", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		counter, err := NewCounter(registry, "requests_total", "Total requests", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		counter.WithLabelValues().Add(100)

		expected := `
# HELP requests_total Total requests
# TYPE requests_total counter
requests_total 100
`
		if err := testutil.GatherAndCompare(registry, strings.NewReader(expected), "requests_total"); err != nil {
			t.Errorf("unexpected metric value: %v", err)
		}
	})

	t.Run("returns existing counter when registered twice", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		first, err := NewCounter(registry, "duplicate_counter", "Duplicate counter", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		second, err := NewCounter(registry, "duplicate_counter", "Duplicate counter", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if first != second {
			t.Fatalf("expected existing counter to be returned on duplicate registration")
		}
	})

	t.Run("returns error when descriptor mismatches", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		_, err := NewCounter(registry, "mismatch_counter", "first help", []string{"a"})
		if err != nil {
			t.Fatalf("unexpected error on first register: %v", err)
		}

		_, err = NewCounter(registry, "mismatch_counter", "other help", []string{"b"})
		if err == nil {
			t.Fatalf("expected error on descriptor mismatch, got nil")
		}
	})
}

func TestNewHistogram(t *testing.T) {
	t.Run("creates and registers histogram with labels", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		histogram, err := NewHistogram(registry, "request_size_bytes", "Request size in bytes", []string{"endpoint"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if histogram == nil {
			t.Fatal("expected histogram to be created, got nil")
		}

		// Observe some values
		histogram.WithLabelValues("/api/users").Observe(100)
		histogram.WithLabelValues("/api/users").Observe(500)
		histogram.WithLabelValues("/api/posts").Observe(1000)

		// Verify histogram appears in registry output
		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range gathered {
			if mf.GetName() == "request_size_bytes" {
				found = true
				break
			}
		}

		if !found {
			t.Error("histogram not found in registry")
		}
	})

	t.Run("histogram with custom buckets", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		buckets := []float64{100, 500, 1000, 5000, 10000}
		histogram, err := NewHistogram(registry, "payload_bytes", "Payload size", []string{}, buckets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		histogram.WithLabelValues().Observe(250)  // falls in 500 bucket
		histogram.WithLabelValues().Observe(750)  // falls in 1000 bucket
		histogram.WithLabelValues().Observe(2000) // falls in 5000 bucket

		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range gathered {
			if mf.GetName() == "payload_bytes" {
				found = true
				if len(mf.GetMetric()) > 0 {
					// Verify it exists and has histogram data
					metric := mf.GetMetric()[0]
					h := metric.GetHistogram()
					if h.GetSampleCount() != 3 {
						t.Errorf("expected 3 samples, got %d", h.GetSampleCount())
					}
				}
				break
			}
		}

		if !found {
			t.Error("histogram not found in registry")
		}
	})

	t.Run("histogram uses default buckets when nil", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		histogram, err := NewHistogram(registry, "latency_seconds", "Latency", []string{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		histogram.WithLabelValues().Observe(0.1)

		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		for _, mf := range gathered {
			if mf.GetName() == "latency_seconds" {
				if len(mf.GetMetric()) > 0 {
					h := mf.GetMetric()[0].GetHistogram()
					// DefBuckets has 11 buckets
					if len(h.GetBucket()) != 11 {
						t.Errorf("expected 11 default buckets, got %d", len(h.GetBucket()))
					}
				}
				return
			}
		}

		t.Error("histogram not found in registry")
	})

	t.Run("returns error when descriptor mismatches", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		_, err := NewHistogram(registry, "mismatch_hist", "help1", []string{"a"}, nil)
		if err != nil {
			t.Fatalf("unexpected error on first register: %v", err)
		}

		_, err = NewHistogram(registry, "mismatch_hist", "help2", []string{"b"}, nil)
		if err == nil {
			t.Fatalf("expected error on descriptor mismatch, got nil")
		}
	})
}

func TestNewGauge(t *testing.T) {
	t.Run("creates and registers gauge with labels", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		gauge, err := NewGauge(registry, "active_connections", "Active connections", []string{"pool"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if gauge == nil {
			t.Fatal("expected gauge to be created, got nil")
		}

		// Set values
		gauge.WithLabelValues("postgres").Set(10)
		gauge.WithLabelValues("redis").Set(5)

		// Verify gauge appears in registry output
		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		found := false
		for _, mf := range gathered {
			if mf.GetName() == "active_connections" {
				found = true
				if len(mf.GetMetric()) != 2 {
					t.Errorf("expected 2 metric variants, got %d", len(mf.GetMetric()))
				}
				break
			}
		}

		if !found {
			t.Error("gauge not found in registry")
		}
	})

	t.Run("gauge set, inc, dec operations", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		gauge, err := NewGauge(registry, "queue_size", "Queue size", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gauge.WithLabelValues().Set(10)
		gauge.WithLabelValues().Inc()
		gauge.WithLabelValues().Inc()
		gauge.WithLabelValues().Dec()

		expected := `
# HELP queue_size Queue size
# TYPE queue_size gauge
queue_size 11
`
		if err := testutil.GatherAndCompare(registry, strings.NewReader(expected), "queue_size"); err != nil {
			t.Errorf("unexpected metric value: %v", err)
		}
	})

	t.Run("gauge add and sub operations", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		gauge, err := NewGauge(registry, "temperature", "Temperature", []string{"location"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gauge.WithLabelValues("room1").Set(20)
		gauge.WithLabelValues("room1").Add(5)
		gauge.WithLabelValues("room1").Sub(3)

		expected := `
# HELP temperature Temperature
# TYPE temperature gauge
temperature{location="room1"} 22
`
		if err := testutil.GatherAndCompare(registry, strings.NewReader(expected), "temperature"); err != nil {
			t.Errorf("unexpected metric value: %v", err)
		}
	})

	t.Run("returns existing gauge when registered twice", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		first, err := NewGauge(registry, "duplicate_gauge", "Duplicate gauge", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		second, err := NewGauge(registry, "duplicate_gauge", "Duplicate gauge", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if first != second {
			t.Fatalf("expected existing gauge to be returned on duplicate registration")
		}
	})

	t.Run("returns error when descriptor mismatches", func(t *testing.T) {
		registry := prometheus.NewRegistry()
		_, err := NewGauge(registry, "mismatch_gauge", "help1", []string{"a"})
		if err != nil {
			t.Fatalf("unexpected error on first register: %v", err)
		}

		_, err = NewGauge(registry, "mismatch_gauge", "help2", []string{"b"})
		if err == nil {
			t.Fatalf("expected error on descriptor mismatch, got nil")
		}
	})
}

func TestMetricsAppearInRegistry(t *testing.T) {
	t.Run("all custom metrics appear in single registry", func(t *testing.T) {
		registry := prometheus.NewRegistry()

		// Create different types of custom metrics
		counter, err := NewCounter(registry, "events_total", "Total events", []string{"type"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		histogram, err := NewHistogram(registry, "duration_seconds", "Duration", []string{}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		gauge, err := NewGauge(registry, "pending_items", "Pending items", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Use the metrics
		counter.WithLabelValues("click").Inc()
		histogram.WithLabelValues().Observe(0.5)
		gauge.WithLabelValues().Set(42)

		// Verify all metrics appear in registry
		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		metricNames := make(map[string]bool)
		for _, mf := range gathered {
			metricNames[mf.GetName()] = true
		}

		expectedMetrics := []string{"events_total", "duration_seconds", "pending_items"}
		for _, name := range expectedMetrics {
			if !metricNames[name] {
				t.Errorf("metric %q not found in registry", name)
			}
		}
	})

	t.Run("custom metrics work alongside NewMetricsRegistry", func(t *testing.T) {
		// Get the standard registry with HTTP metrics
		registry, httpMetrics := NewMetricsRegistry()

		// Add custom metrics to the same registry
		customCounter, err := NewCounter(registry, "custom_events_total", "Custom events", []string{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		customCounter.WithLabelValues().Inc()

		// Use HTTP metrics too
		httpMetrics.IncRequest("GET", "/api/test", "200")

		// Verify both appear
		gathered, err := registry.Gather()
		if err != nil {
			t.Fatalf("failed to gather metrics: %v", err)
		}

		metricNames := make(map[string]bool)
		for _, mf := range gathered {
			metricNames[mf.GetName()] = true
		}

		// Check HTTP metrics
		if !metricNames["http_requests_total"] {
			t.Error("http_requests_total not found in registry")
		}

		// Check custom metrics
		if !metricNames["custom_events_total"] {
			t.Error("custom_events_total not found in registry")
		}
	})
}
