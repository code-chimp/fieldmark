using FieldMark.Domain.Entities.Reference;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class ViolationCategoryConfiguration : IEntityTypeConfiguration<ViolationCategory>
{
    public void Configure(EntityTypeBuilder<ViolationCategory> builder)
    {
        builder.ToTable("violation_category", "domain");
        builder.HasKey(v => v.Id);

        builder.Property(v => v.Code).HasMaxLength(32).IsRequired();
        builder.Property(v => v.Name).HasMaxLength(200).IsRequired();
        builder.Property(v => v.TradeTypeId).IsRequired(false);
        builder.Property(v => v.DefaultSeverity).HasMaxLength(16).IsRequired();
        builder.Property(v => v.Description).IsRequired(false);
        builder.Property(v => v.Active).IsRequired();
    }
}
