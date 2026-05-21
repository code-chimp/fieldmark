"""Canonical conceptual-role names — single Django-side source of truth.

Mirrored in dotnet_auth, django_auth, and fiber_auth. The string values
are persisted as-is (Group.name on Django; ASPNetRole.NormalizedName on .NET;
fiber_auth.user_roles.role on Go), so they must match across stacks.
"""

import logging
from enum import StrEnum

logger = logging.getLogger(__name__)


class Role(StrEnum):
    ADMIN = "ADMIN"
    COMPLIANCE_OFFICER = "COMPLIANCE_OFFICER"
    INSPECTOR = "INSPECTOR"
    SITE_SUPERVISOR = "SITE_SUPERVISOR"
    EXECUTIVE = "EXECUTIVE"


# Title-cased display label per role (AC #4, Story 1.13).
LABELS: dict[Role, str] = {
    Role.ADMIN: "Admin",
    Role.COMPLIANCE_OFFICER: "Compliance Officer",
    Role.INSPECTOR: "Inspector",
    Role.SITE_SUPERVISOR: "Site Supervisor",
    Role.EXECUTIVE: "Executive",
}

# Badge token per role — pairs with .badge from Basecoat (AC #4, Story 1.13).
BADGE_TOKENS: dict[Role, str] = {
    Role.ADMIN: "danger",
    Role.COMPLIANCE_OFFICER: "info",
    Role.INSPECTOR: "warning",
    Role.SITE_SUPERVISOR: "neutral",
    Role.EXECUTIVE: "success",
}


def get_badge_token(role_name: str) -> str:
    """Return the CSS badge modifier token for a role name string.

    Returns "unknown" for any name not in the documented vocabulary and emits
    a server-side warning so unmapped roles are surfaced in logs without
    raising an exception (AC2.4 / Story 1.14).
    """
    try:
        role = Role(role_name)
        return BADGE_TOKENS[role]
    except ValueError:
        logger.warning("Unknown role badge token: %r", role_name)
        return "unknown"

