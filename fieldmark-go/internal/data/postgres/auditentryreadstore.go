package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const AuditPageSize = 100

type AuditPage struct {
	BeforeOccurredAt *time.Time
	BeforeID         *uuid.UUID
}

type AuditCursor struct {
	OccurredAt time.Time
	ID         uuid.UUID
}

type AuditEntryRow struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	ActorName   string
	Action      string
	BeforeState json.RawMessage
	AfterState  json.RawMessage
	Metadata    json.RawMessage
}

type AuditPageResult struct {
	Rows       []AuditEntryRow
	NextCursor *AuditCursor
}

type AuditEntryReadStore interface {
	ListByProject(context.Context, uuid.UUID, AuditPage) (AuditPageResult, error)
}

type auditEntryReadStorePg struct {
	pool *pgxpool.Pool
}

func NewAuditEntryReadStore(pool *pgxpool.Pool) AuditEntryReadStore {
	return &auditEntryReadStorePg{pool: pool}
}

func (s *auditEntryReadStorePg) ListByProject(
	ctx context.Context,
	projectID uuid.UUID,
	page AuditPage,
) (AuditPageResult, error) {
	const baseSQL = `
		SELECT
			a.id,
			a.occurred_at,
			COALESCE(NULLIF(BTRIM(u.display_name), ''), NULLIF(BTRIM(u.username), ''), '') AS actor_name,
			a.action,
			a.before_state,
			a.after_state,
			a.metadata
		FROM domain.audit_entry a
		LEFT JOIN fiber_auth.users u ON u.id = a.actor_id
		WHERE a.project_id = $1
	`
	sql := baseSQL
	args := []any{projectID}
	if page.BeforeOccurredAt != nil && page.BeforeID != nil {
		sql += ` AND (a.occurred_at, a.id) < ($2, $3)`
		args = append(args, *page.BeforeOccurredAt, *page.BeforeID)
	}
	sql += ` ORDER BY a.occurred_at DESC, a.id DESC LIMIT 101`

	rows, err := s.pool.Query(ctx, sql, args...)
	if err != nil {
		return AuditPageResult{}, err
	}
	defer rows.Close()

	var out []AuditEntryRow
	for rows.Next() {
		var row AuditEntryRow
		if err := rows.Scan(
			&row.ID,
			&row.OccurredAt,
			&row.ActorName,
			&row.Action,
			&row.BeforeState,
			&row.AfterState,
			&row.Metadata,
		); err != nil {
			return AuditPageResult{}, err
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return AuditPageResult{}, err
	}

	result := AuditPageResult{Rows: out}
	if len(out) > AuditPageSize {
		last := out[AuditPageSize-1]
		result.Rows = out[:AuditPageSize]
		result.NextCursor = &AuditCursor{OccurredAt: last.OccurredAt, ID: last.ID}
	}
	return result, nil
}
