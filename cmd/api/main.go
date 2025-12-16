package main

import (
	"fmt"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/infra/config"
)

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Print startup info (no sensitive values)
	fmt.Printf("%s starting...\n", cfg.ServiceName)
	fmt.Printf("  Environment: %s\n", cfg.Env)
	fmt.Printf("  Port: %d\n", cfg.Port)
	fmt.Printf("  Log Level: %s\n", cfg.LogLevel)

	// TODO: Wire up dependencies and start server
	// This will be implemented in subsequent stories:
	// - Story 1.5: Health & Readiness Endpoints
	// - Story 1.6: Makefile Commands
}
