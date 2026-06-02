using System.Net;
using System.Text.RegularExpressions;
using System.Linq;
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

    private async Task<(string Status, AuditEntry? Audit)> ReadProjectAndLatestAuditAsync(Guid projectId)
    {
        using var factory = _pg.CreateFactory();
        using var scope = factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<FieldMarkDbContext>();
        var project = await db.Projects.AsNoTracking().FirstAsync(p => p.Id == projectId);
        var audit = await db.AuditEntries.AsNoTracking()
            .Where(a => a.ProjectId == projectId)
            .OrderByDescending(a => a.OccurredAt)
            .FirstOrDefaultAsync();
        return (project.Status.ToString(), audit);
    }

    private async Task<int> CountAuditEntriesAsync(Guid projectId)
    {
        using var factory = _pg.CreateFactory();
        using var scope = factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<FieldMarkDbContext>();
        return await db.AuditEntries.AsNoTracking().CountAsync(a => a.ProjectId == projectId);
    }

    private static int CountOobRegions(string html) =>
        Regex.Count(html, "hx-swap-oob=");

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
        html.Should().NotContain("id=\"project-detail\"");
        html.Should().Contain("id=\"project-header-strip\"");
        html.Should().Contain("id=\"project-detail-tabstrip\"");
        html.Should().Contain("id=\"project-detail-tab-content\"");
        html.Should().Contain("id=\"violation-detail\"");
        html.Should().NotContain("<html");
    }

    [Fact]
    public async Task ProjectsDetail_FullPage_PreservesStandaloneWrapper()
    {
        var id = await CreateProjectAsync();
        var client = await CreateAuthenticatedClientAsync();
        var resp = await client.GetAsync($"/projects/{id}");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("<div id=\"project-detail\">");
        html.Should().Contain("id=\"project-header-strip\"");
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
    public async Task ProjectPlaceOnHold_Get_ReturnsReasonForm()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Get, $"/projects/{id}/place-on-hold");
        req.Headers.Add("HX-Request", "true");
        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("role=\"form\"");
        html.Should().Contain($"hx-post=\"/projects/{id}/place-on-hold\"");
        html.Should().Contain("hx-target=\"#project-detail\"");
        html.Should().Contain("name=\"reason\"");
    }

    [Fact]
    public async Task ProjectResume_Get_ReturnsReasonForm()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Get, $"/projects/{id}/resume");
        req.Headers.Add("HX-Request", "true");
        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("role=\"form\"");
        html.Should().Contain($"hx-post=\"/projects/{id}/resume\"");
        html.Should().Contain("hx-target=\"#project-detail\"");
        html.Should().Contain("name=\"reason\"");
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_Success_RendersThreeRegionShape_AndPersistsAudit()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", "Weather delay")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().NotContain("id=\"project-detail\"");
        html.Should().Contain("id=\"compliance-tile\"");
        html.Should().Contain("hx-swap-oob=\"true\"");
        html.Should().Contain("hx-swap-oob=\"afterbegin:#audit-log\"");
        CountOobRegions(html).Should().Be(2);

        var (status, audit) = await ReadProjectAndLatestAuditAsync(id);
        status.Should().Be("OnHold");
        audit.Should().NotBeNull();
        audit!.Action.Should().Be("ProjectPlacedOnHold");
        audit.BeforeState!.RootElement.GetProperty("status").GetString().Should().Be("Active");
        audit.AfterState!.RootElement.GetProperty("status").GetString().Should().Be("OnHold");
        audit.Metadata!.RootElement.GetProperty("reason").GetString().Should().Be("Weather delay");
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_BlankReason_Returns422_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", "")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("Couldn&#x27;t submit transition");
        html.Should().Contain("id=\"reason-error\"");
        html.Should().Contain("Reason is required.");
        html.Should().Contain("aria-invalid=\"true\"");
        html.Should().Contain("aria-describedby=\"reason-error\"");
        CountOobRegions(html).Should().Be(0);
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_TooLongReason_Returns422_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", new string('x', 501))]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("Reason must be 500 characters or fewer.");
        html.Should().Contain("Couldn&#x27;t submit transition");
        CountOobRegions(html).Should().Be(0);
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_ControlCharReason_Returns422_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", "bad\u0001reason")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("Reason contains invalid control characters.");
        html.Should().Contain("Couldn&#x27;t submit transition");
        CountOobRegions(html).Should().Be(0);
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_ForbiddenForNonAdmin_Returns403_WithoutAudit()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("marisol");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", "nope")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
        var body = await resp.Content.ReadAsStringAsync();
        body.Should().Be("You do not have permission to access this page.");
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_UnknownId_Returns404()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{Guid.NewGuid()}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", "missing")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.NotFound);
    }

    [Fact]
    public async Task ProjectResume_Post_Success_RendersThreeRegionShape_AndPersistsAudit()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/resume")
        {
            Content = new FormUrlEncodedContent([new("reason", "crew available")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().NotContain("id=\"project-detail\"");
        html.Should().Contain("id=\"compliance-tile\"");
        html.Should().Contain("hx-swap-oob=\"true\"");
        html.Should().Contain("hx-swap-oob=\"afterbegin:#audit-log\"");
        CountOobRegions(html).Should().Be(2);

        var (status, audit) = await ReadProjectAndLatestAuditAsync(id);
        status.Should().Be("Active");
        audit.Should().NotBeNull();
        audit!.Action.Should().Be("ProjectResumed");
        audit.BeforeState!.RootElement.GetProperty("status").GetString().Should().Be("OnHold");
        audit.AfterState!.RootElement.GetProperty("status").GetString().Should().Be("Active");
        audit.Metadata!.RootElement.GetProperty("reason").GetString().Should().Be("crew available");
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_XssReason_IsEscapedOn422()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        const string payload = "<script>alert(1)</script>\u0001";
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", payload)]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("&lt;script&gt;alert(1)&lt;/script&gt;");
        html.Should().NotContain("<script>alert(1)</script>");
        CountOobRegions(html).Should().Be(0);
    }

    [Fact]
    public async Task ProjectResume_Post_FromActive_Returns409_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "Active");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/resume")
        {
            Content = new FormUrlEncodedContent([new("reason", "stale request")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.Conflict);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().NotContain("id=\"project-detail\"");
        html.Should().Contain("Couldn&#x27;t resume project");
        html.Should().Contain("Project is not on hold");
        CountOobRegions(html).Should().Be(0);
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectPlaceOnHold_Post_FromOnHold_Returns409_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/place-on-hold")
        {
            Content = new FormUrlEncodedContent([new("reason", "stale request")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.Conflict);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().NotContain("id=\"project-detail\"");
        html.Should().Contain("Couldn&#x27;t place project on hold");
        html.Should().Contain("Project is already on hold");
        CountOobRegions(html).Should().Be(0);
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectResume_Post_ForbiddenForNonAdmin_Returns403_WithoutAudit()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("marisol");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/resume")
        {
            Content = new FormUrlEncodedContent([new("reason", "nope")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
        var body = await resp.Content.ReadAsStringAsync();
        body.Should().Be("You do not have permission to access this page.");
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectResume_Post_BlankReason_IsAccepted()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/resume")
        {
            Content = new FormUrlEncodedContent([new("reason", "")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        CountOobRegions(html).Should().Be(2);

        var (status, audit) = await ReadProjectAndLatestAuditAsync(id);
        status.Should().Be("Active");
        audit.Should().NotBeNull();
        audit!.Metadata!.RootElement.GetProperty("reason").GetString().Should().Be("");
    }

    [Fact]
    public async Task ProjectResume_Post_TooLongReason_Returns422_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/resume")
        {
            Content = new FormUrlEncodedContent([new("reason", new string('x', 501))]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("Reason must be 500 characters or fewer.");
        html.Should().Contain("Couldn&#x27;t submit transition");
        CountOobRegions(html).Should().Be(0);
        (await CountAuditEntriesAsync(id)).Should().Be(0);
    }

    [Fact]
    public async Task ProjectResume_Post_ControlCharReason_Returns422_WithoutOob()
    {
        var id = await CreateProjectRowAsync(status: "OnHold");
        var client = await CreateAuthenticatedClientAsync("aisha");
        var req = new HttpRequestMessage(HttpMethod.Post, $"/projects/{id}/resume")
        {
            Content = new FormUrlEncodedContent([new("reason", "bad\u0001reason")]),
        };
        req.Headers.Add("HX-Request", "true");

        var resp = await client.SendAsync(req);
        resp.StatusCode.Should().Be(HttpStatusCode.UnprocessableEntity);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("Reason contains invalid control characters.");
        html.Should().Contain("Couldn&#x27;t submit transition");
        CountOobRegions(html).Should().Be(0);
        (await CountAuditEntriesAsync(id)).Should().Be(0);
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
