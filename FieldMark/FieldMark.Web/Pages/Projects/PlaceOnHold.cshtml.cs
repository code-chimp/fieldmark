using FieldMark.Data.Auditing;
using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using Microsoft.AspNetCore.Mvc;

namespace FieldMark.Web.Pages.Projects;

public sealed class PlaceOnHoldModel : ProjectDetailPageModelBase
{
    public PlaceOnHoldModel(FieldMarkDbContext db, AuthDbContext authDb, IReferenceReader reference, IAuditAppender audit)
        : base(db, authDb, reference, audit)
    {
    }

    public Task<IActionResult> OnGetAsync(Guid id, CancellationToken ct) =>
        OnGetTransitionAsync(id, ProjectTransitionKind.PlaceOnHold, ct);

    public Task<IActionResult> OnPostAsync(Guid id, CancellationToken ct) =>
        PostTransitionAsync(id, ProjectTransitionKind.PlaceOnHold, ct);
}
