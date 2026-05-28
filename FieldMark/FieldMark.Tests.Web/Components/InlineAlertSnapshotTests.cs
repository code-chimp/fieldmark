using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class InlineAlertSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    public static TheoryData<string, string> Variants =>
        new()
        {
            { "danger", "danger" },
            { "warning", "warning" },
            { "info", "info" },
            { "success", "success" },
            { "unknown", "notice" },
        };

    private static ExpandoObject AlertModel(
        string severity,
        string title = "Action blocked",
        string message = "Resolve open violations before closing.",
        string meta = "Project PM-104"
    ) => Model(("Severity", severity), ("Title", title), ("Message", message), ("Meta", meta));

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task InlineAlertVariantMatchesCanonical(string variant, string severity)
    {
        await AssertSnapshot("inline_alert", "InlineAlert", variant, AlertModel(severity));
    }

    [Fact]
    public async Task InlineAlertEscapesUserStrings()
    {
        var html = await RenderPartial(
            "InlineAlert",
            AlertModel(
                "danger",
                "<script>alert(1)</script>",
                "<script>alert(1)</script>",
                "<script>alert(1)</script>"
            )
        );
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>alert(1)</script>");
    }

    [Fact]
    public async Task InlineAlertUnknownFallbackClass()
    {
        var html = await RenderPartial("InlineAlert", AlertModel("notice"));
        html.Should().Contain("alert-unknown");
    }

    [Fact]
    public void InlineAlertTemplateDoesNotUseHtmlRaw()
    {
        File.ReadAllText(
                RepoPath(
                    "FieldMark",
                    "FieldMark.Web",
                    "Pages",
                    "Shared",
                    "Components",
                    "_InlineAlert.cshtml"
                )
            )
            .Should()
            .NotContain("Html.Raw");
    }
}
