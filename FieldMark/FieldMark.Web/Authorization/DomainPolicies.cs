using System.Security.Claims;
using FieldMark.Domain.ValueObjects;

namespace FieldMark.Web.Authorization;

/// <summary>
/// The single .NET-side authorization decision call site (FR5).
/// Epic 1: role-only checks (entity-scope rules deferred to Epic 2+).
/// </summary>
public static class DomainPolicies
{
    // Action → roles permitted. Stories from Epic 2+ register their actions
    // by appending to this map at composition time (see RegisterAction).
    // Story 1.12 ships the map empty — there are no live actions in Epic 1.
    private static readonly Dictionary<string, HashSet<string>> ActionRoleMap = new();

    /// <summary>
    /// Register an action → permitted-roles mapping. Call from Program.cs
    /// or from a per-aggregate registrar (e.g., ProjectPolicies.Register()).
    /// </summary>
    public static void RegisterAction(string action, params Role[] roles)
    {
        if (!ActionRoleMap.TryGetValue(action, out var set))
        {
            set = new HashSet<string>();
            ActionRoleMap[action] = set;
        }
        foreach (var role in roles)
            set.Add(role.Name);
    }

    /// <summary>
    /// Return true if the user is authenticated and permitted to perform
    /// action (optionally scoped to entityId).
    /// </summary>
    public static bool Can(ClaimsPrincipal user, string action, Guid? entityId = null)
    {
        if (user.Identity is not { IsAuthenticated: true })
            return false;
        if (!ActionRoleMap.TryGetValue(action, out var permittedRoles))
            return false;

        foreach (var permittedRole in permittedRoles)
        {
            if (user.IsInRole(permittedRole))
                return EvaluateEntityScope(action, entityId);
        }
        return false;
    }

    // Single extension point for Epic 2+ entity-scope rules (e.g.,
    // "Site Supervisor can act on a Violation only if assigned to it").
    // Today every action is role-coarse; flip individual entries to do
    // entity-scope work when that story arrives.
    private static bool EvaluateEntityScope(string action, Guid? entityId) => true;

    // Test-only escape hatch. Production callers must use RegisterAction.
    internal static void ResetForTests() => ActionRoleMap.Clear();
}
