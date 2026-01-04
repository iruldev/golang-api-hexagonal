package redact_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
)

// Unit tests for redaction logic.

func TestNewPIIRedactor(t *testing.T) {
	cfg := domain.RedactorConfig{EmailMode: domain.EmailModeFull}
	r := redact.NewPIIRedactor(cfg)
	assert.NotNil(t, r)
}

func TestNewPIIRedactor_NormalizesConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"mixed case", "Partial", domain.EmailModePartial},
		{"upper case", "FULL", domain.EmailModeFull},
		{"whitespace", "  partial  ", domain.EmailModePartial},
		{"empty defaults to full", "", domain.EmailModeFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: tt.input})
			if tt.expected == domain.EmailModePartial {
				res := r.RedactMap(map[string]any{"email": "test@example.com"})
				assert.Equal(t, "te***@example.com", res["email"])
			} else {
				res := r.RedactMap(map[string]any{"email": "test@example.com"})
				assert.Equal(t, "[REDACTED]", res["email"])
			}
		})
	}
}

func TestPIIRedactor_RedactMap_NilInput(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})
	result := r.RedactMap(nil)
	assert.Nil(t, result)
}

func TestPIIRedactor_RedactMap_EmptyMap(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})
	result := r.RedactMap(map[string]any{})
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestPIIRedactor_Redact(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	t.Run("Map", func(t *testing.T) {
		input := map[string]any{"password": "secret"}
		result := r.Redact(input)
		resMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "[REDACTED]", resMap["password"])
	})

	t.Run("Slice", func(t *testing.T) {
		input := []any{map[string]any{"password": "secret"}}
		result := r.Redact(input)
		resSlice, ok := result.([]any)
		require.True(t, ok)
		require.Len(t, resSlice, 1)
		resMap, ok := resSlice[0].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "[REDACTED]", resMap["password"])
	})

	t.Run("Primitive", func(t *testing.T) {
		input := "safe"
		result := r.Redact(input)
		assert.Equal(t, "safe", result)
	})
}

func TestPIIRedactor_RedactMap_StandardPIIFields(t *testing.T) {
	tests := []struct {
		name  string
		field string
	}{
		{"password", "password"},
		{"token", "token"},
		{"secret", "secret"},
		{"authorization", "authorization"},
		{"creditCard", "creditCard"}, // Explicit test case as requested
		{"credit_card", "credit_card"},
		{"ssn", "ssn"},
	}

	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				tt.field: "sensitive-value",
			}
			result := r.RedactMap(input)
			assert.Equal(t, "[REDACTED]", result[tt.field])
		})
	}
}

func TestPIIRedactor_RedactMap_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		field string
	}{
		{"lowercase password", "password"},
		{"uppercase PASSWORD", "PASSWORD"},
		{"mixed case Password", "Password"},
		{"mixed case PaSsWoRd", "PaSsWoRd"},
		{"lowercase email", "email"},
		{"uppercase EMAIL", "EMAIL"},
		{"mixed case Email", "Email"},
	}

	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				tt.field: "sensitive-value",
			}
			result := r.RedactMap(input)
			assert.Equal(t, "[REDACTED]", result[tt.field])
		})
	}
}

func TestPIIRedactor_RedactMap_EmailFullMode(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"email": "john.doe@example.com",
	}
	result := r.RedactMap(input)
	assert.Equal(t, "[REDACTED]", result["email"])
}

func TestPIIRedactor_RedactMap_EmailPartialMode(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"normal email", "john.doe@example.com", "jo***@example.com"},
		{"short local part", "a@example.com", "***@example.com"},
		{"two char local part", "ab@example.com", "ab***@example.com"},
		{"single char local", "j@x.com", "***@x.com"},
	}

	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModePartial})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				"email": tt.email,
			}
			result := r.RedactMap(input)
			assert.Equal(t, tt.expected, result["email"])
		})
	}
}

func TestPIIRedactor_RedactMap_EmailNonStringValue(t *testing.T) {
	// If email field has non-string value, it should be fully redacted
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModePartial})

	tests := []struct {
		name  string
		value any
	}{
		{"integer", 12345},
		{"boolean", true},
		{"null", nil},
		{"slice", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				"email": tt.value,
			}
			result := r.RedactMap(input)
			assert.Equal(t, "[REDACTED]", result["email"])
		})
	}
}

// TestRedactedValue verifies the constant is exported correctly.
func TestRedactedValue(t *testing.T) {
	assert.Equal(t, "[REDACTED]", redact.RedactedValue)
}

func TestPIIRedactor_APIKeyRedaction(t *testing.T) {
	tests := []struct {
		name  string
		field string
	}{
		{"apikey", "apikey"},
		{"api_key", "api_key"},
		{"apiKey", "apiKey"},
		{"API_KEY", "API_KEY"},
	}

	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := map[string]any{
				tt.field: "sensitive-value",
			}
			result := r.RedactMap(input)
			assert.Equal(t, "[REDACTED]", result[tt.field])
		})
	}
}
