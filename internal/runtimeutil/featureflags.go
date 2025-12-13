// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"errors"
	"os"
	"strings"
)

// ErrFlagNotFound indicates the feature flag was not found (used in strict mode).
var ErrFlagNotFound = errors.New("featureflag: flag not found")

// ErrInvalidFlagName indicates the flag name is invalid.
var ErrInvalidFlagName = errors.New("featureflag: invalid flag name")

// EvalContext provides context for advanced feature flag evaluation.
// Used for percentage rollouts, user targeting, and attribute-based flags.
type EvalContext struct {
	// UserID for user-specific targeting
	UserID string

	// Attributes for custom targeting rules
	Attributes map[string]interface{}

	// Percentage for gradual rollouts (0-100)
	// Some providers use this for consistent hashing
	Percentage float64
}

// FeatureFlagProvider defines feature flag abstraction for swappable implementations.
// Implement this interface for LaunchDarkly, Split.io, ConfigCat, or other providers.
//
// Usage Example:
//
//	// Check if feature is enabled
//	enabled, err := provider.IsEnabled(ctx, "new_dashboard")
//	if enabled {
//	    // Render new dashboard
//	}
//
//	// Context-based evaluation for user targeting
//	evalCtx := runtimeutil.EvalContext{UserID: "user-123"}
//	enabled, err := provider.IsEnabledForContext(ctx, "beta_feature", evalCtx)
type FeatureFlagProvider interface {
	// IsEnabled checks if a feature flag is enabled.
	// Returns the flag value and any error (e.g., ErrFlagNotFound in strict mode).
	IsEnabled(ctx context.Context, flag string) (bool, error)

	// IsEnabledForContext checks if a feature flag is enabled with evaluation context.
	// Use for user targeting, percentage rollouts, or attribute-based rules.
	IsEnabledForContext(ctx context.Context, flag string, evalContext EvalContext) (bool, error)
}

// EnvFFOption is a functional option for configuring EnvFeatureFlagProvider.
type EnvFFOption func(*EnvFeatureFlagProvider)

// EnvFeatureFlagProvider reads feature flags from environment variables.
// This is the default implementation for simple feature flag usage.
type EnvFeatureFlagProvider struct {
	prefix       string // Default: "FF_"
	defaultValue bool   // Default: false (fail-closed)
	strictMode   bool   // Default: false
}

// NewEnvFeatureFlagProvider creates a new EnvFeatureFlagProvider with the given options.
//
// Example:
//
//	provider := runtimeutil.NewEnvFeatureFlagProvider(
//	    runtimeutil.WithEnvPrefix("FEATURE_"),
//	    runtimeutil.WithEnvDefaultValue(true),
//	)
//	enabled, _ := provider.IsEnabled(ctx, "dark_mode")
func NewEnvFeatureFlagProvider(opts ...EnvFFOption) FeatureFlagProvider {
	p := &EnvFeatureFlagProvider{
		prefix:       "FF_",
		defaultValue: false,
		strictMode:   false,
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// IsEnabled checks if a feature flag is enabled by reading the corresponding environment variable.
// The flag name is converted to uppercase with the configured prefix (default: FF_).
// For example, "my_feature" becomes "FF_MY_FEATURE".
func (p *EnvFeatureFlagProvider) IsEnabled(ctx context.Context, flag string) (bool, error) {
	// Check context cancellation first
	if ctx.Err() != nil {
		return p.defaultValue, ctx.Err()
	}

	if err := validateFlagName(flag); err != nil {
		return p.defaultValue, err
	}

	envKey := p.flagToEnvKey(flag)
	value := os.Getenv(envKey)

	if value == "" {
		if p.strictMode {
			return p.defaultValue, ErrFlagNotFound
		}
		return p.defaultValue, nil
	}

	return parseTruthy(value), nil
}

// IsEnabledForContext checks if a feature flag is enabled with evaluation context.
// Note: The env provider does not support context-based evaluation, so this method
// behaves the same as IsEnabled. Use a more advanced provider (e.g., LaunchDarkly)
// for user targeting and percentage rollouts.
func (p *EnvFeatureFlagProvider) IsEnabledForContext(ctx context.Context, flag string, _ EvalContext) (bool, error) {
	return p.IsEnabled(ctx, flag)
}

// flagToEnvKey converts a flag name to an environment variable key.
// Example: "my_feature" -> "FF_MY_FEATURE", "beta-feature" -> "FF_BETA_FEATURE"
func (p *EnvFeatureFlagProvider) flagToEnvKey(flag string) string {
	upper := strings.ToUpper(strings.ReplaceAll(flag, "-", "_"))
	return p.prefix + upper
}

// validateFlagName checks if the flag name is valid.
func validateFlagName(flag string) error {
	if flag == "" {
		return ErrInvalidFlagName
	}
	// Check for invalid characters (only alphanumeric, underscore, hyphen allowed)
	for _, c := range flag {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-') {
			return ErrInvalidFlagName
		}
	}
	return nil
}

// parseTruthy parses a string value as a boolean.
// Returns true for: "true", "1", "enabled", "on", "yes" (case-insensitive)
// Returns false for all other values.
func parseTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "enabled", "on", "yes":
		return true
	default:
		return false
	}
}

// WithEnvPrefix sets the environment variable prefix.
// Default is "FF_".
//
// Example:
//
//	provider := runtimeutil.NewEnvFeatureFlagProvider(
//	    runtimeutil.WithEnvPrefix("FEATURE_"),
//	)
//	// "dark_mode" will check "FEATURE_DARK_MODE"
func WithEnvPrefix(prefix string) EnvFFOption {
	return func(p *EnvFeatureFlagProvider) {
		p.prefix = prefix
	}
}

// WithEnvDefaultValue sets the default value for unconfigured flags.
// Default is false (fail-closed).
//
// Example:
//
//	provider := runtimeutil.NewEnvFeatureFlagProvider(
//	    runtimeutil.WithEnvDefaultValue(true), // All unconfigured flags default to enabled
//	)
func WithEnvDefaultValue(defaultValue bool) EnvFFOption {
	return func(p *EnvFeatureFlagProvider) {
		p.defaultValue = defaultValue
	}
}

// WithEnvStrictMode enables strict mode where unknown flags return ErrFlagNotFound.
// Default is false (silent failure with default value).
//
// Example:
//
//	provider := runtimeutil.NewEnvFeatureFlagProvider(
//	    runtimeutil.WithEnvStrictMode(true),
//	)
//	enabled, err := provider.IsEnabled(ctx, "unknown_flag")
//	// err == ErrFlagNotFound
func WithEnvStrictMode(strict bool) EnvFFOption {
	return func(p *EnvFeatureFlagProvider) {
		p.strictMode = strict
	}
}

// NopFeatureFlagProvider is a no-op provider that returns a fixed value.
// Use for testing or when feature flags should be disabled.
type NopFeatureFlagProvider struct {
	defaultEnabled bool
}

// NewNopFeatureFlagProvider creates a new NopFeatureFlagProvider.
// The defaultEnabled parameter determines what value is returned for all flags.
//
// Example:
//
//	// All flags disabled (for testing)
//	provider := runtimeutil.NewNopFeatureFlagProvider(false)
//
//	// All flags enabled (for testing)
//	provider := runtimeutil.NewNopFeatureFlagProvider(true)
func NewNopFeatureFlagProvider(defaultEnabled bool) FeatureFlagProvider {
	return &NopFeatureFlagProvider{defaultEnabled: defaultEnabled}
}

// IsEnabled returns the configured default value for any flag.
// Validates flag name for consistency with other providers.
func (p *NopFeatureFlagProvider) IsEnabled(ctx context.Context, flag string) (bool, error) {
	// Check context cancellation first
	if ctx.Err() != nil {
		return p.defaultEnabled, ctx.Err()
	}

	// Validate flag name for consistency with EnvFeatureFlagProvider
	if err := validateFlagName(flag); err != nil {
		return p.defaultEnabled, err
	}

	return p.defaultEnabled, nil
}

// IsEnabledForContext returns the configured default value for any flag.
// Validates flag name for consistency with other providers.
func (p *NopFeatureFlagProvider) IsEnabledForContext(ctx context.Context, flag string, _ EvalContext) (bool, error) {
	return p.IsEnabled(ctx, flag)
}
