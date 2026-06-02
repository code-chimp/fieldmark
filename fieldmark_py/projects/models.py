"""Django mappings for the `domain.project` aggregate (Stories 2.1 / 2.8).

All models in this module are `Meta.managed = False` with schema-qualified
`db_table` values that force Postgres to interpret `domain` as the schema.
Django migrations never touch `domain.*` (ADR-014); the DDL at
`docker/postgres/init/010_domain_tables.sql` is binding.

See docs/reference/project-create-form-contract.md for the form contract.
"""

from __future__ import annotations

import uuid
from datetime import date
from typing import TYPE_CHECKING

from django.core.validators import MaxValueValidator, MinValueValidator
from django.db import models

from .errors import InvalidProjectTransition

if TYPE_CHECKING:
    from collections.abc import Sequence


class ProjectStatus(models.TextChoices):
    """Persisted strings match the DDL CHECK constraint on
    `domain.project.status` exactly — PascalCase per
    `010_domain_tables.sql:71`. The DDL is binding; the epic AC's
    SCREAMING_SNAKE_CASE note is superseded.
    """

    ACTIVE = "Active", "Active"
    ON_HOLD = "OnHold", "On Hold"
    CLOSED = "Closed", "Closed"


class Project(models.Model):
    id = models.UUIDField(primary_key=True)
    code = models.CharField(max_length=32, unique=True)
    name = models.CharField(max_length=200)
    # DJ001 suppressed: the DDL declares this column nullable; preserving
    # NULL vs empty-string distinction is required to round-trip faithfully.
    description = models.TextField(null=True, blank=True)  # noqa: DJ001
    status = models.CharField(max_length=16, choices=ProjectStatus.choices)
    start_date = models.DateField()
    target_completion_date = models.DateField(null=True, blank=True)
    actual_closed_at = models.DateTimeField(null=True, blank=True)
    compliance_score = models.IntegerField(
        validators=[MinValueValidator(0), MaxValueValidator(100)],
        default=100,
    )
    created_at = models.DateTimeField()
    updated_at = models.DateTimeField()

    class Meta:
        managed = False
        db_table = 'domain"."project'

    def __str__(self) -> str:
        return f"{self.code} ({self.name})"

    def can_place_on_hold(self) -> bool:
        return self.status == ProjectStatus.ACTIVE

    def can_resume(self) -> bool:
        return self.status == ProjectStatus.ON_HOLD

    def can_close(self) -> bool:
        # Status-only gate for Story 2.11; Epic 6 adds additional closure checks.
        return self.status == ProjectStatus.ACTIVE

    def place_on_hold(self, reason: str) -> None:
        _ = reason
        if self.status != ProjectStatus.ACTIVE:
            raise InvalidProjectTransition("Project is already on hold")
        self.status = ProjectStatus.ON_HOLD

    def resume(self, reason: str | None = None) -> None:
        _ = reason
        if self.status != ProjectStatus.ON_HOLD:
            raise InvalidProjectTransition("Project is not on hold")
        self.status = ProjectStatus.ACTIVE

    @classmethod
    def create(
        cls,
        *,
        code: str,
        name: str,
        description: str | None,
        start_date: date,
        target_completion_date: date | None,
        trade_scope_ids: Sequence[uuid.UUID],
        inspector_ids: Sequence[uuid.UUID],
    ) -> tuple[Project, list[ProjectTradeScope], list[ProjectInspector]]:
        """Factory method — call inside ``transaction.atomic()`` in the handler.

        Raises ``ValueError`` for invalid arguments. Request-level validation
        (lengths, allowlists, CSRF) is the form's job; this method enforces
        domain invariants.
        """
        code = (code or "").strip()
        name = (name or "").strip()
        description_value = (description or "").strip() or None

        if not code:
            raise ValueError("code is required")
        if not name:
            raise ValueError("name is required")
        if not trade_scope_ids:
            raise ValueError("at least one trade scope is required")
        if target_completion_date is not None and target_completion_date < start_date:
            raise ValueError("target_completion_date must be on or after start_date")

        project_id = uuid.uuid4()
        project = cls(
            id=project_id,
            code=code,
            name=name,
            description=description_value,
            status=ProjectStatus.ACTIVE,
            start_date=start_date,
            target_completion_date=target_completion_date,
            compliance_score=100,
        )

        scopes = [
            ProjectTradeScope(project_id=project_id, trade_type_id=tid)
            for tid in trade_scope_ids
        ]
        inspectors = [
            ProjectInspector(project_id=project_id, user_id=uid)
            for uid in inspector_ids
        ]
        return project, scopes, inspectors


class JobSite(models.Model):
    id = models.UUIDField(primary_key=True)
    # DO_NOTHING because the cascade is DDL-owned (ADR-014). Django must not
    # attempt to enforce or override the FK behavior.
    project = models.ForeignKey(
        Project, on_delete=models.DO_NOTHING, db_column="project_id", related_name="job_sites"
    )
    label = models.CharField(max_length=120)
    # DJ001 suppressed: DDL-declared nullable; see Project.description.
    address = models.CharField(max_length=300, null=True, blank=True)  # noqa: DJ001

    class Meta:
        managed = False
        db_table = 'domain"."job_site'

    def __str__(self) -> str:
        return self.label


class ProjectTradeScope(models.Model):
    pk = models.CompositePrimaryKey("project_id", "trade_type_id")
    project = models.ForeignKey(
        Project,
        on_delete=models.DO_NOTHING,
        db_column="project_id",
        related_name="trade_scopes",
    )
    # trade_type_id is an opaque UUID at this story; the TradeType model lands
    # in Story 2.3. No FK declared on the Django side; DDL owns the constraint.
    trade_type_id = models.UUIDField()

    class Meta:
        managed = False
        db_table = 'domain"."project_trade_scope'

    def __str__(self) -> str:
        return f"{self.project_id}/{self.trade_type_id}"


class ProjectInspector(models.Model):
    pk = models.CompositePrimaryKey("project_id", "user_id")
    project = models.ForeignKey(
        Project,
        on_delete=models.DO_NOTHING,
        db_column="project_id",
        related_name="inspector_assignments",
    )
    # ADR-012: opaque reference to a framework-local auth user. No FK to
    # django_auth from domain.* tables.
    user_id = models.UUIDField()

    class Meta:
        managed = False
        db_table = 'domain"."project_inspector'

    def __str__(self) -> str:
        return f"{self.project_id}/{self.user_id}"
