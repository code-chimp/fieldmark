from dataclasses import dataclass
from datetime import UTC, datetime, timedelta

from django.conf import settings
from django.contrib.auth import authenticate, login, logout
from django.contrib.auth.decorators import login_not_required
from django.db import connection
from django.db import models
from django.http import (
    HttpRequest,
    HttpResponse,
    HttpResponseBadRequest,
    HttpResponseNotFound,
)
from django.shortcuts import redirect, render
from django.utils.http import url_has_allowed_host_and_scheme
from django.views.decorators.http import require_http_methods, require_POST

from fieldmark.authz import can, register_action
from fieldmark.roles import Role
from projects.models import Project, ProjectStatus

register_action("dashboard.view", Role.ADMIN, Role.COMPLIANCE_OFFICER, Role.INSPECTOR, Role.SITE_SUPERVISOR, Role.EXECUTIVE)

_ALLOWED_THEMES = {"system", "light", "dark"}


@dataclass
class LoginFieldError:
    username: str | None = None
    password: str | None = None
    general: str | None = None

    def has_any(self) -> bool:
        return bool(self.username or self.password or self.general)

    @property
    def field_error_count(self) -> int:
        return sum(1 for v in (self.username, self.password) if v is not None)


@login_not_required
@require_http_methods(["GET", "POST"])
def login_view(request: HttpRequest) -> HttpResponse:
    if request.user.is_authenticated:
        next_url = request.GET.get("next", "/")
        if url_has_allowed_host_and_scheme(next_url, allowed_hosts={request.get_host()}):
            return redirect(next_url)
        return redirect("/")

    errors = LoginFieldError()
    username = ""
    # Form posts the destination as "return_url"; Django's redirect appends it as "next".
    next_url = request.POST.get("return_url") or request.POST.get("next") or request.GET.get("next") or ""

    if request.method == "POST":
        username = (request.POST.get("username") or "").strip()
        password = request.POST.get("password") or ""
        if not username:
            errors.username = "Username is required."
        if not password:
            errors.password = "Password is required."
        if not errors.has_any():
            user = authenticate(request, username=username, password=password)
            if user is None:
                errors.general = "Invalid username or password."
                errors.username = ""
                errors.password = ""
            else:
                login(request, user)
                if url_has_allowed_host_and_scheme(next_url, allowed_hosts={request.get_host()}):
                    return redirect(next_url)
                return redirect("/")

    status = 422 if (request.method == "POST" and errors.has_any()) else 200
    return render(
        request,
        "_login.html",
        {"errors": errors, "username": username, "next": next_url},
        status=status,
    )


@login_not_required
@require_http_methods(["GET", "POST"])
def logout_view(request: HttpRequest) -> HttpResponse:
    logout(request)
    return redirect("/login")


def home(request: HttpRequest) -> HttpResponse:
    return redirect("/dashboard")


def dashboard(request: HttpRequest) -> HttpResponse:
    if not can(request.user, "dashboard.view"):
        return HttpResponse("You do not have permission to access this page.", status=403)

    portfolio_avg = Project.objects.exclude(status=ProjectStatus.CLOSED).aggregate(v=models.Avg("compliance_score"))["v"]
    portfolio_score = None if portfolio_avg is None else round(float(portfolio_avg))

    project_count = Project.objects.count()
    active_count = Project.objects.filter(status=ProjectStatus.ACTIVE).count()

    now_utc = datetime.now(UTC)
    week_start = datetime.combine((now_utc - timedelta(days=now_utc.weekday())).date(), datetime.min.time(), tzinfo=UTC)
    week_end = week_start + timedelta(days=7)

    with connection.cursor() as cur:
        cur.execute("SELECT count(*) FROM domain.violation")
        violation_count = cur.fetchone()[0]
        cur.execute("""SELECT severity, count(*) FROM domain.violation
            WHERE status IN ('Open','InProgress') AND due_at < now()
            GROUP BY severity""")
        severity_counts = {row[0]: row[1] for row in cur.fetchall()}
        overdue_total = sum(severity_counts.values())
        cur.execute("SELECT count(*) FROM domain.inspection")
        inspection_count = cur.fetchone()[0]
        cur.execute(
            """SELECT count(*) FROM domain.inspection
            WHERE scheduled_for >= %s
            AND scheduled_for < %s""",
            [week_start, week_end],
        )
        week_count = cur.fetchone()[0]

    overdue_violations = None if violation_count == 0 else overdue_total
    inspections_week = None if inspection_count == 0 else week_count
    breakdown_parts = []
    for label in ("Critical", "High", "Medium", "Low"):
        count = severity_counts.get(label, 0)
        if count > 0:
            breakdown_parts.append(f"{count} {label}")
    overdue_breakdown = "" if overdue_violations in (None, 0) else ", ".join(breakdown_parts)
    context = dashboard_context_from_raw(
        portfolio_score=portfolio_score,
        project_count=project_count,
        active_count=active_count,
        violation_count=violation_count,
        overdue_total=overdue_total,
        overdue_breakdown=overdue_breakdown,
        inspection_count=inspection_count,
        week_count=week_count,
    )
    return render(request, "dashboard/index.html", context)


def dashboard_context_from_raw(
    *,
    portfolio_score: int | None,
    project_count: int,
    active_count: int,
    violation_count: int,
    overdue_total: int,
    overdue_breakdown: str,
    inspection_count: int,
    week_count: int,
) -> dict[str, object]:
    active_projects = None if project_count == 0 else active_count
    overdue_violations = None if violation_count == 0 else overdue_total
    inspections_week = None if inspection_count == 0 else week_count
    return {
        "portfolio_score": portfolio_score,
        "overdue_violations": overdue_violations,
        "overdue_breakdown": "" if overdue_violations in (None, 0) else overdue_breakdown,
        "active_projects": active_projects,
        "inspections_week": inspections_week,
    }


def privacy(request):
    return render(request, "pages/privacy.html")


def compliance_tile(request):
    return render(request, "fragments/compliance_tile.html")


@login_not_required
@require_POST
def set_theme(request):
    value = request.POST.get("value", "")
    if value not in _ALLOWED_THEMES:
        return HttpResponseBadRequest()
    response = HttpResponse(status=204)
    response.set_cookie("fm_theme", value, max_age=31536000, path="/", samesite="Lax")
    response["HX-Trigger"] = "theme-changed"
    return response


@login_not_required
@require_http_methods(["GET"])
def entity_rail_fixture(request: HttpRequest) -> HttpResponse:
    """Debug-only fixture page for the EntityRail responsive-collapse Playwright test.
    Gated behind DEBUG=True; excluded from make parity via dump_routes __test__ prefix rule.
    """
    if not settings.DEBUG:
        return HttpResponseNotFound()
    return render(request, "debug/entity_rail_fixture.html")


@login_not_required
@require_http_methods(["GET"])
def tab_strip_fixture(request: HttpRequest) -> HttpResponse:
    """Debug-only fixture page for the TabStrip keyboard-navigation Playwright test (AC6 / Story 2.7).
    Gated behind DEBUG=True; excluded from make parity via dump_routes __test__ prefix rule.
    Variant is selected via ?variant=<name> query string (defaults to summary-active).
    """
    if not settings.DEBUG:
        return HttpResponseNotFound()
    variant = request.GET.get("variant", "summary-active")
    project_tabs = [
        {"id": "tab-summary", "label": "Summary", "hx_get": "/projects/__ID__/summary", "hx_target": "#project-detail-tab-content", "badge_count": None},
        {"id": "tab-inspections", "label": "Inspections", "hx_get": "/projects/__ID__/inspections", "hx_target": "#project-detail-tab-content", "badge_count": None},
        {"id": "tab-violations", "label": "Violations", "hx_get": "/projects/__ID__/violations", "hx_target": "#project-detail-tab-content", "badge_count": None},
        {"id": "tab-audit", "label": "Audit", "hx_get": "/projects/__ID__/audit", "hx_target": "#project-detail-tab-content", "badge_count": None},
    ]
    single_tab = [{"id": "tab-only", "label": "Only Tab", "hx_get": "/__tab__/only", "hx_target": "#__panel__", "badge_count": None}]
    return render(request, "debug/tab_strip_fixture.html", {
        "variant": variant,
        "project_tabs": project_tabs,
        "single_tab_tabs": single_tab,
    })
