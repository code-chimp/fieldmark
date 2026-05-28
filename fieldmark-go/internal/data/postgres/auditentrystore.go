package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
)

// ErrAuditEntryNil and ErrAuditTxNil are returned by AuditEntryStore.Append
// when a caller passes a nil pointer. Surfaced as errors rather than a panic
// so a programmer error in a handler fails the transaction cleanly through
// the canonical request flow's error path, never a process crash.
var (
	ErrAuditEntryNil = errors.New("auditentrystore: entry must not be nil")
	ErrAuditTxNil    = errors.New("auditentrystore: tx must not be nil")
)

// AuditEntryStore is the narrow append-only interface for domain.audit_entry.
// The helper does NOT open or commit a transaction — callers own the
// transaction lifecycle per the canonical request flow
// (architecture.md §Process Patterns). FR39 requires the AuditEntry write to
// share the transaction of the surrounding mutation.
type AuditEntryStore interface {
	// Append inserts the entry using the supplied tx. The entry's ID and
	// OccurredAt are server-assigned on insert and populated on the passed
	// entry pointer on success — ID is pre-filled if uuid.Nil because the
	// DDL has no DEFAULT on id; OccurredAt is read back from the
	// RETURNING clause.
	Append(ctx context.Context, tx pgx.Tx, entry *entities.AuditEntry) error
}

type auditEntryStorePg struct{}

// NewAuditEntryStore returns the stateless pgx-backed AuditEntryStore.
// No pgxpool here — callers thread their own transaction in via Append.
func NewAuditEntryStore() AuditEntryStore { return &auditEntryStorePg{} }

const auditEntryInsertSQL = `
	INSERT INTO domain.audit_entry (
		id, actor_id, action, entity_type, entity_id, project_id,
		before_state, after_state, metadata
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, occurred_at
`

func (s *auditEntryStorePg) Append(
	ctx context.Context,
	tx pgx.Tx,
	entry *entities.AuditEntry,
) error {
	if entry == nil {
		return ErrAuditEntryNil
	}
	if tx == nil {
		return ErrAuditTxNil
	}
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	return tx.QueryRow(
		ctx,
		auditEntryInsertSQL,
		entry.ID,
		entry.ActorID,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		entry.ProjectID,
		entry.BeforeState,
		entry.AfterState,
		entry.Metadata,
	).Scan(&entry.ID, &entry.OccurredAt)
}
