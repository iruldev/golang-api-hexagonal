// Package main is the entry point for the backend service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
	"github.com/iruldev/golang-api-hexagonal/internal/config"
	httpx "github.com/iruldev/golang-api-hexagonal/internal/interface/http"
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

	// Create chi router with versioned API routes (Story 3.1)
	router := httpx.NewRouter(cfg)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
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

	// Shutdown tracer provider to flush remaining spans (Story 3.5)
	if httpx.TracerShutdown != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpx.TracerShutdown(ctx); err != nil {
			log.Printf("Tracer shutdown error: %v", err)
		}
	}

	log.Println("Server shutdown complete")
	os.Exit(0)
}
