"""Cross-stack login-form snapshot test.

Renders GET /login, extracts the <form id="login-form"> block, strips the
CSRF hidden input, normalises whitespace, and asserts byte-equality against
fieldmark_shared/components/login-form.example.html.
"""

import re
from pathlib import Path

import pytest
from django.test import Client

_CANONICAL = (
    Path(__file__).resolve().parents[3]
    / "fieldmark_shared"
    / "components"
    / "login-form.example.html"
)

_CSRF_RE = re.compile(
    r'<input[^>]*name=["\']csrfmiddlewaretoken["\'][^>]*/?>',
    re.IGNORECASE,
)
_FORM_RE = re.compile(
    r'(<form\b[^>]*\bid=["\']login-form["\'][^>]*>.*?</form>)',
    re.DOTALL | re.IGNORECASE,
)


def _normalise(html: str) -> str:
    """Strip CSRF noise, collapse whitespace, trim each line."""
    html = _CSRF_RE.sub("", html)
    lines = [line.strip() for line in html.splitlines()]
    # Remove blank lines that appear between stripped lines
    collapsed = "\n".join(line for line in lines if line)
    return collapsed.strip()


def _extract_form(html: str) -> str:
    m = _FORM_RE.search(html)
    assert m, "Could not find <form id='login-form'> in rendered /login response"
    return m.group(1)


@pytest.mark.django_db
def test_login_form_matches_canonical_snapshot():
    client = Client()
    resp = client.get("/login")
    assert resp.status_code == 200

    rendered = resp.content.decode()
    form_block = _extract_form(rendered)
    actual = _normalise(form_block)

    canonical_raw = _CANONICAL.read_text(encoding="utf-8")
    expected = _normalise(canonical_raw)

    assert actual == expected, (
        "Django login form does not match canonical snapshot.\n"
        f"Expected:\n{expected}\n\nActual:\n{actual}"
    )
