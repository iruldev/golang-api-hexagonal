package fxmodule

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/resilience"
)

// TestResilienceModule_ProvidesAllDependencies tests that the ResilienceModule
// correctly provides all expected dependencies for injection.
func TestResilienceModule_ProvidesAllDependencies(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")

	app := fxtest.New(t,
		fx.Provide(config.Load),
		fx.Provide(func() *prometheus.Registry {
			return prometheus.NewRegistry()
		}),
		fx.Provide(func() *slog.Logger {
			return slog.Default()
		}),
		// Provide resilience components directly (not via module to avoid init conflicts)
		fx.Provide(provideResilienceConfig),
		fx.Provide(provideCircuitBreakerMetrics),
		fx.Provide(provideCircuitBreakerPresets),
		fx.Provide(provideRetryMetrics),
		fx.Provide(provideRetrier),
		fx.Provide(provideTimeoutMetrics),
		fx.Provide(provideTimeoutPresets),
		fx.Provide(provideBulkheadMetrics),
		fx.Provide(provideBulkheadPresets),
		fx.Provide(provideShutdownMetrics),
		fx.Provide(provideShutdownCoordinator),
		fx.Provide(provideResilienceWrapper),
		fx.Invoke(func(
			resCfg resilience.ResilienceConfig,
			cbPresets *resilience.CircuitBreakerPresets,
			retrier resilience.Retrier,
			timeoutPresets *resilience.TimeoutPresets,
			bulkheadPresets *resilience.BulkheadPresets,
			shutdownCoord resilience.ShutdownCoordinator,
			wrapper resilience.ResilienceWrapper,
		) {
			// Verify ResilienceConfig is populated
			if resCfg.CircuitBreaker.MaxRequests == 0 {
				t.Error("CircuitBreaker config not loaded")
			}
			if resCfg.Retry.MaxAttempts == 0 {
				t.Error("Retry config not loaded")
			}
			if resCfg.Timeout.Default == 0 {
				t.Error("Timeout config not loaded")
			}
			if resCfg.Bulkhead.MaxConcurrent == 0 {
				t.Error("Bulkhead config not loaded")
			}

			// Verify CircuitBreaker presets
			if cbPresets == nil {
				t.Error("CircuitBreaker presets not provided")
			}
			cb := cbPresets.ForDatabase()
			if cb == nil {
				t.Error("CircuitBreaker.ForDatabase returned nil")
			}
			if cb.Name() != "database" {
				t.Errorf("Expected CB name 'database', got '%s'", cb.Name())
			}

			// Verify Retrier
			if retrier == nil {
				t.Error("Retrier not provided")
			}
			if retrier.Name() != "default" {
				t.Errorf("Expected retrier name 'default', got '%s'", retrier.Name())
			}

			// Verify Timeout presets
			if timeoutPresets == nil {
				t.Error("Timeout presets not provided")
			}
			timeout := timeoutPresets.ForDatabase()
			if timeout == nil {
				t.Error("TimeoutPresets.ForDatabase returned nil")
			}

			// Verify Bulkhead presets
			if bulkheadPresets == nil {
				t.Error("Bulkhead presets not provided")
			}
			bulkhead := bulkheadPresets.ForDatabase()
			if bulkhead == nil {
				t.Error("BulkheadPresets.ForDatabase returned nil")
			}

			// Verify ShutdownCoordinator
			if shutdownCoord == nil {
				t.Error("ShutdownCoordinator not provided")
			}

			// Verify ResilienceWrapper
			if wrapper == nil {
				t.Error("ResilienceWrapper not provided")
			}
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.RequireStart()

	select {
	case <-ctx.Done():
		t.Fatal("App start timed out")
	default:
	}

	app.RequireStop()
}

// TestResilienceModule_ComponentsUseConfiguration verifies that injected
// components are configured based on ResilienceConfig values.
func TestResilienceModule_ComponentsUseConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")
	// Set custom configuration values
	t.Setenv("TIMEOUT_DEFAULT", "3s")
	t.Setenv("TIMEOUT_DATABASE", "2s")
	t.Setenv("TIMEOUT_EXTERNAL_API", "10s")

	app := fxtest.New(t,
		fx.Provide(config.Load),
		fx.Provide(func() *prometheus.Registry {
			return prometheus.NewRegistry()
		}),
		fx.Provide(func() *slog.Logger {
			return slog.Default()
		}),
		fx.Provide(provideResilienceConfig),
		fx.Provide(provideTimeoutMetrics),
		fx.Provide(provideTimeoutPresets),
		fx.Invoke(func(
			timeoutPresets *resilience.TimeoutPresets,
		) {
			// Verify timeouts are configured from environment
			if timeoutPresets.DefaultDuration() != 3*time.Second {
				t.Errorf("Expected default timeout 3s, got %v", timeoutPresets.DefaultDuration())
			}
			if timeoutPresets.DatabaseDuration() != 2*time.Second {
				t.Errorf("Expected database timeout 2s, got %v", timeoutPresets.DatabaseDuration())
			}
			if timeoutPresets.ExternalAPIDuration() != 10*time.Second {
				t.Errorf("Expected external API timeout 10s, got %v", timeoutPresets.ExternalAPIDuration())
			}
		}),
	)

	app.RequireStart()
	app.RequireStop()
}

// TestResilienceModule_WrapperComposesComponents verifies that the
// ResilienceWrapper correctly composes all resilience components.
func TestResilienceModule_WrapperComposesComponents(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
	t.Setenv("JWT_SECRET", "test-secret-key-at-least-32-chars")

	var wrapper resilience.ResilienceWrapper

	app := fxtest.New(t,
		fx.Provide(config.Load),
		fx.Provide(func() *prometheus.Registry {
			return prometheus.NewRegistry()
		}),
		fx.Provide(func() *slog.Logger {
			return slog.Default()
		}),
		fx.Provide(provideResilienceConfig),
		fx.Provide(provideCircuitBreakerMetrics),
		fx.Provide(provideCircuitBreakerPresets),
		fx.Provide(provideRetryMetrics),
		fx.Provide(provideRetrier),
		fx.Provide(provideTimeoutMetrics),
		fx.Provide(provideTimeoutPresets),
		fx.Provide(provideBulkheadMetrics),
		fx.Provide(provideBulkheadPresets),
		fx.Provide(provideShutdownMetrics),
		fx.Provide(provideShutdownCoordinator),
		fx.Provide(provideResilienceWrapper),
		fx.Populate(&wrapper),
	)

	app.RequireStart()
	defer app.RequireStop()

	// Execute operation through wrapper
	called := false
	err := wrapper.Execute(context.Background(), "test-operation", func(ctx context.Context) error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !called {
		t.Error("Operation was not executed")
	}
}
