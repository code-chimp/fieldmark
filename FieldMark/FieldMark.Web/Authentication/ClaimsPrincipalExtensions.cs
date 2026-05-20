using System.Security.Claims;

namespace FieldMark.Web.Authentication;

public static class ClaimsPrincipalExtensions
{
    public static Guid GetActorId(this ClaimsPrincipal user)
    {
        var raw = user.FindFirstValue(ClaimTypes.NameIdentifier);
        if (string.IsNullOrWhiteSpace(raw) || !Guid.TryParse(raw, out var id))
        {
            throw new InvalidOperationException(
                "GetActorId called on an unauthenticated or claim-less principal. "
                    + "Guard with User.Identity.IsAuthenticated or use the [Authorize] attribute."
            );
        }
        return id;
    }

    public static IReadOnlyList<string> GetConceptualRoles(this ClaimsPrincipal user) =>
        user.FindAll(ClaimTypes.Role).Select(c => c.Value).ToList();
}
