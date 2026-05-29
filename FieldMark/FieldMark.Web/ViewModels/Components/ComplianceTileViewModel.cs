// Contract: docs/reference/component-canonical-examples.md
using System.Globalization;

namespace FieldMark.Web.ViewModels.Components;

public record ComplianceTileViewModel(int? Score, string Label, string Id)
{
    public string DisplayValue =>
        Score is null || Score < 0 || Score > 100
            ? "—"
            : Score.Value.ToString(CultureInfo.InvariantCulture);

    public (string ValueClass, string ThresholdWord, string ThresholdClass, bool RenderP) Band =>
        ResolveBand(Score);

    public static (
        string ValueClass,
        string ThresholdWord,
        string ThresholdClass,
        bool RenderP
    ) ResolveBand(int? score) =>
        score switch
        {
            null => ("text-neutral", "", "", false),
            var s when s < 0 || s > 100 => ("text-neutral", "", "", false),
            >= 90 => ("text-success", "Healthy", "text-success", true),
            >= 70 => ("text-warning", "Watch", "text-warning", true),
            >= 50 => ("text-warning-strong", "Concern", "text-warning-strong", true),
            _ => ("text-danger", "Critical", "text-danger", true),
        };
}
