"""Canonical conceptual-role names — single Django-side source of truth.

Mirrored in dotnet_auth, django_auth, and fiber_auth. The string values
are persisted as-is (Group.name on Django; ASPNetRole.NormalizedName on .NET;
fiber_auth.user_roles.role on Go), so they must match across stacks.
"""

from enum import StrEnum


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

