package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

const (
	// TypeCleanupOldNotes is the task type for cleaning up old archived notes.
	TypeCleanupOldNotes = "cleanup:old_notes"
)

// CleanupOldNotesPayload is the typed payload for cleanup tasks.
// This task is typically scheduled to run periodically (e.g., daily).
type CleanupOldNotesPayload struct {
	// ArchivedBefore specifies the cutoff time - notes archived before this will be cleaned up.
	// If empty, defaults to 30 days ago.
	ArchivedBefore time.Time `json:"archived_before,omitempty"`

	// DryRun if true, only logs what would be deleted without actually deleting.
	DryRun bool `json:"dry_run,omitempty"`
}

// NewCleanupOldNotesTask creates a new cleanup task with default options.
// By default, cleans up notes archived more than 30 days ago.
func NewCleanupOldNotesTask() (*asynq.Task, error) {
	return NewCleanupOldNotesTaskWithOptions(CleanupOldNotesPayload{
		ArchivedBefore: time.Now().AddDate(0, 0, -30), // 30 days ago
		DryRun:         false,
	})
}

// NewCleanupOldNotesTaskWithOptions creates a cleanup task with custom options.
func NewCleanupOldNotesTaskWithOptions(opts CleanupOldNotesPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("marshal cleanup payload: %w", err)
	}
	return asynq.NewTask(TypeCleanupOldNotes, payload,
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute), // Give cleanup adequate time
	), nil
}

// CleanupOldNotesHandler handles cleanup tasks.
type CleanupOldNotesHandler struct {
	logger *zap.Logger
	// Future: inject repository for actual database operations
	// noteRepo note.Repository
}

// NewCleanupOldNotesHandler creates a handler with injected dependencies.
func NewCleanupOldNotesHandler(logger *zap.Logger) *CleanupOldNotesHandler {
	return &CleanupOldNotesHandler{logger: logger}
}

// Handle processes cleanup tasks.
func (h *CleanupOldNotesHandler) Handle(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)

	// 1. Unmarshal payload
	var p CleanupOldNotesPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error("invalid cleanup payload",
			zap.Error(err),
			zap.String("task_type", TypeCleanupOldNotes),
			zap.String("task_id", taskID),
		)
		return fmt.Errorf("unmarshal payload: %v: %w", err, asynq.SkipRetry)
	}

	// 2. Set defaults if needed
	if p.ArchivedBefore.IsZero() {
		p.ArchivedBefore = time.Now().AddDate(0, 0, -30)
	}

	h.logger.Info("starting cleanup of old notes",
		zap.String("task_type", TypeCleanupOldNotes),
		zap.String("task_id", taskID),
		zap.Time("archived_before", p.ArchivedBefore),
		zap.Bool("dry_run", p.DryRun),
	)

	// 3. Perform cleanup (placeholder - actual implementation would query and delete)
	// In a real implementation:
	// - Query notes where archived_at < p.ArchivedBefore
	// - Delete them (or mark for deletion)
	// - Log count of deleted notes

	// Simulated work - in production, this would be actual database operations
	cleanedCount := 0 // Would be set by actual cleanup logic

	h.logger.Info("completed cleanup of old notes",
		zap.String("task_type", TypeCleanupOldNotes),
		zap.String("task_id", taskID),
		zap.Int("cleaned_count", cleanedCount),
		zap.Bool("dry_run", p.DryRun),
	)

	return nil
}

// CleanupOldNotesJobID generates a unique job ID for deduplication.
// Use this with asynq.TaskID option to prevent duplicate cleanup jobs.
func CleanupOldNotesJobID() string {
	// Use date-based ID to allow one cleanup per day
	return fmt.Sprintf("cleanup:old_notes:%s", time.Now().UTC().Format("2006-01-02"))
}
