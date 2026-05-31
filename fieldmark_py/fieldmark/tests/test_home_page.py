"""Integration tests for the Home route redirect behavior."""

import pytest
from django.contrib.auth.models import Group, User
from django.test import Client


@pytest.mark.django_db
def test_home_unauthenticated_redirects_to_login(db):
    client = Client()
    resp = client.get("/")
    assert resp.status_code == 302
    assert resp["Location"].startswith("/login")

@pytest.mark.django_db
def test_home_authenticated_admin_redirects_to_dashboard():
    group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(username="test_admin_home", password="pass")
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 302
    assert resp["Location"] == "/dashboard"


@pytest.mark.django_db
def test_home_authenticated_redirects_to_dashboard_for_non_admin():
    group, _ = Group.objects.get_or_create(name="INSPECTOR")
    user = User.objects.create_user(username="test_inspector_home", password="pass")
    user.groups.set([group])

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 302
    assert resp["Location"] == "/dashboard"


@pytest.mark.django_db
def test_home_authenticated_redirects_to_dashboard_for_user_without_role():
    user = User.objects.create_user(username="test_wordmark_home", password="pass")

    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 302
    assert resp["Location"] == "/dashboard"
