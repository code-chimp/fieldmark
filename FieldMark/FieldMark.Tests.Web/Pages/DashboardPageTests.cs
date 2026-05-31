using System.Net;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;

namespace FieldMark.Tests.Web.Pages;

[Collection(AuthTests.Name)]
public sealed class DashboardPageTests(PostgresFixture pg)
{
    private readonly PostgresFixture _pg = pg;

    private HttpClient CreateClient(bool allowAutoRedirect = false) =>
        _pg.CreateFactory().CreateClient(new WebApplicationFactoryClientOptions
        {
            AllowAutoRedirect = allowAutoRedirect,
            HandleCookies = true,
        });

    private async Task<HttpClient> CreateAuthenticatedClientAsync(string username)
    {
        var client = CreateClient();
        var form = new FormUrlEncodedContent(new[]
        {
            new KeyValuePair<string, string>("username", username),
            new KeyValuePair<string, string>("password", "FieldMark!2026"),
        });
        await client.PostAsync("/login", form);
        return client;
    }

    [Fact]
    public async Task Dashboard_Unauthenticated_RedirectsToLogin()
    {
        var client = CreateClient();
        var resp = await client.GetAsync("/dashboard");
        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.PathAndQuery.Should().StartWith("/login");
    }

    [Fact]
    public async Task Dashboard_AuthorizedRole_Renders200()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/dashboard");
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
    }

    [Fact]
    public async Task Dashboard_NoRoleUser_Returns403()
    {
        var client = await CreateAuthenticatedClientAsync("testuser");
        var resp = await client.GetAsync("/dashboard");
        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
    }

    [Fact]
    public async Task Dashboard_RendersTileIdsAndResponsiveGridClasses()
    {
        var client = await CreateAuthenticatedClientAsync("marisol");
        var resp = await client.GetAsync("/dashboard");
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("id=\"compliance-tile-portfolio\"");
        html.Should().Contain("id=\"overdue-violations-tile\"");
        html.Should().Contain("id=\"active-projects-tile\"");
        html.Should().Contain("id=\"inspections-week-tile\"");
        html.Should().Contain("id=\"compliance-tile-portfolio\" role=\"status\"");
        html.Should().Contain("id=\"overdue-violations-tile\" role=\"status\"");
        html.Should().Contain("id=\"active-projects-tile\" role=\"status\"");
        html.Should().Contain("id=\"inspections-week-tile\" role=\"status\"");
        html.Should().Contain("grid-cols-1");
        html.Should().Contain("md:grid-cols-2");
        html.Should().Contain("xl:grid-cols-4");
    }
}
