package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

const reasonMaxLen = 500

var controlCharPattern = regexp.MustCompile(`[\x00-\x1F\x7F]`)

func validateReason(reason string, required bool) string {
	v := strings.TrimSpace(reason)
	if required && v == "" {
		return "Reason is required."
	}
	if utf8.RuneCountInString(v) > reasonMaxLen {
		return fmt.Sprintf("Reason must be %d characters or fewer.", reasonMaxLen)
	}
	if controlCharPattern.MatchString(v) {
		return "Reason contains invalid control characters."
	}
	return ""
}

func (h *ProjectsDetailHandlers) renderTransitionForm(c fiber.Ctx, id uuid.UUID, actionPath, submitLabel, title string, required bool, reason, reasonErr string, status int) error {
	m := fiber.Map{
		"ProjectID":   id.String(),
		"ActionPath":  actionPath,
		"SubmitLabel": submitLabel,
		"Title":       title,
		"Required":    required,
		"Reason":      reason,
		"ReasonError": reasonErr,
	}
	if reasonErr != "" {
		m["Alert"] = viewmodels.InlineAlertVM{
			Severity:   "danger",
			AlertClass: "alert-danger",
			Role:       "alert",
			Icon:       "warning",
			Title:      "Couldn't submit transition",
			Message:    reasonErr,
		}
	}
	c.Status(status)
	return c.Render("pages/projects_transition_form", m, "")
}

func (h *ProjectsDetailHandlers) GetProjectPlaceOnHold(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.place_on_hold", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if _, err := h.Projects.Load(c.Context(), id); err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return err
	}
	return h.renderTransitionForm(c, id, fmt.Sprintf("/projects/%s/place-on-hold", id), "Place on hold", "Place project on hold", true, "", "", fiber.StatusOK)
}

func (h *ProjectsDetailHandlers) GetProjectResume(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.resume", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	if _, err := h.Projects.Load(c.Context(), id); err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return err
	}
	return h.renderTransitionForm(c, id, fmt.Sprintf("/projects/%s/resume", id), "Resume", "Resume project", false, "", "", fiber.StatusOK)
}

func (h *ProjectsDetailHandlers) PostProjectPlaceOnHold(c fiber.Ctx) error {
	// ADR-012: Go/Fiber transition endpoints intentionally rely on the documented CSRF exemption.
	return h.postTransition(c, true)
}

func (h *ProjectsDetailHandlers) PostProjectResume(c fiber.Ctx) error {
	// ADR-012: Go/Fiber transition endpoints intentionally rely on the documented CSRF exemption.
	return h.postTransition(c, false)
}

func (h *ProjectsDetailHandlers) postTransition(c fiber.Ctx, hold bool) error {
	action := "project.resume"
	label := "Resume"
	title := "Resume project"
	required := false
	if hold {
		action = "project.place_on_hold"
		label = "Place on hold"
		title = "Place project on hold"
		required = true
	}
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, action, uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}
	reason := c.FormValue("reason")
	if msg := validateReason(reason, required); msg != "" {
		return h.renderTransitionForm(c, id, fmt.Sprintf("/projects/%s/%s", id, map[bool]string{true: "place-on-hold", false: "resume"}[hold]), label, title, required, reason, msg, fiber.StatusUnprocessableEntity)
	}

	tx, err := h.Pool.Begin(c.Context())
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(c.Context()) }()

	project, err := postgres.LoadProjectForUpdateFrom(c.Context(), tx, id)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return err
	}
	scopes, err := postgres.LoadTradeScopesFrom(c.Context(), tx, id)
	if err != nil {
		return err
	}
	inspectors, err := postgres.LoadInspectorsFrom(c.Context(), tx, id)
	if err != nil {
		return err
	}
	beforeStateBytes, _ := json.Marshal(struct {
		Status string `json:"status"`
	}{Status: string(project.Status)})

	if hold {
		err = project.PlaceOnHold(reason)
	} else {
		err = project.Resume(reason)
	}
	if err != nil {
		if errors.Is(err, entities.ErrInvalidProjectTransition) {
			m, bErr := h.buildVMWithLoadedProjectData(c, id, project, scopes, inspectors)
			if bErr != nil {
				return bErr
			}
			m["TransitionError"] = viewmodels.InlineAlertVM{
				Severity:   "danger",
				AlertClass: "alert-danger",
				Role:       "alert",
				Icon:       "warning",
				Title:      map[bool]string{true: "Couldn't place project on hold", false: "Couldn't resume project"}[hold],
				Message:    err.Error(),
			}
			c.Status(fiber.StatusConflict)
			return c.Render("projects_detail_body", m, "")
		}
		return err
	}

	if _, err = tx.Exec(c.Context(), `UPDATE domain.project SET status = $1 WHERE id = $2`, string(project.Status), id); err != nil {
		return err
	}

	afterStateBytes, _ := json.Marshal(struct {
		Status string `json:"status"`
	}{Status: string(project.Status)})
	reasonTrimmed := strings.TrimSpace(reason)
	metaBytes, _ := json.Marshal(struct {
		Reason string `json:"reason"`
	}{Reason: reasonTrimmed})
	actionValue := enums.AuditActionProjectResumed
	if hold {
		actionValue = enums.AuditActionProjectPlacedOnHold
	}
	auditProjectID := id
	auditEntry := &entities.AuditEntry{
		ActorID:     actor.ID,
		Action:      string(actionValue),
		EntityType:  "Project",
		EntityID:    id,
		ProjectID:   &auditProjectID,
		BeforeState: beforeStateBytes,
		AfterState:  afterStateBytes,
		Metadata:    metaBytes,
	}
	if err = h.Audit.Append(c.Context(), tx, auditEntry); err != nil {
		return err
	}
	if err = tx.Commit(c.Context()); err != nil {
		return err
	}

	m, err := h.buildVM(c, id)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return err
	}
	beforeAfterJSONBytes, _ := json.Marshal(struct {
		After  json.RawMessage `json:"after"`
		Before json.RawMessage `json:"before"`
	}{
		After:  afterStateBytes,
		Before: beforeStateBytes,
	})
	now := time.Now().UTC()
	m["AuditRow"] = viewmodels.AuditRowVM{
		Action:          string(actionValue),
		ActionClass:     "badge-audit-action",
		ActorName:       actor.DisplayName,
		OccurredAt:      now.Format(time.RFC3339),
		Absolute:        now.Format(time.RFC3339),
		Relative:        "just now",
		BeforeAfterJSON: string(beforeAfterJSONBytes),
		Expanded:        false,
	}
	m["ComplianceTileOOB"] = components.NewComplianceTileArgs(&project.ComplianceScore, "Compliance", "compliance-tile")
	c.Status(fiber.StatusOK)
	return c.Render("pages/projects_transition_response", m, "")
}
