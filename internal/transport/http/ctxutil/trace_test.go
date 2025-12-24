package ctxutil

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraceID(t *testing.T) {
	t.Run("GetTraceID returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		id := GetTraceID(ctx)
		assert.Empty(t, id)
	})

	t.Run("SetTraceID sets and GetTraceID retrieves value", func(t *testing.T) {
		ctx := context.Background()
		expected := "test-trace-id"
		ctx = SetTraceID(ctx, expected)
		id := GetTraceID(ctx)
		assert.Equal(t, expected, id)
	})
}

func TestSpanID(t *testing.T) {
	t.Run("GetSpanID returns empty string when not set", func(t *testing.T) {
		ctx := context.Background()
		id := GetSpanID(ctx)
		assert.Empty(t, id)
	})

	t.Run("SetSpanID sets and GetSpanID retrieves value", func(t *testing.T) {
		ctx := context.Background()
		expected := "test-span-id"
		ctx = SetSpanID(ctx, expected)
		id := GetSpanID(ctx)
		assert.Equal(t, expected, id)
	})
	t.Run("EmptySpanID constant", func(t *testing.T) {
		assert.Equal(t, "0000000000000000", EmptySpanID)
		assert.Len(t, EmptySpanID, 16)
	})
}

func TestConstants(t *testing.T) {
	t.Run("EmptyTraceID is correct", func(t *testing.T) {
		assert.Equal(t, "00000000000000000000000000000000", EmptyTraceID)
		assert.Len(t, EmptyTraceID, 32)
	})
}
