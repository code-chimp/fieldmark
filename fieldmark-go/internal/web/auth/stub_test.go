package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
)

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("readBody: %v", err)
	}
	return string(b)
}

func TestResolveUsername_CookieWinsOverHeader(t *testing.T) {
	a := fiber.New()
	a.Get("/probe", func(c fiber.Ctx) error {
		return c.SendString(resolveUsername(c))
	})

	req := httptest.NewRequest("GET", "/probe", nil)
	req.AddCookie(&http.Cookie{Name: cookieName, Value: "marisol"})
	req.Header.Set(headerName, "diego")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("a.Test: %v", err)
	}
	body := readBody(t, resp)
	if body != "marisol" {
		t.Fatalf("want marisol, got %q", body)
	}
}

func TestResolveUsername_HeaderWhenNoCookie(t *testing.T) {
	a := fiber.New()
	a.Get("/probe", func(c fiber.Ctx) error {
		return c.SendString(resolveUsername(c))
	})

	req := httptest.NewRequest("GET", "/probe", nil)
	req.Header.Set(headerName, "diego")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("a.Test: %v", err)
	}
	body := readBody(t, resp)
	if body != "diego" {
		t.Fatalf("want diego, got %q", body)
	}
}

func TestResolveUsername_EnvWhenNoCookieAndNoHeader(t *testing.T) {
	t.Setenv(envVar, "kenji")

	a := fiber.New()
	a.Get("/probe", func(c fiber.Ctx) error {
		return c.SendString(resolveUsername(c))
	})

	req := httptest.NewRequest("GET", "/probe", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("a.Test: %v", err)
	}
	body := readBody(t, resp)
	if body != "kenji" {
		t.Fatalf("want kenji, got %q", body)
	}
}

func TestResolveUsername_EmptyWhenNoneProvided(t *testing.T) {
	t.Setenv(envVar, "")

	a := fiber.New()
	a.Get("/probe", func(c fiber.Ctx) error {
		return c.SendString(resolveUsername(c))
	})

	req := httptest.NewRequest("GET", "/probe", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("a.Test: %v", err)
	}
	body := readBody(t, resp)
	if body != "" {
		t.Fatalf("want empty string, got %q", body)
	}
}

func TestRequireAuth_RedirectsAnonymousToLogin(t *testing.T) {
	a := fiber.New()
	a.Use(func(c fiber.Ctx) error {
		c.Locals(localsKey, app.Anonymous())
		return c.Next()
	})
	a.Get("/secure", RequireAuth(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := a.Test(httptest.NewRequest("GET", "/secure", nil))
	if err != nil {
		t.Fatalf("a.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusFound {
		t.Fatalf("want 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Fatalf("want /login, got %q", loc)
	}
}

func TestRequireAuth_PassesAuthenticatedThrough(t *testing.T) {
	a := fiber.New()
	a.Use(func(c fiber.Ctx) error {
		c.Locals(localsKey, &app.Actor{
			ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Username: "marisol",
			Role:     "ADMIN",
		})
		return c.Next()
	})
	a.Get("/secure", RequireAuth(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	resp, err := a.Test(httptest.NewRequest("GET", "/secure", nil))
	if err != nil {
		t.Fatalf("a.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("want 200, got %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if body != "ok" {
		t.Fatalf("want ok, got %q", body)
	}
}
