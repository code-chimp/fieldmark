using System.Text.Json;
using System.Text.Json.Nodes;
using System.Text.RegularExpressions;
using System.Linq;
using FieldMark.Data.Auditing;
using FieldMark.Data.Context;
using FieldMark.Data.Reference;
using FieldMark.Domain.Entities;
using FieldMark.Domain.Exceptions;
using FieldMark.Domain.ValueObjects;
using FieldMark.Web.Authentication;
using FieldMark.Web.Authorization;
using FieldMark.Web.ViewModels.Components;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Web.Pages.Projects;

public abstract partial class ProjectDetailPageModelBase : PageModel
{
    private readonly FieldMarkDbContext _db;
    private readonly AuthDbContext _authDb;
    private readonly IReferenceReader _reference;
    private readonly IAuditAppender _audit;

    protected ProjectDetailPageModelBase(FieldMarkDbContext db, AuthDbContext authDb, IReferenceReader reference, IAuditAppender audit)
    {
        _db = db;
        _authDb = authDb;
        _reference = reference;
        _audit = audit;
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

    public string ProjectName { get; protected set; } = string.Empty;
    public string ProjectCode { get; protected set; } = string.Empty;
    public string ProjectStatus { get; protected set; } = string.Empty;
    public int ComplianceScore { get; protected set; }
    public SummaryVm? Summary { get; protected set; }
    public IReadOnlyList<TabSpec> Tabs { get; protected set; } = [];
    public int ActiveTabIndex { get; protected set; }
    public bool IsTabResponse { get; protected set; }
    public bool IsTransitionFormResponse { get; protected set; }
    public bool IsTransitionSuccessResponse { get; protected set; }
    public string ActiveTabId { get; protected set; } = "tab-summary";
    public string CurrentTab { get; protected set; } = "summary";
    public object? TransitionError { get; protected set; }
    public object? AuditRow { get; protected set; }
    public IReadOnlyList<object> AuditRows { get; protected set; } = [];
    public string? AuditLoadMorePath { get; protected set; }
    public bool SuppressAuditEmptyState { get; protected set; }
    public string TransitionActionPath { get; protected set; } = string.Empty;
    public string TransitionSubmitLabel { get; protected set; } = string.Empty;
    public string TransitionTitle { get; protected set; } = string.Empty;
    public bool TransitionReasonRequired { get; protected set; }
    public string TransitionReason { get; protected set; } = string.Empty;
    public string? TransitionReasonError { get; protected set; }

    protected async Task<IActionResult> HandleDetailGetAsync(Guid id, string? tab, CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "project.read"))
        {
            Response.StatusCode = 403;
            return Content("You do not have permission to access this page.");
        }

        if (!await LoadDetailAsync(id, tab, ct))
            return NotFound();

        var isHtmx = Request.Headers.TryGetValue("HX-Request", out var htmxVal)
            && string.Equals(htmxVal, "true", StringComparison.OrdinalIgnoreCase);

        if (tab is not null && ActiveTabIndex == -1)
            return NotFound();

        if (!isHtmx && tab is not null)
            return Redirect($"/projects/{id}");

        return Page();
    }

    protected async Task<IActionResult> HandleAuditLogGetAsync(Guid id, string? beforeOccurredAt, string? beforeId, CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "project.read"))
        {
            Response.StatusCode = 403;
            return Content("You do not have permission to access this page.");
        }

        if (!await _db.Projects.AsNoTracking().AnyAsync(p => p.Id == id, ct))
            return NotFound();

        DateTimeOffset? cursorTs = null;
        Guid? cursorId = null;
        if (!string.IsNullOrWhiteSpace(beforeOccurredAt) || !string.IsNullOrWhiteSpace(beforeId))
        {
            if (!DateTimeOffset.TryParse(beforeOccurredAt, out var parsedTs) || !Guid.TryParse(beforeId, out var parsedId))
                return BadRequest("Invalid cursor.");
            cursorTs = parsedTs;
            cursorId = parsedId;
        }

        await LoadAuditPageAsync(id, cursorTs, cursorId, ct);
        return Page();
    }

    protected async Task<IActionResult> PostTransitionAsync(Guid id, ProjectTransitionKind transition, CancellationToken ct)
    {
        var currentTab = NormalizeProjectTab(Request.Form["current_tab"].FirstOrDefault());
        var action = transition == ProjectTransitionKind.PlaceOnHold ? "project.place_on_hold" : "project.resume";
        if (!DomainPolicies.Can(User, action))
            return StatusCode(403, "You do not have permission to access this page.");

        var rawReason = (Request.Form["reason"].FirstOrDefault() ?? "").Trim();
        var reasonError = ValidateReason(rawReason, transition == ProjectTransitionKind.PlaceOnHold);
        if (reasonError is not null)
        {
            SetTransitionForm(id, transition, rawReason, reasonError, currentTab);
            Response.StatusCode = StatusCodes.Status422UnprocessableEntity;
            return Page();
        }

        await using var tx = await _db.Database.BeginTransactionAsync(ct);
        try
        {
            var project = await _db.Projects
                .FromSqlInterpolated($@"
                    SELECT id, code, name, description, status, start_date, target_completion_date,
                           actual_closed_at, compliance_score, created_at, updated_at
                    FROM domain.project
                    WHERE id = {id}
                    FOR UPDATE")
                .FirstOrDefaultAsync(ct);
            if (project is null)
                return NotFound();

            var beforeState = JsonDocument.Parse(JsonSerializer.Serialize(new { status = project.Status.ToString() }));

            if (transition == ProjectTransitionKind.PlaceOnHold)
                project.PlaceOnHold(rawReason);
            else
                project.Resume(string.IsNullOrWhiteSpace(rawReason) ? null : rawReason);

            var afterState = JsonDocument.Parse(JsonSerializer.Serialize(new { status = project.Status.ToString() }));
            var metadata = JsonDocument.Parse(JsonSerializer.Serialize(new { reason = rawReason }));
            var auditAction = transition == ProjectTransitionKind.PlaceOnHold
                ? AuditAction.ProjectPlacedOnHold
                : AuditAction.ProjectResumed;

            _audit.Append(
                actorId: User.GetActorId(),
                action: auditAction,
                entityType: "Project",
                entityId: id,
                projectId: id,
                beforeState: beforeState,
                afterState: afterState,
                metadata: metadata
            );

            await _db.SaveChangesAsync(ct);
            var latestAudit = await _db.AuditEntries.AsNoTracking()
                .Where(a => a.ProjectId == id)
                .OrderByDescending(a => a.OccurredAt)
                .ThenByDescending(a => a.Id)
                .FirstAsync(ct);
            await tx.CommitAsync(ct);

            if (!await LoadDetailAsync(id, currentTab, ct))
                return NotFound();
            var now = DateTimeOffset.UtcNow;
            var currentAuditRow = new
            {
                Action = auditAction.AsString(),
                ActorName = User.Identity?.Name ?? "",
                OccurredAt = now.ToString("O"),
                Absolute = now.ToString("O"),
                Relative = "just now",
                BeforeAfterJson = JsonSerializer.Serialize(new
                {
                    after = new { status = ProjectStatus },
                    before = JsonSerializer.Deserialize<object>(beforeState.RootElement.GetRawText()),
                }),
                Expanded = false,
            };
            if (currentTab == "audit")
            {
                await LoadAuditPageAsync(id, latestAudit.OccurredAt, latestAudit.Id, ct);
                SuppressAuditEmptyState = true;
                AuditRows = ((IEnumerable<object>)[currentAuditRow]).Concat(AuditRows).ToList();
                AuditRow = null;
            }
            else
            {
                AuditRow = currentAuditRow;
            }
            IsTransitionSuccessResponse = true;
            return Page();
        }
        catch (InvalidProjectTransitionException ex)
        {
            await tx.RollbackAsync(ct);
            if (!await LoadDetailAsync(id, currentTab, ct))
                return NotFound();
            TransitionError = new
            {
                Severity = "danger",
                Title = transition == ProjectTransitionKind.PlaceOnHold
                    ? "Couldn't place project on hold"
                    : "Couldn't resume project",
                Message = ex.Message,
                Meta = "",
            };
            Response.StatusCode = StatusCodes.Status409Conflict;
            return Page();
        }
        catch
        {
            await tx.RollbackAsync(ct);
            throw;
        }
    }

    protected async Task<IActionResult> OnGetTransitionAsync(Guid id, ProjectTransitionKind transition, CancellationToken ct)
    {
        var action = transition == ProjectTransitionKind.PlaceOnHold ? "project.place_on_hold" : "project.resume";
        if (!DomainPolicies.Can(User, action))
            return StatusCode(403, "You do not have permission to access this page.");

        var project = await _db.Projects.AsNoTracking().FirstOrDefaultAsync(p => p.Id == id, ct);
        if (project is null)
            return NotFound();

        SetTransitionForm(id, transition, "", null, NormalizeProjectTab(Request.Query["current_tab"].FirstOrDefault()));
        return Page();
    }

    protected async Task<bool> LoadDetailAsync(Guid id, string? tab, CancellationToken ct)
    {
        var project = await _db.Projects.AsNoTracking().FirstOrDefaultAsync(p => p.Id == id, ct);
        if (project is null)
            return false;

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
        var claimMap = userClaims.GroupBy(c => c.UserId).ToDictionary(g => g.Key, g => g.First().ClaimValue);
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
        CurrentTab = NormalizeProjectTab(tabValue);
        ActiveTabIndex = tabValue switch
        {
            "summary" => 0,
            "inspections" => 1,
            "violations" => 2,
            "audit" => 3,
            _ => -1,
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
                null,
                $"/projects/{id}/place-on-hold",
                "#project-action-form",
                "innerHTML",
                "Project is already on hold"
            ),
            new ActionButtonVm(
                "resume-btn",
                DomainPolicies.Can(User, "project.resume"),
                project.CanResume(),
                "Resume",
                null,
                $"/projects/{id}/resume",
                "#project-action-form",
                "innerHTML",
                "Project is not on hold"
            ),
            new ActionButtonVm(
                "close-btn",
                DomainPolicies.Can(User, "project.close"),
                project.CanClose(),
                "Close",
                $"/projects/{id}/close",
                null,
                "#project-detail",
                "outerHTML",
                "Only active projects can be closed"
            )
        );

        IsTabResponse = tab is not null;
        if (ActiveTabIndex == 3)
            await LoadAuditPageAsync(id, null, null, ct);
        return true;
    }

    private async Task LoadAuditPageAsync(Guid projectId, DateTimeOffset? beforeOccurredAt, Guid? beforeId, CancellationToken ct)
    {
        var query = _db.AuditEntries.AsNoTracking().Where(a => a.ProjectId == projectId);
        SuppressAuditEmptyState = false;
        if (beforeOccurredAt.HasValue && beforeId.HasValue)
        {
            query = query.Where(a =>
                a.OccurredAt < beforeOccurredAt.Value
                || (a.OccurredAt == beforeOccurredAt.Value && a.Id.CompareTo(beforeId.Value) < 0));
        }

        var rows = await query
            .OrderByDescending(a => a.OccurredAt)
            .ThenByDescending(a => a.Id)
            .Take(101)
            .ToListAsync(ct);

        var hasMore = rows.Count > 100;
        if (hasMore)
            rows = rows.Take(100).ToList();

        var actorIds = rows.Select(r => r.ActorId).Distinct().ToHashSet();
        var users = await _authDb.Users.AsNoTracking().Where(u => actorIds.Contains(u.Id)).ToListAsync(ct);
        var claims = await _authDb.UserClaims.AsNoTracking()
            .Where(c => actorIds.Contains(c.UserId) && c.ClaimType == "display_name")
            .ToListAsync(ct);
        var claimMap = claims.GroupBy(c => c.UserId).ToDictionary(g => g.Key, g => g.First().ClaimValue);
        var actorMap = users.ToDictionary(
            u => u.Id,
            u => claimMap.TryGetValue(u.Id, out var display) && !string.IsNullOrWhiteSpace(display)
                ? display
                : (u.UserName ?? "").Trim()
        );

        var now = DateTimeOffset.UtcNow;
        AuditRows = rows
            .Select(row =>
            {
                var absolute = row.OccurredAt.ToUniversalTime().ToString("O");
                return (object)new
                {
                    Action = row.Action,
                    ActorName = actorMap.TryGetValue(row.ActorId, out var actorName) ? actorName : "",
                    OccurredAt = absolute,
                    Absolute = absolute,
                    Relative = RelativeAuditTime(row.OccurredAt, now),
                    BeforeAfterJson = RenderAuditJson(row),
                    Expanded = false,
                };
            })
            .ToList();

        AuditLoadMorePath = null;
        if (hasMore && rows.Count > 0)
        {
            var last = rows[^1];
            AuditLoadMorePath = $"/projects/{projectId}/audit-log?before_occurred_at={Uri.EscapeDataString(last.OccurredAt.ToUniversalTime().ToString("O"))}&before_id={last.Id}";
        }
    }

    private static string RelativeAuditTime(DateTimeOffset occurredAt, DateTimeOffset now)
    {
        var seconds = Math.Max(0, (int)(now - occurredAt).TotalSeconds);
        if (seconds < 60) return "just now";
        var minutes = seconds / 60;
        if (minutes < 60) return minutes == 1 ? "1 minute ago" : $"{minutes} minutes ago";
        var hours = minutes / 60;
        if (hours < 24) return hours == 1 ? "1 hour ago" : $"{hours} hours ago";
        var days = hours / 24;
        if (days < 30) return days == 1 ? "1 day ago" : $"{days} days ago";
        var months = days / 30;
        if (months < 12) return months == 1 ? "1 month ago" : $"{months} months ago";
        var years = months / 12;
        return years == 1 ? "1 year ago" : $"{years} years ago";
    }

    private static string? RenderAuditJson(AuditEntry row)
    {
        var payload = new Dictionary<string, JsonNode?>();
        if (row.AfterState is not null) payload["after"] = SortJsonNode(JsonNode.Parse(row.AfterState.RootElement.GetRawText()));
        if (row.BeforeState is not null) payload["before"] = SortJsonNode(JsonNode.Parse(row.BeforeState.RootElement.GetRawText()));
        if (row.Metadata is not null) payload["metadata"] = SortJsonNode(JsonNode.Parse(row.Metadata.RootElement.GetRawText()));
        if (payload.Count == 0) return null;
        return JsonSerializer.Serialize(payload);
    }

    private static JsonNode? SortJsonNode(JsonNode? node)
    {
        if (node is JsonObject obj)
        {
            var sorted = new JsonObject();
            foreach (var kvp in obj.OrderBy(k => k.Key, StringComparer.Ordinal))
                sorted[kvp.Key] = SortJsonNode(kvp.Value);
            return sorted;
        }
        if (node is JsonArray arr)
        {
            var sorted = new JsonArray();
            foreach (var child in arr)
                sorted.Add(SortJsonNode(child));
            return sorted;
        }
        return node?.DeepClone();
    }

    protected void SetTransitionForm(Guid id, ProjectTransitionKind transition, string reason, string? error, string currentTab)
    {
        IsTransitionFormResponse = true;
        CurrentTab = currentTab;
        TransitionActionPath = transition == ProjectTransitionKind.PlaceOnHold ? $"/projects/{id}/place-on-hold" : $"/projects/{id}/resume";
        TransitionSubmitLabel = transition == ProjectTransitionKind.PlaceOnHold ? "Place on hold" : "Resume";
        TransitionTitle = transition == ProjectTransitionKind.PlaceOnHold ? "Place project on hold" : "Resume project";
        TransitionReasonRequired = transition == ProjectTransitionKind.PlaceOnHold;
        TransitionReason = reason;
        TransitionReasonError = error;
    }

    private static string? ValidateReason(string reason, bool required)
    {
        if (required && string.IsNullOrWhiteSpace(reason))
            return "Reason is required.";
        if (reason.EnumerateRunes().Count() > 500)
            return "Reason must be 500 characters or fewer.";
        if (ControlCharPattern().IsMatch(reason))
            return "Reason contains invalid control characters.";
        return null;
    }

    [GeneratedRegex("[\\x00-\\x1F\\x7F]")]
    private static partial Regex ControlCharPattern();

    private static string NormalizeProjectTab(string? tab) =>
        (tab ?? "").Trim().ToLowerInvariant() switch
        {
            "inspections" => "inspections",
            "violations" => "violations",
            "audit" => "audit",
            _ => "summary",
        };

    protected enum ProjectTransitionKind
    {
        PlaceOnHold,
        Resume,
    }
}
