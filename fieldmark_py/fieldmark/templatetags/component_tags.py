# Contract: docs/reference/component-canonical-examples.md
from __future__ import annotations

from django import template

register = template.Library()


@register.filter(name="required_prop")
def required_prop(value: object, prop_name: str) -> object:
    """Raise ValueError if value is empty or whitespace-only.

    Usage in templates: {{ aria_label|required_prop:"aria_label" }}
    Enforces required props in component wrappers without a Python view layer.
    """
    if not (value and str(value).strip()):
        raise ValueError(f"TabStrip: {prop_name} is required and must not be empty or whitespace")
    return value
