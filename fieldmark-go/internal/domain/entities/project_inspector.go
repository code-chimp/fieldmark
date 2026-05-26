package entities

import "github.com/google/uuid"

// ProjectInspector links a Project to an inspector user. UserID is opaque
// per ADR-012 — no FK to fiber_auth or any other auth schema.
type ProjectInspector struct {
	ProjectID uuid.UUID
	UserID    uuid.UUID
}
