using FieldMark.Domain.Entities;
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

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        base.OnModelCreating(modelBuilder);
        modelBuilder.HasDefaultSchema("domain");
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(FieldMarkDbContext).Assembly);
    }
}
