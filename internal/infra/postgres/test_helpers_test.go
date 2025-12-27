package postgres_test

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// dbAdapter adapts *pgxpool.Pool to postgres.Pooler interface for tests.
// This allows tests in postgres_test package to use NewPoolQuerier.
type dbAdapter struct {
	p *pgxpool.Pool
}

func (a *dbAdapter) Ping(ctx context.Context) error {
	return a.p.Ping(ctx)
}

func (a *dbAdapter) Close() {
	a.p.Close()
}

func (a *dbAdapter) Pool() *pgxpool.Pool {
	return a.p
}
