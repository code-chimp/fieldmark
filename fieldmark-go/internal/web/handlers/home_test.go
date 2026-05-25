package handlers_test

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/testutil"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
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
		resolved := auth.ActorFromCtx(c)
		role := domain.Role(resolved.Role)
		return c.Render("pages/home", fiber.Map{
			"Title":           "FieldMark",
			"RoleLabel":       role.Label(),
			"RoleBadgeToken":  role.BadgeToken(),
			"FullName":        resolved.DisplayName,
			"Initials":        viewmodels.Initials(resolved.DisplayName, resolved.Username),
			"Actor":           resolved,
			"FmTheme":         "system",
			"FmThemeNext":     "light",
			"FmThemeResolved": "system",
		})
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

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Errorf("expected redirect to /login, got %q", loc)
	}
}

func TestHomeAuthenticatedAdminRendersRoleBadgeAndPlaceholder(t *testing.T) {
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "<h1>FieldMark</h1>") {
		t.Error("expected <h1>FieldMark</h1>")
	}
	if !strings.Contains(html, "badge-danger") {
		t.Error("expected badge-danger class")
	}
	if !strings.Contains(html, ">Admin<") {
		t.Error("expected >Admin< in badge")
	}
	if !strings.Contains(html, "Your projects will appear here.") {
		t.Error("expected placeholder paragraph")
	}
}

func TestHomeAuthenticatedRendersAvatarMenu(t *testing.T) {
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, "avatar-menu-wrapper") {
		t.Error("expected avatar-menu-wrapper")
	}
	if !strings.Contains(html, "avatar-menu-dropdown") {
		t.Error("expected avatar-menu-dropdown")
	}
	if !strings.Contains(html, `href="/logout"`) {
		t.Error("expected logout anchor href")
	}
}

// TestHomeAuthenticatedPassesAxeCore renders the Home page and runs @axe-core/cli
// against it (AC #6). Skips gracefully when npx is not on PATH — surface skips in CI.
// Manual recipe: npx @axe-core/cli http://localhost:3000/ (authenticated session required).
func TestHomeAuthenticatedPassesAxeCore(t *testing.T) {
	if _, err := exec.LookPath("npx"); err != nil {
		t.Skip("npx not on PATH; install Node.js to enable axe-core WCAG 2.1 AA gate")
	}

	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	f, err := os.CreateTemp("", "fieldmark-home-*.html")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(body); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	f.Close()

	cmd := exec.Command("npx", "@axe-core/cli", "file://"+f.Name())
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("axe-core found WCAG 2.1 AA violations:\n%s", out)
	}
}

// TestHomeTabOrderMatchesContract verifies that the focusable elements required by AC #7
// appear in the correct DOM order: skip-link → brand lockup → theme-toggle pill → avatar button → sign-out.
// DOM order is the primary determinant of tab order when no tabindex attributes override it.
// Full runtime tab-order (CSS display, tabindex) still requires chromedp/Playwright (Epic 7 Story 7.1).
// Manual recipe: open http://localhost:3000/, Tab 5 times, verify sequence above.
func TestHomeTabOrderMatchesContract(t *testing.T) {
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	markers := []struct {
		name string
		text string
	}{
		{"skip-link", `class="skip-link"`},
		{"fm-brand-lockup", `class="fm-brand-lockup"`},
		{"theme-toggle-pill", `class="theme-toggle-pill"`},
		{"avatar-menu button", `class="avatar-menu"`},
		{"sign-out anchor", `href="/logout"`},
	}

	indices := make([]int, len(markers))
	for i, m := range markers {
		idx := strings.Index(html, m.text)
		if idx == -1 {
			t.Errorf("expected %q to be present in HTML", m.name)
		}
		indices[i] = idx
	}

	for i := 1; i < len(indices); i++ {
		if indices[i-1] >= indices[i] {
			t.Errorf("DOM order violation: %q (idx %d) must precede %q (idx %d)",
				markers[i-1].name, indices[i-1], markers[i].name, indices[i])
		}
	}
}

func TestHomeAuthenticatedRendersBrandLockup(t *testing.T) {
	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	if !strings.Contains(html, `class="fm-brand-lockup"`) {
		t.Error("expected fm-brand-lockup class")
	}
	if !strings.Contains(html, `aria-label="FieldMark home"`) {
		t.Error("expected aria-label on brand lockup")
	}
}

var (
	headerRe = regexp.MustCompile(`(?s)<header[^>]*>.*?</header>`)
	mainRe   = regexp.MustCompile(`(?s)<main[^>]*>.*?</main>`)
)

// TestHomeChromeMatchesParitySnapshot renders the Home page with the canonical dev user,
// normalizes, and compares to the committed cross-stack snapshot (AC #8).
func TestHomeChromeMatchesParitySnapshot(t *testing.T) {
	snapshotPath := filepath.Join("..", "..", "..", "..", "_bmad-output", "implementation-artifacts", "_parity-snapshots", "home-chrome.normalized.html")
	snapshotBytes, err := os.ReadFile(snapshotPath)
	if err != nil {
		t.Skipf("parity snapshot not found at %s: %v", snapshotPath, err)
	}
	snapshot := testutil.NormaliseForParity(string(snapshotBytes))

	actor := &app.Actor{ID: uuid.New(), Username: "aisha", DisplayName: "Aisha Patel", Role: string(domain.RoleAdmin)}
	a := buildHomeApp(actor)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	header := headerRe.FindString(html)
	main := mainRe.FindString(html)
	if header == "" || main == "" {
		t.Fatalf("expected <header> and <main> blocks in rendered HTML")
	}
	normalized := testutil.NormaliseForParity(header + "\n" + main)

	if normalized != snapshot {
		t.Errorf("home chrome diverges from parity snapshot.\nGot:\n%s\n\nWant:\n%s", normalized, snapshot)
	}
}
