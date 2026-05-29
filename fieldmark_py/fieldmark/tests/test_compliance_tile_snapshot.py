"""Snapshot tests for the ComplianceTile component."""

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot
from fieldmark.tests.normalize_html import normalise_component

BASE = {
    "score": 95,
    "label": "Compliance",
    "id": "compliance-tile",
}

PORTFOLIO_BASE = {
    "score": 91,
    "label": "Portfolio Compliance",
    "id": "compliance-tile-portfolio",
}


@pytest.mark.parametrize(
    ("variant", "context"),
    [
        ("healthy-project", BASE),
        ("watch-project", {**BASE, "score": 82}),
        ("concern-project", {**BASE, "score": 58}),
        ("critical-project", {**BASE, "score": 37}),
        ("healthy-portfolio", PORTFOLIO_BASE),
        ("critical-portfolio", {**PORTFOLIO_BASE, "score": 42}),
        ("no-data-project", {**BASE, "score": None}),
        ("boundary-90", {**BASE, "score": 90}),
        ("boundary-70", {**BASE, "score": 70}),
        ("boundary-50", {**BASE, "score": 50}),
        ("boundary-49", {**BASE, "score": 49}),
    ],
)
def test_compliance_tile_variant_matches_canonical(
    variant: str, context: dict[str, object]
) -> None:
    assert_component_snapshot(
        "compliance_tile", "components/_compliance_tile.html", variant, context
    )


def test_score_zero_renders_as_critical_not_no_data() -> None:
    html = normalise_component(
        render_to_string("components/_compliance_tile.html", {**BASE, "score": 0})
    )
    assert "text-danger" in html
    assert "Critical" in html
    assert "—" not in html


def test_portfolio_id_passed_through_verbatim() -> None:
    assert_component_snapshot(
        "compliance_tile",
        "components/_compliance_tile.html",
        "healthy-portfolio",
        PORTFOLIO_BASE,
    )


def test_xss_payload_in_label_is_escaped() -> None:
    html = render_to_string(
        "components/_compliance_tile.html",
        {**BASE, "label": "<script>alert(1)</script>"},
    )
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html
    assert "<script>" not in html  # generic check: no raw script tag regardless of payload


def test_whitespace_only_label_does_not_crash() -> None:
    html = render_to_string("components/_compliance_tile.html", {**BASE, "label": "   "})
    assert "<section" in html
    assert "compliance-tile__label" in html


def test_empty_label_does_not_crash() -> None:
    html = render_to_string("components/_compliance_tile.html", {**BASE, "label": ""})
    assert "<section" in html


def test_target_shape_attributes_present() -> None:
    html = render_to_string("components/_compliance_tile.html", BASE)
    assert 'id="compliance-tile"' in html
    assert 'role="status"' in html
    assert 'aria-live="polite"' in html
    assert 'aria-atomic="true"' in html
    assert 'class="compliance-tile"' in html


def test_no_htmx_producer_attributes_emitted() -> None:
    html = render_to_string("components/_compliance_tile.html", BASE)
    for token in ("hx-get", "hx-post", "hx-target", "hx-swap", "hx-trigger", "<script", "onload=", "data-htmx-"):
        assert token not in html, f"Found forbidden token: {token!r}"


def test_compliance_tile_template_does_not_use_safe_filter() -> None:
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / "_compliance_tile.html"
    )
    assert "|safe" not in path.read_text(encoding="utf-8")
