using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class ComplianceTileSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    private static ExpandoObject TileModel(
        int? score = 95,
        string label = "Compliance",
        string id = "compliance-tile"
    ) => Model(("Score", score), ("Label", label), ("Id", id));

    public static TheoryData<string, object> Variants =>
        new()
        {
            { "healthy-project", TileModel(score: 95) },
            { "watch-project", TileModel(score: 82) },
            { "concern-project", TileModel(score: 58) },
            { "critical-project", TileModel(score: 37) },
            {
                "healthy-portfolio",
                TileModel(score: 91, label: "Portfolio Compliance", id: "compliance-tile-portfolio")
            },
            {
                "critical-portfolio",
                TileModel(score: 42, label: "Portfolio Compliance", id: "compliance-tile-portfolio")
            },
            { "no-data-project", TileModel(score: null) },
            { "boundary-90", TileModel(score: 90) },
            { "boundary-70", TileModel(score: 70) },
            { "boundary-50", TileModel(score: 50) },
            { "boundary-49", TileModel(score: 49) },
        };

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task ComplianceTileVariantMatchesCanonical(string variant, object model)
    {
        await AssertSnapshot("compliance_tile", "ComplianceTile", variant, model);
    }

    [Fact]
    public async Task Score0RendersAsCriticalNotNoData()
    {
        var html = await RenderPartial("ComplianceTile", TileModel(score: 0));
        html.Should().Contain("text-danger");
        html.Should().Contain("Critical");
        html.Should().NotContain("—");
    }

    [Fact]
    public async Task PortfolioIdPassedThroughVerbatim()
    {
        await AssertSnapshot(
            "compliance_tile",
            "ComplianceTile",
            "healthy-portfolio",
            TileModel(score: 91, label: "Portfolio Compliance", id: "compliance-tile-portfolio")
        );
    }

    [Fact]
    public async Task XssPayloadInLabelIsEscaped()
    {
        var html = await RenderPartial(
            "ComplianceTile",
            TileModel(score: 95, label: "<script>alert(1)</script>")
        );
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>alert(1)</script>");
        html.Should().NotContain("<script>"); // generic check: no raw script tag regardless of payload
    }

    [Fact]
    public async Task WhitespaceOnlyLabelDoesNotCrash()
    {
        var html = await RenderPartial("ComplianceTile", TileModel(score: 95, label: "   "));
        html.Should().Contain("<section");
        html.Should().Contain("compliance-tile__label");
    }

    [Fact]
    public async Task EmptyLabelDoesNotCrash()
    {
        var html = await RenderPartial("ComplianceTile", TileModel(score: 95, label: ""));
        html.Should().Contain("<section");
    }

    [Fact]
    public async Task TargetShapeAttributesPresent()
    {
        var html = await RenderPartial("ComplianceTile", TileModel());
        html.Should().Contain("id=\"compliance-tile\"");
        html.Should().Contain("role=\"status\"");
        html.Should().Contain("aria-live=\"polite\"");
        html.Should().Contain("aria-atomic=\"true\"");
        html.Should().Contain("class=\"compliance-tile\"");
    }

    [Fact]
    public async Task NoHtmxProducerAttributesEmitted()
    {
        var html = await RenderPartial("ComplianceTile", TileModel());
        html.Should().NotContain("hx-get");
        html.Should().NotContain("hx-post");
        html.Should().NotContain("hx-target");
        html.Should().NotContain("hx-swap");
        html.Should().NotContain("hx-trigger");
        html.Should().NotContain("<script");
        html.Should().NotContain("onload=");
        html.Should().NotContain("data-htmx-");
    }

    [Fact]
    public void ComplianceTileTemplateDoesNotUseHtmlRaw()
    {
        File.ReadAllText(
                RepoPath(
                    "FieldMark",
                    "FieldMark.Web",
                    "Pages",
                    "Shared",
                    "Components",
                    "_ComplianceTile.cshtml"
                )
            )
            .Should()
            .NotContain("Html.Raw");
    }
}
