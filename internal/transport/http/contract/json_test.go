package contract_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/contract"
)

func TestDecodeJSONStrict_ValidJSON(t *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	input := `{"name": "John", "email": "john@example.com"}`
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.NoError(t, err)
	assert.Equal(t, "John", dst.Name)
	assert.Equal(t, "john@example.com", dst.Email)
}

func TestDecodeJSONStrict_RejectsUnknownField(t *testing.T) {
	type payload struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	// "usernmae" is a typo - should be rejected
	input := `{"name": "John", "email": "john@example.com", "usernmae": "typo"}`
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindUnknownField, jsonErr.Kind)
	assert.Equal(t, "usernmae", jsonErr.Field)
	assert.Contains(t, jsonErr.Message, "unknown field")
}

func TestDecodeJSONStrict_RejectsMultipleUnknownFields(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	// Multiple unknown fields - should reject on first
	input := `{"name": "John", "unknownA": "val", "unknownB": "val"}`
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindUnknownField, jsonErr.Kind)
	// First unknown field encountered
	assert.Equal(t, "unknownA", jsonErr.Field)
}

func TestDecodeJSONStrict_InvalidSyntax(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	input := `{"name": "John"` // Missing closing brace
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindEOF, jsonErr.Kind)
}

func TestDecodeJSONStrict_TypeMismatch(t *testing.T) {
	type payload struct {
		Age int `json:"age"`
	}

	input := `{"age": "not-a-number"}` // String instead of int
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindTypeMismatch, jsonErr.Kind)
	assert.Equal(t, "age", jsonErr.Field)
}

func TestDecodeJSONStrict_EmptyBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	input := `` // Empty
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindEOF, jsonErr.Kind)
}

func TestDecodeJSONStrict_RejectsTrailingText(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	// Valid JSON followed by trailing text
	input := `{"name": "John"}extra`
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindTrailingData, jsonErr.Kind)
	assert.Contains(t, jsonErr.Message, "trailing data")
}

func TestDecodeJSONStrict_RejectsTrailingJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	// Valid JSON followed by another JSON object
	input := `{"name": "John"}{"other": "json"}`
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.Error(t, err)

	jsonErr, ok := err.(*contract.JSONDecodeError)
	require.True(t, ok, "error should be *JSONDecodeError")
	assert.Equal(t, contract.JSONDecodeErrorKindTrailingData, jsonErr.Kind)
}

func TestDecodeJSONStrict_AllowsTrailingWhitespace(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}

	// Valid JSON with trailing whitespace (should be OK)
	input := `{"name": "John"}   `
	var dst payload

	err := contract.DecodeJSONStrict(strings.NewReader(input), &dst)

	require.NoError(t, err, "trailing whitespace should be allowed")
	assert.Equal(t, "John", dst.Name)
}
