namespace FieldMark.Domain.Entities;

// user_id is an opaque UUID per ADR-012 — no navigation property to any
// Identity user type. Mapping enforces this (no relationship configured).
public class ProjectInspector
{
    public Guid ProjectId { get; private set; }
    public Guid UserId { get; private set; }

    private ProjectInspector() { }

    // Internal constructor for use by Project.Create factory only.
    internal ProjectInspector(Guid projectId, Guid userId)
    {
        ProjectId = projectId;
        UserId = userId;
    }
}
