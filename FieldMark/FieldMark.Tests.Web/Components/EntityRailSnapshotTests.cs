using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class EntityRailSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    private static ExpandoObject EmptyModel(
        string id = "violation-detail",
        string entityTypeLabel = "Violation"
    ) =>
        Model(
            ("Id", id),
            ("EntityTypeLabel", entityTypeLabel),
            ("EntityLoaded", false),
            ("BodySlot", null),
            ("FooterSlot", null)
        );

    private static ExpandoObject LoadedModel(
        string id = "violation-detail",
        string entityTypeLabel = "Violation",
        string? bodySlot = "__BODY__",
        string? footerSlot = "__FOOTER__"
    ) =>
        Model(
            ("Id", id),
            ("EntityTypeLabel", entityTypeLabel),
            ("EntityLoaded", true),
            ("BodySlot", bodySlot),
            ("FooterSlot", footerSlot)
        );

    public static TheoryData<string, object> Variants =>
        new()
        {
            { "empty-violation", EmptyModel() },
            {
                "empty-inspection",
                EmptyModel(id: "inspection-detail", entityTypeLabel: "Inspection")
            },
            {
                "empty-corrective-action",
                EmptyModel(id: "corrective-action-detail", entityTypeLabel: "Corrective Action")
            },
            { "loaded-shell-violation", LoadedModel() },
            {
                "loaded-shell-inspection",
                LoadedModel(id: "inspection-detail", entityTypeLabel: "Inspection")
            },
            {
                "loaded-shell-corrective-action",
                LoadedModel(id: "corrective-action-detail", entityTypeLabel: "Corrective Action")
            },
        };

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task EntityRailVariantMatchesCanonical(string variant, object model)
    {
        await AssertSnapshot("entity_rail", "EntityRail", variant, model);
    }

    // AC4 — four-case slot/footer-omission coverage
    [Fact]
    public async Task EmptyState_RendersEmptyStateCard()
    {
        var html = await RenderPartial("EntityRail", EmptyModel());
        html.Should().Contain("entity-rail--empty");
        html.Should().Contain("Empty entity rail");
        html.Should().Contain("Select an entity to see its detail here.");
        html.Should().NotContain("entity-rail__body");
    }

    [Fact]
    public async Task LoadedWithBothSlots_RendersBodyAndFooter()
    {
        var html = await RenderPartial(
            "EntityRail",
            LoadedModel(bodySlot: "<p>body</p>", footerSlot: "<button>Save</button>")
        );
        html.Should().Contain("entity-rail--loaded");
        html.Should().Contain("<p>body</p>");
        html.Should().Contain("<button>Save</button>");
        html.Should().Contain("entity-rail__footer");
    }

    [Fact]
    public async Task LoadedWithBodyOnly_OmitsFooterDiv()
    {
        var html = await RenderPartial(
            "EntityRail",
            LoadedModel(bodySlot: "<p>body</p>", footerSlot: null)
        );
        html.Should().Contain("entity-rail__body");
        html.Should().NotContain("entity-rail__footer");
    }

    [Fact]
    public async Task LoadedWithEmptyStringFooter_OmitsFooterDiv()
    {
        // Empty-string footer must be treated the same as null — cross-stack parity with Django/Go.
        var html = await RenderPartial(
            "EntityRail",
            LoadedModel(bodySlot: "<p>body</p>", footerSlot: "")
        );
        html.Should().Contain("entity-rail__body");
        html.Should().NotContain("entity-rail__footer");
    }

    [Fact]
    public async Task LoadedWithNoSlots_RendersHeaderAndEmptyBody()
    {
        var html = await RenderPartial("EntityRail", LoadedModel(bodySlot: null, footerSlot: null));
        html.Should().Contain("entity-rail--loaded");
        html.Should().Contain("entity-rail__header");
        html.Should().Contain("entity-rail__body");
        html.Should().NotContain("entity-rail__footer");
    }

    // AC8 — XSS round-trip: entity_type_label is framework-escaped (non-slot prop)
    [Fact]
    public async Task XssPayloadInEntityTypeLabelIsEscaped()
    {
        var html = await RenderPartial(
            "EntityRail",
            EmptyModel(entityTypeLabel: "<script>alert(1)</script>")
        );
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt; detail");
        html.Should().NotContain("<script>alert(1)</script>");
        html.Should().NotContain("<script>");
    }

    [Fact]
    public async Task XssPayloadInLoadedLabelEscapedInSpanAndAriaLabel()
    {
        var html = await RenderPartial(
            "EntityRail",
            LoadedModel(entityTypeLabel: "<script>alert(1)</script>")
        );
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>");
    }

    // AC8 §category 9 — empty/whitespace entity_type_label does not crash
    [Theory]
    [InlineData("")]
    [InlineData("   ")]
    public async Task WhitespaceOrEmptyLabelDoesNotCrash(string label)
    {
        var html = await RenderPartial("EntityRail", EmptyModel(entityTypeLabel: label));
        html.Should().Contain("<aside");
        html.Should().Contain("entity-rail");
    }

    // AC3 — no HTMX producer attributes on dismiss button
    [Fact]
    public async Task LoadedShell_DismissButtonHasNoHtmxAttributes()
    {
        var html = await RenderPartial("EntityRail", LoadedModel());
        foreach (
            var forbidden in new[]
            {
                "hx-get",
                "hx-post",
                "hx-target",
                "hx-swap",
                "hx-trigger",
                "onclick=",
            }
        )
        {
            html.Should().NotContain(forbidden, $"dismiss button must not emit {forbidden}");
        }
    }

    // [AI-Review] Guard: wrapper must not throw when caller omits optional slot keys
    [Fact]
    public async Task EmptyStateWithoutSlotKeys_DoesNotThrow()
    {
        // Pass a dynamic model without BodySlot/FooterSlot keys; must not throw RuntimeBinderException.
        var modelWithoutSlots = Model(
            ("Id", "violation-detail"),
            ("EntityTypeLabel", "Violation"),
            ("EntityLoaded", false)
        );
        var html = await RenderPartial("EntityRail", modelWithoutSlots);
        html.Should().Contain("<aside");
        html.Should().Contain("entity-rail--empty");
        html.Should().NotContain("entity-rail__body");
    }

    // AC9 — scoped grep guard: exactly two Html.Raw in _EntityRail.cshtml
    [Fact]
    public void EntityRailTemplateHasExactlyTwoHtmlRawOccurrences()
    {
        var content = File.ReadAllText(
            RepoPath(
                "FieldMark",
                "FieldMark.Web",
                "Pages",
                "Shared",
                "Components",
                "_EntityRail.cshtml"
            )
        );
        var count = System.Text.RegularExpressions.Regex.Count(content, @"Html\.Raw");
        count
            .Should()
            .Be(
                2,
                "exactly two Html.Raw calls are permitted in _EntityRail.cshtml — one for body slot, one for footer slot"
            );
    }

    // Verify other component wrappers still have zero Html.Raw
    [Theory]
    [InlineData("_StatusBadge.cshtml")]
    [InlineData("_InlineAlert.cshtml")]
    [InlineData("_AuditRow.cshtml")]
    [InlineData("_DashboardTile.cshtml")]
    [InlineData("_ComplianceTile.cshtml")]
    public void OtherComponentWrappersDoNotUseHtmlRaw(string filename)
    {
        var content = File.ReadAllText(
            RepoPath("FieldMark", "FieldMark.Web", "Pages", "Shared", "Components", filename)
        );
        content.Should().NotContain("Html.Raw", $"{filename} must not use Html.Raw");
    }
}
