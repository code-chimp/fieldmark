"""Snapshot and accessibility tests for the ActionButton template component."""

from pathlib import Path

from bs4 import BeautifulSoup
from django.template.loader import render_to_string

from fieldmark.tests.normalize_html import extract_variant, normalise_component

_CANONICAL = (
    Path(__file__).resolve().parents[3]
    / "fieldmark_shared"
    / "components"
    / "action_button.example.html"
)

# Canonical fixture values matching action_button.example.html.
_FIXTURE = {
    "id": "ab-fixture-1",
    "label": "Approve Resolution",
    "hx_post": "/violations/00000000-0000-0000-0000-000000000001/corrective-actions/00000000-0000-0000-0000-000000000002/approve",
    "hx_target": "#violation-detail",
    "disabled_reason": "Awaiting review",
}


def _render(permission: bool, state_allows: bool) -> str:
    return render_to_string(
        "components/_action_button.html",
        {**_FIXTURE, "permission": permission, "state_allows": state_allows},
    )


def test_action_button_permission_false_renders_empty():
    html = normalise_component(_render(permission=False, state_allows=False))
    assert html == "", f"Expected empty output for absent variant, got: {html!r}"


def test_action_button_disabled_variant_matches_canonical_snapshot():
    actual = normalise_component(_render(permission=True, state_allows=False))
    canonical = extract_variant(_CANONICAL.read_text(encoding="utf-8"), "disabled")
    assert actual == canonical, (
        f"Django disabled variant does not match canonical snapshot.\n"
        f"Expected: {canonical!r}\nActual:   {actual!r}"
    )


def test_action_button_present_variant_matches_canonical_snapshot():
    actual = normalise_component(_render(permission=True, state_allows=True))
    canonical = extract_variant(_CANONICAL.read_text(encoding="utf-8"), "present")
    assert actual == canonical, (
        f"Django present variant does not match canonical snapshot.\n"
        f"Expected: {canonical!r}\nActual:   {actual!r}"
    )


def test_action_button_disabled_variant_has_screen_reader_reason():
    html = _render(permission=True, state_allows=False)
    soup = BeautifulSoup(html, "html.parser")

    button = soup.find("button")
    assert button is not None, "Disabled variant must render a <button>"
    assert button.has_attr("disabled"), "Button must carry the disabled attribute"
    assert button.get("aria-disabled") == "true"
    assert button.get("tabindex") == "0"
    assert button.get("data-tooltip") == "Awaiting review"

    described_by = button.get("aria-describedby")
    assert described_by == "ab-fixture-1-reason"

    sr_span = soup.find("span", id=described_by)
    assert sr_span is not None, "sr-only reason span must exist in the DOM"
    assert "sr-only" in sr_span.get("class", [])
    assert sr_span.get_text(strip=True) == "Awaiting review"
