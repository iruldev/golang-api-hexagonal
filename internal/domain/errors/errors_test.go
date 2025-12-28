package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDomainError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "error without wrapped error",
			err:      New(ErrCodeUserNotFound, "user not found"),
			expected: "ERR_USER_NOT_FOUND: user not found",
		},
		{
			name:     "error with wrapped error",
			err:      Wrap(ErrCodeInternal, "database error", errors.New("connection failed")),
			expected: "ERR_INTERNAL: database error: connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestDomainError_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		target   error
		expected bool
	}{
		{
			name:     "same error code matches",
			err:      NewUserNotFound("123"),
			target:   ErrUserNotFound,
			expected: true,
		},
		{
			name:     "different error code does not match",
			err:      NewUserNotFound("123"),
			target:   ErrEmailExists,
			expected: false,
		},
		{
			name:     "sentinel errors match",
			err:      ErrUserNotFound,
			target:   ErrUserNotFound,
			expected: true,
		},
		{
			name:     "wrapped error with same code matches",
			err:      Wrap(ErrCodeUserNotFound, "wrapped", errors.New("cause")),
			target:   ErrUserNotFound,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, errors.Is(tt.err, tt.target))
		})
	}
}

func TestDomainError_As(t *testing.T) {
	t.Run("can extract DomainError", func(t *testing.T) {
		err := NewUserNotFound("user-123")

		var domainErr *DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, ErrCodeUserNotFound, domainErr.Code)
		assert.Contains(t, domainErr.Message, "user-123")
	})

	t.Run("can extract from wrapped error", func(t *testing.T) {
		cause := errors.New("database timeout")
		err := Wrap(ErrCodeInternal, "operation failed", cause)

		var domainErr *DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, ErrCodeInternal, domainErr.Code)
	})
}

func TestDomainError_Unwrap(t *testing.T) {
	t.Run("unwrap returns underlying error", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := Wrap(ErrCodeInternal, "db error", cause)

		u, ok := err.(interface{ Unwrap() error })
		require.True(t, ok)
		unwrapped := u.Unwrap()
		assert.Equal(t, cause, unwrapped)
	})

	t.Run("unwrap returns nil for errors without cause", func(t *testing.T) {
		err := New(ErrCodeUserNotFound, "not found")

		u, ok := err.(interface{ Unwrap() error })
		require.True(t, ok)
		assert.Nil(t, u.Unwrap())
	})

	t.Run("errors.Is finds wrapped error", func(t *testing.T) {
		cause := errors.New("timeout")
		err := Wrap(ErrCodeInternal, "operation failed", cause)

		assert.True(t, errors.Is(err, cause))
	})
}

func TestErrorCode_String(t *testing.T) {
	assert.Equal(t, "ERR_USER_NOT_FOUND", ErrCodeUserNotFound.String())
	assert.Equal(t, "ERR_INTERNAL", ErrCodeInternal.String())
}

func TestNew(t *testing.T) {
	err := New(ErrCodeValidation, "field required")

	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeValidation, domainErr.Code)
	assert.Equal(t, "field required", domainErr.Message)
	assert.Nil(t, domainErr.Err)
}

func TestWrap(t *testing.T) {
	cause := errors.New("original error")
	err := Wrap(ErrCodeInternal, "wrapper message", cause)

	domainErr, ok := err.(*DomainError)
	require.True(t, ok)
	assert.Equal(t, ErrCodeInternal, domainErr.Code)
	assert.Equal(t, "wrapper message", domainErr.Message)
	assert.Equal(t, cause, domainErr.Err)
}

func TestSentinelErrors(t *testing.T) {
	// Verify all sentinel errors have correct codes
	tests := []struct {
		err  error
		code ErrorCode
	}{
		{ErrUserNotFound, ErrCodeUserNotFound},
		{ErrEmailExists, ErrCodeEmailExists},
		{ErrInvalidEmail, ErrCodeInvalidEmail},
		{ErrAuditNotFound, ErrCodeAuditNotFound},
		{ErrInternal, ErrCodeInternal},
		{ErrValidation, ErrCodeValidation},
		{ErrNotFound, ErrCodeNotFound},
		{ErrUnauthorized, ErrCodeUnauthorized},
		{ErrForbidden, ErrCodeForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.code.String(), func(t *testing.T) {
			var domainErr *DomainError
			require.True(t, errors.As(tt.err, &domainErr))
			assert.Equal(t, tt.code, domainErr.Code)
		})
	}
}

func TestConvenienceConstructors(t *testing.T) {
	t.Run("NewUserNotFound", func(t *testing.T) {
		err := NewUserNotFound("abc-123")
		var domainErr *DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, ErrCodeUserNotFound, domainErr.Code)
		assert.Contains(t, domainErr.Message, "abc-123")
		assert.True(t, errors.Is(err, ErrUserNotFound))
	})

	t.Run("NewEmailExists", func(t *testing.T) {
		err := NewEmailExists("test@example.com")
		var domainErr *DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, ErrCodeEmailExists, domainErr.Code)
		assert.Contains(t, domainErr.Message, "test@example.com")
		assert.True(t, errors.Is(err, ErrEmailExists))
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("email is required")
		var domainErr *DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, ErrCodeValidation, domainErr.Code)
		assert.Equal(t, "email is required", domainErr.Message)
	})

	t.Run("NewInternalError", func(t *testing.T) {
		cause := errors.New("db connection failed")
		err := NewInternalError("failed to save", cause)
		var domainErr *DomainError
		require.True(t, errors.As(err, &domainErr))
		assert.Equal(t, ErrCodeInternal, domainErr.Code)
		assert.True(t, errors.Is(err, cause))
	})
}
