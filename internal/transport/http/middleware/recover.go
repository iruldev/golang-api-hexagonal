// Package middleware provides HTTP middleware components.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// panicCounter tracks the number of panics recovered by the Recoverer middleware.
// Labels: method, path (route pattern is not available in panic context, so use path).
var panicCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "http",
		Subsystem: "server",
		Name:      "panics_total",
		Help:      "Total number of panics recovered by the Recoverer middleware",
	},
	[]string{"method", "path"},
)

// Recoverer returns a middleware that recovers from panics, logs them,
// and returns an RFC 7807 Problem Details response with error code SYS-001.
//
// The middleware:
//   - Catches any panic that occurs in downstream handlers
//   - Logs the panic with full stack trace for debugging (not exposed to client)
//   - Increments a Prometheus counter for monitoring
//   - Returns a safe RFC 7807 response with Code="SYS-001" and status 500
//   - Includes request_id and trace_id in the error response for correlation
//
// Example usage:
//
//	r := chi.NewRouter()
//	r.Use(middleware.RequestID)
//	r.Use(middleware.Recoverer(logger))
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					// Capture context values for logging
					requestID := ctxutil.GetRequestID(r.Context())
					traceID := ctxutil.GetTraceID(r.Context())

					// Log the panic with full stack trace for debugging
					// Stack trace is logged but NOT exposed to client
					logArgs := []any{
						"panic", rec,
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path,
						"request_id", requestID,
					}

					// Add trace_id if present (tracing may be disabled)
					if traceID != "" && traceID != ctxutil.EmptyTraceID {
						logArgs = append(logArgs, "trace_id", traceID)
					}

					logger.ErrorContext(r.Context(), "panic recovered", logArgs...)

					// Increment Prometheus counter for monitoring
					panicCounter.WithLabelValues(r.Method, r.URL.Path).Inc()

					// Return RFC 7807 Problem Details with SYS-001
					// Use safe generic message - DO NOT expose panic details to client
					info := contract.GetErrorCodeInfo(contract.CodeSysInternal)
					problem := contract.NewProblemWithType(
						contract.ProblemTypeURL(info.ProblemTypeSlug),
						info.HTTPStatus,
						info.Title,
						info.DetailTemplate, // "An internal error occurred" - safe for client
					)
					problem.Code = info.Code
					problem.RequestID = requestID
					if traceID != "" && traceID != ctxutil.EmptyTraceID {
						problem.TraceID = traceID
					}
					problem.Instance = r.URL.Path

					contract.WriteProblem(w, problem)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
