"""
Management command: seed_groups

Seeds the five canonical conceptual-role Groups required by FieldMark's
authorization model (Architecture D7). Idempotent — safe to re-run.
"""

from django.contrib.auth.models import Group
from django.core.management.base import BaseCommand

from fieldmark.roles import Role

CANONICAL_GROUPS = tuple(r.value for r in Role)


class Command(BaseCommand):
    help = "Seed canonical conceptual-role Groups (idempotent)"

    def handle(self, *args, **options):
        for name in CANONICAL_GROUPS:
            _, created = Group.objects.get_or_create(name=name)
            verb = "created" if created else "exists"
            self.stdout.write(f"{verb}: {name}")
