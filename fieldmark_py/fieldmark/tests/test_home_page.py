"""Integration tests for the Home page (AC #1, #4, #5, #6, #7, #8, Story 1.13)."""

import os
import re
import shutil
import subprocess
import tempfile

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client

from .normalize_html import normalise_for_parity

_BLOCK_RE = re.compile(r"(?s)<(header|main)[^>]*>.*?</\1>")


@pytest.mark.django_db
def test_home_unauthenticated_redirects_to_login():
    client = Client()
    resp = client.get("/")
    assert resp.status_code == 302
    assert resp["Location"].startswith("/login")


@pytest.mark.django_db
def test_home_authenticated_admin_renders_role_badge_and_placeholder():
    group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(username="test_admin_home", password="pass")
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 200
    content = resp.content.decode()
    assert "<h1>FieldMark</h1>" in content
    assert "badge-danger" in content
    assert ">Admin<" in content
    assert "Your projects will appear here." in content


@pytest.mark.django_db
def test_home_authenticated_renders_avatar_menu():
    group, _ = Group.objects.get_or_create(name="INSPECTOR")
    user = User.objects.create_user(username="test_inspector_home", password="pass")
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 200
    content = resp.content.decode()
    assert "avatar-menu-wrapper" in content
    assert "avatar-menu-dropdown" in content
    assert 'href="/logout"' in content


@pytest.mark.django_db
def test_home_authenticated_renders_wordmark_in_nav():
    user = User.objects.create_user(username="test_wordmark_home", password="pass")

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 200
    content = resp.content.decode()
    assert 'class="fm-wordmark"' in content
    assert 'aria-label="FieldMark home"' in content


@pytest.mark.django_db
def test_home_authenticated_passes_axe_core():
    """AC #6: zero WCAG 2.1 AA violations under axe-core.

    Skips gracefully when npx is not on PATH.
    Manual recipe: npx @axe-core/cli http://localhost:8000/ (authenticated session required).
    """
    if shutil.which("npx") is None:
        pytest.skip("npx not on PATH; install Node.js to enable axe-core WCAG 2.1 AA gate")

    group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(username="axe_admin", password="pass")
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")
    assert resp.status_code == 200

    with tempfile.NamedTemporaryFile(suffix=".html", delete=False, mode="wb") as f:
        f.write(resp.content)
        tmp_path = f.name

    result = subprocess.run(["npx", "@axe-core/cli", "file://" + tmp_path], capture_output=True, text=True)
    assert result.returncode == 0, f"axe-core found WCAG 2.1 AA violations:\n{result.stdout}\n{result.stderr}"


@pytest.mark.django_db
def test_home_tab_order_matches_contract():
    """AC #7: DOM-order check for the required focus sequence.

    Verifies skip-link → wordmark → theme-toggle → avatar button → sign-out appear in DOM order.
    DOM order is the primary determinant of tab order when no tabindex attributes override it.
    Full runtime focus-order verification (CSS, tabindex) still requires pytest-playwright (Epic 7).
    Manual recipe: open http://localhost:8000/, Tab 5 times, verify sequence above.
    """
    group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(username="tab_order_admin", password="pass")
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")
    assert resp.status_code == 200
    html = resp.content.decode()

    markers = [
        ("skip-link", 'class="skip-link"'),
        ("fm-wordmark", 'class="fm-wordmark"'),
        ("theme-toggle", 'class="theme-toggle"'),
        ("avatar-menu button", 'class="avatar-menu"'),
        ("sign-out anchor", 'href="/logout"'),
    ]
    indices = []
    for name, text in markers:
        idx = html.find(text)
        assert idx != -1, f"expected {name!r} to be present in HTML"
        indices.append(idx)

    for i in range(1, len(indices)):
        assert indices[i - 1] < indices[i], (
            f"DOM order violation: {markers[i-1][0]!r} (idx {indices[i-1]}) must precede "
            f"{markers[i][0]!r} (idx {indices[i]})"
        )


@pytest.mark.django_db
def test_home_chrome_matches_parity_snapshot():
    """AC #8: normalized chrome byte-matches the committed cross-stack snapshot."""
    snapshot_path = os.path.join(
        os.path.dirname(__file__),
        "..",
        "..",
        "..",
        "_bmad-output",
        "implementation-artifacts",
        "_parity-snapshots",
        "home-chrome.normalized.html",
    )
    snapshot_path = os.path.normpath(snapshot_path)
    if not os.path.exists(snapshot_path):
        pytest.skip(f"parity snapshot not found at {snapshot_path}")

    snapshot = normalise_for_parity(open(snapshot_path).read())

    group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(
        username="aisha_parity", first_name="Aisha", last_name="Patel", password="pass"
    )
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")
    assert resp.status_code == 200
    html = resp.content.decode()

    full_blocks = [m.group(0) for m in _BLOCK_RE.finditer(html)]
    assert len(full_blocks) >= 2, f"expected <header> and <main> blocks, got {len(full_blocks)}"

    normalized = normalise_for_parity("\n".join(full_blocks))
    assert normalized == snapshot, (
        f"home chrome diverges from parity snapshot.\nGot:\n{normalized}\n\nWant:\n{snapshot}"
    )


@pytest.mark.django_db
def test_home_authenticated_no_role_renders_neutral_badge():
    user = User.objects.create_user(username="test_norole_home", password="pass")

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 200
    content = resp.content.decode()
    # No role → badge token defaults to neutral; badge span present but empty label
    assert "badge-neutral" in content
