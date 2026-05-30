// POST /projects/ — project create handler.
// GET /projects/ → 405 (Story 2.9 will register the list here).
// See docs/reference/project-create-form-contract.md.
using System.Text.Json;
using System.Text.RegularExpressions;
using FieldMark.Data.Auditing;
using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using FieldMark.Domain.Entities;
using FieldMark.Domain.ValueObjects;
using FieldMark.Web.Authentication;
using FieldMark.Web.Authorization;
using FieldMark.Web.ViewModels.Projects;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;
using Npgsql;

namespace FieldMark.Web.Pages.Projects;

public sealed partial class IndexModel : PageModel
{
    private readonly FieldMarkDbContext _db;
    private readonly AuthDbContext _authDb;
    private readonly IAuditAppender _audit;
    private readonly IReferenceReader _reference;
    private readonly UserManager<IdentityUser<Guid>> _userManager;

    public IndexModel(
        FieldMarkDbContext db,
        AuthDbContext authDb,
        IAuditAppender audit,
        IReferenceReader reference,
        UserManager<IdentityUser<Guid>> userManager
    )
    {
        _db = db;
        _authDb = authDb;
        _audit = audit;
        _reference = reference;
        _userManager = userManager;
    }

    // GET /projects/ → 405 Method Not Allowed (Story 2.9 will register the list route here).
    [IgnoreAntiforgeryToken]
    public IActionResult OnGet()
    {
        Response.Headers.Allow = "POST";
        return StatusCode(405);
    }

    // POST /projects/ — canonical form submit. Antiforgery validated by Razor Pages framework.
    public async Task<IActionResult> OnPostAsync(CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "project.create"))
            return StatusCode(403, "You do not have permission to access this page.");

        // --- Collect raw form values (canonical snake_case names per contract doc) ---
        var rawCode = (Request.Form["code"].FirstOrDefault() ?? "").Trim();
        var rawName = (Request.Form["name"].FirstOrDefault() ?? "").Trim();
        var rawDescription = Request.Form["description"].FirstOrDefault()?.Trim();
        var rawStartDate = Request.Form["start_date"].FirstOrDefault() ?? "";
        var rawTargetDate = Request.Form["target_completion_date"].FirstOrDefault() ?? "";
        var rawTradeIds = Request.Form["trade_scope_ids"].ToList();
        var rawInspectorIds = Request.Form["inspector_ids"].ToList();

        var errors = new Dictionary<string, string>();

        // --- Validate code ---
        if (string.IsNullOrWhiteSpace(rawCode))
            errors["code"] = "Code is required.";
        else if (rawCode.Length > 32)
            errors["code"] = "Code must be 32 characters or fewer.";
        else if (!CodePattern().IsMatch(rawCode))
        {
            if (rawCode.StartsWith('-'))
                errors["code"] = "Code must start with a letter or digit.";
            else
                errors["code"] = "Code must contain only uppercase letters, digits, and hyphens.";
        }

        // --- Validate name ---
        if (string.IsNullOrWhiteSpace(rawName))
            errors["name"] = "Name is required.";
        else if (rawName.Length > 200)
            errors["name"] = "Name must be 200 characters or fewer.";

        // --- Validate description ---
        var description = string.IsNullOrWhiteSpace(rawDescription) ? null : rawDescription;
        if (description?.Length > 10000)
            errors["description"] = "Description must be 10,000 characters or fewer.";

        // --- Validate start_date ---
        DateOnly startDate = default;
        if (string.IsNullOrWhiteSpace(rawStartDate))
            errors["start_date"] = "Start date is required.";
        else if (!DateOnly.TryParseExact(rawStartDate, "yyyy-MM-dd", out startDate))
            errors["start_date"] = "Start date must be a valid date (YYYY-MM-DD).";

        // --- Validate target_completion_date ---
        DateOnly? targetDate = null;
        if (!string.IsNullOrWhiteSpace(rawTargetDate))
        {
            if (!DateOnly.TryParseExact(rawTargetDate, "yyyy-MM-dd", out var td))
                errors["target_completion_date"] = "Target completion date must be a valid date.";
            else if (!errors.ContainsKey("start_date") && td < startDate)
                errors["target_completion_date"] =
                    "Target completion date must be on or after the start date.";
            else
                targetDate = td;
        }

        // --- Validate trade_scope_ids ---
        var tradeScopeIds = new List<Guid>();
        if (rawTradeIds.Count == 0)
        {
            errors["trade_scope_ids"] = "At least one trade scope is required.";
        }
        else
        {
            var hasMalformedTrade = false;
            foreach (var raw in rawTradeIds)
            {
                if (Guid.TryParse(raw, out var g))
                    tradeScopeIds.Add(g);
                else
                    hasMalformedTrade = true;
            }
            if (hasMalformedTrade)
                errors["trade_scope_ids"] =
                    "One or more selected trade types are no longer available. Please reselect.";
        }

        // --- Validate inspector_ids ---
        var inspectorIds = new List<Guid>();
        {
            var hasMalformedInspector = false;
            foreach (var raw in rawInspectorIds)
            {
                if (Guid.TryParse(raw, out var g))
                    inspectorIds.Add(g);
                else
                    hasMalformedInspector = true;
            }
            if (hasMalformedInspector)
                errors["inspector_ids"] =
                    "One or more selected inspectors are no longer available. Please reselect.";
        }

        // Deduplicate — duplicate selections are redundant and would cause a
        // composite-PK 23505 on project_trade_scope / project_inspector.
        tradeScopeIds = tradeScopeIds.Distinct().ToList();
        inspectorIds = inspectorIds.Distinct().ToList();

        if (errors.Count > 0)
            return await Return422Async(
                rawCode,
                rawName,
                rawDescription ?? "",
                rawStartDate,
                rawTargetDate,
                tradeScopeIds,
                inspectorIds,
                errors,
                ct
            );

        // --- DB reference validation inside the transaction ---
        await using var tx = await _db.Database.BeginTransactionAsync(ct);
        try
        {
            var activeTradeIds = await _db
                .TradeTypes.Where(t => t.Active)
                .Select(t => t.Id)
                .ToHashSetAsync(ct);

            var invalidTrades = tradeScopeIds.Where(id => !activeTradeIds.Contains(id)).ToList();
            if (invalidTrades.Count > 0)
                errors["trade_scope_ids"] =
                    "One or more selected trade types are no longer available. Please reselect.";

            if (inspectorIds.Count > 0)
            {
                // Note: UserManager reads dotnet_auth via AuthDbContext (separate DbContext
                // from the domain FieldMarkDbContext transaction). Enlisting AuthDbContext in
                // the same physical transaction would require sharing a connection, which
                // conflicts with EF Core's connection-pooling model. The resulting TOCTOU
                // window (role change between read and INSERT) is the same accepted risk
                // described in Dev Notes §"Validate-then-write race" — microsecond window,
                // recoverable state, no additional mitigation introduced in this story.
                // Filter locked-out users for parity with Django's is_active=True check.
                var nowValidation = DateTimeOffset.UtcNow;
                var validInspectors = (await _userManager.GetUsersInRoleAsync(Role.Inspector.Name))
                    .Where(u => u.LockoutEnd == null || u.LockoutEnd < nowValidation)
                    .Select(u => u.Id)
                    .ToHashSet();
                var invalidInspectors = inspectorIds.Where(id => !validInspectors.Contains(id)).ToList();
                if (invalidInspectors.Count > 0)
                    errors["inspector_ids"] =
                        "One or more selected inspectors are no longer available. Please reselect.";
            }

            if (errors.Count > 0)
            {
                await tx.RollbackAsync(ct);
                return await Return422Async(
                    rawCode,
                    rawName,
                    rawDescription ?? "",
                    rawStartDate,
                    rawTargetDate,
                    tradeScopeIds,
                    inspectorIds,
                    errors,
                    ct
                );
            }

            // --- Call entity method ---
            var created = Project.Create(
                rawCode,
                rawName,
                description,
                startDate,
                targetDate,
                tradeScopeIds,
                inspectorIds
            );

            // --- Persist (FK order: project → joins → audit) ---
            _db.Projects.Add(created.Project);
            _db.ProjectTradeScopes.AddRange(created.Scopes);
            _db.ProjectInspectors.AddRange(created.Inspectors);

            var sortedTradeIds = tradeScopeIds.Order().Select(id => id.ToString()).ToList();
            var sortedInspIds = inspectorIds.Order().Select(id => id.ToString()).ToList();
            var afterState = JsonDocument.Parse(
                JsonSerializer.Serialize(
                    new
                    {
                        code = created.Project.Code,
                        compliance_score = 100,
                        description = (object?)description,
                        inspector_ids = sortedInspIds,
                        name = created.Project.Name,
                        start_date = startDate.ToString("yyyy-MM-dd"),
                        status = "Active",
                        target_completion_date = (object?)(targetDate.HasValue
                            ? targetDate.Value.ToString("yyyy-MM-dd")
                            : null),
                        trade_scope_ids = sortedTradeIds,
                    }
                )
            );

            _audit.Append(
                actorId: User.GetActorId(),
                action: AuditAction.ProjectCreated,
                entityType: "Project",
                entityId: created.Project.Id,
                projectId: created.Project.Id,
                beforeState: null,
                afterState: afterState,
                metadata: null
            );

            await _db.SaveChangesAsync(ct);
            await tx.CommitAsync(ct);

            var isHtmx =
                Request.Headers.TryGetValue("HX-Request", out var htmxVal)
                && htmxVal == "true";

            if (isHtmx)
            {
                Response.Headers.Append("HX-Redirect", $"/projects/{created.Project.Id}");
                return new OkResult();
            }

            // Non-HTMX: POST/Redirect/GET with 303 See Other (preserves GET on follow-up).
            Response.StatusCode = StatusCodes.Status303SeeOther;
            Response.Headers.Location = $"/projects/{created.Project.Id}";
            return new EmptyResult();
        }
        catch (DbUpdateException ex) when (IsCodeUniqueViolation(ex))
        {
            await tx.RollbackAsync(ct);
            errors["code"] = "A project with this code already exists.";
            return await Return422Async(
                rawCode,
                rawName,
                rawDescription ?? "",
                rawStartDate,
                rawTargetDate,
                tradeScopeIds,
                inspectorIds,
                errors,
                ct
            );
        }
        catch
        {
            await tx.RollbackAsync(ct);
            throw;
        }
    }

    private async Task<IActionResult> Return422Async(
        string code,
        string name,
        string description,
        string startDate,
        string targetDate,
        List<Guid> selectedTradeIds,
        List<Guid> selectedInspectorIds,
        Dictionary<string, string> errors,
        CancellationToken ct
    )
    {
        var tradeTypes = await _reference.ListTradeTypesAsync(ct);
        // Filter locked-out users for parity with Django's is_active=True check.
        var now = DateTimeOffset.UtcNow;
        var allInspectors422 = await _userManager.GetUsersInRoleAsync(Role.Inspector.Name);
        var inspectorUsers = allInspectors422
            .Where(u => u.LockoutEnd == null || u.LockoutEnd < now)
            .ToList();
        var displayNames = await GetDisplayNamesAsync(inspectorUsers.Select(u => u.Id));

        var vm = new ProjectCreateFormVm
        {
            Code = code,
            Name = name,
            Description = description,
            StartDate = startDate,
            TargetCompletionDate = targetDate,
            SelectedTradeTypeIds = selectedTradeIds,
            SelectedInspectorIds = selectedInspectorIds,
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
            FieldErrors = errors,
        };

        Response.StatusCode = StatusCodes.Status422UnprocessableEntity;
        return Partial("Projects/Shared/_ProjectCreateForm", vm);
    }

    private async Task<Dictionary<Guid, string>> GetDisplayNamesAsync(IEnumerable<Guid> userIds)
    {
        var ids = userIds.ToHashSet();
        // GroupBy then take first value to guard against duplicate display_name claims.
        var claims = await _authDb
            .UserClaims.Where(c => ids.Contains(c.UserId) && c.ClaimType == "display_name")
            .ToListAsync();
        return claims
            .GroupBy(c => c.UserId)
            .ToDictionary(g => g.Key, g => g.First().ClaimValue ?? "");
    }

    // Only catch the code UNIQUE violation — not other 23505s (e.g. composite PK
    // collisions on project_trade_scope under concurrent retry). The DDL's inline
    // UNIQUE on domain.project.code auto-generates the constraint name "project_code_key".
    private static bool IsCodeUniqueViolation(DbUpdateException ex) =>
        ex.InnerException is PostgresException pg
        && pg.SqlState == "23505"
        && pg.ConstraintName == "project_code_key";

    [GeneratedRegex(@"^[A-Z0-9][A-Z0-9-]*$")]
    private static partial Regex CodePattern();
}
