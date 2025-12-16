package response

import (
	"net/http"
	"testing"

	domainerrors "github.com/iruldev/golang-api-hexagonal/internal/domain/errors"
)

func TestMapError_DomainError_InsufficientRole(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeInsufficientRole, "insufficient role")
	status, code := MapError(err)

	if status != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", status)
	}
	if code != domainerrors.CodeInsufficientRole {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeInsufficientRole, code)
	}
}

func TestMapError_DomainError_InsufficientPermission(t *testing.T) {
	err := domainerrors.NewDomain(domainerrors.CodeInsufficientPermission, "insufficient permission")
	status, code := MapError(err)

	if status != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", status)
	}
	if code != domainerrors.CodeInsufficientPermission {
		t.Errorf("Expected code %s, got %s", domainerrors.CodeInsufficientPermission, code)
	}
}
