package domain_test

import (
	"errors"
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
	"github.com/stretchr/testify/assert"
)

func TestErrorAliases(t *testing.T) {
	tests := []struct {
		name     string
		alias    error
		original error
	}{
		{"ErrUserNotFound", domain.ErrUserNotFound, domainerrors.ErrUserNotFound},
		{"ErrEmailAlreadyExists", domain.ErrEmailAlreadyExists, domainerrors.ErrEmailExists},
		{"ErrInvalidEmail", domain.ErrInvalidEmail, domainerrors.ErrInvalidEmail},
		{"ErrInvalidFirstName", domain.ErrInvalidFirstName, domainerrors.ErrInvalidFirstName},
		{"ErrInvalidLastName", domain.ErrInvalidLastName, domainerrors.ErrInvalidLastName},
		{"ErrAuditEventNotFound", domain.ErrAuditEventNotFound, domainerrors.ErrAuditNotFound},
		{"ErrInvalidEventType", domain.ErrInvalidEventType, domainerrors.ErrInvalidEventType},
		{"ErrInvalidEntityType", domain.ErrInvalidEntityType, domainerrors.ErrInvalidEntityType},
		{"ErrInvalidEntityID", domain.ErrInvalidEntityID, domainerrors.ErrInvalidEntityID},
		{"ErrInvalidID", domain.ErrInvalidID, domainerrors.ErrInvalidID},
		{"ErrInvalidTimestamp", domain.ErrInvalidTimestamp, domainerrors.ErrInvalidTimestamp},
		{"ErrInvalidPayload", domain.ErrInvalidPayload, domainerrors.ErrInvalidPayload},
		{"ErrInvalidRequestID", domain.ErrInvalidRequestID, domainerrors.ErrInvalidRequestID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Same(t, tt.original, tt.alias, "Alias should point to the exact same error instance")
			assert.True(t, errors.Is(tt.alias, tt.original), "Alias should be usable with errors.Is")
		})
	}
}
