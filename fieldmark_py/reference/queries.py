"""Single read surface for reference-data consumers."""

from __future__ import annotations

from reference.models import ComplianceRule, TradeType, ViolationCategory


def list_trade_types() -> list[TradeType]:
    return list(TradeType.objects.order_by("code"))


def list_violation_categories() -> list[ViolationCategory]:
    return list(ViolationCategory.objects.order_by("code"))


def list_compliance_rules() -> list[ComplianceRule]:
    return list(ComplianceRule.objects.order_by("code"))
