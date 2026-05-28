using System.Text.Json;
using FieldMark.Domain.ValueObjects;

namespace FieldMark.Data.Auditing;

// The single .NET-side helper for FR39/FR40. Handlers call Append() inside
// their own open transaction; the implementation does NOT call SaveChanges
// and does NOT open a transaction. See docs/reference/audit-actions.md
// and the canonical request flow in architecture.md §Process Patterns.
public interface IAuditAppender
{
    void Append(
        Guid actorId,
        AuditAction action,
        string entityType,
        Guid entityId,
        Guid? projectId = null,
        JsonDocument? beforeState = null,
        JsonDocument? afterState = null,
        JsonDocument? metadata = null
    );
}
