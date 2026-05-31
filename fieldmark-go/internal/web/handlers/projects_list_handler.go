// Package handlers — projects_list_handler.go
//
// GET /projects — projects list page with AG Grid SSRM panel.
// See docs/reference/ag-grid-ssrm-contract.md
package handlers

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

// ProjectsListHandlers groups dependencies for the list page.
type ProjectsListHandlers struct{}

// GetProjectsList handles GET /projects.
func (h *ProjectsListHandlers) GetProjectsList(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "project.read", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}

	canCreate := auth.Can(actor, "project.create", uuid.Nil)

	theme, next := themeEntries(c)
	if actor == nil {
		actor = app.Anonymous()
	}
	role := domain.Role(actor.Role)

	return c.Render("pages/projects_index", fiber.Map{
		"Title":           "Projects",
		"FmTheme":         theme,
		"FmThemeNext":     next,
		"FmThemeResolved": theme,
		"Actor":           actor,
		"RoleLabel":       role.Label(),
		"RoleBadgeToken":  role.BadgeToken(),
		"FullName":        actor.DisplayName,
		"Initials":        viewmodels.Initials(actor.DisplayName, actor.Username),
		"CanCreate":       canCreate,
	})
}
