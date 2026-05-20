package auth

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/app"
)

// localsKey is the c.Locals() key under which the request's *app.Actor
// is stored. Use ActorFromCtx to read it; never type-assert directly.
const localsKey = "user"

// cookieName / headerName are the two request-borne identifier carriers.
// The cookie is set by Story 1.11's /login user-switcher; the header is
// for scripted tests and ad-hoc curl flows. envVar is the deployment-
// fixed fallback (set in docker-compose or .env for parity scenarios).
const (
	cookieName = "X-FieldMark-Actor"
	headerName = "X-FieldMark-Actor"
	envVar     = "FIELDMARK_STUB_ACTOR"
)

// CookieName returns the cookie name carrying the resolved actor username.
// Exposed for the login/logout handlers; do not use elsewhere.
func CookieName() string { return cookieName }

// StubAuthMiddleware returns a Fiber middleware that hydrates an *app.Actor
// onto c.Locals(localsKey) for every request. Resolution order: cookie,
// header, env var, anonymous. Lookup failure (DB error or miss) falls
// back to anonymous; the application remains navigable so the developer
// can see logs and fix the auth store. ADR-012 stub posture: this is
// intentional and not a production-grade pattern.
func StubAuthMiddleware(pool *pgxpool.Pool) fiber.Handler {
	// Warn at startup if a deployment-fixed identity is configured — every
	// request without a cookie or header will resolve to this principal, so
	// a mis-set env var causes all users to share one identity silently.
	if v := strings.TrimSpace(os.Getenv(envVar)); v != "" {
		log.Printf("auth: %s=%q is set — requests with no cookie/header will resolve as %q", envVar, v, v)
	}

	return func(c fiber.Ctx) error {
		username := resolveUsername(c)
		if username == "" || username == "anonymous" {
			c.Locals(localsKey, app.Anonymous())
			return c.Next()
		}
		actor, err := lookupByUsername(c.Context(), pool, username)
		if err != nil {
			log.Printf("auth: lookup error for %q: %v (binding anonymous)", username, err)
			c.Locals(localsKey, app.Anonymous())
			return c.Next()
		}
		if actor == nil {
			c.Locals(localsKey, app.Anonymous())
			return c.Next()
		}
		c.Locals(localsKey, actor)
		return c.Next()
	}
}

// RequireAuth returns a middleware that 302-redirects unauthenticated
// requests to /login. NOT applied to any route in Story 1.9; Story 1.11
// mounts it on business routes once /login exists.
func RequireAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		actor := ActorFromCtx(c)
		if actor.IsAnonymous() {
			return c.Redirect().Status(fiber.StatusFound).To("/login")
		}
		return c.Next()
	}
}

// ActorFromCtx reads the hydrated *app.Actor from c.Locals. Returns
// app.Anonymous() if the middleware did not run or stored an unexpected
// type (defensive — never panic on a missing or wrong-typed locals).
func ActorFromCtx(c fiber.Ctx) *app.Actor {
	v := c.Locals(localsKey)
	if a, ok := v.(*app.Actor); ok && a != nil {
		return a
	}
	return app.Anonymous()
}

func resolveUsername(c fiber.Ctx) string {
	if v := strings.TrimSpace(c.Cookies(cookieName)); v != "" {
		return v
	}
	if v := strings.TrimSpace(c.Get(headerName)); v != "" {
		return v
	}
	return strings.TrimSpace(os.Getenv(envVar))
}
