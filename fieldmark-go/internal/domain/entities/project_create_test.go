package entities_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
)

var (
	testToday    = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	testTradeID1 = uuid.New()
	testTradeID2 = uuid.New()
	testInspID   = uuid.New()
)

func strPtr(s string) *string        { return &s }
func timePtr(t time.Time) *time.Time { return &t }

func makeProject(t *testing.T, opts ...func(*projectCreateOpts)) *entities.CreatedProject {
	t.Helper()
	o := &projectCreateOpts{
		code:         "BLDG-A",
		name:         "Building A",
		description:  nil,
		startDate:    testToday,
		targetDate:   nil,
		tradeIDs:     []uuid.UUID{testTradeID1},
		inspectorIDs: []uuid.UUID{},
	}
	for _, fn := range opts {
		fn(o)
	}
	result, err := entities.CreateProject(
		o.code,
		o.name,
		o.description,
		o.startDate,
		o.targetDate,
		o.tradeIDs,
		o.inspectorIDs,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}

type projectCreateOpts struct {
	code         string
	name         string
	description  *string
	startDate    time.Time
	targetDate   *time.Time
	tradeIDs     []uuid.UUID
	inspectorIDs []uuid.UUID
}

// ── Happy path ────────────────────────────────────────────────────────────────

func TestCreateProject_ActiveStatus(t *testing.T) {
	got := makeProject(t)
	if got.Project.Status != enums.ProjectStatusActive {
		t.Errorf("status = %q; want %q", got.Project.Status, enums.ProjectStatusActive)
	}
}

func TestCreateProject_GeneratesUniqueIDs(t *testing.T) {
	r1 := makeProject(t)
	r2 := makeProject(t)
	if r1.Project.ID == uuid.Nil {
		t.Error("ID is nil")
	}
	if r1.Project.ID == r2.Project.ID {
		t.Error("two calls produced the same ID")
	}
}

func TestCreateProject_ComplianceScore100(t *testing.T) {
	got := makeProject(t)
	if got.Project.ComplianceScore != 100 {
		t.Errorf("ComplianceScore = %d; want 100", got.Project.ComplianceScore)
	}
}

func TestCreateProject_TrimsCodeAndName(t *testing.T) {
	got := makeProject(t, func(o *projectCreateOpts) {
		o.code = "  BLDG-A  "
		o.name = "  Building A  "
	})
	if got.Project.Code != "BLDG-A" {
		t.Errorf("Code = %q; want %q", got.Project.Code, "BLDG-A")
	}
	if got.Project.Name != "Building A" {
		t.Errorf("Name = %q; want %q", got.Project.Name, "Building A")
	}
}

func TestCreateProject_TradeScopes(t *testing.T) {
	got := makeProject(t, func(o *projectCreateOpts) {
		o.tradeIDs = []uuid.UUID{testTradeID1, testTradeID2}
	})
	if len(got.Scopes) != 2 {
		t.Fatalf("Scopes len = %d; want 2", len(got.Scopes))
	}
	for _, s := range got.Scopes {
		if s.ProjectID != got.Project.ID {
			t.Errorf("Scope.ProjectID = %v; want %v", s.ProjectID, got.Project.ID)
		}
	}
}

func TestCreateProject_EmptyInspectors(t *testing.T) {
	got := makeProject(t)
	if len(got.Inspectors) != 0 {
		t.Errorf("Inspectors len = %d; want 0", len(got.Inspectors))
	}
}

func TestCreateProject_Inspectors(t *testing.T) {
	got := makeProject(t, func(o *projectCreateOpts) {
		o.inspectorIDs = []uuid.UUID{testInspID}
	})
	if len(got.Inspectors) != 1 {
		t.Fatalf("Inspectors len = %d; want 1", len(got.Inspectors))
	}
	if got.Inspectors[0].UserID != testInspID {
		t.Errorf("Inspector.UserID = %v; want %v", got.Inspectors[0].UserID, testInspID)
	}
}

func TestCreateProject_NilDescription_StoresNil(t *testing.T) {
	got := makeProject(t, func(o *projectCreateOpts) { o.description = nil })
	if got.Project.Description != nil {
		t.Errorf("Description = %v; want nil", got.Project.Description)
	}
}

func TestCreateProject_WhitespaceDescription_StoresNil(t *testing.T) {
	got := makeProject(t, func(o *projectCreateOpts) { o.description = strPtr("   ") })
	if got.Project.Description != nil {
		t.Errorf("Description = %v; want nil", got.Project.Description)
	}
}

func TestCreateProject_Description_Trimmed(t *testing.T) {
	got := makeProject(t, func(o *projectCreateOpts) { o.description = strPtr("  hello  ") })
	if got.Project.Description == nil || *got.Project.Description != "hello" {
		t.Errorf("Description = %v; want %q", got.Project.Description, "hello")
	}
}

// ── Validation errors ─────────────────────────────────────────────────────────

func TestCreateProject_EmptyCode_Error(t *testing.T) {
	_, err := entities.CreateProject("", "name", nil, testToday, nil, []uuid.UUID{testTradeID1}, nil)
	if !errors.Is(err, entities.ErrInvalidArgument) {
		t.Errorf("err = %v; want ErrInvalidArgument", err)
	}
}

func TestCreateProject_WhitespaceCode_Error(t *testing.T) {
	_, err := entities.CreateProject("  ", "name", nil, testToday, nil, []uuid.UUID{testTradeID1}, nil)
	if !errors.Is(err, entities.ErrInvalidArgument) {
		t.Errorf("err = %v; want ErrInvalidArgument", err)
	}
}

func TestCreateProject_EmptyName_Error(t *testing.T) {
	_, err := entities.CreateProject("CODE", "", nil, testToday, nil, []uuid.UUID{testTradeID1}, nil)
	if !errors.Is(err, entities.ErrInvalidArgument) {
		t.Errorf("err = %v; want ErrInvalidArgument", err)
	}
}

func TestCreateProject_EmptyTradeScopes_Error(t *testing.T) {
	_, err := entities.CreateProject("CODE", "name", nil, testToday, nil, []uuid.UUID{}, nil)
	if !errors.Is(err, entities.ErrInvalidArgument) {
		t.Errorf("err = %v; want ErrInvalidArgument", err)
	}
}

func TestCreateProject_TargetBeforeStart_Error(t *testing.T) {
	target := testToday.AddDate(0, -1, 0)
	_, err := entities.CreateProject("CODE", "name", nil, testToday, &target, []uuid.UUID{testTradeID1}, nil)
	if !errors.Is(err, entities.ErrInvalidArgument) {
		t.Errorf("err = %v; want ErrInvalidArgument", err)
	}
}

func TestCreateProject_TargetEqualsStart_Succeeds(t *testing.T) {
	_, err := entities.CreateProject("CODE", "name", nil, testToday, timePtr(testToday), []uuid.UUID{testTradeID1}, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
