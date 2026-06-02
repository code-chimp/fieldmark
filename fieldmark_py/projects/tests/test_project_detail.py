from __future__ import annotations

import uuid
from datetime import date

import pytest
from django.contrib.auth.models import Group, User
from django.db import connection
from django.test import Client

from audit.models import AuditEntry
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


def _count_oob_regions(html: str) -> int:
    return html.count('hx-swap-oob=')


def test_project_detail_hx_returns_fragment(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'id="project-detail"' not in html
    assert 'id="project-header-strip"' in html
    assert 'id="project-detail-tabstrip"' in html
    assert 'id="project-detail-tab-content"' in html
    assert 'id="violation-detail"' in html
    assert "<html" not in html.lower()


def test_project_detail_full_page_wraps_fragment(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert '<div id="project-detail">' in html
    assert 'id="project-header-strip"' in html


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
    assert f'hx-get="/projects/{p.id}/place-on-hold"' in html
    assert 'hx-target="#project-action-form"' in html
    assert 'hx-swap="innerHTML"' in html
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


def test_project_place_on_hold_get_renders_reason_form(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/place-on-hold", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'role="form"' in html
    assert f'hx-post="/projects/{p.id}/place-on-hold"' in html
    assert 'hx-target="#project-detail"' in html
    assert 'name="reason"' in html


def test_project_place_on_hold_get_forbidden_for_exec(executive_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(executive_user)
    resp = c.get(f"/projects/{p.id}/place-on-hold", HTTP_HX_REQUEST="true")
    assert resp.status_code == 403


def test_project_resume_get_renders_reason_form(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/resume", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'role="form"' in html
    assert f'hx-post="/projects/{p.id}/resume"' in html
    assert 'hx-target="#project-detail"' in html
    assert 'name="reason"' in html


def test_project_place_on_hold_post_success_renders_three_region_shape_and_persists_audit(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": "Weather delay"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'id="project-detail"' not in html
    assert 'id="compliance-tile"' in html
    assert 'hx-swap-oob="true"' in html
    assert 'hx-swap-oob="afterbegin:#audit-log"' in html
    assert _count_oob_regions(html) == 2
    p.refresh_from_db()
    assert p.status == ProjectStatus.ON_HOLD
    audit = AuditEntry.objects.filter(project_id=p.id).order_by("-occurred_at").first()
    assert audit is not None
    assert audit.action == "ProjectPlacedOnHold"
    assert audit.before_state == {"status": "Active"}
    assert audit.after_state == {"status": "OnHold"}
    assert audit.metadata == {"reason": "Weather delay"}


def test_project_resume_post_success_renders_three_region_shape_and_persists_audit(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/resume", {"reason": "Crew available"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'id="project-detail"' not in html
    assert 'id="compliance-tile"' in html
    assert 'hx-swap-oob="true"' in html
    assert 'hx-swap-oob="afterbegin:#audit-log"' in html
    assert _count_oob_regions(html) == 2
    p.refresh_from_db()
    assert p.status == ProjectStatus.ACTIVE
    audit = AuditEntry.objects.filter(project_id=p.id).order_by("-occurred_at").first()
    assert audit is not None
    assert audit.action == "ProjectResumed"
    assert audit.before_state == {"status": "OnHold"}
    assert audit.after_state == {"status": "Active"}
    assert audit.metadata == {"reason": "Crew available"}


def test_project_place_on_hold_post_blank_reason_returns_422_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": ""}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 422
    html = resp.content.decode()
    assert "Couldn't submit transition" in html
    assert 'id="reason-error"' in html
    assert "Reason is required." in html
    assert 'aria-invalid="true"' in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_place_on_hold_post_too_long_reason_returns_422_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": "x" * 501}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 422
    html = resp.content.decode()
    assert "Reason must be 500 characters or fewer." in html
    assert "Couldn't submit transition" in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_place_on_hold_post_control_char_reason_returns_422_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": "bad\x01reason"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 422
    html = resp.content.decode()
    assert "Reason contains invalid control characters." in html
    assert "Couldn't submit transition" in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_resume_post_from_active_returns_409_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/resume", {"reason": "stale request"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 409
    html = resp.content.decode()
    assert 'id="project-detail"' not in html
    assert "Couldn't resume project" in html
    assert "Project is not on hold" in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_place_on_hold_post_from_on_hold_returns_409_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": "stale request"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 409
    html = resp.content.decode()
    assert 'id="project-detail"' not in html
    assert "Couldn't place project on hold" in html
    assert "Project is already on hold" in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_place_on_hold_post_forbidden_for_exec_returns_403_without_audit(executive_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(executive_user)
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": "nope"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 403
    assert resp.content.decode() == "You do not have permission to access this page."
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_resume_post_forbidden_for_exec_returns_403_without_audit(executive_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(executive_user)
    resp = c.post(f"/projects/{p.id}/resume", {"reason": "nope"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 403
    assert resp.content.decode() == "You do not have permission to access this page."
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_place_on_hold_post_unknown_id_returns_404(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    c = _login(admin_user)
    resp = c.post(f"/projects/{uuid.uuid4()}/place-on-hold", {"reason": "missing"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 404


def test_project_place_on_hold_post_xss_reason_is_escaped_on_422(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    payload = "<script>alert(1)</script>\x01"
    resp = c.post(f"/projects/{p.id}/place-on-hold", {"reason": payload}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 422
    html = resp.content.decode()
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html
    assert _count_oob_regions(html) == 0


def test_project_resume_post_blank_reason_is_accepted(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/resume", {"reason": ""}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert _count_oob_regions(html) == 2
    p.refresh_from_db()
    assert p.status == ProjectStatus.ACTIVE
    audit = AuditEntry.objects.filter(project_id=p.id).order_by("-occurred_at").first()
    assert audit is not None
    assert audit.metadata == {"reason": ""}


def test_project_resume_post_too_long_reason_returns_422_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/resume", {"reason": "x" * 501}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 422
    html = resp.content.decode()
    assert "Reason must be 500 characters or fewer." in html
    assert "Couldn't submit transition" in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()


def test_project_resume_post_control_char_reason_returns_422_without_oob(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ON_HOLD)
    c = _login(admin_user)
    resp = c.post(f"/projects/{p.id}/resume", {"reason": "bad\x01reason"}, HTTP_HX_REQUEST="true")
    assert resp.status_code == 422
    html = resp.content.decode()
    assert "Reason contains invalid control characters." in html
    assert "Couldn't submit transition" in html
    assert _count_oob_regions(html) == 0
    assert not AuditEntry.objects.filter(project_id=p.id).exists()
