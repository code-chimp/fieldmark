using FieldMark.Domain.Entities.Reference;

namespace FieldMark.Web.ViewModels.Projects;

/// <summary>
/// View model for the project-create form partial and 422 re-render.
/// Used by Create.cshtml (GET /projects/new) and Index.cshtml.cs (POST /projects/ 422).
/// See docs/reference/project-create-form-contract.md.
/// </summary>
public sealed class ProjectCreateFormVm
{
    // Current / echo-back values
    public string Code { get; init; } = "";
    public string Name { get; init; } = "";
    public string Description { get; init; } = "";
    public string StartDate { get; init; } = "";
    public string TargetCompletionDate { get; init; } = "";
    public IReadOnlyList<Guid> SelectedTradeTypeIds { get; init; } = [];
    public IReadOnlyList<Guid> SelectedInspectorIds { get; init; } = [];

    // Reference data for option lists
    public IReadOnlyList<TradeType> AvailableTradeTypes { get; init; } = [];
    public IReadOnlyList<InspectorOption> AvailableInspectors { get; init; } = [];

    // Validation
    public IReadOnlyDictionary<string, string> FieldErrors { get; init; } =
        new Dictionary<string, string>();

    public int ErrorCount => FieldErrors.Count;
    public bool HasErrors => ErrorCount > 0;
}

/// <summary>Inspector user summary for the inspector multi-select option list.</summary>
public sealed record InspectorOption(Guid Id, string Label);
