"""Auth flow integration tests — login, logout, unauthenticated redirect."""

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client


@pytest.fixture()
def client() -> Client:
    return Client()


@pytest.fixture()
def dev_user(db) -> User:
    """A seeded user with a known password, SITE_SUPERVISOR role."""
    group, _ = Group.objects.get_or_create(name="SITE_SUPERVISOR")
    u = User.objects.create_user(username="pat_flow_test", password="FieldMark!2026")
    u.groups.set([group])
    return u


@pytest.mark.django_db
def test_get_business_route_unauthenticated_redirects_to_login(client):
    resp = client.get("/")
    assert resp.status_code == 302
    assert resp["Location"].startswith("/login")


@pytest.mark.django_db
def test_get_login_while_authenticated_redirects_home(dev_user, client):
    client.force_login(dev_user)
    resp = client.get("/login")
    assert resp.status_code == 302
    assert resp["Location"] == "/"


@pytest.mark.django_db
def test_post_login_valid_credentials_redirects_home(dev_user, client):
    resp = client.post("/login", {"username": "pat_flow_test", "password": "FieldMark!2026"})
    assert resp.status_code == 302
    assert resp["Location"] == "/"


@pytest.mark.django_db
def test_post_login_empty_username_returns_422(client):
    resp = client.post("/login", {"username": "", "password": "whatever"})
    assert resp.status_code == 422
    content = resp.content.decode()
    assert 'id="login-errors"' in content
    assert 'role="alert"' in content


@pytest.mark.django_db
def test_post_login_empty_password_returns_422(client):
    resp = client.post("/login", {"username": "someone", "password": ""})
    assert resp.status_code == 422
    content = resp.content.decode()
    assert 'id="login-errors"' in content


@pytest.mark.django_db
def test_post_login_wrong_password_returns_422_no_session(dev_user, client):
    resp = client.post("/login", {"username": "pat_flow_test", "password": "wrong"})
    assert resp.status_code == 422
    content = resp.content.decode()
    assert 'id="login-errors"' in content
    assert 'role="alert"' in content
    # No session cookie set on failed login.
    assert not client.session.get("_auth_user_id")


@pytest.mark.django_db
def test_post_login_with_return_url_redirects_to_destination(dev_user, client):
    resp = client.post(
        "/login",
        {"username": "pat_flow_test", "password": "FieldMark!2026", "return_url": "/compliance"},
    )
    assert resp.status_code == 302
    assert resp["Location"] == "/compliance"


@pytest.mark.django_db
def test_post_logout_clears_session_and_redirects(dev_user, client):
    client.force_login(dev_user)
    # Django's test client requires CSRF exemption or enforce_csrf_checks=False (default).
    resp = client.post("/logout")
    assert resp.status_code == 302
    assert resp["Location"] == "/login"
    # Subsequent business-route request must redirect to login.
    follow = client.get("/")
    assert follow.status_code == 302
    assert follow["Location"].startswith("/login")
