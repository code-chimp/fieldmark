using FieldMark.Data.Context;
using FieldMark.Domain.Entities.Reference;
using Microsoft.EntityFrameworkCore;

namespace FieldMark.Data.Reference;

public sealed class ReferenceReader(FieldMarkDbContext db) : IReferenceReader
{
    public async Task<IReadOnlyList<TradeType>> ListTradeTypesAsync(
        CancellationToken ct = default
    ) => await db.TradeTypes.AsNoTracking().OrderBy(t => t.Code).ToListAsync(ct);

    public async Task<IReadOnlyList<ViolationCategory>> ListViolationCategoriesAsync(
        CancellationToken ct = default
    ) => await db.ViolationCategories.AsNoTracking().OrderBy(v => v.Code).ToListAsync(ct);

    public async Task<IReadOnlyList<ComplianceRule>> ListComplianceRulesAsync(
        CancellationToken ct = default
    ) => await db.ComplianceRules.AsNoTracking().OrderBy(r => r.Code).ToListAsync(ct);
}
