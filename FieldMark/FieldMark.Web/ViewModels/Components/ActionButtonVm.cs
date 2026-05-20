namespace FieldMark.Web.ViewModels.Components;

public sealed record ActionButtonVm(
    string Id,
    bool Permission,
    bool StateAllows,
    string Label,
    string HxPost,
    string HxTarget,
    string? DisabledReason = null
);
