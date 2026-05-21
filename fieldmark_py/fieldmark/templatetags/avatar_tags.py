from django import template
from django.contrib.auth.models import AbstractBaseUser

from fieldmark.avatar import initials
from fieldmark.roles import LABELS, Role

register = template.Library()


@register.filter
def avatar_initials(user: AbstractBaseUser) -> str:
    full_name = getattr(user, "get_full_name", lambda: "")()
    return initials(full_name or None, user.get_username())


@register.filter
def user_role_label(user: AbstractBaseUser) -> str:
    canonical = {r.value for r in Role}
    group_names = sorted(
        name
        for name in user.groups.values_list("name", flat=True)  # type: ignore[attr-defined]
        if name in canonical
    )
    role_name = group_names[0] if group_names else ""
    try:
        role: Role | None = Role(role_name)
    except ValueError:
        role = None
    return LABELS[role] if role is not None else ""
