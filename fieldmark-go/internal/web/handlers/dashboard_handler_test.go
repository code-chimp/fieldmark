package handlers_test

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	"github.com/code-chimp/fieldmark-go/internal/app"
	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	"github.com/code-chimp/fieldmark-go/internal/web/handlers"
	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

func init() {
	auth.RegisterAction("dashboard.view",
		domain.RoleAdmin,
		domain.RoleComplianceOfficer,
		domain.RoleInspector,
		domain.RoleSiteSupervisor,
		domain.RoleExecutive,
	)
}

func makeDashboardApp(actor *app.Actor) *fiber.App {
	a := newTestApp()
	if actor != nil {
		a.Use(injectActor(actor))
	}
	h := &handlers.DashboardHandlers{Pool: nil}
	a.Get("/dashboard", auth.RequireAuth(), h.GetDashboard)
	return a
}

func TestDashboard_NoPermissionRole_Returns403(t *testing.T) {
	actor := &app.Actor{ID: uuid.New(), Username: "norole", Role: "NONEXISTENT_ROLE"}
	a := makeDashboardApp(actor)
	req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)

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

func TestDashboard_Unauthenticated_RedirectsToLogin(t *testing.T) {
	a := makeDashboardApp(nil)
	req, _ := http.NewRequest(http.MethodGet, "/dashboard", nil)

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

func TestDashboardTemplate_ContainsRenderedTileIdsAndRoleStatus(t *testing.T) {
	a := newTestApp()
	portfolio := components.NewComplianceTileArgs(nil, "Portfolio Compliance", "compliance-tile-portfolio")
	a.Get("/dashboard-render-test", func(c fiber.Ctx) error {
		return c.Render("pages/dashboard", fiber.Map{
			"PortfolioTile": portfolio,
			"OverdueTile": viewmodels.DashboardTileVM{TileID: "overdue-violations-tile", Label: "Overdue Violations", DisplayValue: "", ValueClass: " text-danger", RoleStatus: true},
			"ActiveProjectsTile": viewmodels.DashboardTileVM{TileID: "active-projects-tile", Label: "Active Projects", DisplayValue: "0", ValueClass: " text-info", RoleStatus: true},
			"InspectionsWeekTile": viewmodels.DashboardTileVM{TileID: "inspections-week-tile", Label: "Inspections This Week", DisplayValue: "0", ValueClass: " text-neutral", RoleStatus: true},
			"FmTheme": "system", "FmThemeNext": "light", "FmThemeResolved": "system",
		})
	})
	req, _ := http.NewRequest(http.MethodGet, "/dashboard-render-test", nil)
	resp, err := a.Test(req)
	if err != nil {
		t.Fatalf("test request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	for _, needle := range []string{
		`id="compliance-tile-portfolio" role="status"`,
		`id="overdue-violations-tile" role="status"`,
		`id="active-projects-tile" role="status"`,
		`id="inspections-week-tile" role="status"`,
		"grid-cols-1",
		"md:grid-cols-2",
		"xl:grid-cols-4",
		"data-grid-rowclick=\"navigate\"",
	} {
		if !strings.Contains(html, needle) {
			t.Errorf("expected dashboard template to contain %q", needle)
		}
	}
}

func TestIsoWeekBoundsUTC_PinsMondayStartAndNextMondayEnd(t *testing.T) {
	now := time.Date(2026, time.June, 3, 14, 30, 0, 0, time.UTC) // Wednesday
	start, end := handlers.IsoWeekBoundsUTCForTest(now)
	if start != time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected week start: %s", start)
	}
	if end != time.Date(2026, time.June, 8, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected week end: %s", end)
	}
}

func TestDashboardStatsFromRaw_EmptySetsMapToNil(t *testing.T) {
	stats := handlers.DashboardStatsFromRawForTest(nil, 0, 0, 0, 0, "", 0, 0)
	if stats.PortfolioScore != nil || stats.ActiveProjects != nil || stats.OverdueViolations != nil || stats.InspectionsWeek != nil {
		t.Fatalf("expected nil values for empty source sets")
	}
}

func TestDashboardStatsFromRaw_ExistingDataPreservesZeroCounts(t *testing.T) {
	portfolio := 90
	stats := handlers.DashboardStatsFromRawForTest(&portfolio, 5, 0, 4, 0, "", 3, 0)
	if stats.PortfolioScore == nil || *stats.PortfolioScore != 90 {
		t.Fatalf("expected portfolio score 90")
	}
	if stats.ActiveProjects == nil || *stats.ActiveProjects != 0 {
		t.Fatalf("expected active projects 0")
	}
	if stats.OverdueViolations == nil || *stats.OverdueViolations != 0 {
		t.Fatalf("expected overdue violations 0")
	}
	if stats.InspectionsWeek == nil || *stats.InspectionsWeek != 0 {
		t.Fatalf("expected inspections week 0")
	}
}
