"""Django mapping for ``domain.audit_entry`` (Story 2.2).

``Meta.managed = False`` and the schema-qualified ``db_table`` mean Django
never CREATE/ALTERs this table (ADR-014); the DDL at
``docker/postgres/init/010_domain_tables.sql:190-211`` is binding.

Write path goes through ``audit.append.append_audit_entry`` so callers cannot
accidentally bypass the keyword-only contract.

The ``project`` FK is intentionally NOT declared as ``ForeignKey(Project)`` —
the audit app must stay free of an import dependency on ``projects.``. The
DB-level FK is DDL-owned.
"""

from __future__ import annotations

import uuid

from django.db import models
from django.db.models.functions import Now


class AuditEntry(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    # DDL has `DEFAULT now()`. `db_default=Now()` tells Django to omit the
    # column from the INSERT when no value is supplied, letting Postgres
    # assign the server timestamp (Django 5+).
    occurred_at = models.DateTimeField(db_default=Now())
    actor_id = models.UUIDField()
    action = models.CharField(max_length=64)
    entity_type = models.CharField(max_length=64)
    entity_id = models.UUIDField()
    project_id = models.UUIDField(null=True)
    # JSONField uses jsonb on Postgres via psycopg's native adapter.
    # Nullable (no payload) is distinct from `'null'::jsonb` — do not coalesce.
    before_state = models.JSONField(null=True)
    after_state = models.JSONField(null=True)
    metadata = models.JSONField(null=True)

    class Meta:
        managed = False
        db_table = 'domain"."audit_entry'
        # Suppress Django's automatic CRUD permission rows for this unmanaged
        # table — auth permissions for audit entries live in fieldmark.authz.
        default_permissions = ()

    def __str__(self) -> str:
        return f"{self.action}({self.entity_type}:{self.entity_id})"
