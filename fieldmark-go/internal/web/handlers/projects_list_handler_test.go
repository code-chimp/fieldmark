package handlers_test

// projects_list_handler_test.go — authz and page-render tests for GET /projects.

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

func makeProjectsListApp(actor *app.Actor) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := &handlers.ProjectsListHandlers{}
	a.Get("/projects", auth.RequireAuth(), h.GetProjectsList)
	return a
}

func TestGetProjectsList_Anonymous_RedirectsToLogin(t *testing.T) {
	a := makeProjectsListApp(nil)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/login" {
		t.Errorf("expected redirect to /login, got %s", loc)
	}
}

func TestGetProjectsList_NoRole_Returns403(t *testing.T) {
	noPermActor := &app.Actor{
		ID:       uuid.MustParse("99000000-0000-0000-0000-000000000099"),
		Username: "noperm",
		Role:     "NONEXISTENT_ROLE",
	}
	a := makeProjectsListApp(noPermActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 403, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestGetProjectsList_AuthenticatedAdmin_Returns200WithH1(t *testing.T) {
	a := makeProjectsListApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200, got %d; body: %s", resp.StatusCode, b)
	}
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "<h1>Projects</h1>") {
		t.Errorf("expected <h1>Projects</h1> in response body")
	}
}

func TestGetProjectsList_AuthenticatedCO_Returns200(t *testing.T) {
	a := makeProjectsListApp(coActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200 for CO, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestGetProjectsList_AuthenticatedInspector_Returns200(t *testing.T) {
	inspectorActor := &app.Actor{
		ID:       uuid.MustParse("98000000-0000-0000-0000-000000000098"),
		Username: "inspector",
		Role:     string(domain.RoleInspector),
	}
	a := makeProjectsListApp(inspectorActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 200 for Inspector, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestGetProjectsList_RendersGridContainer(t *testing.T) {
	a := makeProjectsListApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, `data-grid-endpoint="/grid/projects"`) {
		t.Error("expected data-grid-endpoint attribute")
	}
	if !strings.Contains(bodyStr, `data-grid-target="#project-detail"`) {
		t.Error("expected data-grid-target attribute")
	}
	if !strings.Contains(bodyStr, `ag-theme-quartz`) {
		t.Error("expected ag-theme-quartz class on grid container")
	}
}

func TestGetProjectsList_RendersProjectDetailTarget(t *testing.T) {
	a := makeProjectsListApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, `id="project-detail"`) {
		t.Error("expected #project-detail aside element")
	}
	if !strings.Contains(bodyStr, `tabindex="-1"`) {
		t.Error("expected tabindex=-1 on #project-detail")
	}
}

func TestGetProjectsList_RendersNoscriptFallback(t *testing.T) {
	a := makeProjectsListApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, "<noscript>") {
		t.Error("expected <noscript> fallback for no-JS visitors")
	}
}

func TestGetProjectsList_AdminSeesNewProjectButton(t *testing.T) {
	a := makeProjectsListApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if !strings.Contains(bodyStr, `href="/projects/new"`) {
		t.Error("expected server-rendered New Project link for admin")
	}
}

func TestGetProjectsList_NonAdminDoesNotSeeNewProjectButton(t *testing.T) {
	a := makeProjectsListApp(coActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects", nil)

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	if strings.Contains(bodyStr, `href="/projects/new"`) {
		t.Error("expected New Project link to be absent for non-admin")
	}
}
