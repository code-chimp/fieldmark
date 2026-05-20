// Command migrate-fiber-auth applies the framework-local fiber_auth DDL
// (idempotent via CREATE TABLE IF NOT EXISTS). Invoke after `make reset`
// to bring up the Go-stack auth tables. ADR-012 stub posture: this is
// not a general-purpose migration runner; real auth migration tooling
// lands when the deferred Go-auth epic begins.
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	fam "github.com/code-chimp/fieldmark-go/internal/data/postgres/migrations/fiber_auth"
)

func main() {
	dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("migrate-fiber-auth: pool: %v", err)
	}
	defer pool.Close()

	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Fatalf("migrate-fiber-auth: begin: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }() // no-op after commit

	if _, err := tx.Exec(ctx, fam.InitialSQL); err != nil {
		log.Fatalf("migrate-fiber-auth: exec: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("migrate-fiber-auth: commit: %v", err)
	}

	log.Println("fiber_auth: schema applied (idempotent)")
}
