// Package contract provides HTTP transport layer contracts including
// RFC 7807 Problem Details for machine-readable error responses.
//
// This file provides error code mapping utilities that convert various error types
// to their corresponding error codes from the taxonomy defined in codes.go.
package contract

import (
	"errors"
	"strings"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

// codeGetter is an interface for errors that have an error code.
// This allows us to check for resilience errors without importing infra/resilience.
type codeGetter interface {
	GetCode() string
}

// MapErrorToCode determines the error code for any error type.
// It follows a priority order:
//  1. domain.DomainError → mapped via domain code lookup
//  2. app.AppError → mapped via app code lookup
//  3. Errors implementing codeGetter with RES-xxx codes → resilience errors
//  4. Unknown errors → SYS-001 (Internal Server Error)
//
// This function provides a centralized way to determine error codes
// without needing to know the specific error type.
func MapErrorToCode(err error) string {
	if err == nil {
		return CodeSysInternal
	}

	// Priority 1: Check for DomainError
	var domainErr *domainerrors.DomainError
	if errors.As(err, &domainErr) {
		return mapDomainErrorToCode(domainErr)
	}

	// Priority 2: Check for AppError
	var appErr *app.AppError
	if errors.As(err, &appErr) {
		return mapAppErrorToCode(appErr)
	}

	// Priority 3: Check for errors with GetCode() method (e.g., ResilienceError)
	// This avoids importing infra/resilience directly (depguard requirement)
	var codeErr codeGetter
	if errors.As(err, &codeErr) {
		code := codeErr.GetCode()
		if strings.HasPrefix(code, "RES-") {
			return mapResilienceCode(code)
		}
	}

	// Default: Unknown error → SYS-001
	return CodeSysInternal
}

// mapDomainErrorToCode maps domain error codes to the new taxonomy.
func mapDomainErrorToCode(domainErr *domainerrors.DomainError) string {
	if domainErr == nil {
		return CodeSysInternal
	}

	// Translate legacy domain code to new taxonomy
	return TranslateLegacyCode(string(domainErr.Code))
}

// mapAppErrorToCode maps app error codes to the new taxonomy.
func mapAppErrorToCode(appErr *app.AppError) string {
	if appErr == nil {
		return CodeSysInternal
	}

	// Translate legacy app code to new taxonomy
	return TranslateLegacyCode(appErr.Code)
}

// mapResilienceCode maps resilience error codes.
// Resilience errors already use the correct RES-xxx format.
func mapResilienceCode(code string) string {
	switch code {
	case CodeResCircuitOpen:
		return CodeResCircuitOpen
	case CodeResBulkheadFull:
		return CodeResBulkheadFull
	case CodeResTimeoutExceeded:
		return CodeResTimeoutExceeded
	case CodeResMaxRetriesExceeded:
		return CodeResMaxRetriesExceeded
	default:
		// Unknown resilience code, fallback to service unavailable
		return CodeSysUnavailable
	}
}

// MapErrorToHTTPStatus returns the HTTP status code for any error type.
// Convenience function that combines MapErrorToCode with HTTPStatusForCode.
func MapErrorToHTTPStatus(err error) int {
	code := MapErrorToCode(err)
	return HTTPStatusForCode(code)
}

// IsClientError returns true if the error should result in a 4xx response.
func IsClientError(err error) bool {
	status := MapErrorToHTTPStatus(err)
	return status >= 400 && status < 500
}

// IsServerError returns true if the error should result in a 5xx response.
func IsServerError(err error) bool {
	status := MapErrorToHTTPStatus(err)
	return status >= 500
}
