package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestID_String(t *testing.T) {
	tests := []struct {
		name     string
		id       ID
		expected string
	}{
		{
			name:     "valid ID returns string",
			id:       ID("abc-123"),
			expected: "abc-123",
		},
		{
			name:     "empty ID returns empty string",
			id:       ID(""),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestID_IsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		id       ID
		expected bool
	}{
		{
			name:     "empty ID returns true",
			id:       ID(""),
			expected: true,
		},
		{
			name:     "non-empty ID returns false",
			id:       ID("abc-123"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.id.IsEmpty()
			assert.Equal(t, tt.expected, result)
		})
	}
}
