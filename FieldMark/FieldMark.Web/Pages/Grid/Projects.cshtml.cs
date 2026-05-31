// POST /grid/projects — AG Grid SSRM data endpoint.
// Read-only query — no transaction, no audit entry, no state change.
// CSRF: [IgnoreAntiforgeryToken] — read-only; AG Grid datasource does not
// send antiforgery tokens. Rationale documented in contract doc.
// See docs/reference/ag-grid-ssrm-contract.md
using System.Text.Json;
using FieldMark.Data.Context;
using FieldMark.Domain.Entities;
using FieldMark.Web.Authorization;
using FieldMark.Web.Grid;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.RazorPages;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Web.Pages.Grid;

[IgnoreAntiforgeryToken]
public sealed class ProjectsGridModel : PageModel
{
    private static readonly JsonSerializerOptions _camelCase = new()
    {
        PropertyNamingPolicy = JsonNamingPolicy.CamelCase,
    };

    private readonly FieldMarkDbContext _db;

    public ProjectsGridModel(FieldMarkDbContext db) => _db = db;

    // GET → 405 Method Not Allowed (endpoint is POST-only).
    public IActionResult OnGet()
    {
        Response.Headers.Allow = "POST";
        return StatusCode(405);
    }

    public async Task<IActionResult> OnPostAsync(CancellationToken ct)
    {
        if (!DomainPolicies.Can(User, "project.read"))
            return new JsonResult(new { error = "forbidden" }) { StatusCode = 403 };

        SsrmRequest req;
        try
        {
            var body = await new StreamReader(Request.Body).ReadToEndAsync(ct);
            req = JsonSerializer.Deserialize<SsrmRequest>(body, _camelCase)
                ?? throw new JsonException("null body");
        }
        catch (JsonException)
        {
            return new JsonResult(new { error = "invalid request body" }) { StatusCode = 400 };
        }

        SsrmParsed parsed;
        try
        {
            parsed = SsrmParser.Parse(req);
        }
        catch (SsrmValidationException ex)
        {
            return new JsonResult(new { error = ex.Message }) { StatusCode = 400 };
        }

        if (parsed.MatchNothing)
            return new JsonResult(new { rows = Array.Empty<object>(), lastRow = 0 });

        var baseQuery = BuildFilteredQuery(parsed);
        var total = await baseQuery.CountAsync(ct);

        // Manual projection (NFR6) — exactly seven columns, no AutoMapper.
        var sortedQuery = ApplySort(baseQuery, parsed)
            .Skip(parsed.Offset)
            .Take(parsed.Limit);

        var rows = await sortedQuery
            .Select(p => new
            {
                id = p.Id,
                code = p.Code,
                name = p.Name,
                status = p.Status.ToString(),
                compliance_score = p.ComplianceScore,
                start_date = p.StartDate.ToString("yyyy-MM-dd"),
                target_completion_date = p.TargetCompletionDate.HasValue
                    ? p.TargetCompletionDate.Value.ToString("yyyy-MM-dd")
                    : null,
            })
            .ToListAsync(ct);

        return new JsonResult(new { rows, lastRow = total });
    }

    private IQueryable<Project> BuildFilteredQuery(SsrmParsed parsed)
    {
        var q = _db.Projects.AsNoTracking();

        foreach (var (colId, f) in parsed.FiltersByColId)
        {
            q = (colId, f.FilterType) switch
            {
                ("status", "set") => q.Where(p => f.Values!.Contains(p.Status.ToString())),
                ("code", "text") => ApplyCodeFilter(q, f),
                ("name", "text") => ApplyNameFilter(q, f),
                ("compliance_score", "number") => ApplyNumberFilter(q, f),
                ("start_date", "date") => ApplyStartDateFilter(q, f),
                ("target_completion_date", "date") => ApplyTargetDateFilter(q, f),
                _ => q,
            };
        }

        return q;
    }

    private static IQueryable<Project> ApplyCodeFilter(IQueryable<Project> q, SsrmFilterEntry f)
    {
        var val = f.FilterAsString ?? "";
        return f.Type switch
        {
            "blank" => q.Where(p => p.Code == ""),
            "notBlank" => q.Where(p => p.Code != ""),
            "equals" => q.Where(p => p.Code == val),
            "notEqual" => q.Where(p => p.Code != val),
            "contains" => q.Where(p => EF.Functions.ILike(p.Code, $"%{val}%")),
            "notContains" => q.Where(p => !EF.Functions.ILike(p.Code, $"%{val}%")),
            "startsWith" => q.Where(p => EF.Functions.ILike(p.Code, $"{val}%")),
            "endsWith" => q.Where(p => EF.Functions.ILike(p.Code, $"%{val}")),
            _ => q,
        };
    }

    private static IQueryable<Project> ApplyNameFilter(IQueryable<Project> q, SsrmFilterEntry f)
    {
        var val = f.FilterAsString ?? "";
        return f.Type switch
        {
            "blank" => q.Where(p => p.Name == ""),
            "notBlank" => q.Where(p => p.Name != ""),
            "equals" => q.Where(p => p.Name == val),
            "notEqual" => q.Where(p => p.Name != val),
            "contains" => q.Where(p => EF.Functions.ILike(p.Name, $"%{val}%")),
            "notContains" => q.Where(p => !EF.Functions.ILike(p.Name, $"%{val}%")),
            "startsWith" => q.Where(p => EF.Functions.ILike(p.Name, $"{val}%")),
            "endsWith" => q.Where(p => EF.Functions.ILike(p.Name, $"%{val}")),
            _ => q,
        };
    }

    private static IQueryable<Project> ApplyNumberFilter(
        IQueryable<Project> q,
        SsrmFilterEntry f
    )
    {
        // Parser guarantees Filter/FilterTo are present for value-requiring operators.
        var val = (int)(f.FilterAsDouble ?? 0);
        var valTo = (int)(f.FilterToAsDouble ?? 0);
        return f.Type switch
        {
            "blank" => q.Where(_ => false), // compliance_score NOT NULL
            "notBlank" => q,
            "equals" => q.Where(p => p.ComplianceScore == val),
            "notEqual" => q.Where(p => p.ComplianceScore != val),
            "greaterThan" => q.Where(p => p.ComplianceScore > val),
            "greaterThanOrEqual" => q.Where(p => p.ComplianceScore >= val),
            "lessThan" => q.Where(p => p.ComplianceScore < val),
            "lessThanOrEqual" => q.Where(p => p.ComplianceScore <= val),
            "inRange" => q.Where(p => p.ComplianceScore >= val && p.ComplianceScore <= valTo),
            _ => q,
        };
    }

    private static IQueryable<Project> ApplyStartDateFilter(IQueryable<Project> q, SsrmFilterEntry f)
    {
        // Parser has already validated dateFrom/dateTo are valid YYYY-MM-DD strings for
        // operators that need them. ParseExact is safe here.
        DateOnly.TryParseExact(f.DateFrom ?? "", "yyyy-MM-dd", out var from);
        DateOnly.TryParseExact(f.DateTo ?? "", "yyyy-MM-dd", out var to);
        return f.Type switch
        {
            "blank" => q.Where(_ => false), // start_date NOT NULL
            "notBlank" => q,
            "equals" => q.Where(p => p.StartDate == from),
            "notEqual" => q.Where(p => p.StartDate != from),
            "greaterThan" => q.Where(p => p.StartDate > from),
            "lessThan" => q.Where(p => p.StartDate < from),
            "inRange" => q.Where(p => p.StartDate >= from && p.StartDate <= to),
            _ => q,
        };
    }

    private static IQueryable<Project> ApplyTargetDateFilter(
        IQueryable<Project> q,
        SsrmFilterEntry f
    )
    {
        // Parser has already validated dateFrom/dateTo for operators that need them.
        DateOnly.TryParseExact(f.DateFrom ?? "", "yyyy-MM-dd", out var from);
        DateOnly.TryParseExact(f.DateTo ?? "", "yyyy-MM-dd", out var to);
        return f.Type switch
        {
            "blank" => q.Where(p => p.TargetCompletionDate == null),
            "notBlank" => q.Where(p => p.TargetCompletionDate != null),
            "equals" => q.Where(p => p.TargetCompletionDate == from),
            "notEqual" => q.Where(p => p.TargetCompletionDate != from),
            "greaterThan" => q.Where(p => p.TargetCompletionDate > from),
            "lessThan" => q.Where(p => p.TargetCompletionDate < from),
            "inRange" => q.Where(p =>
                p.TargetCompletionDate >= from && p.TargetCompletionDate <= to),
            _ => q,
        };
    }

    private static IQueryable<Project> ApplySort(IQueryable<Project> q, SsrmParsed parsed)
    {
        IOrderedQueryable<Project>? ordered = null;
        bool first = true;

        foreach (var (colId, desc) in parsed.SortEntries)
        {
            if (first)
            {
                ordered = (colId, desc) switch
                {
                    ("code", false) => q.OrderBy(p => p.Code),
                    ("code", true) => q.OrderByDescending(p => p.Code),
                    ("name", false) => q.OrderBy(p => p.Name),
                    ("name", true) => q.OrderByDescending(p => p.Name),
                    ("status", false) => q.OrderBy(p => p.Status),
                    ("status", true) => q.OrderByDescending(p => p.Status),
                    ("compliance_score", false) => q.OrderBy(p => p.ComplianceScore),
                    ("compliance_score", true) => q.OrderByDescending(p => p.ComplianceScore),
                    ("start_date", false) => q.OrderBy(p => p.StartDate),
                    ("start_date", true) => q.OrderByDescending(p => p.StartDate),
                    ("target_completion_date", false) => q.OrderBy(p => p.TargetCompletionDate),
                    ("target_completion_date", true) => q.OrderByDescending(p => p.TargetCompletionDate),
                    ("id", false) => q.OrderBy(p => p.Id),
                    ("id", true) => q.OrderByDescending(p => p.Id),
                    _ => q.OrderBy(p => p.Code),
                };
                first = false;
            }
            else
            {
                ordered = (colId, desc) switch
                {
                    ("code", false) => ordered!.ThenBy(p => p.Code),
                    ("code", true) => ordered!.ThenByDescending(p => p.Code),
                    ("name", false) => ordered!.ThenBy(p => p.Name),
                    ("name", true) => ordered!.ThenByDescending(p => p.Name),
                    ("status", false) => ordered!.ThenBy(p => p.Status),
                    ("status", true) => ordered!.ThenByDescending(p => p.Status),
                    ("compliance_score", false) => ordered!.ThenBy(p => p.ComplianceScore),
                    ("compliance_score", true) => ordered!.ThenByDescending(p => p.ComplianceScore),
                    ("start_date", false) => ordered!.ThenBy(p => p.StartDate),
                    ("start_date", true) => ordered!.ThenByDescending(p => p.StartDate),
                    ("target_completion_date", false) => ordered!.ThenBy(p => p.TargetCompletionDate),
                    ("target_completion_date", true) => ordered!.ThenByDescending(p => p.TargetCompletionDate),
                    ("id", false) => ordered!.ThenBy(p => p.Id),
                    ("id", true) => ordered!.ThenByDescending(p => p.Id),
                    _ => ordered!.ThenBy(p => p.Id),
                };
            }
        }

        return ordered ?? q.OrderBy(p => p.Code).ThenBy(p => p.Id);
    }
}
