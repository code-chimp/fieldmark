"""Snapshot tests for the AuditRow component."""

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot

BASE = {
    "action": "ProjectCreated",
    "actor_name": "Aisha Stone",
    "occurred_at": "2026-05-28T14:20:01Z",
    "absolute": "2026-05-28 14:20:01 UTC",
    "relative": "3 minutes ago",
    "before_after_json": "",
    "expanded": False,
}


@pytest.mark.parametrize(
    ("variant", "context"),
    [
        ("default", BASE),
        (
            "with-disclosure-collapsed",
            {
                **BASE,
                "before_after_json": '{"after":{"status":"ACTIVE"},"before":{"status":"DRAFT"}}',
            },
        ),
        (
            "with-disclosure-expanded",
            {
                **BASE,
                "before_after_json": '{"after":{"status":"ACTIVE"},"before":{"status":"DRAFT"}}',
                "expanded": True,
            },
        ),
        ("unknown-action", {**BASE, "action": "UnknownAction"}),
        ("empty-actor", {**BASE, "actor_name": ""}),
    ],
)
def test_audit_row_variant_matches_canonical(variant: str, context: dict[str, object]):
    assert_component_snapshot(
        "audit_row", "components/_audit_row.html", variant, context
    )


def test_audit_row_whitespace_only_actor_matches_empty_actor_canonical():
    assert_component_snapshot(
        "audit_row",
        "components/_audit_row.html",
        "empty-actor",
        {**BASE, "actor_name": "   "},
    )


def test_audit_row_unknown_action_fallback_class():
    html = render_to_string(
        "components/_audit_row.html",
        {**BASE, "action": "UnknownAction"},
    )
    assert "badge-unknown" in html


def test_audit_row_escapes_json_text():
    html = render_to_string(
        "components/_audit_row.html",
        {**BASE, "before_after_json": "<script>alert(1)</script>", "expanded": True},
    )
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html


def test_audit_row_template_does_not_use_safe_filter():
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / "_audit_row.html"
    )
    assert "|safe" not in path.read_text(encoding="utf-8")
