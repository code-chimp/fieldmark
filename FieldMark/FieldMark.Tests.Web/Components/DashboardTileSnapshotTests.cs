using System.Dynamic;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;

namespace FieldMark.Tests.Web.Components;

[Collection(AuthTests.Name)]
public sealed class DashboardTileSnapshotTests(PostgresFixture pg) : ComponentRenderFixture(pg)
{
    private static ExpandoObject TileModel(
        string value = "12",
        string secondary = "",
        string valueColor = "",
        bool roleStatus = false
    ) =>
        Model(
            ("TileId", "open-violations-tile"),
            ("Label", "Open Violations"),
            ("Value", value),
            ("Secondary", secondary),
            ("ValueColor", valueColor),
            ("RoleStatus", roleStatus)
        );

    public static TheoryData<string, object> Variants =>
        new()
        {
            { "populated", TileModel() },
            { "zero-value", TileModel(value: "0") },
            { "populated-with-secondary", TileModel(secondary: "3 critical") },
            { "populated-with-color", TileModel(valueColor: "danger") },
            { "empty", TileModel(value: "") },
            { "status-region", TileModel(roleStatus: true) },
        };

    [Theory]
    [MemberData(nameof(Variants))]
    public async Task DashboardTileVariantMatchesCanonical(string variant, object model)
    {
        await AssertSnapshot("dashboard_tile", "DashboardTile", variant, model);
    }

    [Fact]
    public async Task DashboardTileWhitespaceOnlyValueMatchesEmptyCanonical()
    {
        await AssertSnapshot("dashboard_tile", "DashboardTile", "empty", TileModel(value: "   "));
    }

    [Fact]
    public void DashboardTileTemplateDoesNotUseHtmlRaw()
    {
        File.ReadAllText(
                RepoPath(
                    "FieldMark",
                    "FieldMark.Web",
                    "Pages",
                    "Shared",
                    "Components",
                    "_DashboardTile.cshtml"
                )
            )
            .Should()
            .NotContain("Html.Raw");
    }
}
