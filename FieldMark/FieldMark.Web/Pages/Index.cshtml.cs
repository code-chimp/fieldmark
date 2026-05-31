using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages;

public sealed class IndexModel : PageModel
{
    public string RoleLabel { get; private set; } = string.Empty;
    public string RoleBadgeToken { get; private set; } = "unknown";
    public string FullName { get; private set; } = string.Empty;
    public string Initials { get; private set; } = "??";

    public IActionResult OnGet()
    {
        return Redirect("/dashboard");
    }
}
