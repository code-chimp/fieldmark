using FieldMark.Domain.Entities.Reference;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class TradeTypeConfiguration : IEntityTypeConfiguration<TradeType>
{
    public void Configure(EntityTypeBuilder<TradeType> builder)
    {
        builder.ToTable("trade_type", "domain");
        builder.HasKey(t => t.Id);

        builder.Property(t => t.Code).HasMaxLength(32).IsRequired();
        builder.Property(t => t.Name).HasMaxLength(120).IsRequired();
        builder.Property(t => t.Description).IsRequired(false);
        builder.Property(t => t.Active).IsRequired();
    }
}
