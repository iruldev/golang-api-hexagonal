package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/interface/http/httpx"
	"github.com/iruldev/golang-api-hexagonal/internal/observability"
)

// Metrics middleware records HTTP request metrics (count, duration).
// It captures method, path, status for http_requests_total counter
// and method, path for http_request_duration_seconds histogram.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		rw := httpx.NewResponseWriter(w)

		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		method := r.Method
		path := r.URL.Path
		status := strconv.Itoa(rw.StatusCode())

		observability.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		observability.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}
