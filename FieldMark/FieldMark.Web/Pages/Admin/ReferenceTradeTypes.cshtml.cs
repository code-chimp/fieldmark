using FieldMark.Data.Reference;
using FieldMark.Domain.Entities.Reference;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Admin;

[Authorize(Roles = "ADMIN")]
public sealed class ReferenceTradeTypesModel(IReferenceReader references) : PageModel
{
    public IReadOnlyList<TradeType> TradeTypes { get; private set; } = [];

    public async Task OnGetAsync(CancellationToken ct)
    {
        TradeTypes = await references.ListTradeTypesAsync(ct);
    }
}
