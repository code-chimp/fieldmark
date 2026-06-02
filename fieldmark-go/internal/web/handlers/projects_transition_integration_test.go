//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/data/postgres"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
)

func openHandlerPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("FIELDMARK_DATABASE_URL"))
	if dsn == "" {
		dsn = "postgres://fieldmark:fieldmark@localhost:5432/fieldmark"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Postgres not reachable at %s: %v", dsn, err)
	}
	return pool
}

func makeProjectsTransitionIntegrationApp(pool *pgxpool.Pool) *fiber.App {
	auth.ResetForTests()
	auth.RegisterAction("project.read",
		domain.RoleAdmin, domain.RoleComplianceOfficer, domain.RoleInspector,
		domain.RoleSiteSupervisor, domain.RoleExecutive,
	)
	auth.RegisterAction("project.place_on_hold", domain.RoleAdmin)
	auth.RegisterAction("project.resume", domain.RoleAdmin)
	auth.RegisterAction("project.close", domain.RoleAdmin)

	a := newTestApp()
	a.Use(injectActor(adminActor))
	h := &handlers.ProjectsDetailHandlers{
		Pool:      pool,
		Projects:  postgres.NewProjectStore(pool),
		Reference: postgres.NewReferenceStore(pool),
		Audit:     postgres.NewAuditEntryStore(),
	}
	a.Post("/projects/:id/place-on-hold", auth.RequireAuth(), h.PostProjectPlaceOnHold)
	a.Post("/projects/:id/resume", auth.RequireAuth(), h.PostProjectResume)
	return a
}

func insertProjectRow(t *testing.T, pool *pgxpool.Pool, status string) uuid.UUID {
	t.Helper()

	id := uuid.New()
	code := "PD-" + strings.ToUpper(uuid.NewString()[:8])
	_, err := pool.Exec(context.Background(), `
		INSERT INTO domain.project
			(id, code, name, description, status, start_date, target_completion_date, compliance_score, created_at, updated_at)
		VALUES
			($1, $2, $3, NULL, $4, $5, NULL, 100, now(), now())
	`, id, code, "Go Transition Integration", status, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM domain.audit_entry WHERE project_id = $1`, id)
		_, _ = pool.Exec(context.Background(), `DELETE FROM domain.project WHERE id = $1`, id)
	})
	return id
}

func countOobRegions(body string) int {
	return strings.Count(body, `hx-swap-oob=`)
}

func TestPostProjectPlaceOnHold_Success_RendersThreeRegionShape_AndPersistsAudit(t *testing.T) {
	pool := openHandlerPool(t)
	defer pool.Close()

	id := insertProjectRow(t, pool, "Active")
	app := makeProjectsTransitionIntegrationApp(pool)

	form := url.Values{"reason": {"Weather delay"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/"+id.String()+"/place-on-hold", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d; body=%s", resp.StatusCode, string(b))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	if strings.Contains(body, `id="project-detail"`) {
		t.Fatalf("expected inner fragment only; body=%s", body)
	}
	if !strings.Contains(body, `id="compliance-tile"`) || !strings.Contains(body, `hx-swap-oob="afterbegin:#audit-log"`) {
		t.Fatalf("expected three-region response shape; body=%s", body)
	}
	if countOobRegions(body) != 2 {
		t.Fatalf("expected exactly 2 oob regions, got %d; body=%s", countOobRegions(body), body)
	}

	var status string
	if err := pool.QueryRow(context.Background(), `SELECT status FROM domain.project WHERE id = $1`, id).Scan(&status); err != nil {
		t.Fatalf("select project status: %v", err)
	}
	if status != "OnHold" {
		t.Fatalf("status = %s; want OnHold", status)
	}

	var action string
	var beforeStateBytes []byte
	var afterStateBytes []byte
	var metadataBytes []byte
	if err := pool.QueryRow(context.Background(),
		`SELECT action, before_state, after_state, metadata FROM domain.audit_entry WHERE project_id = $1 ORDER BY occurred_at DESC LIMIT 1`,
		id,
	).Scan(&action, &beforeStateBytes, &afterStateBytes, &metadataBytes); err != nil {
		t.Fatalf("select audit row: %v", err)
	}
	if action != "ProjectPlacedOnHold" {
		t.Fatalf("action = %s; want ProjectPlacedOnHold", action)
	}
	var beforeState map[string]string
	if err := json.Unmarshal(beforeStateBytes, &beforeState); err != nil {
		t.Fatalf("unmarshal before_state: %v", err)
	}
	if beforeState["status"] != "Active" {
		t.Fatalf("before_state.status = %q; want Active", beforeState["status"])
	}
	var afterState map[string]string
	if err := json.Unmarshal(afterStateBytes, &afterState); err != nil {
		t.Fatalf("unmarshal after_state: %v", err)
	}
	if afterState["status"] != "OnHold" {
		t.Fatalf("after_state.status = %q; want OnHold", afterState["status"])
	}
	var metadata map[string]string
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if metadata["reason"] != "Weather delay" {
		t.Fatalf("metadata.reason = %q; want Weather delay", metadata["reason"])
	}
}

func TestPostProjectResume_FromActive_Returns409_WithoutOob(t *testing.T) {
	pool := openHandlerPool(t)
	defer pool.Close()

	id := insertProjectRow(t, pool, "Active")
	app := makeProjectsTransitionIntegrationApp(pool)

	form := url.Values{"reason": {"stale request"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/"+id.String()+"/resume", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 409, got %d; body=%s", resp.StatusCode, string(b))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	if strings.Contains(body, `id="project-detail"`) {
		t.Fatalf("expected inner fragment only; body=%s", body)
	}
	if !strings.Contains(body, `Couldn&#39;t resume project`) || !strings.Contains(body, `Project is not on hold`) {
		t.Fatalf("expected inline conflict alert; body=%s", body)
	}
	if countOobRegions(body) != 0 {
		t.Fatalf("expected zero oob regions, got %d; body=%s", countOobRegions(body), body)
	}

	var auditCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM domain.audit_entry WHERE project_id = $1`, id).Scan(&auditCount); err != nil {
		t.Fatalf("count audit rows: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("audit rows = %d; want 0", auditCount)
	}
}

func TestPostProjectPlaceOnHold_FromOnHold_Returns409_WithoutOob(t *testing.T) {
	pool := openHandlerPool(t)
	defer pool.Close()

	id := insertProjectRow(t, pool, "OnHold")
	app := makeProjectsTransitionIntegrationApp(pool)

	form := url.Values{"reason": {"stale request"}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/"+id.String()+"/place-on-hold", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 409, got %d; body=%s", resp.StatusCode, string(b))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	if strings.Contains(body, `id="project-detail"`) {
		t.Fatalf("expected inner fragment only; body=%s", body)
	}
	if !strings.Contains(body, `Couldn&#39;t place project on hold`) || !strings.Contains(body, `Project is already on hold`) {
		t.Fatalf("expected inline conflict alert; body=%s", body)
	}
	if countOobRegions(body) != 0 {
		t.Fatalf("expected zero oob regions, got %d; body=%s", countOobRegions(body), body)
	}

	var auditCount int
	if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM domain.audit_entry WHERE project_id = $1`, id).Scan(&auditCount); err != nil {
		t.Fatalf("count audit rows: %v", err)
	}
	if auditCount != 0 {
		t.Fatalf("audit rows = %d; want 0", auditCount)
	}
}

func TestPostProjectResume_BlankReason_IsAccepted_AndPersistsAudit(t *testing.T) {
	pool := openHandlerPool(t)
	defer pool.Close()

	id := insertProjectRow(t, pool, "OnHold")
	app := makeProjectsTransitionIntegrationApp(pool)

	form := url.Values{"reason": {""}}
	req, _ := http.NewRequest(http.MethodPost, "/projects/"+id.String()+"/resume", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("HX-Request", "true")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d; body=%s", resp.StatusCode, string(b))
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)
	if countOobRegions(body) != 2 {
		t.Fatalf("expected exactly 2 oob regions, got %d; body=%s", countOobRegions(body), body)
	}

	var status string
	if err := pool.QueryRow(context.Background(), `SELECT status FROM domain.project WHERE id = $1`, id).Scan(&status); err != nil {
		t.Fatalf("select project status: %v", err)
	}
	if status != "Active" {
		t.Fatalf("status = %s; want Active", status)
	}

	var action string
	var beforeStateBytes []byte
	var afterStateBytes []byte
	var metadataBytes []byte
	if err := pool.QueryRow(context.Background(),
		`SELECT action, before_state, after_state, metadata FROM domain.audit_entry WHERE project_id = $1 ORDER BY occurred_at DESC LIMIT 1`,
		id,
	).Scan(&action, &beforeStateBytes, &afterStateBytes, &metadataBytes); err != nil {
		t.Fatalf("select audit row: %v", err)
	}
	if action != "ProjectResumed" {
		t.Fatalf("action = %s; want ProjectResumed", action)
	}
	var beforeState map[string]string
	if err := json.Unmarshal(beforeStateBytes, &beforeState); err != nil {
		t.Fatalf("unmarshal before_state: %v", err)
	}
	if beforeState["status"] != "OnHold" {
		t.Fatalf("before_state.status = %q; want OnHold", beforeState["status"])
	}
	var afterState map[string]string
	if err := json.Unmarshal(afterStateBytes, &afterState); err != nil {
		t.Fatalf("unmarshal after_state: %v", err)
	}
	if afterState["status"] != "Active" {
		t.Fatalf("after_state.status = %q; want Active", afterState["status"])
	}
	var metadata map[string]string
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		t.Fatalf("unmarshal metadata: %v", err)
	}
	if metadata["reason"] != "" {
		t.Fatalf("metadata.reason = %q; want empty string", metadata["reason"])
	}
}
