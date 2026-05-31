package handlers

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/code-chimp/fieldmark-go/internal/domain"
	"github.com/code-chimp/fieldmark-go/internal/web/auth"
	components "github.com/code-chimp/fieldmark-go/internal/web/templates/components"
	"github.com/code-chimp/fieldmark-go/internal/web/viewmodels"
)

type DashboardHandlers struct {
	Pool *pgxpool.Pool
	Now  func() time.Time
}

func (h *DashboardHandlers) GetDashboard(c fiber.Ctx) error {
	actor := auth.ActorFromCtx(c)
	if !auth.Can(actor, "dashboard.view", uuid.Nil) {
		c.Status(fiber.StatusForbidden)
		return c.SendString("You do not have permission to access this page.")
	}
	stats, err := h.readStats(c)
	if err != nil {
		return err
	}
	role := domain.Role(actor.Role)
	theme, next := themeEntries(c)
	return c.Render("pages/dashboard", fiber.Map{
		"Title":             "Compliance Dashboard",
		"FmTheme":           theme,
		"FmThemeNext":       next,
		"FmThemeResolved":   theme,
		"Actor":             actor,
		"RoleLabel":         role.Label(),
		"RoleBadgeToken":    role.BadgeToken(),
		"FullName":          actor.DisplayName,
		"Initials":          viewmodels.Initials(actor.DisplayName, actor.Username),
		"PortfolioTile": components.NewComplianceTileArgs(stats.PortfolioScore, "Portfolio Compliance", "compliance-tile-portfolio"),
		"OverdueTile": viewmodels.DashboardTileVM{
			TileID:       "overdue-violations-tile",
			Label:        "Overdue Violations",
			DisplayValue: intPtrToString(stats.OverdueViolations),
			Secondary:    stats.OverdueBreakdown,
			ValueClass:   " text-danger",
			RoleStatus:   true,
		},
		"ActiveProjectsTile": viewmodels.DashboardTileVM{
			TileID:       "active-projects-tile",
			Label:        "Active Projects",
			DisplayValue: intPtrToString(stats.ActiveProjects),
			ValueClass:   " text-info",
			RoleStatus:   true,
		},
		"InspectionsWeekTile": viewmodels.DashboardTileVM{
			TileID:       "inspections-week-tile",
			Label:        "Inspections This Week",
			DisplayValue: intPtrToString(stats.InspectionsWeek),
			ValueClass:   " text-neutral",
			RoleStatus:   true,
		},
	})
}

type dashboardStats struct {
	PortfolioScore    *int
	OverdueViolations *int
	OverdueBreakdown  string
	ActiveProjects    *int
	InspectionsWeek   *int
}

func dashboardStatsFromRaw(
	portfolioScore *int,
	projectCount int,
	activeCount int,
	violationCount int,
	overdueTotal int,
	overdueBreakdown string,
	inspectionCount int,
	weekCount int,
) dashboardStats {
	var activeProjects *int
	if projectCount > 0 {
		activeProjects = &activeCount
	}
	var overdueViolations *int
	if violationCount > 0 {
		overdueViolations = &overdueTotal
	}
	var inspectionsWeek *int
	if inspectionCount > 0 {
		inspectionsWeek = &weekCount
	}
	if overdueViolations == nil || *overdueViolations == 0 {
		overdueBreakdown = ""
	}
	return dashboardStats{
		PortfolioScore:    portfolioScore,
		OverdueViolations: overdueViolations,
		OverdueBreakdown:  overdueBreakdown,
		ActiveProjects:    activeProjects,
		InspectionsWeek:   inspectionsWeek,
	}
}

func intPtrToString(v *int) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%d", *v)
}

func (h *DashboardHandlers) readStats(c fiber.Ctx) (dashboardStats, error) {
	if h.Pool == nil {
		return dashboardStats{}, nil
	}
	ctx := c.Context()
	var rawPortfolio *float64
	if err := h.Pool.QueryRow(ctx, `SELECT AVG(compliance_score)::float8 FROM domain.project WHERE status <> 'Closed'`).Scan(&rawPortfolio); err != nil {
		return dashboardStats{}, err
	}
	var portfolioScore *int
	if rawPortfolio != nil {
		v := int(math.RoundToEven(*rawPortfolio))
		portfolioScore = &v
	}

	var projectCount, activeCount int
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM domain.project`).Scan(&projectCount); err != nil {
		return dashboardStats{}, err
	}
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM domain.project WHERE status = 'Active'`).Scan(&activeCount); err != nil {
		return dashboardStats{}, err
	}

	var violationCount int
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM domain.violation`).Scan(&violationCount); err != nil {
		return dashboardStats{}, err
	}
	rows, err := h.Pool.Query(ctx, `SELECT severity, count(*) FROM domain.violation WHERE status IN ('Open','InProgress') AND due_at < now() GROUP BY severity`)
	if err != nil {
		return dashboardStats{}, err
	}
	defer rows.Close()
	counts := map[string]int{"Critical": 0, "High": 0, "Medium": 0, "Low": 0}
	overdueTotal := 0
	for rows.Next() {
		var sev string
		var n int
		if err := rows.Scan(&sev, &n); err != nil {
			return dashboardStats{}, err
		}
		counts[sev] = n
		overdueTotal += n
	}
	if err := rows.Err(); err != nil {
		return dashboardStats{}, err
	}
	parts := make([]string, 0)
	for _, sev := range []string{"Critical", "High", "Medium", "Low"} {
		if n := counts[sev]; n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, sev))
		}
	}
	overdueBreakdown := ""
	if violationCount > 0 && overdueTotal > 0 {
		overdueBreakdown = strings.Join(parts, ", ")
	}

	var inspectionCount int
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM domain.inspection`).Scan(&inspectionCount); err != nil {
		return dashboardStats{}, err
	}
	nowFn := h.Now
	if nowFn == nil {
		nowFn = time.Now
	}
	weekStart, weekEnd := isoWeekBoundsUTC(nowFn())
	var weekCount int
	if err := h.Pool.QueryRow(ctx, `SELECT count(*) FROM domain.inspection WHERE scheduled_for >= $1 AND scheduled_for < $2`, weekStart, weekEnd).Scan(&weekCount); err != nil {
		return dashboardStats{}, err
	}
	return dashboardStatsFromRaw(
		portfolioScore,
		projectCount,
		activeCount,
		violationCount,
		overdueTotal,
		overdueBreakdown,
		inspectionCount,
		weekCount,
	), nil
}

func isoWeekBoundsUTC(now time.Time) (time.Time, time.Time) {
	utc := now.UTC()
	delta := int(utc.Weekday()) - 1
	if utc.Weekday() == time.Sunday {
		delta = 6
	}
	weekStart := time.Date(utc.Year(), utc.Month(), utc.Day()-delta, 0, 0, 0, 0, time.UTC)
	return weekStart, weekStart.Add(7 * 24 * time.Hour)
}

// IsoWeekBoundsUTCForTest exposes week-bound computation for black-box tests.
func IsoWeekBoundsUTCForTest(now time.Time) (time.Time, time.Time) {
	return isoWeekBoundsUTC(now)
}

// DashboardStatsFromRawForTest exposes null-vs-zero mapping for tests.
func DashboardStatsFromRawForTest(
	portfolioScore *int,
	projectCount int,
	activeCount int,
	violationCount int,
	overdueTotal int,
	overdueBreakdown string,
	inspectionCount int,
	weekCount int,
) dashboardStats {
	return dashboardStatsFromRaw(
		portfolioScore,
		projectCount,
		activeCount,
		violationCount,
		overdueTotal,
		overdueBreakdown,
		inspectionCount,
		weekCount,
	)
}
