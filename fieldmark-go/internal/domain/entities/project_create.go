// Package entities — project_create.go
//
// CreateProject is the domain factory for the project aggregate. It produces
// the Project and its join-table rows so the handler can persist all four
// row-sets in one transaction without reconstructing IDs.
//
// See docs/reference/project-create-form-contract.md for the form contract.
package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

// ErrInvalidArgument is the sentinel for domain-invariant violations in
// CreateProject. Callers use errors.Is to check; errors.As to inspect the
// wrapped message.
var ErrInvalidArgument = errors.New("invalid argument")

// CreatedProject groups the new Project and its join-table rows so the handler
// can persist all four row-sets without reconstructing IDs from scratch.
type CreatedProject struct {
	Project    *Project
	Scopes     []ProjectTradeScope
	Inspectors []ProjectInspector
}

// CreateProject creates a new Project with its join-table rows.
//
// Returns ErrInvalidArgument (wrapped) when domain invariants are violated.
// Request-level validation (lengths, allowlists, CSRF) is the handler's job;
// this function enforces domain invariants only.
func CreateProject(
	code string,
	name string,
	description *string,
	startDate time.Time,
	targetCompletionDate *time.Time,
	tradeScopeIDs []uuid.UUID,
	inspectorIDs []uuid.UUID,
) (*CreatedProject, error) {
	code = strings.TrimSpace(code)
	name = strings.TrimSpace(name)

	if code == "" {
		return nil, fmt.Errorf("%w: code is required", ErrInvalidArgument)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidArgument)
	}
	if len(tradeScopeIDs) == 0 {
		return nil, fmt.Errorf("%w: at least one trade scope is required", ErrInvalidArgument)
	}
	if targetCompletionDate != nil && targetCompletionDate.Before(startDate) {
		return nil, fmt.Errorf(
			"%w: target_completion_date must be on or after start_date",
			ErrInvalidArgument,
		)
	}

	var desc *string
	if description != nil {
		trimmed := strings.TrimSpace(*description)
		if trimmed != "" {
			desc = &trimmed
		}
	}

	projectID := uuid.New()
	project := &Project{
		ID:                   projectID,
		Code:                 code,
		Name:                 name,
		Description:          desc,
		Status:               enums.ProjectStatusActive,
		StartDate:            startDate,
		TargetCompletionDate: targetCompletionDate,
		ComplianceScore:      100,
	}

	scopes := make([]ProjectTradeScope, len(tradeScopeIDs))
	for i, tid := range tradeScopeIDs {
		scopes[i] = ProjectTradeScope{ProjectID: projectID, TradeTypeID: tid}
	}

	inspectors := make([]ProjectInspector, len(inspectorIDs))
	for i, uid := range inspectorIDs {
		inspectors[i] = ProjectInspector{ProjectID: projectID, UserID: uid}
	}

	return &CreatedProject{
		Project:    project,
		Scopes:     scopes,
		Inspectors: inspectors,
	}, nil
}
