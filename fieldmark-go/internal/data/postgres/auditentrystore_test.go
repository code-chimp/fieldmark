//go:build integration

// Story 2.2 AC5 — Go transactional integrity for AuditEntryStore.Append.
//
// Two tests:
//   - Rollback: open tx, insert project + audit entry, rollback, verify on a
//     fresh pool connection that nothing persists.
//   - Commit: same setup but commit, verify both rows are present, clean up
//     (audit_entry first because project_id FK has no ON DELETE CASCADE).
//
// The cleanup DELETE is test-only — production app code never issues
// UPDATE/DELETE against domain.audit_entry (DDL comment at
// docker/postgres/init/010_domain_tables.sql:187-189).
package postgres_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

func TestAuditAppender_RollbackLeavesNoAuditRow(t *testing.T) {
	pool := openPool(t)
	defer pool.Close()

	ctx := context.Background()
	projectID := uuid.New()
	actorID := uuid.New()
	code := "AUD_" + strings.ToUpper(uuid.NewString()[:10])

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO domain.project
			(id, code, name, status, start_date, compliance_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
	`, projectID, code, "Audit Smoke Project", "Active", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 100); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("insert project: %v", err)
	}

	store := postgres.NewAuditEntryStore()
	entry := &entities.AuditEntry{
		ActorID:    actorID,
		Action:     string(enums.AuditActionProjectCreated),
		EntityType: "Project",
		EntityID:   projectID,
		ProjectID:  &projectID,
	}
	if err := store.Append(ctx, tx, entry); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("append: %v", err)
	}
	if entry.ID == uuid.Nil {
		t.Fatal("expected store to set entry.ID")
	}
	if entry.OccurredAt.IsZero() {
		t.Fatal("expected store to set entry.OccurredAt from RETURNING")
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback: %v", err)
	}

	// Fresh connection — neither row should persist.
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM domain.project WHERE id = $1`, projectID).Scan(&n); err != nil {
		t.Fatalf("project count: %v", err)
	}
	if n != 0 {
		t.Fatalf("rollback persisted project: count = %d, want 0", n)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM domain.audit_entry WHERE entity_id = $1`, projectID).Scan(&n); err != nil {
		t.Fatalf("audit count: %v", err)
	}
	if n != 0 {
		t.Fatalf("rollback persisted audit row: count = %d, want 0", n)
	}
}

func TestAuditAppender_CommitPersistsThenCleanupSucceeds(t *testing.T) {
	pool := openPool(t)
	defer pool.Close()

	ctx := context.Background()
	projectID := uuid.New()
	actorID := uuid.New()
	code := "AUD_" + strings.ToUpper(uuid.NewString()[:10])

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO domain.project
			(id, code, name, status, start_date, compliance_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, now(), now())
	`, projectID, code, "Audit Smoke Project", "Active", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), 100); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("insert project: %v", err)
	}

	store := postgres.NewAuditEntryStore()
	entry := &entities.AuditEntry{
		ActorID:    actorID,
		Action:     string(enums.AuditActionProjectPlacedOnHold),
		EntityType: "Project",
		EntityID:   projectID,
		ProjectID:  &projectID,
	}
	if err := store.Append(ctx, tx, entry); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("append: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Cleanup runs unconditionally — audit_entry first (no ON DELETE CASCADE).
	defer func() {
		if _, err := pool.Exec(ctx, `DELETE FROM domain.audit_entry WHERE entity_id = $1`, projectID); err != nil {
			t.Errorf("cleanup audit_entry: %v", err)
		}
		if _, err := pool.Exec(ctx, `DELETE FROM domain.project WHERE id = $1`, projectID); err != nil {
			t.Errorf("cleanup project: %v", err)
		}
	}()

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM domain.project WHERE id = $1`, projectID).Scan(&n); err != nil {
		t.Fatalf("project count: %v", err)
	}
	if n != 1 {
		t.Fatalf("project count = %d, want 1", n)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM domain.audit_entry WHERE entity_id = $1`, projectID).Scan(&n); err != nil {
		t.Fatalf("audit count: %v", err)
	}
	if n != 1 {
		t.Fatalf("audit count = %d, want 1", n)
	}

	// Persisted action string is the canonical PascalCase form verbatim.
	var persisted string
	if err := pool.QueryRow(ctx, `SELECT action FROM domain.audit_entry WHERE entity_id = $1`, projectID).Scan(&persisted); err != nil {
		t.Fatalf("action lookup: %v", err)
	}
	if persisted != "ProjectPlacedOnHold" {
		t.Fatalf("persisted action = %q, want ProjectPlacedOnHold", persisted)
	}
}
