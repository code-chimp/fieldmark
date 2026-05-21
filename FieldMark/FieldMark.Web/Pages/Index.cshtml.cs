using System.Security.Claims;
using FieldMark.Domain.Services;
using FieldMark.Domain.ValueObjects;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages;

public sealed partial class IndexModel(ILogger<IndexModel> logger) : PageModel
{
    public string RoleLabel { get; private set; } = string.Empty;
    public string RoleBadgeToken { get; private set; } = "unknown";
    public string FullName { get; private set; } = string.Empty;
    public string Initials { get; private set; } = "??";

    public void OnGet()
    {
        // Two-key sort: canonical roles first (so a user with both a canonical and an
        // unknown role always displays the canonical badge), then alphabetically within
        // each tier. A pure lexical sort would let "ANALYST" outrank "COMPLIANCE_OFFICER"
        // and render badge-unknown incorrectly. The warning branch below still fires when
        // the selected role is unknown (i.e., the user has no canonical role at all).
        var canonicalNames = new HashSet<string>(
            Role.All.Select(r => r.Name),
            StringComparer.Ordinal
        );
        var roleName =
            User.Claims.Where(c => c.Type == ClaimTypes.Role)
                .Select(c => c.Value)
                .OrderByDescending(v => canonicalNames.Contains(v))
                .ThenBy(v => v, StringComparer.Ordinal)
                .FirstOrDefault()
            ?? string.Empty;

        var role = Role.All.FirstOrDefault(r => r.Name == roleName);
        if (role is not null)
        {
            RoleLabel = role.Label;
            RoleBadgeToken = role.BadgeToken;
        }
        else if (!string.IsNullOrEmpty(roleName))
        {
            LogUnknownRoleBadgeToken(roleName);
        }

        FullName = User.FindFirstValue("display_name") ?? User.Identity?.Name ?? string.Empty;
        Initials = AvatarInitials.From(FullName, User.Identity?.Name);
    }

    [LoggerMessage(
        Level = LogLevel.Warning,
        Message = "Unknown role badge token for role: {RoleName}"
    )]
    private partial void LogUnknownRoleBadgeToken(string roleName);
}
