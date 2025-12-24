package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/iruldev/golang-api-hexagonal/transport/http"

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

		// Store trace ID and span ID in context for logging correlation
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()
		ctx = ctxutil.SetTraceID(ctx, traceID)
		ctx = ctxutil.SetSpanID(ctx, spanID)

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
	// First, check ctxutil context (set by Tracing middleware)
	if id := ctxutil.GetTraceID(ctx); id != "" && id != ctxutil.EmptyTraceID {
		return id
	}

	// Fallback to span context in case middleware order changes or trace ID is only in the span.
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.HasTraceID() {
		return ""
	}

	traceID := spanCtx.TraceID().String()
	if traceID == ctxutil.EmptyTraceID {
		return ""
	}

	return traceID
}

// GetSpanID retrieves the span ID from the context.
// Returns an empty string (16 zero chars) if no active span is present.
// Span ID format: 16 hex characters (64 bits).
func GetSpanID(ctx context.Context) string {
	// First, check ctxutil context (set by Tracing middleware)
	if id := ctxutil.GetSpanID(ctx); id != "" && id != ctxutil.EmptySpanID {
		return id
	}

	// Fallback to OTel span context
	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.HasSpanID() {
		return ""
	}
	spanID := spanCtx.SpanID().String()
	// Check for zero span ID (invalid/no tracing)
	if spanID == ctxutil.EmptySpanID {
		return ""
	}
	return spanID
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
