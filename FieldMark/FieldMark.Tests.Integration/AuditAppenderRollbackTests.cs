using System;
using System.Threading.Tasks;
using FieldMark.Data.Auditing;
using FieldMark.Data.Context;
using FieldMark.Domain.ValueObjects;
using FluentAssertions;
using Microsoft.EntityFrameworkCore;
using Npgsql;

namespace FieldMark.Tests.Integration;

// Story 2.2 AC5 — load-bearing test for FR39/FR57. Proves the .NET
// AuditAppender writes inside the caller's transaction and that a rollback
// leaves zero orphaned audit rows. Plus a positive commit test with manual
// DELETE cleanup (audit_entry is append-only at app level — the cleanup
// lives only in test code, per the DDL comment).
[Collection(PostgresCollection.Name)]
public sealed class AuditAppenderRollbackTests
{
    private readonly PostgresContainerFixture _pg;

    public AuditAppenderRollbackTests(PostgresContainerFixture pg) => _pg = pg;

    [Fact]
    public async Task Append_inside_transaction_then_rollback_leaves_no_audit_row()
    {
        var projectId = Guid.NewGuid();
        var actorId = Guid.NewGuid();
        var code = $"AUD_{Guid.NewGuid():N}".Substring(0, 16);

        await using var conn = new NpgsqlConnection(_pg.ConnectionString);
        await conn.OpenAsync();
        await using var tx = await conn.BeginTransactionAsync();

        await InsertProjectAsync(conn, tx, projectId, code, status: "Active");

        var options = new DbContextOptionsBuilder<FieldMarkDbContext>()
            .UseNpgsql(conn)
            .UseSnakeCaseNamingConvention()
            .Options;

        await using (var ctx = new FieldMarkDbContext(options))
        {
            await ctx.Database.UseTransactionAsync(tx);
            var appender = new AuditAppender(ctx);
            appender.Append(
                actorId: actorId,
                action: AuditAction.ProjectCreated,
                entityType: "Project",
                entityId: projectId,
                projectId: projectId
            );
            await ctx.SaveChangesAsync();
        }

        await tx.RollbackAsync();

        // Fresh connection — confirm both project and audit row vanished.
        await using var check = new NpgsqlConnection(_pg.ConnectionString);
        await check.OpenAsync();
        (await CountAsync(check, "SELECT count(*) FROM domain.project WHERE id = @id", projectId))
            .Should()
            .Be(0, "project rollback must not persist");
        (
            await CountAsync(
                check,
                "SELECT count(*) FROM domain.audit_entry WHERE entity_id = @id",
                projectId
            )
        )
            .Should()
            .Be(0, "audit rollback must not leave an orphan row");
    }

    [Fact]
    public async Task Append_then_commit_persists_both_rows_then_cleanup_succeeds()
    {
        var projectId = Guid.NewGuid();
        var actorId = Guid.NewGuid();
        var code = $"AUD_{Guid.NewGuid():N}".Substring(0, 16);

        await using (var conn = new NpgsqlConnection(_pg.ConnectionString))
        {
            await conn.OpenAsync();
            await using var tx = await conn.BeginTransactionAsync();

            await InsertProjectAsync(conn, tx, projectId, code, status: "Active");

            var options = new DbContextOptionsBuilder<FieldMarkDbContext>()
                .UseNpgsql(conn)
                .UseSnakeCaseNamingConvention()
                .Options;

            await using (var ctx = new FieldMarkDbContext(options))
            {
                await ctx.Database.UseTransactionAsync(tx);
                var appender = new AuditAppender(ctx);
                appender.Append(
                    actorId: actorId,
                    action: AuditAction.ProjectPlacedOnHold,
                    entityType: "Project",
                    entityId: projectId,
                    projectId: projectId
                );
                await ctx.SaveChangesAsync();
            }

            await tx.CommitAsync();
        }

        try
        {
            // Fresh connection — both rows are present after commit.
            await using var check = new NpgsqlConnection(_pg.ConnectionString);
            await check.OpenAsync();
            (
                await CountAsync(
                    check,
                    "SELECT count(*) FROM domain.project WHERE id = @id",
                    projectId
                )
            )
                .Should()
                .Be(1);
            (
                await CountAsync(
                    check,
                    "SELECT count(*) FROM domain.audit_entry WHERE entity_id = @id",
                    projectId
                )
            )
                .Should()
                .Be(1);

            // Verify the persisted action string is the canonical PascalCase form.
            await using var actionCheck = new NpgsqlCommand(
                "SELECT action FROM domain.audit_entry WHERE entity_id = @id",
                check
            );
            actionCheck.Parameters.AddWithValue("id", projectId);
            var persistedAction = (string?)await actionCheck.ExecuteScalarAsync();
            persistedAction.Should().Be("ProjectPlacedOnHold");
        }
        finally
        {
            // Cleanup runs unconditionally so an assertion failure above does
            // not leak the committed test rows. Audit_entry first —
            // project_id references domain.project with no ON DELETE CASCADE.
            // Cleanup is test-only; production code never issues
            // UPDATE/DELETE against domain.audit_entry.
            await using var cleanup = new NpgsqlConnection(_pg.ConnectionString);
            await cleanup.OpenAsync();
            await ExecuteAsync(
                cleanup,
                "DELETE FROM domain.audit_entry WHERE entity_id = @id",
                projectId
            );
            await ExecuteAsync(cleanup, "DELETE FROM domain.project WHERE id = @id", projectId);
        }
    }

    private static async Task InsertProjectAsync(
        NpgsqlConnection conn,
        NpgsqlTransaction tx,
        Guid id,
        string code,
        string status
    )
    {
        await using var insert = new NpgsqlCommand(
            @"INSERT INTO domain.project
                (id, code, name, status, start_date, compliance_score, created_at, updated_at)
              VALUES
                (@id, @code, @name, @status, @start_date, @score, @created, @updated)",
            conn,
            tx
        );
        insert.Parameters.AddWithValue("id", id);
        insert.Parameters.AddWithValue("code", code);
        insert.Parameters.AddWithValue("name", "Audit Smoke Project");
        insert.Parameters.AddWithValue("status", status);
        insert.Parameters.AddWithValue("start_date", new DateOnly(2026, 1, 1));
        insert.Parameters.AddWithValue("score", 100);
        insert.Parameters.AddWithValue("created", DateTimeOffset.UtcNow);
        insert.Parameters.AddWithValue("updated", DateTimeOffset.UtcNow);
        await insert.ExecuteNonQueryAsync();
    }

    private static async Task<long> CountAsync(NpgsqlConnection conn, string sql, Guid id)
    {
        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("id", id);
        return (long)(await cmd.ExecuteScalarAsync() ?? 0L);
    }

    private static async Task ExecuteAsync(NpgsqlConnection conn, string sql, Guid id)
    {
        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("id", id);
        await cmd.ExecuteNonQueryAsync();
    }
}
