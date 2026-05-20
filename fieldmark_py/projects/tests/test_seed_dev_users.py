import pytest
from django.contrib.auth import get_user_model
from django.contrib.auth.models import Group
from django.core.management import call_command

from tools.models import DevUserUuid

User = get_user_model()


@pytest.fixture
def groups_seeded(db: None) -> None:
    for name in ("ADMIN", "COMPLIANCE_OFFICER", "INSPECTOR", "SITE_SUPERVISOR", "EXECUTIVE"):
        Group.objects.get_or_create(name=name)


@pytest.mark.django_db
def test_seed_dev_users_creates_six_users(groups_seeded: None) -> None:
    call_command("seed_dev_users")
    assert User.objects.count() == 6
    assert DevUserUuid.objects.count() == 6
    assert User.objects.get(username="marisol").groups.filter(name="COMPLIANCE_OFFICER").exists()
    assert User.objects.get(username="testuser").groups.count() == 0


@pytest.mark.django_db
def test_seed_dev_users_is_idempotent(groups_seeded: None) -> None:
    call_command("seed_dev_users")
    baseline_user_ids = set(User.objects.values_list("id", flat=True))
    baseline_uuid_ids = set(DevUserUuid.objects.values_list("user_id", flat=True))

    call_command("seed_dev_users")

    assert set(User.objects.values_list("id", flat=True)) == baseline_user_ids
    assert set(DevUserUuid.objects.values_list("user_id", flat=True)) == baseline_uuid_ids
