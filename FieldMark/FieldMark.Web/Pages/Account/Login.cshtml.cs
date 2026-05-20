using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Account;

[AllowAnonymous]
public class LoginModel : PageModel
{
    private readonly SignInManager<IdentityUser<Guid>> _signInManager;

    public LoginModel(SignInManager<IdentityUser<Guid>> signInManager)
    {
        _signInManager = signInManager;
    }

    [BindProperty]
    public string Username { get; set; } = "";

    [BindProperty]
    public string Password { get; set; } = "";

    [BindProperty(Name = "return_url", SupportsGet = true)]
    public string? ReturnUrl { get; set; }

    public bool HasErrors => !ModelState.IsValid;

    public IDictionary<string, string?> FieldErrors { get; } = new Dictionary<string, string?>();

    public IActionResult OnGet()
    {
        if (User.Identity?.IsAuthenticated == true)
        {
            return LocalRedirect(Url.IsLocalUrl(ReturnUrl) ? ReturnUrl! : "/");
        }
        return Page();
    }

    public async Task<IActionResult> OnPostAsync()
    {
        if (string.IsNullOrWhiteSpace(Username))
        {
            FieldErrors["field-username"] = "Username is required.";
            ModelState.AddModelError("Username", "Username is required.");
        }

        if (string.IsNullOrWhiteSpace(Password))
        {
            FieldErrors["field-password"] = "Password is required.";
            ModelState.AddModelError("Password", "Password is required.");
        }

        if (!ModelState.IsValid)
        {
            Response.StatusCode = 422;
            return Page();
        }

        var result = await _signInManager.PasswordSignInAsync(
            Username, Password, isPersistent: true, lockoutOnFailure: false);

        if (result.Succeeded)
        {
            if (Url.IsLocalUrl(ReturnUrl))
            {
                return LocalRedirect(ReturnUrl);
            }
            return LocalRedirect("/");
        }

        ModelState.AddModelError(string.Empty, "Invalid username or password.");
        FieldErrors["field-username"] = "";
        FieldErrors["field-password"] = "";
        Response.StatusCode = 422;
        return Page();
    }
}
