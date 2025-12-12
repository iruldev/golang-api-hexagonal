package note

import "errors"

// Domain-specific errors for the Note entity.
var (
	// ErrEmptyTitle indicates the note title is empty.
	ErrEmptyTitle = errors.New("note: title cannot be empty")

	// ErrTitleTooLong indicates the note title exceeds maximum length.
	ErrTitleTooLong = errors.New("note: title exceeds maximum length")

	// ErrNoteNotFound indicates the requested note was not found.
	ErrNoteNotFound = errors.New("note: not found")
)
