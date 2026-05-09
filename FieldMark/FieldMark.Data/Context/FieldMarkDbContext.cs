using Microsoft.EntityFrameworkCore;

namespace FieldMark.Data.Context;

public class FieldMarkDbContext : DbContext
{
    public FieldMarkDbContext(DbContextOptions<FieldMarkDbContext> options)
        : base(options) { }

    // DbSet<Project> Projects;
}
