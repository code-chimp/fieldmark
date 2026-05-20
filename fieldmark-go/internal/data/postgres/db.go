// Package postgres provides database connectivity and persistence adapters
// for the FieldMark Go stack. All SQL targets the infrastructure-owned
// domain schema (domain.*) plus the framework-local fiber_auth schema.
// This package owns nothing in domain — it only reads and writes to
// tables created by docker/postgres/init scripts; fiber_auth DDL is
// applied by cmd/migrate-fiber-auth.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a pgxpool against the FieldMark PostgreSQL database and
// validates it with a Ping. The caller is responsible for closing the
// pool via pool.Close() at shutdown.
//
// dsn must be a valid libpq-style connection string or URL, e.g.:
//
//	postgres://fieldmark:fieldmark@localhost:5432/fieldmark
func Connect(dsn string) (*pgxpool.Pool, error) {
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: pool open: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return pool, nil
}
