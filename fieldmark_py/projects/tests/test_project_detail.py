from __future__ import annotations

import re
import uuid
from datetime import date
from datetime import UTC, datetime, timedelta
from html.parser import HTMLParser
from pathlib import Path

import pytest
from django.contrib.auth.models import Group, User
from django.db import connection
from django.test import Client

from audit.models import AuditEntry
from fieldmark.authz import _reset_for_tests, register_action
from fieldmark.roles import Role
from fieldmark.tests.normalize_html import extract_variant, normalise_component
from projects.models import Project, ProjectStatus
from tools.models import DevUserUuid


pytestmark = [pytest.mark.django_db, pytest.mark.integration]

_CANONICAL = (
    Path(__file__).resolve().parents[3]
    / "docs"
    / "reference"
    / "fixtures"
    / "project-audit-log-canonical.html"
)


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


def _normalise_audit_log_html(html: str) -> str:
    html = re.sub(
        r"[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}",
        "00000000-0000-0000-0000-000000000000",
        html,
    )
    html = re.sub(r'datetime="[^"]+"', 'datetime="TIMESTAMP"', html)
    html = re.sub(r'title="[^"]+"', 'title="TIMESTAMP"', html)
    html = re.sub(r"before_occurred_at=[^\"&]+", "before_occurred_at=TIMESTAMP_ENCODED", html)
    html = re.sub(r"(<time[^>]*>)(.*?)(</time>)", r"\1RELATIVE_TIME\3", html)
    return normalise_component(html)


def _extract_audit_panel(html: str) -> str:
    class _PanelParser(HTMLParser):
        def __init__(self) -> None:
            super().__init__(convert_charrefs=False)
            self._depth = 0
            self._capture = False
            self.parts: list[str] = []

        def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
            attrs_dict = dict(attrs)
            if tag == "div" and attrs_dict.get("id") == "project-detail-tab-content" and not self._capture:
                self._capture = True
                self._depth = 1
                self.parts.append(self.get_starttag_text())
                return
            if self._capture:
                if tag == "div":
                    self._depth += 1
                self.parts.append(self.get_starttag_text())

        def handle_endtag(self, tag: str) -> None:
            if not self._capture:
                return
            self.parts.append(f"</{tag}>")
            if tag == "div":
                self._depth -= 1
                if self._depth == 0:
                    self._capture = False

        def handle_startendtag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
            if self._capture:
                self.parts.append(self.get_starttag_text())

        def handle_data(self, data: str) -> None:
            if self._capture:
                self.parts.append(data)

        def handle_entityref(self, name: str) -> None:
            if self._capture:
                self.parts.append(f"&{name};")

        def handle_charref(self, name: str) -> None:
            if self._capture:
                self.parts.append(f"&#{name};")

    parser = _PanelParser()
    parser.feed(html)
    return "".join(parser.parts)


def _extract_first_audit_row_and_load_more(html: str) -> str:
    class _ListItemParser(HTMLParser):
        def __init__(self) -> None:
            super().__init__(convert_charrefs=False)
            self._capture: str | None = None
            self._depth = 0
            self.parts: dict[str, list[str]] = {"row": [], "load_more": []}

        def _current(self) -> list[str] | None:
            return self.parts[self._capture] if self._capture else None

        def handle_starttag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
            attrs_dict = dict(attrs)
            classes = set((attrs_dict.get("class") or "").split())
            if tag == "li" and self._capture is None:
                if "audit-row" in classes:
                    self._capture = "row"
                    self._depth = 1
                    self.parts["row"].append(self.get_starttag_text())
                    return
                if attrs_dict.get("id") == "audit-log-load-more":
                    self._capture = "load_more"
                    self._depth = 1
                    self.parts["load_more"].append(self.get_starttag_text())
                    return
            current = self._current()
            if current is not None:
                if tag == "li":
                    self._depth += 1
                current.append(self.get_starttag_text())

        def handle_endtag(self, tag: str) -> None:
            current = self._current()
            if current is None:
                return
            current.append(f"</{tag}>")
            if tag == "li":
                self._depth -= 1
                if self._depth == 0:
                    self._capture = None

        def handle_startendtag(self, tag: str, attrs: list[tuple[str, str | None]]) -> None:
            current = self._current()
            if current is not None:
                current.append(self.get_starttag_text())

        def handle_data(self, data: str) -> None:
            current = self._current()
            if current is not None:
                current.append(data)

        def handle_entityref(self, name: str) -> None:
            current = self._current()
            if current is not None:
                current.append(f"&{name};")

        def handle_charref(self, name: str) -> None:
            current = self._current()
            if current is not None:
                current.append(f"&#{name};")

    parser = _ListItemParser()
    parser.feed(html)
    return f"{''.join(parser.parts['row'])}{''.join(parser.parts['load_more'])}"


def _create_audit_entry(
    project_id: uuid.UUID,
    *,
    action: str = "ProjectPlacedOnHold",
    actor_id: uuid.UUID | None = None,
    occurred_at: datetime | None = None,
    metadata: dict[str, object] | None = None,
) -> AuditEntry:
    return AuditEntry.objects.create(
        id=uuid.uuid4(),
        occurred_at=occurred_at or datetime.now(UTC),
        actor_id=actor_id or uuid.uuid4(),
        action=action,
        entity_type="Project",
        entity_id=project_id,
        project_id=project_id,
        before_state={"status": "Active"},
        after_state={"status": "OnHold"},
        metadata=metadata,
    )


def _create_actor_user(
    *,
    username: str,
    display_name: str | None = None,
) -> DevUserUuid:
    first_name = ""
    last_name = ""
    if display_name:
        parts = display_name.split(" ", 1)
        first_name = parts[0]
        last_name = parts[1] if len(parts) > 1 else ""
    user = User.objects.create_user(
        username=username,
        password="pass",
        first_name=first_name,
        last_name=last_name,
    )
    return DevUserUuid.objects.create(user=user)


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


def test_project_detail_tab_audit_renders_live_audit_log(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    _create_audit_entry(
        p.id,
        metadata={"reason": "<script>alert(1)</script>"},
        occurred_at=datetime(2026, 6, 3, 12, 0, tzinfo=UTC),
    )
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'id="audit-log"' in html
    assert 'aria-live="polite"' in html
    assert 'data-audit-action="ProjectPlacedOnHold"' in html
    assert 'Load more' not in html
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html


def test_project_detail_tab_audit_unknown_action_uses_badge_unknown(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    _create_audit_entry(p.id, action="ProjectReticulatedSpline")
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert "badge-unknown" in html


def test_project_detail_tab_audit_unresolvable_actor_renders_question_mark_fallback(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    _create_audit_entry(p.id, actor_id=uuid.uuid4())
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert '<span class="audit-row__initials">??</span>' in html


def test_project_detail_tab_audit_whitespace_username_renders_question_mark_fallback(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    actor = _create_actor_user(username="   ")
    _create_audit_entry(p.id, actor_id=actor.uuid)
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert '<span class="audit-row__initials">??</span>' in html


def test_project_detail_tab_audit_empty_username_renders_question_mark_fallback(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    actor = _create_actor_user(username="")
    _create_audit_entry(p.id, actor_id=actor.uuid)
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert '<span class="audit-row__initials">??</span>' in html


def test_project_detail_tab_audit_actor_display_xss_is_escaped(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    payload = "<script>alert(1)</script>"
    actor = _create_actor_user(username=f"actor_{uuid.uuid4().hex[:8]}", display_name=payload)
    _create_audit_entry(
        p.id,
        actor_id=actor.uuid,
        metadata={"reason": payload},
    )
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert "&lt;script&gt;alert(1)&lt;/script&gt;" in html
    assert "<script>alert(1)</script>" not in html


def test_project_detail_tab_audit_empty_panel_matches_canonical(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    expected = extract_variant(_CANONICAL.read_text(encoding="utf-8"), "panel-empty")
    assert _normalise_audit_log_html(_extract_audit_panel(html)) == expected


def test_project_detail_tab_audit_no_role_forbidden(no_role_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = _login(no_role_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 403
    assert resp.content.decode() == "You do not have permission to access this page."


def test_project_detail_tab_audit_empty_state(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'id="audit-log"' in html
    assert "No audit entries recorded for this project yet." in html


def test_project_audit_log_load_more_returns_next_page(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    base_time = datetime(2026, 6, 3, 15, 0, tzinfo=UTC)
    for i in range(101):
        _create_audit_entry(
            p.id,
            action="ProjectResumed" if i % 2 else "ProjectPlacedOnHold",
            occurred_at=base_time - timedelta(minutes=i),
            metadata={"reason": f"reason-{i}"},
        )

    c = _login(admin_user)
    first = c.get(f"/projects/{p.id}/tabs/audit", HTTP_HX_REQUEST="true")
    assert first.status_code == 200
    first_html = first.content.decode()
    assert first_html.count('class="audit-row"') == 100
    assert 'id="audit-log-load-more"' in first_html

    oldest = AuditEntry.objects.filter(project_id=p.id).order_by("-occurred_at", "-id")[99]
    cursor_ts = oldest.occurred_at.astimezone(UTC).isoformat().replace("+00:00", "Z")
    next_page = c.get(
        f"/projects/{p.id}/audit-log",
        {"before_occurred_at": cursor_ts, "before_id": str(oldest.id)},
        HTTP_HX_REQUEST="true",
    )
    assert next_page.status_code == 200
    next_html = next_page.content.decode()
    assert next_html.count('class="audit-row"') == 1
    assert "Load more" not in next_html


def test_project_audit_log_first_page_shape_matches_canonical(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    base_time = datetime(2026, 6, 3, 15, 0, tzinfo=UTC)
    for i in range(101):
        _create_audit_entry(
            p.id,
            occurred_at=base_time - timedelta(minutes=i),
            metadata={"reason": f"reason-{i}"},
        )
    c = _login(admin_user)
    resp = c.get(f"/projects/{p.id}/audit-log", HTTP_HX_REQUEST="true")
    assert resp.status_code == 200
    html = resp.content.decode()
    expected = extract_variant(_CANONICAL.read_text(encoding="utf-8"), "fragment-with-row-and-load-more")
    assert _normalise_audit_log_html(_extract_first_audit_row_and_load_more(html)) == expected


def test_project_audit_log_unauth_redirects_to_login():
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project()
    c = Client()
    resp = c.get(f"/projects/{p.id}/audit-log")
    assert resp.status_code in (301, 302)
    assert "/login" in resp["Location"]


def test_project_audit_log_invalid_cursor_returns_400(admin_user: User):
    if not _has_domain_table("project") or not _has_domain_table("audit_entry"):
        pytest.skip("domain tables not present on default test DB")
    p = _create_project()
    _create_audit_entry(p.id)
    c = _login(admin_user)
    resp = c.get(
        f"/projects/{p.id}/audit-log",
        {"before_occurred_at": "nope", "before_id": "bad"},
        HTTP_HX_REQUEST="true",
    )
    assert resp.status_code == 400
    assert resp.content.decode() == "Invalid cursor."


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


def test_project_place_on_hold_post_current_tab_audit_keeps_audit_panel_live(admin_user: User):
    if not _has_domain_table("project"):
        pytest.skip("domain.project not present on default test DB")
    p = _create_project(status=ProjectStatus.ACTIVE)
    c = _login(admin_user)
    resp = c.post(
        f"/projects/{p.id}/place-on-hold",
        {"reason": "Weather delay", "current_tab": "audit"},
        HTTP_HX_REQUEST="true",
    )
    assert resp.status_code == 200
    html = resp.content.decode()
    assert 'aria-labelledby="tab-audit"' in html
    assert 'id="audit-log"' in html
    assert 'id="tab-audit"' in html
    assert 'aria-selected="true"' in html
    assert 'hx-swap-oob="afterbegin:#audit-log"' not in html
    assert html.count('data-audit-action="ProjectPlacedOnHold"') == 1
    assert "No audit entries recorded for this project yet." not in html


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
