using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.AspNetCore.Hosting;
using Microsoft.AspNetCore.Identity;
using Microsoft.Extensions.DependencyInjection;

namespace FieldMark.Web.SeedData;

public static class DevUsersSeeder
{
    private sealed record ManifestEntry(
        Guid Id,
        string Username,
        string DisplayName,
        string Password,
        string? Role);

    private sealed record Manifest(List<ManifestEntry> Users);

    private static readonly JsonSerializerOptions JsonOptions = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower,
        ReadCommentHandling  = JsonCommentHandling.Skip,
    };

    public static async Task SeedAsync(
        IServiceProvider services,
        IWebHostEnvironment env,
        CancellationToken ct)
    {
        var manifestPath = Path.GetFullPath(
            Path.Combine(
                env.ContentRootPath, "..", "..",
                "docker", "postgres", "init", "seed-uuids", "dev-users.json"));

        if (!File.Exists(manifestPath))
            throw new InvalidOperationException(
                $"DevUsersSeeder: manifest not found at {manifestPath}");

        var json = await File.ReadAllTextAsync(manifestPath, ct);
        var manifest = JsonSerializer.Deserialize<Manifest>(json, JsonOptions)
            ?? throw new InvalidOperationException("DevUsersSeeder: manifest parse returned null");

        var userManager = services.GetRequiredService<UserManager<IdentityUser<Guid>>>();

        foreach (var entry in manifest.Users)
        {
            var existing = await userManager.FindByIdAsync(entry.Id.ToString());
            if (existing is null)
            {
                var user = new IdentityUser<Guid>
                {
                    Id                 = entry.Id,
                    UserName           = entry.Username,
                    NormalizedUserName = entry.Username.ToUpperInvariant(),
                    Email              = $"{entry.Username}@fieldmark.local",
                    NormalizedEmail    = $"{entry.Username}@fieldmark.local".ToUpperInvariant(),
                    EmailConfirmed     = true,
                    SecurityStamp      = Guid.NewGuid().ToString(),
                };
                var create = await userManager.CreateAsync(user, entry.Password);
                if (!create.Succeeded)
                    throw new InvalidOperationException(
                        $"DevUsersSeeder: CreateAsync failed for {entry.Username}: " +
                        string.Join("; ", create.Errors.Select(e => e.Description)));
                existing = user;
            }
            else
            {
                // Converge username, email, and password to manifest values so a re-seed
                // after a manual edit doesn't leave the database in a mismatched state.
                existing.UserName           = entry.Username;
                existing.NormalizedUserName = entry.Username.ToUpperInvariant();
                existing.Email              = $"{entry.Username}@fieldmark.local";
                existing.NormalizedEmail    = $"{entry.Username}@fieldmark.local".ToUpperInvariant();
                var update = await userManager.UpdateAsync(existing);
                if (!update.Succeeded)
                    throw new InvalidOperationException(
                        $"DevUsersSeeder: UpdateAsync failed for {entry.Username}: " +
                        string.Join("; ", update.Errors.Select(e => e.Description)));

                var token  = await userManager.GeneratePasswordResetTokenAsync(existing);
                var reset  = await userManager.ResetPasswordAsync(existing, token, entry.Password);
                if (!reset.Succeeded)
                    throw new InvalidOperationException(
                        $"DevUsersSeeder: ResetPasswordAsync failed for {entry.Username}: " +
                        string.Join("; ", reset.Errors.Select(e => e.Description)));
            }

            // Prune stale roles, then ensure the desired role is assigned.
            var currentRoles = await userManager.GetRolesAsync(existing);
            var desired = entry.Role is not null
                ? new HashSet<string>(StringComparer.OrdinalIgnoreCase) { entry.Role }
                : new HashSet<string>(StringComparer.OrdinalIgnoreCase);

            foreach (var stale in currentRoles.Where(r => !desired.Contains(r)))
            {
                var remove = await userManager.RemoveFromRoleAsync(existing, stale);
                if (!remove.Succeeded)
                    throw new InvalidOperationException(
                        $"DevUsersSeeder: RemoveFromRoleAsync({stale}) failed for {entry.Username}: " +
                        string.Join("; ", remove.Errors.Select(e => e.Description)));
            }

            if (entry.Role is not null && !currentRoles.Contains(entry.Role, StringComparer.OrdinalIgnoreCase))
            {
                var add = await userManager.AddToRoleAsync(existing, entry.Role);
                if (!add.Succeeded)
                    throw new InvalidOperationException(
                        $"DevUsersSeeder: AddToRoleAsync({entry.Role}) failed for {entry.Username}: " +
                        string.Join("; ", add.Errors.Select(e => e.Description)));
            }
        }
    }
}
