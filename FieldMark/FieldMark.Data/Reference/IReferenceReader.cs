using FieldMark.Domain.Entities.Reference;

namespace FieldMark.Data.Reference;

public interface IReferenceReader
{
    Task<IReadOnlyList<TradeType>> ListTradeTypesAsync(CancellationToken ct = default);
    Task<IReadOnlyList<ViolationCategory>> ListViolationCategoriesAsync(
        CancellationToken ct = default
    );
    Task<IReadOnlyList<ComplianceRule>> ListComplianceRulesAsync(CancellationToken ct = default);
}
