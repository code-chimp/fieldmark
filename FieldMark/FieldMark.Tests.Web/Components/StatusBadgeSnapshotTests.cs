using FieldMark.Tests.Web.Fixtures;
using FieldMark.Tests.Web.Helpers;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class StatusBadgeSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    public static TheoryData<string, object> Variants =>
        new()
        {
            {
                "project-active",
                Model(("Entity", "Project"), ("Value", "Active"), ("Severity", ""))
            },
            {
                "project-on-hold",
                Model(("Entity", "Project"), ("Value", "OnHold"), ("Severity", ""))
            },
            {
                "project-closed",
                Model(("Entity", "Project"), ("Value", "Closed"), ("Severity", ""))
            },
            {
                "inspection-scheduled",
                Model(("Entity", "Inspection"), ("Value", "Scheduled"), ("Severity", ""))
            },
            {
                "inspection-in-progress",
                Model(("Entity", "Inspection"), ("Value", "InProgress"), ("Severity", ""))
            },
            {
                "inspection-completed-pass",
                Model(("Entity", "Inspection"), ("Value", "CompletedPass"), ("Severity", ""))
            },
            {
                "inspection-completed-conditional",
                Model(("Entity", "Inspection"), ("Value", "CompletedConditional"), ("Severity", ""))
            },
            {
                "inspection-completed-fail",
                Model(("Entity", "Inspection"), ("Value", "CompletedFail"), ("Severity", ""))
            },
            {
                "inspection-cancelled",
                Model(("Entity", "Inspection"), ("Value", "Cancelled"), ("Severity", ""))
            },
            {
                "violation-open-critical-high",
                Model(("Entity", "Violation"), ("Value", "Open"), ("Severity", "Critical"))
            },
            {
                "violation-open-medium-low",
                Model(("Entity", "Violation"), ("Value", "Open"), ("Severity", "Medium"))
            },
            {
                "violation-in-progress",
                Model(("Entity", "Violation"), ("Value", "InProgress"), ("Severity", ""))
            },
            {
                "violation-resolved",
                Model(("Entity", "Violation"), ("Value", "Resolved"), ("Severity", ""))
            },
            {
                "violation-voided",
                Model(("Entity", "Violation"), ("Value", "Voided"), ("Severity", ""))
            },
            {
                "corrective-action-submitted",
                Model(("Entity", "CorrectiveAction"), ("Value", "Submitted"), ("Severity", ""))
            },
            {
                "corrective-action-under-review",
                Model(("Entity", "CorrectiveAction"), ("Value", "UnderReview"), ("Severity", ""))
            },
            {
                "corrective-action-approved",
                Model(("Entity", "CorrectiveAction"), ("Value", "Approved"), ("Severity", ""))
            },
            {
                "corrective-action-rejected",
                Model(("Entity", "CorrectiveAction"), ("Value", "Rejected"), ("Severity", ""))
            },
            {
                "severity-critical",
                Model(("Entity", "Severity"), ("Value", "Critical"), ("Severity", ""))
            },
            { "severity-high", Model(("Entity", "Severity"), ("Value", "High"), ("Severity", "")) },
            {
                "severity-medium",
                Model(("Entity", "Severity"), ("Value", "Medium"), ("Severity", ""))
            },
            { "severity-low", Model(("Entity", "Severity"), ("Value", "Low"), ("Severity", "")) },
            { "unknown", Model(("Entity", "Violation"), ("Value", "Foobar"), ("Severity", "")) },
        };

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task StatusBadgeVariantMatchesCanonical(string variant, object model)
    {
        await AssertSnapshot("status_badge", "StatusBadge", variant, model);
    }

    [Fact]
    public async Task StatusBadgeUnknownFallbackClass()
    {
        var html = NormaliseHtml.NormaliseComponent(
            await RenderPartial(
                "StatusBadge",
                Model(("Entity", "Violation"), ("Value", "Foobar"), ("Severity", ""))
            )
        );
        html.Should().Contain("badge-unknown");
    }

    [Fact]
    public void StatusBadgeTemplateDoesNotUseHtmlRaw()
    {
        File.ReadAllText(
                RepoPath(
                    "FieldMark",
                    "FieldMark.Web",
                    "Pages",
                    "Shared",
                    "Components",
                    "_StatusBadge.cshtml"
                )
            )
            .Should()
            .NotContain("Html.Raw");
    }
}
