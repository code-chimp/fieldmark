from django.http import HttpResponse, HttpResponseBadRequest
from django.shortcuts import render
from django.views.decorators.http import require_POST

_ALLOWED_THEMES = {"system", "light", "dark"}


def home(request):
    return render(request, "pages/home.html")


def privacy(request):
    return render(request, "pages/privacy.html")


def compliance_tile(request):
    return render(request, "fragments/compliance_tile.html")


@require_POST
def set_theme(request):
    value = request.POST.get("value", "")
    if value not in _ALLOWED_THEMES:
        return HttpResponseBadRequest()
    response = HttpResponse(status=204)
    response.set_cookie("fm_theme", value, max_age=31536000, path="/", samesite="Lax")
    response["HX-Trigger"] = "theme-changed"
    return response
