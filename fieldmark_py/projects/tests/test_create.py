"""Unit tests for Project.create classmethod (Story 2.8)."""

from __future__ import annotations

import uuid
from datetime import date

import pytest

from projects.models import Project

TODAY = date(2026, 6, 1)
TRADE_ID_1 = uuid.uuid4()
TRADE_ID_2 = uuid.uuid4()
INSPECTOR_ID = uuid.uuid4()


def make_project(**kwargs):
    defaults = dict(
        code="BLDG-A",
        name="Building A",
        description=None,
        start_date=TODAY,
        target_completion_date=None,
        trade_scope_ids=[TRADE_ID_1],
        inspector_ids=[],
    )
    defaults.update(kwargs)
    return Project.create(**defaults)


# ── Happy path ──────────────────────────────────────────────────────────────

def test_create_returns_active_status():
    project, _, _ = make_project()
    assert project.status == "Active"


def test_create_generates_unique_ids():
    p1, _, _ = make_project()
    p2, _, _ = make_project()
    assert p1.id is not None
    assert p1.id != p2.id


def test_create_compliance_score_100():
    project, _, _ = make_project()
    assert project.compliance_score == 100


def test_create_trims_code_and_name():
    project, _, _ = make_project(code="  BLDG-A  ", name="  Building A  ")
    assert project.code == "BLDG-A"
    assert project.name == "Building A"


def test_create_trade_scopes_linked_to_project():
    project, scopes, _ = make_project(trade_scope_ids=[TRADE_ID_1, TRADE_ID_2])
    assert len(scopes) == 2
    for scope in scopes:
        assert scope.project_id == project.id
    assert {s.trade_type_id for s in scopes} == {TRADE_ID_1, TRADE_ID_2}


def test_create_inspectors_linked_to_project():
    project, _, inspectors = make_project(inspector_ids=[INSPECTOR_ID])
    assert len(inspectors) == 1
    assert inspectors[0].project_id == project.id
    assert inspectors[0].user_id == INSPECTOR_ID


def test_create_empty_inspectors_returns_empty_list():
    _, _, inspectors = make_project(inspector_ids=[])
    assert inspectors == []


def test_create_none_description_stores_none():
    project, _, _ = make_project(description=None)
    assert project.description is None


def test_create_whitespace_description_stores_none():
    project, _, _ = make_project(description="   ")
    assert project.description is None


def test_create_description_trimmed():
    project, _, _ = make_project(description="  hello  ")
    assert project.description == "hello"


# ── Validation errors ────────────────────────────────────────────────────────

def test_create_empty_code_raises():
    with pytest.raises(ValueError, match="code"):
        make_project(code="")


def test_create_whitespace_code_raises():
    with pytest.raises(ValueError, match="code"):
        make_project(code="   ")


def test_create_empty_name_raises():
    with pytest.raises(ValueError, match="name"):
        make_project(name="")


def test_create_empty_trade_scope_ids_raises():
    with pytest.raises(ValueError, match="trade scope"):
        make_project(trade_scope_ids=[])


def test_create_target_before_start_raises():
    with pytest.raises(ValueError, match="target_completion_date"):
        make_project(
            start_date=date(2026, 6, 1),
            target_completion_date=date(2026, 5, 1),
        )


def test_create_target_equal_to_start_succeeds():
    # boundary: target == start is allowed
    project, _, _ = make_project(
        start_date=date(2026, 6, 1),
        target_completion_date=date(2026, 6, 1),
    )
    assert project.target_completion_date == date(2026, 6, 1)
