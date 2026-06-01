using System.Net;
using FieldMark.Data.Context;
using FieldMark.Domain.Entities;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Tests.Web.Pages;

[Collection(AuthTests.Name)]
public sealed class ProjectsDetailPageTests(PostgresFixture pg)
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
        var client = CreateClient();
        var form = new FormUrlEncodedContent(
            [
                new("username", username),
                new("password", "FieldMark!2026"),
            ]
        );
        await client.PostAsync("/login", form);
        return client;
    }

    private async Task<Guid> CreateProjectAsync()
    {
        using var factory = _pg.CreateFactory();
        using var scope = factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<FieldMarkDbContext>();

        var tradeId = await db.TradeTypes.AsNoTracking().Select(t => t.Id).FirstAsync();
        var code = $"PD-{Guid.NewGuid().ToString("N")[..8].ToUpperInvariant()}";
        var created = Project.Create(code, "Project Detail Test", null, new DateOnly(2026, 6, 1), null, [tradeId], []);
        db.Projects.Add(created.Project);
        db.ProjectTradeScopes.AddRange(created.Scopes);
        await db.SaveChangesAsync();
        return created.Project.Id;
    }

    private async Task<Guid> CreateProjectRowAsync(
        string status = "Active",
        string? name = "Project Detail Test",
        string? description = null,
        DateOnly? targetCompletionDate = null
    )
    {
        using var factory = _pg.CreateFactory();
        using var scope = factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<FieldMarkDbContext>();
        var id = Guid.NewGuid();
        var code = $"PD-{Guid.NewGuid().ToString("N")[..8].ToUpperInvariant()}";
        var hasDescription = description is not null;
        var hasTargetCompletionDate = targetCompletionDate is not null;
        if (!hasDescription && !hasTargetCompletionDate)
        {
            await db.Database.ExecuteSqlRawAsync(
                """
                INSERT INTO domain.project
                  (id, code, name, description, status, start_date, target_completion_date, compliance_score, created_at, updated_at)
                VALUES
                  ({0}, {1}, {2}, NULL, {3}, {4}, NULL, 100, now(), now())
                """,
                id,
                code,
                name ?? "Project Detail Test",
                status,
                new DateOnly(2026, 6, 1)
            );
        }
        else if (hasDescription && hasTargetCompletionDate)
        {
            await db.Database.ExecuteSqlRawAsync(
                """
                INSERT INTO domain.project
                  (id, code, name, description, status, start_date, target_completion_date, compliance_score, created_at, updated_at)
                VALUES
                  ({0}, {1}, {2}, {3}, {4}, {5}, {6}, 100, now(), now())
                """,
                id,
                code,
                name ?? "Project Detail Test",
                (object)description!,
                status,
                new DateOnly(2026, 6, 1),
                (object)targetCompletionDate!.Value
            );
        }
        else if (hasDescription)
        {
            await db.Database.ExecuteSqlRawAsync(
                """
                INSERT INTO domain.project
                  (id, code, name, description, status, start_date, target_completion_date, compliance_score, created_at, updated_at)
                VALUES
                  ({0}, {1}, {2}, {3}, {4}, {5}, NULL, 100, now(), now())
                """,
                id,
                code,
                name ?? "Project Detail Test",
                (object)description!,
                status,
                new DateOnly(2026, 6, 1)
            );
        }
        else
        {
            await db.Database.ExecuteSqlRawAsync(
                """
                INSERT INTO domain.project
                  (id, code, name, description, status, start_date, target_completion_date, compliance_score, created_at, updated_at)
                VALUES
                  ({0}, {1}, {2}, NULL, {3}, {4}, {5}, 100, now(), now())
                """,
                id,
                code,
                name ?? "Project Detail Test",
                status,
                new DateOnly(2026, 6, 1),
                (object)targetCompletionDate!.Value
            );
        }
        return id;
    }

    [Fact]
    public async Task ProjectsDetail_Unauthenticated_RedirectsToLogin()
    {
        var id = await CreateProjectAsync();
        var client = CreateClient();
        var resp = await client.GetAsync($"/projects/{id}");
        resp.StatusCode.Should().Be(HttpStatusCode.Found);
    }

    [Fact]
    public async Task ProjectsDetail_HtmxMode_ReturnsBodyFragmentOnly()
    {
        var id = await CreateProjectAsync();
        var client = await CreateAuthenticatedClientAsync();
        var req = new HttpRequestMessage(HttpMethod.Get, $"/projects/{id}");
        req.Headers.Add("HX-Request", "true");
        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("id=\"project-header-strip\"");
        html.Should().Contain("id=\"project-detail-tabstrip\"");
        html.Should().Contain("id=\"project-detail-tab-content\"");
        html.Should().Contain("id=\"violation-detail\"");
        html.Should().NotContain("<html");
    }

    [Fact]
    public async Task ProjectsDetailTab_Violations_ReturnsPanelAndOobTabstrip()
    {
        var id = await CreateProjectAsync();
        var client = await CreateAuthenticatedClientAsync();
        var req = new HttpRequestMessage(HttpMethod.Get, $"/projects/{id}/tabs/violations");
        req.Headers.Add("HX-Request", "true");
        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("aria-labelledby=\"tab-violations\"");
        html.Should().Contain("hx-swap-oob=\"outerHTML\"");
        html.Should().Contain("id=\"project-detail-tabstrip\"");
        html.Should().Contain("id=\"tab-violations\"");
        html.Should().Contain("aria-selected=\"true\"");
    }

    [Fact]
    public async Task ProjectsDetail_AdminActive_ShowsHoldAndClose_DisablesResume()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("id=\"place-on-hold-btn\"");
        html.Should().Contain("id=\"close-btn\"");
        html.Should().Contain("id=\"resume-btn\"");
        html.Should().Contain("aria-describedby=\"resume-btn-reason\"");
    }

    [Fact]
    public async Task ProjectsDetail_AdminOnHold_ShowsResume_DisablesHoldAndClose()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("id=\"resume-btn\"");
        html.Should().Contain("id=\"place-on-hold-btn\"");
        html.Should().Contain("id=\"close-btn\"");
        html.Should().Contain("aria-describedby=\"place-on-hold-btn-reason\"");
        html.Should().Contain("aria-describedby=\"close-btn-reason\"");
    }

    [Fact]
    public async Task ProjectsDetail_NonAdmin_HidesAllActionButtons()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("marisol");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().NotContain("id=\"place-on-hold-btn\"");
        html.Should().NotContain("id=\"resume-btn\"");
        html.Should().NotContain("id=\"close-btn\"");
    }

    [Fact]
    public async Task ProjectsDetail_AdminClosed_ShowsAllButtonsDisabled()
    {
        var id = await CreateProjectRowAsync(status: "Closed");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("id=\"place-on-hold-btn\"");
        html.Should().Contain("id=\"resume-btn\"");
        html.Should().Contain("id=\"close-btn\"");
        html.Should().Contain("aria-describedby=\"place-on-hold-btn-reason\"");
        html.Should().Contain("aria-describedby=\"resume-btn-reason\"");
        html.Should().Contain("aria-describedby=\"close-btn-reason\"");
    }

    [Fact]
    public async Task ProjectsDetail_NoRoleUser_ReturnsForbidden()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("testuser");
        var resp = await client.GetAsync($"/projects/{id}");
        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
    }

    [Fact]
    public async Task ProjectsDetail_EmptyFields_ShowFallbacks()
    {
        var id = await CreateProjectRowAsync(status: "Active", description: null, targetCompletionDate: null);
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("No inspectors assigned");
        html.Should().Contain("—");
    }

    [Fact]
    public async Task ProjectsDetail_WhitespaceDescription_ShowsFallback()
    {
        var id = await CreateProjectRowAsync(status: "Active", description: "   ", targetCompletionDate: null);
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("<dt>Description</dt><dd>&#x2014;</dd>");
    }

    [Fact]
    public async Task ProjectsDetail_XssPayloads_AreEscaped()
    {
        const string payload = "<script>alert(1)</script>";
        var id = await CreateProjectRowAsync(status: "Active", name: payload, description: payload);
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync($"/projects/{id}");
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>alert(1)</script>");
    }
}
