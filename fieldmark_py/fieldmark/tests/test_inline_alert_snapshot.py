"""Snapshot tests for the InlineAlert component."""

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot

BASE = {
    "title": "Action blocked",
    "message": "Resolve open violations before closing.",
    "meta": "Project PM-104",
}


@pytest.mark.parametrize("variant", ["danger", "warning", "info", "success", "unknown"])
def test_inline_alert_variant_matches_canonical(variant: str):
    severity = "notice" if variant == "unknown" else variant
    assert_component_snapshot(
        "inline_alert",
        "components/_inline_alert.html",
        variant,
        {**BASE, "severity": severity},
    )


def test_inline_alert_escapes_user_strings():
    html = render_to_string(
        "components/_inline_alert.html",
        {
            "severity": "danger",
            "title": "<script>alert(1)</script>",
            "message": "<script>alert(1)</script>",
            "meta": "<script>alert(1)</script>",
        },
    )
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html


def test_inline_alert_unknown_fallback_class():
    html = render_to_string(
        "components/_inline_alert.html",
        {**BASE, "severity": "notice"},
    )
    assert "alert-unknown" in html


def test_inline_alert_template_does_not_use_safe_filter():
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / "_inline_alert.html"
    )
    assert "|safe" not in path.read_text(encoding="utf-8")
