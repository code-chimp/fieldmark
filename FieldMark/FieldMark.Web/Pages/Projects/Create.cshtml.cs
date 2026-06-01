// GET /projects/new — renders the project-create form.
// See docs/reference/project-create-form-contract.md.
using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using FieldMark.Domain.ValueObjects;
using FieldMark.Web.Authorization;
using FieldMark.Web.ViewModels.Projects;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Web.Pages.Projects;

public sealed class CreateModel : PageModel
{
    private readonly IReferenceReader _reference;
    private readonly AuthDbContext _authDb;
    private readonly UserManager<IdentityUser<Guid>> _userManager;

    public CreateModel(
        IReferenceReader reference,
        AuthDbContext authDb,
        UserManager<IdentityUser<Guid>> userManager
    )
    {
        _reference = reference;
        _authDb = authDb;
        _userManager = userManager;
    }

    public ProjectCreateFormVm Form { get; private set; } = new();

    public async Task<IActionResult> OnGetAsync(CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "project.create"))
            return StatusCode(403, "You do not have permission to access this page.");

        Form = await BuildEmptyFormVm(ct);
        return Page();
    }

    private async Task<ProjectCreateFormVm> BuildEmptyFormVm(CancellationToken ct)
    {
        var tradeTypes = await _reference.ListTradeTypesAsync(ct);
        // Filter locked-out users for parity with Django's is_active=True check.
        var now = DateTimeOffset.UtcNow;
        var allInspectors = await _userManager.GetUsersInRoleAsync(Role.Inspector.Name);
        var inspectorUsers = allInspectors
            .Where(u => u.LockoutEnd == null || u.LockoutEnd < now)
            .ToList();
        var displayNames = await GetDisplayNamesAsync(inspectorUsers.Select(u => u.Id));

        return new ProjectCreateFormVm
        {
            AvailableTradeTypes = tradeTypes.Where(t => t.Active).ToList(),
            AvailableInspectors = inspectorUsers
                .Select(u =>
                    new InspectorOption(
                        u.Id,
                        displayNames.GetValueOrDefault(u.Id, u.UserName ?? u.Id.ToString())
                    )
                )
                .OrderBy(o => o.Label)
                .ToList(),
        };
    }

    private async Task<Dictionary<Guid, string>> GetDisplayNamesAsync(IEnumerable<Guid> userIds)
    {
        var ids = userIds.ToHashSet();
        // GroupBy then take the first value to guard against duplicate display_name claims
        // (shouldn't occur in practice but would cause ToDictionaryAsync to throw).
        var claims = await _authDb
            .UserClaims.Where(c => ids.Contains(c.UserId) && c.ClaimType == "display_name")
            .ToListAsync();
        return claims
            .GroupBy(c => c.UserId)
            .ToDictionary(g => g.Key, g => g.First().ClaimValue ?? "");
    }
}
