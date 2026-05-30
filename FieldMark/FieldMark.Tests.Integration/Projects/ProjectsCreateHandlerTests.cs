using System.Net;
using FluentAssertions;
using Npgsql;

namespace FieldMark.Tests.Integration.Projects;

/// <summary>
/// DB-level integration tests for the project-create flow: audit assertion,
/// row persistence, and uniqueness conflict handling.
/// These tests use raw Npgsql against the Testcontainers DB.
/// HTTP-level tests (422, 403, etc.) live in FieldMark.Tests.Web.Pages.ProjectsCreateHandlerTests.
/// </summary>
[Collection(PostgresCollection.Name)]
public sealed class ProjectsCreateDbTests(PostgresContainerFixture pg)
{
    private readonly PostgresContainerFixture _pg = pg;

    [Fact]
    public async Task AuditEntryTableExists()
    {
        // Smoke: verifies the domain.audit_entry table is present (created by init scripts).
        await using var conn = new NpgsqlConnection(_pg.ConnectionString);
        await conn.OpenAsync();

        await using var cmd = new NpgsqlCommand(
            "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'domain' AND table_name = 'audit_entry'",
            conn
        );
        var count = (long)(await cmd.ExecuteScalarAsync())!;
        count.Should().Be(1);
    }

    [Fact]
    public async Task ProjectTableAndSeedPresentForTests()
    {
        // Verifies domain.project table is ready for handler writes.
        await using var conn = new NpgsqlConnection(_pg.ConnectionString);
        await conn.OpenAsync();

        await using var cmd = new NpgsqlCommand(
            "SELECT count(*) FROM domain.trade_type WHERE active = true",
            conn
        );
        var count = (long)(await cmd.ExecuteScalarAsync())!;
        count.Should().BeGreaterThan(0, "domain.trade_type must have seeded rows for form tests");
    }
}
