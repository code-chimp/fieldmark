"""Story 2.2 AC6 — Django conformance gate.

The ``AuditAction.values`` set must match the canonical fixture at
``docs/reference/audit-actions.json`` exactly. Pure unit test — no DB.
"""

from __future__ import annotations

import json
from pathlib import Path

from audit.actions import AuditAction


def _locate_fixture() -> Path:
    """Walk up from this file until docs/reference/audit-actions.json is found."""

    for parent in Path(__file__).resolve().parents:
        candidate = parent / "docs" / "reference" / "audit-actions.json"
        if candidate.exists():
            return candidate
    raise FileNotFoundError(
        f"Could not locate docs/reference/audit-actions.json walking up from {__file__}"
    )


def test_audit_action_matches_canonical_fixture() -> None:
    fixture = _locate_fixture()
    canonical_list: list[str] = json.loads(fixture.read_text())["actions"]

    # Cardinality first — set equality alone masks duplicate fixture entries.
    duplicates = sorted({x for x in canonical_list if canonical_list.count(x) > 1})
    assert not duplicates, f"audit-actions.json contains duplicate entries: {duplicates}"

    canonical = set(canonical_list)
    native = {choice.value for choice in AuditAction}

    missing_from_native = canonical - native
    extras_in_native = native - canonical

    assert not missing_from_native, (
        f"canonical actions missing from Django AuditAction: {sorted(missing_from_native)}"
    )
    assert not extras_in_native, (
        f"Django AuditAction has extras not in fixture: {sorted(extras_in_native)}"
    )
