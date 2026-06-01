using FieldMark.Domain.ValueObjects;

namespace FieldMark.Domain.Entities;

// Property bag + factory for domain.project.
// See docs/reference/project-create-form-contract.md for the form contract.
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

    /// <summary>
    /// Status-only gate for Story 2.11 action affordances.
    /// Epic 6 adds additional closure gate checks (open violations / required inspections).
    /// </summary>
    public bool CanPlaceOnHold() => Status == ProjectStatus.Active;

    /// <summary>
    /// Status-only gate for Story 2.11 action affordances.
    /// </summary>
    public bool CanResume() => Status == ProjectStatus.OnHold;

    /// <summary>
    /// Status-only gate for Story 2.11 action affordances.
    /// Epic 6 adds additional closure gate checks (open violations / required inspections).
    /// </summary>
    public bool CanClose() => Status == ProjectStatus.Active;

    /// <summary>
    /// Creates a new Project with its join collections. Returns a
    /// <see cref="CreatedProject"/> wrapper so the handler can persist all
    /// four row-sets without guessing IDs.
    /// </summary>
    /// <exception cref="ArgumentException">Required string is null or empty.</exception>
    /// <exception cref="ArgumentOutOfRangeException">targetCompletionDate is before startDate, or tradeScopeIds is empty.</exception>
    public static CreatedProject Create(
        string code,
        string name,
        string? description,
        DateOnly startDate,
        DateOnly? targetCompletionDate,
        IReadOnlyList<Guid> tradeScopeIds,
        IReadOnlyList<Guid> inspectorIds
    )
    {
        if (string.IsNullOrWhiteSpace(code))
            throw new ArgumentException("Code is required.", nameof(code));
        if (string.IsNullOrWhiteSpace(name))
            throw new ArgumentException("Name is required.", nameof(name));
        if (tradeScopeIds.Count == 0)
            throw new ArgumentOutOfRangeException(
                nameof(tradeScopeIds),
                "At least one trade scope is required."
            );
        if (targetCompletionDate.HasValue && targetCompletionDate.Value < startDate)
            throw new ArgumentOutOfRangeException(
                nameof(targetCompletionDate),
                "Target completion date must be on or after the start date."
            );

        var id = Guid.NewGuid();
        var project = new Project
        {
            Id = id,
            Code = code.Trim(),
            Name = name.Trim(),
            Description = string.IsNullOrWhiteSpace(description) ? null : description.Trim(),
            Status = ProjectStatus.Active,
            StartDate = startDate,
            TargetCompletionDate = targetCompletionDate,
            ComplianceScore = 100,
        };

        var scopes = tradeScopeIds
            .Select(tid => new ProjectTradeScope(id, tid))
            .ToArray();

        var inspectors = inspectorIds
            .Select(uid => new ProjectInspector(id, uid))
            .ToArray();

        return new CreatedProject(project, scopes, inspectors);
    }
}

/// <summary>Wrapper returned by <see cref="Project.Create"/> carrying the new
/// aggregate root and its join-table rows so the handler can persist them as a
/// single unit without reconstructing IDs.</summary>
public sealed record CreatedProject(
    Project Project,
    ProjectTradeScope[] Scopes,
    ProjectInspector[] Inspectors
);
