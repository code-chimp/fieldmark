using System.Net;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Tests.Web;

[Collection(AuthTests.Name)]
public sealed class AuthFlowTests(PostgresFixture pg)
{
    private readonly PostgresFixture _pg = pg;

    private HttpClient CreateClient(bool allowAutoRedirect = false) =>
        _pg.CreateFactory().CreateClient(new WebApplicationFactoryClientOptions
        {
            AllowAutoRedirect = allowAutoRedirect,
            HandleCookies = true,
        });

    [Fact]
    public async Task Get_BusinessRoute_WhileUnauthenticated_Redirects302ToLogin()
    {
        var client = CreateClient();
        var resp = await client.GetAsync("/");

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.PathAndQuery.Should().StartWith("/login");
    }

    [Fact]
    public async Task Post_Login_WithValidCredentials_RedirectsToHome()
    {
        var client = CreateClient();
        var form = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("username", "marisol"),
            new KeyValuePair<string, string>("password", "FieldMark!2026"),
        });

        var resp = await client.PostAsync("/login", form);

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.ToString().Should().Be("/");
    }

    [Fact]
    public async Task Post_Login_WithValidCredentialsAndReturnUrl_RedirectsToDestination()
    {
        var client = CreateClient();
        var form = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("username", "marisol"),
            new KeyValuePair<string, string>("password", "FieldMark!2026"),
            new KeyValuePair<string, string>("return_url", "/compliance"),
        });

        var resp = await client.PostAsync("/login", form);

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.ToString().Should().Be("/compliance");
    }

    [Fact]
    public async Task Post_Login_WithInvalidPassword_Returns422AndDoesNotSetSessionCookie()
    {
        var client = CreateClient();
        var form = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("username", "marisol"),
            new KeyValuePair<string, string>("password", "wrongpassword"),
        });

        var resp = await client.PostAsync("/login", form);
        var body = await resp.Content.ReadAsStringAsync();

        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);

        // No Identity session cookie on failure.
        var cookies = resp.Headers.TryGetValues("Set-Cookie", out var vals) ? vals : [];
        cookies.Should().NotContain(c => c.Contains("Identity.Application"),
            because: "no session cookie must be set on failed login");

        body.Should().Contain("id=\"login-errors\"");
        body.Should().Contain("role=\"alert\"");
    }

    [Fact]
    public async Task Post_Logout_TerminatesSessionAndRedirectsToLogin()
    {
        var client = CreateClient();

        // Sign in.
        var loginForm = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("username", "marisol"),
            new KeyValuePair<string, string>("password", "FieldMark!2026"),
        });
        var loginResp = await client.PostAsync("/login", loginForm);
        loginResp.StatusCode.Should().Be(HttpStatusCode.Found);

        // Get a page to extract an antiforgery token for logout.
        var loginPageResp = await client.GetAsync("/login");
        var loginHtml = await loginPageResp.Content.ReadAsStringAsync();
        var token = ExtractAntiforgeryToken(loginHtml);

        // Logout.
        var logoutForm = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("__RequestVerificationToken", token),
        });
        var logoutResp = await client.PostAsync("/logout", logoutForm);

        logoutResp.StatusCode.Should().Be(HttpStatusCode.Found);
        logoutResp.Headers.Location!.ToString().Should().Be("/login");

        // Subsequent GET to a business route must redirect to /login.
        var afterLogout = await client.GetAsync("/");
        afterLogout.StatusCode.Should().Be(HttpStatusCode.Found);
        afterLogout.Headers.Location!.PathAndQuery.Should().StartWith("/login");
    }

    [Fact]
    public async Task Post_AuthzProbe_AsSiteSupervisor_Returns403WithoutLeakingState()
    {
        // Register a test-only probe route requiring ADMIN role via a startup filter.
        var factory = _pg.CreateFactory().WithWebHostBuilder(b =>
        {
            b.ConfigureServices(services =>
            {
                services.AddSingleton<Microsoft.AspNetCore.Hosting.IStartupFilter>(
                    new ProbeRouteStartupFilter());
            });
        });

        var client = factory.CreateClient(new WebApplicationFactoryClientOptions
        {
            AllowAutoRedirect = false,
            HandleCookies = true,
        });

        // Sign in as pat (SITE_SUPERVISOR — not ADMIN).
        var loginForm = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("username", "pat"),
            new KeyValuePair<string, string>("password", "FieldMark!2026"),
        });
        await client.PostAsync("/login", loginForm);

        var resp = await client.PostAsync("/__authz_probe", new StringContent(""));
        var body = await resp.Content.ReadAsStringAsync();

        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
        var stateLeakStrings = new[]
            { "Active", "OnHold", "Closed", "InProgress", "Open", "Resolved", "Voided" };
        foreach (var s in stateLeakStrings)
        {
            body.Should().NotContain(s, because: "403 must not leak entity state");
        }
    }

    private static string ExtractAntiforgeryToken(string html)
    {
        const string needle = "name=\"__RequestVerificationToken\" value=\"";
        var idx = html.IndexOf(needle, StringComparison.Ordinal);
        if (idx < 0) return "";
        var start = idx + needle.Length;
        var end = html.IndexOf('"', start);
        return end > start ? html[start..end] : "";
    }

    // Registered only in test code — not in production. Verified absent from make parity.
    private sealed class ProbeRouteStartupFilter : Microsoft.AspNetCore.Hosting.IStartupFilter
    {
        public Action<IApplicationBuilder> Configure(Action<IApplicationBuilder> next) =>
            app =>
            {
                // Branch before the main pipeline so MapRazorPages never sees this path.
                // UseAuthentication in the branch hydrates ctx.User from the session cookie.
                app.Map("/__authz_probe", probeApp =>
                {
                    probeApp.UseAuthentication();
                    probeApp.Run(async ctx =>
                    {
                        if (ctx.Request.Method != "POST")
                        {
                            ctx.Response.StatusCode = StatusCodes.Status405MethodNotAllowed;
                            return;
                        }
                        if (!ctx.User.Identity?.IsAuthenticated ?? true)
                        {
                            ctx.Response.StatusCode = StatusCodes.Status401Unauthorized;
                            return;
                        }
                        if (!ctx.User.IsInRole("ADMIN"))
                        {
                            ctx.Response.StatusCode = StatusCodes.Status403Forbidden;
                            await ctx.Response.WriteAsync("Forbidden.");
                            return;
                        }
                        ctx.Response.StatusCode = StatusCodes.Status200OK;
                        await ctx.Response.WriteAsync("OK.");
                    });
                });
                next(app);
            };
    }
}
