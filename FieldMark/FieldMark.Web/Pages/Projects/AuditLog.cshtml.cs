using FieldMark.Data.Auditing;
using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Projects;

public sealed class AuditLogModel : ProjectDetailPageModelBase
{
    public AuditLogModel(FieldMarkDbContext db, AuthDbContext authDb, IReferenceReader reference, IAuditAppender audit)
        : base(db, authDb, reference, audit)
    {
    }

    public Task<IActionResult> OnGetAsync(
        Guid id,
        [FromQuery(Name = "before_occurred_at")] string? beforeOccurredAt,
        [FromQuery(Name = "before_id")] string? beforeId,
        CancellationToken ct
    ) => HandleAuditLogGetAsync(id, beforeOccurredAt, beforeId, ct);
}
