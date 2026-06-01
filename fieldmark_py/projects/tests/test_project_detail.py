from __future__ import annotations

import uuid
from datetime import date

import pytest
from django.contrib.auth.models import Group, User
from django.db import connection
from django.test import Client

from fieldmark.authz import _reset_for_tests, register_action
from fieldmark.roles import Role
from projects.models import Project, ProjectStatus


pytestmark = [pytest.mark.django_db, pytest.mark.integration]


@pytest.fixture(autouse=True)
def _register_permissions():
    _reset_for_tests()
    register_action("project.read", Role.ADMIN, Role.COMPLIANCE_OFFICER, Role.INSPECTOR, Role.SITE_SUPERVISOR, Role.EXECUTIVE)
    register_action("project.place_on_hold", Role.ADMIN)
    register_action("project.resume", Role.ADMIN)
    register_action("project.close", Role.ADMIN)
    yield
    _reset_for_tests()


@pytest.fixture
def admin_user() -> User:
    group, _ = Group.objects.get_or_create(name=Role.ADMIN.value)
    user = User.objects.create_user(username=f"pd_admin_{uuid.uuid4().hex[:8]}", password="pass")
    user.groups.set([group])
    return user


@pytest.fixture
def executive_user() -> User:
    group, _ = Group.objects.get_or_create(name=Role.EXECUTIVE.value)
    user = User.objects.create_user(username=f"pd_exec_{uuid.uuid4().hex[:8]}", password="pass")
    user.groups.set([group])
    return user


@pytest.fixture
def no_role_user() -> User:
    return User.objects.create_user(username=f"pd_norole_{uuid.uuid4().hex[:8]}", password="pass")


def _login(user: User) -> Client:
    c = Client()
    c.force_login(user)
    return c


def _create_project(
    *,
    status: ProjectStatus = ProjectStatus.ACTIVE,
    name: str = "Project Detail Django",
    description: str | None = None,
    target_completion_date: date | None = None,
) -> Project:
    pid = uuid.uuid4()
    return Project.objects.create(
        id=pid,
        code=f"PD-{pid.hex[:6].upper()}",
        name=name,
        description=description,
        status=status,
        start_date=date(2026, 6, 1),
        target_completion_date=target_completion_date,
        actual_closed_at=None,
        compliance_score=100,
    )


def _has_domain_table(table: str) -> bool:
    with connection.cursor() as cur:
        cur.execute("SELECT to_regclass(%s)", [f"domain.{table}"])
        row = cur.fetchone()
    return row is not None and row[0] is not None


def test_project_detail_hx_returns_fragment(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'id="project-detail"' in html
    assert 'id="project-header-strip"' in html
    assert 'id="project-detail-tabstrip"' in html
    assert 'id="project-detail-tab-content"' in html
    assert 'id="violation-detail"' in html
    assert "<html" not in html.lower()


def test_project_detail_tab_violations_returns_panel_and_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/violations", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'aria-labelledby="tab-violations"' in html
    assert 'hx-swap-oob="outerHTML"' in html
    assert 'id="project-detail-tabstrip"' in html
    assert 'id="tab-violations"' in html
    assert 'aria-selected="true"' in html


def test_project_detail_tab_non_htmx_redirects_to_detail(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/violations")
    assert resp.status_code in (301, 302)
    assert resp["Location"].endswith(f"/projects/{p.id}")


def test_project_detail_no_role_user_forbidden(no_role_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(no_role_user)
    resp = c.get(f"/projects/{p.id}")
    assert resp.status_code == 403


def test_project_detail_admin_active_shows_hold_close_and_disables_resume(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    html = c.get(f"/projects/{p.id}").content.decode()
    assert 'id="place-on-hold-btn"' in html
    assert 'id="resume-btn"' in html
    assert 'id="close-btn"' in html
    assert 'aria-describedby="resume-btn-reason"' in html


def test_project_detail_admin_onhold_shows_resume_and_disables_others(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    html = c.get(f"/projects/{p.id}").content.decode()
    assert 'id="place-on-hold-btn"' in html
    assert 'id="resume-btn"' in html
    assert 'id="close-btn"' in html
    assert 'aria-describedby="place-on-hold-btn-reason"' in html
    assert 'aria-describedby="close-btn-reason"' in html


def test_project_detail_admin_closed_disables_all_actions(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.CLOSED)
    c = _login(admin_user)
    html = c.get(f"/projects/{p.id}").content.decode()
    assert 'id="place-on-hold-btn"' in html
    assert 'id="resume-btn"' in html
    assert 'id="close-btn"' in html
    assert 'aria-describedby="place-on-hold-btn-reason"' in html
    assert 'aria-describedby="resume-btn-reason"' in html
    assert 'aria-describedby="close-btn-reason"' in html


def test_project_detail_executive_hides_all_action_buttons(executive_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(executive_user)
    html = c.get(f"/projects/{p.id}").content.decode()
    assert 'id="place-on-hold-btn"' not in html
    assert 'id="resume-btn"' not in html
    assert 'id="close-btn"' not in html


def test_project_detail_xss_payloads_are_escaped(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    payload = "<script>alert(1)</script>"
    p = _create_project(name=payload, description=payload, target_completion_date=date(2026, 6, 30))
    c = _login(admin_user)
    html = c.get(f"/projects/{p.id}").content.decode()
    assert html.count("&lt;script&gt;alert(1)&lt;/script&gt;") >= 2
    assert "<script>alert(1)</script>" not in html
