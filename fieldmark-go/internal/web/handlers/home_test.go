package handlers_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
)

// injectActor returns a middleware that sets a pre-built Actor into locals,
// simulating what StubAuthMiddleware does after a successful DB lookup.
func injectActor(actor *app.Actor) fiber.Handler {
	return func(c fiber.Ctx) error {
		c.Locals("user", actor)
		return c.Next()
	}
}

func buildHomeApp(actor *app.Actor) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	a.Get("/", auth.RequireAuth(), func(c fiber.Ctx) error {
		return c.Redirect().To("/dashboard")
	})
	return a
}

func TestHomeUnauthenticatedRedirectsToLogin(t *testing.T) {
	a := buildHomeApp(nil)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected 302 or 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Errorf("expected redirect to /login, got %q", loc)
	}
}

func TestHomeAuthenticatedAdminRedirectsToDashboard(t *testing.T) {
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Errorf("expected 302 or 303, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/dashboard" {
		t.Errorf("expected redirect to /dashboard, got %q", loc)
	}
}
