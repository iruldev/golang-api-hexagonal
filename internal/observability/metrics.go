// Package observability provides observability utilities for the application.
package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP metrics for monitoring API performance.
var (
	// HTTPRequestsTotal counts total HTTP requests by method, path, and status.
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration measures HTTP request duration in seconds.
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// Job metrics for monitoring worker performance.
var (
	// JobProcessedTotal counts total job executions by task_type, queue, status.
	JobProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "job_processed_total",
			Help: "Total jobs processed",
		},
		[]string{"task_type", "queue", "status"},
	)

	// JobDurationSeconds measures job execution duration in seconds.
	JobDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_duration_seconds",
			Help:    "Job duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"task_type", "queue"},
	)
)
