//go:build !integration

package app

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		inner := errors.New("inner error")
		appErr := &AppError{
			Op:      "GetUser",
			Code:    CodeUserNotFound,
			Message: "User not found",
			Err:     inner,
		}

		assert.Equal(t, "GetUser: User not found: inner error", appErr.Error())
	})

	t.Run("without wrapped error", func(t *testing.T) {
		appErr := &AppError{
			Op:      "GetUser",
			Code:    CodeUserNotFound,
			Message: "User not found",
			Err:     nil,
		}

		assert.Equal(t, "GetUser: User not found", appErr.Error())
	})
}

func TestAppError_Unwrap(t *testing.T) {
	inner := errors.New("wrapped")
	appErr := &AppError{
		Op:      "GetUser",
		Code:    CodeInternalError,
		Message: "boom",
		Err:     inner,
	}

	assert.ErrorIs(t, appErr, inner)
	assert.Equal(t, inner, appErr.Unwrap())
}
