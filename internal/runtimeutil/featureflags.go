// Package runtimeutil provides runtime utility interfaces for external services.
package runtimeutil

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"time"
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

// FeatureFlagState represents the state of a feature flag including metadata.
type FeatureFlagState struct {
	Name        string `json:"name"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description,omitempty"`
	UpdatedAt   string `json:"updated_at"` // ISO 8601 format
}

// AdminFeatureFlagProvider extends FeatureFlagProvider with write operations
// for administrative feature flag management.
//
// Usage Example:
//
//	// Enable a flag
//	err := provider.SetEnabled(ctx, "new_dashboard", true)
//
//	// List all flags
//	flags, err := provider.List(ctx)
//
//	// Get specific flag
//	state, err := provider.Get(ctx, "new_dashboard")
type AdminFeatureFlagProvider interface {
	FeatureFlagProvider

	// SetEnabled updates a flag's enabled state.
	// Returns ErrFlagNotFound if the flag doesn't exist.
	SetEnabled(ctx context.Context, flag string, enabled bool) error

	// List returns all known feature flags with their current states.
	List(ctx context.Context) ([]FeatureFlagState, error)

	// Get returns a specific flag's state.
	// Returns ErrFlagNotFound if the flag doesn't exist.
	Get(ctx context.Context, flag string) (FeatureFlagState, error)
}

// InMemoryFeatureFlagStore implements AdminFeatureFlagProvider with in-memory storage.
// Thread-safe via sync.RWMutex. State changes are lost on restart.
//
// This implementation is suitable for development and testing.
// For production, use a persistent store (Redis, database).
type InMemoryFeatureFlagStore struct {
	mu          sync.RWMutex
	flags       map[string]FeatureFlagState
	envProvider *EnvFeatureFlagProvider
}

// InMemoryFFOption is a functional option for configuring InMemoryFeatureFlagStore.
type InMemoryFFOption func(*InMemoryFeatureFlagStore)

// NewInMemoryFeatureFlagStore creates a new InMemoryFeatureFlagStore.
// It initializes from the provided EnvFeatureFlagProvider for reading initial env values.
//
// Example:
//
//	store := runtimeutil.NewInMemoryFeatureFlagStore(
//	    runtimeutil.WithInitialFlags(map[string]bool{
//	        "new_dashboard": true,
//	        "dark_mode": false,
//	    }),
//	)
func NewInMemoryFeatureFlagStore(opts ...InMemoryFFOption) *InMemoryFeatureFlagStore {
	s := &InMemoryFeatureFlagStore{
		flags: make(map[string]FeatureFlagState),
		envProvider: &EnvFeatureFlagProvider{
			prefix:       "FF_",
			defaultValue: false,
			strictMode:   false,
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithInitialFlags sets initial flags when creating the store.
func WithInitialFlags(flags map[string]bool) InMemoryFFOption {
	return func(s *InMemoryFeatureFlagStore) {
		now := timeNow().Format("2006-01-02T15:04:05Z07:00")
		for name, enabled := range flags {
			s.flags[name] = FeatureFlagState{
				Name:      name,
				Enabled:   enabled,
				UpdatedAt: now,
			}
		}
	}
}

// WithFlagDescriptions sets descriptions for flags.
func WithFlagDescriptions(descriptions map[string]string) InMemoryFFOption {
	return func(s *InMemoryFeatureFlagStore) {
		for name, desc := range descriptions {
			if state, exists := s.flags[name]; exists {
				state.Description = desc
				s.flags[name] = state
			}
		}
	}
}

// timeNow is a variable for testing purposes.
var timeNow = time.Now

// IsEnabled checks if a feature flag is enabled.
// First checks dynamic state, falls back to env provider.
func (s *InMemoryFeatureFlagStore) IsEnabled(ctx context.Context, flag string) (bool, error) {
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err := validateFlagName(flag); err != nil {
		return false, err
	}

	s.mu.RLock()
	state, exists := s.flags[flag]
	s.mu.RUnlock()

	if exists {
		return state.Enabled, nil
	}

	// Fall back to env provider
	return s.envProvider.IsEnabled(ctx, flag)
}

// IsEnabledForContext checks if a feature flag is enabled with evaluation context.
// Same as IsEnabled - context evaluation not supported in in-memory store.
func (s *InMemoryFeatureFlagStore) IsEnabledForContext(ctx context.Context, flag string, _ EvalContext) (bool, error) {
	return s.IsEnabled(ctx, flag)
}

// SetEnabled updates a flag's enabled state.
// Creates the flag if it doesn't exist (auto-registration).
func (s *InMemoryFeatureFlagStore) SetEnabled(ctx context.Context, flag string, enabled bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if err := validateFlagName(flag); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := timeNow().Format("2006-01-02T15:04:05Z07:00")

	state, exists := s.flags[flag]
	if !exists {
		// Auto-register unknown flags
		state = FeatureFlagState{
			Name: flag,
		}
	}

	state.Enabled = enabled
	state.UpdatedAt = now
	s.flags[flag] = state

	return nil
}

// List returns all known feature flags.
func (s *InMemoryFeatureFlagStore) List(ctx context.Context) ([]FeatureFlagState, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]FeatureFlagState, 0, len(s.flags))
	for _, state := range s.flags {
		result = append(result, state)
	}

	return result, nil
}

// Get returns a specific flag's state.
// Falls back to environment provider if not in store (read-only state).
func (s *InMemoryFeatureFlagStore) Get(ctx context.Context, flag string) (FeatureFlagState, error) {
	if ctx.Err() != nil {
		return FeatureFlagState{}, ctx.Err()
	}

	if err := validateFlagName(flag); err != nil {
		return FeatureFlagState{}, err
	}

	s.mu.RLock()
	state, exists := s.flags[flag]
	s.mu.RUnlock()

	if exists {
		return state, nil
	}

	// Fall back to env provider for consistency with IsEnabled behavior
	enabled, err := s.envProvider.IsEnabled(ctx, flag)
	if err != nil {
		// If env provider returns ErrFlagNotFound, propagate it
		return FeatureFlagState{}, err
	}

	// Synthesize read-only state from env
	return FeatureFlagState{
		Name:        flag,
		Enabled:     enabled,
		Description: "(from environment)",
		UpdatedAt:   "", // Unknown for env-based flags
	}, nil
}
