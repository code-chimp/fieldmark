from dataclasses import dataclass

from django.contrib.auth import authenticate, login, logout
from django.contrib.auth.decorators import login_not_required
from django.http import HttpRequest, HttpResponse, HttpResponseBadRequest
from django.shortcuts import redirect, render
from django.utils.http import url_has_allowed_host_and_scheme
from django.views.decorators.http import require_http_methods, require_POST

from fieldmark.roles import BADGE_TOKENS, LABELS, Role

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
    canonical = {r.value for r in Role}
    group_names = sorted(
        name for name in request.user.groups.values_list("name", flat=True) if name in canonical
    )
    role_name = group_names[0] if group_names else ""
    try:
        role: Role | None = Role(role_name)
    except ValueError:
        role = None
    return render(
        request,
        "pages/home.html",
        {
            "role_label": LABELS[role] if role is not None else "",
            "role_badge_token": BADGE_TOKENS[role] if role is not None else "neutral",
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
