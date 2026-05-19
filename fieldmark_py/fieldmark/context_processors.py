_ALLOWED = {"system", "light", "dark"}
_CYCLE = {"system": "light", "light": "dark", "dark": "system"}


def theme(request):
    raw = request.COOKIES.get("fm_theme", "system")
    current = raw if raw in _ALLOWED else "system"
    return {
        "fm_theme": current,
        "fm_theme_next": _CYCLE[current],
        "fm_theme_resolved": current,
    }
