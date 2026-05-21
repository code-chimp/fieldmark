using System.Security.Claims;
using FieldMark.Domain.Services;
using FieldMark.Domain.ValueObjects;

namespace FieldMark.Web.ViewModels;

/// <summary>View model for the _AvatarMenu partial (Story 1.13).</summary>
public sealed class AvatarMenuVm
{
    public string FullName { get; }
    public string Initials { get; }
    public string RoleLabel { get; }

    public AvatarMenuVm(ClaimsPrincipal user)
    {
        FullName = user.FindFirstValue("display_name") ?? user.Identity?.Name ?? string.Empty;

        var canonicalNames = new HashSet<string>(
            Role.All.Select(r => r.Name),
            StringComparer.Ordinal
        );
        var roleName =
            user.Claims.Where(c => c.Type == ClaimTypes.Role && canonicalNames.Contains(c.Value))
                .Select(c => c.Value)
                .OrderBy(s => s, StringComparer.Ordinal)
                .FirstOrDefault()
            ?? string.Empty;

        var role = Role.All.FirstOrDefault(r => r.Name == roleName);
        RoleLabel = role?.Label ?? string.Empty;

        Initials = AvatarInitials.From(FullName, user.Identity?.Name);
    }
}
