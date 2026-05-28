"""The single Django-side ``append_audit_entry`` helper for FR39/FR40.

Callers wrap the surrounding work in ``with transaction.atomic():``; this
function only issues the INSERT. A helper that opened its own transaction
would break FR39 — a rollback elsewhere in the handler would leave an orphan
audit row. See ``docs/reference/audit-actions.md`` for the canonical action
list and the casing convention.
"""

from __future__ import annotations

import uuid
from typing import Any

from .actions import AuditAction
from .models import AuditEntry


def append_audit_entry(
    *,
    actor_id: uuid.UUID,
    action: AuditAction,
    entity_type: str,
    entity_id: uuid.UUID,
    project_id: uuid.UUID | None = None,
    before_state: dict[str, Any] | None = None,
    after_state: dict[str, Any] | None = None,
    metadata: dict[str, Any] | None = None,
) -> AuditEntry:
    """Append an AuditEntry inside the caller's open ``transaction.atomic()`` block.

    Keyword-only — six similar UUID/JSON arguments would silently mis-bind
    positionally. ``action.value`` persists the PascalCase string verbatim
    (FR40).
    """

    return AuditEntry.objects.create(
        actor_id=actor_id,
        action=action.value,
        entity_type=entity_type,
        entity_id=entity_id,
        project_id=project_id,
        before_state=before_state,
        after_state=after_state,
        metadata=metadata,
    )
