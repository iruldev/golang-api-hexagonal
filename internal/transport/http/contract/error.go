// Package contract provides HTTP transport layer contracts including
// RFC 7807 Problem Details for machine-readable error responses.
package contract

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// ProblemBaseURL is the default base URL for problem type URIs.
const ProblemBaseURL = "https://configured-url-missing.com/problems/"

var problemBaseURL atomic.Value // string

func init() {
	problemBaseURL.Store(ProblemBaseURL)
}

// CodeServiceUnavailable is kept for backward compatibility if needed,
// though preferred usage is contract.CodeSysUnavailable.
const CodeServiceUnavailable = CodeSysUnavailable

func SetProblemBaseURL(baseURL string) error {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return fmt.Errorf("problem base URL is empty")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("problem base URL must be an absolute URL (scheme + host)")
	}
	if !strings.HasSuffix(trimmed, "/") {
		trimmed += "/"
	}
	problemBaseURL.Store(trimmed)
	return nil
}

// problemTypeURL returns the RFC 7807 type URL.
func problemTypeURL(slug string) string {
	baseURL, ok := problemBaseURL.Load().(string)
	if !ok || baseURL == "" {
		baseURL = ProblemBaseURL
	}
	return baseURL + slug
}

// ProblemTypeURL returns the RFC 7807 type URL for the given slug.
// Exported for use by middleware packages that need to construct problem details.
func ProblemTypeURL(slug string) string {
	return problemTypeURL(slug)
}

func validationErrorsFromAppError(appErr *app.AppError) []ValidationError {
	if appErr == nil || appErr.Err == nil {
		return nil
	}

	type fieldMessageError interface {
		Field() string
		Message() string
	}

	var fm fieldMessageError
	if errors.As(appErr.Err, &fm) {
		field := strings.TrimSpace(fm.Field())
		message := strings.TrimSpace(fm.Message())
		if field == "" {
			field = "validation"
		}
		if message == "" {
			message = safeValidationMessage(appErr)
		}
		return []ValidationError{{Field: field, Message: message}}
	}

	switch {
	case errors.Is(appErr.Err, domain.ErrInvalidEmail):
		return []ValidationError{{Field: "email", Message: "must be a valid email address"}}
	case errors.Is(appErr.Err, domain.ErrInvalidFirstName):
		return []ValidationError{{Field: "firstName", Message: "must not be empty"}}
	case errors.Is(appErr.Err, domain.ErrInvalidLastName):
		return []ValidationError{{Field: "lastName", Message: "must not be empty"}}
	default:
		return []ValidationError{{Field: "validation", Message: safeValidationMessage(appErr)}}
	}
}

func safeValidationMessage(appErr *app.AppError) string {
	if appErr == nil {
		return "Validation failed"
	}
	message := strings.TrimSpace(appErr.Message)
	if message == "" {
		return "Validation failed"
	}
	return message
}

// WriteProblemJSON writes an RFC 7807 error response.
func WriteProblemJSON(w http.ResponseWriter, r *http.Request, err error) {
	// Priority 1: Check for DomainError (Story 3.2)
	var domainErr *domainerrors.DomainError
	if errors.As(err, &domainErr) {
		problem := FromDomainError(r, domainErr)
		WriteProblem(w, problem)
		return
	}

	var appErr *app.AppError
	if !errors.As(err, &appErr) {
		// Unknown error → internal error (don't expose details)
		appErr = &app.AppError{
			Op:      "unknown",
			Code:    app.CodeInternalError,
			Message: "An internal error occurred",
			Err:     err,
		}
	}

	problem := FromAppError(r, appErr)
	WriteProblem(w, problem)
}

// NewValidationProblem creates a ProblemDetail for validation errors.
// Keep for backward compatibility, returns *Problem now.
func NewValidationProblem(r *http.Request, validationErrors []ValidationError) *Problem {
	// Convert ValidationError to FieldError
	var fieldErrors []FieldError
	if len(validationErrors) > 0 {
		fieldErrors = make([]FieldError, len(validationErrors))
		for i, ve := range validationErrors {
			// Convert ValidationError to FieldError
			fieldErrors[i] = FieldError(ve)

		}
	}

	return NewFieldValidationProblem(r, fieldErrors)
}

// WriteValidationError writes a validation error response.
func WriteValidationError(w http.ResponseWriter, r *http.Request, validationErrors []ValidationError) {
	problem := NewValidationProblem(r, validationErrors)
	WriteProblem(w, problem)
}
