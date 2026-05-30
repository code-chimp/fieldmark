// Package handlers — projects_detail_handler.go
//
// GET /projects/:id — minimal stub. Story 2.11 replaces this with the full
// Project Detail screen. Exists so the HX-Redirect from Story 2.8 lands
// somewhere meaningful.
package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
)

// ProjectsDetailHandlers groups the stub detail handler dependencies.
type ProjectsDetailHandlers struct {
	Projects postgres.ProjectStore
}

// GetProjectsDetail handles GET /projects/:id (stub).
func (h *ProjectsDetailHandlers) GetProjectsDetail(c fiber.Ctx) error {
	rawID := c.Params("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		return c.SendStatus(fiber.StatusNotFound)
	}

	project, err := h.Projects.Load(c.Context(), id)
	if err != nil {
		if errors.Is(err, postgres.ErrProjectNotFound) {
			return c.SendStatus(fiber.StatusNotFound)
		}
		return err
	}

	return c.Render("pages/projects_detail", fiber.Map{
		"Title":       project.Name,
		"ProjectName": project.Name,
	})
}
