//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

// Story 2.1 AC5 — Go round-trip smoke for the domain.project mapping.
//
// Pattern mirrors DomainRollbackSmokeTests on the .NET side: a single
// pgx.Tx spans the INSERT and the read; rolling that transaction back
// at teardown means nothing reaches disk — no commit-plus-delete window
// where a crashed test could leak a row.
func TestProjectStore_LoadRoundTrip(t *testing.T) {
	pool := openPool(t)
	t.Cleanup(pool.Close)

	ctx := context.Background()
	id := uuid.New()
	code := "P_" + uuid.NewString()[:8]
	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	targetDate := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	closedAt := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	createdAt := time.Date(2026, 1, 10, 9, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 11, 10, 0, 0, 0, time.UTC)
	desc := "round-trip"

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	// Belt-and-braces: rollback on any test exit path. Once we call
	// Rollback explicitly below, a second call is a no-op.
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`INSERT INTO domain.project
		    (id, code, name, description, status,
		     start_date, target_completion_date, actual_closed_at,
		     compliance_score, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		id, code, "Smoke Project", desc, string(enums.ProjectStatusOnHold),
		startDate, targetDate, closedAt, 87, createdAt, updatedAt,
	); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Exercise the production scan logic via the exported LoadProjectFrom
	// helper, passing the open tx so the read sees the uncommitted insert.
	loaded, err := postgres.LoadProjectFrom(ctx, tx, id)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.ID != id {
		t.Errorf("id = %v, want %v", loaded.ID, id)
	}
	if loaded.Code != code {
		t.Errorf("code = %q, want %q", loaded.Code, code)
	}
	if loaded.Name != "Smoke Project" {
		t.Errorf("name = %q, want %q", loaded.Name, "Smoke Project")
	}
	if loaded.Description == nil || *loaded.Description != desc {
		t.Errorf("description = %v, want %q", loaded.Description, desc)
	}
	if loaded.Status != enums.ProjectStatusOnHold {
		t.Errorf("status = %q, want %q", loaded.Status, enums.ProjectStatusOnHold)
	}
	if !loaded.StartDate.Equal(startDate) {
		t.Errorf("start_date = %v, want %v", loaded.StartDate, startDate)
	}
	if loaded.TargetCompletionDate == nil || !loaded.TargetCompletionDate.Equal(targetDate) {
		t.Errorf("target_completion_date = %v, want %v", loaded.TargetCompletionDate, targetDate)
	}
	if loaded.ActualClosedAt == nil || !loaded.ActualClosedAt.Equal(closedAt) {
		t.Errorf("actual_closed_at = %v, want %v", loaded.ActualClosedAt, closedAt)
	}
	if loaded.ComplianceScore != 87 {
		t.Errorf("compliance_score = %d, want 87", loaded.ComplianceScore)
	}
	if !loaded.CreatedAt.Equal(createdAt) {
		t.Errorf("created_at = %v, want %v", loaded.CreatedAt, createdAt)
	}
	if !loaded.UpdatedAt.Equal(updatedAt) {
		t.Errorf("updated_at = %v, want %v", loaded.UpdatedAt, updatedAt)
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Confirm the rollback was honored — no row escaped the transaction.
	var leaked int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM domain.project WHERE id = $1`, id,
	).Scan(&leaked); err != nil {
		t.Fatalf("post-rollback count: %v", err)
	}
	if leaked != 0 {
		t.Fatalf("rollback persisted row: count = %d, want 0", leaked)
	}
}

// LoadWithRelations smokes the snapshot-tx read path against a committed
// project. Cleanup deletes the row before the pool closes (t.Cleanup runs
// LIFO; pool.Close registered first, delete registered second).
func TestProjectStore_LoadWithRelations_SmokesSnapshotPath(t *testing.T) {
	pool := openPool(t)
	t.Cleanup(pool.Close)

	ctx := context.Background()
	id := uuid.New()
	code := "P_" + uuid.NewString()[:8]
	now := time.Now().UTC()

	if _, err := pool.Exec(ctx,
		`INSERT INTO domain.project
		    (id, code, name, status, start_date, compliance_score, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, code, "Snapshot Smoke", string(enums.ProjectStatusActive),
		now, 100, now, now,
	); err != nil {
		t.Fatalf("insert: %v", err)
	}
	t.Cleanup(func() {
		if _, err := pool.Exec(context.Background(),
			`DELETE FROM domain.project WHERE id = $1`, id); err != nil {
			t.Logf("cleanup delete: %v", err)
		}
	})

	store := postgres.NewProjectStore(pool)
	project, sites, scopes, inspectors, err := store.LoadWithRelations(ctx, id)
	if err != nil {
		t.Fatalf("load with relations: %v", err)
	}
	if project.ID != id {
		t.Errorf("project.id = %v, want %v", project.ID, id)
	}
	if len(sites) != 0 || len(scopes) != 0 || len(inspectors) != 0 {
		t.Errorf("expected empty relations, got %d sites, %d scopes, %d inspectors",
			len(sites), len(scopes), len(inspectors))
	}
}

func TestProjectStore_LoadNotFound(t *testing.T) {
	pool := openPool(t)
	t.Cleanup(pool.Close)

	store := postgres.NewProjectStore(pool)
	_, err := store.Load(context.Background(), uuid.New())
	if !errors.Is(err, postgres.ErrProjectNotFound) {
		t.Fatalf("err = %v, want ErrProjectNotFound", err)
	}
}

// Compile-time check: pgx.Tx satisfies the exported Querier interface so
// callers (including the round-trip smoke above) can use LoadProjectFrom
// with a transaction.
var _ postgres.Querier = (pgx.Tx)(nil)
