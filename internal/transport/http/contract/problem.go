// Package contract provides HTTP transport layer contracts including
// RFC 7807 Problem Details for machine-readable error responses.
//
// This file implements Problem type using github.com/moogar0880/problems
// for RFC 7807 compliance with project-specific extensions.
package contract

import (
	"encoding/json"
	"net/http"

	"github.com/moogar0880/problems"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/iruldev/golang-api-hexagonal/internal/transport/http/ctxutil"
)

const (
	ProblemTypeValidationErrorSlug    = "validation-error"
	ProblemTypeNotFoundSlug           = "not-found"
	ProblemTypeConflictSlug           = "conflict"
	ProblemTypeInternalErrorSlug      = "internal-error"
	ProblemTypeUnauthorizedSlug       = "unauthorized"
	ProblemTypeForbiddenSlug          = "forbidden"
	ProblemTypeRateLimitSlug          = "rate-limit-exceeded"
	ProblemTypeServiceUnavailableSlug = "service-unavailable"

	ContentTypeProblemJSON = "application/problem+json"
)

// FieldError represents a single field validation error with code.
// This extends ValidationError to include an error code per field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// ValidationError represents a single field validation error.
// Kept for backward compatibility.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Problem represents an RFC 7807 Problem Details response with project-specific extensions.
// It embeds the moogar0880/problems DefaultProblem for core RFC 7807 compliance
// and adds extension fields for error correlation and validation details.
//
// Thread-safety: Problem instances are not safe for concurrent modification.
// Create a new Problem for each error response.
// Problem represents an RFC 7807 Problem Details response with project-specific extensions.
// It embeds the moogar0880/problems DefaultProblem for core RFC 7807 compliance
// and adds extension fields for error correlation and validation details.
type Problem struct {
	*problems.DefaultProblem

	// Extension fields
	Code      string `json:"code,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	TraceID   string `json:"trace_id,omitempty"`

	// Errors contains per-field validation errors (AC2)
	Errors []FieldError `json:"errors,omitempty"`

	// ValidationErrors is for backward compatibility
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

// NewProblem creates a new Problem from a problems.DefaultProblem.
// It copies the core RFC 7807 fields from the library type.
func NewProblem(status int, title, detail string) *Problem {
	baseProblem := problems.NewStatusProblem(status)
	baseProblem.Detail = detail
	baseProblem.Title = title // Override title if provided

	return &Problem{
		DefaultProblem: baseProblem,
	}
}

// NewProblemWithType creates a Problem with a custom type URI.
func NewProblemWithType(typeURI string, status int, title, detail string) *Problem {
	baseProblem := problems.NewStatusProblem(status)
	baseProblem.Type = typeURI
	baseProblem.Title = title
	baseProblem.Detail = detail

	return &Problem{
		DefaultProblem: baseProblem,
	}
}

// NewFieldValidationProblem creates a Problem for validation errors with FieldError details.
// It populates the Errors field with per-field validation details including error codes,
// and extracts request_id and trace_id from the request context.
//
// This function extends the existing NewValidationProblem by supporting FieldError
// which includes a Code field for per-field error codes.
func NewFieldValidationProblem(r *http.Request, fieldErrors []FieldError) *Problem {
	baseProblem := problems.NewDetailedProblem(http.StatusBadRequest, "One or more fields failed validation")
	baseProblem.Type = problemTypeURL(ProblemTypeValidationErrorSlug)
	baseProblem.Title = "Validation Error"

	problem := &Problem{
		DefaultProblem:   baseProblem,
		Code:             app.CodeValidationError,
		Errors:           fieldErrors,
		ValidationErrors: fieldErrorsToValidationErrors(fieldErrors),
	}

	if r != nil {
		problem.Instance = r.URL.Path
		populateProblemIDs(r, problem)
	}

	return problem
}

// FromAppError creates a Problem from an app.AppError.
// It maps the error code to HTTP status and problem type using the error registry,
// and extracts request_id and trace_id from the request context.
func FromAppError(r *http.Request, appErr *app.AppError) *Problem {
	if appErr == nil {
		// Use CodeSysInternal as default for nil error
		def := GetErrorCodeInfo(CodeSysInternal)
		return NewProblem(def.HTTPStatus, def.Title, def.DetailTemplate)
	}

	// Translate legacy code to new taxonomy if applicable (Story 2.2)
	code := TranslateLegacyCode(appErr.Code)

	// Get metadata from registry
	info := GetErrorCodeInfo(code)

	var fieldErrors []FieldError
	if code == CodeValRequired || code == CodeValInvalidFormat || appErr.Code == app.CodeValidationError {
		fieldErrors = convertValidationErrors(appErr)
		// Fallback for generic validation errors without specific fields
		if len(fieldErrors) == 0 {
			fieldErrors = []FieldError{{Field: "validation", Message: safeValidationMessage(appErr)}}
		}
	}

	// Use safe detail for 5xx errors to avoid leaking internal info
	detail := appErr.Message
	if info.HTTPStatus >= 500 {
		detail = info.DetailTemplate
	}

	baseProblem := problems.NewDetailedProblem(info.HTTPStatus, detail)
	baseProblem.Type = problemTypeURL(info.ProblemTypeSlug)
	baseProblem.Title = info.Title

	problem := &Problem{
		DefaultProblem:   baseProblem,
		Code:             info.Code,
		Errors:           fieldErrors,
		ValidationErrors: fieldErrorsToValidationErrors(fieldErrors),
	}

	if r != nil {
		problem.Instance = r.URL.Path
		populateProblemIDs(r, problem)
	}

	return problem
}

// FromDomainError creates a Problem from a domain.DomainError.
// It maps the domain error code to HTTP status and problem type,
// and extracts request_id and trace_id from the request context.
func FromDomainError(r *http.Request, domainErr *domainerrors.DomainError) *Problem {
	if domainErr == nil {
		def := GetErrorCodeInfo(CodeSysInternal)
		return NewProblem(def.HTTPStatus, def.Title, def.DetailTemplate)
	}

	// Use string(domainErr.Code) to look up in new registry
	// Domain codes like "ERR_USER_NOT_FOUND" will be translated via GetErrorCodeInfo's internal lookup
	// or we can explicitly translate if we want to enforce new format output
	code := TranslateLegacyCode(string(domainErr.Code))
	info := GetErrorCodeInfo(code)

	var fieldErrors []FieldError
	if info.ProblemTypeSlug == ProblemTypeValidationErrorSlug {
		fieldErrors = mapDomainValidationErrors(domainErr)
	}

	baseProblem := problems.NewDetailedProblem(info.HTTPStatus, domainErr.Message)
	baseProblem.Type = problemTypeURL(info.ProblemTypeSlug)
	baseProblem.Title = info.Title

	problem := &Problem{
		DefaultProblem:   baseProblem,
		Code:             info.Code,
		Errors:           fieldErrors,
		ValidationErrors: fieldErrorsToValidationErrors(fieldErrors),
	}

	if r != nil {
		problem.Instance = r.URL.Path
		populateProblemIDs(r, problem)
	}

	return problem
}

// populateProblemIDs extracts request_id and trace_id from context and sets them on the Problem.
func populateProblemIDs(r *http.Request, problem *Problem) {
	if r == nil || problem == nil {
		return
	}

	problem.RequestID = ctxutil.GetRequestID(r.Context())
	if traceID := ctxutil.GetTraceID(r.Context()); traceID != "" && traceID != ctxutil.EmptyTraceID {
		problem.TraceID = traceID
	}
}

// safeAppErrorDetail returns a safe error message that doesn't expose internal details for 5xx errors.
func safeAppErrorDetail(appErr *app.AppError) string {
	if appErr == nil {
		return "An internal error occurred"
	}
	if HTTPStatusForCode(appErr.Code) >= 500 {
		return "An internal error occurred"
	}
	return appErr.Message
}

// convertValidationErrors converts app.AppError validation errors to FieldError slice.
func convertValidationErrors(appErr *app.AppError) []FieldError {
	if appErr == nil {
		return nil
	}

	// Use existing validation error extraction logic
	legacyErrors := validationErrorsFromAppError(appErr)
	if len(legacyErrors) == 0 {
		return nil
	}

	fieldErrors := make([]FieldError, len(legacyErrors))
	for i, ve := range legacyErrors {
		fieldErrors[i] = FieldError{
			Field:   ve.Field,
			Message: ve.Message,
		}
	}
	return fieldErrors
}

// mapDomainValidationErrors maps domain validation errors to FieldError slice.
func mapDomainValidationErrors(domainErr *domainerrors.DomainError) []FieldError {
	if domainErr == nil {
		return nil
	}

	switch domainErr.Code {
	case domainerrors.ErrCodeInvalidEmail:
		return []FieldError{{Field: "email", Message: "must be a valid email address"}}
	case domainerrors.ErrCodeInvalidFirstName:
		return []FieldError{{Field: "firstName", Message: "must not be empty"}}
	case domainerrors.ErrCodeInvalidLastName:
		return []FieldError{{Field: "lastName", Message: "must not be empty"}}
	default:
		if domainErr.Message != "" {
			return []FieldError{{Field: "validation", Message: domainErr.Message}}
		}
		return nil
	}
}

// WriteProblem writes the Problem as an RFC 7807 JSON response.
// It sets the Content-Type to application/problem+json and writes the status code.
func WriteProblem(w http.ResponseWriter, problem *Problem) {
	if problem == nil {
		problem = NewProblem(http.StatusInternalServerError, "Internal Server Error", "An internal error occurred")
	}
	if problem.Status == 0 {
		problem.Status = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", ContentTypeProblemJSON)
	w.WriteHeader(problem.Status)

	// We use json.NewEncoder which respects the json tags on the Problem struct.
	// Note: We explicitly ignore the encoding error here because we've already written
	// the status code and headers. If encoding fails, we cannot send a different
	// status code or useful error response at this point.
	_ = json.NewEncoder(w).Encode(problem)
}

// fieldErrorsToValidationErrors converts FieldError slice to ValidationError slice for backward compatibility.
func fieldErrorsToValidationErrors(fieldErrors []FieldError) []ValidationError {
	if len(fieldErrors) == 0 {
		return nil
	}
	legacyErrors := make([]ValidationError, len(fieldErrors))
	for i, fe := range fieldErrors {
		legacyErrors[i] = ValidationError{
			Field:   fe.Field,
			Message: fe.Message,
		}
	}
	return legacyErrors
}
