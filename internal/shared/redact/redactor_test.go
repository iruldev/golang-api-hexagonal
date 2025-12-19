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

func TestRedactAndMarshal_SliceInput(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	input := []any{
		map[string]any{"password": "secret"},
		map[string]any{"name": "test"},
	}

	result, err := redact.RedactAndMarshal(r, input)
	require.NoError(t, err)

	var output []any
	err = json.Unmarshal(result, &output)
	require.NoError(t, err)
	require.Len(t, output, 2)

	item0 := output[0].(map[string]any)
	assert.Equal(t, "[REDACTED]", item0["password"])
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

	// Verify we didn't panic
	assert.NotNil(t, result)

	// Traverse result to verify depth limit behavior
	// The implementation now returns a marker map at max depth (fail-safe)
	depth := 0
	curr := result
	for {
		next, ok := curr["next"].(map[string]any)
		if !ok {
			// Check if we hit the marker
			if val, exists := curr["_REDACTED_"]; exists {
				assert.Equal(t, "Max Recursion Depth Exceeded", val)
				break
			}
			// If no marker and no next, just break (end of chain)
			break
		}
		curr = next
		depth++
	}

	// Should have traversed up to near MaxRecursionDepth levels before hitting empty map
	assert.True(t, depth >= 98, "Should handle at least 98 levels before stopping")
}

func TestPIIRedactor_RecursionLimit_PIINotLeaked(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	// Create map with PII at depth 101 (beyond MaxRecursionDepth)
	deepMap := make(map[string]any)
	current := deepMap
	for i := 0; i < 102; i++ {
		next := make(map[string]any)
		if i == 101 {
			// This password is at depth 102, beyond MaxRecursionDepth
			next["password"] = "secret-that-must-not-leak"
		}
		current["next"] = next
		current = next
	}

	result := r.RedactMap(deepMap)

	// Traverse to the deepest point we can reach
	curr := result
	foundPassword := false
	for i := 0; i < 120; i++ {
		// Check if password leaked at this level
		if pwd, exists := curr["password"]; exists {
			if pwd == "secret-that-must-not-leak" {
				foundPassword = true
				break
			}
		}

		next, ok := curr["next"].(map[string]any)
		if !ok {
			break
		}
		curr = next
	}

	// Password should NOT have leaked
	assert.False(t, foundPassword, "Password should NOT leak beyond max recursion depth")
}

func TestRedactAndMarshal_UnmarshalableStruct(t *testing.T) {
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	// Channels cannot be marshaled to JSON
	type BadStruct struct {
		Ch chan int `json:"ch"`
	}

	input := BadStruct{Ch: make(chan int)}

	// Now returns fail-safe nil (marshaled to "null") instead of erroring or returning potentially leaked data
	result, err := redact.RedactAndMarshal(r, input)
	assert.NoError(t, err)
	assert.Equal(t, []byte("null"), result)
}

func TestPIIRedactor_Redact_Struct(t *testing.T) {
	// New test to verify direct struct redaction capability
	r := redact.NewPIIRedactor(domain.RedactorConfig{EmailMode: domain.EmailModeFull})

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		APIKey   string `json:"api_key"`
	}

	input := User{
		Name:     "John",
		Password: "secret",
		APIKey:   "123-abc-key",
	}

	// Redact should now treat struct as map (via JSON conversion) and redact it
	result := r.Redact(input)

	resMap, ok := result.(map[string]any)
	require.True(t, ok, "Struct should be converted to map")

	assert.Equal(t, "John", resMap["name"])
	assert.Equal(t, "[REDACTED]", resMap["password"])
	assert.Equal(t, "[REDACTED]", resMap["api_key"])
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
