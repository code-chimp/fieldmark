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
            lines.append(f"get {path}")


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
