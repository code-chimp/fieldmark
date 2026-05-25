using System.Diagnostics;
using System.Net;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;

namespace FieldMark.Tests.Web.Pages;

[Collection(AuthTests.Name)]
public sealed class HomePageTests(PostgresFixture pg)
{
    private readonly PostgresFixture _pg = pg;

    private HttpClient CreateClient(bool allowAutoRedirect = false) =>
        _pg.CreateFactory()
            .CreateClient(
                new WebApplicationFactoryClientOptions
                {
                    AllowAutoRedirect = allowAutoRedirect,
                    HandleCookies = true,
                }
            );

    private async Task<HttpClient> CreateAuthenticatedClientAsync(string username = "aisha")
    {
        var client = CreateClient(allowAutoRedirect: true);
        var form = new FormUrlEncodedContent(
            new[]
            {
                new KeyValuePair<string, string>("username", username),
                new KeyValuePair<string, string>("password", "FieldMark!2026"),
            }
        );
        await client.PostAsync("/login", form);
        return client;
    }

    [Fact]
    public async Task Home_Unauthenticated_RedirectsToLogin()
    {
        var client = CreateClient();
        var resp = await client.GetAsync("/");

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.PathAndQuery.Should().StartWith("/login");
    }

    [Fact]
    public async Task Home_AuthenticatedAdmin_RendersCanonicalMarkup()
    {
        // aisha is seeded as ADMIN
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("<h1>FieldMark</h1>");
        html.Should().Contain("badge-danger");
        html.Should().Contain(">Admin<");
        html.Should().Contain("Your projects will appear here.");
    }

    [Fact]
    public async Task Home_AuthenticatedAdmin_RendersAvatarMenuWithInitials()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("avatar-menu-wrapper");
        html.Should().Contain("avatar-menu-dropdown");
        html.Should().Contain("href=\"/logout\"");
    }

    /// <summary>
    /// AC #6: zero WCAG 2.1 AA violations under axe-core.
    /// Renders the Home page via WebApplicationFactory, writes the HTML to a temp file,
    /// and runs @axe-core/cli against it (file:// path — no live HTTP server required).
    /// Skips gracefully when npx is not on PATH; surface skips in CI.
    /// Manual recipe: npx @axe-core/cli http://localhost:5xxx/ (authenticated session).
    /// </summary>
    [Fact]
    public async Task Home_AuthenticatedAdmin_PassesAxeCore()
    {
        var npx = FindOnPath("npx");
        if (npx is null)
            throw new Xunit.Sdk.XunitException(
                "AC #6 gate cannot run because `npx` is not on PATH. Install Node.js so @axe-core/cli can execute."
            );

        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/");
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();

        var tmp = Path.GetTempFileName() + ".html";
        await File.WriteAllTextAsync(tmp, html);
        try
        {
            using var proc = Process.Start(
                new ProcessStartInfo
                {
                    FileName = npx,
                    Arguments = $"@axe-core/cli file://{tmp}",
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                }
            )!;
            await proc.WaitForExitAsync();
            var output =
                await proc.StandardOutput.ReadToEndAsync()
                + await proc.StandardError.ReadToEndAsync();
            proc.ExitCode.Should().Be(0, $"axe-core found WCAG 2.1 AA violations:\n{output}");
        }
        finally
        {
            File.Delete(tmp);
        }
    }

    /// <summary>
    /// AC #7: DOM-order check for the required focus sequence:
    /// skip-link → brand lockup → theme-toggle pill → avatar button → sign-out link.
    /// DOM order is the primary determinant of tab order when no tabindex attributes are present.
    /// Full runtime focus-order verification (CSS, tabindex overrides) still requires Playwright;
    /// wire it in Epic 7 (Story 7.1). Manual recipe: open app, Tab 5 times, verify sequence.
    /// </summary>
    [Fact]
    public async Task Home_TabOrderMatchesContract()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/");
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();

        var idxSkipLink = html.IndexOf("class=\"skip-link\"", StringComparison.Ordinal);
        var idxWordmark = html.IndexOf("class=\"fm-brand-lockup\"", StringComparison.Ordinal);
        var idxThemeToggle = html.IndexOf("class=\"theme-toggle-pill\"", StringComparison.Ordinal);
        var idxAvatarBtn = html.IndexOf("class=\"avatar-menu\"", StringComparison.Ordinal);
        var idxSignOut = html.IndexOf("href=\"/logout\"", StringComparison.Ordinal);

        idxSkipLink.Should().BeGreaterThan(-1, "skip-link must be present");
        idxWordmark.Should().BeGreaterThan(-1, "fm-brand-lockup must be present");
        idxThemeToggle.Should().BeGreaterThan(-1, "theme-toggle-pill must be present");
        idxAvatarBtn.Should().BeGreaterThan(-1, "avatar-menu button must be present");
        idxSignOut.Should().BeGreaterThan(-1, "sign-out anchor must be present");

        idxSkipLink.Should().BeLessThan(idxWordmark, "skip-link must precede brand lockup in DOM");
        idxWordmark
            .Should()
            .BeLessThan(idxThemeToggle, "brand lockup must precede theme-toggle in DOM");
        idxThemeToggle
            .Should()
            .BeLessThan(idxAvatarBtn, "theme-toggle must precede avatar button in DOM");
        idxAvatarBtn
            .Should()
            .BeLessThan(idxSignOut, "avatar button must precede sign-out link in DOM");
    }

    /// <summary>
    /// AC2.4: IndexModel defaults RoleBadgeToken to "unknown" when the user has no canonical role.
    /// Uses Microsoft.Extensions.Logging.Abstractions so no real DI infrastructure is needed.
    /// </summary>
    [Fact]
    public void IndexModel_NoRole_DefaultBadgeTokenIsUnknown()
    {
        var logger = Microsoft
            .Extensions
            .Logging
            .Abstractions
            .NullLogger<FieldMark.Web.Pages.IndexModel>
            .Instance;
        var model = new FieldMark.Web.Pages.IndexModel(logger);

        model
            .RoleBadgeToken.Should()
            .Be(
                "unknown",
                "IndexModel must default to 'unknown' so the CSS fallback renders for users with no role"
            );
    }

    /// <summary>
    /// AC2.4 / Story 1.14 regression guard: when a ClaimsPrincipal carries both a
    /// canonical role (COMPLIANCE_OFFICER) and an unknown role (ANALYST), the
    /// canonical badge token must be selected even though "ANALYST" sorts before
    /// "COMPLIANCE_OFFICER" in ordinal order. A pure lexical sort would pick ANALYST
    /// and render badge-unknown incorrectly.
    /// </summary>
    [Fact]
    public void IndexModel_MixedCanonicalAndUnknownRole_PrefersCanonicalBadgeToken()
    {
        var logger = Microsoft
            .Extensions
            .Logging
            .Abstractions
            .NullLogger<FieldMark.Web.Pages.IndexModel>
            .Instance;
        var model = new FieldMark.Web.Pages.IndexModel(logger);

        // "ANALYST" (unknown) sorts before "COMPLIANCE_OFFICER" (canonical) in ordinal
        // order — a pure lexical sort would pick ANALYST → badge-unknown (the regression).
        var claims = new[]
        {
            new System.Security.Claims.Claim(System.Security.Claims.ClaimTypes.Role, "ANALYST"),
            new System.Security.Claims.Claim(
                System.Security.Claims.ClaimTypes.Role,
                "COMPLIANCE_OFFICER"
            ),
        };
        var identity = new System.Security.Claims.ClaimsIdentity(claims, "Test");
        var principal = new System.Security.Claims.ClaimsPrincipal(identity);

        var httpContext = new Microsoft.AspNetCore.Http.DefaultHttpContext { User = principal };
        model.PageContext = new Microsoft.AspNetCore.Mvc.RazorPages.PageContext
        {
            HttpContext = httpContext,
        };

        model.OnGet();

        model
            .RoleBadgeToken.Should()
            .Be(
                "info",
                "COMPLIANCE_OFFICER is canonical and must be preferred over the lexically-earlier unknown role ANALYST"
            );
        model.RoleLabel.Should().Be("Compliance Officer");
    }

    private static string? FindOnPath(string executable)
    {
        var paths = Environment.GetEnvironmentVariable("PATH")?.Split(Path.PathSeparator) ?? [];
        foreach (var dir in paths)
        {
            var full = Path.Combine(dir, executable);
            if (File.Exists(full))
                return full;
            // Windows
            var fullExe = full + ".cmd";
            if (File.Exists(fullExe))
                return fullExe;
        }
        return null;
    }

    [Fact]
    public async Task Home_AuthenticatedAdmin_RendersBrandLockupInNav()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("class=\"fm-brand-lockup\"");
        html.Should().Contain("aria-label=\"FieldMark home\"");
    }
}
