package auth

import "github.com/gofiber/fiber/v3"

// RequireRole returns a middleware that 403s if the authenticated actor does
// not hold the given conceptual role. Unauthenticated actors are redirected to
// /login instead (same as RequireAuth). The full authz.Can primitive with
// entity-scope rules lands in Story 1.12; this is the minimal role-gate needed
// for AC #6 integration-test probes in Story 1.11.
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
