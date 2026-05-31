// GET /projects — project list page with AG Grid SSRM panel.
// See docs/reference/ag-grid-ssrm-contract.md
using FieldMark.Web.Authorization;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Projects;

public sealed class ListModel : PageModel
{
    public bool CanCreate { get; private set; }

    public IActionResult OnGet()
    {
        if (!DomainPolicies.Can(User, "project.read"))
        {
            Response.StatusCode = 403;
            return Content("You do not have permission to access this page.");
        }

        CanCreate = DomainPolicies.Can(User, "project.create");
        return Page();
    }
}
