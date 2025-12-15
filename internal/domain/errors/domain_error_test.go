package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestDomainError_Error(t *testing.T) {
	err := NewDomain(CodeNotFound, "note not found")
	expected := "note not found"

	if err.Error() != expected {
		t.Errorf("Expected Error() = %q, got %q", expected, err.Error())
	}
}

func TestDomainError_Unwrap(t *testing.T) {
	cause := errors.New("database error")
	err := NewDomainWithCause(CodeInternalError, "failed to fetch note", cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Errorf("Expected Unwrap() to return cause, got %v", unwrapped)
	}
}

func TestDomainError_Unwrap_NoCause(t *testing.T) {
	err := NewDomain(CodeNotFound, "not found")

	if err.Unwrap() != nil {
		t.Error("Expected Unwrap() to return nil when no cause")
	}
}

func TestNewDomain(t *testing.T) {
	err := NewDomain(CodeNotFound, "resource not found")

	if err.Code != CodeNotFound {
		t.Errorf("Expected Code = %q, got %q", CodeNotFound, err.Code)
	}
	if err.Message != "resource not found" {
		t.Errorf("Expected Message = %q, got %q", "resource not found", err.Message)
	}
	if err.Hint != "" {
		t.Errorf("Expected Hint to be empty, got %q", err.Hint)
	}
}

func TestNewDomainWithHint(t *testing.T) {
	err := NewDomainWithHint(
		CodeValidationError,
		"invalid email format",
		"email must be a valid email address",
	)

	if err.Code != CodeValidationError {
		t.Errorf("Expected Code = %q, got %q", CodeValidationError, err.Code)
	}
	if err.Message != "invalid email format" {
		t.Errorf("Expected Message = %q, got %q", "invalid email format", err.Message)
	}
	if err.Hint != "email must be a valid email address" {
		t.Errorf("Expected Hint = %q, got %q", "email must be a valid email address", err.Hint)
	}
}

func TestNewDomainWithCause(t *testing.T) {
	cause := errors.New("sql: no rows")
	err := NewDomainWithCause(CodeNotFound, "note not found", cause)

	if err.Code != CodeNotFound {
		t.Errorf("Expected Code = %q, got %q", CodeNotFound, err.Code)
	}
	if err.Message != "note not found" {
		t.Errorf("Expected Message = %q, got %q", "note not found", err.Message)
	}
	if err.Unwrap() != cause {
		t.Errorf("Expected cause to be unwrapped")
	}
}

func TestDomainError_Is_MatchingCode(t *testing.T) {
	err := NewDomain(CodeNotFound, "resource not found")
	target := &DomainError{Code: CodeNotFound}

	if !errors.Is(err, target) {
		t.Error("Expected errors.Is() to return true for matching code")
	}
}

func TestDomainError_Is_DifferentCode(t *testing.T) {
	err := NewDomain(CodeNotFound, "resource not found")
	target := &DomainError{Code: CodeForbidden}

	if errors.Is(err, target) {
		t.Error("Expected errors.Is() to return false for different code")
	}
}

func TestDomainError_Is_EmptyTargetCode(t *testing.T) {
	err := NewDomain(CodeNotFound, "resource not found")
	target := &DomainError{}

	// When target has no code, just check type match
	if !errors.Is(err, target) {
		t.Error("Expected errors.Is() to return true for empty target code (type match)")
	}
}

func TestDomainError_Is_NonDomainError(t *testing.T) {
	err := NewDomain(CodeNotFound, "resource not found")
	target := errors.New("some other error")

	if errors.Is(err, target) {
		t.Error("Expected errors.Is() to return false for non-DomainError target")
	}
}

func TestDomainError_As(t *testing.T) {
	err := NewDomain(CodeNotFound, "resource not found")

	var domainErr *DomainError
	if !errors.As(err, &domainErr) {
		t.Fatal("Expected errors.As() to succeed")
	}

	if domainErr.Code != CodeNotFound {
		t.Errorf("Expected Code = %q, got %q", CodeNotFound, domainErr.Code)
	}
	if domainErr.Message != "resource not found" {
		t.Errorf("Expected Message = %q, got %q", "resource not found", domainErr.Message)
	}
}

func TestDomainError_As_Wrapped(t *testing.T) {
	original := NewDomain(CodeValidationError, "validation failed")
	wrapped := fmt.Errorf("handler error: %w", original)

	var domainErr *DomainError
	if !errors.As(wrapped, &domainErr) {
		t.Fatal("Expected errors.As() to succeed on wrapped error")
	}

	if domainErr.Code != CodeValidationError {
		t.Errorf("Expected Code = %q, got %q", CodeValidationError, domainErr.Code)
	}
}

func TestIsDomainError_Success(t *testing.T) {
	err := NewDomain(CodeNotFound, "not found")

	domainErr := IsDomainError(err)
	if domainErr == nil {
		t.Fatal("Expected IsDomainError to return non-nil")
	}

	if domainErr.Code != CodeNotFound {
		t.Errorf("Expected Code = %q, got %q", CodeNotFound, domainErr.Code)
	}
}

func TestIsDomainError_Wrapped(t *testing.T) {
	original := NewDomain(CodeForbidden, "access denied")
	wrapped := fmt.Errorf("service error: %w", original)

	domainErr := IsDomainError(wrapped)
	if domainErr == nil {
		t.Fatal("Expected IsDomainError to return non-nil for wrapped error")
	}

	if domainErr.Code != CodeForbidden {
		t.Errorf("Expected Code = %q, got %q", CodeForbidden, domainErr.Code)
	}
}

func TestIsDomainError_NotDomainError(t *testing.T) {
	err := errors.New("regular error")

	if IsDomainError(err) != nil {
		t.Error("Expected IsDomainError to return nil for non-DomainError")
	}
}

func TestIsDomainError_Nil(t *testing.T) {
	if IsDomainError(nil) != nil {
		t.Error("Expected IsDomainError to return nil for nil error")
	}
}

func TestDomainError_Is_WrappedWithCause(t *testing.T) {
	cause := errors.New("sql: no rows in result set")
	err := NewDomainWithCause(CodeNotFound, "note not found", cause)

	// Should match the underlying cause with errors.Is
	if !errors.Is(err, cause) {
		t.Error("Expected errors.Is() to find wrapped cause")
	}

	// Should also match DomainError with same code
	target := &DomainError{Code: CodeNotFound}
	if !errors.Is(err, target) {
		t.Error("Expected errors.Is() to match DomainError code")
	}
}

func TestDomainError_ErrorInterface(t *testing.T) {
	var err error = NewDomain(CodeNotFound, "not found")

	// Verify it implements error interface
	if err.Error() != "not found" {
		t.Errorf("Expected error message 'not found', got %q", err.Error())
	}
}
