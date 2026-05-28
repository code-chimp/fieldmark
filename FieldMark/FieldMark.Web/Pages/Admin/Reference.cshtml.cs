using System.Text.Json;
using FieldMark.Data.Reference;
using FieldMark.Domain.Entities.Reference;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc.RazorPages;

namespace FieldMark.Web.Pages.Admin;

[Authorize(Roles = "ADMIN")]
public sealed class ReferenceModel(IReferenceReader references) : PageModel
{
    public IReadOnlyList<TradeType> TradeTypes { get; private set; } = [];
    public IReadOnlyList<ViolationCategory> ViolationCategories { get; private set; } = [];
    public IReadOnlyList<ComplianceRuleRow> ComplianceRules { get; private set; } = [];

    public async Task OnGetAsync(CancellationToken ct)
    {
        TradeTypes = await references.ListTradeTypesAsync(ct);
        ViolationCategories = await references.ListViolationCategoriesAsync(ct);
        ComplianceRules = (await references.ListComplianceRulesAsync(ct))
            .Select(r => new ComplianceRuleRow(
                r.Code,
                r.Name,
                r.Description,
                r.RuleKind,
                JsonSerializer.Serialize(r.Parameters.RootElement),
                r.Active
            ))
            .ToList();
    }

    public sealed record ComplianceRuleRow(
        string Code,
        string Name,
        string Description,
        string RuleKind,
        string ParametersJson,
        bool Active
    );
}
