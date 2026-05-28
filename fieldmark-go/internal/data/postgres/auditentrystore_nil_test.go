package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
)

// Story 2.2 review patch — AuditEntryStore.Append must return a typed error
// rather than panic when callers pass nil inputs (programmer error from a
// handler should fail the transaction cleanly, not crash the process).
func TestAuditEntryStore_Append_NilEntry(t *testing.T) {
	store := postgres.NewAuditEntryStore()
	err := store.Append(context.Background(), nil, nil)
	if !errors.Is(err, postgres.ErrAuditEntryNil) {
		t.Fatalf("nil entry: want ErrAuditEntryNil, got %v", err)
	}
}

func TestAuditEntryStore_Append_NilTx(t *testing.T) {
	store := postgres.NewAuditEntryStore()
	err := store.Append(context.Background(), nil, &entities.AuditEntry{})
	if !errors.Is(err, postgres.ErrAuditTxNil) {
		t.Fatalf("nil tx: want ErrAuditTxNil, got %v", err)
	}
}
