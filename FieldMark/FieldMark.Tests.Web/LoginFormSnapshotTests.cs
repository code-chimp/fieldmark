using FieldMark.Tests.Web.Fixtures;
using FieldMark.Tests.Web.Helpers;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;

namespace FieldMark.Tests.Web;

[Collection(AuthTests.Name)]
public sealed class LoginFormSnapshotTests(PostgresFixture pg)
{
    private readonly PostgresFixture _pg = pg;

    [Fact]
    public async Task GetLogin_FormBlock_MatchesCanonicalExample()
    {
        var client = _pg.CreateFactory()
            .CreateClient(
                new WebApplicationFactoryClientOptions
                {
                    AllowAutoRedirect = false,
                    HandleCookies = true,
                }
            );

        var resp = await client.GetAsync("/login");
        resp.IsSuccessStatusCode.Should().BeTrue();
        var html = await resp.Content.ReadAsStringAsync();

        var actual = NormaliseHtml.ExtractLoginForm(html);
        actual.Should().NotBeEmpty(because: "GET /login must render the login form");

        // Load canonical reference from fieldmark_shared.
        // Path: repo-root/fieldmark_shared/components/login-form.example.html
        var repoRoot = FindRepoRoot();
        var canonicalPath = Path.Combine(
            repoRoot,
            "fieldmark_shared",
            "components",
            "login-form.example.html"
        );
        var canonical = NormaliseHtml.ExtractLoginForm(await File.ReadAllTextAsync(canonicalPath));

        actual
            .Should()
            .Be(
                canonical,
                because: "the rendered login form must be byte-identical to the canonical example "
                    + "(minus antiforgery token)"
            );
    }

    private static string FindRepoRoot()
    {
        var dir = new DirectoryInfo(AppContext.BaseDirectory);
        while (dir != null && !File.Exists(Path.Combine(dir.FullName, "Makefile")))
        {
            dir = dir.Parent;
        }
        return dir?.FullName
            ?? throw new InvalidOperationException(
                "Could not locate repo root (no Makefile found)."
            );
    }
}
