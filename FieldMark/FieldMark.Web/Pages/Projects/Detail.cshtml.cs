using FieldMark.Data.Auditing;
using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using Microsoft.AspNetCore.Mvc;

namespace FieldMark.Web.Pages.Projects;

public sealed partial class DetailModel : ProjectDetailPageModelBase
{
    public DetailModel(FieldMarkDbContext db, AuthDbContext authDb, IReferenceReader reference, IAuditAppender audit)
        : base(db, authDb, reference, audit)
    {
    }

    public Task<IActionResult> OnGetAsync(Guid id, string? tab, CancellationToken ct) =>
        HandleDetailGetAsync(id, tab, ct);
}
