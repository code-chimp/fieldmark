# Contract: docs/reference/component-canonical-examples.md
from __future__ import annotations

from typing import Any

from django import template

register = template.Library()

_BANDS: list[tuple[int, dict[str, Any]]] = [
    (90, {"value_class": "text-success", "threshold_word": "Healthy", "threshold_class": "text-success", "render_p": True}),
    (70, {"value_class": "text-warning", "threshold_word": "Watch", "threshold_class": "text-warning", "render_p": True}),
    (50, {"value_class": "text-warning-strong", "threshold_word": "Concern", "threshold_class": "text-warning-strong", "render_p": True}),
    (0,  {"value_class": "text-danger", "threshold_word": "Critical", "threshold_class": "text-danger", "render_p": True}),
]

_NO_DATA: dict[str, Any] = {
    "value_class": "text-neutral",
    "threshold_word": "",
    "threshold_class": "",
    "render_p": False,
}


@register.simple_tag
def compliance_band(score: int | None) -> dict[str, Any]:
    """Return the threshold-band dict (including display_value) for a compliance score.

    Non-int, bool, out-of-range, and None scores render as the no-data variant.
    Guards against TypeError on non-int inputs before any numeric comparison.
    """
    if score is None or not isinstance(score, int) or isinstance(score, bool):
        return {**_NO_DATA, "display_value": "—"}
    if score < 0 or score > 100:
        return {**_NO_DATA, "display_value": "—"}
    for threshold, band in _BANDS:
        if score >= threshold:
            return {**band, "display_value": str(score)}
    return {**_NO_DATA, "display_value": "—"}
