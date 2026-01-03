package redact_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/shared/redact"
)

// RedactAndMarshal, recursion limit, struct handling, and performance tests.

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
