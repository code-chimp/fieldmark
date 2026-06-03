package handlers

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
)

func (h *ProjectsDetailHandlers) GetProjectAuditLog(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.read", uuid.Nil) {
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

	page := postgres.AuditPage{}
	rawBeforeOccurredAt := c.Query("before_occurred_at")
	rawBeforeID := c.Query("before_id")
	if rawBeforeOccurredAt != "" || rawBeforeID != "" {
		if rawBeforeOccurredAt == "" || rawBeforeID == "" {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid cursor.")
		}
		beforeOccurredAt, err := time.Parse(time.RFC3339, rawBeforeOccurredAt)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid cursor.")
		}
		beforeID, err := uuid.Parse(rawBeforeID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid cursor.")
		}
		page.BeforeOccurredAt = &beforeOccurredAt
		page.BeforeID = &beforeID
	}

	auditRows, auditLoadMorePath, err := h.loadAuditPage(c, id, page)
	if err != nil {
		return err
	}
	return c.Render("pages/projects_audit_log_items", fiber.Map{
		"AuditRows":         auditRows,
		"AuditLoadMorePath": auditLoadMorePath,
	}, "")
}
