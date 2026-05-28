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
    user = User.objects.create_user(username="reference_admin", password="FieldMark!2026")
    user.groups.set([group])
    return user


@pytest.mark.django_db
def test_reference_index_renders_three_sections_in_order(monkeypatch, admin_user: User) -> None:
    trade_id = uuid.UUID("a1b2c3d4-0001-0001-0001-000000000001")
    monkeypatch.setattr(
        "reference.queries.list_trade_types",
        lambda: [
            SimpleNamespace(
                code="ELEC",
                name="Electrical",
                description=None,
                active=True,
            )
        ],
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
    resp = client.get("/admin/reference")
    html = resp.content.decode()

    assert resp.status_code == 200
    assert html.index("<h2>Trade Types</h2>") < html.index("<h2>Violation Categories</h2>")
    assert html.index("<h2>Violation Categories</h2>") < html.index("<h2>Compliance Rules</h2>")
    assert html.count("<tbody>") == 3
    assert html.count("<tr>") == 6
    assert "OPEN_VIOLATION_GATE" in html
    assert '{"blocking_statuses":["Open","InProgress"]}' in unescape(html)
    assert ">None<" not in html
