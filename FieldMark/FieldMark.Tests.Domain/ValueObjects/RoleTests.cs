using FieldMark.Domain.ValueObjects;
using FluentAssertions;

namespace FieldMark.Tests.Domain.ValueObjects;

public class RoleTests
{
    [Theory]
    [InlineData("ADMIN", "Admin", "danger")]
    [InlineData("COMPLIANCE_OFFICER", "Compliance Officer", "info")]
    [InlineData("INSPECTOR", "Inspector", "warning")]
    [InlineData("SITE_SUPERVISOR", "Site Supervisor", "neutral")]
    [InlineData("EXECUTIVE", "Executive", "success")]
    public void Role_LabelAndBadgeToken_MatchCanonicalMapping(
        string name,
        string expectedLabel,
        string expectedToken
    )
    {
        var role = Role.Parse(name);
        role.Label.Should().Be(expectedLabel);
        role.BadgeToken.Should().Be(expectedToken);
    }

    [Fact]
    public void Role_All_ContainsFiveCanonicalRoles()
    {
        Role.All.Should().HaveCount(5);
    }

    [Fact]
    public void Role_Parse_UnknownName_ThrowsArgumentException()
    {
        var act = () => Role.Parse("UNKNOWN_ROLE");
        act.Should().Throw<ArgumentException>();
    }

    [Fact]
    public void Role_All_FirstOrDefault_UnknownName_ReturnsNull()
    {
        // IndexModel uses FirstOrDefault (not Parse) so unknown roles return null rather than throwing.
        var role = Role.All.FirstOrDefault(r => r.Name == "UNKNOWN_ROLE");
        role.Should().BeNull();
    }
}
