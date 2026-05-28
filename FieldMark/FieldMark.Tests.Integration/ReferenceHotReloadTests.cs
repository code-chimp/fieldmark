using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using FluentAssertions;
using Microsoft.EntityFrameworkCore;
using Npgsql;

namespace FieldMark.Tests.Integration;

[Collection(PostgresCollection.Name)]
public sealed class ReferenceHotReloadTests(PostgresContainerFixture pg)
{
    [Fact]
    public async Task Reader_sees_reference_updates_without_recreation()
    {
        const string code = "OPEN_VIOLATION_GATE";
        const string updatedName = "Open Violation Closure Gate (UPDATED)";

        var options = new DbContextOptionsBuilder<FieldMarkDbContext>()
            .UseNpgsql(pg.ConnectionString)
            .UseSnakeCaseNamingConvention()
            .Options;

        await using var ctx = new FieldMarkDbContext(options);
        var reader = new ReferenceReader(ctx);

        var before = (await reader.ListComplianceRulesAsync()).Single(r => r.Code == code);
        var originalName = before.Name;
        before.Name.Should().NotBe(updatedName);

        await using var conn = new NpgsqlConnection(pg.ConnectionString);
        await conn.OpenAsync();

        try
        {
            await using (
                var update = new NpgsqlCommand(
                    "UPDATE domain.compliance_rule SET name = @name WHERE code = @code",
                    conn
                )
            )
            {
                update.Parameters.AddWithValue("name", updatedName);
                update.Parameters.AddWithValue("code", code);
                await update.ExecuteNonQueryAsync();
            }

            var after = (await reader.ListComplianceRulesAsync()).Single(r => r.Code == code);
            after.Name.Should().Be(updatedName);
        }
        finally
        {
            await using var restore = new NpgsqlCommand(
                "UPDATE domain.compliance_rule SET name = @name WHERE code = @code",
                conn
            );
            restore.Parameters.AddWithValue("name", originalName);
            restore.Parameters.AddWithValue("code", code);
            await restore.ExecuteNonQueryAsync();
        }
    }
}
