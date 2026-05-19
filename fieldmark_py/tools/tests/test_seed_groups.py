"""Integration tests for the seed_groups management command."""

import pytest
from django.contrib.auth.models import Group
from django.core.management import call_command

CANONICAL = {"ADMIN", "COMPLIANCE_OFFICER", "INSPECTOR", "SITE_SUPERVISOR", "EXECUTIVE"}


@pytest.mark.django_db
def test_seed_groups_creates_five_canonical_groups():
    call_command("seed_groups")
    names = set(Group.objects.values_list("name", flat=True))
    assert CANONICAL <= names
    assert Group.objects.filter(name__in=CANONICAL).count() == 5


@pytest.mark.django_db
def test_seed_groups_is_idempotent():
    call_command("seed_groups")
    ids_first = {g.id for g in Group.objects.filter(name__in=CANONICAL)}
    call_command("seed_groups")
    ids_second = {g.id for g in Group.objects.filter(name__in=CANONICAL)}
    assert ids_first == ids_second
    assert Group.objects.filter(name__in=CANONICAL).count() == 5
