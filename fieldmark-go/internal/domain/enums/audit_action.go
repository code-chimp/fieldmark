// Package enums — audit_action.go.
//
// Canonical audit-action constants for the Go stack. The persisted form in
// domain.audit_entry.action is the string value verbatim — PascalCase
// present-tense past-form per docs/reference/audit-actions.md.
//
// Source of truth: docs/reference/audit-actions.md +
// docs/reference/audit-actions.json. Adding or removing a member requires the
// Change Procedure documented there. AllAuditActions below is the iteration
// target for audit_action_conformance_test.go.
package enums

type AuditAction string

const (
	AuditActionProjectCreated                 AuditAction = "ProjectCreated"
	AuditActionProjectPlacedOnHold            AuditAction = "ProjectPlacedOnHold"
	AuditActionProjectResumed                 AuditAction = "ProjectResumed"
	AuditActionProjectClosed                  AuditAction = "ProjectClosed"
	AuditActionInspectionScheduled            AuditAction = "InspectionScheduled"
	AuditActionInspectionStarted              AuditAction = "InspectionStarted"
	AuditActionInspectionCompleted            AuditAction = "InspectionCompleted"
	AuditActionInspectionCancelled            AuditAction = "InspectionCancelled"
	AuditActionViolationOpened                AuditAction = "ViolationOpened"
	AuditActionViolationAssigned              AuditAction = "ViolationAssigned"
	AuditActionViolationVoided                AuditAction = "ViolationVoided"
	AuditActionCorrectiveActionSubmitted      AuditAction = "CorrectiveActionSubmitted"
	AuditActionCorrectiveActionTakenForReview AuditAction = "CorrectiveActionTakenForReview"
	AuditActionCorrectiveActionApproved       AuditAction = "CorrectiveActionApproved"
	AuditActionCorrectiveActionRejected       AuditAction = "CorrectiveActionRejected"
)

// AllAuditActions is the exhaustive slice of canonical action values. The
// order matches the Change Procedure list in docs/reference/audit-actions.md.
var AllAuditActions = []AuditAction{
	AuditActionProjectCreated,
	AuditActionProjectPlacedOnHold,
	AuditActionProjectResumed,
	AuditActionProjectClosed,
	AuditActionInspectionScheduled,
	AuditActionInspectionStarted,
	AuditActionInspectionCompleted,
	AuditActionInspectionCancelled,
	AuditActionViolationOpened,
	AuditActionViolationAssigned,
	AuditActionViolationVoided,
	AuditActionCorrectiveActionSubmitted,
	AuditActionCorrectiveActionTakenForReview,
	AuditActionCorrectiveActionApproved,
	AuditActionCorrectiveActionRejected,
}
