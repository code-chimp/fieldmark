package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
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

func themeMap(c fiber.Ctx) fiber.Map {
	current := resolveFmTheme(c)
	return fiber.Map{
		"FmTheme":         current,
		"FmThemeNext":     fmThemeCycle[current],
		"FmThemeResolved": current,
	}
}

func main() {
	dumpRoutes := flag.Bool("dump-routes", false, "print normalized route inventory and exit")
	flag.Parse()

	// --- Template engine --------------------------------------------------
	// html.New walks internal/web/templates/ and loads all *.html files.
	// The Layout option wraps every c.Render() call in layouts/base.html
	// unless the handler passes an empty layout string explicitly.
	engine := html.New("./internal/web/templates", ".html")
	engine.Layout("base")
	engine.AddFunc("noescape", func(s string) string { return s })

	// --- Application ------------------------------------------------------
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(logger.New())

	// Static assets: /static/** → internal/web/static/
	app.Use("/static", static.New("./internal/web/static"))

	// --- Routes -----------------------------------------------------------

	// Full page — dashboard
	app.Get("/", func(c fiber.Ctx) error {
		m := themeMap(c)
		m["Title"] = "Dashboard"
		return c.Render("pages/dashboard", m)
	})

	// Full page — privacy policy
	app.Get("/privacy", func(c fiber.Ctx) error {
		m := themeMap(c)
		m["Title"] = "Privacy"
		return c.Render("pages/privacy", m)
	})

	// HTMX fragment — compliance tile (no layout wrapper)
	app.Get("/fragments/compliance-tile", func(c fiber.Ctx) error {
		return c.Render("fragments/compliance_tile", fiber.Map{}, "")
	})

	// POST /preferences/theme — set fm_theme cookie and signal client listener.
	// No CSRF middleware is mounted on this stack (auth is deferred; story 1-9 wires
	// Fiber authentication). Theme preference is non-security-sensitive UI state — a
	// CSRF attack would only flip a visitor's colour scheme. This is the intentional
	// parallel to .NET's [IgnoreAntiforgeryToken] on the equivalent Razor Page handler.
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

	// --- Dump routes and exit (parity tooling) ----------------------------
	// Checked AFTER route registration so GetRoutes reflects the full inventory,
	// but BEFORE database connect so no live DB is needed for a route dump.
	if *dumpRoutes {
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
		return
	}

	// --- Database ---------------------------------------------------------
	// Opened after the dump-routes short-circuit so parity tooling never
	// requires a live database connection.
	dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	conn, err := postgres.Connect(dsn)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()
	log.Println("database connection validated")

	// --- Listen -----------------------------------------------------------
	log.Fatal(app.Listen(":3000"))
}
