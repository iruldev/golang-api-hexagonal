// Package middleware provides HTTP middleware components.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/metrics"
)

// Metrics is a middleware that records HTTP request metrics.
// It captures method, route pattern (from Chi), and response status code.
func Metrics(recorder metrics.HTTPMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			ww := NewResponseWrapper(w)

			// Process request
			next.ServeHTTP(ww, r)

			// Get route pattern from Chi context (e.g., /api/v1/users/{id})
			routePattern := r.URL.Path
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if rp := rctx.RoutePattern(); rp != "" {
					routePattern = rp
				}
			}

			// Record metrics
			recorder.IncRequest(
				r.Method,
				routePattern,
				strconv.Itoa(ww.Status()),
			)

			recorder.ObserveRequestDuration(
				r.Method,
				routePattern,
				time.Since(start).Seconds(),
			)
		})
	}
}
