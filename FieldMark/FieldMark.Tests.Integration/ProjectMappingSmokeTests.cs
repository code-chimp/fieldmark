using System;
using System.Threading.Tasks;
using FieldMark.Data.Context;
using FieldMark.Domain.ValueObjects;
using FluentAssertions;
using Microsoft.EntityFrameworkCore;
using Npgsql;

namespace FieldMark.Tests.Integration;

// Story 2.1 AC5 — round-trip smoke for the domain.project mapping plus a
// count probe per relation table to prove the JobSite / ProjectTradeScope /
// ProjectInspector configurations compile and read against the live DDL.
//
// Pattern follows DomainRollbackSmokeTests: a single NpgsqlConnection
// holds an open transaction across the raw INSERT and the EF Core read;
// rolling that transaction back means nothing reaches disk — no
// commit-plus-delete cleanup window where a crashed test could leak.
[Collection(PostgresCollection.Name)]
public sealed class ProjectMappingSmokeTests
{
    private readonly PostgresContainerFixture _pg;

    public ProjectMappingSmokeTests(PostgresContainerFixture pg) => _pg = pg;

    [Fact]
    public async Task Project_round_trips_through_EF_Core_mapping()
    {
        var id = Guid.NewGuid();
        var code = $"P_{Guid.NewGuid():N}".Substring(0, 16);
        var startDate = new DateOnly(2026, 1, 15);
        var targetDate = new DateOnly(2026, 12, 31);
        var closedAt = new DateTimeOffset(2026, 6, 1, 12, 0, 0, TimeSpan.Zero);
        var createdAt = new DateTimeOffset(2026, 1, 10, 9, 0, 0, TimeSpan.Zero);
        var updatedAt = new DateTimeOffset(2026, 1, 11, 10, 0, 0, TimeSpan.Zero);

        await using var conn = new NpgsqlConnection(_pg.ConnectionString);
        await conn.OpenAsync();
        await using var tx = await conn.BeginTransactionAsync();

        await using (
            var insert = new NpgsqlCommand(
                @"INSERT INTO domain.project
                    (id, code, name, description, status,
                     start_date, target_completion_date, actual_closed_at,
                     compliance_score, created_at, updated_at)
                  VALUES
                    (@id, @code, @name, @description, @status,
                     @start_date, @target_completion_date, @actual_closed_at,
                     @compliance_score, @created_at, @updated_at)",
                conn,
                tx
            )
        )
        {
            insert.Parameters.AddWithValue("id", id);
            insert.Parameters.AddWithValue("code", code);
            insert.Parameters.AddWithValue("name", "Smoke Project");
            insert.Parameters.AddWithValue("description", "round-trip");
            insert.Parameters.AddWithValue("status", "OnHold");
            insert.Parameters.AddWithValue("start_date", startDate);
            insert.Parameters.AddWithValue("target_completion_date", targetDate);
            insert.Parameters.AddWithValue("actual_closed_at", closedAt);
            insert.Parameters.AddWithValue("compliance_score", 87);
            insert.Parameters.AddWithValue("created_at", createdAt);
            insert.Parameters.AddWithValue("updated_at", updatedAt);
            await insert.ExecuteNonQueryAsync();
        }

        var options = new DbContextOptionsBuilder<FieldMarkDbContext>()
            .UseNpgsql(conn)
            .UseSnakeCaseNamingConvention()
            .Options;

        await using (var ctx = new FieldMarkDbContext(options))
        {
            // Enlist the DbContext in the open transaction so the EF read
            // sees the uncommitted INSERT on the same connection.
            await ctx.Database.UseTransactionAsync(tx);

            var loaded = await ctx.Projects.AsNoTracking().SingleAsync(p => p.Id == id);

            loaded.Id.Should().Be(id);
            loaded.Code.Should().Be(code);
            loaded.Name.Should().Be("Smoke Project");
            loaded.Description.Should().Be("round-trip");
            loaded.Status.Should().Be(ProjectStatus.OnHold);
            loaded.StartDate.Should().Be(startDate);
            loaded.TargetCompletionDate.Should().Be(targetDate);
            loaded.ActualClosedAt.Should().Be(closedAt);
            loaded.ComplianceScore.Should().Be(87);
            loaded.CreatedAt.Should().Be(createdAt);
            loaded.UpdatedAt.Should().Be(updatedAt);
        }

        await tx.RollbackAsync();

        // Confirm the rollback was honored — no row escaped the transaction.
        await using var checkConn = new NpgsqlConnection(_pg.ConnectionString);
        await checkConn.OpenAsync();
        await using var check = new NpgsqlCommand(
            "SELECT count(*) FROM domain.project WHERE id = @id",
            checkConn
        );
        check.Parameters.AddWithValue("id", id);
        var leaked = (long)(await check.ExecuteScalarAsync() ?? 0L);
        leaked.Should().Be(0, "rollback must not persist the row");
    }

    [Fact]
    public async Task Relation_tables_are_readable_via_DbSet_count_query()
    {
        var options = new DbContextOptionsBuilder<FieldMarkDbContext>()
            .UseNpgsql(_pg.ConnectionString)
            .UseSnakeCaseNamingConvention()
            .Options;

        await using var ctx = new FieldMarkDbContext(options);

        // Smoke that the three relation-table mappings compile and translate.
        // We only assert non-negative counts; the canonical seed leaves these
        // tables empty so a strict equality would be brittle.
        (await ctx.JobSites.CountAsync())
            .Should()
            .BeGreaterThanOrEqualTo(0);
        (await ctx.ProjectTradeScopes.CountAsync()).Should().BeGreaterThanOrEqualTo(0);
        (await ctx.ProjectInspectors.CountAsync()).Should().BeGreaterThanOrEqualTo(0);
    }
}
