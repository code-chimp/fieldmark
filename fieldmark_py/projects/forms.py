"""ProjectCreateForm — validates and cleans project-create inputs.

Field names match the canonical contract in
docs/reference/project-create-form-contract.md exactly.
"""

from __future__ import annotations

import re
import uuid

from django import forms

_CODE_RE = re.compile(r"^[A-Z0-9][A-Z0-9-]*$")


class _LenientMultipleChoiceField(forms.MultipleChoiceField):
    """MultipleChoiceField that skips the built-in choice-membership check.

    Django's default MultipleChoiceField validates each submitted value against
    the choices list before calling clean_<field>. When a choice is stale (the
    option was valid when the page loaded but is no longer active), the default
    validator fires "Select a valid choice. X is not one of the available
    choices." instead of the AC-required contract message. Skipping the built-in
    check lets clean_trade_scope_ids / clean_inspector_ids emit the canonical
    "no longer available" message for every invalid-UUID case.
    """

    def valid_value(self, value: str) -> bool:  # noqa: ARG002
        return True


class ProjectCreateForm(forms.Form):
    code = forms.CharField(
        max_length=32,
        strip=True,
        error_messages={"required": "Code is required.", "max_length": "Code must be 32 characters or fewer."},
    )
    name = forms.CharField(
        max_length=200,
        strip=True,
        error_messages={"required": "Name is required.", "max_length": "Name must be 200 characters or fewer."},
    )
    description = forms.CharField(
        required=False,
        strip=True,
        max_length=10000,
        widget=forms.Textarea,
        error_messages={"max_length": "Description must be 10,000 characters or fewer."},
    )
    start_date = forms.DateField(
        input_formats=["%Y-%m-%d"],
        error_messages={
            "required": "Start date is required.",
            "invalid": "Start date must be a valid date (YYYY-MM-DD).",
        },
    )
    target_completion_date = forms.DateField(
        required=False,
        input_formats=["%Y-%m-%d"],
        error_messages={"invalid": "Target completion date must be a valid date."},
    )
    trade_scope_ids = _LenientMultipleChoiceField(
        choices=[],  # populated at runtime
        error_messages={"required": "At least one trade scope is required."},
    )
    inspector_ids = _LenientMultipleChoiceField(
        required=False,
        choices=[],  # populated at runtime
    )

    def __init__(self, *args, trade_type_choices=(), inspector_choices=(), **kwargs):
        super().__init__(*args, **kwargs)
        self.fields["trade_scope_ids"].choices = list(trade_type_choices)
        self.fields["inspector_ids"].choices = list(inspector_choices)

    def clean_code(self) -> str:
        code = str(self.cleaned_data.get("code") or "").strip()
        if not code:
            raise forms.ValidationError("Code is required.")
        if not _CODE_RE.match(code):
            if code.startswith("-"):
                raise forms.ValidationError("Code must start with a letter or digit.")
            raise forms.ValidationError(
                "Code must contain only uppercase letters, digits, and hyphens."
            )
        return code

    def clean_description(self) -> str | None:
        description = str(self.cleaned_data.get("description") or "").strip()
        return description or None

    def clean_trade_scope_ids(self) -> list[uuid.UUID]:
        raw = self.cleaned_data.get("trade_scope_ids", [])
        if not raw:
            raise forms.ValidationError("At least one trade scope is required.")
        try:
            return [uuid.UUID(v) for v in raw]
        except (ValueError, AttributeError) as exc:
            raise forms.ValidationError(
                "One or more selected trade types are no longer available. Please reselect."
            ) from exc

    def clean_inspector_ids(self) -> list[uuid.UUID]:
        raw = self.cleaned_data.get("inspector_ids", [])
        try:
            return [uuid.UUID(v) for v in raw]
        except (ValueError, AttributeError) as exc:
            raise forms.ValidationError(
                "One or more selected inspectors are no longer available. Please reselect."
            ) from exc

    def clean(self):
        cleaned = super().clean()
        start = cleaned.get("start_date")
        target = cleaned.get("target_completion_date")
        if start and target and target < start:
            self.add_error(
                "target_completion_date",
                "Target completion date must be on or after the start date.",
            )
        return cleaned
