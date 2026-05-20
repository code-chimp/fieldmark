using System.Security.Claims;
using FieldMark.Domain.ValueObjects;
using FieldMark.Web.Authorization;
using FluentAssertions;

namespace FieldMark.Tests.Web.Authorization;

/// <summary>
/// Pure-logic unit tests for <see cref="DomainPolicies.Can"/>.
/// No server, no database — hand-built ClaimsPrincipal only.
/// </summary>
public sealed class CanTests : IDisposable
{
    // Reset the static ActionRoleMap before each test to prevent bleed.
    public CanTests() => DomainPolicies.ResetForTests();

    public void Dispose() => DomainPolicies.ResetForTests();

    [Fact]
    public void Can_AnonymousActor_ReturnsFalse()
    {
        DomainPolicies.RegisterAction("test.allow_admin", Role.Admin);
        var anonymous = new ClaimsPrincipal(new ClaimsIdentity());

        DomainPolicies.Can(anonymous, "test.allow_admin").Should().BeFalse();
    }

    [Fact]
    public void Can_AdminActor_ReturnsTrueForAdminScopedAction()
    {
        DomainPolicies.RegisterAction("test.allow_admin", Role.Admin);
        var admin = BuildPrincipal("ADMIN");

        DomainPolicies.Can(admin, "test.allow_admin").Should().BeTrue();
    }

    [Fact]
    public void Can_NonAdminActor_ReturnsFalseForAdminScopedAction()
    {
        DomainPolicies.RegisterAction("test.allow_admin", Role.Admin);
        var supervisor = BuildPrincipal("SITE_SUPERVISOR");

        DomainPolicies.Can(supervisor, "test.allow_admin").Should().BeFalse();
    }

    [Fact]
    public void Can_UnknownAction_ReturnsFalse()
    {
        var admin = BuildPrincipal("ADMIN");

        DomainPolicies.Can(admin, "test.unmapped").Should().BeFalse();
    }

    private static ClaimsPrincipal BuildPrincipal(string role)
    {
        var claims = new[]
        {
            new Claim(ClaimTypes.NameIdentifier, Guid.NewGuid().ToString()),
            new Claim(ClaimTypes.Name, "testuser"),
            new Claim(ClaimTypes.Role, role),
        };
        return new ClaimsPrincipal(new ClaimsIdentity(claims, authenticationType: "Test"));
    }
}
