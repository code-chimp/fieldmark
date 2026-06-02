// Package handlers — projects_detail_handler.go
//
// GET /projects/:id — minimal stub. Story 2.11 replaces this with the full
// Project Detail screen. Exists so the HX-Redirect from Story 2.8 lands
// somewhere meaningful.
package handlers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

// ProjectsDetailHandlers groups the stub detail handler dependencies.
type ProjectsDetailHandlers struct {
	Pool      *pgxpool.Pool
	Projects  postgres.ProjectStore
	Reference postgres.ReferenceStore
	Audit     postgres.AuditEntryStore
}

type projectSummaryVM struct {
	Code           string
	Name           string
	StartDate      string
	TargetDate     string
	Description    string
	TradeNames     []string
	InspectorNames []string
	PlaceOnHold    viewmodels.ActionButtonVM
	Resume         viewmodels.ActionButtonVM
	Close          viewmodels.ActionButtonVM
}

func projectTabs(id uuid.UUID) []components.TabSpec {
	base := fmt.Sprintf("/projects/%s/tabs", id.String())
	return []components.TabSpec{
		{ID: "tab-summary", Label: "Summary", HxGet: base + "/summary", HxTarget: "#project-detail-tab-content"},
		{ID: "tab-inspections", Label: "Inspections", HxGet: base + "/inspections", HxTarget: "#project-detail-tab-content"},
		{ID: "tab-violations", Label: "Violations", HxGet: base + "/violations", HxTarget: "#project-detail-tab-content"},
		{ID: "tab-audit", Label: "Audit", HxGet: base + "/audit", HxTarget: "#project-detail-tab-content"},
	}
}

func (h *ProjectsDetailHandlers) loadInspectorNames(c fiber.Ctx, inspectorIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	out := map[uuid.UUID]string{}
	if len(inspectorIDs) == 0 || h.Pool == nil {
		return out, nil
	}
	rows, err := h.Pool.Query(c.Context(), `
		SELECT id, COALESCE(NULLIF(display_name, ''), username)
		FROM fiber_auth.users
		WHERE id = ANY($1::uuid[])
	`, inspectorIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uuid.UUID
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[id] = name
	}
	return out, rows.Err()
}

func (h *ProjectsDetailHandlers) buildVM(c fiber.Ctx, id uuid.UUID) (fiber.Map, error) {
	project, _, scopes, inspectors, err := h.Projects.LoadWithRelations(c.Context(), id)
	if err != nil {
		return nil, err
	}
	return h.buildVMWithLoadedProjectData(c, id, project, scopes, inspectors)
}

func (h *ProjectsDetailHandlers) buildVMWithLoadedProjectData(
	c fiber.Ctx,
	id uuid.UUID,
	project *entities.Project,
	scopes []entities.ProjectTradeScope,
	inspectors []entities.ProjectInspector,
) (fiber.Map, error) {
	tradeTypes, err := h.Reference.ListTradeTypes(c.Context())
	if err != nil {
		return nil, err
	}
	tradeByID := map[uuid.UUID]entities.TradeType{}
	for _, t := range tradeTypes {
		tradeByID[t.ID] = t
	}
	tradeNames := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if t, ok := tradeByID[s.TradeTypeID]; ok {
			if t.Active {
				tradeNames = append(tradeNames, t.Name)
			} else {
				tradeNames = append(tradeNames, fmt.Sprintf("%s (inactive)", t.Name))
			}
		}
	}
	inspectorIDs := make([]uuid.UUID, 0, len(inspectors))
	for _, i := range inspectors {
		inspectorIDs = append(inspectorIDs, i.UserID)
	}
	inspectorMap, err := h.loadInspectorNames(c, inspectorIDs)
	if err != nil {
		return nil, err
	}
	inspectorNames := make([]string, 0, len(inspectorIDs))
	for _, iid := range inspectorIDs {
		if n, ok := inspectorMap[iid]; ok {
			inspectorNames = append(inspectorNames, n)
		}
	}
	actor := auth.ActorFromCtx(c)
	summary := projectSummaryVM{
		Code:           project.Code,
		Name:           project.Name,
		StartDate:      project.StartDate.Format("2006-01-02"),
		TargetDate:     "—",
		Description:    "—",
		TradeNames:     tradeNames,
		InspectorNames: inspectorNames,
		PlaceOnHold: viewmodels.ActionButtonVM{
			ID: "place-on-hold-btn", Permission: auth.Can(actor, "project.place_on_hold", uuid.Nil),
			StateAllows: project.CanPlaceOnHold(), Label: "Place on Hold",
			HxGet: fmt.Sprintf("/projects/%s/place-on-hold", id), HxTarget: "#project-action-form", HxSwap: "innerHTML", DisabledReason: "Project is already on hold",
		},
		Resume: viewmodels.ActionButtonVM{
			ID: "resume-btn", Permission: auth.Can(actor, "project.resume", uuid.Nil),
			StateAllows: project.CanResume(), Label: "Resume",
			HxGet: fmt.Sprintf("/projects/%s/resume", id), HxTarget: "#project-action-form", HxSwap: "innerHTML", DisabledReason: "Project is not on hold",
		},
		Close: viewmodels.ActionButtonVM{
			ID: "close-btn", Permission: auth.Can(actor, "project.close", uuid.Nil),
			StateAllows: project.CanClose(), Label: "Close",
			HxPost: fmt.Sprintf("/projects/%s/close", id), HxTarget: "#project-detail", HxSwap: "outerHTML", DisabledReason: "Only active projects can be closed",
		},
	}
	if project.TargetCompletionDate != nil {
		summary.TargetDate = project.TargetCompletionDate.Format("2006-01-02")
	}
	if project.Description != nil && strings.TrimSpace(*project.Description) != "" {
		summary.Description = *project.Description
	}
	activeIndex := 0
	tab := strings.ToLower(c.Params("tab"))
	panel := "project_detail_summary_panel"
	activeTabID := "tab-summary"
	switch tab {
	case "", "summary":
	case "inspections":
		activeIndex = 1
		panel = "project_detail_inspections_panel"
		activeTabID = "tab-inspections"
	case "violations":
		activeIndex = 2
		panel = "project_detail_violations_panel"
		activeTabID = "tab-violations"
	case "audit":
		activeIndex = 3
		panel = "project_detail_audit_panel"
		activeTabID = "tab-audit"
	default:
		return nil, fiber.ErrNotFound
	}
	status := viewmodels.ResolveStatusBadge("project", string(project.Status))
	return fiber.Map{
		"Title":          project.Name,
		"ProjectCode":    project.Code,
		"ProjectName":    project.Name,
		"ProjectID":      id.String(),
		"ProjectStatus":  status,
		"ComplianceTile": components.NewComplianceTileArgs(&project.ComplianceScore, "Compliance", "compliance-tile"),
		"Tabs": components.TabStripArgs{
			ID: "project-detail-tabstrip", AriaLabel: "Project Detail Tabs", Tabs: projectTabs(id), ActiveIndex: activeIndex,
		},
		"Summary":       summary,
		"Rail":          components.EntityRailArgs{ID: "violation-detail", EntityTypeLabel: "Violation", EntityLoaded: false},
		"ActiveTabID":   activeTabID,
		"PanelTemplate": panel,
		"IsTabResponse": tab != "",
	}, nil
}

// GetProjectsDetail handles GET /projects/:id.
func (h *ProjectsDetailHandlers) GetProjectsDetail(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.read", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}
	rawID := c.Params("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	m, err := h.buildVM(c, id)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		if errors.Is(err, fiber.ErrNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return err
	}
	isTab, _ := m["IsTabResponse"].(bool)
	if c.Get("HX-Request") == "true" {
		if isTab {
			tabs, ok := m["Tabs"].(components.TabStripArgs)
			if !ok {
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			m["TabOob"] = components.TabStripArgs{
				ID: "project-detail-tabstrip", AriaLabel: "Project Detail Tabs", Tabs: projectTabs(id), ActiveIndex: tabs.ActiveIndex, HxSwapOob: "outerHTML",
			}
			return c.Render("pages/projects_detail_tab_response", m, "")
		}
		return c.Render("projects_detail_body", m, "")
	}
	if isTab {
		return c.Redirect().To(fmt.Sprintf("/projects/%s", id))
	}
	return c.Render("pages/projects_detail", m)
}
