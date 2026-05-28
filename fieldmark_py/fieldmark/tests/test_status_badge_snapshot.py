"""Snapshot tests for the StatusBadge component."""

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot
from fieldmark.tests.normalize_html import normalise_component

CASES: dict[str, dict[str, object]] = {
    "project-active": {"entity": "Project", "value": "Active", "severity": None},
    "project-on-hold": {"entity": "Project", "value": "OnHold", "severity": None},
    "project-closed": {"entity": "Project", "value": "Closed", "severity": None},
    "inspection-scheduled": {
        "entity": "Inspection",
        "value": "Scheduled",
        "severity": None,
    },
    "inspection-in-progress": {
        "entity": "Inspection",
        "value": "InProgress",
        "severity": None,
    },
    "inspection-completed-pass": {
        "entity": "Inspection",
        "value": "CompletedPass",
        "severity": None,
    },
    "inspection-completed-conditional": {
        "entity": "Inspection",
        "value": "CompletedConditional",
        "severity": None,
    },
    "inspection-completed-fail": {
        "entity": "Inspection",
        "value": "CompletedFail",
        "severity": None,
    },
    "inspection-cancelled": {
        "entity": "Inspection",
        "value": "Cancelled",
        "severity": None,
    },
    "violation-open-critical-high": {
        "entity": "Violation",
        "value": "Open",
        "severity": "Critical",
    },
    "violation-open-medium-low": {
        "entity": "Violation",
        "value": "Open",
        "severity": "Medium",
    },
    "violation-in-progress": {
        "entity": "Violation",
        "value": "InProgress",
        "severity": None,
    },
    "violation-resolved": {
        "entity": "Violation",
        "value": "Resolved",
        "severity": None,
    },
    "violation-voided": {"entity": "Violation", "value": "Voided", "severity": None},
    "corrective-action-submitted": {
        "entity": "CorrectiveAction",
        "value": "Submitted",
        "severity": None,
    },
    "corrective-action-under-review": {
        "entity": "CorrectiveAction",
        "value": "UnderReview",
        "severity": None,
    },
    "corrective-action-approved": {
        "entity": "CorrectiveAction",
        "value": "Approved",
        "severity": None,
    },
    "corrective-action-rejected": {
        "entity": "CorrectiveAction",
        "value": "Rejected",
        "severity": None,
    },
    "severity-critical": {"entity": "Severity", "value": "Critical", "severity": None},
    "severity-high": {"entity": "Severity", "value": "High", "severity": None},
    "severity-medium": {"entity": "Severity", "value": "Medium", "severity": None},
    "severity-low": {"entity": "Severity", "value": "Low", "severity": None},
    "unknown": {"entity": "Violation", "value": "Foobar", "severity": None},
}


@pytest.mark.parametrize("variant", CASES)
def test_status_badge_variant_matches_canonical(variant: str):
    assert_component_snapshot(
        "status_badge",
        "components/_status_badge.html",
        variant,
        CASES[variant],
    )


def test_status_badge_unknown_fallback_class():
    html = normalise_component(
        render_to_string(
            "components/_status_badge.html",
            {"entity": "Violation", "value": "Foobar", "severity": None},
        )
    )
    assert "badge-unknown" in html


def test_status_badge_template_does_not_use_safe_filter():
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / "_status_badge.html"
    )
    assert "|safe" not in path.read_text(encoding="utf-8")
