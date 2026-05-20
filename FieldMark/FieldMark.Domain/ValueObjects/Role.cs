namespace FieldMark.Domain.ValueObjects;

/// <summary>
/// Conceptual role of an authenticated FieldMark user. The five canonical
/// names are persisted in dotnet_auth, django_auth, and fiber_auth — this
/// type is the single .NET-side source of truth for them.
/// </summary>
public sealed record Role
{
    public string Name { get; }

    private Role(string name) => Name = name;

    public static readonly Role Admin = new("ADMIN");
    public static readonly Role ComplianceOfficer = new("COMPLIANCE_OFFICER");
    public static readonly Role Inspector = new("INSPECTOR");
    public static readonly Role SiteSupervisor = new("SITE_SUPERVISOR");
    public static readonly Role Executive = new("EXECUTIVE");

    public static IReadOnlyList<Role> All { get; } =
        new[] { Admin, ComplianceOfficer, Inspector, SiteSupervisor, Executive };

    public static Role Parse(string name) =>
        All.FirstOrDefault(r => r.Name == name)
        ?? throw new ArgumentException($"Unknown role name: {name}", nameof(name));

    public override string ToString() => Name;
}
