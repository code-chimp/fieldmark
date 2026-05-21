using System.Security.Claims;
using FieldMark.Domain.Services;
using FieldMark.Domain.ValueObjects;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages;

public class IndexModel : PageModel
{
    public string RoleLabel { get; private set; } = string.Empty;
    public string RoleBadgeToken { get; private set; } = "neutral";
    public string FullName { get; private set; } = string.Empty;
    public string Initials { get; private set; } = "??";

    public void OnGet()
    {
        var canonicalNames = new HashSet<string>(
            Role.All.Select(r => r.Name),
            StringComparer.Ordinal
        );
        var roleName =
            User.Claims.Where(c => c.Type == ClaimTypes.Role && canonicalNames.Contains(c.Value))
                .Select(c => c.Value)
                .OrderBy(s => s, StringComparer.Ordinal)
                .FirstOrDefault()
            ?? string.Empty;

        var role = Role.All.FirstOrDefault(r => r.Name == roleName);
        if (role is not null)
        {
            RoleLabel = role.Label;
            RoleBadgeToken = role.BadgeToken;
        }

        FullName = User.FindFirstValue("display_name") ?? User.Identity?.Name ?? string.Empty;
        Initials = AvatarInitials.From(FullName, User.Identity?.Name);
    }
}
