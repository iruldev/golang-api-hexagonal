package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// NoteArchivePayload is the typed payload for note archive tasks.
type NoteArchivePayload struct {
	NoteID uuid.UUID `json:"note_id"`
}

// NewNoteArchiveTask creates a new note archive task with default options.
func NewNoteArchiveTask(noteID uuid.UUID) (*asynq.Task, error) {
	payload, err := json.Marshal(NoteArchivePayload{NoteID: noteID})
	if err != nil {
		return nil, fmt.Errorf("marshal note archive payload: %w", err)
	}
	return asynq.NewTask(TypeNoteArchive, payload, asynq.MaxRetry(3)), nil
}

// NoteArchiveHandler handles note archive tasks.
// Use NewNoteArchiveHandler to create with injected dependencies.
type NoteArchiveHandler struct {
	logger *zap.Logger
}

// NewNoteArchiveHandler creates a handler with injected logger.
func NewNoteArchiveHandler(logger *zap.Logger) *NoteArchiveHandler {
	return &NoteArchiveHandler{logger: logger}
}

// Handle processes note archive tasks.
func (h *NoteArchiveHandler) Handle(ctx context.Context, t *asynq.Task) error {
	taskID, _ := asynq.GetTaskID(ctx)

	var p NoteArchivePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		h.logger.Error("invalid payload",
			zap.Error(err),
			zap.String("task_type", TypeNoteArchive),
			zap.String("task_id", taskID),
		)
		return fmt.Errorf("unmarshal payload: %v: %w", err, asynq.SkipRetry)
	}

	if p.NoteID == uuid.Nil {
		h.logger.Error("missing note_id",
			zap.String("task_type", TypeNoteArchive),
			zap.String("task_id", taskID),
		)
		return fmt.Errorf("note_id is required: %w", asynq.SkipRetry)
	}

	h.logger.Info("archiving note",
		zap.String("task_type", TypeNoteArchive),
		zap.String("task_id", taskID),
		zap.String("note_id", p.NoteID.String()),
	)

	// TODO: Implement actual archive logic when business requirement exists
	// For sample, just log success
	h.logger.Info("note archived successfully",
		zap.String("task_type", TypeNoteArchive),
		zap.String("task_id", taskID),
		zap.String("note_id", p.NoteID.String()),
	)

	return nil
}
