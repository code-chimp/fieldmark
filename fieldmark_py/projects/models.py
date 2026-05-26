"""Django mappings for the `domain.project` aggregate (Story 2.1).

All models in this module are `Meta.managed = False` with schema-qualified
`db_table` values that force Postgres to interpret `domain` as the schema.
Django migrations never touch `domain.*` (ADR-014); the DDL at
`docker/postgres/init/010_domain_tables.sql` is binding.

Behavior methods (`place_on_hold`, `resume`, `close`, etc.) land in their
consuming stories (2.8, 2.12, Epic 6). This story is mapping only.
"""

from __future__ import annotations

from django.core.validators import MaxValueValidator, MinValueValidator
from django.db import models


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
