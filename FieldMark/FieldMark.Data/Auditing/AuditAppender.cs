using System.Text.Json;
using FieldMark.Data.Context;
using FieldMark.Domain.Entities;
using FieldMark.Domain.ValueObjects;

namespace FieldMark.Data.Auditing;

public sealed class AuditAppender : IAuditAppender
{
    private readonly FieldMarkDbContext _db;

    public AuditAppender(FieldMarkDbContext db) => _db = db;

    public void Append(
        Guid actorId,
        AuditAction action,
        string entityType,
        Guid entityId,
        Guid? projectId = null,
        JsonDocument? beforeState = null,
        JsonDocument? afterState = null,
        JsonDocument? metadata = null
    )
    {
        _db.AuditEntries.Add(
            new AuditEntry(
                actorId,
                action.AsString(),
                entityType,
                entityId,
                projectId,
                beforeState,
                afterState,
                metadata
            )
        );
    }
}
