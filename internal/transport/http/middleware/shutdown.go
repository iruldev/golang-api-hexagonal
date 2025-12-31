// Package middleware provides HTTP middleware for the transport layer.
package middleware

import (
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

// ShutdownCoordinator is the interface required for graceful shutdown middleware.
// This interface is defined in the transport layer to avoid importing the infra layer.
// The implementation in internal/infra/resilience.ShutdownCoordinator satisfies this interface.
type ShutdownCoordinator interface {
	// IncrementActive increments the active request counter.
	// Returns false if shutdown has been initiated (caller should reject the request).
	IncrementActive() bool

	// DecrementActive decrements the active request counter.
	DecrementActive()
}

const (
	// ShutdownRetryAfterSeconds is the default Retry-After header value in seconds.
	// This is a reasonable default that works for most drain period configurations.
	// Note: This value is intentionally hardcoded rather than derived from DrainPeriod
	// to keep the middleware interface simple (DIP pattern - minimal interface).
	// For custom Retry-After values, consider extending the ShutdownCoordinator interface.
	ShutdownRetryAfterSeconds = "30"
)

// Shutdown returns a middleware that rejects new requests during graceful shutdown.
// It tracks in-flight requests and returns 503 Service Unavailable during shutdown.
//
// This middleware should be placed early in the chain (after RequestID, Logger, Recoverer)
// to reject requests before consuming rate limit quota and to track all requests.
func Shutdown(coord ShutdownCoordinator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try to increment active count (will fail if shutting down)
			if !coord.IncrementActive() {
				// Shutting down - reject the request with 503
				w.Header().Set("Retry-After", ShutdownRetryAfterSeconds)
				w.Header().Set("Connection", "close")

				problem := contract.NewProblem(http.StatusServiceUnavailable, "Service Unavailable", "Server is shutting down. Please retry later.")
				problem.Type = contract.ProblemTypeURL(contract.ProblemTypeServiceUnavailableSlug)
				problem.Code = contract.CodeServiceUnavailable
				problem.Instance = r.URL.Path

				// Populate request_id and trace_id from context
				problem.RequestID = ctxutil.GetRequestID(r.Context())
				if traceID := ctxutil.GetTraceID(r.Context()); traceID != "" && traceID != ctxutil.EmptyTraceID {
					problem.TraceID = traceID
				}

				contract.WriteProblem(w, problem)
				return
			}

			// Decrement active count when request completes
			defer coord.DecrementActive()

			next.ServeHTTP(w, r)
		})
	}
}
