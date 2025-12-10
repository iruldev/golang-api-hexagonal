// Package main is the entry point for the backend service.
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/iruldev/golang-api-hexagonal/internal/app"
)

func main() {
	// Load port from environment (Story 1.3 config preparation)
	port := os.Getenv("APP_HTTP_PORT")
	if port == "" {
		port = "8080"
	}

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
