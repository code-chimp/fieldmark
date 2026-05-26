using FieldMark.Domain.Entities;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class ProjectTradeScopeConfiguration : IEntityTypeConfiguration<ProjectTradeScope>
{
    public void Configure(EntityTypeBuilder<ProjectTradeScope> builder)
    {
        builder.ToTable("project_trade_scope", "domain");
        builder.HasKey(x => new { x.ProjectId, x.TradeTypeId });

        builder.Property(x => x.ProjectId).IsRequired();
        builder.Property(x => x.TradeTypeId).IsRequired();

        builder
            .HasOne<Project>()
            .WithMany()
            .HasForeignKey(x => x.ProjectId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
