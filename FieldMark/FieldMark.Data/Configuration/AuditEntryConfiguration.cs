using FieldMark.Domain.Entities;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class AuditEntryConfiguration : IEntityTypeConfiguration<AuditEntry>
{
    public void Configure(EntityTypeBuilder<AuditEntry> builder)
    {
        builder.ToTable("audit_entry", "domain");
        builder.HasKey(a => a.Id);

        // The DDL sets `DEFAULT now()` on occurred_at; tell EF Core the server
        // assigns it so inserts skip the column and reads pick the value back.
        builder.Property(a => a.OccurredAt).HasDefaultValueSql("now()").ValueGeneratedOnAdd();

        builder.Property(a => a.ActorId).IsRequired();
        builder.Property(a => a.Action).HasMaxLength(64).IsRequired();
        builder.Property(a => a.EntityType).HasMaxLength(64).IsRequired();
        builder.Property(a => a.EntityId).IsRequired();
        builder.Property(a => a.ProjectId).IsRequired(false);

        // Npgsql maps System.Text.Json.JsonDocument to jsonb natively. Nullable
        // (no payload) is distinct from `'null'::jsonb` — do not coalesce.
        builder.Property(a => a.BeforeState).HasColumnType("jsonb");
        builder.Property(a => a.AfterState).HasColumnType("jsonb");
        builder.Property(a => a.Metadata).HasColumnType("jsonb");

        // idx_audit_entity and idx_audit_project are DDL-owned (see
        // docker/postgres/init/010_domain_tables.sql). No HasIndex calls here —
        // `make parity` would flag a phantom .NET-only index. Same resolution
        // pattern as Story 2.1 used for Project.code UNIQUE.
    }
}
