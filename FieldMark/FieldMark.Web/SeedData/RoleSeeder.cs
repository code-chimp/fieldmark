using Microsoft.AspNetCore.Identity;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Web.SeedData;

internal static class RoleSeeder
{
    private static readonly string[] CanonicalRoles =
    {
        "ADMIN",
        "COMPLIANCE_OFFICER",
        "INSPECTOR",
        "SITE_SUPERVISOR",
        "EXECUTIVE",
    };

    internal static async Task SeedAsync(IServiceProvider services, CancellationToken ct)
    {
        var roleManager = services.GetRequiredService<RoleManager<IdentityRole<Guid>>>();
        foreach (var name in CanonicalRoles)
        {
            if (!await roleManager.RoleExistsAsync(name))
            {
                var role = new IdentityRole<Guid>(name) { Id = Guid.NewGuid() };
                var result = await roleManager.CreateAsync(role);
                if (!result.Succeeded)
                {
                    throw new InvalidOperationException(
                        $"Failed to seed role '{name}': {string.Join("; ", result.Errors.Select(e => e.Description))}"
                    );
                }
            }
        }
    }
}
