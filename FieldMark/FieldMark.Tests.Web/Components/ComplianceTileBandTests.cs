using FieldMark.Tests.Web.Fixtures;
using FieldMark.Web.ViewModels.Components;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

/// <summary>
/// Unit tests for ComplianceTileViewModel.ResolveBand — direct tuple assertions per AC3.
/// Each row names the boundary it exercises so a failure surfaces the exact band edge.
/// Synchronous; does not render HTML.
/// </summary>
[Collection(AuthTests.Name)]
public sealed class ComplianceTileBandTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    [Theory]
    [InlineData(null, "text-neutral", "", "", false)]
    [InlineData(100, "text-success", "Healthy", "text-success", true)]
    [InlineData(90, "text-success", "Healthy", "text-success", true)]
    [InlineData(89, "text-warning", "Watch", "text-warning", true)]
    [InlineData(70, "text-warning", "Watch", "text-warning", true)]
    [InlineData(69, "text-warning-strong", "Concern", "text-warning-strong", true)]
    [InlineData(50, "text-warning-strong", "Concern", "text-warning-strong", true)]
    [InlineData(49, "text-danger", "Critical", "text-danger", true)]
    [InlineData(0, "text-danger", "Critical", "text-danger", true)]
    [InlineData(-1, "text-neutral", "", "", false)]
    [InlineData(101, "text-neutral", "", "", false)]
    public void ResolveBandReturnsTupleForScore(
        int? score,
        string expectedValueClass,
        string expectedThresholdWord,
        string expectedThresholdClass,
        bool expectedRenderP
    )
    {
        var (valueClass, thresholdWord, thresholdClass, renderP) =
            ComplianceTileViewModel.ResolveBand(score);

        var label = score?.ToString(System.Globalization.CultureInfo.InvariantCulture) ?? "null";
        valueClass.Should().Be(expectedValueClass, $"score={label} value class");
        thresholdWord.Should().Be(expectedThresholdWord, $"score={label} threshold word");
        thresholdClass.Should().Be(expectedThresholdClass, $"score={label} threshold class");
        renderP.Should().Be(expectedRenderP, $"score={label} renderP");
    }

    [Fact]
    public void DisplayValueUsesInvariantCulture()
    {
        var vm = new ComplianceTileViewModel(Score: 95, Label: "Compliance", Id: "compliance-tile");
        vm.DisplayValue.Should().Be("95");
    }
}
