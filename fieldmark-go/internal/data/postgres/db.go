// Package postgres provides database connectivity and persistence adapters
// for the FieldMark Go stack. All SQL targets the infrastructure-owned
// domain schema (domain.*). This package owns nothing in domain — it only
// reads and writes to tables created by docker/postgres/init scripts.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Connect opens a connection to the FieldMark PostgreSQL database and
// validates it with a Ping. The caller is responsible for closing the
// connection via conn.Close(ctx) when done.
//
// dsn must be a valid libpq-style connection string or URL, e.g.:
//
//	postgres://fieldmark:fieldmark@localhost:5432/fieldmark
func Connect(dsn string) (*pgx.Conn, error) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return conn, nil
}
