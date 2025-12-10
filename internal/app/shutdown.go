// Package app provides application shutdown handling.
package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ShutdownTimeout is the maximum time to wait for in-flight requests to complete.
const ShutdownTimeout = 30 * time.Second

// GracefulShutdown handles OS signals and shuts down the server gracefully.
// It blocks until SIGINT or SIGTERM is received, then initiates server shutdown.
// The done channel receives nil on successful shutdown, or an error if shutdown fails.
func GracefulShutdown(server *http.Server, done chan<- error) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit) // Clean up signal handler to prevent goroutine leak

	<-quit // Block until signal received

	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		done <- err
		return
	}
	done <- nil
}
