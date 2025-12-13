package runtimeutil

import (
	"context"
	"os"
	"testing"
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
