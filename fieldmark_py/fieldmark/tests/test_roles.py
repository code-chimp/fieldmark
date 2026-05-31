"""Unit tests for roles.py — badge token resolution including unknown-token path (AC2.4 / Story 1.14)."""

import logging

import pytest

from fieldmark.roles import get_badge_token


def test_get_badge_token_known_roles():
    assert get_badge_token("ADMIN") == "danger"
    assert get_badge_token("COMPLIANCE_OFFICER") == "info"
    assert get_badge_token("INSPECTOR") == "warning"
    assert get_badge_token("SITE_SUPERVISOR") == "neutral"
    assert get_badge_token("EXECUTIVE") == "success"


def test_get_badge_token_unknown_returns_unknown_string(caplog):
    with caplog.at_level(logging.WARNING, logger="fieldmark.roles"):
        result = get_badge_token("UNKNOWN_ROLE")

    assert result == "unknown"
    assert any("Unknown role badge token" in r.message for r in caplog.records)
    assert any("UNKNOWN_ROLE" in r.message for r in caplog.records)


def test_get_badge_token_empty_string_returns_unknown_string(caplog):
    with caplog.at_level(logging.WARNING, logger="fieldmark.roles"):
        result = get_badge_token("")

    assert result == "unknown"


@pytest.mark.django_db
def test_home_page_no_role_redirects_to_dashboard():
    """Authenticated users land on the dashboard route via home redirect."""
    from django.contrib.auth.models import User
    from django.test import Client

    user = User.objects.create_user(username="test_no_role_1114", password="pass")
    client = Client()
    client.force_login(user)
    resp = client.get("/")

    assert resp.status_code == 302
    assert resp["Location"] == "/dashboard"
