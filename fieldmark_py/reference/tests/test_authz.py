from __future__ import annotations

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client

from fieldmark.roles import Role


@pytest.fixture()
def client() -> Client:
    return Client()


@pytest.fixture()
def role_user(db) -> User:
    user = User.objects.create_user(username="reference_role_user", password="FieldMark!2026")
    return user


@pytest.mark.django_db
@pytest.mark.parametrize(
    "role",
    [
        Role.COMPLIANCE_OFFICER,
        Role.INSPECTOR,
        Role.SITE_SUPERVISOR,
        Role.EXECUTIVE,
    ],
)
def test_reference_index_non_admin_returns_403_without_reference_state(
    monkeypatch, client: Client, role_user: User, role: Role
) -> None:
    def fail_if_called():
        raise AssertionError("reference queries must not run for non-admin users")

    monkeypatch.setattr("reference.queries.list_trade_types", fail_if_called)
    monkeypatch.setattr("reference.queries.list_violation_categories", fail_if_called)
    monkeypatch.setattr("reference.queries.list_compliance_rules", fail_if_called)
    group, _ = Group.objects.get_or_create(name=role.value)
    role_user.groups.set([group])
    client.force_login(role_user)

    resp = client.get("/admin/reference")
    body = resp.content.decode()

    assert resp.status_code == 403
    for protected in [
        "ELEC",
        "ELEC_NO_GFCI",
        "OPEN_VIOLATION_GATE",
        "rule_kind",
        "parameters",
    ]:
        assert protected not in body
