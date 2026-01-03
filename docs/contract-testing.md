# Contract Testing with Pact

This document describes the Pact contract testing setup for the golang-api-hexagonal API.

## Overview

[Pact](https://pact.io) is a contract testing tool that ensures API consumers and providers agree on the contract (request/response format). This project uses `pact-go v2` for Go-based contract testing.

## Prerequisites

### Install Pact FFI Library

The Pact FFI (Foreign Function Interface) library must be installed on your system:

```bash
# Install the pact-go CLI and FFI library
make pact-install

# Or manually:
go install github.com/pact-foundation/pact-go/v2@latest
pact-go install
```

## Running Contract Tests

### Consumer Tests

Consumer tests define the expected contract from the consumer's perspective:

```bash
make test-contract-consumer
```

This generates Pact contract files in `test/contract/pacts/`.

### Provider Verification

Provider verification checks that the provider implementation matches the contracts:

```bash
# Start the server first
make run

# In another terminal, run provider verification
make test-contract-provider
```

### All Contract Tests

```bash
make test-contract
```

## Project Structure

```
test/contract/
├── pact_setup_test.go   # Base configuration
├── consumer_test.go     # Consumer contract definitions
├── provider_test.go     # Provider verification tests
└── pacts/               # Generated contract files
    └── APIConsumer-golang-api-hexagonal.json
```

## Consumer Tests Included

| Test | Endpoint | Description |
|------|----------|-------------|
| TestConsumerHealthEndpoint | GET /healthz | Liveness probe |
| TestConsumerReadinessEndpoint | GET /readyz | Readiness probe with checks |
| TestConsumerListUsers | GET /api/v1/users | User list with pagination |
| TestConsumerGetUserByID | GET /api/v1/users/{id} | Single user retrieval |
| TestConsumerGetUserNotFound | GET /api/v1/users/{id} | 404 RFC 7807 error |
| TestConsumerRateLimitExceeded | GET /api/v1/users | 429 rate limit error |

## Pact Broker (Optional)

For CI/CD integration, you can publish contracts to a Pact Broker:

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `PACT_BROKER_URL` | Pact Broker base URL | Yes (for broker) |
| `PACT_BROKER_TOKEN` | API token for authentication | Yes (for broker) |
| `PACT_LOG_LEVEL` | Log level (TRACE, DEBUG, INFO, WARN, ERROR) | No (default: WARN) |
| `PACT_PROVIDER_TEST` | Set to "true" to run provider tests | No |
| `PROVIDER_BASE_URL` | Provider URL for verification | No (default: localhost:8080) |

### Using PactFlow

[PactFlow](https://pactflow.io) is a managed Pact Broker service:

1. Create an account at https://pactflow.io
2. Get your API token from Settings → API Tokens
3. Set environment variables:
   ```bash
   export PACT_BROKER_URL="https://your-org.pactflow.io"
   export PACT_BROKER_TOKEN="your-api-token"
   ```

### Self-Hosted Broker

Run a self-hosted Pact Broker with Docker:

```yaml
# docker-compose.pact.yml
services:
  pact-broker:
    image: pactfoundation/pact-broker:latest
    ports:
      - "9292:9292"
    environment:
      PACT_BROKER_DATABASE_URL: postgres://postgres:password@db/pact_broker
      PACT_BROKER_BASIC_AUTH_USERNAME: pact
      PACT_BROKER_BASIC_AUTH_PASSWORD: pact
```

## Troubleshooting

### "Pact FFI library not found"

Run `make pact-install` or `pact-go install` to install the native library.

### Consumer tests fail to generate pact files

Ensure the `test/contract/pacts/` directory exists and is writable.

### Provider verification fails

1. Ensure the provider is running at the expected URL
2. Check that state handlers match the provider states in consumer tests
3. Review the `PACT_LOG_LEVEL=DEBUG` output for details

### Build tag warnings in IDE

The "No packages found" warnings for contract test files are expected. These files use `//go:build contract` tags and only compile when running with `-tags=contract`.

## CI Integration

Contract tests run automatically in CI:

1. The `contract-tests` job runs after the main `ci` job
2. Consumer tests run to generate pact files
3. Pact files are uploaded as artifacts for inspection

Provider verification in CI requires either:
- A running provider service (for integration testing)
- A Pact Broker with published contracts

## References

- [Pact Documentation](https://docs.pact.io)
- [pact-go v2 Repository](https://github.com/pact-foundation/pact-go)
- [Consumer Contract Testing](https://docs.pact.io/getting_started/how_pact_works)
