using FieldMark.Domain.Entities;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class JobSiteConfiguration : IEntityTypeConfiguration<JobSite>
{
    public void Configure(EntityTypeBuilder<JobSite> builder)
    {
        builder.ToTable("job_site", "domain");
        builder.HasKey(j => j.Id);

        builder.Property(j => j.ProjectId).IsRequired();
        builder.Property(j => j.Label).HasMaxLength(120).IsRequired();
        builder.Property(j => j.Address).HasMaxLength(300);

        builder
            .HasOne<Project>()
            .WithMany()
            .HasForeignKey(j => j.ProjectId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
