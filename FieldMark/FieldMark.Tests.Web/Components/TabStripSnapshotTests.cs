using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class TabStripSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    private static ExpandoObject Tab(
        string id,
        string label,
        string hxGet,
        string hxTarget,
        int? badgeCount = null
    )
    {
        IDictionary<string, object?> t = new ExpandoObject();
        t["Id"] = id;
        t["Label"] = label;
        t["HxGet"] = hxGet;
        t["HxTarget"] = hxTarget;
        t["BadgeCount"] = (object?)badgeCount;
        return (ExpandoObject)t;
    }

    private static readonly List<ExpandoObject> ProjectDetailTabs =
    [
        Tab("tab-summary", "Summary", "/projects/__ID__/summary", "#project-detail-tab-content"),
        Tab(
            "tab-inspections",
            "Inspections",
            "/projects/__ID__/inspections",
            "#project-detail-tab-content"
        ),
        Tab(
            "tab-violations",
            "Violations",
            "/projects/__ID__/violations",
            "#project-detail-tab-content"
        ),
        Tab("tab-audit", "Audit", "/projects/__ID__/audit", "#project-detail-tab-content"),
    ];

    private static ExpandoObject StripModel(
        string? id,
        string ariaLabel,
        int activeIndex,
        List<ExpandoObject> tabs
    ) => Model(("Id", (object?)id), ("AriaLabel", ariaLabel), ("ActiveIndex", activeIndex), ("Tabs", tabs));

    public static TheoryData<string, object> Variants =>
        new()
        {
            {
                "project-detail-four-tabs-summary-active",
                StripModel(
                    "project-detail-tabstrip",
                    "Project Detail Tabs",
                    0,
                    ProjectDetailTabs
                )
            },
            {
                "project-detail-four-tabs-violations-active",
                StripModel(
                    "project-detail-tabstrip",
                    "Project Detail Tabs",
                    2,
                    ProjectDetailTabs
                )
            },
            {
                "project-detail-four-tabs-with-badges",
                StripModel(
                    "project-detail-tabstrip",
                    "Project Detail Tabs",
                    0,
                    [
                        Tab(
                            "tab-summary",
                            "Summary",
                            "/projects/__ID__/summary",
                            "#project-detail-tab-content"
                        ),
                        Tab(
                            "tab-inspections",
                            "Inspections",
                            "/projects/__ID__/inspections",
                            "#project-detail-tab-content",
                            12
                        ),
                        Tab(
                            "tab-violations",
                            "Violations",
                            "/projects/__ID__/violations",
                            "#project-detail-tab-content",
                            3
                        ),
                        Tab(
                            "tab-audit",
                            "Audit",
                            "/projects/__ID__/audit",
                            "#project-detail-tab-content",
                            147
                        ),
                    ]
                )
            },
            {
                "two-tabs-minimal",
                StripModel(
                    "two-tabs-strip",
                    "Open Closed Tabs",
                    0,
                    [
                        Tab("tab-open", "Open", "/__tab__/open", "#__panel__"),
                        Tab("tab-closed", "Closed", "/__tab__/closed", "#__panel__"),
                    ]
                )
            },
            {
                "single-tab",
                StripModel(
                    "single-tab-strip",
                    "Single Tab",
                    0,
                    [Tab("tab-only", "Only Tab", "/__tab__/only", "#__panel__")]
                )
            },
            {
                "badge-zero",
                StripModel(
                    "project-detail-tabstrip",
                    "Project Detail Tabs",
                    0,
                    [
                        Tab(
                            "tab-summary",
                            "Summary",
                            "/projects/__ID__/summary",
                            "#project-detail-tab-content"
                        ),
                        Tab(
                            "tab-inspections",
                            "Inspections",
                            "/projects/__ID__/inspections",
                            "#project-detail-tab-content",
                            12
                        ),
                        Tab(
                            "tab-violations",
                            "Violations",
                            "/projects/__ID__/violations",
                            "#project-detail-tab-content",
                            0
                        ),
                        Tab(
                            "tab-audit",
                            "Audit",
                            "/projects/__ID__/audit",
                            "#project-detail-tab-content",
                            147
                        ),
                    ]
                )
            },
            {
                "badge-large",
                StripModel(
                    "project-detail-tabstrip",
                    "Project Detail Tabs",
                    0,
                    [
                        Tab(
                            "tab-summary",
                            "Summary",
                            "/projects/__ID__/summary",
                            "#project-detail-tab-content"
                        ),
                        Tab(
                            "tab-inspections",
                            "Inspections",
                            "/projects/__ID__/inspections",
                            "#project-detail-tab-content",
                            9999
                        ),
                        Tab(
                            "tab-violations",
                            "Violations",
                            "/projects/__ID__/violations",
                            "#project-detail-tab-content"
                        ),
                        Tab(
                            "tab-audit",
                            "Audit",
                            "/projects/__ID__/audit",
                            "#project-detail-tab-content"
                        ),
                    ]
                )
            },
        };

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task TabStripVariantMatchesCanonical(string variant, object model)
    {
        await AssertSnapshot("tab_strip", "TabStrip", variant, model);
    }

    [Fact]
    public async Task TabStrip_ActiveIndexZero_FirstTabHasTabindex0()
    {
        var html = await RenderPartial(
            "TabStrip",
            StripModel("strip", "Tabs", 0, ProjectDetailTabs)
        );
        // active tab has tabindex 0; all others have -1
        var tabindexZeroCount = System.Text.RegularExpressions.Regex.Count(html, @"tabindex=""0""");
        var tabindexNegCount = System.Text.RegularExpressions.Regex.Count(html, @"tabindex=""-1""");
        tabindexZeroCount.Should().Be(1, "exactly one tab should have tabindex=0");
        tabindexNegCount.Should().Be(3, "three inactive tabs should have tabindex=-1");
    }

    [Fact]
    public async Task TabStrip_ActiveIndexTwo_ThirdTabAriaSelectedTrue()
    {
        var html = await RenderPartial(
            "TabStrip",
            StripModel("strip", "Tabs", 2, ProjectDetailTabs)
        );
        html.Should().Contain("id=\"tab-violations\"");
        // violations tab (index 2) should be selected
        var selectedCount = System.Text.RegularExpressions.Regex.Count(html, @"aria-selected=""true""");
        selectedCount.Should().Be(1, "exactly one tab should be aria-selected=true");
        var notSelectedCount = System.Text.RegularExpressions.Regex.Count(html, @"aria-selected=""false""");
        notSelectedCount.Should().Be(3, "three tabs should be aria-selected=false");
    }

    [Fact]
    public async Task TabStrip_ActiveIndexLastTab_LastTabActive()
    {
        var html = await RenderPartial(
            "TabStrip",
            StripModel("strip", "Tabs", 3, ProjectDetailTabs)
        );
        // the last tab (index 3 = tab-audit) should have tabindex=0 and aria-selected=true
        html.Should().Contain("id=\"tab-audit\"");
        var selectedCount = System.Text.RegularExpressions.Regex.Count(html, @"aria-selected=""true""");
        selectedCount.Should().Be(1);
    }

    [Fact]
    public async Task Badge_Count12_RendersWithText12AndAriaLabel()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "/a", "#panel", 12),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain("class=\"badge tab-strip__badge\"");
        html.Should().Contain(">12<");
        html.Should().Contain("aria-label=\"12 unread\"");
    }

    [Fact]
    public async Task Badge_Count0_RendersWithText0()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "/a", "#panel", 0),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain("class=\"badge tab-strip__badge\"");
        html.Should().Contain(">0<");
        html.Should().Contain("aria-label=\"0 unread\"");
    }

    [Fact]
    public async Task Badge_Null_NoBadgeElementRendered()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "/a", "#panel"),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().NotContain("tab-strip__badge");
    }

    [Fact]
    public async Task Badge_Count9999_RendersWithText9999NoTruncation()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "/a", "#panel", 9999),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain(">9999<");
        html.Should().NotContain("99+");
        html.Should().NotContain("max-width");
    }

    [Fact]
    public async Task Badge_NegativeCount_RendersVerbatimNotCoerced()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "/a", "#panel", -1),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain(">-1<");
    }

    [Fact]
    public async Task MissingAriaLabel_Throws()
    {
        var act = async () =>
            await RenderPartial("TabStrip", StripModel("strip", "", 0, ProjectDetailTabs));
        await act.Should().ThrowAsync<InvalidOperationException>();
    }

    [Fact]
    public async Task AllButtons_HaveTypeButton()
    {
        var html = await RenderPartial(
            "TabStrip",
            StripModel("strip", "Tabs", 0, ProjectDetailTabs)
        );
        var typeButtonCount = System.Text.RegularExpressions.Regex.Count(html, @"type=""button""");
        typeButtonCount.Should().Be(
            ProjectDetailTabs.Count,
            "every tab button must have type=button"
        );
    }

    [Fact]
    public async Task XSS_LabelIsEscaped()
    {
        var payload = "<script>alert(1)</script>";
        var tabs = new List<ExpandoObject> { Tab("tab-xss", payload, "/a", "#panel") };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>");
    }

    [Fact]
    public async Task XSS_AriaLabelIsEscaped()
    {
        var payload = "<script>alert(1)</script>";
        var html = await RenderPartial(
            "TabStrip",
            StripModel("strip", payload, 0, ProjectDetailTabs)
        );
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>");
    }

    [Fact]
    public async Task XSS_HxGetIsEscaped()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "javascript:alert(1)", "#panel"),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain("hx-get=\"javascript:alert(1)\"");
        html.Should().NotContain("hx-get=javascript:");
    }

    [Fact]
    public async Task XSS_HxTargetIsEscaped()
    {
        var tabs = new List<ExpandoObject>
        {
            Tab("tab-a", "Alpha", "/a", "#<script>"),
        };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().NotContain("<script>");
    }

    [Fact]
    public async Task EmptyLabel_DoesNotCrash()
    {
        var tabs = new List<ExpandoObject> { Tab("tab-a", "", "/a", "#panel") };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain("tab-strip__label");
    }

    [Fact]
    public async Task WhitespaceOnlyLabel_DoesNotCrash()
    {
        var tabs = new List<ExpandoObject> { Tab("tab-a", "   ", "/a", "#panel") };
        var html = await RenderPartial("TabStrip", StripModel("strip", "Tabs", 0, tabs));
        html.Should().Contain("tab-strip__label");
    }

    [Fact]
    public void TabStripTemplateDoesNotUseHtmlRaw()
    {
        File.ReadAllText(
                RepoPath(
                    "FieldMark",
                    "FieldMark.Web",
                    "Pages",
                    "Shared",
                    "Components",
                    "_TabStrip.cshtml"
                )
            )
            .Should()
            .NotContain("Html.Raw");
    }
}
