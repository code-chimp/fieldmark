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
		return c.Render("pages/dashboard", fiber.Map{
			"Title": "Dashboard",
		})
	})

	// Full page — privacy policy
	app.Get("/privacy", func(c fiber.Ctx) error {
		return c.Render("pages/privacy", fiber.Map{
			"Title": "Privacy",
		}, "")
	})

	// HTMX fragment — compliance tile (no layout wrapper)
	app.Get("/fragments/compliance-tile", func(c fiber.Ctx) error {
		return c.Render("fragments/compliance_tile", fiber.Map{}, "")
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
