using System;
using System.Threading.Tasks;
using FluentAssertions;
using Npgsql;

namespace FieldMark.Tests.Integration;

// Smoke test for action item A3 from the Epic 1 retro: prove the harness can
// open a transaction, write to a domain.* table, roll back, and observe that
// nothing was persisted. Story 2.2 (audit_entry helper) will lean on this
// pattern to verify append_audit_entry's transactional semantics.
[Collection(PostgresCollection.Name)]
public sealed class DomainRollbackSmokeTests
{
    private readonly PostgresContainerFixture _pg;

    public DomainRollbackSmokeTests(PostgresContainerFixture pg) => _pg = pg;

    [Fact]
    public async Task Insert_then_rollback_does_not_persist()
    {
        var id = Guid.NewGuid();
        var code = $"TEST_{Guid.NewGuid():N}".Substring(0, 16);

        await using var conn = new NpgsqlConnection(_pg.ConnectionString);
        await conn.OpenAsync();

        await using (var tx = await conn.BeginTransactionAsync())
        {
            await using (var insert = new NpgsqlCommand(
                "INSERT INTO domain.trade_type (id, code, name) VALUES (@id, @code, @name)",
                conn, tx))
            {
                insert.Parameters.AddWithValue("id", id);
                insert.Parameters.AddWithValue("code", code);
                insert.Parameters.AddWithValue("name", "Rollback smoke");
                await insert.ExecuteNonQueryAsync();
            }

            // Visible inside the open transaction.
            await using var checkInside = new NpgsqlCommand(
                "SELECT count(*) FROM domain.trade_type WHERE code = @code", conn, tx);
            checkInside.Parameters.AddWithValue("code", code);
            var insideCount = (long)(await checkInside.ExecuteScalarAsync() ?? 0L);
            insideCount.Should().Be(1);

            await tx.RollbackAsync();
        }

        await using var checkAfter = new NpgsqlCommand(
            "SELECT count(*) FROM domain.trade_type WHERE code = @code", conn);
        checkAfter.Parameters.AddWithValue("code", code);
        var afterCount = (long)(await checkAfter.ExecuteScalarAsync() ?? 0L);
        afterCount.Should().Be(0, "rollback must not persist the row");
    }

    [Fact]
    public async Task Seed_reference_data_is_present()
    {
        await using var conn = new NpgsqlConnection(_pg.ConnectionString);
        await conn.OpenAsync();

        await using var cmd = new NpgsqlCommand(
            "SELECT count(*) FROM domain.trade_type WHERE active", conn);
        var count = (long)(await cmd.ExecuteScalarAsync() ?? 0L);
        count.Should().BeGreaterThan(0, "init scripts should have populated reference data");
    }
}
