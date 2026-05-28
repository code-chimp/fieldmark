"""Shared fixtures for component snapshot tests."""

from __future__ import annotations

from pathlib import Path

from django.template.loader import render_to_string

from fieldmark.tests.normalize_html import extract_variant, normalise_component


def canonical_component(component: str) -> str:
    path = (
        Path(__file__).resolve().parents[3]
        / "fieldmark_shared"
        / "components"
        / component
        / "canonical.html"
    )
    return path.read_text(encoding="utf-8")


def assert_component_snapshot(
    component: str, template: str, variant: str, context: dict[str, object]
) -> None:
    actual = normalise_component(render_to_string(template, context))
    canonical = extract_variant(canonical_component(component), variant)
    assert actual == canonical, (
        f"{component} {variant} does not match canonical snapshot.\n"
        f"Expected: {canonical!r}\nActual:   {actual!r}"
    )
