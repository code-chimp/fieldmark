"""Seed dev users from the shared UUID manifest.

Reads docker/postgres/init/seed-uuids/dev-users.json (UTF-8 JSON, schema in
docker/postgres/init/seed-uuids/dev-users.schema.json) and writes six users
into django_auth.auth_user, mapping each to its canonical UUID via the
tools.DevUserUuid side table. Role assignment uses Django Groups
(seeded by tools.management.commands.seed_groups in Story 1.8).

Idempotence: update_or_create + Group.set(...) — re-running with no
manifest changes produces zero net mutations. Side-table UUID approach
documented in Story 1.10 AC #4 (auth_user.id is BIGSERIAL; the canonical
UUID lives in django_auth.dev_user_uuid).
"""

from __future__ import annotations

import json
from pathlib import Path

from django.conf import settings
from django.contrib.auth import get_user_model
from django.contrib.auth.models import Group
from django.core.management.base import BaseCommand, CommandError
from django.db import transaction

from tools.models import DevUserUuid

User = get_user_model()


class Command(BaseCommand):
    help = "Seed dev users from docker/postgres/init/seed-uuids/dev-users.json (idempotent)."

    def handle(self, *args: object, **options: object) -> None:
        manifest_path: Path = (
            settings.BASE_DIR.parent
            / "docker"
            / "postgres"
            / "init"
            / "seed-uuids"
            / "dev-users.json"
        )
        if not manifest_path.exists():
            raise CommandError(f"seed_dev_users: manifest not found at {manifest_path}")

        data = json.loads(manifest_path.read_text(encoding="utf-8"))
        entries = data["users"]

        created_count = 0
        updated_count = 0

        with transaction.atomic():
            for entry in entries:
                first, *rest = entry["display_name"].split(" ", 1)
                last = rest[0] if rest else ""

                user, was_created = User.objects.update_or_create(
                    username=entry["username"],
                    defaults={
                        "first_name": first,
                        "last_name": last,
                        "email": f"{entry['username']}@fieldmark.local",
                        "is_active": True,
                    },
                )
                # set_password invokes Django's configured PASSWORD_HASHERS (PBKDF2 by
                # default). Always call save() after set_password.
                user.set_password(entry["password"])
                user.save()

                DevUserUuid.objects.update_or_create(
                    user=user,
                    defaults={"uuid": entry["id"]},
                )

                if entry["role"]:
                    try:
                        group = Group.objects.get(name=entry["role"])
                    except Group.DoesNotExist as exc:
                        raise CommandError(
                            f"seed_dev_users: Group '{entry['role']}' missing. "
                            f"Run `manage.py seed_groups` first (Story 1.8)."
                        ) from exc
                    user.groups.set([group])
                else:
                    user.groups.clear()

                if was_created:
                    created_count += 1
                    self.stdout.write(f"  + created {entry['username']} ({entry['role'] or 'no-role'})")
                else:
                    updated_count += 1
                    self.stdout.write(f"  ~ updated {entry['username']} ({entry['role'] or 'no-role'})")

        self.stdout.write(
            self.style.SUCCESS(
                f"seed_dev_users: {len(entries)} users ({created_count} created, {updated_count} updated)"
            )
        )
