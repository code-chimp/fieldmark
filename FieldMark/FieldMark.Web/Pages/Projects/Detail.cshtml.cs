// GET /projects/{id} — minimal stub redirect target for Story 2.8.
// Story 2.11 replaces this with the full Project Detail screen.
// Authentication is required via the global fallback policy (Program.cs).
using FieldMark.Data.Context;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Web.Pages.Projects;

public sealed class DetailModel : PageModel
{
    private readonly FieldMarkDbContext _db;

    public DetailModel(FieldMarkDbContext db) => _db = db;

    public string ProjectName { get; private set; } = string.Empty;

    public async Task<IActionResult> OnGetAsync(Guid id, CancellationToken ct)
    {
        var project = await _db
            .Projects.AsNoTracking()
            .Where(p => p.Id == id)
            .Select(p => new { p.Name })
            .FirstOrDefaultAsync(ct);

        if (project is null)
            return NotFound();

        ProjectName = project.Name;
        return Page();
    }
}
