using FieldMark.Domain.Entities.Reference;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace FieldMark.Data.Configuration;

public sealed class ComplianceRuleConfiguration : IEntityTypeConfiguration<ComplianceRule>
{
    public void Configure(EntityTypeBuilder<ComplianceRule> builder)
    {
        builder.ToTable("compliance_rule", "domain");
        builder.HasKey(r => r.Id);

        builder.Property(r => r.Code).HasMaxLength(64).IsRequired();
        builder.Property(r => r.Name).HasMaxLength(200).IsRequired();
        builder.Property(r => r.Description).IsRequired();
        builder.Property(r => r.RuleKind).HasMaxLength(32).IsRequired();
        builder.Property(r => r.Parameters).HasColumnType("jsonb").IsRequired();
        builder.Property(r => r.Active).IsRequired();
    }
}
