//go:build contract

// Package contract contains Pact contract testing infrastructure for the golang-api-hexagonal API.
// Contract tests verify that the API meets the expectations defined by consumers.
//
// Prerequisites:
//   - Install Pact FFI: go install github.com/pact-foundation/pact-go/v2/command/pact-go@latest && pact-go install
//
// Run consumer tests: make test-contract-consumer
// Run provider tests: make test-contract-provider
// Run all: make test-contract
package contract

import (
	"os"
	"path/filepath"
)

const (
	// ProviderName is the name of this service as a Pact provider
	ProviderName = "golang-api-hexagonal"

	// DefaultConsumerName is the default consumer name for tests
	DefaultConsumerName = "APIConsumer"

	// PactDir is the directory where generated pact files are stored
	PactDir = "./pacts"
)

// PactConfig holds configuration for Pact tests
type PactConfig struct {
	// Consumer is the name of the consumer application
	Consumer string
	// Provider is the name of the provider application
	Provider string
	// PactDir is the directory to write pact files
	PactDir string
	// LogLevel controls Pact logging verbosity (TRACE, DEBUG, INFO, WARN, ERROR, NONE)
	LogLevel string
}

// DefaultConfig returns a PactConfig with sensible defaults
func DefaultConfig() PactConfig {
	logLevel := os.Getenv("PACT_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "WARN"
	}

	return PactConfig{
		Consumer: DefaultConsumerName,
		Provider: ProviderName,
		PactDir:  getPactDir(),
		LogLevel: logLevel,
	}
}

// getPactDir returns the absolute path to the pacts directory
func getPactDir() string {
	// Try to get the test directory relative to this file
	if wd, err := os.Getwd(); err == nil {
		pactDir := filepath.Join(wd, "pacts")
		if _, err := os.Stat(pactDir); err == nil {
			return pactDir
		}
		// Create if it doesn't exist
		if err := os.MkdirAll(pactDir, 0755); err == nil {
			return pactDir
		}
	}
	return "./pacts"
}
