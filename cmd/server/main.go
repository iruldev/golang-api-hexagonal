// Package main is the entry point for the backend service.
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
)

func main() {
	// Load and validate configuration (Epic 2: Configuration & Environment)
	cfg, err := config.Load()
	if err != nil {
		// Exit code 1 with clear error message (Story 2.5)
		log.Fatalf("Configuration error: %v", err)
	}

	// Use typed config instead of raw os.Getenv
	port := fmt.Sprintf("%d", cfg.App.HTTPPort)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: nil, // Router will be added in Story 3.x
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal (Story 1.4 graceful shutdown)
	done := make(chan error, 1)
	go app.GracefulShutdown(server, done)

	if err := <-done; err != nil {
		log.Printf("Shutdown error: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
	os.Exit(0)
}
