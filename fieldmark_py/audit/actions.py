"""Canonical audit-action constants for the Django stack (Story 2.2).

The persisted form in ``domain.audit_entry.action`` is the PascalCase string
value. Symbol names follow the Python idiom (``SCREAMING_SNAKE_CASE``); the
conformance test (``audit/tests/test_action_conformance.py``) proves the
choice values round-trip to the canonical fixture exactly.

Source of truth: ``docs/reference/audit-actions.md`` +
``docs/reference/audit-actions.json``. Adding or removing a member requires
the Change Procedure documented there.
"""

from __future__ import annotations

from django.db import models


class AuditAction(models.TextChoices):
    PROJECT_CREATED = "ProjectCreated", "ProjectCreated"
    PROJECT_PLACED_ON_HOLD = "ProjectPlacedOnHold", "ProjectPlacedOnHold"
    PROJECT_RESUMED = "ProjectResumed", "ProjectResumed"
    PROJECT_CLOSED = "ProjectClosed", "ProjectClosed"
    INSPECTION_SCHEDULED = "InspectionScheduled", "InspectionScheduled"
    INSPECTION_STARTED = "InspectionStarted", "InspectionStarted"
    INSPECTION_COMPLETED = "InspectionCompleted", "InspectionCompleted"
    INSPECTION_CANCELLED = "InspectionCancelled", "InspectionCancelled"
    VIOLATION_OPENED = "ViolationOpened", "ViolationOpened"
    VIOLATION_ASSIGNED = "ViolationAssigned", "ViolationAssigned"
    VIOLATION_VOIDED = "ViolationVoided", "ViolationVoided"
    CORRECTIVE_ACTION_SUBMITTED = (
        "CorrectiveActionSubmitted",
        "CorrectiveActionSubmitted",
    )
    CORRECTIVE_ACTION_TAKEN_FOR_REVIEW = (
        "CorrectiveActionTakenForReview",
        "CorrectiveActionTakenForReview",
    )
    CORRECTIVE_ACTION_APPROVED = "CorrectiveActionApproved", "CorrectiveActionApproved"
    CORRECTIVE_ACTION_REJECTED = "CorrectiveActionRejected", "CorrectiveActionRejected"
