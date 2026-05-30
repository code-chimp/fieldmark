package handlers_test

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/domain/entities"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

// ─── Stubs ────────────────────────────────────────────────────────────────────

// projectReferenceStoreStub satisfies both handlers.ReferenceStore and
// postgres.ProjectStore interfaces for unit tests.
type projectReferenceStoreStub struct {
	tradeTypes []entities.TradeType
}

func (s *projectReferenceStoreStub) ListTradeTypes(_ context.Context) ([]entities.TradeType, error) {
	return s.tradeTypes, nil
}
func (s *projectReferenceStoreStub) ListViolationCategories(_ context.Context) ([]entities.ViolationCategory, error) {
	return nil, nil
}
func (s *projectReferenceStoreStub) ListComplianceRules(_ context.Context) ([]entities.ComplianceRule, error) {
	return nil, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func buildProjectsCreateApp(actor *app.Actor) *handlers.ProjectsCreateHandlers {
	return &handlers.ProjectsCreateHandlers{
		Pool:      nil, // nil pool — must not be used in 403 paths
		Reference: &projectReferenceStoreStub{},
		Projects:  nil,
		Audit:     nil,
	}
}

func makeProjectsCreateFiberApp(actor *app.Actor) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := buildProjectsCreateApp(actor)
	a.Get("/projects/new", auth.RequireAuth(), h.GetProjectsNew)
	a.Post("/projects/", auth.RequireAuth(), h.PostProjectsCreate)
	return a
}

var (
	adminActor = &app.Actor{
		ID:       uuid.MustParse("372da3c7-1cf5-4455-9f01-005117e48d76"),
		Username: "aisha",
		Role:     string(domain.RoleAdmin),
	}
	nonAdminActor = &app.Actor{
		ID:       uuid.MustParse("4470e3c8-e754-419b-8b45-acb90fc5c7e9"),
		Username: "marisol",
		Role:     string(domain.RoleComplianceOfficer),
	}
)

func init() {
	// Register action so Can() is aware of it.
	auth.RegisterAction("project.create", domain.RoleAdmin)
}

// ─── GET /projects/new — 403 ──────────────────────────────────────────────────

func TestGetProjectsNew_Anonymous_RedirectsToLogin(t *testing.T) {
	a := makeProjectsCreateFiberApp(nil)
	req, _ := http.NewRequest(http.MethodGet, "/projects/new", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Errorf("status = %d; want 302", resp.StatusCode)
	}
}

func TestGetProjectsNew_NonAdmin_Returns403(t *testing.T) {
	a := makeProjectsCreateFiberApp(nonAdminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/new", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d; want 403", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if len(body) == 0 {
		t.Error("403 response body is empty")
	}
}

// ─── POST /projects/ — 403 ───────────────────────────────────────────────────

func TestPostProjectsCreate_NonAdmin_Returns403(t *testing.T) {
	a := makeProjectsCreateFiberApp(nonAdminActor)
	req, _ := http.NewRequest(http.MethodPost, "/projects/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("status = %d; want 403", resp.StatusCode)
	}
}

func TestPostProjectsCreate_Anonymous_RedirectsToLogin(t *testing.T) {
	a := makeProjectsCreateFiberApp(nil)
	req, _ := http.NewRequest(http.MethodPost, "/projects/", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusFound {
		t.Errorf("status = %d; want 302", resp.StatusCode)
	}
}

// ─── Method-not-allowed ───────────────────────────────────────────────────────

func TestGetProjectsCollection_Returns405(t *testing.T) {
	// Register only POST for /projects/ — Fiber returns 405 on GET.
	a := newTestApp()
	a.Use(injectActor(adminActor))
	h := buildProjectsCreateApp(adminActor)
	a.Post("/projects/", auth.RequireAuth(), h.PostProjectsCreate)

	req, _ := http.NewRequest(http.MethodGet, "/projects/", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test: %v", err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("status = %d; want 405", resp.StatusCode)
	}
	allow := resp.Header.Get("Allow")
	if allow == "" {
		t.Error("Allow header missing from 405 response")
	}
}
