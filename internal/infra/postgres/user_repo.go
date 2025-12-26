package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/iruldev/golang-api-hexagonal/internal/domain"
	"github.com/iruldev/golang-api-hexagonal/internal/infra/postgres/sqlcgen"
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

// getDBTX extracts the underlying pgx interface from the domain.Querier.
// This is necessary because sqlc generated code requires exact pgx types (DBTX).
func getDBTX(q domain.Querier) (sqlcgen.DBTX, error) {
	switch v := q.(type) {
	case *PoolQuerier:
		pool := v.pool.Pool()
		if pool == nil {
			return nil, fmt.Errorf("database not connected")
		}
		return pool, nil
	case *TxQuerier:
		return v.tx, nil
	default:
		return nil, fmt.Errorf("unsupported querier type: %T", q)
	}
}

// Create stores a new user in the database.
// It returns domain.ErrEmailAlreadyExists if the email is already taken.
func (r *UserRepo) Create(ctx context.Context, q domain.Querier, user *domain.User) error {
	const op = "userRepo.Create"

	dbtx, err := getDBTX(q)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

	// Parse domain.ID to uuid
	uid, err := uuid.Parse(string(user.ID))
	if err != nil {
		return fmt.Errorf("%s: parse ID: %w", op, err)
	}

	// Prepare params
	params := sqlcgen.CreateUserParams{
		ID:        pgtype.UUID{Bytes: uid, Valid: true},
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: pgtype.Timestamptz{Time: user.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: user.UpdatedAt, Valid: true},
	}

	if err := queries.CreateUser(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
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

	dbtx, err := getDBTX(q)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

	uid, err := uuid.Parse(string(id))
	if err != nil {
		return nil, fmt.Errorf("%s: parse ID: %w", op, err)
	}

	dbUser, err := queries.GetUserByID(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", op, domain.ErrUserNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Map generic UUID bytes back to string
	// sqlc with pgx/v5 pgtype.UUID uses Bytes [16]byte.
	// We need to convert that to uuid.UUID to get String().
	uuidVal, err := uuid.FromBytes(dbUser.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("%s: invalid uuid from db: %w", op, err)
	}

	return &domain.User{
		ID:        domain.ID(uuidVal.String()),
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}, nil
}

// List retrieves users with pagination.
// Returns the slice of users, total count of matching users, and any error.
// Results are ordered by created_at DESC, id DESC.
func (r *UserRepo) List(ctx context.Context, q domain.Querier, params domain.ListParams) ([]domain.User, int, error) {
	const op = "userRepo.List"

	dbtx, err := getDBTX(q)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", op, err)
	}
	queries := sqlcgen.New(dbtx)

	// Get total count
	count, err := queries.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: count: %w", op, err)
	}

	if count == 0 {
		return []domain.User{}, 0, nil
	}

	// Get paginated results
	listParams := sqlcgen.ListUsersParams{
		Limit:  int32(params.Limit()),
		Offset: int32(params.Offset()),
	}

	dbUsers, err := queries.ListUsers(ctx, listParams)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: list: %w", op, err)
	}

	users := make([]domain.User, len(dbUsers))
	for i, u := range dbUsers {
		uuidVal, err := uuid.FromBytes(u.ID.Bytes[:])
		if err != nil {
			return nil, 0, fmt.Errorf("%s: invalid uuid from db: %w", op, err)
		}
		users[i] = domain.User{
			ID:        domain.ID(uuidVal.String()),
			Email:     u.Email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			CreatedAt: u.CreatedAt.Time,
			UpdatedAt: u.UpdatedAt.Time,
		}
	}

	return users, int(count), nil
}

// Ensure UserRepo implements domain.UserRepository at compile time.
var _ domain.UserRepository = (*UserRepo)(nil)
