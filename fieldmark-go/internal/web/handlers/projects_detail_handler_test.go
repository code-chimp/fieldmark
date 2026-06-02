package handlers_test

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

type projectStoreStub struct{}

var stubProjectStatus = enums.ProjectStatusActive
var stubProjectName = "Project Detail Go"
var stubProjectDescription *string
var stubProjectTargetDate *time.Time

func (projectStoreStub) Load(context.Context, uuid.UUID) (*entities.Project, error) { return nil, postgres.ErrProjectNotFound }
func (projectStoreStub) CreateInTx(context.Context, pgx.Tx, *entities.CreatedProject) error {
	return nil
}
func (projectStoreStub) LoadWithRelations(context.Context, uuid.UUID) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	p := &entities.Project{
		ID:              uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		Code:            "PD-001",
		Name:            stubProjectName,
		Status:          stubProjectStatus,
		StartDate:       now,
		TargetCompletionDate: stubProjectTargetDate,
		Description:     stubProjectDescription,
		ComplianceScore: 100,
	}
	return p, nil, nil, nil, nil
}

type projectDetailReferenceStoreStub struct{}

func (projectDetailReferenceStoreStub) ListTradeTypes(context.Context) ([]entities.TradeType, error) {
	return nil, nil
}
func (projectDetailReferenceStoreStub) ListViolationCategories(context.Context) ([]entities.ViolationCategory, error) {
	return nil, nil
}
func (projectDetailReferenceStoreStub) ListComplianceRules(context.Context) ([]entities.ComplianceRule, error) {
	return nil, nil
}

func makeProjectsDetailApp(actor *app.Actor) *fiber.App {
	auth.ResetForTests()
	stubProjectStatus = enums.ProjectStatusActive
	stubProjectName = "Project Detail Go"
	stubProjectDescription = nil
	stubProjectTargetDate = nil
	auth.RegisterAction("project.read", "ADMIN", "COMPLIANCE_OFFICER", "INSPECTOR", "SITE_SUPERVISOR", "EXECUTIVE")
	auth.RegisterAction("project.place_on_hold", "ADMIN")
	auth.RegisterAction("project.resume", "ADMIN")
	auth.RegisterAction("project.close", "ADMIN")
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := &handlers.ProjectsDetailHandlers{
		Pool:      nil,
		Projects:  projectStoreStub{},
		Reference: projectDetailReferenceStoreStub{},
	}
	a.Get("/projects/:id", auth.RequireAuth(), h.GetProjectsDetail)
	a.Get("/projects/:id/tabs/:tab", auth.RequireAuth(), h.GetProjectsDetail)
	return a
}

func TestGetProjectsDetail_UnauthenticatedRedirects(t *testing.T) {
	a := makeProjectsDetailApp(nil)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 302/303, got %d", resp.StatusCode)
	}
}

func TestGetProjectsDetail_HxRequestReturnsFragment(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	html := string(b)
	if strings.Contains(html, `id="project-detail"`) || !strings.Contains(html, `id="project-header-strip"`) || !strings.Contains(html, `id="project-detail-tabstrip"`) {
		t.Fatalf("missing canonical ids; body=%s", html)
	}
	if strings.Contains(strings.ToLower(html), "<html") {
		t.Fatalf("expected body fragment only for HX request")
	}
}

func TestGetProjectsDetail_FullPageWrapsFragment(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	html := string(b)
	if !strings.Contains(html, `<div id="project-detail">`) || !strings.Contains(html, `id="project-header-strip"`) {
		t.Fatalf("expected standalone wrapper around detail fragment; body=%s", html)
	}
}

func TestGetProjectsDetailTab_NonHtmxRedirectsToDetail(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/violations", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 302/303, got %d", resp.StatusCode)
	}
}

func TestGetProjectsDetail_NoRoleForbidden(t *testing.T) {
	noRoleActor := &app.Actor{ID: uuid.New(), Username: "norole", Role: "NO_ROLE"}
	a := makeProjectsDetailApp(noRoleActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
}

func TestGetProjectsDetail_AdminClosedAllButtonsDisabled(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	stubProjectStatus = enums.ProjectStatusClosed
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	for _, id := range []string{`id="place-on-hold-btn"`, `id="resume-btn"`, `id="close-btn"`} {
		if !strings.Contains(html, id) {
			t.Fatalf("expected %s", id)
		}
	}
	for _, reason := range []string{
		`aria-describedby="place-on-hold-btn-reason"`,
		`aria-describedby="resume-btn-reason"`,
		`aria-describedby="close-btn-reason"`,
	} {
		if !strings.Contains(html, reason) {
			t.Fatalf("expected %s", reason)
		}
	}
}

func TestGetProjectsDetail_AdminActiveShowsHoldCloseAndDisablesResume(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	stubProjectStatus = enums.ProjectStatusActive
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `id="place-on-hold-btn"`) || !strings.Contains(html, `id="close-btn"`) || !strings.Contains(html, `id="resume-btn"`) {
		t.Fatalf("expected all action button ids")
	}
	if !strings.Contains(html, `aria-describedby="resume-btn-reason"`) {
		t.Fatalf("expected disabled resume reason")
	}
}

func TestGetProjectsDetail_AdminOnHoldShowsResumeAndDisablesOthers(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	stubProjectStatus = enums.ProjectStatusOnHold
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `id="place-on-hold-btn"`) || !strings.Contains(html, `id="close-btn"`) || !strings.Contains(html, `id="resume-btn"`) {
		t.Fatalf("expected all action button ids")
	}
	if !strings.Contains(html, `aria-describedby="place-on-hold-btn-reason"`) || !strings.Contains(html, `aria-describedby="close-btn-reason"`) {
		t.Fatalf("expected disabled reasons for hold and close")
	}
}

func TestGetProjectsDetail_ExecutiveHidesButtons(t *testing.T) {
	executiveActor := &app.Actor{ID: uuid.New(), Username: "eli", Role: string(domain.RoleExecutive)}
	a := makeProjectsDetailApp(executiveActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if strings.Contains(html, `id="place-on-hold-btn"`) || strings.Contains(html, `id="resume-btn"`) || strings.Contains(html, `id="close-btn"`) {
		t.Fatalf("expected no action buttons for executive")
	}
}

func TestGetProjectsDetail_XssPayloadEscaped(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	t.Cleanup(func() {
		stubProjectName = "Project Detail Go"
		stubProjectDescription = nil
		stubProjectTargetDate = nil
	})
	payload := "<script>alert(1)</script>"
	stubProjectName = payload
	stubProjectDescription = &payload
	d := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)
	stubProjectTargetDate = &d
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped payload")
	}
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Fatalf("unexpected raw payload")
	}
}

func TestGetProjectsDetailTab_Violations_ReturnsPanelAndOob(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/violations", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	b, _ := io.ReadAll(resp.Body)
	html := string(b)
	if !strings.Contains(html, `aria-labelledby="tab-violations"`) {
		t.Fatalf("missing violations panel aria-labelledby")
	}
	if !strings.Contains(html, `hx-swap-oob="outerHTML"`) {
		t.Fatalf("missing OOB tabstrip markup")
	}
}
