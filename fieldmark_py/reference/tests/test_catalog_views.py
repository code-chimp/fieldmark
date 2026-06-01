from __future__ import annotations

import uuid
from html import unescape
from types import SimpleNamespace

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client

from fieldmark.roles import Role


@pytest.fixture()
def admin_user(db) -> User:
    group, _ = Group.objects.get_or_create(name=Role.ADMIN.value)
    user = User.objects.create_user(username="reference_catalog_admin", password="FieldMark!2026")
    user.groups.set([group])
    return user


@pytest.mark.django_db
@pytest.mark.parametrize(
    ("path", "heading", "expected"),
    [
        ("/admin/reference/trade-types", "Trade Types", "ELEC"),
        ("/admin/reference/violation-categories", "Violation Categories", "ELEC_NO_GFCI"),
        ("/admin/reference/compliance-rules", "Compliance Rules", "OPEN_VIOLATION_GATE"),
    ],
)
def test_reference_catalog_admin_pages_render(monkeypatch, admin_user: User, path: str, heading: str, expected: str) -> None:
    trade_id = uuid.UUID("a1b2c3d4-0001-0001-0001-000000000001")
    monkeypatch.setattr(
        "reference.queries.list_trade_types",
        lambda: [SimpleNamespace(code="ELEC", name="Electrical", description=None, active=True)],
    )
    monkeypatch.setattr(
        "reference.queries.list_violation_categories",
        lambda: [
            SimpleNamespace(
                code="ELEC_NO_GFCI",
                name="Missing GFCI Protection",
                trade_type=trade_id,
                default_severity="High",
                description=None,
                active=True,
            )
        ],
    )
    monkeypatch.setattr(
        "reference.queries.list_compliance_rules",
        lambda: [
            SimpleNamespace(
                code="OPEN_VIOLATION_GATE",
                name="Open Violation Closure Gate",
                description="Blocks closure with open violations",
                rule_kind="ClosureGate",
                parameters={"blocking_statuses": ["Open", "InProgress"]},
                active=True,
            )
        ],
    )

    client = Client()
    client.force_login(admin_user)
    resp = client.get(path)
    html = resp.content.decode()

    assert resp.status_code == 200
    assert f"<h1>{heading}</h1>" in html
    assert expected in html
    assert 'aria-label="Reference catalogs"' in html
    assert 'href="/admin/reference"' in html
    if path == "/admin/reference/trade-types":
        assert 'href="/admin/reference/trade-types"' not in html
    elif path == "/admin/reference/violation-categories":
        assert 'href="/admin/reference/violation-categories"' not in html
    elif path == "/admin/reference/compliance-rules":
        assert 'href="/admin/reference/compliance-rules"' not in html
    if path == "/admin/reference/compliance-rules":
        assert '{"blocking_statuses":["Open","InProgress"]}' in unescape(html)


@pytest.mark.django_db
@pytest.mark.parametrize("path", [
    "/admin/reference/trade-types",
    "/admin/reference/violation-categories",
    "/admin/reference/compliance-rules",
])
def test_reference_catalog_empty_state(monkeypatch, admin_user: User, path: str) -> None:
    monkeypatch.setattr("reference.queries.list_trade_types", lambda: [])
    monkeypatch.setattr("reference.queries.list_violation_categories", lambda: [])
    monkeypatch.setattr("reference.queries.list_compliance_rules", lambda: [])

    client = Client()
    client.force_login(admin_user)
    resp = client.get(path)
    html = resp.content.decode()

    assert resp.status_code == 200
    if path.endswith("trade-types"):
        assert "No trade types defined." in html
    elif path.endswith("violation-categories"):
        assert "No violation categories defined." in html
    else:
        assert "No compliance rules defined." in html
