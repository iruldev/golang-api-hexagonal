package fxmodule_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/fx"

	fxmodule "github.com/iruldev/golang-api-hexagonal/internal/infra/fx"
)

func TestModule_GraphIsValid(t *testing.T) {
	err := fx.ValidateApp(fxmodule.Module)
	assert.NoError(t, err, "Fx dependency graph should be valid")
}
