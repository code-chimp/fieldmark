"""View-level tests for the project-create form (Story 2.8).

Tests are grouped:
  - AUTH: 403 / login-redirect (no DB needed, no domain queries)
  - METHOD: 405 on GET /projects/ (no DB needed)
  - FORM: 422 validation cases (monkeypatched reference data; Django test DB for auth)
  - INTEGRATION: happy path + audit (marked; requires domain.* schema via make up)

See docs/reference/project-create-form-contract.md for the contract.
"""

from __future__ import annotations

import uuid
from typing import Any

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client

from fieldmark.authz import _reset_for_tests, register_action
from fieldmark.roles import Role

TRADE_ID_ELEC = uuid.UUID("a1b2c3d4-0001-0001-0001-000000000001")
INSPECTOR_UUID = uuid.UUID("93435c4e-f246-432f-b9fd-6fc5fbce5e9f")  # Ravi (INSPECTOR)
ADMIN_UUID = uuid.UUID("372da3c7-1cf5-4455-9f01-005117e48d76")  # Aisha (ADMIN)

# Choices used in all validation tests.
TRADE_CHOICES = [(str(TRADE_ID_ELEC), "Electrical")]
INSPECTOR_CHOICES = [(str(INSPECTOR_UUID), "Ravi Kumar")]


@pytest.fixture(autouse=True)
def _register_permission():
    """Ensure project.create is registered for ADMIN."""
    _reset_for_tests()
    register_action("project.create", Role.ADMIN)
    yield
    _reset_for_tests()


@pytest.fixture
def admin_user(db) -> User:
    group, _ = Group.objects.get_or_create(name=Role.ADMIN.value)
    user = User.objects.create_user(username="test_admin", password="pass")
    user.groups.set([group])
    return user


@pytest.fixture
def compliance_user(db) -> User:
    group, _ = Group.objects.get_or_create(name=Role.COMPLIANCE_OFFICER.value)
    user = User.objects.create_user(username="test_compliance", password="pass")
    user.groups.set([group])
    return user


def _client_login(user: User) -> Client:
    c = Client()
    c.force_login(user)
    return c


def _fake_reference_data(monkeypatch):
    """Patch _get_reference_data to return fixed stub choices."""
    monkeypatch.setattr(
        "projects.views._get_reference_data",
        lambda: (TRADE_CHOICES, INSPECTOR_CHOICES),
    )


# ─── AUTH ─────────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_get_projects_new_unauthenticated_redirects_to_login() -> None:
    c = Client()
    resp = c.get("/projects/new")
    assert resp.status_code == 302
    assert "/login" in resp["Location"]


@pytest.mark.django_db
def test_get_projects_new_non_admin_returns_403(compliance_user: User, monkeypatch) -> None:
    _fake_reference_data(monkeypatch)
    c = _client_login(compliance_user)
    resp = c.get("/projects/new")
    assert resp.status_code == 403


@pytest.mark.django_db
def test_post_projects_non_admin_returns_403(compliance_user: User) -> None:
    c = _client_login(compliance_user)
    resp = c.post("/projects/", {})
    assert resp.status_code == 403


# ─── METHOD ───────────────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_get_projects_collection_returns_405(admin_user: User) -> None:
    c = _client_login(admin_user)
    resp = c.get("/projects/")
    assert resp.status_code == 405
    assert resp.get("Allow") is not None


# ─── FORM RENDERING ───────────────────────────────────────────────────────────

@pytest.mark.django_db
def test_get_projects_new_renders_form(admin_user: User, monkeypatch) -> None:
    _fake_reference_data(monkeypatch)
    c = _client_login(admin_user)
    resp = c.get("/projects/new")
    assert resp.status_code == 200
    content = resp.content.decode()
    assert 'name="code"' in content
    assert 'name="name"' in content
    assert 'name="trade_scope_ids"' in content


# ─── VALIDATION (422) ─────────────────────────────────────────────────────────

def _post_form(admin_user: User, monkeypatch, **fields) -> Any:
    _fake_reference_data(monkeypatch)
    c = _client_login(admin_user)
    # Default data; caller overrides per test case.
    # Setting a key to None removes it from the data dict (no submission).
    base = {
        "code": "PROJ-A",
        "name": "Test Project",
        "start_date": "2026-06-01",
        "trade_scope_ids": str(TRADE_ID_ELEC),
    }
    merged = {**base, **fields}
    data = {k: v for k, v in merged.items() if v is not None}
    return c.post("/projects/", data)


@pytest.mark.django_db
@pytest.mark.parametrize(
    "overrides, expected_error",
    [
        ({"code": ""}, "Code is required."),
        ({"code": "   "}, "Code is required."),
        ({"code": "A" * 33}, "Code must be 32 characters or fewer."),
        ({"code": "abc!"}, "Code must contain only uppercase letters, digits, and hyphens."),
        ({"code": "-PROJ"}, "Code must start with a letter or digit."),
        ({"name": ""}, "Name is required."),
        ({"name": "N" * 201}, "Name must be 200 characters or fewer."),
        ({"description": "D" * 10001}, "Description must be 10,000 characters or fewer."),
        ({"start_date": ""}, "Start date is required."),
        ({"start_date": "not-a-date"}, "Start date must be a valid date (YYYY-MM-DD)."),
        ({"trade_scope_ids": None}, "At least one trade scope is required."),  # omit trade_scope_ids
    ],
)
def test_post_projects_validation_422(
    admin_user: User, monkeypatch, overrides: dict, expected_error: str
) -> None:
    resp = _post_form(admin_user, monkeypatch, **overrides)
    assert resp.status_code == 422, f"expected 422, got {resp.status_code}"
    content = resp.content.decode()
    assert expected_error in content, f"expected '{expected_error}' in response"


@pytest.mark.django_db
def test_post_projects_target_before_start_422(admin_user: User, monkeypatch) -> None:
    resp = _post_form(
        admin_user,
        monkeypatch,
        start_date="2026-06-01",
        target_completion_date="2026-05-01",
    )
    assert resp.status_code == 422
    content = resp.content.decode()
    assert "Target completion date must be on or after the start date." in content


@pytest.mark.django_db
def test_post_projects_422_no_oob_swaps(admin_user: User, monkeypatch) -> None:
    """UX Pattern 3: 422 response body must not contain hx-swap-oob."""
    resp = _post_form(admin_user, monkeypatch, code="")
    assert resp.status_code == 422
    assert "hx-swap-oob" not in resp.content.decode()


@pytest.mark.django_db
def test_post_projects_422_shows_inline_alert(admin_user: User, monkeypatch) -> None:
    resp = _post_form(admin_user, monkeypatch, code="")
    assert resp.status_code == 422
    content = resp.content.decode()
    assert "Couldn't create the project" in content
    assert 'role="alert"' in content


@pytest.mark.django_db
def test_post_projects_422_multiple_errors(admin_user: User, monkeypatch) -> None:
    """Multi-error case: all three fields show errors; alert says '3 errors'."""
    _fake_reference_data(monkeypatch)
    c = _client_login(admin_user)
    # Empty code + empty start_date + no trade scope = 3 errors
    resp = c.post("/projects/", {"name": "X", "code": "", "start_date": ""})  # no trade_scope_ids
    assert resp.status_code == 422
    content = resp.content.decode()
    assert "3 error" in content


@pytest.mark.django_db
def test_post_projects_csrf_token_present_in_get(admin_user: User, monkeypatch) -> None:
    """Django: form contains csrfmiddlewaretoken hidden input."""
    _fake_reference_data(monkeypatch)
    c = _client_login(admin_user)
    resp = c.get("/projects/new")
    assert resp.status_code == 200
    assert "csrfmiddlewaretoken" in resp.content.decode()


# ─── XSS / security (AC8 category 6) ─────────────────────────────────────────

@pytest.mark.django_db
def test_post_projects_xss_name_echoed_escaped(admin_user: User, monkeypatch) -> None:
    """XSS payload in name field must be escaped in 422 re-render."""
    _fake_reference_data(monkeypatch)
    c = _client_login(admin_user)
    xss = "<script>alert(1)</script>"
    resp = c.post(
        "/projects/",
        {"code": "PROJ-A", "name": xss, "start_date": "2026-06-01", "trade_scope_ids": ""},
    )
    assert resp.status_code == 422
    content = resp.content.decode()
    assert "&lt;script&gt;" in content
    assert "<script>" not in content
