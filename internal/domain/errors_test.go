package domain

import (
	"errors"
	"testing"
)

func TestErrNotFound_Is(t *testing.T) {
	err := ErrNotFound
	if !errors.Is(err, ErrNotFound) {
		t.Error("Expected errors.Is to return true for ErrNotFound")
	}
}

func TestErrValidation_Is(t *testing.T) {
	err := ErrValidation
	if !errors.Is(err, ErrValidation) {
		t.Error("Expected errors.Is to return true for ErrValidation")
	}
}

func TestErrUnauthorized_Is(t *testing.T) {
	err := ErrUnauthorized
	if !errors.Is(err, ErrUnauthorized) {
		t.Error("Expected errors.Is to return true for ErrUnauthorized")
	}
}

func TestWrapError_Unwrap(t *testing.T) {
	wrapped := WrapError(ErrNotFound, "user not found")

	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("Expected wrapped error to unwrap to ErrNotFound")
	}
}

func TestWrapError_Message(t *testing.T) {
	wrapped := WrapError(ErrNotFound, "user with id 123 not found")

	if wrapped.Error() != "user with id 123 not found" {
		t.Errorf("Expected custom message, got %s", wrapped.Error())
	}
}

func TestWrapError_DifferentTypes(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
	}{
		{"NotFound", ErrNotFound, "resource not found"},
		{"Validation", ErrValidation, "invalid input"},
		{"Unauthorized", ErrUnauthorized, "invalid token"},
		{"Forbidden", ErrForbidden, "access denied"},
		{"Conflict", ErrConflict, "already exists"},
		{"Internal", ErrInternal, "something went wrong"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WrapError(tt.err, tt.message)
			if !errors.Is(wrapped, tt.err) {
				t.Errorf("Expected wrapped error to match %v", tt.err)
			}
			if wrapped.Error() != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, wrapped.Error())
			}
		})
	}
}
