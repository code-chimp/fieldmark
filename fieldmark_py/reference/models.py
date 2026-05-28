
"""Read-only Django mappings for FieldMark reference data."""

from __future__ import annotations

import uuid

from django.db import models


class TradeType(models.Model):
    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    code = models.CharField(max_length=32, unique=True)
    name = models.CharField(max_length=120)
    description = models.TextField(null=True, blank=True)  # noqa: DJ001
    active = models.BooleanField(default=True)

    class Meta:
        managed = False
        db_table = 'domain"."trade_type'
        default_permissions = ()

    def __str__(self) -> str:
        return f"{self.code} ({self.name})"


class ViolationCategory(models.Model):
    class Severity(models.TextChoices):
        LOW = "Low", "Low"
        MEDIUM = "Medium", "Medium"
        HIGH = "High", "High"
        CRITICAL = "Critical", "Critical"

    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    code = models.CharField(max_length=32, unique=True)
    name = models.CharField(max_length=200)
    trade_type = models.UUIDField(null=True, db_column="trade_type_id")
    default_severity = models.CharField(max_length=16, choices=Severity.choices)
    description = models.TextField(null=True, blank=True)  # noqa: DJ001
    active = models.BooleanField(default=True)

    class Meta:
        managed = False
        db_table = 'domain"."violation_category'
        default_permissions = ()

    def __str__(self) -> str:
        return f"{self.code} ({self.name})"


class ComplianceRule(models.Model):
    class RuleKind(models.TextChoices):
        SCORING_PENALTY = "ScoringPenalty", "Scoring Penalty"
        CLOSURE_GATE = "ClosureGate", "Closure Gate"

    id = models.UUIDField(primary_key=True, default=uuid.uuid4)
    code = models.CharField(max_length=64, unique=True)
    name = models.CharField(max_length=200)
    description = models.TextField()
    rule_kind = models.CharField(max_length=32, choices=RuleKind.choices)
    parameters = models.JSONField()
    active = models.BooleanField(default=True)

    class Meta:
        managed = False
        db_table = 'domain"."compliance_rule'
        default_permissions = ()

    def __str__(self) -> str:
        return f"{self.code} ({self.name})"
