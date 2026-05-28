using System.Text.Json;

namespace FieldMark.Domain.Entities;

// Write-once value object for domain.audit_entry. The persisted Action column
// is stored as a string rather than the AuditAction enum so an unrecognized DB
// value surfaces as a deserialization failure on read in a future audit-read
// path (Story 2.13) rather than silently corrupting the row.
//
// Helpers (IAuditAppender) live in FieldMark.Data; this type is pure data.
public sealed class AuditEntry
{
    public Guid Id { get; private set; }
    public DateTimeOffset OccurredAt { get; private set; }
    public Guid ActorId { get; private set; }
    public string Action { get; private set; } = string.Empty;
    public string EntityType { get; private set; } = string.Empty;
    public Guid EntityId { get; private set; }
    public Guid? ProjectId { get; private set; }
    public JsonDocument? BeforeState { get; private set; }
    public JsonDocument? AfterState { get; private set; }
    public JsonDocument? Metadata { get; private set; }

    private AuditEntry() { }

    public AuditEntry(
        Guid actorId,
        string action,
        string entityType,
        Guid entityId,
        Guid? projectId = null,
        JsonDocument? beforeState = null,
        JsonDocument? afterState = null,
        JsonDocument? metadata = null
    )
    {
        Id = Guid.NewGuid();
        ActorId = actorId;
        Action = action;
        EntityType = entityType;
        EntityId = entityId;
        ProjectId = projectId;
        BeforeState = beforeState;
        AfterState = afterState;
        Metadata = metadata;
    }
}
