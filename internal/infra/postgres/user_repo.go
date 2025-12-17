package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
)

const (
	// pgUniqueViolation is the PostgreSQL error code for unique constraint violations.
	pgUniqueViolation = "23505"
)

// UserRepo implements domain.UserRepository for PostgreSQL.
type UserRepo struct{}

// NewUserRepo creates a new UserRepo instance.
func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

// Create stores a new user in the database.
// It returns domain.ErrEmailAlreadyExists if the email is already taken.
func (r *UserRepo) Create(ctx context.Context, q domain.Querier, user *domain.User) error {
	const op = "userRepo.Create"

	// Parse domain.ID to uuid.UUID at repository boundary
	id, err := uuid.Parse(string(user.ID))
	if err != nil {
		return fmt.Errorf("%s: parse ID: %w", op, err)
	}

	_, err = q.Exec(ctx, `
		INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, id, user.Email, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			// Only map the specific unique constraint we own for email.
			// Other unique violations (e.g., duplicate primary key) should not be reported as email conflicts.
			if pgErr.ConstraintName == "uniq_users_email" {
				return fmt.Errorf("%s: %w", op, domain.ErrEmailAlreadyExists)
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetByID retrieves a user by their ID.
// It returns domain.ErrUserNotFound if no user exists with the given ID.
func (r *UserRepo) GetByID(ctx context.Context, q domain.Querier, id domain.ID) (*domain.User, error) {
	const op = "userRepo.GetByID"

	// Parse domain.ID to uuid.UUID
	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, fmt.Errorf("%s: parse ID: %w", op, err)
	}

	row := q.QueryRow(ctx, `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users WHERE id = $1
	`, uid)

	// Type assert to rowScanner interface
	scanner, ok := row.(rowScanner)
	if !ok {
		return nil, fmt.Errorf("%s: invalid querier type", op)
	}

	var user domain.User
	var dbID uuid.UUID
	err = scanner.Scan(&dbID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user.ID = domain.ID(dbID.String())
	return &user, nil
}

// List retrieves users with pagination.
// Returns the slice of users, total count of matching users, and any error.
// Results are ordered by created_at DESC, id DESC.
func (r *UserRepo) List(ctx context.Context, q domain.Querier, params domain.ListParams) ([]domain.User, int, error) {
	const op = "userRepo.List"

	// Get total count
	countRow := q.QueryRow(ctx, `SELECT COUNT(*) FROM users`)
	countScanner, ok := countRow.(rowScanner)
	if !ok {
		return nil, 0, fmt.Errorf("%s: invalid querier type for count", op)
	}

	var totalCount int
	if err := countScanner.Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("%s: count: %w", op, err)
	}

	// If no results, return early
	if totalCount == 0 {
		return []domain.User{}, 0, nil
	}

	// Get paginated results
	rows, err := q.Query(ctx, `
		SELECT id, email, first_name, last_name, created_at, updated_at
		FROM users
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2
	`, params.Limit(), params.Offset())
	if err != nil {
		return nil, 0, fmt.Errorf("%s: query: %w", op, err)
	}

	// Type assert to rowsScanner interface
	scanner, ok := rows.(rowsScanner)
	if !ok {
		return nil, 0, fmt.Errorf("%s: invalid querier type for rows", op)
	}
	defer scanner.Close()

	var users []domain.User
	for scanner.Next() {
		var user domain.User
		var dbID uuid.UUID
		if err := scanner.Scan(&dbID, &user.Email, &user.FirstName, &user.LastName, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("%s: scan: %w", op, err)
		}
		user.ID = domain.ID(dbID.String())
		users = append(users, user)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, fmt.Errorf("%s: rows: %w", op, err)
	}

	return users, totalCount, nil
}

// Ensure UserRepo implements domain.UserRepository at compile time.
var _ domain.UserRepository = (*UserRepo)(nil)
