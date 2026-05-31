using FieldMark.Web.Authorization;
using FieldMark.Web.Dashboard;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Dashboard;

public sealed class IndexModel(DashboardStatsReader statsReader) : PageModel
{
    public DashboardStats Stats { get; private set; } = new(null, null, string.Empty, null, null);

    public async Task<IActionResult> OnGetAsync(CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "dashboard.view"))
        {
            Response.StatusCode = 403;
            return Content("You do not have permission to access this page.");
        }

        Stats = await statsReader.ReadAsync(DateTimeOffset.UtcNow, ct);
        return Page();
    }
}
