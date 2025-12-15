package response

import (
	"context"
	"errors"
	"net/http"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// MapError maps a domain error to HTTP status code and error code.
// First checks for DomainError type (preferred), then falls back to
// legacy sentinel errors for backward compatibility.
//
// Returns:
//   - status: HTTP status code
//   - code: Error code string in UPPER_SNAKE format (e.g., NOT_FOUND)
func MapError(err error) (status int, code string) {
	// First, check if it's a DomainError (preferred path)
	if domainErr := domainerrors.IsDomainError(err); domainErr != nil {
		return mapDomainErrorCode(domainErr.Code), domainErr.Code
	}

	// Fall back to legacy sentinel errors for backward compatibility
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, CodeNotFound
	case errors.Is(err, domain.ErrValidation):
		return http.StatusUnprocessableEntity, domainerrors.CodeValidationError
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, CodeUnauthorized
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, CodeForbidden
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, CodeConflict
	case errors.Is(err, domain.ErrInternal):
		return http.StatusInternalServerError, domainerrors.CodeInternalError
	case errors.Is(err, domain.ErrTimeout):
		return http.StatusGatewayTimeout, CodeTimeout
	default:
		// Unknown errors are treated as internal server errors
		return http.StatusInternalServerError, domainerrors.CodeInternalError
	}
}

// mapDomainErrorCode maps a domain error code to HTTP status.
func mapDomainErrorCode(code string) int {
	switch code {
	case domainerrors.CodeNotFound:
		return http.StatusNotFound
	case domainerrors.CodeValidationError:
		return http.StatusUnprocessableEntity
	case domainerrors.CodeUnauthorized:
		return http.StatusUnauthorized
	case domainerrors.CodeForbidden:
		return http.StatusForbidden
	case domainerrors.CodeConflict:
		return http.StatusConflict
	case domainerrors.CodeInternalError:
		return http.StatusInternalServerError
	case domainerrors.CodeTimeout:
		return http.StatusGatewayTimeout
	case domainerrors.CodeRateLimitExceeded:
		return http.StatusTooManyRequests
	case domainerrors.CodeBadRequest:
		return http.StatusBadRequest
	case domainerrors.CodeTokenExpired:
		return http.StatusUnauthorized
	case domainerrors.CodeTokenInvalid:
		return http.StatusUnauthorized
	default:
		// CRITICAL: Unmapped domain error code encountered.
		// This indicates a developer forgot to add a mapping for a new error code.
		return http.StatusInternalServerError
	}
}

// MapErrorWithHint maps a domain error and extracts hint if available.
// Returns status, code, message, and optional hint.
func MapErrorWithHint(err error) (status int, code, message, hint string) {
	if domainErr := domainerrors.IsDomainError(err); domainErr != nil {
		return mapDomainErrorCode(domainErr.Code), domainErr.Code, domainErr.Message, domainErr.Hint
	}

	// Fall back for non-DomainError
	status, code = MapError(err)
	return status, code, err.Error(), ""
}

// HandleError writes an error response based on domain error type.
// Automatically maps the domain error to HTTP status and error code.
// Uses new Envelope format from Story 2.1.
//
// Example:
//
//	user, err := service.GetUser(id)
//	if err != nil {
//	    response.HandleError(w, err)
//	    return
//	}
//	response.Success(w, user)
func HandleError(w http.ResponseWriter, err error) {
	status, code := MapError(err)
	Error(w, status, code, err.Error())
}

// HandleErrorCtx writes an error response with context for trace_id.
// Uses new Envelope format with meta.trace_id.
func HandleErrorCtx(w http.ResponseWriter, ctx context.Context, err error) {
	if domainErr := domainerrors.IsDomainError(err); domainErr != nil {
		status := mapDomainErrorCode(domainErr.Code)
		if domainErr.Hint != "" {
			ErrorEnvelopeWithHint(w, ctx, status, domainErr.Code, domainErr.Message, domainErr.Hint)
		} else {
			ErrorEnvelope(w, ctx, status, domainErr.Code, domainErr.Message)
		}
		return
	}

	// Fall back for non-DomainError
	status, code := MapError(err)
	ErrorEnvelope(w, ctx, status, code, err.Error())
}
