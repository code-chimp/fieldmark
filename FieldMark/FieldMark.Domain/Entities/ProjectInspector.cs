namespace FieldMark.Domain.Entities;

// user_id is an opaque UUID per ADR-012 — no navigation property to any
// Identity user type. Mapping enforces this (no relationship configured).
public class ProjectInspector
{
    public Guid ProjectId { get; private set; }
    public Guid UserId { get; private set; }

    private ProjectInspector() { }
}
