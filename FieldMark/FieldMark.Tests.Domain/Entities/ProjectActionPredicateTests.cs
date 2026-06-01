using System.Reflection;
using FieldMark.Domain.Entities;
using FieldMark.Domain.ValueObjects;
using FluentAssertions;

namespace FieldMark.Tests.Domain.Entities;

public class ProjectActionPredicateTests
{
    private static readonly DateOnly Today = new(2026, 6, 1);
    private static readonly Guid TradeId = Guid.NewGuid();

    [Theory]
    [InlineData(ProjectStatus.Active, true, false, true)]
    [InlineData(ProjectStatus.OnHold, false, true, false)]
    [InlineData(ProjectStatus.Closed, false, false, false)]
    public void ActionPredicates_MatchStatusGates(
        ProjectStatus status,
        bool canPlaceOnHold,
        bool canResume,
        bool canClose
    )
    {
        var created = Project.Create("P-1", "Project", null, Today, null, [TradeId], []);
        var project = created.Project;

        typeof(Project)
            .GetProperty(nameof(Project.Status), BindingFlags.Instance | BindingFlags.Public)!
            .SetValue(project, status);

        project.CanPlaceOnHold().Should().Be(canPlaceOnHold);
        project.CanResume().Should().Be(canResume);
        project.CanClose().Should().Be(canClose);
    }
}
