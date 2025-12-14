// Package note provides the PostgreSQL repository adapter for Note entities.
// This bridges the sqlc-generated queries to the domain repository interface.
package note

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	notedom "github.com/iruldev/golang-api-hexagonal/internal/domain/note"
)

// Repository implements note.Repository interface using PostgreSQL.
type Repository struct {
	pool    *pgxpool.Pool
	queries *Queries
}

// NewRepository creates a new PostgreSQL note repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: New(pool),
	}
}

// Create persists a new note to the database.
func (r *Repository) Create(ctx context.Context, note *notedom.Note) error {
	// Generate ID if not already set
	if note.ID == uuid.Nil {
		note.ID = uuid.New()
	}

	now := time.Now()
	var pgID pgtype.UUID
	_ = pgID.Scan(note.ID.String())

	params := CreateNoteParams{
		ID:        pgID,
		Title:     note.Title,
		Content:   pgtype.Text{String: note.Content, Valid: note.Content != ""},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	created, err := r.queries.CreateNote(ctx, params)
	if err != nil {
		return fmt.Errorf("create note: %w", err)
	}

	// Update the note with database-generated values
	note.ID = uuidFromPgtype(created.ID)
	note.CreatedAt = created.CreatedAt.Time
	note.UpdatedAt = created.UpdatedAt.Time

	return nil
}

// Get retrieves a note by ID.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*notedom.Note, error) {
	var pgID pgtype.UUID
	_ = pgID.Scan(id.String())

	row, err := r.queries.GetNote(ctx, pgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, notedom.ErrNoteNotFound
		}
		return nil, fmt.Errorf("get note: %w", err)
	}

	return &notedom.Note{
		ID:        uuidFromPgtype(row.ID),
		Title:     row.Title,
		Content:   row.Content.String,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}, nil
}

// List retrieves notes with pagination using sqlc-generated queries (FR21: Type-safe SQL).
func (r *Repository) List(ctx context.Context, limit, offset int) ([]*notedom.Note, int64, error) {
	// Get total count using sqlc-generated CountNotes
	total, err := r.queries.CountNotes(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count notes: %w", err)
	}

	// Get paginated results using sqlc-generated ListNotes
	params := ListNotesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	}

	rows, err := r.queries.ListNotes(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list notes: %w", err)
	}

	notes := make([]*notedom.Note, 0, len(rows))
	for _, row := range rows {
		notes = append(notes, &notedom.Note{
			ID:        uuidFromPgtype(row.ID),
			Title:     row.Title,
			Content:   row.Content.String,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		})
	}

	return notes, total, nil
}

// Update persists changes to an existing note.
func (r *Repository) Update(ctx context.Context, note *notedom.Note) error {
	var pgID pgtype.UUID
	_ = pgID.Scan(note.ID.String())

	now := time.Now()
	params := UpdateNoteParams{
		ID:        pgID,
		Title:     note.Title,
		Content:   pgtype.Text{String: note.Content, Valid: note.Content != ""},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	updated, err := r.queries.UpdateNote(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notedom.ErrNoteNotFound
		}
		return fmt.Errorf("update note: %w", err)
	}

	// Update timestamps from database
	note.UpdatedAt = updated.UpdatedAt.Time

	return nil
}

// Delete removes a note by ID.
// Uses :execrows to check if a row was deleted, avoiding an extra query.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	var pgID pgtype.UUID
	_ = pgID.Scan(id.String())

	// DeleteNote returns the number of rows affected
	rowsAffected, err := r.queries.DeleteNote(ctx, pgID)
	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	if rowsAffected == 0 {
		return notedom.ErrNoteNotFound
	}

	return nil
}

// uuidFromPgtype converts a pgtype.UUID to uuid.UUID.
func uuidFromPgtype(pg pgtype.UUID) uuid.UUID {
	if !pg.Valid {
		return uuid.Nil
	}
	return uuid.UUID(pg.Bytes)
}
