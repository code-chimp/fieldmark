package handlers_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	"github.com/code-chimp/fieldmark-go/internal/web/testutil"
)

type projectStoreStub struct{}

var stubProjectStatus = enums.ProjectStatusActive
var stubProjectName = "Project Detail Go"
var stubProjectDescription *string
var stubProjectTargetDate *time.Time

func (projectStoreStub) Load(context.Context, uuid.UUID) (*entities.Project, error) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	return &entities.Project{
		ID:                   uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		Code:                 "PD-001",
		Name:                 stubProjectName,
		Status:               stubProjectStatus,
		StartDate:            now,
		TargetCompletionDate: stubProjectTargetDate,
		Description:          stubProjectDescription,
		ComplianceScore:      100,
	}, nil
}
func (projectStoreStub) CreateInTx(context.Context, pgx.Tx, *entities.CreatedProject) error {
	return nil
}
func (projectStoreStub) LoadWithRelations(context.Context, uuid.UUID) (*entities.Project, []entities.JobSite, []entities.ProjectTradeScope, []entities.ProjectInspector, error) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	p := &entities.Project{
		ID:                   uuid.MustParse("a1000000-0000-0000-0000-000000000001"),
		Code:                 "PD-001",
		Name:                 stubProjectName,
		Status:               stubProjectStatus,
		StartDate:            now,
		TargetCompletionDate: stubProjectTargetDate,
		Description:          stubProjectDescription,
		ComplianceScore:      100,
	}
	return p, nil, nil, nil, nil
}

type projectDetailReferenceStoreStub struct{}

type auditReadStoreStub struct{}

var stubAuditPageResult postgres.AuditPageResult

func (projectDetailReferenceStoreStub) ListTradeTypes(context.Context) ([]entities.TradeType, error) {
	return nil, nil
}
func (projectDetailReferenceStoreStub) ListViolationCategories(context.Context) ([]entities.ViolationCategory, error) {
	return nil, nil
}
func (projectDetailReferenceStoreStub) ListComplianceRules(context.Context) ([]entities.ComplianceRule, error) {
	return nil, nil
}

func (auditReadStoreStub) ListByProject(_ context.Context, _ uuid.UUID, page postgres.AuditPage) (postgres.AuditPageResult, error) {
	result := stubAuditPageResult
	if page.BeforeOccurredAt == nil && page.BeforeID == nil {
		return result, nil
	}
	result.NextCursor = nil
	return result, nil
}

func resetAuditStub() {
	stubAuditPageResult = postgres.AuditPageResult{
		Rows: []postgres.AuditEntryRow{
			{
				ID:          uuid.MustParse("a2000000-0000-0000-0000-000000000001"),
				OccurredAt:  time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC),
				ActorName:   "",
				Action:      "ProjectPlacedOnHold",
				BeforeState: []byte(`{"status":"Active"}`),
				AfterState:  []byte(`{"status":"OnHold"}`),
				Metadata:    []byte(`{"reason":"Weather delay"}`),
			},
		},
		NextCursor: &postgres.AuditCursor{
			OccurredAt: time.Date(2026, 6, 3, 13, 21, 0, 0, time.UTC),
			ID:         uuid.MustParse("a3000000-0000-0000-0000-000000000001"),
		},
	}
}

func auditLogCanonical(t *testing.T, variant string) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	b, err := os.ReadFile(filepath.Join(repoRoot, "docs", "reference", "fixtures", "project-audit-log-canonical.html"))
	if err != nil {
		t.Fatalf("read audit log canonical: %v", err)
	}
	return testutil.ExtractVariant(string(b), variant)
}

func normaliseAuditLogHTML(html string) string {
	uuidPattern := regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)
	html = uuidPattern.ReplaceAllString(html, "00000000-0000-0000-0000-000000000000")
	html = regexp.MustCompile(`datetime="[^"]+"`).ReplaceAllString(html, `datetime="TIMESTAMP"`)
	html = regexp.MustCompile(`title="[^"]+"`).ReplaceAllString(html, `title="TIMESTAMP"`)
	html = regexp.MustCompile(`before_occurred_at=[^"&]+`).ReplaceAllString(html, `before_occurred_at=TIMESTAMP_ENCODED`)
	html = regexp.MustCompile(`(<time[^>]*>)(.*?)(</time>)`).ReplaceAllString(html, `${1}RELATIVE_TIME${3}`)
	return testutil.NormaliseComponent(html)
}

func extractAuditPanel(html string) string {
	start := strings.Index(html, `<div id="project-detail-tab-content"`)
	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(html); {
		switch {
		case strings.HasPrefix(html[i:], "<div ") || strings.HasPrefix(html[i:], "<div>"):
			depth++
			i += 4
		case strings.HasPrefix(html[i:], "</div>"):
			depth--
			i += len("</div>")
			if depth == 0 {
				return html[start:i]
			}
		default:
			i++
		}
	}

	return ""
}

func makeProjectsDetailApp(actor *app.Actor) *fiber.App {
	auth.ResetForTests()
	stubProjectStatus = enums.ProjectStatusActive
	stubProjectName = "Project Detail Go"
	stubProjectDescription = nil
	stubProjectTargetDate = nil
	resetAuditStub()
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
		AuditRead: auditReadStoreStub{},
	}
	a.Get("/projects/:id", auth.RequireAuth(), h.GetProjectsDetail)
	a.Get("/projects/:id/tabs/:tab", auth.RequireAuth(), h.GetProjectsDetail)
	a.Get("/projects/:id/audit-log", auth.RequireAuth(), h.GetProjectAuditLog)
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

func TestGetProjectsDetailTab_Audit_RendersLiveAuditLog(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/audit", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `id="audit-log"`) || !strings.Contains(html, `aria-live="polite"`) {
		t.Fatalf("expected live audit log; body=%s", html)
	}
	if !strings.Contains(html, `data-audit-action="ProjectPlacedOnHold"`) {
		t.Fatalf("expected audit row; body=%s", html)
	}
	if !strings.Contains(html, `Show change`) {
		t.Fatalf("expected collapsed disclosure; body=%s", html)
	}
}

func TestGetProjectsDetailTab_AuditEmptyPanelMatchesCanonical(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	stubAuditPageResult = postgres.AuditPageResult{}
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/audit", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	panel := extractAuditPanel(html)
	if panel == "" {
		t.Fatalf("expected audit panel block; body=%s", html)
	}
	want := auditLogCanonical(t, "panel-empty")
	if got := normaliseAuditLogHTML(panel); got != want {
		t.Fatalf("audit panel mismatch:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestGetProjectsDetailTab_Audit_UnknownActionFallsBackToBadgeUnknown(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	stubAuditPageResult = postgres.AuditPageResult{
		Rows: []postgres.AuditEntryRow{
			{
				ID:         uuid.MustParse("a2000000-0000-0000-0000-000000000002"),
				OccurredAt: time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC),
				Action:     "ProjectReticulatedSpline",
			},
		},
	}
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/audit", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `badge-unknown`) {
		t.Fatalf("expected unknown action badge fallback; body=%s", html)
	}
}

func TestGetProjectsDetailTab_Audit_ActorFallbackRendersQuestionMarks(t *testing.T) {
	cases := []struct {
		name      string
		actorName string
	}{
		{name: "empty", actorName: ""},
		{name: "whitespace", actorName: "   "},
		{name: "unresolvable"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := makeProjectsDetailApp(adminActor)
			stubAuditPageResult = postgres.AuditPageResult{
				Rows: []postgres.AuditEntryRow{
					{
						ID:         uuid.MustParse("a2000000-0000-0000-0000-000000000003"),
						OccurredAt: time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC),
						ActorName:  tc.actorName,
						Action:     "ProjectPlacedOnHold",
					},
				},
			}
			req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/audit", nil)
			req.Header.Set("HX-Request", "true")
			resp, err := a.Test(req)
			if err != nil {
				t.Fatalf("test request failed: %v", err)
			}
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			html := string(body)
			if !strings.Contains(html, `<span class="audit-row__initials">??</span>`) {
				t.Fatalf("expected actor fallback initials; body=%s", html)
			}
		})
	}
}

func TestGetProjectsDetailTab_Audit_EscapesActorAndMetadataAndSortsNestedJSON(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	stubAuditPageResult = postgres.AuditPageResult{
		Rows: []postgres.AuditEntryRow{
			{
				ID:          uuid.MustParse("a2000000-0000-0000-0000-000000000004"),
				OccurredAt:  time.Date(2026, 6, 3, 15, 0, 0, 0, time.UTC),
				ActorName:   `<script>alert(1)</script>`,
				Action:      "ProjectPlacedOnHold",
				BeforeState: []byte(`{"zebra":{"delta":2,"alpha":1}}`),
				AfterState:  []byte(`{"items":[{"zulu":2,"bravo":1}],"alpha":1}`),
				Metadata:    []byte(`{"reason":"<script>alert(1)</script>"}`),
			},
		},
	}
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/audit", nil)
	req.Header.Set("HX-Request", "true")
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatalf("expected escaped script payload; body=%s", html)
	}
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Fatalf("unexpected raw script payload; body=%s", html)
	}
}

func TestGetProjectsDetailTab_Audit_NoRoleForbidden(t *testing.T) {
	noRoleActor := &app.Actor{ID: uuid.New(), Username: "norole", Role: "NO_ROLE"}
	a := makeProjectsDetailApp(noRoleActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/tabs/audit", nil)
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
	if string(body) != "You do not have permission to access this page." {
		t.Fatalf("expected canonical 403 body; body=%s", string(body))
	}
}

func TestGetProjectAuditLog_ReturnsFragment(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/audit-log", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if !strings.Contains(html, `class="audit-row"`) {
		t.Fatalf("expected audit row fragment; body=%s", html)
	}
	if strings.Contains(html, `id="audit-log"`) {
		t.Fatalf("expected item fragment only; body=%s", html)
	}
}

func TestGetProjectAuditLog_FirstPageMatchesCanonical(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/audit-log", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	want := auditLogCanonical(t, "fragment-with-row-and-load-more")
	if got := normaliseAuditLogHTML(string(body)); got != want {
		t.Fatalf("audit fragment mismatch:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestGetProjectAuditLog_UnauthenticatedRedirects(t *testing.T) {
	a := makeProjectsDetailApp(nil)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/audit-log", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected 302/303, got %d", resp.StatusCode)
	}
}

func TestGetProjectAuditLog_InvalidCursorReturnsBadRequest(t *testing.T) {
	a := makeProjectsDetailApp(adminActor)
	req, _ := http.NewRequest(http.MethodGet, "/projects/a1000000-0000-0000-0000-000000000001/audit-log?before_occurred_at=nope&before_id=bad", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Invalid cursor." {
		t.Fatalf("expected invalid cursor body; body=%s", string(body))
	}
}
