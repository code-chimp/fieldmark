"""Snapshot tests for the DashboardTile component."""

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot
from fieldmark.tests.normalize_html import normalise_component

BASE = {
    "tile_id": "open-violations-tile",
    "label": "Open Violations",
    "value": "12",
    "secondary": "",
    "value_color": "",
    "role_status": False,
}


@pytest.mark.parametrize(
    ("variant", "context"),
    [
        ("populated", BASE),
        ("zero-value", {**BASE, "value": "0"}),
        ("populated-with-secondary", {**BASE, "secondary": "3 critical"}),
        ("populated-with-color", {**BASE, "value_color": "danger"}),
        ("empty", {**BASE, "value": ""}),
        ("status-region", {**BASE, "role_status": True}),
    ],
)
def test_dashboard_tile_variant_matches_canonical(
    variant: str, context: dict[str, object]
):
    assert_component_snapshot(
        "dashboard_tile", "components/_dashboard_tile.html", variant, context
    )


@pytest.mark.parametrize("value", ["0", 0])
def test_dashboard_tile_zero_value_renders_zero(value: object):
    html = normalise_component(
        render_to_string("components/_dashboard_tile.html", {**BASE, "value": value})
    )

    assert ">0</p>" in html
    assert ">—</p>" not in html


def test_dashboard_tile_whitespace_only_value_matches_empty_canonical():
    assert_component_snapshot(
        "dashboard_tile",
        "components/_dashboard_tile.html",
        "empty",
        {**BASE, "value": "   "},
    )


def test_dashboard_tile_template_does_not_use_safe_filter():
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / "_dashboard_tile.html"
    )
    assert "|safe" not in path.read_text(encoding="utf-8")
