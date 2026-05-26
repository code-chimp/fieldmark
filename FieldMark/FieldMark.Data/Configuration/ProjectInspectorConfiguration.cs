using FieldMark.Domain.Entities;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class ProjectInspectorConfiguration : IEntityTypeConfiguration<ProjectInspector>
{
    public void Configure(EntityTypeBuilder<ProjectInspector> builder)
    {
        builder.ToTable("project_inspector", "domain");
        builder.HasKey(x => new { x.ProjectId, x.UserId });

        builder.Property(x => x.ProjectId).IsRequired();
        // ADR-012: user_id is an opaque UUID; no relational FK to any auth schema.
        builder.Property(x => x.UserId).IsRequired();

        builder
            .HasOne<Project>()
            .WithMany()
            .HasForeignKey(x => x.ProjectId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
