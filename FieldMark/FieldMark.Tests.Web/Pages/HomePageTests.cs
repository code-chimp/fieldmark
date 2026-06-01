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
        _pg.CreateFactory().CreateClient(
            new WebApplicationFactoryClientOptions
            {
                AllowAutoRedirect = allowAutoRedirect,
                HandleCookies = true,
            }
        );

    private async Task<HttpClient> CreateAuthenticatedClientAsync(string username = "aisha")
    {
        var client = CreateClient(allowAutoRedirect: false);
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
        resp.Headers.Location!.ToString().Should().Contain("/login");
    }

    [Fact]
    public async Task Home_AuthenticatedAdmin_RedirectsToDashboard()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/");

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.ToString().Should().EndWith("/dashboard");
    }

    [Fact]
    public void IndexModel_NoRole_DefaultBadgeTokenIsUnknown()
    {
        var model = new FieldMark.Web.Pages.IndexModel();
        model.RoleBadgeToken.Should().Be("unknown");
    }
}
