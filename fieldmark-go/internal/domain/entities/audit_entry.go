package entities

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditEntry mirrors domain.audit_entry. Write-once value object; behavior
// methods are intentionally absent (Story 2.2). The Action field is a plain
// string rather than the enums.AuditAction type so unrecognized DB values
// surface to the caller without enum coercion — Story 2.13's audit-log read
// path can decide how to surface the divergence.
//
// JSONB columns use json.RawMessage so the storage layer is opaque about
// payload shape. Nil means absence-of-payload (semantically distinct from
// `'null'::jsonb`); the helper must preserve the distinction.
type AuditEntry struct {
	ID          uuid.UUID
	OccurredAt  time.Time
	ActorID     uuid.UUID
	Action      string
	EntityType  string
	EntityID    uuid.UUID
	ProjectID   *uuid.UUID
	BeforeState json.RawMessage
	AfterState  json.RawMessage
	Metadata    json.RawMessage
}
