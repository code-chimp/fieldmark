using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using FieldMark.Domain.ValueObjects;
using FieldMark.Web.Authorization;
using FieldMark.Web.ViewModels.Components;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Web.Pages.Projects;

public sealed class DetailModel : PageModel
{
    private readonly FieldMarkDbContext _db;
    private readonly AuthDbContext _authDb;
    private readonly IReferenceReader _reference;

    public DetailModel(FieldMarkDbContext db, AuthDbContext authDb, IReferenceReader reference)
    {
        _db = db;
        _authDb = authDb;
        _reference = reference;
    }

    public sealed record TabSpec(string Id, string Label, string HxGet, string HxTarget, int? BadgeCount = null);
    public sealed record SummaryVm(
        string Code,
        string Name,
        DateOnly StartDate,
        DateOnly? TargetCompletionDate,
        string? Description,
        IReadOnlyList<string> TradeNames,
        IReadOnlyList<string> InspectorNames,
        ActionButtonVm PlaceOnHold,
        ActionButtonVm Resume,
        ActionButtonVm Close
    );

    public string ProjectName { get; private set; } = string.Empty;
    public string ProjectCode { get; private set; } = string.Empty;
    public string ProjectStatus { get; private set; } = string.Empty;
    public int ComplianceScore { get; private set; }
    public SummaryVm? Summary { get; private set; }
    public IReadOnlyList<TabSpec> Tabs { get; private set; } = [];
    public int ActiveTabIndex { get; private set; }
    public bool IsTabResponse { get; private set; }
    public string ActiveTabId { get; private set; } = "tab-summary";

    public async Task<IActionResult> OnGetAsync(Guid id, string? tab, CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "project.read"))
        {
            Response.StatusCode = 403;
            return Content("You do not have permission to access this page.");
        }

        var project = await _db
            .Projects.AsNoTracking()
            .FirstOrDefaultAsync(p => p.Id == id, ct);

        if (project is null)
            return NotFound();

        var scopes = await _db.ProjectTradeScopes.AsNoTracking().Where(x => x.ProjectId == id).ToListAsync(ct);
        var inspectors = await _db.ProjectInspectors.AsNoTracking().Where(x => x.ProjectId == id).ToListAsync(ct);
        var tradeTypes = await _reference.ListTradeTypesAsync(ct);

        var tradeById = tradeTypes.ToDictionary(t => t.Id, t => t);
        var tradeNames = scopes
            .Where(s => tradeById.ContainsKey(s.TradeTypeId))
            .Select(s =>
            {
                var trade = tradeById[s.TradeTypeId];
                return trade.Active ? trade.Name : $"{trade.Name} (inactive)";
            })
            .ToList();

        var inspectorIds = inspectors.Select(i => i.UserId).ToHashSet();
        var users = await _authDb.Users.AsNoTracking().Where(u => inspectorIds.Contains(u.Id)).ToListAsync(ct);
        var userClaims = await _authDb.UserClaims.AsNoTracking()
            .Where(c => inspectorIds.Contains(c.UserId) && c.ClaimType == "display_name")
            .ToListAsync(ct);
        var claimMap = userClaims
            .GroupBy(c => c.UserId)
            .ToDictionary(g => g.Key, g => g.First().ClaimValue);
        var userMap = users.ToDictionary(
            u => u.Id,
            u => claimMap.TryGetValue(u.Id, out var display) && !string.IsNullOrWhiteSpace(display)
                ? display
                : u.UserName ?? u.Id.ToString()
        );
        var inspectorNames = inspectors.Where(i => userMap.ContainsKey(i.UserId)).Select(i => userMap[i.UserId]).ToList();

        ProjectName = project.Name;
        ProjectCode = project.Code;
        ProjectStatus = project.Status.ToString();
        ComplianceScore = project.ComplianceScore;

        var tabValue = (tab ?? "summary").ToLowerInvariant();
        ActiveTabIndex = tabValue switch
        {
            "summary" => 0,
            "inspections" => 1,
            "violations" => 2,
            "audit" => 3,
            _ => 0,
        };
        ActiveTabId = ActiveTabIndex switch
        {
            1 => "tab-inspections",
            2 => "tab-violations",
            3 => "tab-audit",
            _ => "tab-summary",
        };

        Tabs =
        [
            new("tab-summary", "Summary", $"/projects/{id}/tabs/summary", "#project-detail-tab-content"),
            new("tab-inspections", "Inspections", $"/projects/{id}/tabs/inspections", "#project-detail-tab-content"),
            new("tab-violations", "Violations", $"/projects/{id}/tabs/violations", "#project-detail-tab-content"),
            new("tab-audit", "Audit", $"/projects/{id}/tabs/audit", "#project-detail-tab-content"),
        ];

        Summary = new SummaryVm(
            project.Code,
            project.Name,
            project.StartDate,
            project.TargetCompletionDate,
            project.Description,
            tradeNames,
            inspectorNames,
            new ActionButtonVm(
                "place-on-hold-btn",
                DomainPolicies.Can(User, "project.place_on_hold"),
                project.CanPlaceOnHold(),
                "Place on Hold",
                $"/projects/{id}/place-on-hold",
                "#project-detail",
                "Project is already on hold"
            ),
            new ActionButtonVm(
                "resume-btn",
                DomainPolicies.Can(User, "project.resume"),
                project.CanResume(),
                "Resume",
                $"/projects/{id}/resume",
                "#project-detail",
                "Project is not on hold"
            ),
            new ActionButtonVm(
                "close-btn",
                DomainPolicies.Can(User, "project.close"),
                project.CanClose(),
                "Close",
                $"/projects/{id}/close",
                "#project-detail",
                "Only active projects can be closed"
            )
        );

        var isHtmx = Request.Headers.TryGetValue("HX-Request", out var htmxVal) && string.Equals(htmxVal, "true", StringComparison.OrdinalIgnoreCase);
        IsTabResponse = tab is not null;

        if (tab is not null && tabValue is not ("summary" or "inspections" or "violations" or "audit"))
            return NotFound();

        if (!isHtmx && tab is not null)
            return Redirect($"/projects/{id}");

        return Page();
    }
}
