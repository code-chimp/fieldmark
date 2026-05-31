"""Tests for grid/ssrm.py SSRM parser and POST /grid/projects endpoint.

Unit tests (parser) run without a database.
Integration tests that hit the database are in test_grid_projects_integration.py
(if present; skipped when DB unavailable).
"""

from __future__ import annotations

import pytest

from grid.ssrm import SsrmError, parse_ssrm_request


def req(start=0, end=10, sort=None, filter_model=None):
    return {
        "startRow": start,
        "endRow": end,
        "sortModel": sort or [],
        "filterModel": filter_model or {},
    }


# ─── Pagination bounds ────────────────────────────────────────────────────────

def test_parse_negative_start_row_raises():
    with pytest.raises(SsrmError, match="startRow"):
        parse_ssrm_request(req(start=-1))


def test_parse_end_not_greater_than_start_raises():
    with pytest.raises(SsrmError, match="endRow"):
        parse_ssrm_request(req(start=5, end=5))


def test_parse_page_size_exceeds_max_raises():
    with pytest.raises(SsrmError, match="page size"):
        parse_ssrm_request(req(start=0, end=1001))


def test_parse_valid_pagination_sets_limit_offset():
    p = parse_ssrm_request(req(start=10, end=20))
    assert p.limit == 10
    assert p.offset == 10


# ─── Sort allowlist ───────────────────────────────────────────────────────────

def test_parse_sort_entry_not_dict_raises():
    with pytest.raises(SsrmError, match="must be objects"):
        parse_ssrm_request(req(sort=["not_a_dict"]))


def test_parse_unknown_sort_col_raises():
    with pytest.raises(SsrmError, match="unknown column"):
        parse_ssrm_request(req(sort=[{"colId": "description", "sort": "asc"}]))


def test_parse_invalid_sort_direction_raises():
    with pytest.raises(SsrmError, match="invalid sort direction"):
        parse_ssrm_request(req(sort=[{"colId": "code", "sort": "DESC"}]))


def test_parse_injection_in_sort_col_raises():
    with pytest.raises(SsrmError, match="unknown column"):
        parse_ssrm_request(req(sort=[{"colId": "code; DROP TABLE domain.project --", "sort": "asc"}]))


def test_parse_valid_sort_always_has_tiebreaker():
    p = parse_ssrm_request(req(sort=[{"colId": "compliance_score", "sort": "asc"}]))
    assert p.order_fields[-1] == "id"


def test_parse_empty_sort_has_tiebreaker():
    p = parse_ssrm_request(req())
    assert "id" in p.order_fields


# ─── Filter allowlist ─────────────────────────────────────────────────────────

def test_parse_unknown_filter_col_raises():
    with pytest.raises(SsrmError, match="unknown column"):
        parse_ssrm_request(req(filter_model={"description": {"filterType": "text", "type": "contains", "filter": "x"}}))


def test_parse_injection_in_filter_col_raises():
    with pytest.raises(SsrmError, match="unknown column"):
        parse_ssrm_request(req(filter_model={"code; DROP TABLE domain.project --": {"filterType": "text", "type": "contains", "filter": "x"}}))


def test_parse_invalid_text_operator_raises():
    with pytest.raises(SsrmError, match="invalid operator"):
        parse_ssrm_request(req(filter_model={"code": {"filterType": "text", "type": "INVALID", "filter": "x"}}))


def test_parse_invalid_number_operator_raises():
    with pytest.raises(SsrmError, match="invalid operator"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "number", "type": "INVALID", "filter": 50}}))


def test_parse_invalid_date_operator_raises():
    with pytest.raises(SsrmError, match="invalid operator"):
        parse_ssrm_request(req(filter_model={"start_date": {"filterType": "date", "type": "INVALID", "dateFrom": "2026-01-01"}}))


def test_parse_invalid_status_value_raises():
    with pytest.raises(SsrmError, match="invalid status value"):
        parse_ssrm_request(req(filter_model={"status": {"filterType": "set", "values": ["Active", "INVALID"]}}))


def test_parse_injection_in_status_value_raises():
    with pytest.raises(SsrmError, match="invalid status value"):
        parse_ssrm_request(req(filter_model={"status": {"filterType": "set", "values": ["Active' OR '1'='1"]}}))


def test_parse_empty_status_values_sets_match_nothing():
    p = parse_ssrm_request(req(filter_model={"status": {"filterType": "set", "values": []}}))
    assert p.match_nothing is True


# ─── Body shape guards ────────────────────────────────────────────────────────

def test_parse_boolean_start_row_raises():
    # bool is a subclass of int; True == 1 would otherwise pass the int check.
    with pytest.raises(SsrmError, match="startRow"):
        parse_ssrm_request({"startRow": True, "endRow": 10, "sortModel": [], "filterModel": {}})


def test_parse_boolean_end_row_raises():
    with pytest.raises(SsrmError, match="endRow"):
        parse_ssrm_request({"startRow": 0, "endRow": True, "sortModel": [], "filterModel": {}})


def test_parse_date_filter_invalid_calendar_date_raises():
    # Regex passes for "2026-13-99" but fromisoformat should reject it.
    with pytest.raises(SsrmError, match="invalid calendar date"):
        parse_ssrm_request(req(filter_model={"start_date": {"filterType": "date", "type": "equals", "dateFrom": "2026-13-99"}}))


def test_parse_date_filter_valid_calendar_date_passes():
    p = parse_ssrm_request(req(filter_model={"start_date": {"filterType": "date", "type": "equals", "dateFrom": "2026-06-15"}}))
    assert p.match_nothing is False


def test_parse_data_not_dict_raises():
    with pytest.raises(SsrmError, match="JSON object"):
        parse_ssrm_request([])  # type: ignore[arg-type]


def test_parse_sort_model_not_list_raises():
    with pytest.raises(SsrmError, match="array"):
        parse_ssrm_request({"startRow": 0, "endRow": 10, "sortModel": "bad", "filterModel": {}})


def test_parse_filter_model_not_dict_raises():
    with pytest.raises(SsrmError, match="object"):
        parse_ssrm_request({"startRow": 0, "endRow": 10, "sortModel": [], "filterModel": []})


# ─── Number filter operand validation ────────────────────────────────────────

def test_parse_number_filter_missing_val_raises():
    with pytest.raises(SsrmError, match="numeric 'filter'"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "number", "type": "equals"}}))


def test_parse_number_filter_string_val_raises():
    with pytest.raises(SsrmError, match="numeric 'filter'"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "number", "type": "greaterThan", "filter": "not-a-number"}}))


def test_parse_number_filter_inrange_missing_filterTo_raises():
    with pytest.raises(SsrmError, match="filterTo"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "number", "type": "inRange", "filter": 50}}))


def test_parse_text_filter_numeric_value_raises():
    # AG Grid text filter sends "filter" as a string; a number is a malformed payload.
    with pytest.raises(SsrmError, match="must be a string"):
        parse_ssrm_request(req(filter_model={"code": {"filterType": "text", "type": "contains", "filter": 42}}))


def test_parse_text_filter_string_value_passes():
    p = parse_ssrm_request(req(filter_model={"code": {"filterType": "text", "type": "contains", "filter": "BLDG"}}))
    assert p.match_nothing is False


def test_parse_number_filter_boolean_true_raises():
    # True == 1 but booleans are not valid numeric operands per the contract.
    with pytest.raises(SsrmError, match="numeric 'filter'"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "number", "type": "equals", "filter": True}}))


def test_parse_number_filter_boolean_false_raises():
    with pytest.raises(SsrmError, match="numeric 'filter'"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "number", "type": "greaterThan", "filter": False}}))


# ─── Date filter operand validation ───────────────────────────────────────────

def test_parse_date_filter_missing_dateFrom_raises():
    with pytest.raises(SsrmError, match="dateFrom"):
        parse_ssrm_request(req(filter_model={"start_date": {"filterType": "date", "type": "equals"}}))


def test_parse_date_filter_malformed_date_raises():
    with pytest.raises(SsrmError, match="YYYY-MM-DD"):
        parse_ssrm_request(req(filter_model={"start_date": {"filterType": "date", "type": "equals", "dateFrom": "not-a-date"}}))


def test_parse_date_filter_inrange_missing_dateTo_raises():
    with pytest.raises(SsrmError, match="dateTo"):
        parse_ssrm_request(req(filter_model={"start_date": {"filterType": "date", "type": "inRange", "dateFrom": "2026-01-01"}}))


# ─── Per-column filterType allowlist ─────────────────────────────────────────

def test_parse_wrong_filter_type_for_column_raises():
    # compliance_score only accepts "number"; sending "text" must → 400.
    with pytest.raises(SsrmError, match="filterType"):
        parse_ssrm_request(req(filter_model={"compliance_score": {"filterType": "text", "type": "contains", "filter": "50"}}))


def test_parse_text_filter_type_on_date_column_raises():
    with pytest.raises(SsrmError, match="filterType"):
        parse_ssrm_request(req(filter_model={"start_date": {"filterType": "text", "type": "contains", "filter": "2026"}}))


def test_parse_number_filter_type_on_text_column_raises():
    with pytest.raises(SsrmError, match="filterType"):
        parse_ssrm_request(req(filter_model={"code": {"filterType": "number", "type": "equals", "filter": 42}}))
