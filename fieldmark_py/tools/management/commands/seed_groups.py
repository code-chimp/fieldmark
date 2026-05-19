"""
Management command: seed_groups

Seeds the five canonical conceptual-role Groups required by FieldMark's
authorization model (Architecture D7). Idempotent — safe to re-run.
"""

from django.contrib.auth.models import Group
from django.core.management.base import BaseCommand

CANONICAL_GROUPS = (
    "ADMIN",
    "COMPLIANCE_OFFICER",
    "INSPECTOR",
    "SITE_SUPERVISOR",
    "EXECUTIVE",
)


class Command(BaseCommand):
    help = "Seed canonical conceptual-role Groups (idempotent)"

    def handle(self, *args, **options):
        for name in CANONICAL_GROUPS:
            _, created = Group.objects.get_or_create(name=name)
            verb = "created" if created else "exists"
            self.stdout.write(f"{verb}: {name}")
