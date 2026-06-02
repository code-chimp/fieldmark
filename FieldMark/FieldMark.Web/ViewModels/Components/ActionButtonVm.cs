namespace FieldMark.Web.ViewModels.Components;

public sealed record ActionButtonVm(
    string Id,
    bool Permission,
    bool StateAllows,
    string Label,
    string? HxPost,
    string? HxGet,
    string HxTarget,
    string HxSwap = "outerHTML",
    string? DisabledReason = null
);
