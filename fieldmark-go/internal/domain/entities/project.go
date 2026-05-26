// Package entities holds plain field-bag types for the domain aggregates.
// Behavior methods are added by the consuming stories (2.8 / 2.12 / Epic 6).
package entities

import (
	"time"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

// Project mirrors domain.project. Date columns map to time.Time (DDL DATE);
// timestamps map to time.Time (DDL TIMESTAMPTZ). Nullable columns use *T.
type Project struct {
	ID                   uuid.UUID
	Code                 string
	Name                 string
	Description          *string
	Status               enums.ProjectStatus
	StartDate            time.Time
	TargetCompletionDate *time.Time
	ActualClosedAt       *time.Time
	ComplianceScore      int
	CreatedAt            time.Time
	UpdatedAt            time.Time
}
