package tasks

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCleanupOldNotesTask(t *testing.T) {
	// Act
	task, err := NewCleanupOldNotesTask()

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, TypeCleanupOldNotes, task.Type())

	// Verify payload is valid JSON
	var payload CleanupOldNotesPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)

	// Should default to ~30 days ago (with some tolerance for test timing)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	assert.WithinDuration(t, thirtyDaysAgo, payload.ArchivedBefore, time.Minute)
	assert.False(t, payload.DryRun)
}

func TestNewCleanupOldNotesTaskWithOptions(t *testing.T) {
	tests := []struct {
		name           string
		opts           CleanupOldNotesPayload
		wantDryRun     bool
		wantTimeOffset time.Duration
	}{
		{
			name: "custom time and dry run",
			opts: CleanupOldNotesPayload{
				ArchivedBefore: time.Now().AddDate(0, 0, -7), // 7 days ago
				DryRun:         true,
			},
			wantDryRun:     true,
			wantTimeOffset: 7 * 24 * time.Hour,
		},
		{
			name: "90 days cleanup",
			opts: CleanupOldNotesPayload{
				ArchivedBefore: time.Now().AddDate(0, 0, -90),
				DryRun:         false,
			},
			wantDryRun:     false,
			wantTimeOffset: 90 * 24 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			task, err := NewCleanupOldNotesTaskWithOptions(tt.opts)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, TypeCleanupOldNotes, task.Type())

			var payload CleanupOldNotesPayload
			err = json.Unmarshal(task.Payload(), &payload)
			require.NoError(t, err)

			assert.Equal(t, tt.wantDryRun, payload.DryRun)
		})
	}
}

func TestCleanupOldNotesHandler_Handle(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		wantErr   bool
		skipRetry bool
	}{
		{
			name:    "valid payload with defaults",
			payload: []byte(`{}`),
			wantErr: false,
		},
		{
			name:    "valid payload with custom time",
			payload: []byte(`{"archived_before":"2025-01-01T00:00:00Z","dry_run":false}`),
			wantErr: false,
		},
		{
			name:    "valid payload dry run",
			payload: []byte(`{"dry_run":true}`),
			wantErr: false,
		},
		{
			name:      "invalid json",
			payload:   []byte(`{invalid}`),
			wantErr:   true,
			skipRetry: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			logger := zap.NewNop()
			h := NewCleanupOldNotesHandler(logger)
			task := asynq.NewTask(TypeCleanupOldNotes, tt.payload)

			// Act
			err := h.Handle(context.Background(), task)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.skipRetry {
					require.ErrorIs(t, err, asynq.SkipRetry)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCleanupOldNotesJobID(t *testing.T) {
	// Act
	id1 := CleanupOldNotesJobID()
	id2 := CleanupOldNotesJobID()

	// Assert - same day should produce same ID (for deduplication)
	assert.Equal(t, id1, id2)
	assert.Contains(t, id1, "cleanup:old_notes:")
	assert.Contains(t, id1, time.Now().UTC().Format("2006-01-02"))
}

func TestTypeCleanupOldNotes_Constant(t *testing.T) {
	// Assert - verify the task type follows naming convention
	assert.Equal(t, "cleanup:old_notes", TypeCleanupOldNotes)
}
