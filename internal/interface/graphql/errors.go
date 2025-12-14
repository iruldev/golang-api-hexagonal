package graphql

import (
	"context"
	"errors"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/domain/note"
)

// MapError maps domain errors to GraphQL errors with appropriate extensions.
func MapError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	var code string
	var msg = err.Error()

	switch {
	case errors.Is(err, domain.ErrNotFound) || errors.Is(err, note.ErrNoteNotFound):
		code = "NOT_FOUND"
	case errors.Is(err, domain.ErrValidation) || errors.Is(err, note.ErrEmptyTitle) || errors.Is(err, note.ErrTitleTooLong):
		code = "BAD_REQUEST"
	case errors.Is(err, domain.ErrUnauthorized):
		code = "UNAUTHORIZED"
	case errors.Is(err, domain.ErrForbidden):
		code = "FORBIDDEN"
	case errors.Is(err, domain.ErrConflict):
		code = "CONFLICT"
	case errors.Is(err, domain.ErrInternal):
		code = "INTERNAL_SERVER_ERROR"
	default:
		code = "INTERNAL_SERVER_ERROR"
	}

	return &gqlerror.Error{
		Message: msg,
		Path:    graphql.GetPath(ctx),
		Extensions: map[string]interface{}{
			"code": code,
		},
	}
}
