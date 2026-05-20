"""Tests for fieldmark.authn.current_actor."""

import uuid

import pytest
from django.contrib.auth.models import AnonymousUser, Group, User
from django.test import RequestFactory

from fieldmark.authn import ANONYMOUS, current_actor
from tools.models import DevUserUuid


@pytest.mark.django_db
def test_anonymous_request_returns_anonymous():
    request = RequestFactory().get("/")
    request.user = AnonymousUser()
    assert current_actor(request) == ANONYMOUS


@pytest.mark.django_db
def test_authenticated_user_returns_uuid_username_and_roles():
    group = Group.objects.create(name="ADMIN")
    u = User.objects.create(username="aisha_authn_test")
    u.groups.set([group])
    canonical = uuid.uuid4()
    DevUserUuid.objects.create(user_id=u.pk, uuid=canonical)

    request = RequestFactory().get("/")
    request.user = u

    actor = current_actor(request)
    assert actor.id == canonical
    assert actor.username == "aisha_authn_test"
    assert actor.roles == ("ADMIN",)
