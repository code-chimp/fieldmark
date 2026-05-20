using System.Security.Claims;
using FieldMark.Web.Authentication;
using FluentAssertions;

namespace FieldMark.Tests.Web;

public class ClaimsPrincipalExtensionsTests
{
    [Fact]
    public void GetActorId_ReturnsGuid_FromNameIdentifierClaim()
    {
        var expected = Guid.Parse("01923456-7890-7abc-def0-123456789abc");
        var principal = BuildPrincipal(new Claim(ClaimTypes.NameIdentifier, expected.ToString()));

        var result = principal.GetActorId();

        result.Should().Be(expected);
    }

    [Fact]
    public void GetActorId_ThrowsWhenClaimMissing()
    {
        var principal = new ClaimsPrincipal(new ClaimsIdentity());

        var act = () => principal.GetActorId();

        act.Should().Throw<InvalidOperationException>();
    }

    [Fact]
    public void GetActorId_ThrowsWhenClaimNotGuid()
    {
        var principal = BuildPrincipal(new Claim(ClaimTypes.NameIdentifier, "not-a-guid"));

        var act = () => principal.GetActorId();

        act.Should().Throw<InvalidOperationException>();
    }

    [Fact]
    public void GetConceptualRoles_ReturnsAllRoleClaims()
    {
        var principal = BuildPrincipal(
            new Claim(ClaimTypes.Role, "ADMIN"),
            new Claim(ClaimTypes.Role, "COMPLIANCE_OFFICER")
        );

        var roles = principal.GetConceptualRoles();

        roles.Should().Equal("ADMIN", "COMPLIANCE_OFFICER");
    }

    private static ClaimsPrincipal BuildPrincipal(params Claim[] claims) =>
        new(new ClaimsIdentity(claims, "test"));
}
