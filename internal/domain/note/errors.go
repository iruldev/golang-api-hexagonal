package note

import (
	"errors"

	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// Domain-specific errors for the Note entity using central error codes.
// Use these for new code instead of the deprecated sentinel errors below.
var (
	// NewErrEmptyTitle creates a VALIDATION_ERROR for empty title.
	NewErrEmptyTitle = func() *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeValidationError, "note: title cannot be empty")
	}

	// NewErrTitleTooLong creates a VALIDATION_ERROR for title exceeding max length.
	NewErrTitleTooLong = func() *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeValidationError, "note: title exceeds maximum length")
	}

	// NewErrNoteNotFound creates a NOT_FOUND error for note.
	NewErrNoteNotFound = func() *domainerrors.DomainError {
		return domainerrors.NewDomain(domainerrors.CodeNotFound, "note: not found")
	}
)

// Deprecated: Use NewErr* functions instead.
// These sentinel errors are kept for backward compatibility.
// They will be removed in a future version.
var (
	// Deprecated: Use NewErrEmptyTitle() instead.
	// ErrEmptyTitle indicates the note title is empty.
	ErrEmptyTitle = errors.New("note: title cannot be empty")

	// Deprecated: Use NewErrTitleTooLong() instead.
	// ErrTitleTooLong indicates the note title exceeds maximum length.
	ErrTitleTooLong = errors.New("note: title exceeds maximum length")

	// Deprecated: Use NewErrNoteNotFound() instead.
	// ErrNoteNotFound indicates the requested note was not found.
	ErrNoteNotFound = errors.New("note: not found")
)
