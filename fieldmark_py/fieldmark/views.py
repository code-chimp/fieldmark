from dataclasses import dataclass

from django.conf import settings
from django.contrib.auth import authenticate, login, logout
from django.contrib.auth.decorators import login_not_required
from django.http import (
    HttpRequest,
    HttpResponse,
    HttpResponseBadRequest,
    HttpResponseNotFound,
)
from django.shortcuts import redirect, render
from django.utils.http import url_has_allowed_host_and_scheme
from django.views.decorators.http import require_http_methods, require_POST

from fieldmark.roles import LABELS, Role, get_badge_token

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
    # Two-key selection: prefer canonical roles over unknown ones so a user with
    # both a canonical group (e.g. COMPLIANCE_OFFICER) and an unknown group
    # (e.g. ANALYST) always displays the correct canonical badge. A pure lexical
    # sort lets "ANALYST" outrank "COMPLIANCE_OFFICER" → badge-unknown regression.
    # Within each tier (canonical / unknown), names are sorted alphabetically for
    # stable selection. The warning branch in get_badge_token still fires when the
    # selected role_name is unknown (i.e., the user has no canonical group at all).
    all_names = list(request.user.groups.values_list("name", flat=True))
    canonical_set = {r.value for r in Role}
    canonical_sorted = sorted(n for n in all_names if n in canonical_set)
    unknown_sorted = sorted(n for n in all_names if n not in canonical_set)
    role_name = canonical_sorted[0] if canonical_sorted else (unknown_sorted[0] if unknown_sorted else "")
    try:
        role: Role | None = Role(role_name)
    except ValueError:
        role = None
    return render(
        request,
        "pages/home.html",
        {
            "role_label": LABELS[role] if role is not None else "",
            "role_badge_token": get_badge_token(role_name) if role_name else "unknown",
        },
    )


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
