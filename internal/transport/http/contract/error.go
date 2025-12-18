// Package contract provides HTTP transport layer contracts including
// RFC 7807 Problem Details for machine-readable error responses.
package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

// ProblemBaseURL is the default base URL for problem type URIs.
const ProblemBaseURL = "https://api.example.com/problems/"

var problemBaseURL atomic.Value // string

func init() {
	problemBaseURL.Store(ProblemBaseURL)
}

const (
	ProblemTypeValidationErrorSlug = "validation-error"
	ProblemTypeNotFoundSlug        = "not-found"
	ProblemTypeConflictSlug        = "conflict"
	ProblemTypeInternalErrorSlug   = "internal-error"
	ProblemTypeUnauthorizedSlug    = "unauthorized"
	ProblemTypeForbiddenSlug       = "forbidden"
	ProblemTypeRateLimitSlug       = "rate-limit-exceeded"
)

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

// ProblemDetail represents an RFC 7807 Problem Details response.
type ProblemDetail struct {
	Type             string            `json:"type"`
	Title            string            `json:"title"`
	Status           int               `json:"status"`
	Detail           string            `json:"detail"`
	Instance         string            `json:"instance"`
	Code             string            `json:"code"`
	ValidationErrors []ValidationError `json:"validationErrors,omitempty"`
}

// ValidationError represents a single field validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// mapCodeToStatus maps AppError.Code to HTTP status code.
func mapCodeToStatus(code string) int {
	switch code {
	case app.CodeUserNotFound:
		return http.StatusNotFound // 404
	case app.CodeEmailExists:
		return http.StatusConflict // 409
	case app.CodeValidationError:
		return http.StatusBadRequest // 400
	case app.CodeRequestTooLarge:
		return http.StatusRequestEntityTooLarge // 413
	case app.CodeUnauthorized:
		return http.StatusUnauthorized // 401
	case app.CodeForbidden:
		return http.StatusForbidden // 403
	case app.CodeRateLimitExceeded:
		return http.StatusTooManyRequests // 429
	case app.CodeInternalError:
		return http.StatusInternalServerError // 500
	default:
		return http.StatusInternalServerError // 500
	}
}

// codeToTitle returns a human-readable title for the error code.
func codeToTitle(code string) string {
	switch code {
	case app.CodeUserNotFound:
		return "User Not Found"
	case app.CodeEmailExists:
		return "Email Already Exists"
	case app.CodeValidationError:
		return "Validation Error"
	case app.CodeRequestTooLarge:
		return "Request Entity Too Large"
	case app.CodeUnauthorized:
		return "Unauthorized"
	case app.CodeForbidden:
		return "Forbidden"
	case app.CodeRateLimitExceeded:
		return "Too Many Requests"
	case app.CodeInternalError:
		return "Internal Server Error"
	default:
		return "Internal Server Error"
	}
}

func codeToTypeSlug(code string) string {
	switch code {
	case app.CodeValidationError:
		return ProblemTypeValidationErrorSlug
	case app.CodeUserNotFound:
		return ProblemTypeNotFoundSlug
	case app.CodeEmailExists:
		return ProblemTypeConflictSlug
	case app.CodeRequestTooLarge:
		return ProblemTypeValidationErrorSlug
	case app.CodeUnauthorized:
		return ProblemTypeUnauthorizedSlug
	case app.CodeForbidden:
		return ProblemTypeForbiddenSlug
	case app.CodeRateLimitExceeded:
		return ProblemTypeRateLimitSlug
	case app.CodeInternalError:
		return ProblemTypeInternalErrorSlug
	default:
		return ProblemTypeInternalErrorSlug
	}
}

// problemTypeURL returns the RFC 7807 type URL.
func problemTypeURL(slug string) string {
	baseURL, ok := problemBaseURL.Load().(string)
	if !ok || baseURL == "" {
		baseURL = ProblemBaseURL
	}
	return baseURL + slug
}

// safeDetail returns a safe error message (no internal details for 5xx).
func safeDetail(appErr *app.AppError) string {
	if mapCodeToStatus(appErr.Code) >= 500 {
		return "An internal error occurred"
	}
	return appErr.Message
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

func writeProblemJSON(w http.ResponseWriter, status int, problem ProblemDetail) {
	payload, err := json.Marshal(problem)
	if err != nil {
		w.Header().Set("Content-Type", "application/problem+json")
		w.WriteHeader(http.StatusInternalServerError)
		instanceJSON, _ := json.Marshal(problem.Instance)
		_, _ = w.Write([]byte(`{"type":"` + problemTypeURL(ProblemTypeInternalErrorSlug) + `","title":"Internal Server Error","status":500,"detail":"An internal error occurred","instance":` + string(instanceJSON) + `,"code":"INTERNAL_ERROR"}`))
		return
	}

	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_, err = w.Write(payload)
	if err != nil {
		return
	}
}

// WriteProblemJSON writes an RFC 7807 error response.
func WriteProblemJSON(w http.ResponseWriter, r *http.Request, err error) {
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

	status := mapCodeToStatus(appErr.Code)
	var validationErrors []ValidationError
	if appErr.Code == app.CodeValidationError {
		validationErrors = validationErrorsFromAppError(appErr)
		if len(validationErrors) == 0 {
			validationErrors = []ValidationError{{Field: "validation", Message: safeValidationMessage(appErr)}}
		}
	}

	problem := ProblemDetail{
		Type:             problemTypeURL(codeToTypeSlug(appErr.Code)),
		Title:            codeToTitle(appErr.Code),
		Status:           status,
		Detail:           safeDetail(appErr),
		Instance:         r.URL.Path,
		Code:             appErr.Code,
		ValidationErrors: validationErrors,
	}

	writeProblemJSON(w, status, problem)
}

// NewValidationProblem creates a ProblemDetail for validation errors.
func NewValidationProblem(r *http.Request, validationErrors []ValidationError) *ProblemDetail {
	return &ProblemDetail{
		Type:             problemTypeURL(ProblemTypeValidationErrorSlug),
		Title:            "Validation Error",
		Status:           http.StatusBadRequest,
		Detail:           "One or more fields failed validation",
		Instance:         r.URL.Path,
		Code:             app.CodeValidationError,
		ValidationErrors: validationErrors,
	}
}

// WriteValidationError writes a validation error response.
func WriteValidationError(w http.ResponseWriter, r *http.Request, validationErrors []ValidationError) {
	problem := NewValidationProblem(r, validationErrors)
	writeProblemJSON(w, http.StatusBadRequest, *problem)
}
