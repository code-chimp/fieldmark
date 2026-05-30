// Package handlers — projects_create_handler.go
//
// GET  /projects/new → GetProjectsNew
// POST /projects/    → PostProjectsCreate
//
// See docs/reference/project-create-form-contract.md for the full contract.
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

func pluralS(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

var codePattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9-]*$`)

// ProjectsCreateHandlers groups the project-create handler dependencies.
type ProjectsCreateHandlers struct {
	Pool      *pgxpool.Pool
	Reference ReferenceStore
	Projects  postgres.ProjectStore
	Audit     postgres.AuditEntryStore
}

// inspectorUser is a lightweight summary for the inspector select option.
type inspectorUser struct {
	ID          uuid.UUID
	Username    string
	DisplayName string
}

// loadInspectorUsers queries fiber_auth for users with the INSPECTOR role.
// Parity note: fiber_auth.users has no active/lockout column (ADR-012 stub
// posture — the real-auth epic will add per-user lockout). All role members are
// treated as active. Django adds is_active=True; .NET filters LockoutEnd; Go
// cannot distinguish until the real-auth epic lands. Documented per AC9.
func loadInspectorUsers(ctx context.Context, pool *pgxpool.Pool) ([]inspectorUser, error) {
	rows, err := pool.Query(ctx, `
		SELECT u.id, u.username, u.display_name
		  FROM fiber_auth.users u
		  JOIN fiber_auth.user_roles ur ON u.id = ur.user_id
		 WHERE ur.role = $1
		 ORDER BY u.display_name ASC`,
		string(domain.RoleInspector),
	)
	if err != nil {
		return nil, fmt.Errorf("projects: load inspectors: %w", err)
	}
	defer rows.Close()

	var out []inspectorUser
	for rows.Next() {
		var u inspectorUser
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName); err != nil {
			return nil, fmt.Errorf("projects: scan inspector: %w", err)
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (h *ProjectsCreateHandlers) baseCreateMap(c fiber.Ctx) fiber.Map {
	m := fiber.Map{}
	current := "system"
	v := c.Cookies("fm_theme", "system")
	if v == "light" || v == "dark" {
		current = v
	}
	m["FmTheme"] = current
	m["FmThemeNext"] = map[string]string{"system": "light", "light": "dark", "dark": "system"}[current]
	m["FmThemeResolved"] = current
	actor := auth.ActorFromCtx(c)
	if actor == nil {
		actor = app.Anonymous()
	}
	m["Actor"] = actor
	return m
}

// GetProjectsNew handles GET /projects/new — renders the empty create form.
func (h *ProjectsCreateHandlers) GetProjectsNew(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.create", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	tradeTypes, err := h.Reference.ListTradeTypes(c.Context())
	if err != nil {
		slog.Error("projects: list trade types", "err", err)
		return err
	}
	inspectors, err := loadInspectorUsers(c.Context(), h.Pool)
	if err != nil {
		slog.Error("projects: load inspectors", "err", err)
		return err
	}

	m := h.baseCreateMap(c)
	m["Title"] = "Create Project"
	m["TradeTypes"] = tradeTypes
	m["Inspectors"] = inspectors
	m["Values"] = map[string]string{}
	m["SelectedTradeIDs"] = map[string]bool{}
	m["SelectedInspIDs"] = map[string]bool{}
	m["Errors"] = map[string]string{}
	m["ErrorCount"] = 0
	m["Alert"] = nil

	return c.Render("pages/projects_create", m)
}

// projectCreatedAfterState is the JSON snapshot written to audit_entry.after_state.
// Fields declared in alphabetical order so encoding/json emits them in that order.
type projectCreatedAfterState struct {
	Code                 string   `json:"code"`
	ComplianceScore      int      `json:"compliance_score"`
	Description          *string  `json:"description"`
	InspectorIDs         []string `json:"inspector_ids"`
	Name                 string   `json:"name"`
	StartDate            string   `json:"start_date"`
	Status               string   `json:"status"`
	TargetCompletionDate *string  `json:"target_completion_date"`
	TradeScopeIDs        []string `json:"trade_scope_ids"`
}

// PostProjectsCreate handles POST /projects/ — validates input, creates the project.
func (h *ProjectsCreateHandlers) PostProjectsCreate(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.create", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	// --- Parse raw form values (canonical snake_case names per contract doc) ---
	rawCode := strings.TrimSpace(c.FormValue("code"))
	rawName := strings.TrimSpace(c.FormValue("name"))
	rawDescription := strings.TrimSpace(c.FormValue("description"))
	rawStartDate := c.FormValue("start_date")
	rawTargetDate := c.FormValue("target_completion_date")
	rawTradeIDs := c.Request().PostArgs().PeekMulti("trade_scope_ids")
	rawInspIDs := c.Request().PostArgs().PeekMulti("inspector_ids")

	tradeIDStrs := make([]string, len(rawTradeIDs))
	for i, b := range rawTradeIDs {
		tradeIDStrs[i] = string(b)
	}
	inspIDStrs := make([]string, len(rawInspIDs))
	for i, b := range rawInspIDs {
		inspIDStrs[i] = string(b)
	}

	errs := map[string]string{}

	// --- Validate code ---
	if rawCode == "" {
		errs["code"] = "Code is required."
	} else if len(rawCode) > 32 {
		errs["code"] = "Code must be 32 characters or fewer."
	} else if !codePattern.MatchString(rawCode) {
		if strings.HasPrefix(rawCode, "-") {
			errs["code"] = "Code must start with a letter or digit."
		} else {
			errs["code"] = "Code must contain only uppercase letters, digits, and hyphens."
		}
	}

	// --- Validate name ---
	if rawName == "" {
		errs["name"] = "Name is required."
	} else if len(rawName) > 200 {
		errs["name"] = "Name must be 200 characters or fewer."
	}

	// --- Validate description ---
	var description *string
	if rawDescription != "" {
		if len(rawDescription) > 10000 {
			errs["description"] = "Description must be 10,000 characters or fewer."
		} else {
			description = &rawDescription
		}
	}

	// --- Validate start_date ---
	var startDate time.Time
	if rawStartDate == "" {
		errs["start_date"] = "Start date is required."
	} else {
		var parseErr error
		startDate, parseErr = time.Parse("2006-01-02", rawStartDate)
		if parseErr != nil {
			errs["start_date"] = "Start date must be a valid date (YYYY-MM-DD)."
		}
	}

	// --- Validate target_completion_date ---
	var targetDate *time.Time
	if rawTargetDate != "" {
		td, parseErr := time.Parse("2006-01-02", rawTargetDate)
		if parseErr != nil {
			errs["target_completion_date"] = "Target completion date must be a valid date."
		} else if _, startHasErr := errs["start_date"]; !startHasErr && td.Before(startDate) {
			errs["target_completion_date"] =
				"Target completion date must be on or after the start date."
		} else {
			targetDate = &td
		}
	}

	// --- Validate trade_scope_ids ---
	tradeScopeIDs := make([]uuid.UUID, 0, len(tradeIDStrs))
	if len(tradeIDStrs) == 0 {
		errs["trade_scope_ids"] = "At least one trade scope is required."
	} else {
		for _, s := range tradeIDStrs {
			id, err := uuid.Parse(s)
			if err != nil {
				errs["trade_scope_ids"] = "One or more selected trade types are no longer available. Please reselect."
				break
			}
			tradeScopeIDs = append(tradeScopeIDs, id)
		}
	}

	// --- Validate inspector_ids ---
	inspectorIDs := make([]uuid.UUID, 0, len(inspIDStrs))
	for _, s := range inspIDStrs {
		id, err := uuid.Parse(s)
		if err != nil {
			errs["inspector_ids"] = "One or more selected inspectors are no longer available. Please reselect."
			break
		}
		inspectorIDs = append(inspectorIDs, id)
	}

	// Deduplicate — duplicate submissions are redundant and would hit composite-PK
	// 23505 on project_trade_scope / project_inspector if not removed here.
	tradeScopeIDs = deduplicateUUIDs(tradeScopeIDs)
	inspectorIDs = deduplicateUUIDs(inspectorIDs)

	if len(errs) > 0 {
		return h.render422(c, rawCode, rawName, rawDescription, rawStartDate, rawTargetDate,
			tradeIDStrs, inspIDStrs, errs)
	}

	// --- Begin transaction ---
	tx, err := h.Pool.Begin(c.Context())
	if err != nil {
		return fmt.Errorf("projects: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(c.Context()) }()

	// --- Validate trade IDs against DB ---
	activeTrades, err := loadActiveTradeScopeIDs(c.Context(), tx)
	if err != nil {
		return err
	}
	for _, id := range tradeScopeIDs {
		if !activeTrades[id] {
			errs["trade_scope_ids"] =
				"One or more selected trade types are no longer available. Please reselect."
		}
	}

	// --- Validate inspector IDs against fiber_auth (inside the transaction) ---
	// Using tx instead of h.Pool puts the read inside the same MVCC snapshot as
	// the writes. fiber_auth and domain are both in the same Postgres instance, so
	// the transaction can span both schemas.
	if len(inspectorIDs) > 0 {
		validInspectors, err := loadValidInspectorIDs(c.Context(), tx)
		if err != nil {
			return err
		}
		for _, id := range inspectorIDs {
			if !validInspectors[id] {
				errs["inspector_ids"] =
					"One or more selected inspectors are no longer available. Please reselect."
				break
			}
		}
	}

	if len(errs) > 0 {
		_ = tx.Rollback(c.Context())
		return h.render422(c, rawCode, rawName, rawDescription, rawStartDate, rawTargetDate,
			tradeIDStrs, inspIDStrs, errs)
	}

	// --- Call entity method ---
	created, domainErr := entities.CreateProject(
		rawCode, rawName, description, startDate, targetDate, tradeScopeIDs, inspectorIDs,
	)
	if domainErr != nil {
		return fmt.Errorf("projects: entity create: %w", domainErr)
	}

	// --- Persist ---
	if err = h.Projects.CreateInTx(c.Context(), tx, created); err != nil {
		var pgErr *pgconn.PgError
		// Only surface the code-collision message when the violating constraint is
		// exactly the project.code UNIQUE constraint. Other 23505 violations (e.g.
		// a duplicate composite PK on project_trade_scope in a retry) should not
		// be misreported as a code collision.
		if errors.As(err, &pgErr) && pgErr.Code == "23505" &&
			pgErr.ConstraintName == "project_code_key" {
			_ = tx.Rollback(c.Context())
			errs["code"] = "A project with this code already exists."
			return h.render422(c, rawCode, rawName, rawDescription, rawStartDate, rawTargetDate,
				tradeIDStrs, inspIDStrs, errs)
		}
		return err
	}

	// --- Audit entry ---
	sortedTradeIDs := uuidSliceToStrings(tradeScopeIDs)
	sort.Strings(sortedTradeIDs)
	sortedInspIDs := uuidSliceToStrings(inspectorIDs)
	sort.Strings(sortedInspIDs)

	var targetDateStr *string
	if targetDate != nil {
		s := targetDate.Format("2006-01-02")
		targetDateStr = &s
	}

	afterStateBytes, marshalErr := json.Marshal(projectCreatedAfterState{
		Code:                 created.Project.Code,
		ComplianceScore:      100,
		Description:          created.Project.Description,
		InspectorIDs:         sortedInspIDs,
		Name:                 created.Project.Name,
		StartDate:            startDate.Format("2006-01-02"),
		Status:               string(enums.ProjectStatusActive),
		TargetCompletionDate: targetDateStr,
		TradeScopeIDs:        sortedTradeIDs,
	})
	if marshalErr != nil {
		return fmt.Errorf("projects: marshal after_state: %w", marshalErr)
	}

	projectID := created.Project.ID
	auditEntry := &entities.AuditEntry{
		ActorID:    actor.ID,
		Action:     string(enums.AuditActionProjectCreated),
		EntityType: "Project",
		EntityID:   projectID,
		ProjectID:  &projectID,
		AfterState: json.RawMessage(afterStateBytes),
	}
	if err = h.Audit.Append(c.Context(), tx, auditEntry); err != nil {
		return fmt.Errorf("projects: audit append: %w", err)
	}

	if err = tx.Commit(c.Context()); err != nil {
		return fmt.Errorf("projects: commit: %w", err)
	}

	// --- Respond ---
	redirectURL := fmt.Sprintf("/projects/%s", created.Project.ID)
	isHtmx := c.Get("HX-Request") == "true"
	if isHtmx {
		c.Set("HX-Redirect", redirectURL)
		return c.SendStatus(fiber.StatusOK)
	}
	return c.Redirect().Status(fiber.StatusSeeOther).To(redirectURL)
}

func (h *ProjectsCreateHandlers) render422(
	c fiber.Ctx,
	code, name, description, startDate, targetDate string,
	tradeIDStrs, inspIDStrs []string,
	errs map[string]string,
) error {
	tradeTypes, err := h.Reference.ListTradeTypes(c.Context())
	if err != nil {
		slog.Error("projects: list trade types on 422", "err", err)
		return err
	}
	inspectors, err := loadInspectorUsers(c.Context(), h.Pool)
	if err != nil {
		slog.Error("projects: load inspectors on 422", "err", err)
		return err
	}

	selectedTrades := strSliceToSet(tradeIDStrs)
	selectedInspectors := strSliceToSet(inspIDStrs)

	errCount := len(errs)
	alertMsg := fmt.Sprintf("%d error%s must be resolved before this project can be created.",
		errCount, pluralS(errCount))
	alertVM := viewmodels.InlineAlertVM{
		Severity:   "danger",
		AlertClass: "alert-danger",
		Role:       "alert",
		Icon:       "warning",
		Title:      "Couldn't create the project",
		Message:    alertMsg,
	}

	m := fiber.Map{
		"TradeTypes":       tradeTypes,
		"Inspectors":       inspectors,
		"Values":           map[string]string{"code": code, "name": name, "description": description, "start_date": startDate, "target_completion_date": targetDate},
		"SelectedTradeIDs": selectedTrades,
		"SelectedInspIDs":  selectedInspectors,
		"Errors":           errs,
		"ErrorCount":       errCount,
		"Alert":            alertVM,
	}

	c.Status(fiber.StatusUnprocessableEntity)
	return c.Render("project_create_form", m, "")
}

func loadActiveTradeScopeIDs(ctx context.Context, q pgx.Tx) (map[uuid.UUID]bool, error) {
	rows, err := q.Query(ctx, `SELECT id FROM domain.trade_type WHERE active = true`)
	if err != nil {
		return nil, fmt.Errorf("projects: load active trades: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("projects: scan trade id: %w", err)
		}
		out[id] = true
	}
	return out, rows.Err()
}

// loadValidInspectorIDs accepts a postgres.Querier so callers can pass either
// a *pgxpool.Pool (for display rendering) or a pgx.Tx (for within-transaction
// validation). When called from within a transaction the read participates in
// the same MVCC snapshot as the writes, closing the TOCTOU window.
func loadValidInspectorIDs(ctx context.Context, q postgres.Querier) (map[uuid.UUID]bool, error) {
	rows, err := q.Query(ctx, `
		SELECT u.id FROM fiber_auth.users u
		JOIN fiber_auth.user_roles ur ON u.id = ur.user_id
		WHERE ur.role = $1`,
		string(domain.RoleInspector),
	)
	if err != nil {
		return nil, fmt.Errorf("projects: load inspector ids: %w", err)
	}
	defer rows.Close()

	out := make(map[uuid.UUID]bool)
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("projects: scan inspector id: %w", err)
		}
		out[id] = true
	}
	return out, rows.Err()
}

// deduplicateUUIDs removes duplicate UUIDs preserving first-occurrence order.
func deduplicateUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	out := ids[:0:0]
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			out = append(out, id)
		}
	}
	return out
}

func strSliceToSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, s := range items {
		m[s] = true
	}
	return m
}

func uuidSliceToStrings(ids []uuid.UUID) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	return out
}
