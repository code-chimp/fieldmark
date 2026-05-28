using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class AuditRowSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    private static ExpandoObject RowModel(
        string action = "ProjectCreated",
        string actor = "Aisha Stone",
        string json = "",
        bool expanded = false
    ) =>
        Model(
            ("Action", action),
            ("ActorName", actor),
            ("OccurredAt", "2026-05-28T14:20:01Z"),
            ("Absolute", "2026-05-28 14:20:01 UTC"),
            ("Relative", "3 minutes ago"),
            ("BeforeAfterJson", json),
            ("Expanded", expanded)
        );

    public static TheoryData<string, object> Variants =>
        new()
        {
            { "default", RowModel() },
            {
                "with-disclosure-collapsed",
                RowModel(json: """{"after":{"status":"ACTIVE"},"before":{"status":"DRAFT"}}""")
            },
            {
                "with-disclosure-expanded",
                RowModel(
                    json: """{"after":{"status":"ACTIVE"},"before":{"status":"DRAFT"}}""",
                    expanded: true
                )
            },
            { "unknown-action", RowModel(action: "UnknownAction") },
            { "empty-actor", RowModel(actor: "") },
        };

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task AuditRowVariantMatchesCanonical(string variant, object model)
    {
        await AssertSnapshot("audit_row", "AuditRow", variant, model);
    }

    [Fact]
    public async Task AuditRowWhitespaceOnlyActorMatchesEmptyActorCanonical()
    {
        await AssertSnapshot("audit_row", "AuditRow", "empty-actor", RowModel(actor: "   "));
    }

    [Fact]
    public async Task AuditRowUnknownActionFallbackClass()
    {
        var html = await RenderPartial("AuditRow", RowModel(action: "UnknownAction"));
        html.Should().Contain("badge-unknown");
    }

    [Fact]
    public async Task AuditRowEscapesJsonText()
    {
        var html = await RenderPartial(
            "AuditRow",
            RowModel(json: """<script>alert(1)</script>""", expanded: true)
        );
        html.Should().Contain("""&lt;script&gt;alert(1)&lt;/script&gt;""");
        html.Should().NotContain("""<script>alert(1)</script>""");
    }

    [Fact]
    public void AuditRowTemplateDoesNotUseHtmlRaw()
    {
        File.ReadAllText(
                RepoPath(
                    "FieldMark",
                    "FieldMark.Web",
                    "Pages",
                    "Shared",
                    "Components",
                    "_AuditRow.cshtml"
                )
            )
            .Should()
            .NotContain("Html.Raw");
    }
}
