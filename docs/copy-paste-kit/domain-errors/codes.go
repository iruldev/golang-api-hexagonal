// Package errors provides domain-level error codes and error types.
// This is a copy-paste kit example of how to structure domain errors.
package errors

const (
	// Standard error codes.
	CodeInternal    ErrorCode = "INTERNAL_ERROR"
	CodeNotFound    ErrorCode = "NOT_FOUND"
	CodeInvalidArgs ErrorCode = "INVALID_ARGUMENTS"
)
