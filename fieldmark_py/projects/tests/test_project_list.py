"""Tests for GET /projects — project list page (Story 2.9)."""

from __future__ import annotations

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client

from fieldmark.authz import _reset_for_tests, register_action
from fieldmark.roles import Role


@pytest.fixture(autouse=True)
def _register_permissions():
    """Ensure project.read (and project.create) are registered for this test run."""
    _reset_for_tests()
    register_action("project.read", Role.ADMIN, Role.COMPLIANCE_OFFICER, Role.INSPECTOR,
                    Role.SITE_SUPERVISOR, Role.EXECUTIVE)
    register_action("project.create", Role.ADMIN)
    yield
    _reset_for_tests()


@pytest.fixture
def admin_user(db) -> User:
    group, _ = Group.objects.get_or_create(name=Role.ADMIN.value)
    user = User.objects.create_user(username="test_admin_list", password="pass")
    user.groups.set([group])
    return user


@pytest.fixture
def co_user(db) -> User:
    group, _ = Group.objects.get_or_create(name=Role.COMPLIANCE_OFFICER.value)
    user = User.objects.create_user(username="test_co_list", password="pass")
    user.groups.set([group])
    return user


def _login(user: User) -> Client:
    c = Client()
    c.force_login(user)
    return c


@pytest.mark.django_db
def test_project_list_unauthenticated_redirects(client: Client):
    resp = client.get("/projects")
    assert resp.status_code in (301, 302)


@pytest.mark.django_db
def test_project_list_authenticated_admin_returns_200(admin_user):
    c = _login(admin_user)
    resp = c.get("/projects")
    assert resp.status_code == 200


@pytest.mark.django_db
def test_project_list_renders_h1(admin_user):
    c = _login(admin_user)
    resp = c.get("/projects")
    assert b"<h1>Projects</h1>" in resp.content


@pytest.mark.django_db
def test_project_list_renders_grid_container(admin_user):
    c = _login(admin_user)
    resp = c.get("/projects")
    assert b'data-grid-endpoint="/grid/projects"' in resp.content
    assert b'data-grid-target="#project-detail"' in resp.content
    assert b"ag-theme-quartz" in resp.content


@pytest.mark.django_db
def test_project_list_renders_project_detail_target(admin_user):
    c = _login(admin_user)
    resp = c.get("/projects")
    assert b'id="project-detail"' in resp.content
    assert b'tabindex="-1"' in resp.content


@pytest.mark.django_db
def test_project_list_renders_noscript_fallback(admin_user):
    c = _login(admin_user)
    resp = c.get("/projects")
    assert b"<noscript>" in resp.content


@pytest.mark.django_db
def test_project_list_admin_sees_new_project_button(admin_user):
    c = _login(admin_user)
    resp = c.get("/projects")
    assert b'href="/projects/new"' in resp.content, "admin must see server-rendered New Project link"


@pytest.mark.django_db
def test_project_list_non_admin_no_new_project_button(co_user):
    c = _login(co_user)
    resp = c.get("/projects")
    assert resp.status_code == 200
    assert b'href="/projects/new"' not in resp.content, "non-admin must not see New Project link"
