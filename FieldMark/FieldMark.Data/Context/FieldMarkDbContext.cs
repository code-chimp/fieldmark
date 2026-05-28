using FieldMark.Domain.Entities;
using FieldMark.Domain.Entities.Reference;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Data.Context;

public class FieldMarkDbContext : DbContext
{
    public FieldMarkDbContext(DbContextOptions<FieldMarkDbContext> options)
        : base(options) { }

    public DbSet<Project> Projects => Set<Project>();
    public DbSet<JobSite> JobSites => Set<JobSite>();
    public DbSet<ProjectTradeScope> ProjectTradeScopes => Set<ProjectTradeScope>();
    public DbSet<ProjectInspector> ProjectInspectors => Set<ProjectInspector>();
    public DbSet<AuditEntry> AuditEntries => Set<AuditEntry>();
    public DbSet<TradeType> TradeTypes => Set<TradeType>();
    public DbSet<ViolationCategory> ViolationCategories => Set<ViolationCategory>();
    public DbSet<ComplianceRule> ComplianceRules => Set<ComplianceRule>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);
        modelBuilder.HasDefaultSchema("domain");
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(FieldMarkDbContext).Assembly);
    }
}
