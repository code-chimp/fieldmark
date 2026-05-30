"""Snapshot tests for the TabStrip component."""

from __future__ import annotations

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot

TEMPLATE = "components/_tab_strip.html"

# Canonical four Project Detail tabs (no badges)
_PROJECT_TABS_NO_BADGE = [
    {
        "id": "tab-summary",
        "label": "Summary",
        "hx_get": "/projects/__ID__/summary",
        "hx_target": "#project-detail-tab-content",
        "badge_count": None,
    },
    {
        "id": "tab-inspections",
        "label": "Inspections",
        "hx_get": "/projects/__ID__/inspections",
        "hx_target": "#project-detail-tab-content",
        "badge_count": None,
    },
    {
        "id": "tab-violations",
        "label": "Violations",
        "hx_get": "/projects/__ID__/violations",
        "hx_target": "#project-detail-tab-content",
        "badge_count": None,
    },
    {
        "id": "tab-audit",
        "label": "Audit",
        "hx_get": "/projects/__ID__/audit",
        "hx_target": "#project-detail-tab-content",
        "badge_count": None,
    },
]

SUMMARY_ACTIVE = {
    "id": "project-detail-tabstrip",
    "aria_label": "Project Detail Tabs",
    "active_index": 0,
    "tabs": _PROJECT_TABS_NO_BADGE,
}

VIOLATIONS_ACTIVE = {
    "id": "project-detail-tabstrip",
    "aria_label": "Project Detail Tabs",
    "active_index": 2,
    "tabs": _PROJECT_TABS_NO_BADGE,
}


@pytest.mark.parametrize(
    ("variant", "context"),
    [
        ("project-detail-four-tabs-summary-active", SUMMARY_ACTIVE),
        ("project-detail-four-tabs-violations-active", VIOLATIONS_ACTIVE),
        (
            "project-detail-four-tabs-with-badges",
            {
                "id": "project-detail-tabstrip",
                "aria_label": "Project Detail Tabs",
                "active_index": 0,
                "tabs": [
                    {**_PROJECT_TABS_NO_BADGE[0]},
                    {**_PROJECT_TABS_NO_BADGE[1], "badge_count": 12},
                    {**_PROJECT_TABS_NO_BADGE[2], "badge_count": 3},
                    {**_PROJECT_TABS_NO_BADGE[3], "badge_count": 147},
                ],
            },
        ),
        (
            "two-tabs-minimal",
            {
                "id": "two-tabs-strip",
                "aria_label": "Open Closed Tabs",
                "active_index": 0,
                "tabs": [
                    {
                        "id": "tab-open",
                        "label": "Open",
                        "hx_get": "/__tab__/open",
                        "hx_target": "#__panel__",
                        "badge_count": None,
                    },
                    {
                        "id": "tab-closed",
                        "label": "Closed",
                        "hx_get": "/__tab__/closed",
                        "hx_target": "#__panel__",
                        "badge_count": None,
                    },
                ],
            },
        ),
        (
            "single-tab",
            {
                "id": "single-tab-strip",
                "aria_label": "Single Tab",
                "active_index": 0,
                "tabs": [
                    {
                        "id": "tab-only",
                        "label": "Only Tab",
                        "hx_get": "/__tab__/only",
                        "hx_target": "#__panel__",
                        "badge_count": None,
                    }
                ],
            },
        ),
        (
            "badge-zero",
            {
                "id": "project-detail-tabstrip",
                "aria_label": "Project Detail Tabs",
                "active_index": 0,
                "tabs": [
                    {**_PROJECT_TABS_NO_BADGE[0]},
                    {**_PROJECT_TABS_NO_BADGE[1], "badge_count": 12},
                    {**_PROJECT_TABS_NO_BADGE[2], "badge_count": 0},
                    {**_PROJECT_TABS_NO_BADGE[3], "badge_count": 147},
                ],
            },
        ),
        (
            "badge-large",
            {
                "id": "project-detail-tabstrip",
                "aria_label": "Project Detail Tabs",
                "active_index": 0,
                "tabs": [
                    {**_PROJECT_TABS_NO_BADGE[0]},
                    {**_PROJECT_TABS_NO_BADGE[1], "badge_count": 9999},
                    {**_PROJECT_TABS_NO_BADGE[2]},
                    {**_PROJECT_TABS_NO_BADGE[3]},
                ],
            },
        ),
    ],
)
def test_tab_strip_variant_matches_canonical(variant: str, context: dict[str, object]) -> None:
    assert_component_snapshot("tab_strip", TEMPLATE, variant, context)


def test_active_index_zero_first_tab_has_tabindex_0() -> None:
    html = render_to_string(TEMPLATE, SUMMARY_ACTIVE)
    assert html.count('tabindex="0"') == 1
    assert html.count('tabindex="-1"') == 3


def test_active_index_two_third_tab_aria_selected_true() -> None:
    html = render_to_string(TEMPLATE, VIOLATIONS_ACTIVE)
    assert html.count('aria-selected="true"') == 1
    assert html.count('aria-selected="false"') == 3


def test_active_index_last_last_tab_active() -> None:
    ctx = {**SUMMARY_ACTIVE, "active_index": 3}
    html = render_to_string(TEMPLATE, ctx)
    assert html.count('aria-selected="true"') == 1
    assert 'id="tab-audit"' in html


def test_badge_count_12_renders_badge() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "badge_count": 12}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert "tab-strip__badge" in html
    assert ">12<" in html
    assert 'aria-label="12 unread"' in html


def test_badge_count_0_renders_badge_with_zero() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "badge_count": 0}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert "tab-strip__badge" in html
    assert ">0<" in html
    assert 'aria-label="0 unread"' in html


def test_badge_count_none_no_badge_element() -> None:
    html = render_to_string(TEMPLATE, SUMMARY_ACTIVE)
    assert "tab-strip__badge" not in html


def test_badge_count_9999_no_truncation() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "badge_count": 9999}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert ">9999<" in html
    assert "99+" not in html


def test_badge_negative_renders_verbatim() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "badge_count": -1}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert ">-1<" in html


def test_all_buttons_have_type_button() -> None:
    html = render_to_string(TEMPLATE, SUMMARY_ACTIVE)
    assert html.count('type="button"') == 4


def test_xss_label_is_escaped() -> None:
    payload = "<script>alert(1)</script>"
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "label": payload}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>" not in html


def test_xss_aria_label_is_escaped() -> None:
    payload = "<script>alert(1)</script>"
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "aria_label": payload})
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>" not in html


def test_xss_hx_get_is_escaped() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "hx_get": "javascript:alert(1)"}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    # hx-get should contain the value but auto-escaped
    assert "javascript:alert(1)" in html
    assert "<script>" not in html


def test_empty_label_does_not_crash() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "label": ""}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert "tab-strip__label" in html


def test_whitespace_only_label_does_not_crash() -> None:
    tabs = [{**_PROJECT_TABS_NO_BADGE[0], "label": "   "}]
    html = render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "tabs": tabs})
    assert "tab-strip__label" in html


def test_missing_aria_label_raises() -> None:
    """required_prop filter must raise ValueError when aria_label is empty."""
    import pytest
    with pytest.raises(ValueError, match="aria_label is required"):
        render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "aria_label": ""})


def test_whitespace_aria_label_raises() -> None:
    """required_prop filter must raise ValueError when aria_label is whitespace-only."""
    import pytest
    with pytest.raises(ValueError, match="aria_label is required"):
        render_to_string(TEMPLATE, {**SUMMARY_ACTIVE, "aria_label": "   "})


def test_safe_filter_not_in_template() -> None:
    template_path = (
        Path(__file__).resolve().parents[2] / "templates" / "components" / "_tab_strip.html"
    )
    content = template_path.read_text(encoding="utf-8")
    assert "|safe" not in content, "_tab_strip.html must not use |safe filter"
