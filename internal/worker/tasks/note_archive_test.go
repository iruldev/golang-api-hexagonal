package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewNoteArchiveTask(t *testing.T) {
	tests := []struct {
		name    string
		noteID  uuid.UUID
		wantErr bool
	}{
		{
			name:    "valid note ID",
			noteID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			wantErr: false,
		},
		{
			name:    "nil note ID",
			noteID:  uuid.Nil,
			wantErr: false, // Task creation succeeds, validation happens in handler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			task, err := NewNoteArchiveTask(tt.noteID)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, task)
			assert.Equal(t, TypeNoteArchive, task.Type())
			assert.NotEmpty(t, task.Payload())
		})
	}
}

func TestNewNoteArchiveHandler(t *testing.T) {
	// Arrange
	logger := zap.NewNop()

	// Act
	handler := NewNoteArchiveHandler(logger)

	// Assert
	require.NotNil(t, handler)
	assert.NotNil(t, handler.logger)
}

func TestNoteArchiveHandler_Handle(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		wantErr   bool
		skipRetry bool
	}{
		{
			name:    "valid payload",
			payload: []byte(`{"note_id":"550e8400-e29b-41d4-a716-446655440000"}`),
			wantErr: false,
		},
		{
			name:      "invalid json",
			payload:   []byte(`{invalid}`),
			wantErr:   true,
			skipRetry: true,
		},
		{
			name:      "nil note_id",
			payload:   []byte(`{"note_id":"00000000-0000-0000-0000-000000000000"}`),
			wantErr:   true,
			skipRetry: true,
		},
		{
			name:      "empty payload",
			payload:   []byte(`{}`),
			wantErr:   true,
			skipRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := zap.NewNop()
			handler := NewNoteArchiveHandler(logger)
			task := asynq.NewTask(TypeNoteArchive, tt.payload)

			// Act
			err := handler.Handle(context.Background(), task)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				if tt.skipRetry {
					assert.True(t, errors.Is(err, asynq.SkipRetry))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTypeNoteArchive(t *testing.T) {
	// Assert task type constant is correctly defined
	assert.Equal(t, "note:archive", TypeNoteArchive)
}
