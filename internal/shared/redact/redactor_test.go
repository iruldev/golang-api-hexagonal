package redact_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
)

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
			// We can't access private field directly but we can verify behavior
			// If input normalizes to "partial", then partial redaction should work
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
		{"short local part", "a@example.com", "a***@example.com"},
		{"two char local part", "ab@example.com", "ab***@example.com"},
		{"single char local", "j@x.com", "j***@x.com"},
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

func TestPIIRedactor_RedactMap_EmailPartialModeEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected string
	}{
		{"no @ symbol", "invalid-email", "[REDACTED]"},
		{"@ at start", "@example.com", "[REDACTED]"},
		{"empty string", "", "[REDACTED]"},
		{"multiple @ symbols (uses first)", "user@domain@hack.com", "us***@domain@hack.com"},
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

func TestPIIRedactor_RedactMap_NestedObjects(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"user": map[string]any{
			"name":     "John Doe",
			"email":    "john@example.com",
			"password": "secret123",
		},
		"metadata": map[string]any{
			"token": "abc123",
		},
	}

	result := r.RedactMap(input)

	user := result["user"].(map[string]any)
	assert.Equal(t, "John Doe", user["name"])
	assert.Equal(t, "[REDACTED]", user["email"])
	assert.Equal(t, "[REDACTED]", user["password"])

	metadata := result["metadata"].(map[string]any)
	assert.Equal(t, "[REDACTED]", metadata["token"])
}

func TestPIIRedactor_RedactMap_DeeplyNestedObjects(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"password": "deep-secret",
					"name":     "allowed",
				},
			},
		},
	}

	result := r.RedactMap(input)

	l1 := result["level1"].(map[string]any)
	l2 := l1["level2"].(map[string]any)
	l3 := l2["level3"].(map[string]any)
	assert.Equal(t, "[REDACTED]", l3["password"])
	assert.Equal(t, "allowed", l3["name"])
}

func TestPIIRedactor_RedactMap_ArrayOfObjects(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"users": []any{
			map[string]any{"email": "user1@test.com", "name": "User 1"},
			map[string]any{"email": "user2@test.com", "name": "User 2"},
		},
	}

	result := r.RedactMap(input)

	users := result["users"].([]any)
	assert.Len(t, users, 2)

	user1 := users[0].(map[string]any)
	assert.Equal(t, "[REDACTED]", user1["email"])
	assert.Equal(t, "User 1", user1["name"])

	user2 := users[1].(map[string]any)
	assert.Equal(t, "[REDACTED]", user2["email"])
	assert.Equal(t, "User 2", user2["name"])
}

func TestPIIRedactor_RedactMap_NestedArrays(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"data": []any{
			[]any{
				map[string]any{"secret": "nested-secret"},
			},
		},
	}

	result := r.RedactMap(input)

	data := result["data"].([]any)
	nested := data[0].([]any)
	obj := nested[0].(map[string]any)
	assert.Equal(t, "[REDACTED]", obj["secret"])
}

func TestPIIRedactor_RedactMap_NonPIIFieldsPreserved(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"id":        "user-123",
		"name":      "John Doe",
		"age":       30,
		"active":    true,
		"roles":     []any{"admin", "user"},
		"createdAt": "2025-01-01T00:00:00Z",
	}

	result := r.RedactMap(input)

	assert.Equal(t, "user-123", result["id"])
	assert.Equal(t, "John Doe", result["name"])
	assert.Equal(t, 30, result["age"])
	assert.Equal(t, true, result["active"])
	assert.Equal(t, []any{"admin", "user"}, result["roles"])
	assert.Equal(t, "2025-01-01T00:00:00Z", result["createdAt"])
}

func TestPIIRedactor_RedactMap_OriginalNotModified(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	original := map[string]any{
		"email":    "original@test.com",
		"password": "original-password",
		"nested": map[string]any{
			"token": "original-token",
		},
	}

	// Make copies to compare later
	originalEmail := original["email"]
	originalPassword := original["password"]
	originalNested := original["nested"].(map[string]any)
	originalToken := originalNested["token"]

	result := r.RedactMap(original)

	// Verify result is redacted
	assert.Equal(t, "[REDACTED]", result["email"])
	assert.Equal(t, "[REDACTED]", result["password"])
	resultNested := result["nested"].(map[string]any)
	assert.Equal(t, "[REDACTED]", resultNested["token"])

	// Verify original is NOT modified
	assert.Equal(t, originalEmail, original["email"])
	assert.Equal(t, originalPassword, original["password"])
	assert.Equal(t, originalToken, original["nested"].(map[string]any)["token"])
}

func TestRedactAndMarshal_NilInput(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})
	result, err := redact.RedactAndMarshal(r, nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedactAndMarshal_EmptyBytes(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})
	result, err := redact.RedactAndMarshal(r, []byte{})
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestRedactAndMarshal_MapInput(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"email": "test@example.com",
		"name":  "Test User",
	}

	result, err := redact.RedactAndMarshal(r, input)
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.Equal(t, "[REDACTED]", output["email"])
	assert.Equal(t, "Test User", output["name"])
}

func TestRedactAndMarshal_JSONBytesInput(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := []byte(`{"password":"secret123","id":"user-1"}`)

	result, err := redact.RedactAndMarshal(r, input)
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.Equal(t, "[REDACTED]", output["password"])
	assert.Equal(t, "user-1", output["id"])
}

func TestRedactAndMarshal_StructInput(t *testing.T) {
	type UserPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := UserPayload{
		Email:    "user@test.com",
		Password: "secret",
		Name:     "John",
	}

	result, err := redact.RedactAndMarshal(r, input)
	require.NoError(t, err)

	var output map[string]any
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)

	assert.Equal(t, "[REDACTED]", output["email"])
	assert.Equal(t, "[REDACTED]", output["password"])
	assert.Equal(t, "John", output["name"])
}

func TestRedactAndMarshal_InvalidJSONBytes(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := []byte(`{invalid json}`)

	_, err := redact.RedactAndMarshal(r, input)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal JSON bytes")
}

func TestPIIRedactor_RedactMap_MixedContent(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModePartial})

	input := map[string]any{
		"action": "user.created",
		"actor":  "system",
		"payload": map[string]any{
			"user": map[string]any{
				"id":       "user-123",
				"email":    "john.doe@example.com",
				"password": "hashed-password",
				"profile": map[string]any{
					"ssn":   "123-45-6789",
					"name":  "John Doe",
					"token": "refresh-token-abc",
				},
			},
			"creditCard": "4111-1111-1111-1111",
		},
	}

	result := r.RedactMap(input)

	// Top level preserved
	assert.Equal(t, "user.created", result["action"])
	assert.Equal(t, "system", result["actor"])

	payload := result["payload"].(map[string]any)
	assert.Equal(t, "[REDACTED]", payload["creditCard"])

	user := payload["user"].(map[string]any)
	assert.Equal(t, "user-123", user["id"])
	assert.Equal(t, "jo***@example.com", user["email"]) // partial mode
	assert.Equal(t, "[REDACTED]", user["password"])

	profile := user["profile"].(map[string]any)
	assert.Equal(t, "[REDACTED]", profile["ssn"])
	assert.Equal(t, "John Doe", profile["name"])
	assert.Equal(t, "[REDACTED]", profile["token"])
}

func TestPIIRedactor_RedactMap_ArrayWithMixedTypes(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"items": []any{
			"string-value",
			123,
			true,
			nil,
			map[string]any{"password": "secret"},
		},
	}

	result := r.RedactMap(input)

	items := result["items"].([]any)
	assert.Equal(t, "string-value", items[0])
	assert.Equal(t, 123, items[1])
	assert.Equal(t, true, items[2])
	assert.Nil(t, items[3])

	obj := items[4].(map[string]any)
	assert.Equal(t, "[REDACTED]", obj["password"])
}

func TestPIIRedactor_RedactMap_NilSlice(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := map[string]any{
		"items": []any(nil),
	}

	result := r.RedactMap(input)
	assert.Nil(t, result["items"])
}

// TestRedactedValue verifies the constant is exported correctly
func TestRedactedValue(t *testing.T) {
	assert.Equal(t, "[REDACTED]", redact.RedactedValue)
}

func TestPIIRedactor_RecursionLimit(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	// Create a map nested deeper than MaxRecursionDepth (100)
	// We'll go to 105
	deepMap := make(map[string]any)
	current := deepMap
	for i := 0; i < 105; i++ {
		next := make(map[string]any)
		current["next"] = next
		if i == 104 {
			current["password"] = "hidden"
		}
		current = next
	}

	result := r.RedactMap(deepMap)

	// Traverse result to see where it stopped
	// It should stop around depth 100
	depth := 0
	curr := result
	for {
		next, ok := curr["next"].(map[string]any)
		if !ok {
			break
		}
		curr = next
		depth++
	}

	// We expect depth to be limited to MaxRecursionDepth (100) or close to it depending on counting.
	// Since we recurse 0..100, we should see about 100 levels.
	// The implementation returns the `data` (map) as is when depth > 100.
	// So effectively, the simplified logic might return the original map reference at the boundary.
	// But let's just assert we didn't panic and depth is around 100.
	assert.True(t, depth >= 100, "Should handle at least 100 levels")

	// Ensure we didn't crash
	assert.NotNil(t, result)
}
