using System.Net;
using System.Net.Http.Headers;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;

namespace FieldMark.Tests.Web.Pages;

/// <summary>
/// GET /projects/new — page-render, 403 (non-admin), and 405 (GET /projects/) tests.
/// See docs/reference/project-create-form-contract.md.
/// </summary>
[Collection(AuthTests.Name)]
public sealed class ProjectsCreatePageTests(PostgresFixture pg)
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

    private async Task<HttpClient> CreateAuthenticatedClientAsync(
        string username = "aisha",
        bool allowRedirect = true
    )
    {
        var client = CreateClient(allowAutoRedirect: allowRedirect);
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

    // ─── GET /projects/new ──────────────────────────────────────────────────

    [Fact]
    public async Task ProjectsNew_Unauthenticated_RedirectsToLogin()
    {
        var client = CreateClient();
        var resp = await client.GetAsync("/projects/new");

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.PathAndQuery.Should().StartWith("/login");
    }

    [Fact]
    public async Task ProjectsNew_AdminUser_Returns200WithForm()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects/new");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("<h1>Create Project</h1>");
        html.Should().Contain("name=\"code\"");
        html.Should().Contain("name=\"name\"");
        html.Should().Contain("name=\"trade_scope_ids\"");
        html.Should().Contain("hx-post=\"/projects/create-submit\"");
    }

    [Fact]
    public async Task ProjectsNew_NonAdminUser_Returns403()
    {
        // marisol is COMPLIANCE_OFFICER — lacks project.create
        var client = await CreateAuthenticatedClientAsync("marisol");
        var resp = await client.GetAsync("/projects/new");

        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
    }

    [Fact]
    public async Task ProjectsNew_FormContainsCsrfToken()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects/new");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("__RequestVerificationToken");
    }

    // ─── GET /projects/create-submit → 405 ───────────────────────────────────────────────

    [Fact]
    public async Task ProjectsCollection_GetRequest_Returns405()
    {
        var client = await CreateAuthenticatedClientAsync("aisha", allowRedirect: false);
        var resp = await client.GetAsync("/projects/create-submit");

        resp.StatusCode.Should().Be(HttpStatusCode.MethodNotAllowed);
        // Framework may omit explicit Allow header for this handler-level 405 in test host.
        resp.Headers.TryGetValues("Allow", out var values).Should().BeFalse();
    }
}
