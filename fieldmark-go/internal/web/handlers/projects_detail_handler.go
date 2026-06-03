// Package handlers — projects_detail_handler.go
//
// GET /projects/:id — minimal stub. Story 2.11 replaces this with the full
// Project Detail screen. Exists so the HX-Redirect from Story 2.8 lands
// somewhere meaningful.
package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"slices"
	"strings"
	"time"

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
	AuditRead postgres.AuditEntryReadStore
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

func normalizeProjectTab(tab string) string {
	switch strings.ToLower(strings.TrimSpace(tab)) {
	case "inspections", "violations", "audit":
		return strings.ToLower(strings.TrimSpace(tab))
	default:
		return "summary"
	}
}

func (h *ProjectsDetailHandlers) buildVM(c fiber.Ctx, id uuid.UUID) (fiber.Map, error) {
	project, _, scopes, inspectors, err := h.Projects.LoadWithRelations(c.Context(), id)
	if err != nil {
		return nil, err
	}
	return h.buildVMWithLoadedProjectData(c, id, project, scopes, inspectors, c.Params("tab"))
}

func (h *ProjectsDetailHandlers) buildVMWithLoadedProjectData(
	c fiber.Ctx,
	id uuid.UUID,
	project *entities.Project,
	scopes []entities.ProjectTradeScope,
	inspectors []entities.ProjectInspector,
	tabValue string,
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
	tab := normalizeProjectTab(tabValue)
	panel := "project_detail_summary_panel"
	activeTabID := "tab-summary"
	switch tab {
	case "summary":
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
	m := fiber.Map{
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
		"IsTabResponse": tabValue != "",
	}
	if activeTabID == "tab-audit" {
		auditRows, auditLoadMorePath, err := h.loadAuditPage(c, id, postgres.AuditPage{})
		if err != nil {
			return nil, err
		}
		m["AuditRows"] = auditRows
		m["AuditLoadMorePath"] = auditLoadMorePath
		m["SuppressAuditEmptyState"] = false
	}
	return m, nil
}

func relativeAuditTime(occurredAt time.Time, now time.Time) string {
	seconds := int(now.Sub(occurredAt).Seconds())
	if seconds < 0 {
		seconds = 0
	}
	if seconds < 60 {
		return "just now"
	}
	minutes := seconds / 60
	if minutes < 60 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	days := hours / 24
	if days < 30 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	months := days / 30
	if months < 12 {
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}
	years := months / 12
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

func renderAuditJSON(row postgres.AuditEntryRow) string {
	payload := map[string]any{}
	if len(row.AfterState) > 0 {
		var after any
		if json.Unmarshal(row.AfterState, &after) == nil {
			payload["after"] = after
		}
	}
	if len(row.BeforeState) > 0 {
		var before any
		if json.Unmarshal(row.BeforeState, &before) == nil {
			payload["before"] = before
		}
	}
	if len(row.Metadata) > 0 {
		var metadata any
		if json.Unmarshal(row.Metadata, &metadata) == nil {
			payload["metadata"] = metadata
		}
	}
	if len(payload) == 0 {
		return ""
	}
	b, err := marshalSortedJSON(payload)
	if err != nil {
		return ""
	}
	return string(b)
}

func marshalSortedJSON(value any) ([]byte, error) {
	switch v := value.(type) {
	case map[string]any:
		keys := slices.Sorted(maps.Keys(v))
		var buf bytes.Buffer
		buf.WriteByte('{')
		for i, key := range keys {
			if i > 0 {
				buf.WriteByte(',')
			}
			keyJSON, err := json.Marshal(key)
			if err != nil {
				return nil, err
			}
			buf.Write(keyJSON)
			buf.WriteByte(':')
			childJSON, err := marshalSortedJSON(v[key])
			if err != nil {
				return nil, err
			}
			buf.Write(childJSON)
		}
		buf.WriteByte('}')
		return buf.Bytes(), nil
	case []any:
		var buf bytes.Buffer
		buf.WriteByte('[')
		for i, child := range v {
			if i > 0 {
				buf.WriteByte(',')
			}
			childJSON, err := marshalSortedJSON(child)
			if err != nil {
				return nil, err
			}
			buf.Write(childJSON)
		}
		buf.WriteByte(']')
		return buf.Bytes(), nil
	default:
		return json.Marshal(v)
	}
}

func auditActionClass(action string) string {
	switch action {
	case "ProjectCreated", "ProjectPlacedOnHold", "ProjectResumed", "ProjectClosed",
		"InspectionScheduled", "InspectionStarted", "InspectionCompleted", "InspectionCancelled",
		"ViolationOpened", "ViolationAssigned", "ViolationVoided",
		"CorrectiveActionSubmitted", "CorrectiveActionTakenForReview",
		"CorrectiveActionApproved", "CorrectiveActionRejected":
		return "badge-audit-action"
	default:
		return "badge-unknown"
	}
}

func (h *ProjectsDetailHandlers) loadAuditPage(
	c fiber.Ctx,
	projectID uuid.UUID,
	page postgres.AuditPage,
) ([]viewmodels.AuditRowVM, string, error) {
	if h.AuditRead == nil {
		log.Printf("warn: ProjectsDetailHandlers.loadAuditPage called without AuditRead store for project %s", projectID.String())
		return nil, "", nil
	}
	result, err := h.AuditRead.ListByProject(c.Context(), projectID, page)
	if err != nil {
		return nil, "", err
	}
	now := time.Now().UTC()
	rows := make([]viewmodels.AuditRowVM, 0, len(result.Rows))
	for _, row := range result.Rows {
		absolute := row.OccurredAt.UTC().Format(time.RFC3339)
		rows = append(rows, viewmodels.AuditRowVM{
			Action:          row.Action,
			ActionClass:     auditActionClass(row.Action),
			ActorName:       row.ActorName,
			OccurredAt:      absolute,
			Absolute:        absolute,
			Relative:        relativeAuditTime(row.OccurredAt.UTC(), now),
			BeforeAfterJSON: renderAuditJSON(row),
			Expanded:        false,
		})
	}
	loadMorePath := ""
	if result.NextCursor != nil {
		loadMorePath = fmt.Sprintf(
			"/projects/%s/audit-log?before_occurred_at=%s&before_id=%s",
			projectID.String(),
			result.NextCursor.OccurredAt.UTC().Format(time.RFC3339),
			result.NextCursor.ID.String(),
		)
	}
	return rows, loadMorePath, nil
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
