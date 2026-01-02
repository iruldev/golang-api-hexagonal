package main_test

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain_Smoke builds the binary and runs a smoke test
// to verify both servers start up and listen on configured ports.
func TestMain_Smoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping smoke test in short mode")
	}

	// 1. Build the binary
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "api-server")

	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	// Handle running from root or cmd/api
	wd, err := os.Getwd()
	require.NoError(t, err)
	if !strings.HasSuffix(wd, "cmd/api") {
		buildCmd = exec.Command("go", "build", "-o", binaryPath, "./cmd/api")
	}

	out, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "Build failed: %s", string(out))

	// 2. Run the binary
	// We'll let the OS choose ports (value 0) and parse them from logs if possible,
	// BUT main.go logging might output ":0" which doesn't tell us the actual port unless
	// specific logic is there to log the *assigned* addr (net.Listener.Addr()).
	// Looking at main.go:
	// logger.Info("public server listening", slog.String("addr", publicAddr)) -> logs configured addr.
	// If configured as ":0", it logs ":0". It doesn't query the listener for the real port.
	// So we must assign specific free ports ourselves to test reliably.

	// Since main.go logs the *configured* address, not the bound address (unless we change main.go to capture listener),
	// using "0" will result in log "listening on :0", and we won't know where to curl.
	// Strategy: Pick random high ports.
	port := 50000 + (time.Now().UnixNano() % 1000)
	internalPort := port + 1

	cmd := exec.Command(binaryPath)
	env := os.Environ()
	env = append(env, fmt.Sprintf("PORT=%d", port))
	env = append(env, fmt.Sprintf("INTERNAL_PORT=%d", internalPort))
	env = append(env, "LOG_LEVEL=debug")
	// Force DB failure to verify IGNORE_DB_STARTUP_ERROR works
	env = append(env, "DATABASE_URL=postgres://user:pass@invalid-host:5432/db")
	env = append(env, "IGNORE_DB_STARTUP_ERROR=true")
	cmd.Env = env

	// Capture stdout/stderr
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	// stderr, _ := cmd.StderrPipe() // merge?
	cmd.Stderr = cmd.Stdout // merge to stdout

	require.NoError(t, cmd.Start())

	// Ensure cleanup
	defer func() {
		_ = cmd.Process.Signal(os.Interrupt)
		_ = cmd.Wait()
	}()

	// 3. Monitor logs for startup readiness
	ready := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(stdout)
		publicUp := false
		internalUp := false

		for scanner.Scan() {
			line := scanner.Text()
			t.Logf("[SERVER LOG] %s", line) // Enabled for debug

			if strings.Contains(line, "public server listening") {
				publicUp = true
			}
			if strings.Contains(line, "internal server listening") {
				internalUp = true
			}

			if publicUp && internalUp {
				close(ready)
				return // Success
			}
		}
	}()

	// Wait up to 10 seconds for startup
	select {
	case <-ready:
		t.Log("Servers started successfully")
	case <-time.After(10 * time.Second):
		t.Fatal("Timed out waiting for servers to start")
	}

	// 4. Verify Endpoints

	// Public /health
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/healthz", port))
	if assert.NoError(t, err) {
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "{}")
	}

	// Internal /metrics
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/metrics", internalPort))
	if assert.NoError(t, err) {
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// Should contain some prometheus output
		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "go_goroutines")
	}

	// Verify Internal endpoint NOT on Public port
	resp, err = http.Get(fmt.Sprintf("http://localhost:%d/metrics", port))
	if assert.NoError(t, err) {
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "/metrics should NOT be on public port")
	}
}
