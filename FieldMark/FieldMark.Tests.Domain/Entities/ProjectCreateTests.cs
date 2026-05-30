using FieldMark.Domain.Entities;
using FieldMark.Domain.ValueObjects;
using FluentAssertions;

namespace FieldMark.Tests.Domain.Entities;

public class ProjectCreateTests
{
    private static readonly DateOnly Today = new(2026, 6, 1);
    private static readonly Guid TradeId1 = Guid.NewGuid();
    private static readonly Guid TradeId2 = Guid.NewGuid();
    private static readonly Guid InspectorId = Guid.NewGuid();

    private static CreatedProject HappyPath(
        string code = "BLDG-A",
        string name = "Building A",
        string? description = null,
        DateOnly? startDate = null,
        DateOnly? targetCompletion = null,
        IReadOnlyList<Guid>? trades = null,
        IReadOnlyList<Guid>? inspectors = null
    ) =>
        Project.Create(
            code,
            name,
            description,
            startDate ?? Today,
            targetCompletion,
            trades ?? [TradeId1],
            inspectors ?? []
        );

    [Fact]
    public void Create_HappyPath_ReturnsProjectWithActiveStatus()
    {
        var result = HappyPath();
        result.Project.Status.Should().Be(ProjectStatus.Active);
    }

    [Fact]
    public void Create_HappyPath_GeneratesNewId()
    {
        var r1 = HappyPath();
        var r2 = HappyPath();
        r1.Project.Id.Should().NotBe(Guid.Empty);
        r1.Project.Id.Should().NotBe(r2.Project.Id);
    }

    [Fact]
    public void Create_HappyPath_ComplianceScoreIs100()
    {
        HappyPath().Project.ComplianceScore.Should().Be(100);
    }

    [Fact]
    public void Create_HappyPath_TrimsCodeAndName()
    {
        var result = HappyPath(code: "  BLDG-A  ", name: "  Building A  ");
        result.Project.Code.Should().Be("BLDG-A");
        result.Project.Name.Should().Be("Building A");
    }

    [Fact]
    public void Create_HappyPath_TradeScopes_MatchedToProjectId()
    {
        var result = HappyPath(trades: [TradeId1, TradeId2]);
        result.Scopes.Should().HaveCount(2);
        result.Scopes.Should().OnlyContain(s => s.ProjectId == result.Project.Id);
        result.Scopes.Select(s => s.TradeTypeId).Should().BeEquivalentTo(new[] { TradeId1, TradeId2 });
    }

    [Fact]
    public void Create_HappyPath_Inspectors_MatchedToProjectId()
    {
        var result = HappyPath(inspectors: [InspectorId]);
        result.Inspectors.Should().HaveCount(1);
        result.Inspectors[0].ProjectId.Should().Be(result.Project.Id);
        result.Inspectors[0].UserId.Should().Be(InspectorId);
    }

    [Fact]
    public void Create_EmptyInspectors_ReturnsEmptyInspectorArray()
    {
        HappyPath(inspectors: []).Inspectors.Should().BeEmpty();
    }

    [Fact]
    public void Create_NullDescription_StoresNull()
    {
        HappyPath(description: null).Project.Description.Should().BeNull();
    }

    [Fact]
    public void Create_WhitespaceDescription_StoresNull()
    {
        HappyPath(description: "   ").Project.Description.Should().BeNull();
    }

    [Fact]
    public void Create_NonNullDescription_TrimsAndStores()
    {
        HappyPath(description: "  hello  ").Project.Description.Should().Be("hello");
    }

    [Fact]
    public void Create_EmptyCode_ThrowsArgumentException()
    {
        var act = () => HappyPath(code: "");
        act.Should().Throw<ArgumentException>().WithParameterName("code");
    }

    [Fact]
    public void Create_WhitespaceCode_ThrowsArgumentException()
    {
        var act = () => HappyPath(code: "   ");
        act.Should().Throw<ArgumentException>().WithParameterName("code");
    }

    [Fact]
    public void Create_EmptyName_ThrowsArgumentException()
    {
        var act = () => HappyPath(name: "");
        act.Should().Throw<ArgumentException>().WithParameterName("name");
    }

    [Fact]
    public void Create_EmptyTradeScopeIds_ThrowsArgumentOutOfRangeException()
    {
        var act = () => HappyPath(trades: []);
        act.Should().Throw<ArgumentOutOfRangeException>().WithParameterName("tradeScopeIds");
    }

    [Fact]
    public void Create_TargetBeforeStart_ThrowsArgumentOutOfRangeException()
    {
        var act = () => HappyPath(
            startDate: new DateOnly(2026, 6, 1),
            targetCompletion: new DateOnly(2026, 5, 1)
        );
        act.Should().Throw<ArgumentOutOfRangeException>().WithParameterName("targetCompletionDate");
    }

    [Fact]
    public void Create_TargetEqualsStart_Succeeds()
    {
        var act = () => HappyPath(
            startDate: new DateOnly(2026, 6, 1),
            targetCompletion: new DateOnly(2026, 6, 1)
        );
        act.Should().NotThrow();
    }
}
