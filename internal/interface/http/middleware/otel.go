package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Otel middleware wraps the handler with OpenTelemetry tracing.
// It creates a span for each request with the given operation name.
// The span context is propagated to downstream handlers.
func Otel(operation string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, operation)
	}
}
