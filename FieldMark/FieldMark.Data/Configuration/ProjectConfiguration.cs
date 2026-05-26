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
        builder.Property(p => p.CreatedAt).IsRequired();
        builder.Property(p => p.UpdatedAt).IsRequired();

        // The DDL declares `code VARCHAR(32) UNIQUE`. We model it as an
        // AlternateKey rather than HasIndex so EF Core does not introduce a
        // phantom index entry in pg_indexes — `make parity` would flag a
        // .NET-only index. The DDL owns the uniqueness constraint.
        builder.HasAlternateKey(p => p.Code);
    }
}
