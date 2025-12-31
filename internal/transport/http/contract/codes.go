// Package contract provides HTTP transport layer contracts including
// RFC 7807 Problem Details for machine-readable error responses.
//
// This file defines the centralized error code registry with taxonomy.
// Error codes follow the format {CATEGORY}-{NNN} where:
//   - CATEGORY is a short identifier for the error domain
//   - NNN is a 3-digit number within the category's reserved range
//
// # Error Code Categories
//
// | Category | Prefix | Range   | HTTP Status | Description                        |
// |----------|--------|---------|-------------|------------------------------------|
// | AUTH     | AUTH   | 001-099 | 401         | Authentication (token, credentials)|
// | AUTHZ    | AUTHZ  | 001-099 | 403         | Authorization (permissions)        |
// | VAL      | VAL    | 001-199 | 400         | Validation (input errors)          |
// | USR      | USR    | 001-099 | 400/404/409 | User domain specific               |
// | DB       | DB     | 001-099 | 500/503     | Database operations                |
// | SYS      | SYS    | 001-099 | 500/503     | System/infrastructure              |
// | RATE     | RATE   | 001-099 | 429         | Rate limiting                      |
// | RES      | RES    | 001-099 | 503         | Resilience (circuit breaker, etc.) |
//
// # Stability
//
// Error codes in this file are STABLE and must not be changed once published.
// Adding new codes is allowed, but modifying or removing existing codes will
// break backward compatibility with API clients.
//
// # Legacy Code Support
//
// The registry also supports legacy ERR_* format codes from domain/app layers
// via the GetErrorCodeInfo function, which handles both formats transparently.
package contract

import (
	"net/http"
	"strings"
)

// -----------------------------------------------------------------------------
// AUTH - Authentication errors (HTTP 401)
// Reserved range: AUTH-001 to AUTH-099
// -----------------------------------------------------------------------------

const (
	// CodeAuthExpiredToken indicates an expired authentication token.
	CodeAuthExpiredToken = "AUTH-001"

	// CodeAuthInvalidToken indicates a malformed or invalid token signature.
	CodeAuthInvalidToken = "AUTH-002"

	// CodeAuthMissingToken indicates no authentication token was provided.
	CodeAuthMissingToken = "AUTH-003"

	// CodeAuthInvalidCredentials indicates invalid username/password combination.
	CodeAuthInvalidCredentials = "AUTH-004"
)

// -----------------------------------------------------------------------------
// AUTHZ - Authorization errors (HTTP 403)
// Reserved range: AUTHZ-001 to AUTHZ-099
// -----------------------------------------------------------------------------

const (
	// CodeAuthzForbidden indicates access to the resource is forbidden.
	CodeAuthzForbidden = "AUTHZ-001"

	// CodeAuthzInsufficientPermissions indicates the user lacks required permissions.
	CodeAuthzInsufficientPermissions = "AUTHZ-002"
)

// -----------------------------------------------------------------------------
// VAL - Validation errors (HTTP 400)
// Reserved range: VAL-001 to VAL-199
// Range 001-099: General validation errors
// Range 100-199: Reserved for specific features (e.g., idempotency)
// -----------------------------------------------------------------------------

const (
	// CodeValRequired indicates a required field is missing.
	CodeValRequired = "VAL-001"

	// CodeValInvalidFormat indicates an invalid format (generic).
	CodeValInvalidFormat = "VAL-002"

	// CodeValInvalidEmail indicates an invalid email address format.
	CodeValInvalidEmail = "VAL-003"

	// CodeValTooShort indicates a value is shorter than minimum length.
	CodeValTooShort = "VAL-004"

	// CodeValTooLong indicates a value exceeds maximum length.
	CodeValTooLong = "VAL-005"

	// CodeValOutOfRange indicates a numeric value is out of allowed range.
	CodeValOutOfRange = "VAL-006"

	// CodeValInvalidType indicates an invalid data type.
	CodeValInvalidType = "VAL-007"

	// CodeValInvalidJSON indicates the request body is not valid JSON.
	CodeValInvalidJSON = "VAL-008"

	// CodeValRequestTooLarge indicates the request body exceeds size limit.
	CodeValRequestTooLarge = "VAL-009"

	// CodeValInvalidUUID indicates an invalid UUID format.
	CodeValInvalidUUID = "VAL-010"

	// CodeValIdempotencyConflict indicates an idempotency key conflict.
	// Reserved for Story 2.4: Idempotency Key Middleware.
	CodeValIdempotencyConflict = "VAL-100"

	// CodeValIdempotencyKeyInvalid indicates an invalid idempotency key format.
	// Story 2.4: Idempotency Key Middleware - requires UUID v4 format.
	CodeValIdempotencyKeyInvalid = "VAL-101"
)

// -----------------------------------------------------------------------------
// USR - User domain errors
// Reserved range: USR-001 to USR-099
// HTTP status varies: 400 for validation, 404 for not found, 409 for conflict
// -----------------------------------------------------------------------------

const (
	// CodeUsrNotFound indicates the requested user was not found (HTTP 404).
	CodeUsrNotFound = "USR-001"

	// CodeUsrEmailExists indicates the email is already registered (HTTP 409).
	CodeUsrEmailExists = "USR-002"

	// CodeUsrInvalidField indicates an invalid user field value (HTTP 400).
	CodeUsrInvalidField = "USR-003"
)

// -----------------------------------------------------------------------------
// DB - Database errors (HTTP 500/503)
// Reserved range: DB-001 to DB-099
// -----------------------------------------------------------------------------

const (
	// CodeDBConnection indicates a database connection failure.
	CodeDBConnection = "DB-001"

	// CodeDBQuery indicates a query execution failure.
	CodeDBQuery = "DB-002"

	// CodeDBTransaction indicates a transaction failure.
	CodeDBTransaction = "DB-003"
)

// -----------------------------------------------------------------------------
// SYS - System errors (HTTP 500/503)
// Reserved range: SYS-001 to SYS-099
// -----------------------------------------------------------------------------

const (
	// CodeSysInternal indicates an internal server error.
	CodeSysInternal = "SYS-001"

	// CodeSysUnavailable indicates the service is temporarily unavailable.
	CodeSysUnavailable = "SYS-002"

	// CodeSysConfig indicates a configuration error.
	CodeSysConfig = "SYS-003"
)

// -----------------------------------------------------------------------------
// RATE - Rate limit errors (HTTP 429)
// Reserved range: RATE-001 to RATE-099
// -----------------------------------------------------------------------------

const (
	// CodeRateLimitExceeded indicates the rate limit has been exceeded.
	CodeRateLimitExceeded = "RATE-001"
)

// -----------------------------------------------------------------------------
// RES - Resilience errors (HTTP 503)
// Reserved range: RES-001 to RES-099
//
// NOTE: These codes are defined in internal/infra/resilience/errors.go.
// This section documents them for reference and registry integration.
//
// RES-001: Circuit breaker is open
// RES-002: Bulkhead capacity exceeded
// RES-003: Timeout exceeded
// RES-004: Maximum retry attempts exceeded
// -----------------------------------------------------------------------------

const (
	// CodeResCircuitOpen indicates the circuit breaker is open.
	// Mirrors resilience.ErrCodeCircuitOpen for transport layer access.
	CodeResCircuitOpen = "RES-001"

	// CodeResBulkheadFull indicates bulkhead capacity has been reached.
	// Mirrors resilience.ErrCodeBulkheadFull for transport layer access.
	CodeResBulkheadFull = "RES-002"

	// CodeResTimeoutExceeded indicates an operation timeout.
	// Mirrors resilience.ErrCodeTimeoutExceeded for transport layer access.
	CodeResTimeoutExceeded = "RES-003"

	// CodeResMaxRetriesExceeded indicates retry attempts exhausted.
	// Mirrors resilience.ErrCodeMaxRetriesExceeded for transport layer access.
	CodeResMaxRetriesExceeded = "RES-004"
)

// -----------------------------------------------------------------------------
// ErrorCodeInfo and Registry
// -----------------------------------------------------------------------------

// ErrorCodeInfo provides metadata for an error code including HTTP status,
// human-readable title, and RFC 7807 problem type slug.
type ErrorCodeInfo struct {
	// Code is the error code in {CATEGORY}-{NNN} format.
	Code string

	// Category is the error category prefix (e.g., "AUTH", "VAL").
	Category string

	// Title is a short, human-readable summary of the error type.
	Title string

	// DetailTemplate is a template for the detailed error message.
	DetailTemplate string

	// HTTPStatus is the associated HTTP status code.
	HTTPStatus int

	// ProblemTypeSlug is the slug for the RFC 7807 type URI.
	ProblemTypeSlug string
}

// Default error info for unknown codes.
var defaultErrorInfo = ErrorCodeInfo{
	Code:            CodeSysInternal,
	Category:        "SYS",
	Title:           "Internal Server Error",
	DetailTemplate:  "An internal error occurred",
	HTTPStatus:      http.StatusInternalServerError,
	ProblemTypeSlug: ProblemTypeInternalErrorSlug,
}

// errorCodeInfoRegistry maps error codes to their metadata.
var errorCodeInfoRegistry = map[string]ErrorCodeInfo{
	// AUTH codes
	CodeAuthExpiredToken: {
		Code:            CodeAuthExpiredToken,
		Category:        "AUTH",
		Title:           "Token Expired",
		DetailTemplate:  "The authentication token has expired",
		HTTPStatus:      http.StatusUnauthorized,
		ProblemTypeSlug: ProblemTypeUnauthorizedSlug,
	},
	CodeAuthInvalidToken: {
		Code:            CodeAuthInvalidToken,
		Category:        "AUTH",
		Title:           "Invalid Token",
		DetailTemplate:  "The authentication token is invalid",
		HTTPStatus:      http.StatusUnauthorized,
		ProblemTypeSlug: ProblemTypeUnauthorizedSlug,
	},
	CodeAuthMissingToken: {
		Code:            CodeAuthMissingToken,
		Category:        "AUTH",
		Title:           "Missing Token",
		DetailTemplate:  "No authentication token was provided",
		HTTPStatus:      http.StatusUnauthorized,
		ProblemTypeSlug: ProblemTypeUnauthorizedSlug,
	},
	CodeAuthInvalidCredentials: {
		Code:            CodeAuthInvalidCredentials,
		Category:        "AUTH",
		Title:           "Invalid Credentials",
		DetailTemplate:  "The username or password is incorrect",
		HTTPStatus:      http.StatusUnauthorized,
		ProblemTypeSlug: ProblemTypeUnauthorizedSlug,
	},

	// AUTHZ codes
	CodeAuthzForbidden: {
		Code:            CodeAuthzForbidden,
		Category:        "AUTHZ",
		Title:           "Forbidden",
		DetailTemplate:  "Access to this resource is forbidden",
		HTTPStatus:      http.StatusForbidden,
		ProblemTypeSlug: ProblemTypeForbiddenSlug,
	},
	CodeAuthzInsufficientPermissions: {
		Code:            CodeAuthzInsufficientPermissions,
		Category:        "AUTHZ",
		Title:           "Insufficient Permissions",
		DetailTemplate:  "You do not have permission to perform this action",
		HTTPStatus:      http.StatusForbidden,
		ProblemTypeSlug: ProblemTypeForbiddenSlug,
	},

	// VAL codes
	CodeValRequired: {
		Code:            CodeValRequired,
		Category:        "VAL",
		Title:           "Required Field Missing",
		DetailTemplate:  "A required field is missing",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValInvalidFormat: {
		Code:            CodeValInvalidFormat,
		Category:        "VAL",
		Title:           "Invalid Format",
		DetailTemplate:  "The value has an invalid format",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValInvalidEmail: {
		Code:            CodeValInvalidEmail,
		Category:        "VAL",
		Title:           "Invalid Email",
		DetailTemplate:  "The email address format is invalid",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValTooShort: {
		Code:            CodeValTooShort,
		Category:        "VAL",
		Title:           "Value Too Short",
		DetailTemplate:  "The value is shorter than the minimum length",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValTooLong: {
		Code:            CodeValTooLong,
		Category:        "VAL",
		Title:           "Value Too Long",
		DetailTemplate:  "The value exceeds the maximum length",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValOutOfRange: {
		Code:            CodeValOutOfRange,
		Category:        "VAL",
		Title:           "Value Out of Range",
		DetailTemplate:  "The value is outside the allowed range",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValInvalidType: {
		Code:            CodeValInvalidType,
		Category:        "VAL",
		Title:           "Invalid Type",
		DetailTemplate:  "The value has an invalid type",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValInvalidJSON: {
		Code:            CodeValInvalidJSON,
		Category:        "VAL",
		Title:           "Invalid JSON",
		DetailTemplate:  "The request body is not valid JSON",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValRequestTooLarge: {
		Code:            CodeValRequestTooLarge,
		Category:        "VAL",
		Title:           "Request Too Large",
		DetailTemplate:  "The request body exceeds the size limit",
		HTTPStatus:      http.StatusRequestEntityTooLarge,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValInvalidUUID: {
		Code:            CodeValInvalidUUID,
		Category:        "VAL",
		Title:           "Invalid UUID",
		DetailTemplate:  "The UUID format is invalid",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},
	CodeValIdempotencyConflict: {
		Code:            CodeValIdempotencyConflict,
		Category:        "VAL",
		Title:           "Idempotency Conflict",
		DetailTemplate:  "The idempotency key already exists with different request data",
		HTTPStatus:      http.StatusConflict,
		ProblemTypeSlug: ProblemTypeConflictSlug,
	},
	CodeValIdempotencyKeyInvalid: {
		Code:            CodeValIdempotencyKeyInvalid,
		Category:        "VAL",
		Title:           "Invalid Idempotency Key",
		DetailTemplate:  "The idempotency key must be a valid UUID v4",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},

	// USR codes
	CodeUsrNotFound: {
		Code:            CodeUsrNotFound,
		Category:        "USR",
		Title:           "User Not Found",
		DetailTemplate:  "The requested user was not found",
		HTTPStatus:      http.StatusNotFound,
		ProblemTypeSlug: ProblemTypeNotFoundSlug,
	},
	CodeUsrEmailExists: {
		Code:            CodeUsrEmailExists,
		Category:        "USR",
		Title:           "Email Already Exists",
		DetailTemplate:  "The email address is already registered",
		HTTPStatus:      http.StatusConflict,
		ProblemTypeSlug: ProblemTypeConflictSlug,
	},
	CodeUsrInvalidField: {
		Code:            CodeUsrInvalidField,
		Category:        "USR",
		Title:           "Invalid User Field",
		DetailTemplate:  "A user field has an invalid value",
		HTTPStatus:      http.StatusBadRequest,
		ProblemTypeSlug: ProblemTypeValidationErrorSlug,
	},

	// DB codes
	CodeDBConnection: {
		Code:            CodeDBConnection,
		Category:        "DB",
		Title:           "Database Connection Failed",
		DetailTemplate:  "Unable to connect to the database",
		HTTPStatus:      http.StatusServiceUnavailable,
		ProblemTypeSlug: ProblemTypeServiceUnavailableSlug,
	},
	CodeDBQuery: {
		Code:            CodeDBQuery,
		Category:        "DB",
		Title:           "Database Query Failed",
		DetailTemplate:  "An error occurred while executing the database query",
		HTTPStatus:      http.StatusInternalServerError,
		ProblemTypeSlug: ProblemTypeInternalErrorSlug,
	},
	CodeDBTransaction: {
		Code:            CodeDBTransaction,
		Category:        "DB",
		Title:           "Transaction Failed",
		DetailTemplate:  "The database transaction failed",
		HTTPStatus:      http.StatusInternalServerError,
		ProblemTypeSlug: ProblemTypeInternalErrorSlug,
	},

	// SYS codes
	CodeSysInternal: {
		Code:            CodeSysInternal,
		Category:        "SYS",
		Title:           "Internal Server Error",
		DetailTemplate:  "An internal error occurred",
		HTTPStatus:      http.StatusInternalServerError,
		ProblemTypeSlug: ProblemTypeInternalErrorSlug,
	},
	CodeSysUnavailable: {
		Code:            CodeSysUnavailable,
		Category:        "SYS",
		Title:           "Service Unavailable",
		DetailTemplate:  "The service is temporarily unavailable",
		HTTPStatus:      http.StatusServiceUnavailable,
		ProblemTypeSlug: ProblemTypeServiceUnavailableSlug,
	},
	CodeSysConfig: {
		Code:            CodeSysConfig,
		Category:        "SYS",
		Title:           "Configuration Error",
		DetailTemplate:  "A configuration error occurred",
		HTTPStatus:      http.StatusInternalServerError,
		ProblemTypeSlug: ProblemTypeInternalErrorSlug,
	},

	// RATE codes
	CodeRateLimitExceeded: {
		Code:            CodeRateLimitExceeded,
		Category:        "RATE",
		Title:           "Rate Limit Exceeded",
		DetailTemplate:  "You have exceeded the rate limit. Please retry later.",
		HTTPStatus:      http.StatusTooManyRequests,
		ProblemTypeSlug: ProblemTypeRateLimitSlug,
	},

	// RES codes (mirrors resilience package codes)
	CodeResCircuitOpen: {
		Code:            CodeResCircuitOpen,
		Category:        "RES",
		Title:           "Circuit Breaker Open",
		DetailTemplate:  "The service is temporarily unavailable due to circuit breaker activation",
		HTTPStatus:      http.StatusServiceUnavailable,
		ProblemTypeSlug: ProblemTypeServiceUnavailableSlug,
	},
	CodeResBulkheadFull: {
		Code:            CodeResBulkheadFull,
		Category:        "RES",
		Title:           "Service Overloaded",
		DetailTemplate:  "The service is currently overloaded. Please retry later.",
		HTTPStatus:      http.StatusServiceUnavailable,
		ProblemTypeSlug: ProblemTypeServiceUnavailableSlug,
	},
	CodeResTimeoutExceeded: {
		Code:            CodeResTimeoutExceeded,
		Category:        "RES",
		Title:           "Operation Timeout",
		DetailTemplate:  "The operation timed out",
		HTTPStatus:      http.StatusServiceUnavailable,
		ProblemTypeSlug: ProblemTypeServiceUnavailableSlug,
	},
	CodeResMaxRetriesExceeded: {
		Code:            CodeResMaxRetriesExceeded,
		Category:        "RES",
		Title:           "Retry Limit Exceeded",
		DetailTemplate:  "The operation failed after maximum retry attempts",
		HTTPStatus:      http.StatusServiceUnavailable,
		ProblemTypeSlug: ProblemTypeServiceUnavailableSlug,
	},
}

// legacyToNewCode maps legacy ERR_* format codes to new taxonomy codes.
// This enables backward compatibility with existing code that uses ERR_* format.
//
// Deprecated: New code should use the new {CATEGORY}-{NNN} format directly.
var legacyToNewCode = map[string]string{
	// User domain errors
	"ERR_USER_NOT_FOUND":    CodeUsrNotFound,
	"ERR_USER_EMAIL_EXISTS": CodeUsrEmailExists,

	// Validation errors
	"ERR_VALIDATION":              CodeValRequired,
	"ERR_USER_INVALID_EMAIL":      CodeValInvalidEmail,
	"ERR_USER_INVALID_FIRST_NAME": CodeUsrInvalidField,
	"ERR_USER_INVALID_LAST_NAME":  CodeUsrInvalidField,

	// Auth errors
	"ERR_UNAUTHORIZED": CodeAuthExpiredToken,
	"ERR_FORBIDDEN":    CodeAuthzForbidden,

	// System errors
	"ERR_INTERNAL":            CodeSysInternal,
	"ERR_NOT_FOUND":           CodeUsrNotFound,
	"ERR_CONFLICT":            CodeUsrEmailExists,
	"ERR_REQUEST_TOO_LARGE":   CodeValRequestTooLarge,
	"ERR_RATE_LIMIT_EXCEEDED": CodeRateLimitExceeded,

	// Additional Validation errors (mapped to generic codes)
	// Additional Validation errors (mapped to generic codes)
	"ERR_AUDIT_INVALID_ID":          CodeValInvalidFormat,
	"ERR_AUDIT_INVALID_TIMESTAMP":   CodeValInvalidFormat,
	"ERR_AUDIT_INVALID_PAYLOAD":     CodeValInvalidFormat,
	"ERR_AUDIT_INVALID_REQUEST_ID":  CodeValInvalidUUID,
	"ERR_AUDIT_INVALID_ENTITY_TYPE": CodeValInvalidType,
	"ERR_AUDIT_INVALID_EVENT_TYPE":  CodeValInvalidFormat,
	"ERR_AUDIT_INVALID_ENTITY_ID":   CodeValInvalidUUID,
}

// GetErrorCodeInfo returns metadata for the given error code.
// It handles both new taxonomy format ({CATEGORY}-{NNN}) and legacy ERR_* format.
// Unknown codes return default internal error metadata.
func GetErrorCodeInfo(code string) ErrorCodeInfo {
	// Check new taxonomy first
	if info, ok := errorCodeInfoRegistry[code]; ok {
		return info
	}

	// Check legacy format and translate
	if newCode, ok := legacyToNewCode[code]; ok {
		if info, ok := errorCodeInfoRegistry[newCode]; ok {
			// Return info but with original code preserved
			result := info
			result.Code = code // Keep original code for backward compatibility
			return result
		}
	}

	// Return default for unknown codes
	return defaultErrorInfo
}

// HTTPStatusForCode returns the HTTP status code for the given error code.
// Unknown codes return 500 Internal Server Error.
func HTTPStatusForCode(code string) int {
	return GetErrorCodeInfo(code).HTTPStatus
}

// TitleForCode returns the human-readable title for the given error code.
// Unknown codes return "Internal Server Error".
func TitleForCode(code string) string {
	return GetErrorCodeInfo(code).Title
}

// ProblemTypeForCode returns the RFC 7807 problem type slug for the given error code.
// Unknown codes return the internal error slug.
func ProblemTypeForCode(code string) string {
	return GetErrorCodeInfo(code).ProblemTypeSlug
}

// CategoryForCode extracts the category from an error code.
// For new taxonomy format, it extracts the prefix before the dash.
// For unknown formats, it returns "SYS".
func CategoryForCode(code string) string {
	if idx := strings.Index(code, "-"); idx > 0 {
		return code[:idx]
	}
	// Legacy ERR_* format or unknown
	return "SYS"
}

// IsNewTaxonomyCode checks if the code follows the new {CATEGORY}-{NNN} format.
func IsNewTaxonomyCode(code string) bool {
	if len(code) < 5 { // Minimum: "A-001"
		return false
	}
	idx := strings.Index(code, "-")
	if idx < 1 || idx >= len(code)-3 {
		return false
	}
	// Check if suffix is 3 digits
	suffix := code[idx+1:]
	if len(suffix) != 3 {
		return false
	}
	for _, c := range suffix {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// TranslateLegacyCode translates a legacy ERR_* code to the new taxonomy format.
// If the code is already in new format or not found, it returns the original code.
func TranslateLegacyCode(code string) string {
	if IsNewTaxonomyCode(code) {
		return code
	}
	if newCode, ok := legacyToNewCode[code]; ok {
		return newCode
	}
	return code
}
