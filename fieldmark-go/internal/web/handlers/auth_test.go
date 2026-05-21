package handlers_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/template/html/v2"

	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

// newTestApp creates a Fiber app with the real template engine
// (relative to the project root, two levels up from this package).
func newTestApp() *fiber.App {
	engine := html.New("../templates", ".html")
	engine.AddFunc("noescape", func(s string) string { return s })
	return fiber.New(fiber.Config{Views: engine, ViewsLayout: "base"})
}

func TestPostLogin_EmptyUsername_Returns422(t *testing.T) {
	app := newTestApp()
	h := &handlers.LoginHandlers{Pool: nil}
	app.Post("/login", h.PostLogin)

	req, _ := http.NewRequest(http.MethodPost, "/login",
		bytes.NewBufferString("username="))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 422, got %d; body: %s", resp.StatusCode, body)
	}
}

func TestPostLogout_ClearsCookieAndRedirects(t *testing.T) {
	app := newTestApp()
	h := &handlers.LoginHandlers{Pool: nil}
	app.Post("/logout", h.PostLogout)

	req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Cookie", auth.CookieName()+"=marisol")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "/login" {
		t.Errorf("expected redirect to /login, got %q", location)
	}

	// Response must include a Set-Cookie that clears the actor cookie.
	setCookie := resp.Header.Get("Set-Cookie")
	if setCookie == "" {
		t.Error("expected Set-Cookie header to be present on logout")
	}
}
