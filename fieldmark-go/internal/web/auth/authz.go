package auth

import (
	"sync"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
)

// actionRoleMap holds the action → permitted-roles mapping. Stories from
// Epic 2+ register their actions at composition time. Story 1.12 ships
// empty — Epic 1 has no live actions.
var (
	actionRoleMapMu sync.RWMutex
	actionRoleMap   = map[string]map[domain.Role]struct{}{}
)

// RegisterAction registers an action → permitted-roles mapping.
func RegisterAction(action string, roles ...domain.Role) {
	actionRoleMapMu.Lock()
	defer actionRoleMapMu.Unlock()
	set, ok := actionRoleMap[action]
	if !ok {
		set = map[domain.Role]struct{}{}
		actionRoleMap[action] = set
	}
	for _, r := range roles {
		set[r] = struct{}{}
	}
}

// Can returns true if the actor is authenticated and permitted to perform
// action (optionally scoped to entityID; uuid.Nil means "no entity scope").
// Epic 1: role-only checks (entity-scope rules deferred to Epic 2+).
func Can(actor *app.Actor, action string, entityID uuid.UUID) bool {
	if actor == nil || actor.IsAnonymous() {
		return false
	}
	actionRoleMapMu.RLock()
	permitted, ok := actionRoleMap[action]
	actionRoleMapMu.RUnlock()
	if !ok {
		return false
	}
	if _, hit := permitted[domain.Role(actor.Role)]; !hit {
		return false
	}
	return evaluateEntityScope(action, entityID)
}

// evaluateEntityScope is the single extension point for Epic 2+ entity-scope
// rules (e.g., "Site Supervisor can act on a Violation only if assigned to it").
// Today every action is role-coarse; future stories wire entity-scope evaluators here.
func evaluateEntityScope(_ string, _ uuid.UUID) bool { return true }

// resetForTests clears the map. Test-only — unexported, accessible only within this package.
func resetForTests() {
	actionRoleMapMu.Lock()
	defer actionRoleMapMu.Unlock()
	actionRoleMap = map[string]map[domain.Role]struct{}{}
}

// ResetForTests clears the action policy map for external package tests.
func ResetForTests() {
	resetForTests()
}

// RequireRole returns a middleware that 403s if the authenticated actor does
// not hold the given conceptual role. Unauthenticated actors are redirected to
// /login. Use the domain.Role* constants for the role argument.
func RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		actor := ActorFromCtx(c)
		if actor.IsAnonymous() {
			return c.Redirect().Status(fiber.StatusFound).To("/login")
		}
		if actor.Role != role {
			c.Status(fiber.StatusForbidden)
			return c.SendString("Forbidden.")
		}
		return c.Next()
	}
}
