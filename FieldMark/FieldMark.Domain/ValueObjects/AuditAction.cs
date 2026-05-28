namespace FieldMark.Domain.ValueObjects;

/// <summary>
/// Canonical audit-action enum. The persisted form in
/// <c>domain.audit_entry.action</c> is the symbol's name verbatim — PascalCase
/// present-tense past-form per <c>docs/reference/audit-actions.md</c>.
///
/// Source of truth: <c>docs/reference/audit-actions.md</c> +
/// <c>docs/reference/audit-actions.json</c> conformance fixture.
/// Adding or removing a member requires the Change Procedure documented there.
/// </summary>
public enum AuditAction
{
    ProjectCreated,
    ProjectPlacedOnHold,
    ProjectResumed,
    ProjectClosed,
    InspectionScheduled,
    InspectionStarted,
    InspectionCompleted,
    InspectionCancelled,
    ViolationOpened,
    ViolationAssigned,
    ViolationVoided,
    CorrectiveActionSubmitted,
    CorrectiveActionTakenForReview,
    CorrectiveActionApproved,
    CorrectiveActionRejected,
}

public static class AuditActionExtensions
{
    public static string AsString(this AuditAction action) => action.ToString();
}
