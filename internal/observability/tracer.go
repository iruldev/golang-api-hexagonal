// Package observability provides observability utilities including logging and tracing.
package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

// NewTracerProvider creates a new OpenTelemetry tracer provider.
// Returns the provider, a shutdown function, and any error.
// The shutdown function should be called during graceful shutdown.
func NewTracerProvider(ctx context.Context, cfg *config.ObservabilityConfig) (*sdktrace.TracerProvider, func(context.Context) error, error) {
	// Create OTLP exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.ExporterEndpoint),
		otlptracehttp.WithInsecure(), // For local dev - use TLS in production
	)
	if err != nil {
		return nil, nil, err
	}

	// Create resource with service name
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, tp.Shutdown, nil
}

// GetTraceID extracts trace ID from context.
// Returns empty string if no trace is present.
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().HasTraceID() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// GetSpan returns the current span from context.
// Use this to create child spans or add attributes.
func GetSpan(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// StartSpan starts a new child span with the given name.
// Returns the context with the new span and a function to end the span.
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	tracer := otel.Tracer("golang-api-hexagonal")
	return tracer.Start(ctx, name)
}
