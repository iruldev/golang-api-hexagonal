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
			routePattern := "unmatched"
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if rp := rctx.RoutePattern(); rp != "" {
					routePattern = rp
				}
			}

			// Sanitize method to prevent cardinality explosion (e.g. from malicious arbitrary methods)
			method := r.Method
			switch method {
			case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace:
				// Allowed standard methods
			default:
				method = "OTHER"
			}

			// Record metrics
			recorder.IncRequest(
				method,
				routePattern,
				strconv.Itoa(ww.Status()),
			)

			recorder.ObserveRequestDuration(
				method,
				routePattern,
				time.Since(start).Seconds(),
			)

			recorder.ObserveResponseSize(
				method,
				routePattern,
				float64(ww.BytesWritten()),
			)
		})
	}
}
