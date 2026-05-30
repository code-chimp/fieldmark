using FieldMark.Domain.Entities;
using FieldMark.Domain.ValueObjects;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class ProjectConfiguration : IEntityTypeConfiguration<Project>
{
    public void Configure(EntityTypeBuilder<Project> builder)
    {
        builder.ToTable("project", "domain");
        builder.HasKey(p => p.Id);

        builder.Property(p => p.Code).HasMaxLength(32).IsRequired();
        builder.Property(p => p.Name).HasMaxLength(200).IsRequired();
        builder.Property(p => p.Description);
        builder.Property(p => p.Status).HasConversion<string>().HasMaxLength(16).IsRequired();
        builder.Property(p => p.StartDate).IsRequired();
        builder.Property(p => p.ComplianceScore).IsRequired();
        // The DDL sets `DEFAULT now()` on both columns; tell EF Core the server
        // assigns them so inserts omit these columns and reads pick the values back.
        // Matches the pattern used in AuditEntryConfiguration for occurred_at.
        builder.Property(p => p.CreatedAt).HasDefaultValueSql("now()").ValueGeneratedOnAdd();
        builder.Property(p => p.UpdatedAt).HasDefaultValueSql("now()").ValueGeneratedOnAdd();

        // The DDL declares `code VARCHAR(32) UNIQUE`. We model it as an
        // AlternateKey rather than HasIndex so EF Core does not introduce a
        // phantom index entry in pg_indexes — `make parity` would flag a
        // .NET-only index. The DDL owns the uniqueness constraint.
        builder.HasAlternateKey(p => p.Code);
    }
}
