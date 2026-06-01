using System.Net;
using System.Text.Json;
using FieldMark.Data.Reference;
using FieldMark.Domain.Entities.Reference;
using FieldMark.Data.Context;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.Extensions.DependencyInjection.Extensions;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Tests.Web.Pages;

[Collection(AuthTests.Name)]
public sealed class AdminReferenceCatalogPagesTests(PostgresFixture pg)
{
    private readonly PostgresFixture _pg = pg;

    private HttpClient CreateClient() =>
        _pg.CreateFactory()
            .CreateClient(
                new WebApplicationFactoryClientOptions
                {
                    AllowAutoRedirect = false,
                    HandleCookies = true,
                }
            );

    [Theory]
    [InlineData("/admin/reference/trade-types", "Trade Types", "Code", "Name", "Description", "Active", "ELEC")]
    [InlineData("/admin/reference/violation-categories", "Violation Categories", "Code", "Name", "Trade Type ID", "Default Severity", "Description", "Active", "ELEC_NO_GFCI")]
    [InlineData("/admin/reference/compliance-rules", "Compliance Rules", "Code", "Name", "Description", "Rule Kind", "Parameters", "Active", "OPEN_VIOLATION_GATE")]
    public async Task AdminReferenceCatalog_Admin_RendersExpectedPage(string path, params string[] expectedStrings)
    {
        var client = CreateClient();
        await LoginAsync(client, "aisha");

        var resp = await client.GetAsync(path);
        var html = await resp.Content.ReadAsStringAsync();

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        foreach (var expected in expectedStrings)
        {
            html.Should().Contain(expected);
        }

        html.Should().Contain("aria-label=\"Reference catalogs\"");
        html.Should().Contain("/admin/reference");
        if (path == "/admin/reference/trade-types")
        {
            html.Should().NotContain("/admin/reference/trade-types");
            html.Should().Contain("/admin/reference/violation-categories");
            html.Should().Contain("/admin/reference/compliance-rules");
        }
        else if (path == "/admin/reference/violation-categories")
        {
            html.Should().Contain("/admin/reference/trade-types");
            html.Should().NotContain("/admin/reference/violation-categories");
            html.Should().Contain("/admin/reference/compliance-rules");
        }
        else
        {
            html.Should().Contain("/admin/reference/trade-types");
            html.Should().Contain("/admin/reference/violation-categories");
            html.Should().NotContain("/admin/reference/compliance-rules");
        }
    }

    [Theory]
    [InlineData("marisol")]
    [InlineData("ravi")]
    [InlineData("pat")]
    [InlineData("kenji")]
    public async Task AdminReferenceCatalogs_NonAdmin_Returns403WithoutReferenceState(string username)
    {
        foreach (var path in new[]
                 {
                     "/admin/reference/trade-types",
                     "/admin/reference/violation-categories",
                     "/admin/reference/compliance-rules",
                 })
        {
            var client = CreateClient();
            await LoginAsync(client, username);

            var resp = await client.GetAsync(path);
            var body = await resp.Content.ReadAsStringAsync();

            resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
            body.Should().NotContain("ELEC");
            body.Should().NotContain("ELEC_NO_GFCI");
            body.Should().NotContain("OPEN_VIOLATION_GATE");
            body.Should().NotContain("rule_kind");
            body.Should().NotContain("parameters");
        }
    }

    [Theory]
    [InlineData("/admin/reference/trade-types", "No trade types defined.")]
    [InlineData("/admin/reference/violation-categories", "No violation categories defined.")]
    [InlineData("/admin/reference/compliance-rules", "No compliance rules defined.")]
    public async Task AdminReferenceCatalog_EmptyState_RendersExpectedRow(string path, string emptyMessage)
    {
        using var factory = _pg
            .CreateFactory()
            .WithWebHostBuilder(builder =>
                builder.ConfigureServices(services =>
                {
                    services.RemoveAll<IReferenceReader>();
                    services.AddScoped<IReferenceReader, EmptyReferenceReader>();
                })
            );

        var client = factory.CreateClient(
            new WebApplicationFactoryClientOptions
            {
                AllowAutoRedirect = false,
                HandleCookies = true,
            }
        );
        await LoginAsync(client, "aisha");

        var resp = await client.GetAsync(path);
        var html = await resp.Content.ReadAsStringAsync();

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        html.Should().Contain(emptyMessage);
    }

    private static async Task LoginAsync(HttpClient client, string username)
    {
        var form = new FormUrlEncodedContent(
            new[]
            {
                new KeyValuePair<string, string>("username", username),
                new KeyValuePair<string, string>("password", "FieldMark!2026"),
            }
        );
        var resp = await client.PostAsync("/login", form);
        resp.StatusCode.Should().Be(HttpStatusCode.Found);
    }

    private sealed class EmptyReferenceReader : IReferenceReader
    {
        public Task<IReadOnlyList<TradeType>> ListTradeTypesAsync(CancellationToken ct) =>
            Task.FromResult<IReadOnlyList<TradeType>>([]);

        public Task<IReadOnlyList<ViolationCategory>> ListViolationCategoriesAsync(CancellationToken ct) =>
            Task.FromResult<IReadOnlyList<ViolationCategory>>([]);

        public Task<IReadOnlyList<ComplianceRule>> ListComplianceRulesAsync(CancellationToken ct) =>
            Task.FromResult<IReadOnlyList<ComplianceRule>>([]);
    }
}
