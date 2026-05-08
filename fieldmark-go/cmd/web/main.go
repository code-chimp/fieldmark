package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/gofiber/template/html/v2"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
)

func main() {
	// --- Database ---------------------------------------------------------
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	conn, err := postgres.Connect(dsn)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer conn.Close(nil) //nolint:errcheck
	log.Println("database connection validated")

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

	// HTMX fragment — compliance tile (no layout wrapper)
	app.Get("/fragments/compliance-tile", func(c fiber.Ctx) error {
		return c.Render("fragments/compliance_tile", fiber.Map{}, "")
	})

	// --- Listen -----------------------------------------------------------
	log.Fatal(app.Listen(":3000"))
}
