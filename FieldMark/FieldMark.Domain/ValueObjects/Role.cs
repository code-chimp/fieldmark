namespace FieldMark.Domain.ValueObjects;

/// <summary>
/// Conceptual role of an authenticated FieldMark user. The five canonical
/// names are persisted in dotnet_auth, django_auth, and fiber_auth — this
/// type is the single .NET-side source of truth for them.
/// </summary>
public sealed record Role
{
    public string Name { get; }
    public string Label { get; }
    public string BadgeToken { get; }

    private Role(string name, string label, string badgeToken)
    {
        Name = name;
        Label = label;
        BadgeToken = badgeToken;
    }

    public static readonly Role Admin = new("ADMIN", "Admin", "danger");
    public static readonly Role ComplianceOfficer = new(
        "COMPLIANCE_OFFICER",
        "Compliance Officer",
        "info"
    );
    public static readonly Role Inspector = new("INSPECTOR", "Inspector", "warning");
    public static readonly Role SiteSupervisor = new(
        "SITE_SUPERVISOR",
        "Site Supervisor",
        "neutral"
    );
    public static readonly Role Executive = new("EXECUTIVE", "Executive", "success");

    public static IReadOnlyList<Role> All { get; } =
        new[] { Admin, ComplianceOfficer, Inspector, SiteSupervisor, Executive };

    public static Role Parse(string name) =>
        All.FirstOrDefault(r => r.Name == name)
        ?? throw new ArgumentException($"Unknown role name: {name}", nameof(name));

    public override string ToString() => Name;
}
