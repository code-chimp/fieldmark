from django.contrib.auth.models import Group, User
from django.test import Client
import pytest


def _login_with_role(role: str) -> Client:
    group, _ = Group.objects.get_or_create(name=role)
    user = User.objects.create_user(username=f"{role.lower()}_dashboard", password="pass")
    user.groups.set([group])
    client = Client()
    client.force_login(user)
    return client


@pytest.mark.django_db
def test_dashboard_authenticated_admin_renders_200(db):
    client = _login_with_role("ADMIN")
    resp = client.get("/dashboard")
    assert resp.status_code == 200


@pytest.mark.django_db
def test_dashboard_unauthenticated_redirects_to_login(db):
    client = Client()
    resp = client.get("/dashboard")
    assert resp.status_code == 302
    assert resp["Location"].startswith("/login")


@pytest.mark.django_db
def test_dashboard_no_role_returns_403(db):
    user = User.objects.create_user(username="dashboard_norole", password="pass")
    client = Client()
    client.force_login(user)
    resp = client.get("/dashboard")
    assert resp.status_code == 403


@pytest.mark.django_db
def test_dashboard_renders_tile_ids_and_responsive_classes(db):
    client = _login_with_role("COMPLIANCE_OFFICER")
    resp = client.get("/dashboard")
    html = resp.content.decode()
    assert 'id="compliance-tile-portfolio"' in html
    assert 'id="overdue-violations-tile"' in html
    assert 'id="active-projects-tile"' in html
    assert 'id="inspections-week-tile"' in html
    assert 'id="compliance-tile-portfolio" role="status"' in html
    assert 'id="overdue-violations-tile" role="status"' in html
    assert 'id="active-projects-tile" role="status"' in html
    assert 'id="inspections-week-tile" role="status"' in html
    assert "grid-cols-1" in html
    assert "md:grid-cols-2" in html
    assert "xl:grid-cols-4" in html
