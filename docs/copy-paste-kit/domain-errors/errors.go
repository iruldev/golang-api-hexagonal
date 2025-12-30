package errors

import "errors"

// ErrorCode is a stable error code type.
type ErrorCode string

// DomainError represents a domain-level error with stable code.
type DomainError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// New creates a new DomainError.
func New(code ErrorCode, message string) error {
	return &DomainError{Code: code, Message: message}
}

// Wrap creates a new DomainError wrapping an existing error.
func Wrap(code ErrorCode, message string, err error) error {
	return &DomainError{Code: code, Message: message, Err: err}
}
