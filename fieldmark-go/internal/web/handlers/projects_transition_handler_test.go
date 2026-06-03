package handlers_test

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/domain/enums"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

type projectTransitionNotFoundStoreStub struct{}

type projectTransitionStoreStub struct{}

type projectTransitionAuditReadStoreStub struct{}

type projectTransitionAuditAppendStoreStub struct{}

func (projectTransitionStoreStub) Load(context.Context, uuid.UUID) (*entities.Project, error) {
	return &entities.Project{
		ID:     uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		Code:   "PD-001",
		Name:   "Project Detail Go",
		Status: enums.ProjectStatusActive,
	}, nil
}
func (projectTransitionStoreStub) CreateInTx(context.Context, pgx.Tx, *entities.CreatedProject) error {
	return nil
}
func (projectTransitionStoreStub) LoadWithRelations(context.Context, uuid.UUID) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error) {
	return &entities.Project{
		ID:     uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		Code:   "PD-001",
		Name:   "Project Detail Go",
		Status: enums.ProjectStatusActive,
	}, nil, nil, nil, nil
}

func (projectTransitionNotFoundStoreStub) Load(context.Context, uuid.UUID) (*entities.Project, error) {
	return nil, postgres.ErrProjectNotFound
}
func (projectTransitionNotFoundStoreStub) CreateInTx(context.Context, pgx.Tx, *entities.CreatedProject) error {
	return nil
}
func (projectTransitionNotFoundStoreStub) LoadWithRelations(context.Context, uuid.UUID) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error) {
	return nil, nil, nil, nil, postgres.ErrProjectNotFound
}

func (projectTransitionAuditReadStoreStub) ListByProject(context.Context, uuid.UUID, postgres.AuditPage) (postgres.AuditPageResult, error) {
	return postgres.AuditPageResult{}, nil
}

func (projectTransitionAuditAppendStoreStub) Append(context.Context, pgx.Tx, *entities.AuditEntry) error {
	return nil
}

func makeProjectsTransitionApp(actor *app.Actor, store postgres.ProjectStore) *fiber.App {
	auth.ResetForTests()
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
		Projects:  store,
		Reference: projectDetailReferenceStoreStub{},
		Audit:     projectTransitionAuditAppendStoreStub{},
		AuditRead: projectTransitionAuditReadStoreStub{},
	}
	a.Get("/projects/:id/place-on-hold", auth.RequireAuth(), h.GetProjectPlaceOnHold)
	a.Post("/projects/:id/place-on-hold", auth.RequireAuth(), h.PostProjectPlaceOnHold)
	a.Get("/projects/:id/resume", auth.RequireAuth(), h.GetProjectResume)
	a.Post("/projects/:id/resume", auth.RequireAuth(), h.PostProjectResume)
	return a
}

func TestGetProjectPlaceOnHold_AnonymousRedirects(t *testing.T) {
	a := makeProjectsTransitionApp(nil, projectTransitionStoreStub{})
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 302/303, got %d", resp.StatusCode)
	}
}

func TestGetProjectPlaceOnHold_RendersReasonForm(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `role="form"`) || !strings.Contains(html, `hx-post="/projects/a1000000-0000-0000-0000-000000000001/place-on-hold"`) || !strings.Contains(html, `hx-target="#project-detail"`) {
		t.Fatalf("expected canonical reason form; body=%s", html)
	}
}

func TestGetProjectResume_RendersReasonForm(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/resume", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `role="form"`) || !strings.Contains(html, `hx-post="/projects/a1000000-0000-0000-0000-000000000001/resume"`) || !strings.Contains(html, `hx-target="#project-detail"`) {
		t.Fatalf("expected canonical resume form; body=%s", html)
	}
}

func TestGetProjectPlaceOnHold_ForbiddenForExecutive(t *testing.T) {
	executiveActor := &app.Actor{ID: uuid.New(), Username: "eli", Role: "EXECUTIVE"}
	a := makeProjectsTransitionApp(executiveActor, projectTransitionStoreStub{})
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if html != "You do not have permission to access this page." {
		t.Fatalf("expected canonical 403 body; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 403; body=%s", html)
	}
}

func TestGetProjectPlaceOnHold_UnknownIdReturns404(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionNotFoundStoreStub{})
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestPostProjectPlaceOnHold_BlankReasonReturns422WithoutOob(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {""}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if !strings.Contains(html, `Couldn&#39;t submit transition`) || !strings.Contains(html, `Reason is required.`) || !strings.Contains(html, `aria-invalid="true"`) {
		t.Fatalf("expected validation UI; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 422; body=%s", html)
	}
}

func TestPostProjectPlaceOnHold_ControlCharReasonReturns422WithoutOob(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {"bad\x01reason"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if !strings.Contains(html, `Reason contains invalid control characters.`) || !strings.Contains(html, `Couldn&#39;t submit transition`) {
		t.Fatalf("expected control-char validation; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 422; body=%s", html)
	}
}

func TestPostProjectPlaceOnHold_TooLongReasonReturns422WithoutOob(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {strings.Repeat("x", 501)}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if !strings.Contains(html, `Reason must be 500 characters or fewer.`) || !strings.Contains(html, `Couldn&#39;t submit transition`) {
		t.Fatalf("expected too-long validation UI; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 422; body=%s", html)
	}
}

func TestPostProjectPlaceOnHold_XssReasonEscapedOn422(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	payload := "<script>alert(1)</script>\x01"
	body := url.Values{"reason": {payload}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped payload; body=%s", html)
	}
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Fatalf("unexpected raw payload; body=%s", html)
	}
}

func TestPostProjectPlaceOnHold_ForbiddenForExecutive(t *testing.T) {
	executiveActor := &app.Actor{ID: uuid.New(), Username: "eli", Role: "EXECUTIVE"}
	a := makeProjectsTransitionApp(executiveActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {"nope"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/place-on-hold", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if html != "You do not have permission to access this page." {
		t.Fatalf("expected canonical 403 body; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 403; body=%s", html)
	}
}

func TestPostProjectResume_ForbiddenForExecutive(t *testing.T) {
	executiveActor := &app.Actor{ID: uuid.New(), Username: "eli", Role: "EXECUTIVE"}
	a := makeProjectsTransitionApp(executiveActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {"nope"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/resume", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if html != "You do not have permission to access this page." {
		t.Fatalf("expected canonical 403 body; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 403; body=%s", html)
	}
}

func TestPostProjectResume_TooLongReasonReturns422WithoutOob(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {strings.Repeat("x", 501)}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/resume", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if !strings.Contains(html, `Reason must be 500 characters or fewer.`) || !strings.Contains(html, `Couldn&#39;t submit transition`) {
		t.Fatalf("expected too-long validation UI; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 422; body=%s", html)
	}
}

func TestPostProjectResume_ControlCharReasonReturns422WithoutOob(t *testing.T) {
	a := makeProjectsTransitionApp(adminActor, projectTransitionStoreStub{})
	body := url.Values{"reason": {"bad\x01reason"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/a1000000-0000-0000-0000-000000000001/resume", strings.NewReader(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}
	raw, _ := io.ReadAll(resp.Body)
	html := string(raw)
	if !strings.Contains(html, `Reason contains invalid control characters.`) || !strings.Contains(html, `Couldn&#39;t submit transition`) {
		t.Fatalf("expected control-char validation UI; body=%s", html)
	}
	if strings.Contains(html, `hx-swap-oob=`) {
		t.Fatalf("unexpected OOB markup on 422; body=%s", html)
	}
}
