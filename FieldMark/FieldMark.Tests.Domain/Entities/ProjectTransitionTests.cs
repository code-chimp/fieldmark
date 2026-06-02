using System.Reflection;
using FieldMark.Domain.Entities;
using FieldMark.Domain.Exceptions;
using FieldMark.Domain.ValueObjects;
using FluentAssertions;

namespace FieldMark.Tests.Domain.Entities;

public class ProjectTransitionTests
{
    private static readonly DateOnly Today = new(2026, 6, 1);
    private static readonly Guid TradeId = Guid.NewGuid();

    private static Project Build(ProjectStatus status)
    {
        var project = Project.Create("P-1", "Project", null, Today, null, [TradeId], []).Project;
        typeof(Project)
            .GetProperty(nameof(Project.Status), BindingFlags.Instance | BindingFlags.Public)!
            .SetValue(project, status);
        return project;
    }

    [Fact]
    public void PlaceOnHold_Active_ToOnHold()
    {
        var project = Build(ProjectStatus.Active);
        project.PlaceOnHold("maintenance window");
        project.Status.Should().Be(ProjectStatus.OnHold);
    }

    [Theory]
    [InlineData(ProjectStatus.OnHold)]
    [InlineData(ProjectStatus.Closed)]
    public void PlaceOnHold_NonActive_Throws(ProjectStatus status)
    {
        var project = Build(status);
        var act = () => project.PlaceOnHold("maintenance window");
        act.Should()
            .Throw<InvalidProjectTransitionException>()
            .WithMessage("Project is already on hold");
    }

    [Fact]
    public void Resume_OnHold_ToActive()
    {
        var project = Build(ProjectStatus.OnHold);
        project.Resume("back online");
        project.Status.Should().Be(ProjectStatus.Active);
    }

    [Theory]
    [InlineData(ProjectStatus.Active)]
    [InlineData(ProjectStatus.Closed)]
    public void Resume_NonOnHold_Throws(ProjectStatus status)
    {
        var project = Build(status);
        var act = () => project.Resume("back online");
        act.Should()
            .Throw<InvalidProjectTransitionException>()
            .WithMessage("Project is not on hold");
    }
}
