package handlers_test

// grid_projects_handler_test.go — unit-level authz tests for POST /grid/projects.
// Integration tests (SSRM conformance, filter/sort/page, projection key-set,
// zero/null faithfulness, injection rejection) live in the integration test suite
// (build tag: integration) in grid_projects_integration_test.go.

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

func init() {
	auth.RegisterAction("project.read",
		domain.RoleAdmin, domain.RoleComplianceOfficer, domain.RoleInspector,
		domain.RoleSiteSupervisor, domain.RoleExecutive,
	)
}

var (
	coActor = &app.Actor{
		ID:       uuid.MustParse("10000000-0000-0000-0000-000000000002"),
		Username: "marisol",
		Role:     string(domain.RoleComplianceOfficer),
	}
)

func makeGridProjectsApp(actor *app.Actor) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := &handlers.GridProjectsHandlers{Pool: nil} // nil pool — not reached in 403 paths
	a.Post("/grid/projects", auth.RequireAuth(), h.PostGridProjects)
	return a
}

// ─── Authz ────────────────────────────────────────────────────────────────────

func TestPostGridProjects_Anonymous_Redirects(t *testing.T) {
	a := makeGridProjectsApp(nil)
	body := bytes.NewBufferString(`{"startRow":0,"endRow":10,"sortModel":[],"filterModel":{}}`)
	req, _ := http.NewRequest(http.MethodPost, "/grid/projects", body)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 302 or 403 for anonymous, got %d", resp.StatusCode)
	}
}

func TestPostGridProjects_ComplianceOfficer_Forbidden_Without_Permission(t *testing.T) {
	// CO does have project.read; this tests that 403 doesn't fire when they DO have permission.
	// The real 403 path is tested via a role without any registered permissions.
	noPermActor := &app.Actor{
		ID:       uuid.New(),
		Username: "norole",
		Role:     "NONEXISTENT_ROLE",
	}
	a := makeGridProjectsApp(noPermActor)
	body := bytes.NewBufferString(`{"startRow":0,"endRow":10,"sortModel":[],"filterModel":{}}`)
	req, _ := http.NewRequest(http.MethodPost, "/grid/projects", body)
	req.Header.Set("Content-Type", "application/json")

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

// ─── 400 validation (no DB needed) ───────────────────────────────────────────

func makeGridRequestWithAuth(actor *app.Actor, body string) (*http.Response, error) {
	a := makeGridProjectsApp(actor)
	req, _ := http.NewRequest(http.MethodPost, "/grid/projects", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	return a.Test(req)
}

func TestPostGridProjects_InvalidJSON_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `NOT JSON`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_UnknownColId_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[{"colId":"UNKNOWN","sort":"asc"}],"filterModel":{}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_InjectionColId_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[{"colId":"code; DROP TABLE domain.project --","sort":"asc"}],"filterModel":{}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for injection-style colId, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_InvalidSortDirection_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[{"colId":"code","sort":"INVALID"}],"filterModel":{}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for invalid sort direction, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_NegativeStartRow_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":-1,"endRow":10,"sortModel":[],"filterModel":{}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_PageSizeExceedsMax_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":1001,"sortModel":[],"filterModel":{}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for page size > 1000, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_InvalidStatusValue_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[],"filterModel":{"status":{"filterType":"set","values":["INVALID"]}}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for invalid status value, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_ExplicitNullSortModel_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":null,"filterModel":{}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for explicit null sortModel, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_ExplicitNullFilterModel_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[],"filterModel":null}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for explicit null filterModel, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_NumberValueInTextField_Returns400(t *testing.T) {
	// Go's parseTextFilter unmarshals "filter" into a string field; a JSON number
	// (42) causes json.Unmarshal to fail → error returned → 400.
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[],"filterModel":{"code":{"filterType":"text","type":"contains","filter":42}}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for numeric filter value in text filter, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_WrongFilterTypeForColumn_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[],"filterModel":{"compliance_score":{"filterType":"text","type":"contains","filter":"50"}}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for wrong filterType on compliance_score, got %d; body: %s", resp.StatusCode, b)
	}
}

func TestPostGridProjects_InjectionStatusValue_Returns400(t *testing.T) {
	resp, err := makeGridRequestWithAuth(coActor, `{"startRow":0,"endRow":10,"sortModel":[],"filterModel":{"status":{"filterType":"set","values":["Active' OR '1'='1"]}}}`)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		t.Errorf("expected 400 for injection-style status value, got %d; body: %s", resp.StatusCode, b)
	}
}
