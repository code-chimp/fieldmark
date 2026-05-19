using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Preferences;

// Theme preference is non-security-sensitive UI state; no CSRF protection required.
[IgnoreAntiforgeryToken]
public class ThemeModel : PageModel
{
    private static readonly HashSet<string> AllowedValues = new(StringComparer.Ordinal)
    {
        "system",
        "light",
        "dark",
    };

    public IActionResult OnPost()
    {
        var value = Request.Form["value"].ToString();

        if (!AllowedValues.Contains(value))
            return BadRequest();

        Response.Cookies.Append(
            "fm_theme",
            value,
            new CookieOptions
            {
                Path = "/",
                SameSite = SameSiteMode.Lax,
                MaxAge = TimeSpan.FromSeconds(31536000),
            }
        );

        Response.Headers["HX-Trigger"] = "theme-changed";
        return new StatusCodeResult(204);
    }
}
