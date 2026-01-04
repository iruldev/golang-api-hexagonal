// Package domain contains the core business entities and interfaces.
//
// This package is the innermost layer of the hexagonal architecture,
// containing pure business logic with no external dependencies.
// It defines entities, value objects, repository interfaces (ports),
// and domain errors.
//
// # Layer Boundary Rules
//
// The domain layer has strict import restrictions enforced by depguard:
//
//	| CAN Import    | CANNOT Import                                           |
//	|---------------|---------------------------------------------------------|
//	| stdlib, subpkgs| slog, otel, uuid, http, pgx, app, transport, infra      |
//
// This ensures the domain remains pure and testable without infrastructure.
//
// # Key Implications
//
//   - Entities MUST NOT have JSON tags (transport layer adds them via DTOs)
//   - Domain MUST NOT log directly (return errors instead)
//   - Domain MUST NOT use external packages (no uuid, no http, no pgx)
//   - Repository interfaces define only the contract, not implementation
//
// # Entities
//
// Domain entities represent core business objects:
//
//	type User struct {
//	    ID        ID
//	    Email     string
//	    FirstName string
//	    LastName  string
//	    CreatedAt time.Time
//	    UpdatedAt time.Time
//	}
//
// Entities include validation methods that return domain errors:
//
//	if err := user.Validate(); err != nil {
//	    return err // Returns domain.ErrInvalidEmail, etc.
//	}
//
// # Repository Interfaces (Ports)
//
// Repository interfaces define persistence contracts implemented by infrastructure:
//
//	type UserRepository interface {
//	    Create(ctx context.Context, q Querier, user *User) error
//	    GetByID(ctx context.Context, q Querier, id ID) (*User, error)
//	    List(ctx context.Context, q Querier, params ListParams) ([]User, int, error)
//	}
//
// The Querier interface enables both direct pool and transaction usage:
//
//	// Use with connection pool
//	user, err := repo.GetByID(ctx, pool, id)
//
//	// Use with transaction
//	err := tx.Do(ctx, func(q Querier) error {
//	    return repo.Create(ctx, q, user)
//	})
//
// # Domain Errors
//
// Domain errors are sentinel errors used for logic control.
// This package aliases errors from internal/domain/errors for backward compatibility:
//
//	var (
//	    ErrInvalidEmail     = errors.ErrInvalidEmail
//	    ErrInvalidFirstName = errors.ErrInvalidFirstName
//	    ErrUserNotFound     = errors.ErrUserNotFound
//	)
//
// For creating new structured errors with codes, use the errors subpackage directly:
//
//	return errors.NewDomainError(errors.CodeUserNotFound, "user not found", nil)
//
// # Value Objects
//
// ID is a value object wrapping identity:
//
//	id := domain.NewID() // Generates new ID
//	id := domain.ParseID("uuid-string") // Parses existing ID
//
// Pagination provides standardized list parameters:
//
//	params := domain.ListParams{Page: 1, PageSize: 20}
//
// # Auditing
//
// The Auditable interface marks entities that support audit logging:
//
//	type Auditable interface {
//	    AuditAction() string
//	    AuditDetails() map[string]any
//	}
//
// # See Also
//
//   - ADR-001: Hexagonal Architecture
//   - ADR-002: Layer Boundary Enforcement
//   - internal/domain/errors: Structured domain errors source
//   - domain/errors: Structured domain errors with codes
package domain
