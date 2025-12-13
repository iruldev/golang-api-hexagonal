package runtimeutil_test

import (
	"context"
	"fmt"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/runtimeutil"
)

// ExampleNewEnvFeatureFlagProvider demonstrates basic feature flag usage.
func ExampleNewEnvFeatureFlagProvider() {
	// Set environment variable for the example
	os.Setenv("FF_NEW_DASHBOARD", "true")
	defer os.Unsetenv("FF_NEW_DASHBOARD")

	// Create provider with default settings
	provider := runtimeutil.NewEnvFeatureFlagProvider()

	// Check if feature is enabled
	enabled, err := provider.IsEnabled(context.Background(), "new_dashboard")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if enabled {
		fmt.Println("New dashboard is enabled!")
	} else {
		fmt.Println("New dashboard is disabled")
	}
	// Output: New dashboard is enabled!
}

// ExampleNewEnvFeatureFlagProvider_customPrefix demonstrates using a custom prefix.
func ExampleNewEnvFeatureFlagProvider_customPrefix() {
	// Set environment variable with custom prefix
	os.Setenv("FEATURE_DARK_MODE", "enabled")
	defer os.Unsetenv("FEATURE_DARK_MODE")

	// Create provider with custom prefix
	provider := runtimeutil.NewEnvFeatureFlagProvider(
		runtimeutil.WithEnvPrefix("FEATURE_"),
	)

	enabled, _ := provider.IsEnabled(context.Background(), "dark_mode")
	fmt.Println("Dark mode enabled:", enabled)
	// Output: Dark mode enabled: true
}

// ExampleNewEnvFeatureFlagProvider_defaultValue demonstrates fail-open behavior.
func ExampleNewEnvFeatureFlagProvider_defaultValue() {
	// No environment variable set - uses default value
	provider := runtimeutil.NewEnvFeatureFlagProvider(
		runtimeutil.WithEnvDefaultValue(true), // Fail-open: unconfigured flags are enabled
	)

	enabled, _ := provider.IsEnabled(context.Background(), "unconfigured_feature")
	fmt.Println("Unconfigured feature enabled:", enabled)
	// Output: Unconfigured feature enabled: true
}

// ExampleNewEnvFeatureFlagProvider_strictMode demonstrates strict mode error handling.
func ExampleNewEnvFeatureFlagProvider_strictMode() {
	// Create provider with strict mode
	provider := runtimeutil.NewEnvFeatureFlagProvider(
		runtimeutil.WithEnvStrictMode(true),
	)

	// Try to check an unconfigured flag
	_, err := provider.IsEnabled(context.Background(), "missing_flag")
	if err != nil {
		fmt.Println("Error:", err)
	}
	// Output: Error: featureflag: flag not found
}

// ExampleNewEnvFeatureFlagProvider_contextBased demonstrates context-based evaluation.
func ExampleNewEnvFeatureFlagProvider_contextBased() {
	os.Setenv("FF_BETA_FEATURE", "true")
	defer os.Unsetenv("FF_BETA_FEATURE")

	provider := runtimeutil.NewEnvFeatureFlagProvider()

	// Create evaluation context for user targeting
	evalCtx := runtimeutil.EvalContext{
		UserID: "user-123",
		Attributes: map[string]interface{}{
			"plan":    "premium",
			"country": "US",
		},
		Percentage: 50.0, // For gradual rollouts
	}

	// Check flag with context
	// Note: EnvProvider ignores context, but interface supports it
	enabled, _ := provider.IsEnabledForContext(context.Background(), "beta_feature", evalCtx)
	fmt.Println("Beta feature for user-123:", enabled)
	// Output: Beta feature for user-123: true
}

// ExampleNewNopFeatureFlagProvider demonstrates testing with NopProvider.
func ExampleNewNopFeatureFlagProvider() {
	// Create a provider that returns false for all flags (for testing)
	provider := runtimeutil.NewNopFeatureFlagProvider(false)

	// All flags will be disabled
	enabled, _ := provider.IsEnabled(context.Background(), "any_feature")
	fmt.Println("Any feature enabled:", enabled)
	// Output: Any feature enabled: false
}

// ExampleNewNopFeatureFlagProvider_allEnabled demonstrates testing with all flags enabled.
func ExampleNewNopFeatureFlagProvider_allEnabled() {
	// Create a provider that returns true for all flags (for testing)
	provider := runtimeutil.NewNopFeatureFlagProvider(true)

	enabled1, _ := provider.IsEnabled(context.Background(), "feature_a")
	enabled2, _ := provider.IsEnabled(context.Background(), "feature_b")
	fmt.Println("Feature A:", enabled1, "Feature B:", enabled2)
	// Output: Feature A: true Feature B: true
}

// Example_httpHandler demonstrates using feature flags in an HTTP handler.
func Example_httpHandler() {
	os.Setenv("FF_NEW_UI", "true")
	defer os.Unsetenv("FF_NEW_UI")

	provider := runtimeutil.NewEnvFeatureFlagProvider()

	// Simulated handler logic
	handleRequest := func() string {
		enabled, _ := provider.IsEnabled(context.Background(), "new_ui")
		if enabled {
			return "Rendering new UI"
		}
		return "Rendering classic UI"
	}

	result := handleRequest()
	fmt.Println(result)
	// Output: Rendering new UI
}
