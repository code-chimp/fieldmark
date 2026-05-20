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

