"""Snapshot tests for the EntityRail component."""

from pathlib import Path

import pytest
from django.template.loader import render_to_string

from fieldmark.tests.component_fixtures import assert_component_snapshot

TEMPLATE = "components/_entity_rail.html"

EMPTY_VIOLATION = {
    "id": "violation-detail",
    "entity_type_label": "Violation",
    "entity_loaded": False,
    "body_slot": None,
    "footer_slot": None,
}

LOADED_VIOLATION = {
    "id": "violation-detail",
    "entity_type_label": "Violation",
    "entity_loaded": True,
    "body_slot": "__BODY__",
    "footer_slot": "__FOOTER__",
}


@pytest.mark.parametrize(
    ("variant", "context"),
    [
        ("empty-violation", EMPTY_VIOLATION),
        (
            "empty-inspection",
            {**EMPTY_VIOLATION, "id": "inspection-detail", "entity_type_label": "Inspection"},
        ),
        (
            "empty-corrective-action",
            {
                **EMPTY_VIOLATION,
                "id": "corrective-action-detail",
                "entity_type_label": "Corrective Action",
            },
        ),
        ("loaded-shell-violation", LOADED_VIOLATION),
        (
            "loaded-shell-inspection",
            {**LOADED_VIOLATION, "id": "inspection-detail", "entity_type_label": "Inspection"},
        ),
        (
            "loaded-shell-corrective-action",
            {
                **LOADED_VIOLATION,
                "id": "corrective-action-detail",
                "entity_type_label": "Corrective Action",
            },
        ),
    ],
)
def test_entity_rail_variant_matches_canonical(variant: str, context: dict[str, object]) -> None:
    assert_component_snapshot("entity_rail", TEMPLATE, variant, context)


# AC4 — four-case slot / footer-omission coverage


def test_empty_state_renders_empty_card() -> None:
    html = render_to_string(TEMPLATE, EMPTY_VIOLATION)
    assert "entity-rail--empty" in html
    assert "Empty entity rail" in html
    assert "Select an entity to see its detail here." in html
    assert "entity-rail__body" not in html


def test_loaded_with_both_slots_renders_body_and_footer() -> None:
    ctx = {**LOADED_VIOLATION, "body_slot": "<p>body</p>", "footer_slot": "<button>Save</button>"}
    html = render_to_string(TEMPLATE, ctx)
    assert "entity-rail--loaded" in html
    assert "<p>body</p>" in html
    assert "<button>Save</button>" in html
    assert "entity-rail__footer" in html


def test_loaded_body_only_omits_footer_div() -> None:
    ctx = {**LOADED_VIOLATION, "body_slot": "<p>body</p>", "footer_slot": None}
    html = render_to_string(TEMPLATE, ctx)
    assert "entity-rail__body" in html
    assert "entity-rail__footer" not in html


def test_loaded_no_slots_renders_header_and_empty_body() -> None:
    ctx = {**LOADED_VIOLATION, "body_slot": None, "footer_slot": None}
    html = render_to_string(TEMPLATE, ctx)
    assert "entity-rail--loaded" in html
    assert "entity-rail__header" in html
    assert "entity-rail__body" in html
    assert "entity-rail__footer" not in html


# AC8 — XSS round-trip: entity_type_label is framework-escaped (non-slot prop)


def test_xss_payload_in_entity_type_label_is_escaped() -> None:
    ctx = {**EMPTY_VIOLATION, "entity_type_label": "<script>alert(1)</script>"}
    html = render_to_string(TEMPLATE, ctx)
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html
    assert "<script>" not in html


def test_xss_payload_in_loaded_label_is_escaped_in_span_and_aria() -> None:
    ctx = {**LOADED_VIOLATION, "entity_type_label": "<script>alert(1)</script>"}
    html = render_to_string(TEMPLATE, ctx)
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>" not in html


# AC8 §category 9 — empty/whitespace entity_type_label does not crash


@pytest.mark.parametrize("label", ["", "   "])
def test_whitespace_or_empty_label_does_not_crash(label: str) -> None:
    ctx = {**EMPTY_VIOLATION, "entity_type_label": label}
    html = render_to_string(TEMPLATE, ctx)
    assert "<aside" in html
    assert "entity-rail" in html


# AC3 — no HTMX producer attributes on dismiss button


def test_loaded_shell_dismiss_button_has_no_htmx_attributes() -> None:
    html = render_to_string(TEMPLATE, LOADED_VIOLATION)
    for token in ("hx-get", "hx-post", "hx-target", "hx-swap", "hx-trigger", "onclick="):
        assert token not in html, f"dismiss button must not emit {token!r}"


# AC9 — scoped grep guard: exactly two |safe in _entity_rail.html


def test_entity_rail_template_has_exactly_two_safe_filters() -> None:
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / "_entity_rail.html"
    )
    content = path.read_text(encoding="utf-8")
    count = content.count("|safe")
    assert count == 2, (
        f"exactly two |safe filters are permitted in _entity_rail.html "
        f"(one for body_slot, one for footer_slot); found {count}"
    )


# Verify other component wrappers still have zero |safe


@pytest.mark.parametrize(
    "filename",
    [
        "_status_badge.html",
        "_inline_alert.html",
        "_audit_row.html",
        "_dashboard_tile.html",
        "_compliance_tile.html",
        "_tab_strip.html",
    ],
)
def test_other_component_wrappers_do_not_use_safe_filter(filename: str) -> None:
    path = (
        Path(__file__).resolve().parents[2]
        / "templates"
        / "components"
        / filename
    )
    content = path.read_text(encoding="utf-8")
    assert "|safe" not in content, f"{filename} must not use |safe filter"
