// Package observability provides logging, tracing, and metrics utilities.
package observability

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

// InitTracer initializes the OpenTelemetry tracer provider based on configuration.
// When OTEL_ENABLED is false, it returns a noop tracer provider that doesn't export spans.
// When enabled, it configures an OTLP gRPC exporter to send traces to the configured endpoint.
//
// The caller is responsible for calling Shutdown on the returned TracerProvider
// during graceful shutdown to ensure all spans are flushed.
func InitTracer(ctx context.Context, cfg *config.Config) (*sdktrace.TracerProvider, error) {
	const op = "observability.InitTracer"

	if !cfg.OTELEnabled {
		// Return noop provider - no spans are exported
		return sdktrace.NewTracerProvider(), nil
	}

	// Create OTLP gRPC exporter
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTELExporterEndpoint),
	}

	if cfg.OTELExporterInsecure || isLocalEndpoint(cfg.OTELExporterEndpoint) {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create exporter: %w", op, err)
	}

	// Create resource with service metadata
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create resource: %w", op, err)
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global propagator for W3C Trace Context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

// isLocalEndpoint returns true when the endpoint looks like a local collector (no TLS).
func isLocalEndpoint(endpoint string) bool {
	endpoint = strings.TrimSpace(endpoint)
	return strings.HasPrefix(endpoint, "localhost:") || strings.HasPrefix(endpoint, "127.0.0.1:")
}
