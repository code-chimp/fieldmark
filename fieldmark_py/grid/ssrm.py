"""SSRM request parser for AG Grid server-side row model endpoints.

Validates all client-supplied colIds, operators, and values against strict
allowlists before producing ORM Q objects and ordering tuples.

Contract: docs/reference/ag-grid-ssrm-contract.md
"""

from __future__ import annotations

import re
from dataclasses import dataclass, field
from datetime import date
from typing import Any

from django.db.models import Q


class SsrmError(ValueError):
    """Raised when an SSRM request fails allowlist validation."""


# Allowed colId → ORM field name mapping.
_COL_ALLOWLIST: dict[str, str] = {
    "code": "code",
    "name": "name",
    "status": "status",
    "compliance_score": "compliance_score",
    "start_date": "start_date",
    "target_completion_date": "target_completion_date",
}

_STATUS_ALLOWLIST = frozenset({"Active", "OnHold", "Closed"})

# Each column accepts exactly one filterType. Sending the wrong type → 400.
_COL_FILTER_TYPE: dict[str, str] = {
    "code":                   "text",
    "name":                   "text",
    "status":                 "set",
    "compliance_score":       "number",
    "start_date":             "date",
    "target_completion_date": "date",
}

_TEXT_OPS = frozenset({
    "equals", "notEqual", "contains", "notContains",
    "startsWith", "endsWith", "blank", "notBlank",
})
_NUMBER_OPS = frozenset({
    "equals", "notEqual", "greaterThan", "greaterThanOrEqual",
    "lessThan", "lessThanOrEqual", "inRange", "blank", "notBlank",
})
_DATE_OPS = frozenset({
    "equals", "notEqual", "greaterThan", "lessThan",
    "inRange", "blank", "notBlank",
})

# Strict YYYY-MM-DD pattern — rejects anything Postgres might coerce differently.
_DATE_PATTERN = re.compile(r"^\d{4}-\d{2}-\d{2}$")


@dataclass
class SsrmParsed:
    limit: int
    offset: int
    order_fields: list[str] = field(default_factory=list)
    q_filter: Q = field(default_factory=Q)
    # True when the filter logically matches nothing (empty Set Filter values=[]).
    match_nothing: bool = False


def parse_ssrm_request(data: Any) -> SsrmParsed:
    """Validate and translate an SSRM request dict into SsrmParsed.

    Raises SsrmError with a human-readable message on any validation failure.
    """
    # Guard the root shape — callers may pass any JSON value.
    if not isinstance(data, dict):
        raise SsrmError("request body must be a JSON object")

    start_row = data.get("startRow", 0)
    end_row = data.get("endRow", 100)
    sort_model = data.get("sortModel", [])
    filter_model = data.get("filterModel", {})

    # bool is a subclass of int in Python; reject it explicitly so True/False don't slip through.
    if isinstance(start_row, bool) or not isinstance(start_row, int) or start_row < 0:
        raise SsrmError("startRow must be >= 0")
    if isinstance(end_row, bool) or not isinstance(end_row, int) or end_row <= start_row:
        raise SsrmError("endRow must be greater than startRow")
    if end_row - start_row > 1000:
        raise SsrmError("page size exceeds maximum of 1000")
    if not isinstance(sort_model, list):
        raise SsrmError("sortModel must be an array")
    if not isinstance(filter_model, dict):
        raise SsrmError("filterModel must be an object")

    parsed = SsrmParsed(limit=end_row - start_row, offset=start_row)

    # --- Sort ---
    for entry in sort_model:
        if not isinstance(entry, dict):
            raise SsrmError("sortModel entries must be objects")
        col_id = entry.get("colId", "")
        sort_dir = entry.get("sort", "")
        if col_id not in _COL_ALLOWLIST:
            raise SsrmError(f"unknown column: {col_id}")
        if sort_dir not in ("asc", "desc"):
            raise SsrmError(f"invalid sort direction: {sort_dir}")
        orm_field = _COL_ALLOWLIST[col_id]
        parsed.order_fields.append(f"-{orm_field}" if sort_dir == "desc" else orm_field)

    # Always append tiebreaker for stable pagination.
    parsed.order_fields.append("id")

    # --- Filter ---
    q = Q()
    match_nothing = False
    for col_id, f in filter_model.items():
        if not isinstance(f, dict):
            raise SsrmError(f"filterModel entry for '{col_id}' must be an object")
        if col_id not in _COL_ALLOWLIST:
            raise SsrmError(f"unknown column: {col_id}")
        orm_field = _COL_ALLOWLIST[col_id]
        filter_type = f.get("filterType", "")

        # Reject filterType that doesn't match the column's declared type.
        expected_type = _COL_FILTER_TYPE.get(col_id)
        if expected_type and filter_type != expected_type:
            raise SsrmError(
                f"column '{col_id}' only accepts filterType '{expected_type}', got '{filter_type}'"
            )

        if filter_type == "set":
            if col_id != "status":
                raise SsrmError("set filter only supported on status column")
            values = f.get("values", [])
            if not isinstance(values, list):
                raise SsrmError("set filter 'values' must be an array")
            for v in values:
                if v not in _STATUS_ALLOWLIST:
                    raise SsrmError(f"invalid status value: {v}")
            if not values:
                match_nothing = True
            else:
                q &= Q(**{f"{orm_field}__in": values})

        elif filter_type == "text":
            q &= _parse_text_filter(orm_field, col_id, f)

        elif filter_type == "number":
            q &= _parse_number_filter(orm_field, col_id, f)

        elif filter_type == "date":
            q &= _parse_date_filter(orm_field, col_id, f)

        else:
            raise SsrmError(f"unknown filterType for column {col_id}")

    parsed.q_filter = q
    parsed.match_nothing = match_nothing
    return parsed


def _parse_text_filter(orm_field: str, col_id: str, f: dict) -> Q:
    op = f.get("type", "")
    if op not in _TEXT_OPS:
        raise SsrmError(f"invalid operator '{op}' for column '{col_id}'")
    val = f.get("filter", "")
    # For value-using operators the filter must be a string; a number would silently
    # pass through and produce semantically wrong ORM queries.
    if op not in ("blank", "notBlank") and not isinstance(val, str):
        raise SsrmError(
            f"text filter 'filter' for column '{col_id}' must be a string, got {type(val).__name__}"
        )
    match op:
        case "blank":
            return Q(**{f"{orm_field}__isnull": True}) | Q(**{f"{orm_field}": ""})
        case "notBlank":
            return ~Q(**{f"{orm_field}__isnull": True}) & ~Q(**{f"{orm_field}": ""})
        case "equals":
            return Q(**{f"{orm_field}": val})
        case "notEqual":
            return ~Q(**{f"{orm_field}": val})
        case "contains":
            return Q(**{f"{orm_field}__icontains": val})
        case "notContains":
            return ~Q(**{f"{orm_field}__icontains": val})
        case "startsWith":
            return Q(**{f"{orm_field}__istartswith": val})
        case "endsWith":
            return Q(**{f"{orm_field}__iendswith": val})
    return Q()


def _parse_number_filter(orm_field: str, col_id: str, f: dict) -> Q:
    op = f.get("type", "")
    if op not in _NUMBER_OPS:
        raise SsrmError(f"invalid operator '{op}' for column '{col_id}'")

    # Validate operands for operators that require them.
    _needs_val = op not in ("blank", "notBlank")
    val = f.get("filter")
    val_to = f.get("filterTo")

    # bool is a subclass of int; reject True/False as operands — they are not valid numbers
    # in the SSRM contract and would pass as 1/0, producing silent incorrect matches.
    def _is_numeric(v: object) -> bool:
        return isinstance(v, (int, float)) and not isinstance(v, bool)

    if _needs_val and not _is_numeric(val):
        raise SsrmError(
            f"operator '{op}' for column '{col_id}' requires a numeric 'filter' value"
        )
    if op == "inRange" and not _is_numeric(val_to):
        raise SsrmError(
            f"inRange for column '{col_id}' requires a numeric 'filterTo' value"
        )

    match op:
        case "blank":
            return Q(**{f"{orm_field}__isnull": True})
        case "notBlank":
            return ~Q(**{f"{orm_field}__isnull": True})
        case "equals":
            return Q(**{f"{orm_field}": val})
        case "notEqual":
            return ~Q(**{f"{orm_field}": val})
        case "greaterThan":
            return Q(**{f"{orm_field}__gt": val})
        case "greaterThanOrEqual":
            return Q(**{f"{orm_field}__gte": val})
        case "lessThan":
            return Q(**{f"{orm_field}__lt": val})
        case "lessThanOrEqual":
            return Q(**{f"{orm_field}__lte": val})
        case "inRange":
            return Q(**{f"{orm_field}__gte": val, f"{orm_field}__lte": val_to})
    return Q()


def _parse_date_filter(orm_field: str, col_id: str, f: dict) -> Q:
    op = f.get("type", "")
    if op not in _DATE_OPS:
        raise SsrmError(f"invalid operator '{op}' for column '{col_id}'")

    date_from = f.get("dateFrom", "")
    date_to = f.get("dateTo", "")

    # Operators that need dateFrom.
    _needs_from = op not in ("blank", "notBlank")
    if _needs_from:
        if not isinstance(date_from, str) or not _DATE_PATTERN.match(date_from):
            raise SsrmError(
                f"operator '{op}' for column '{col_id}' requires a valid YYYY-MM-DD 'dateFrom'"
            )
        # Regex only checks shape (YYYY-MM-DD digits); fromisoformat validates calendar correctness
        # (e.g. rejects "2026-13-99" which matches the regex but is not a real date).
        try:
            date.fromisoformat(date_from)
        except ValueError as exc:
            raise SsrmError(
                f"operator '{op}' for column '{col_id}' has an invalid calendar date in 'dateFrom': {date_from!r}"
            ) from exc
    if op == "inRange":
        if not isinstance(date_to, str) or not _DATE_PATTERN.match(date_to):
            raise SsrmError(
                f"inRange for column '{col_id}' requires a valid YYYY-MM-DD 'dateTo'"
            )
        try:
            date.fromisoformat(date_to)
        except ValueError as exc:
            raise SsrmError(
                f"inRange for column '{col_id}' has an invalid calendar date in 'dateTo': {date_to!r}"
            ) from exc

    match op:
        case "blank":
            return Q(**{f"{orm_field}__isnull": True})
        case "notBlank":
            return ~Q(**{f"{orm_field}__isnull": True})
        case "equals":
            return Q(**{f"{orm_field}": date_from})
        case "notEqual":
            return ~Q(**{f"{orm_field}": date_from})
        case "greaterThan":
            return Q(**{f"{orm_field}__gt": date_from})
        case "lessThan":
            return Q(**{f"{orm_field}__lt": date_from})
        case "inRange":
            return Q(**{f"{orm_field}__gte": date_from, f"{orm_field}__lte": date_to})
    return Q()
