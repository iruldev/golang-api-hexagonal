package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListParams_Offset(t *testing.T) {
	tests := []struct {
		name     string
		params   ListParams
		expected int
	}{
		{
			name:     "first page returns 0 offset",
			params:   ListParams{Page: 1, PageSize: 20},
			expected: 0,
		},
		{
			name:     "second page with default page size",
			params:   ListParams{Page: 2, PageSize: 20},
			expected: 20,
		},
		{
			name:     "third page with custom page size",
			params:   ListParams{Page: 3, PageSize: 10},
			expected: 20,
		},
		{
			name:     "zero page treated as first page",
			params:   ListParams{Page: 0, PageSize: 20},
			expected: 0,
		},
		{
			name:     "negative page treated as first page",
			params:   ListParams{Page: -1, PageSize: 20},
			expected: 0,
		},
		{
			name:     "zero page size uses default",
			params:   ListParams{Page: 2, PageSize: 0},
			expected: DefaultPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.params.Offset())
		})
	}
}

func TestListParams_Limit(t *testing.T) {
	tests := []struct {
		name     string
		params   ListParams
		expected int
	}{
		{
			name:     "custom page size",
			params:   ListParams{PageSize: 50},
			expected: 50,
		},
		{
			name:     "zero page size uses default",
			params:   ListParams{PageSize: 0},
			expected: DefaultPageSize,
		},
		{
			name:     "negative page size uses default",
			params:   ListParams{PageSize: -10},
			expected: DefaultPageSize,
		},
		{
			name:     "default page size constant is 20",
			params:   ListParams{},
			expected: 20,
		},
		{
			name:     "page size is clamped to max",
			params:   ListParams{PageSize: MaxPageSize + 1},
			expected: MaxPageSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.params.Limit())
		})
	}
}
