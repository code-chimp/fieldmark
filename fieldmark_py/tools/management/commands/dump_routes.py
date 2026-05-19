"""
Management command: dump_routes

Writes a normalized route inventory to stdout — one line per route,
`METHOD /path`, sorted, all lowercase. Consumed by tools/parity/dump-routes-django.sh.

Framework internals (admin, static assets) are excluded; only application
routes participate in parity checking.
"""

from django.core.management.base import BaseCommand
from django.urls import URLPattern, URLResolver, get_resolver


class Command(BaseCommand):
    help = "Dump application routes for cross-stack parity checking"

    def handle(self, *args, **options):
        lines: list[str] = []
        _collect(get_resolver(), "", lines)
        for line in sorted(set(lines)):
            self.stdout.write(line)


def _collect(resolver: URLResolver, prefix: str, lines: list[str]) -> None:
    for pattern in resolver.url_patterns:
        if isinstance(pattern, URLResolver):
            _collect(pattern, prefix + _strip_regex(str(pattern.pattern)), lines)
        elif isinstance(pattern, URLPattern):
            raw = prefix + _strip_regex(str(pattern.pattern))
            path = _normalize(raw)
            if path is None:
                continue
            for method in _methods_for(pattern.callback):
                lines.append(f"{method} {path}")


_HTTP_METHODS = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}


def _methods_for(callback) -> list[str]:
    # Class-based views expose http_method_names via the view_class attribute.
    view_class = getattr(callback, "view_class", None)
    if view_class is not None:
        return [m for m in view_class.http_method_names if hasattr(view_class, m)]

    # Function-based views decorated with require_http_methods / require_POST.
    # Django <= 5 set request_method_list as a direct attribute on the wrapper.
    methods = getattr(callback, "request_method_list", None)
    if methods is not None:
        return [m.lower() for m in methods]

    # Django 6+ stores it as a closure variable.  Rather than relying on the
    # specific variable name or its index in co_freevars (both are internal
    # implementation details that can change), walk every closure cell and
    # look for a non-empty collection whose items are all known HTTP method
    # strings.  This is version-agnostic and survives decorator refactors.
    for cell in getattr(callback, "__closure__", None) or []:
        try:
            v = cell.cell_contents
        except ValueError:
            continue
        if (
            isinstance(v, (list, tuple))
            and v
            and all(isinstance(m, str) and m.upper() in _HTTP_METHODS for m in v)
        ):
            return [m.lower() for m in v]

    # Undecorated function-based view: assume GET (all current page renders).
    return ["get"]


def _strip_regex(raw: str) -> str:
    """Remove regex anchors ^ and $ from a URL pattern string."""
    return raw.lstrip("^").rstrip("$")


def _normalize(raw: str) -> str | None:
    """
    Normalize a raw URL path to the canonical parity format:
    - Leading slash
    - Lowercase
    - No trailing slash (except root)
    - Returns None for excluded paths
    """
    path = "/" + raw.lstrip("/")
    path = path.lower()
    if len(path) > 1:
        path = path.rstrip("/")

    # Exclude framework internals.
    excluded_prefixes = ("/admin",)
    if any(path == p or path.startswith(p + "/") for p in excluded_prefixes):
        return None

    return path
