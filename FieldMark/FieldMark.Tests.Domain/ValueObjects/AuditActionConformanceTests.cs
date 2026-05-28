using System.Text.Json;
using FieldMark.Domain.ValueObjects;
using FluentAssertions;

namespace FieldMark.Tests.Domain.ValueObjects;

// Story 2.2 AC6 — conformance gate. The native AuditAction enum must match
// the canonical list in docs/reference/audit-actions.json exactly. The fixture
// is the contract; this test prevents per-stack drift.
public class AuditActionConformanceTests
{
    [Fact]
    public void AuditAction_enum_matches_canonical_fixture()
    {
        var fixturePath = LocateFixture();
        var canonicalList = LoadCanonicalList(fixturePath);

        // Cardinality first — set equality alone masks duplicate fixture entries.
        var duplicates = canonicalList
            .GroupBy(x => x)
            .Where(g => g.Count() > 1)
            .Select(g => g.Key)
            .ToList();
        duplicates
            .Should()
            .BeEmpty(
                "audit-actions.json must not contain duplicate entries: [{0}]",
                string.Join(", ", duplicates)
            );

        var canonical = canonicalList.ToHashSet();
        var native = Enum.GetNames<AuditAction>().ToHashSet();

        var missingFromNative = canonical.Except(native).ToList();
        var extrasInNative = native.Except(canonical).ToList();

        missingFromNative
            .Should()
            .BeEmpty(
                "canonical actions are missing from the .NET enum: [{0}]",
                string.Join(", ", missingFromNative)
            );
        extrasInNative
            .Should()
            .BeEmpty(
                "the .NET enum has extra members not in the canonical fixture: [{0}]",
                string.Join(", ", extrasInNative)
            );
    }

    [Fact]
    public void AsString_round_trips_every_enum_member_to_its_symbol_name()
    {
        foreach (var member in Enum.GetValues<AuditAction>())
        {
            member.AsString().Should().Be(member.ToString());
        }
    }

    private static List<string> LoadCanonicalList(string fixturePath)
    {
        var json = File.ReadAllText(fixturePath);
        using var doc = JsonDocument.Parse(json);
        return doc
            .RootElement.GetProperty("actions")
            .EnumerateArray()
            .Select(e => e.GetString()!)
            .ToList();
    }

    // Walk up from the test bin directory to find docs/reference/audit-actions.json.
    // Pattern mirrors PostgresContainerFixture.LocateInitDir so the test runs
    // unmodified under `dotnet test` from any working directory.
    private static string LocateFixture()
    {
        var dir = new DirectoryInfo(AppContext.BaseDirectory);
        while (dir is not null)
        {
            var candidate = Path.Combine(dir.FullName, "docs", "reference", "audit-actions.json");
            if (File.Exists(candidate))
            {
                return candidate;
            }
            dir = dir.Parent;
        }
        throw new FileNotFoundException(
            "Could not locate docs/reference/audit-actions.json relative to AppContext.BaseDirectory."
        );
    }
}
