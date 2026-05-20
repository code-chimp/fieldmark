"""Cross-cutting tools models.

DevUserUuid maps Django auth_user.id → the canonical UUID from
docker/postgres/init/seed-uuids/dev-users.json. This lets domain.audit_entry
rows reference users by UUID across all three stacks (cross-stack audit
parity, ADR-012). Lives in django_auth schema because the side table is
framework-local, never referenced from domain.*.
"""

import uuid

from django.contrib.auth import get_user_model
from django.db import models


class DevUserUuid(models.Model):
    user = models.OneToOneField(get_user_model(), on_delete=models.CASCADE, related_name="dev_uuid")
    uuid = models.UUIDField(unique=True, default=uuid.uuid4, editable=False)

    class Meta:
        db_table = 'django_auth"."dev_user_uuid'
        verbose_name = "dev user UUID mapping"

    def __str__(self) -> str:
        return f"{self.user_id} → {self.uuid}"
