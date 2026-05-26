//go:build integration

// Package postgres integration tests for action item A3 (Epic 1 retro): real-DB
// harness verifying transactional rollback semantics. Story 2.2's
// append_audit_entry helper will reuse this lane to prove its rollback contract.
//
// Build tag isolates the lane from the default `go test ./...` run so the unit
// suite stays hermetic. Drive it with `make test-go-integration` (see top-level
// Makefile) which sets the tag and FIELDMARK_DATABASE_URL.
package postgres_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func openPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Postgres not reachable at %s: %v", dsn, err)
	}
	return pool
}

func TestRollbackDoesNotPersist(t *testing.T) {
	pool := openPool(t)
	defer pool.Close()

	ctx := context.Background()
	code := "TEST_" + strings.ToUpper(uuid.NewString()[:8])

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	if _, err := tx.Exec(ctx,
		`INSERT INTO domain.trade_type (id, code, name) VALUES ($1, $2, $3)`,
		uuid.NewString(), code, "Rollback smoke",
	); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("insert: %v", err)
	}

	// Visible inside the transaction.
	var inside int
	if err := tx.QueryRow(ctx,
		`SELECT count(*) FROM domain.trade_type WHERE code = $1`, code,
	).Scan(&inside); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("inside select: %v", err)
	}
	if inside != 1 {
		_ = tx.Rollback(ctx)
		t.Fatalf("inside count = %d, want 1", inside)
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Fresh connection from the pool — row must not be visible.
	var after int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM domain.trade_type WHERE code = $1`, code,
	).Scan(&after); err != nil {
		t.Fatalf("after select: %v", err)
	}
	if after != 0 {
		t.Fatalf("rollback persisted row: count = %d, want 0", after)
	}
}

func TestReferenceSeedPresent(t *testing.T) {
	pool := openPool(t)
	defer pool.Close()

	var n int
	if err := pool.QueryRow(context.Background(),
		`SELECT count(*) FROM domain.trade_type WHERE active`,
	).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n == 0 {
		t.Fatal("expected init scripts to have populated domain.trade_type")
	}
}
