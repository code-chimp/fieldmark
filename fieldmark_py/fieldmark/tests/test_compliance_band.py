"""Unit tests for the compliance_band template tag (pure function boundary tests)."""

from __future__ import annotations

import pytest

from fieldmark.templatetags.compliance_tile_tags import compliance_band


@pytest.mark.parametrize(
    ("score", "expected_value_class", "expected_threshold_word", "expected_render_p"),
    [
        # no-data and out-of-range
        (None,  "text-neutral",       "",         False),
        (-1,    "text-neutral",       "",         False),
        (101,   "text-neutral",       "",         False),
        # healthy boundaries
        (100,   "text-success",       "Healthy",  True),
        (90,    "text-success",       "Healthy",  True),
        # watch boundaries
        (89,    "text-warning",       "Watch",    True),
        (70,    "text-warning",       "Watch",    True),
        # concern boundaries
        (69,    "text-warning-strong", "Concern", True),
        (50,    "text-warning-strong", "Concern", True),
        # critical boundaries
        (49,    "text-danger",        "Critical", True),
        (0,     "text-danger",        "Critical", True),
    ],
)
def test_compliance_band(
    score: int | None,
    expected_value_class: str,
    expected_threshold_word: str,
    expected_render_p: bool,
) -> None:
    band = compliance_band(score)
    assert band["value_class"] == expected_value_class
    assert band["threshold_word"] == expected_threshold_word
    assert band["render_p"] == expected_render_p
    # display_value is part of the contract: str(score) for valid scores, em-dash for no-data
    expected_display = "—" if not expected_render_p else str(score)
    assert band["display_value"] == expected_display


@pytest.mark.parametrize("non_int_score", ["95", 95.0, True, False, [], {}])
def test_compliance_band_non_int_score_renders_no_data(non_int_score: object) -> None:
    """Non-int inputs must not raise TypeError; they render as no-data."""
    band = compliance_band(non_int_score)  # type: ignore[arg-type]
    assert band["value_class"] == "text-neutral"
    assert band["render_p"] is False
    assert band["display_value"] == "—"
