package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

var (
	fmThemeAllowed = map[string]bool{"system": true, "light": true, "dark": true}
	fmThemeCycle   = map[string]string{"system": "light", "light": "dark", "dark": "system"}
)

func resolveFmTheme(c fiber.Ctx) string {
	v := c.Cookies("fm_theme", "system")
	if !fmThemeAllowed[v] {
		return "system"
	}
	return v
}

// baseMap builds the common view-model fields present on every full-page render:
// theme values and the resolved actor for layout chrome (sign in / sign out).
func baseMap(c fiber.Ctx) fiber.Map {
	current := resolveFmTheme(c)
	actor := auth.ActorFromCtx(c)
	if actor == nil {
		actor = app.Anonymous()
	}
	return fiber.Map{
		"FmTheme":         current,
		"FmThemeNext":     fmThemeCycle[current],
		"FmThemeResolved": current,
		"Actor":           actor,
	}
}

func renderHomeContext(c fiber.Ctx) fiber.Map {
	m := baseMap(c)
	m["Title"] = "FieldMark"

	actor := auth.ActorFromCtx(c)
	if actor == nil {
		actor = app.Anonymous()
	}

	role := domain.Role(actor.Role)
	m["RoleLabel"] = role.Label()
	badgeToken := role.BadgeToken()
	// Only warn when a non-empty role string is unrecognised. An empty actor.Role
	// (anonymous user or user with no role assigned) is expected and must not log.
	if badgeToken == "unknown" && actor.Role != "" {
		slog.Warn("unknown role badge token", "role", actor.Role)
	}
	m["RoleBadgeToken"] = badgeToken
	m["FullName"] = actor.DisplayName
	m["Initials"] = viewmodels.Initials(actor.DisplayName, actor.Username)

	return m
}

func buildApp(pool *pgxpool.Pool) *fiber.App {
	// html.New walks internal/web/templates/ and loads all *.html files.
	// The Layout option wraps every c.Render() call in layouts/base.html
	// unless the handler passes an empty layout string explicitly.
	engine := html.New("./internal/web/templates", ".html")
	engine.AddFunc("noescape", func(s string) string { return s })

	app := fiber.New(fiber.Config{
		Views:       engine,
		ViewsLayout: "base",
	})

	app.Use(logger.New())

	// StubAuthMiddleware runs on every request when a pool is available.
	// Omitted on the -dump-routes path (pool is nil) so route enumeration
	// never requires a live database — preserving the Story 1.3 invariant.
	if pool != nil {
		app.Use(auth.StubAuthMiddleware(pool))
	}

	// Root-level static files served before the /static prefix mount.
	app.Get("/robots.txt", func(c fiber.Ctx) error {
		return c.SendFile("./internal/web/static/robots.txt")
	})
	app.Get("/.well-known/security.txt", func(c fiber.Ctx) error {
		return c.SendFile("./internal/web/static/.well-known/security.txt")
	})

	// Static assets: /static/** → internal/web/static/
	app.Use("/static", static.New("./internal/web/static"))

	return app
}

func registerRoutes(app *fiber.App, pool *pgxpool.Pool) {
	// Auth routes — no RequireAuth; these are the public entry points.
	if pool != nil {
		h := &handlers.LoginHandlers{Pool: pool}
		app.Get("/login", h.GetLogin)
		app.Post("/login", h.PostLogin)
		app.Get("/logout", h.PostLogout)
		app.Post("/logout", h.PostLogout)
	} else {
		// dump-routes path: register stub handlers so the route inventory is complete.
		noop := func(c fiber.Ctx) error { return nil }
		app.Get("/login", noop)
		app.Post("/login", noop)
		app.Get("/logout", noop)
		app.Post("/logout", noop)
	}

	// Business routes — protected by RequireAuth.
	app.Get("/", auth.RequireAuth(), func(c fiber.Ctx) error {
		return c.Render("pages/home", renderHomeContext(c))
	})

	app.Get("/privacy", auth.RequireAuth(), func(c fiber.Ctx) error {
		m := baseMap(c)
		m["Title"] = "Privacy"
		return c.Render("pages/privacy", m)
	})

	app.Get("/fragments/compliance-tile", auth.RequireAuth(), func(c fiber.Ctx) error {
		return c.Render("fragments/compliance_tile", fiber.Map{}, "")
	})

	// POST /preferences/theme — exempt from RequireAuth so the toggle works on /login.
	app.Post("/preferences/theme", func(c fiber.Ctx) error {
		value := c.FormValue("value")
		if !fmThemeAllowed[value] {
			return c.SendStatus(400)
		}
		c.Cookie(&fiber.Cookie{
			Name:     "fm_theme",
			Value:    value,
			Path:     "/",
			MaxAge:   31536000,
			SameSite: "Lax",
		})
		c.Set("HX-Trigger", "theme-changed")
		return c.SendStatus(204)
	})
}

func runDumpRoutes() {
	// Build a minimal app with nil pool — middleware is skipped, no DB needed.
	app := buildApp(nil)
	registerRoutes(app, nil)
	var lines []string
	for _, r := range app.GetRoutes(true) {
		method := strings.ToLower(r.Method)
		path := strings.ToLower(r.Path)
		// Exclude static asset middleware routes.
		if strings.HasPrefix(path, "/static") {
			continue
		}
		// Exclude HEAD auto-mirrors that Fiber adds for every GET.
		if method == "head" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s", method, path))
	}
	sort.Strings(lines)
	for _, l := range lines {
		fmt.Println(l)
	}
}

func runServer() {
	dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	pool, err := postgres.Connect(dsn)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("database connection validated")

	app := buildApp(pool)
	registerRoutes(app, pool)
	log.Fatal(app.Listen(":3000"))
}

func main() {
	dumpRoutes := flag.Bool("dump-routes", false, "print normalized route inventory and exit")
	flag.Parse()

	if *dumpRoutes {
		runDumpRoutes()
		return
	}

	runServer()
}
