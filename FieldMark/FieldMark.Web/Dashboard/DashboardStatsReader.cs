using System.Globalization;
using FieldMark.Data.Context;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Web.Dashboard;

public sealed record DashboardStats(
    int? PortfolioScore,
    int? OverdueViolations,
    string OverdueBreakdown,
    int? ActiveProjects,
    int? InspectionsThisWeek
);

public sealed class DashboardStatsReader(FieldMarkDbContext db)
{
    public async Task<DashboardStats> ReadAsync(DateTimeOffset nowUtc, CancellationToken ct)
    {
        var portfolioAvg = await db
            .Projects.AsNoTracking()
            .Where(p => p.Status != FieldMark.Domain.ValueObjects.ProjectStatus.Closed)
            .Select(p => (double?)p.ComplianceScore)
            .AverageAsync(ct);
        var portfolioScore = portfolioAvg is null ? null : (int?)Math.Round(portfolioAvg.Value);

        var projectCount = await db.Projects.AsNoTracking().CountAsync(ct);
        var activeCount = await db
            .Projects.AsNoTracking()
            .CountAsync(p => p.Status == FieldMark.Domain.ValueObjects.ProjectStatus.Active, ct);
        int? activeProjects = projectCount == 0 ? null : activeCount;

        var weekStart = StartOfIsoWeekUtc(nowUtc);
        var weekEnd = weekStart.AddDays(7);

        await db.Database.OpenConnectionAsync(ct);
        var conn = db.Database.GetDbConnection();
        try
        {

        var violationCount = await ScalarIntAsync(
            conn,
            "SELECT count(*) FROM domain.violation",
            ct
        );
        var overdueTotal = await ScalarIntAsync(
            conn,
            "SELECT count(*) FROM domain.violation WHERE status IN ('Open','InProgress') AND due_at < now()",
            ct
        );
        int? overdue = violationCount == 0 ? null : overdueTotal;

        var overdueBreakdown = await BuildOverdueBreakdownAsync(conn, ct);

        var inspectionCount = await ScalarIntAsync(
            conn,
            "SELECT count(*) FROM domain.inspection",
            ct
        );
        var weekCount = await ScalarIntAsync(
            conn,
            "SELECT count(*) FROM domain.inspection WHERE scheduled_for >= @start AND scheduled_for < @end",
            ct,
            ("@start", weekStart.UtcDateTime),
            ("@end", weekEnd.UtcDateTime)
        );
        int? inspectionsThisWeek = inspectionCount == 0 ? null : weekCount;

            return FromRaw(
                portfolioScore,
                projectCount,
                activeCount,
                violationCount,
                overdueTotal,
                overdueBreakdown,
                inspectionCount,
                weekCount
            );
        }
        finally
        {
            await db.Database.CloseConnectionAsync();
        }
    }

    public static DashboardStats FromRaw(
        int? portfolioScore,
        int projectCount,
        int activeCount,
        int violationCount,
        int overdueTotal,
        string overdueBreakdown,
        int inspectionCount,
        int weekCount
    )
    {
        int? activeProjects = projectCount == 0 ? null : activeCount;
        int? overdueViolations = violationCount == 0 ? null : overdueTotal;
        int? inspectionsThisWeek = inspectionCount == 0 ? null : weekCount;
        return new DashboardStats(
            PortfolioScore: portfolioScore,
            OverdueViolations: overdueViolations,
            OverdueBreakdown:
                overdueViolations == 0 || overdueViolations is null
                    ? string.Empty
                    : overdueBreakdown,
            ActiveProjects: activeProjects,
            InspectionsThisWeek: inspectionsThisWeek
        );
    }

    private static DateTimeOffset StartOfIsoWeekUtc(DateTimeOffset nowUtc)
    {
        var day = (int)nowUtc.UtcDateTime.DayOfWeek;
        var delta = day == 0 ? 6 : day - 1;
        var monday = nowUtc.UtcDateTime.Date.AddDays(-delta);
        return new DateTimeOffset(monday, TimeSpan.Zero);
    }

    private static async Task<int> ScalarIntAsync(
        System.Data.Common.DbConnection conn,
        string sql,
        CancellationToken ct,
        params (string Name, object Value)[] args
    )
    {
        await using var cmd = conn.CreateCommand();
        cmd.CommandText = sql;
        foreach (var (name, value) in args)
        {
            var p = cmd.CreateParameter();
            p.ParameterName = name;
            p.Value = value;
            cmd.Parameters.Add(p);
        }

        var valueObj = await cmd.ExecuteScalarAsync(ct);
        return Convert.ToInt32(valueObj, CultureInfo.InvariantCulture);
    }

    private static async Task<string> BuildOverdueBreakdownAsync(
        System.Data.Common.DbConnection conn,
        CancellationToken ct
    )
    {
        var counts = new Dictionary<string, int>(StringComparer.Ordinal)
        {
            ["Critical"] = 0,
            ["High"] = 0,
            ["Medium"] = 0,
            ["Low"] = 0,
        };

        await using var cmd = conn.CreateCommand();
        cmd.CommandText =
            "SELECT severity, count(*) FROM domain.violation WHERE status IN ('Open','InProgress') AND due_at < now() GROUP BY severity";
        await using var reader = await cmd.ExecuteReaderAsync(ct);
        while (await reader.ReadAsync(ct))
        {
            var severity = reader.GetString(0);
            var count = reader.GetInt32(1);
            if (counts.ContainsKey(severity))
            {
                counts[severity] = count;
            }
        }

        var parts = new List<string>();
        foreach (var key in new[] { "Critical", "High", "Medium", "Low" })
        {
            var value = counts[key];
            if (value > 0)
            {
                parts.Add($"{value} {key}");
            }
        }
        return string.Join(", ", parts);
    }
}
