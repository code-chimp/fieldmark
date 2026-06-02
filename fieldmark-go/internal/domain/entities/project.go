// Package entities holds plain field-bag types for the domain aggregates.
// Behavior methods are added by the consuming stories (2.8 / 2.12 / Epic 6).
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

var ErrInvalidProjectTransition = errors.New("invalid project transition")

type invalidProjectTransitionError struct {
	message string
}

func (e invalidProjectTransitionError) Error() string {
	return e.message
}

func (e invalidProjectTransitionError) Unwrap() error {
	return ErrInvalidProjectTransition
}

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

// CanPlaceOnHold is a status-only gate for Story 2.11 affordance rendering.
func (p Project) CanPlaceOnHold() bool {
	return p.Status == enums.ProjectStatusActive
}

// CanResume is a status-only gate for Story 2.11 affordance rendering.
func (p Project) CanResume() bool {
	return p.Status == enums.ProjectStatusOnHold
}

// CanClose is a status-only gate for Story 2.11; Epic 6 adds closure checks.
func (p Project) CanClose() bool {
	return p.Status == enums.ProjectStatusActive
}

func (p *Project) PlaceOnHold(reason string) error {
	_ = reason
	if p.Status != enums.ProjectStatusActive {
		return invalidProjectTransitionError{message: "Project is already on hold"}
	}
	p.Status = enums.ProjectStatusOnHold
	return nil
}

func (p *Project) Resume(reason string) error {
	_ = reason
	if p.Status != enums.ProjectStatusOnHold {
		return invalidProjectTransitionError{message: "Project is not on hold"}
	}
	p.Status = enums.ProjectStatusActive
	return nil
}
