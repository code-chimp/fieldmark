using System.Net;
using FieldMark.Data.Context;
using FieldMark.Tests.Web.Fixtures;
using FluentAssertions;
using Microsoft.AspNetCore.Mvc.Testing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Tests.Web.Pages;

[Collection(AuthTests.Name)]
public sealed class AdminReferencePageTests(PostgresFixture pg)
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

    [Fact]
    public async Task AdminReference_Admin_RendersThreeSectionsAndSeededRows()
    {
        using var factory = _pg.CreateFactory();
        var client = factory.CreateClient(
            new WebApplicationFactoryClientOptions
            {
                AllowAutoRedirect = false,
                HandleCookies = true,
            }
        );
        await LoginAsync(client, "aisha");

        var resp = await client.GetAsync("/admin/reference");
        var html = await resp.Content.ReadAsStringAsync();

        resp.StatusCode.Should().Be(HttpStatusCode.OK);
        SectionHeadings(html)
            .Should()
            .Equal("Trade Types", "Violation Categories", "Compliance Rules");

        using var scope = factory.Services.CreateScope();
        var db = scope.ServiceProvider.GetRequiredService<FieldMarkDbContext>();
        var expectedRows =
            await db.TradeTypes.CountAsync()
            + await db.ViolationCategories.CountAsync()
            + await db.ComplianceRules.CountAsync();
        CountOccurrences(html, "<tbody>").Should().Be(3);
        CountOccurrences(html, "<tr>").Should().Be(expectedRows + 3);
        html.Should().Contain("ELEC");
        html.Should().Contain("OPEN_VIOLATION_GATE");
    }

    [Theory]
    [InlineData("marisol")]
    [InlineData("ravi")]
    [InlineData("pat")]
    [InlineData("kenji")]
    public async Task AdminReference_NonAdmin_Returns403WithoutReferenceState(string username)
    {
        var client = CreateClient();
        await LoginAsync(client, username);

        var resp = await client.GetAsync("/admin/reference");
        var body = await resp.Content.ReadAsStringAsync();

        resp.StatusCode.Should().Be(HttpStatusCode.Forbidden);
        body.Should().NotContain("ELEC");
        body.Should().NotContain("ELEC_NO_GFCI");
        body.Should().NotContain("OPEN_VIOLATION_GATE");
        body.Should().NotContain("rule_kind");
        body.Should().NotContain("parameters");
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

    private static IEnumerable<string> SectionHeadings(string html)
    {
        var index = 0;
        while (true)
        {
            var start = html.IndexOf("<h2>", index, StringComparison.Ordinal);
            if (start < 0)
            {
                yield break;
            }
            var end = html.IndexOf("</h2>", start, StringComparison.Ordinal);
            yield return html[(start + 4)..end];
            index = end + 5;
        }
    }

    private static int CountOccurrences(string haystack, string needle)
    {
        var count = 0;
        var index = 0;
        while ((index = haystack.IndexOf(needle, index, StringComparison.Ordinal)) >= 0)
        {
            count++;
            index += needle.Length;
        }
        return count;
    }
}
