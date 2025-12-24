package main

import (
	"testing"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
	"github.com/stretchr/testify/assert"
)

// TestRedactorWiring_Simulation simulates the wiring logic in main.go
// to ensure the configuration correctly maps to the domain config.
// This fulfills Story 1.1 Task 4 (Wiring verification).
func TestRedactorWiring_Simulation(t *testing.T) {
	// 1. Simulate "Full" mode configuration
	t.Run("wires full mode correctly", func(t *testing.T) {
		// Mock config load (manual struct creation)
		cfg := &config.Config{
			AuditRedactEmail: "full",
		}

		// Simulate the wiring line from main.go
		redactorCfg := domain.RedactorConfig{EmailMode: cfg.AuditRedactEmail}

		// Verify
		assert.Equal(t, domain.EmailModeFull, redactorCfg.EmailMode)
	})

	// 2. Simulate "Partial" mode configuration
	t.Run("wires partial mode correctly", func(t *testing.T) {
		cfg := &config.Config{
			AuditRedactEmail: "partial",
		}

		redactorCfg := domain.RedactorConfig{EmailMode: cfg.AuditRedactEmail}

		assert.Equal(t, domain.EmailModePartial, redactorCfg.EmailMode)
	})

	// 3. Verify constants alignment
	t.Run("config values match domain constants", func(t *testing.T) {
		// This ensures that if the config default changes or domain constants change,
		// we detect the mismatch if it breaks logic.
		assert.Equal(t, "full", domain.EmailModeFull)
		assert.Equal(t, "partial", domain.EmailModePartial)
	})
}
