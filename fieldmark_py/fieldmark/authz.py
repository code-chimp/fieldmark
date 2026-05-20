"""Authorization decision primitive (FR5).

`can(user, action, entity_id=None) -> bool` is the single Django-side call
site every view uses to decide whether an action is permitted. Epic 1:
role-only checks (entity-scope rules deferred to Epic 2+).
"""

from __future__ import annotations

import uuid
from typing import Any

from fieldmark.roles import Role

# Action → frozenset of permitted role names. Epic 1 ships empty;
# subsequent stories register their actions via register_action().
_ACTION_ROLE_MAP: dict[str, frozenset[str]] = {}


def register_action(action: str, *roles: Role) -> None:
    """Register an action → permitted-roles mapping. Call at app-load
    time (typically a module-level statement in the same package as the
    action's handler — Django signals are banned, so AppConfig.ready()
    is off-limits).
    """
    existing = _ACTION_ROLE_MAP.get(action, frozenset())
    _ACTION_ROLE_MAP[action] = existing | frozenset(r.value for r in roles)


def can(user: Any, action: str, entity_id: uuid.UUID | None = None) -> bool:
    if not getattr(user, "is_authenticated", False):
        return False
    permitted = _ACTION_ROLE_MAP.get(action)
    if permitted is None:
        return False
    user_roles = set(user.groups.values_list("name", flat=True))
    if not (user_roles & permitted):
        return False
    return _evaluate_entity_scope(action, entity_id)


def _evaluate_entity_scope(action: str, entity_id: uuid.UUID | None) -> bool:
    # Single extension point for Epic 2+ entity-scope rules. Today every
    # action is role-coarse; future stories swap this for per-action
    # entity-scope evaluators.
    return True


# Test-only escape hatch. Production callers must use register_action().
def _reset_for_tests() -> None:
    _ACTION_ROLE_MAP.clear()
