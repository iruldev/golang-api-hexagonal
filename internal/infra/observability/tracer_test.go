package observability

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

func TestInitTracer_Disabled(t *testing.T) {
	// Given: OTEL is disabled
	cfg := &config.Config{
		OTELEnabled:          false,
		OTELExporterEndpoint: "",
		OTELExporterInsecure: true,
		ServiceName:          "test-service",
		Env:                  "test",
	}

	// When: InitTracer is called
	tp, err := InitTracer(context.Background(), cfg)

	// Then: no error and a noop provider is returned
	require.NoError(t, err)
	require.NotNil(t, tp)

	// Verify it's a valid tracer provider (can create tracer)
	tracer := tp.Tracer("test")
	assert.NotNil(t, tracer)

	// Cleanup
	err = tp.Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestInitTracer_Enabled_WithValidEndpoint(t *testing.T) {
	// Given: OTEL is enabled with an endpoint
	cfg := &config.Config{
		OTELEnabled:          true,
		OTELExporterEndpoint: "localhost:4317",
		OTELExporterInsecure: true,
		ServiceName:          "test-service",
		Env:                  "test",
	}

	// When: InitTracer is called
	tp, err := InitTracer(context.Background(), cfg)

	// Then: no error and a configured provider is returned
	require.NoError(t, err)
	require.NotNil(t, tp)

	// Verify it's a valid tracer provider (can create tracer)
	tracer := tp.Tracer("test")
	assert.NotNil(t, tracer)

	// Cleanup - shutdown may warn about connection but shouldn't error
	_ = tp.Shutdown(context.Background())
}

func TestInitTracer_ReturnsTracerProvider(t *testing.T) {
	tests := []struct {
		name        string
		otelEnabled bool
		endpoint    string
	}{
		{
			name:        "disabled returns noop provider",
			otelEnabled: false,
			endpoint:    "",
		},
		{
			name:        "enabled with endpoint returns configured provider",
			otelEnabled: true,
			endpoint:    "localhost:4317",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				OTELEnabled:          tt.otelEnabled,
				OTELExporterEndpoint: tt.endpoint,
				OTELExporterInsecure: true,
				ServiceName:          "test-service",
				Env:                  "test",
			}

			tp, err := InitTracer(context.Background(), cfg)

			require.NoError(t, err)
			require.NotNil(t, tp)

			// Verify provider is of correct type
			var _ *sdktrace.TracerProvider = tp

			// Should be able to create a tracer
			tracer := tp.Tracer("test-tracer")
			assert.NotNil(t, tracer)

			// Cleanup
			_ = tp.Shutdown(context.Background())
		})
	}
}
