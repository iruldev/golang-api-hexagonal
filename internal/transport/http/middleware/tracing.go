package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/iruldev/golang-api-hexagonal/transport/http"

// traceIDKey is the context key for storing trace ID.
const traceIDKey contextKey = "traceId"

// Tracing returns a middleware that creates spans for HTTP requests.
// It extracts W3C Trace Context from incoming headers (traceparent) for distributed tracing,
// creates a span with HTTP attributes, and propagates the trace ID via context.
func Tracing(next http.Handler) http.Handler {
	tracer := otel.Tracer(tracerName)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract trace context from incoming headers (W3C Trace Context)
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Start span with HTTP attributes
		routePattern := getRoutePattern(r)
		if routePattern == "" {
			routePattern = r.URL.Path
		}

		ctx, span := tracer.Start(ctx, routePattern,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
			),
		)
		defer span.End()

		// Store trace ID in context for logging correlation
		traceID := span.SpanContext().TraceID().String()
		ctx = context.WithValue(ctx, traceIDKey, traceID)

		// Use response wrapper to capture status code
		ww := NewResponseWrapper(w)

		// Process the request with the new context (keeps Chi route context)
		reqWithCtx := r.WithContext(ctx)
		next.ServeHTTP(ww, reqWithCtx)

		// Resolve final route pattern after Chi has matched the route
		finalRoutePattern := getRoutePattern(reqWithCtx)
		if finalRoutePattern == "" {
			finalRoutePattern = reqWithCtx.URL.Path
		}

		span.SetName(finalRoutePattern)

		// Add status code after request completes
		span.SetAttributes(
			attribute.String("http.route", finalRoutePattern),
			attribute.Int("http.status_code", ww.Status()),
		)
	})
}

// GetTraceID retrieves the trace ID from the context.
// Returns an empty string if tracing is disabled or no trace ID is present.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok && id != "00000000000000000000000000000000" {
		return id
	}
	return ""
}

// getRoutePattern returns the Chi route pattern or falls back to the URL path.
func getRoutePattern(r *http.Request) string {
	routeCtx := chi.RouteContext(r.Context())
	if routeCtx != nil {
		if pattern := routeCtx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}
