using FieldMark.Domain.ValueObjects;

namespace FieldMark.Domain.Entities;

// Property bag for domain.project. Behavior methods (Create, place_on_hold,
// resume, close, RecomputeComplianceScore) land in Stories 2.8 / 2.12 / 6.x.
public class Project
{
    public Guid Id { get; private set; }
    public string Code { get; private set; } = string.Empty;
    public string Name { get; private set; } = string.Empty;
    public string? Description { get; private set; }
    public ProjectStatus Status { get; private set; }
    public DateOnly StartDate { get; private set; }
    public DateOnly? TargetCompletionDate { get; private set; }
    public DateTimeOffset? ActualClosedAt { get; private set; }
    public int ComplianceScore { get; private set; }
    public DateTimeOffset CreatedAt { get; private set; }
    public DateTimeOffset UpdatedAt { get; private set; }

    private Project() { }
}
