# Story 6.5: Define SecretProvider Interface

Status: done

## Story

As a developer,
I want a SecretProvider interface abstraction,
So that I can fetch secrets from Vault, AWS SM, or GCP SM.

## Acceptance Criteria

### AC1: SecretProvider interface defined
**Given** `internal/runtimeutil/secrets.go` exists
**When** I review the interface
**Then** methods include: GetSecret(key), GetSecretWithTTL(key)
**And** default implementation reads from environment

---

## Tasks / Subtasks

- [x] **Task 1: Define SecretProvider interface** (AC: #1)
  - [x] Create GetSecret(ctx, key) method
  - [x] Create GetSecretWithTTL(ctx, key) method

- [x] **Task 2: Create EnvSecretProvider default** (AC: #1)
  - [x] Implement that reads from environment variables

- [x] **Task 3: Verify implementation** (AC: #1)
  - [x] Run `make test` - all pass
  - [x] Run `make lint` - 0 issues

---

## Dev Notes

### SecretProvider Interface

```go
// SecretProvider defines secret management abstraction.
type SecretProvider interface {
    // GetSecret retrieves a secret value by key.
    GetSecret(ctx context.Context, key string) (string, error)

    // GetSecretWithTTL retrieves a secret with its TTL.
    GetSecretWithTTL(ctx context.Context, key string) (Secret, error)
}

type Secret struct {
    Value     string
    ExpiresAt time.Time // Zero if no expiry
}
```

### EnvSecretProvider Default

```go
type EnvSecretProvider struct{}

func (p *EnvSecretProvider) GetSecret(ctx context.Context, key string) (string, error) {
    val := os.Getenv(key)
    if val == "" {
        return "", ErrSecretNotFound
    }
    return val, nil
}
```

### Architecture Compliance

**Layer:** `internal/runtimeutil/`
**Pattern:** Interface abstraction for secret management
**Benefit:** Swappable backends (Vault, AWS SM, GCP SM, env)

### File List

Files to create:
- `internal/runtimeutil/secrets.go` - SecretProvider interface
