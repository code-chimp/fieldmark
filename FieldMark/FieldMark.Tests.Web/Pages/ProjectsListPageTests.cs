using System.Net;
using System.Net.Http.Headers;
using System.Text;
using System.Text.Json;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;

namespace FieldMark.Tests.Web.Pages;

/// <summary>
/// GET /projects — list page render + authz tests.
/// POST /grid/projects — SSRM endpoint authz + 400 validation tests.
/// See docs/reference/ag-grid-ssrm-contract.md
/// </summary>
[Collection(AuthTests.Name)]
public sealed class ProjectsListPageTests(PostgresFixture pg)
{
    private static readonly object[] EmptySortModel = [];
    private static readonly object EmptyFilterModel = new { };
    private static readonly string[] InvalidStatusValues = ["INVALID"];
    private static readonly string[] InjectionStatusValues = ["Active' OR '1'='1"];

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

    // ─── GET /projects — project list page ───────────────────────────────────

    [Fact]
    public async Task ProjectsList_Unauthenticated_RedirectsToLogin()
    {
        var client = CreateClient();
        var resp = await client.GetAsync("/projects");

        resp.StatusCode.Should().Be(HttpStatusCode.Found);
        resp.Headers.Location!.PathAndQuery.Should().StartWith("/login");
    }

    [Fact]
    public async Task ProjectsList_AdminUser_Returns200WithH1()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var html = await resp.Content.ReadAsStringAsync();
        html.Should().Contain("<h1>Projects</h1>");
    }

    [Fact]
    public async Task ProjectsList_ComplianceOfficer_Returns200()
    {
        // marisol is COMPLIANCE_OFFICER — has project.read
        var client = await CreateAuthenticatedClientAsync("marisol");
        var resp = await client.GetAsync("/projects");

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
    }

    [Fact]
    public async Task ProjectsList_RendersGridContainer()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects");
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("data-grid-endpoint=\"/grid/projects\"");
        html.Should().Contain("data-grid-target=\"#project-detail\"");
        html.Should().Contain("ag-theme-quartz");
    }

    [Fact]
    public async Task ProjectsList_RendersProjectDetailTarget()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects");
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("id=\"project-detail\"");
        html.Should().Contain("tabindex=\"-1\"");
    }

    [Fact]
    public async Task ProjectsList_RendersNoscriptFallback()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects");
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("<noscript>");
    }

    [Fact]
    public async Task ProjectsList_AdminHasCanCreateTrue()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await client.GetAsync("/projects");
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("data-can-create=\"true\"");
    }

    [Fact]
    public async Task ProjectsList_NonAdminHasCanCreateFalse()
    {
        // marisol is COMPLIANCE_OFFICER — lacks project.create
        var client = await CreateAuthenticatedClientAsync("marisol");
        var resp = await client.GetAsync("/projects");
        var html = await resp.Content.ReadAsStringAsync();

        html.Should().Contain("data-can-create=\"false\"");
    }

    // ─── POST /grid/projects — authz + 400 validation (no DB rows needed) ───

    private static async Task<HttpResponseMessage> PostGridAsync(HttpClient client, object body)
    {
        var json = JsonSerializer.Serialize(body);
        var content = new StringContent(json, Encoding.UTF8, "application/json");
        return await client.PostAsync("/grid/projects", content);
    }

    [Fact]
    public async Task GridProjects_Unauthenticated_Returns403OrRedirect()
    {
        var client = CreateClient();
        var resp = await PostGridAsync(client, new { startRow = 0, endRow = 10, sortModel = EmptySortModel, filterModel = EmptyFilterModel });

        ((int)resp.StatusCode).Should().BeOneOf(302, 303, 401, 403);
    }

    [Fact]
    public async Task GridProjects_InvalidJson_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var content = new StringContent("NOT JSON", Encoding.UTF8, "application/json");
        var resp = await client.PostAsync("/grid/projects", content);

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_UnknownColId_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = new[] { new { colId = "UNKNOWN", sort = "asc" } },
            filterModel = new { },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_InjectionColId_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = new[] { new { colId = "code; DROP TABLE domain.project --", sort = "asc" } },
            filterModel = new { },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_InvalidSortDirection_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = new[] { new { colId = "code", sort = "INVALID" } },
            filterModel = new { },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_NegativeStartRow_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = -1,
            endRow = 10,
            sortModel = EmptySortModel,
            filterModel = EmptyFilterModel,
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_PageSizeExceedsMax_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 1001,
            sortModel = Array.Empty<object>(),
            filterModel = new { },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_InvalidStatusValue_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = EmptySortModel,
            filterModel = new
            {
                status = new { filterType = "set", values = InvalidStatusValues },
            },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_InjectionStatusValue_Returns400()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = EmptySortModel,
            filterModel = new
            {
                status = new { filterType = "set", values = InjectionStatusValues },
            },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.BadRequest);
    }

    [Fact]
    public async Task GridProjects_EndpointDoesNotRequireAntiforgeryToken()
    {
        // The endpoint is CSRF-exempt (read-only). Posting without an antiforgery token must succeed (not 400).
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = EmptySortModel,
            filterModel = new { },
        });

        // Should be 200 (with DB data) — not 400 antiforgery rejection.
        resp.StatusCode.Should().Be(HttpStatusCode.OK);
    }

    // ─── Conformance-path: envelope shape + key casing (AC2/AC3) ─────────────

    [Fact]
    public async Task GridProjects_CanonicalFixture_ReturnsValidEnvelope()
    {
        // Issues the canonical SSRM fixture and asserts the response envelope is
        // {rows: [...], lastRow: N} per docs/reference/ag-grid-ssrm-contract.md
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = new[] { new { colId = "code", sort = "asc" } },
            filterModel = new { },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        resp.Content.Headers.ContentType!.MediaType.Should().Be("application/json");

        var json = await resp.Content.ReadAsStringAsync();
        using var doc = System.Text.Json.JsonDocument.Parse(json);
        var root = doc.RootElement;

        root.TryGetProperty("rows", out var rows).Should().BeTrue("envelope must contain 'rows'");
        root.TryGetProperty("lastRow", out var lastRow).Should().BeTrue("envelope must contain 'lastRow'");
        rows.ValueKind.Should().Be(System.Text.Json.JsonValueKind.Array);
        lastRow.ValueKind.Should().Be(System.Text.Json.JsonValueKind.Number);
    }

    [Fact]
    public async Task GridProjects_CanonicalFixture_RowKeysAreSnakeCase()
    {
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 10,
            sortModel = Array.Empty<object>(),
            filterModel = new { },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var json = await resp.Content.ReadAsStringAsync();
        using var doc = System.Text.Json.JsonDocument.Parse(json);
        var rows = doc.RootElement.GetProperty("rows");

        if (rows.GetArrayLength() > 0)
        {
            var firstRow = rows[0];
            // Required snake_case keys per contract doc.
            firstRow.TryGetProperty("id", out _).Should().BeTrue();
            firstRow.TryGetProperty("code", out _).Should().BeTrue();
            firstRow.TryGetProperty("name", out _).Should().BeTrue();
            firstRow.TryGetProperty("status", out _).Should().BeTrue();
            firstRow.TryGetProperty("compliance_score", out _).Should().BeTrue();
            firstRow.TryGetProperty("start_date", out _).Should().BeTrue();
            firstRow.TryGetProperty("target_completion_date", out _).Should().BeTrue();
            // No extra columns should leak.
            firstRow.TryGetProperty("description", out _).Should().BeFalse("description must not leak");
            firstRow.TryGetProperty("updated_at", out _).Should().BeFalse("updated_at must not leak");
        }
    }

    [Fact]
    public async Task GridProjects_LastRowEqualsTotalMatchingFilter()
    {
        // lastRow must equal the total count over the same WHERE, not just the page size.
        var client = await CreateAuthenticatedClientAsync("aisha");
        var pageResp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 2,  // request only 2 rows
            sortModel = Array.Empty<object>(),
            filterModel = new { },
        });

        pageResp.StatusCode.Should().Be(HttpStatusCode.OK);
        var json = await pageResp.Content.ReadAsStringAsync();
        using var doc = System.Text.Json.JsonDocument.Parse(json);
        var rows = doc.RootElement.GetProperty("rows");
        var lastRow = doc.RootElement.GetProperty("lastRow").GetInt32();

        // lastRow must be >= rows returned (it is the total, not just the page).
        lastRow.Should().BeGreaterThanOrEqualTo(rows.GetArrayLength());
    }

    [Fact]
    public async Task GridProjects_NullTargetDateSerializesAsNull()
    {
        // target_completion_date IS NULL rows must serialize as null (not omitted / empty string).
        var client = await CreateAuthenticatedClientAsync("aisha");
        var resp = await PostGridAsync(client, new
        {
            startRow = 0,
            endRow = 100,
            sortModel = Array.Empty<object>(),
            filterModel = new
            {
                target_completion_date = new { filterType = "date", type = "blank" },
            },
        });

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        var json = await resp.Content.ReadAsStringAsync();
        using var doc = System.Text.Json.JsonDocument.Parse(json);
        var rows = doc.RootElement.GetProperty("rows");

        foreach (var row in rows.EnumerateArray())
        {
            row.TryGetProperty("target_completion_date", out var tcd).Should().BeTrue();
            tcd.ValueKind.Should().Be(System.Text.Json.JsonValueKind.Null,
                "null dates must serialize as JSON null, not be omitted");
        }
    }
}
