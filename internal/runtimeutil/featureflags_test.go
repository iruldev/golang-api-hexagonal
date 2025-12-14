package runtimeutil

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

func TestEnvFeatureFlagProvider_IsEnabled(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		flag         string
		opts         []EnvFFOption
		want         bool
		wantErr      error
		skipEnvSetup bool
	}{
		// Truthy values
		{
			name:     "true value",
			envVar:   "FF_MY_FEATURE",
			envValue: "true",
			flag:     "my_feature",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "TRUE uppercase",
			envVar:   "FF_MY_FEATURE",
			envValue: "TRUE",
			flag:     "my_feature",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "1 value",
			envVar:   "FF_MY_FEATURE",
			envValue: "1",
			flag:     "my_feature",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "enabled value",
			envVar:   "FF_DARK_MODE",
			envValue: "enabled",
			flag:     "dark_mode",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "ENABLED uppercase",
			envVar:   "FF_DARK_MODE",
			envValue: "ENABLED",
			flag:     "dark_mode",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "on value",
			envVar:   "FF_BETA",
			envValue: "on",
			flag:     "beta",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "yes value",
			envVar:   "FF_NEW_UI",
			envValue: "yes",
			flag:     "new_ui",
			want:     true,
			wantErr:  nil,
		},
		{
			name:     "true with whitespace",
			envVar:   "FF_FEATURE",
			envValue: "  true  ",
			flag:     "feature",
			want:     true,
			wantErr:  nil,
		},
		// Falsy values
		{
			name:     "false value",
			envVar:   "FF_MY_FEATURE",
			envValue: "false",
			flag:     "my_feature",
			want:     false,
			wantErr:  nil,
		},
		{
			name:     "0 value",
			envVar:   "FF_MY_FEATURE",
			envValue: "0",
			flag:     "my_feature",
			want:     false,
			wantErr:  nil,
		},
		{
			name:     "disabled value",
			envVar:   "FF_MY_FEATURE",
			envValue: "disabled",
			flag:     "my_feature",
			want:     false,
			wantErr:  nil,
		},
		{
			name:     "off value",
			envVar:   "FF_MY_FEATURE",
			envValue: "off",
			flag:     "my_feature",
			want:     false,
			wantErr:  nil,
		},
		{
			name:     "no value",
			envVar:   "FF_MY_FEATURE",
			envValue: "no",
			flag:     "my_feature",
			want:     false,
			wantErr:  nil,
		},
		{
			name:     "random string",
			envVar:   "FF_MY_FEATURE",
			envValue: "maybe",
			flag:     "my_feature",
			want:     false,
			wantErr:  nil,
		},
		// Unset env var (default behavior)
		{
			name:         "not set returns default false",
			envVar:       "",
			envValue:     "",
			flag:         "unknown_flag",
			want:         false,
			wantErr:      nil,
			skipEnvSetup: true,
		},
		// Hyphen to underscore conversion
		{
			name:     "hyphen to underscore",
			envVar:   "FF_BETA_FEATURE",
			envValue: "true",
			flag:     "beta-feature",
			want:     true,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if !tt.skipEnvSetup && tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envValue)
			}

			provider := NewEnvFeatureFlagProvider(tt.opts...)
			got, err := provider.IsEnabled(context.Background(), tt.flag)

			// Assert
			if tt.wantErr != nil {
				if err == nil || err.Error() != tt.wantErr.Error() {
					t.Errorf("IsEnabled() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("IsEnabled() unexpected error = %v", err)
			}

			if got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvFeatureFlagProvider_WithCustomPrefix(t *testing.T) {
	t.Setenv("FEATURE_DARK_MODE", "true")

	provider := NewEnvFeatureFlagProvider(WithEnvPrefix("FEATURE_"))
	got, err := provider.IsEnabled(context.Background(), "dark_mode")

	if err != nil {
		t.Errorf("IsEnabled() unexpected error = %v", err)
	}
	if got != true {
		t.Errorf("IsEnabled() = %v, want true", got)
	}
}

func TestEnvFeatureFlagProvider_WithDefaultValue(t *testing.T) {
	// Don't set any env var - test default value behavior
	provider := NewEnvFeatureFlagProvider(WithEnvDefaultValue(true))
	got, err := provider.IsEnabled(context.Background(), "nonexistent_flag_12345")

	if err != nil {
		t.Errorf("IsEnabled() unexpected error = %v", err)
	}
	if got != true {
		t.Errorf("IsEnabled() = %v, want true (default)", got)
	}
}

func TestEnvFeatureFlagProvider_WithStrictMode(t *testing.T) {
	// Don't set any env var - test strict mode behavior
	provider := NewEnvFeatureFlagProvider(WithEnvStrictMode(true))
	got, err := provider.IsEnabled(context.Background(), "nonexistent_flag_12345")

	if err != ErrFlagNotFound {
		t.Errorf("IsEnabled() error = %v, want ErrFlagNotFound", err)
	}
	if got != false {
		t.Errorf("IsEnabled() = %v, want false", got)
	}
}

func TestEnvFeatureFlagProvider_StrictModeWithDefaultValue(t *testing.T) {
	// Test strict mode with custom default value
	provider := NewEnvFeatureFlagProvider(
		WithEnvStrictMode(true),
		WithEnvDefaultValue(true),
	)
	got, err := provider.IsEnabled(context.Background(), "nonexistent_flag_12345")

	if err != ErrFlagNotFound {
		t.Errorf("IsEnabled() error = %v, want ErrFlagNotFound", err)
	}
	// Should return the default value even though error is returned
	if got != true {
		t.Errorf("IsEnabled() = %v, want true (default)", got)
	}
}

func TestEnvFeatureFlagProvider_InvalidFlagName(t *testing.T) {
	tests := []struct {
		name string
		flag string
	}{
		{"empty flag", ""},
		{"flag with space", "my feature"},
		{"flag with special char", "my@feature"},
		{"flag with dot", "my.feature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewEnvFeatureFlagProvider()
			_, err := provider.IsEnabled(context.Background(), tt.flag)

			if err != ErrInvalidFlagName {
				t.Errorf("IsEnabled() error = %v, want ErrInvalidFlagName", err)
			}
		})
	}
}

func TestEnvFeatureFlagProvider_IsEnabledForContext(t *testing.T) {
	t.Setenv("FF_BETA_FEATURE", "true")

	provider := NewEnvFeatureFlagProvider()
	evalCtx := EvalContext{
		UserID: "user-123",
		Attributes: map[string]interface{}{
			"plan": "premium",
		},
		Percentage: 50.0,
	}

	// Note: EnvProvider ignores context - this tests the interface compliance
	got, err := provider.IsEnabledForContext(context.Background(), "beta_feature", evalCtx)

	if err != nil {
		t.Errorf("IsEnabledForContext() unexpected error = %v", err)
	}
	if got != true {
		t.Errorf("IsEnabledForContext() = %v, want true", got)
	}
}

func TestNopFeatureFlagProvider_AllDisabled(t *testing.T) {
	provider := NewNopFeatureFlagProvider(false)

	flags := []string{"feature1", "feature2", "any_flag"}
	for _, flag := range flags {
		got, err := provider.IsEnabled(context.Background(), flag)
		if err != nil {
			t.Errorf("IsEnabled(%s) unexpected error = %v", flag, err)
		}
		if got != false {
			t.Errorf("IsEnabled(%s) = %v, want false", flag, got)
		}
	}
}

func TestNopFeatureFlagProvider_AllEnabled(t *testing.T) {
	provider := NewNopFeatureFlagProvider(true)

	flags := []string{"feature1", "feature2", "any_flag"}
	for _, flag := range flags {
		got, err := provider.IsEnabled(context.Background(), flag)
		if err != nil {
			t.Errorf("IsEnabled(%s) unexpected error = %v", flag, err)
		}
		if got != true {
			t.Errorf("IsEnabled(%s) = %v, want true", flag, got)
		}
	}
}

func TestNopFeatureFlagProvider_IsEnabledForContext(t *testing.T) {
	provider := NewNopFeatureFlagProvider(true)
	evalCtx := EvalContext{UserID: "user-123"}

	got, err := provider.IsEnabledForContext(context.Background(), "any_flag", evalCtx)
	if err != nil {
		t.Errorf("IsEnabledForContext() unexpected error = %v", err)
	}
	if got != true {
		t.Errorf("IsEnabledForContext() = %v, want true", got)
	}
}

func TestEvalContext_Fields(t *testing.T) {
	ctx := EvalContext{
		UserID: "user-123",
		Attributes: map[string]interface{}{
			"plan":    "premium",
			"country": "US",
		},
		Percentage: 50.0,
	}

	if ctx.UserID != "user-123" {
		t.Errorf("UserID = %v, want user-123", ctx.UserID)
	}
	if ctx.Attributes["plan"] != "premium" {
		t.Errorf("Attributes[plan] = %v, want premium", ctx.Attributes["plan"])
	}
	if ctx.Percentage != 50.0 {
		t.Errorf("Percentage = %v, want 50.0", ctx.Percentage)
	}
}

func TestFlagToEnvKey_Conversion(t *testing.T) {
	tests := []struct {
		flag     string
		prefix   string
		expected string
	}{
		{"my_feature", "FF_", "FF_MY_FEATURE"},
		{"beta-feature", "FF_", "FF_BETA_FEATURE"},
		{"DarkMode", "FF_", "FF_DARKMODE"},
		{"feature", "FEATURE_", "FEATURE_FEATURE"},
		{"a-b_c", "FF_", "FF_A_B_C"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			p := &EnvFeatureFlagProvider{prefix: tt.prefix}
			got := p.flagToEnvKey(tt.flag)
			if got != tt.expected {
				t.Errorf("flagToEnvKey(%s) = %v, want %v", tt.flag, got, tt.expected)
			}
		})
	}
}

// Ensure empty env value returns default
func TestEnvFeatureFlagProvider_EmptyEnvValue(t *testing.T) {
	t.Setenv("FF_EMPTY_FLAG", "")

	provider := NewEnvFeatureFlagProvider()
	got, err := provider.IsEnabled(context.Background(), "empty_flag")

	if err != nil {
		t.Errorf("IsEnabled() unexpected error = %v", err)
	}
	// Empty string is treated as unset, returns default
	if got != false {
		t.Errorf("IsEnabled() = %v, want false (default)", got)
	}
}

// Test that provider implements interface
func TestProviderInterface(t *testing.T) {
	var _ FeatureFlagProvider = NewEnvFeatureFlagProvider()
	var _ FeatureFlagProvider = NewNopFeatureFlagProvider(false)
}

// Benchmark tests
func BenchmarkEnvFeatureFlagProvider_IsEnabled(b *testing.B) {
	os.Setenv("FF_BENCHMARK_FLAG", "true")
	defer os.Unsetenv("FF_BENCHMARK_FLAG")

	provider := NewEnvFeatureFlagProvider()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.IsEnabled(ctx, "benchmark_flag")
	}
}

func BenchmarkNopFeatureFlagProvider_IsEnabled(b *testing.B) {
	provider := NewNopFeatureFlagProvider(true)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = provider.IsEnabled(ctx, "any_flag")
	}
}

// Test context cancellation
func TestEnvFeatureFlagProvider_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	provider := NewEnvFeatureFlagProvider()
	_, err := provider.IsEnabled(ctx, "any_flag")

	if err != context.Canceled {
		t.Errorf("IsEnabled() error = %v, want context.Canceled", err)
	}
}

func TestNopFeatureFlagProvider_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	provider := NewNopFeatureFlagProvider(true)
	_, err := provider.IsEnabled(ctx, "any_flag")

	if err != context.Canceled {
		t.Errorf("IsEnabled() error = %v, want context.Canceled", err)
	}
}

// Test NopProvider validates flag name (consistency with EnvProvider)
func TestNopFeatureFlagProvider_InvalidFlagName(t *testing.T) {
	tests := []struct {
		name string
		flag string
	}{
		{"empty flag", ""},
		{"flag with space", "my feature"},
		{"flag with special char", "my@feature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewNopFeatureFlagProvider(true)
			_, err := provider.IsEnabled(context.Background(), tt.flag)

			if err != ErrInvalidFlagName {
				t.Errorf("IsEnabled() error = %v, want ErrInvalidFlagName", err)
			}
		})
	}
}

// Test NopProvider IsEnabledForContext validates flag name
func TestNopFeatureFlagProvider_IsEnabledForContext_InvalidFlag(t *testing.T) {
	provider := NewNopFeatureFlagProvider(true)
	evalCtx := EvalContext{UserID: "user-123"}

	_, err := provider.IsEnabledForContext(context.Background(), "", evalCtx)
	if err != ErrInvalidFlagName {
		t.Errorf("IsEnabledForContext() error = %v, want ErrInvalidFlagName", err)
	}
}

// =============================================================================
// InMemoryFeatureFlagStore Tests (Story 15.2)
// =============================================================================

func TestInMemoryFeatureFlagStore_NewStore(t *testing.T) {
	store := NewInMemoryFeatureFlagStore()

	if store == nil {
		t.Fatal("NewInMemoryFeatureFlagStore() returned nil")
	}
	if store.flags == nil {
		t.Error("flags map should be initialized")
	}
}

func TestInMemoryFeatureFlagStore_WithInitialFlags(t *testing.T) {
	// Mock time for consistent testing
	originalTimeNow := timeNow
	fixedTime := time.Date(2025, 12, 14, 22, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = originalTimeNow }()

	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"new_dashboard": true,
			"dark_mode":     false,
		}),
	)

	// Check new_dashboard is enabled
	enabled, err := store.IsEnabled(context.Background(), "new_dashboard")
	if err != nil {
		t.Errorf("IsEnabled(new_dashboard) error = %v", err)
	}
	if !enabled {
		t.Error("new_dashboard should be enabled")
	}

	// Check dark_mode is disabled
	disabled, err := store.IsEnabled(context.Background(), "dark_mode")
	if err != nil {
		t.Errorf("IsEnabled(dark_mode) error = %v", err)
	}
	if disabled {
		t.Error("dark_mode should be disabled")
	}
}

func TestInMemoryFeatureFlagStore_SetEnabled(t *testing.T) {
	// Mock time
	originalTimeNow := timeNow
	fixedTime := time.Date(2025, 12, 14, 22, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = originalTimeNow }()

	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"feature_a": false,
		}),
	)

	// Enable the flag
	err := store.SetEnabled(context.Background(), "feature_a", true)
	if err != nil {
		t.Errorf("SetEnabled() error = %v", err)
	}

	// Verify it's now enabled
	enabled, err := store.IsEnabled(context.Background(), "feature_a")
	if err != nil {
		t.Errorf("IsEnabled() error = %v", err)
	}
	if !enabled {
		t.Error("feature_a should be enabled after SetEnabled(true)")
	}

	// Disable it again
	err = store.SetEnabled(context.Background(), "feature_a", false)
	if err != nil {
		t.Errorf("SetEnabled() error = %v", err)
	}

	disabled, err := store.IsEnabled(context.Background(), "feature_a")
	if err != nil {
		t.Errorf("IsEnabled() error = %v", err)
	}
	if disabled {
		t.Error("feature_a should be disabled after SetEnabled(false)")
	}
}

func TestInMemoryFeatureFlagStore_SetEnabled_AutoRegisters(t *testing.T) {
	store := NewInMemoryFeatureFlagStore()

	// Set a flag that doesn't exist - should auto-register
	err := store.SetEnabled(context.Background(), "new_flag", true)
	if err != nil {
		t.Errorf("SetEnabled() error = %v", err)
	}

	enabled, err := store.IsEnabled(context.Background(), "new_flag")
	if err != nil {
		t.Errorf("IsEnabled() error = %v", err)
	}
	if !enabled {
		t.Error("new_flag should be enabled after SetEnabled")
	}
}

func TestInMemoryFeatureFlagStore_SetEnabled_InvalidFlagName(t *testing.T) {
	store := NewInMemoryFeatureFlagStore()

	err := store.SetEnabled(context.Background(), "", true)
	if err != ErrInvalidFlagName {
		t.Errorf("SetEnabled() error = %v, want ErrInvalidFlagName", err)
	}
}

func TestInMemoryFeatureFlagStore_List(t *testing.T) {
	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"flag_a": true,
			"flag_b": false,
		}),
	)

	flags, err := store.List(context.Background())
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(flags) != 2 {
		t.Errorf("List() returned %d flags, want 2", len(flags))
	}

	// Check that both flags are present
	flagMap := make(map[string]FeatureFlagState)
	for _, f := range flags {
		flagMap[f.Name] = f
	}

	if _, ok := flagMap["flag_a"]; !ok {
		t.Error("flag_a should be in list")
	}
	if _, ok := flagMap["flag_b"]; !ok {
		t.Error("flag_b should be in list")
	}
}

func TestInMemoryFeatureFlagStore_Get(t *testing.T) {
	// Mock time
	originalTimeNow := timeNow
	fixedTime := time.Date(2025, 12, 14, 22, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = originalTimeNow }()

	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"existing_flag": true,
		}),
	)

	state, err := store.Get(context.Background(), "existing_flag")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if state.Name != "existing_flag" {
		t.Errorf("Name = %v, want existing_flag", state.Name)
	}
	if !state.Enabled {
		t.Error("Enabled should be true")
	}
	if state.UpdatedAt == "" {
		t.Error("UpdatedAt should be set")
	}
}

func TestInMemoryFeatureFlagStore_Get_NotFound_StrictMode(t *testing.T) {
	// With default env provider (non-strict), Get() falls back to env and returns success
	// To test ErrFlagNotFound, we need to test with a flag that doesn't exist in either store or env
	// Since default EnvProvider returns false (not error) for missing flags,
	// ErrFlagNotFound is only returned in strict mode
	store := NewInMemoryFeatureFlagStore()

	// Flag not in store, but env provider returns default (false, nil)
	// So this should return a synthesized state, not ErrFlagNotFound
	state, err := store.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Errorf("Get() error = %v, want nil (env fallback)", err)
	}
	if state.Name != "nonexistent" {
		t.Errorf("Name = %v, want nonexistent", state.Name)
	}
	if state.Enabled != false {
		t.Error("Enabled should be false (default from env)")
	}
}

func TestInMemoryFeatureFlagStore_Get_InvalidFlagName(t *testing.T) {
	store := NewInMemoryFeatureFlagStore()

	_, err := store.Get(context.Background(), "")
	if err != ErrInvalidFlagName {
		t.Errorf("Get() error = %v, want ErrInvalidFlagName", err)
	}
}

func TestInMemoryFeatureFlagStore_WithFlagDescriptions(t *testing.T) {
	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"new_dashboard": true,
		}),
		WithFlagDescriptions(map[string]string{
			"new_dashboard": "New dashboard UI",
		}),
	)

	state, err := store.Get(context.Background(), "new_dashboard")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}

	if state.Description != "New dashboard UI" {
		t.Errorf("Description = %v, want 'New dashboard UI'", state.Description)
	}
}

func TestInMemoryFeatureFlagStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"concurrent_flag": false,
		}),
	)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(enabled bool) {
			defer wg.Done()
			_ = store.SetEnabled(context.Background(), "concurrent_flag", enabled)
		}(i%2 == 0)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.IsEnabled(context.Background(), "concurrent_flag")
		}()
	}

	// Concurrent List
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = store.List(context.Background())
		}()
	}

	wg.Wait()
	// If we get here without deadlock/race, the test passes
}

func TestInMemoryFeatureFlagStore_ContextCancelled(t *testing.T) {
	store := NewInMemoryFeatureFlagStore()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// All methods should return context.Canceled
	_, err := store.IsEnabled(ctx, "any_flag")
	if err != context.Canceled {
		t.Errorf("IsEnabled() error = %v, want context.Canceled", err)
	}

	err = store.SetEnabled(ctx, "any_flag", true)
	if err != context.Canceled {
		t.Errorf("SetEnabled() error = %v, want context.Canceled", err)
	}

	_, err = store.List(ctx)
	if err != context.Canceled {
		t.Errorf("List() error = %v, want context.Canceled", err)
	}

	_, err = store.Get(ctx, "any_flag")
	if err != context.Canceled {
		t.Errorf("Get() error = %v, want context.Canceled", err)
	}
}

func TestInMemoryFeatureFlagStore_ImplementsAdminInterface(t *testing.T) {
	// Compile-time interface check
	var _ AdminFeatureFlagProvider = NewInMemoryFeatureFlagStore()
	var _ FeatureFlagProvider = NewInMemoryFeatureFlagStore()
}

func TestInMemoryFeatureFlagStore_IsEnabledForContext(t *testing.T) {
	store := NewInMemoryFeatureFlagStore(
		WithInitialFlags(map[string]bool{
			"beta_feature": true,
		}),
	)

	evalCtx := EvalContext{UserID: "user-123"}
	enabled, err := store.IsEnabledForContext(context.Background(), "beta_feature", evalCtx)
	if err != nil {
		t.Errorf("IsEnabledForContext() error = %v", err)
	}
	if !enabled {
		t.Error("beta_feature should be enabled")
	}
}

func TestInMemoryFeatureFlagStore_FallsBackToEnv(t *testing.T) {
	t.Setenv("FF_ENV_FLAG", "true")

	store := NewInMemoryFeatureFlagStore()

	// Flag not in store, should fall back to env
	enabled, err := store.IsEnabled(context.Background(), "env_flag")
	if err != nil {
		t.Errorf("IsEnabled() error = %v", err)
	}
	if !enabled {
		t.Error("env_flag should be enabled from environment")
	}
}

func TestInMemoryFeatureFlagStore_Get_FallsBackToEnv(t *testing.T) {
	t.Setenv("FF_ENV_FEATURE", "true")

	store := NewInMemoryFeatureFlagStore()

	// Get() should also fall back to env for consistency
	state, err := store.Get(context.Background(), "env_feature")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if state.Name != "env_feature" {
		t.Errorf("Name = %v, want env_feature", state.Name)
	}
	if !state.Enabled {
		t.Error("Enabled should be true (from environment)")
	}
	if state.Description != "(from environment)" {
		t.Errorf("Description = %v, want '(from environment)'", state.Description)
	}
}
