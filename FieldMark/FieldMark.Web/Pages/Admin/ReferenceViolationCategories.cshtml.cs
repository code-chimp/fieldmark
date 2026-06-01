using FieldMark.Data.Reference;
using FieldMark.Domain.Entities.Reference;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Admin;

[Authorize(Roles = "ADMIN")]
public sealed class ReferenceViolationCategoriesModel(IReferenceReader references) : PageModel
{
    public IReadOnlyList<ViolationCategory> ViolationCategories { get; private set; } = [];

    public async Task OnGetAsync(CancellationToken ct)
    {
        ViolationCategories = await references.ListViolationCategoriesAsync(ct);
    }
}
