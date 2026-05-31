using FieldMark.Web.Dashboard;
using FluentAssertions;

namespace FieldMark.Tests.Web.Pages;

public sealed class DashboardStatsReaderTests
{
    [Fact]
    public void FromRaw_EmptySets_RenderEmDashViaNulls()
    {
        var stats = DashboardStatsReader.FromRaw(
            portfolioScore: null,
            projectCount: 0,
            activeCount: 0,
            violationCount: 0,
            overdueTotal: 0,
            overdueBreakdown: "",
            inspectionCount: 0,
            weekCount: 0
        );

        stats.PortfolioScore.Should().BeNull();
        stats.ActiveProjects.Should().BeNull();
        stats.OverdueViolations.Should().BeNull();
        stats.InspectionsThisWeek.Should().BeNull();
    }

    [Fact]
    public void FromRaw_DataExistsButCountsZero_PreservesZero()
    {
        var stats = DashboardStatsReader.FromRaw(
            portfolioScore: 73,
            projectCount: 3,
            activeCount: 0,
            violationCount: 2,
            overdueTotal: 0,
            overdueBreakdown: "",
            inspectionCount: 4,
            weekCount: 0
        );

        stats.PortfolioScore.Should().Be(73);
        stats.ActiveProjects.Should().Be(0);
        stats.OverdueViolations.Should().Be(0);
        stats.InspectionsThisWeek.Should().Be(0);
    }
}
