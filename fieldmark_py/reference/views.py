
from __future__ import annotations

import json
from dataclasses import dataclass

from django.core.exceptions import PermissionDenied
from django.http import HttpRequest, HttpResponse
from django.shortcuts import render

from fieldmark.roles import Role
from reference import queries


@dataclass(frozen=True)
class ComplianceRuleRow:
    code: str
    name: str
    description: str | None
    rule_kind: str
    parameters_json: str
    active: bool


def _is_admin(request: HttpRequest) -> bool:
    return request.user.is_authenticated and request.user.groups.filter(name=Role.ADMIN.value).exists()


def reference_index(request: HttpRequest) -> HttpResponse:
    if not _is_admin(request):
        raise PermissionDenied("You do not have permission to access this page.")

    return render(
        request,
        "reference/index.html",
        {
            "trade_types": queries.list_trade_types(),
            "violation_categories": queries.list_violation_categories(),
            "compliance_rules": [
                ComplianceRuleRow(
                    code=rule.code,
                    name=rule.name,
                    description=rule.description,
                    rule_kind=rule.rule_kind,
                    parameters_json=json.dumps(rule.parameters, separators=(",", ":")),
                    active=rule.active,
                )
                for rule in queries.list_compliance_rules()
            ],
        },
    )


def trade_types(request: HttpRequest) -> HttpResponse:
    if not _is_admin(request):
        raise PermissionDenied("You do not have permission to access this page.")
    return render(
        request,
        "reference/trade_types.html",
        {
            "trade_types": queries.list_trade_types(),
        },
    )


def violation_categories(request: HttpRequest) -> HttpResponse:
    if not _is_admin(request):
        raise PermissionDenied("You do not have permission to access this page.")
    return render(
        request,
        "reference/violation_categories.html",
        {
            "violation_categories": queries.list_violation_categories(),
        },
    )


def compliance_rules(request: HttpRequest) -> HttpResponse:
    if not _is_admin(request):
        raise PermissionDenied("You do not have permission to access this page.")
    return render(
        request,
        "reference/compliance_rules.html",
        {
            "compliance_rules": [
                ComplianceRuleRow(
                    code=rule.code,
                    name=rule.name,
                    description=rule.description,
                    rule_kind=rule.rule_kind,
                    parameters_json=json.dumps(rule.parameters, separators=(",", ":")),
                    active=rule.active,
                )
                for rule in queries.list_compliance_rules()
            ],
        },
    )
