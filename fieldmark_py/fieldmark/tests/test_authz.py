"""Unit tests for fieldmark.authz.can."""

import pytest
from django.contrib.auth.models import Group, User

from fieldmark.authz import _reset_for_tests, can, register_action
from fieldmark.roles import Role


@pytest.fixture(autouse=True)
def reset_action_map():
    """Reset the module-level ActionRoleMap before and after each test."""
    _reset_for_tests()
    yield
    _reset_for_tests()


@pytest.mark.django_db
def test_can_anonymous_actor_returns_false():
    from django.contrib.auth.models import AnonymousUser

    register_action("test.allow_admin", Role.ADMIN)
    assert can(AnonymousUser(), "test.allow_admin") is False


@pytest.mark.django_db
def test_can_admin_actor_returns_true_for_admin_scoped_action():
    register_action("test.allow_admin", Role.ADMIN)
    group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(username="test_admin_can")
    user.groups.set([group])
    assert can(user, "test.allow_admin") is True


@pytest.mark.django_db
def test_can_non_admin_actor_returns_false_for_admin_scoped_action():
    register_action("test.allow_admin", Role.ADMIN)
    group, _ = Group.objects.get_or_create(name="SITE_SUPERVISOR")
    user = User.objects.create_user(username="test_supervisor_can")
    user.groups.set([group])
    assert can(user, "test.allow_admin") is False


@pytest.mark.django_db
def test_can_unknown_action_returns_false():
    admin_group, _ = Group.objects.get_or_create(name="ADMIN")
    user = User.objects.create_user(username="test_admin_unknown")
    user.groups.set([admin_group])
    assert can(user, "test.unmapped") is False
