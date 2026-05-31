"""Grid views — POST /grid/projects AG Grid SSRM endpoint.

Read-only query — no transaction, no audit entry, no state change.
CSRF: @csrf_exempt because the AG Grid datasource does not send csrfmiddlewaretoken
(read-only endpoint; cross-stack symmetry matches Go's no-CSRF posture).

See docs/reference/ag-grid-ssrm-contract.md
"""

from __future__ import annotations

import json

from django.http import HttpRequest, JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_POST

from fieldmark.authz import can, register_action
from fieldmark.roles import Role
from projects.models import Project

from .ssrm import SsrmError, parse_ssrm_request

# Register project.read for all five roles (portfolio list visible to any authenticated user).
# Rationale: docs/reference/ag-grid-ssrm-contract.md §"Per-Stack Native Implementations".
register_action(
    "project.read",
    Role.ADMIN,
    Role.COMPLIANCE_OFFICER,
    Role.INSPECTOR,
    Role.SITE_SUPERVISOR,
    Role.EXECUTIVE,
)

# Canonical projected columns (NFR6 — manually projected, no AutoMapper).
# See docs/reference/ag-grid-ssrm-contract.md §Row Projection Rules.
_PROJECTION = (
    "id",
    "code",
    "name",
    "status",
    "compliance_score",
    "start_date",
    "target_completion_date",
)


@require_POST
@csrf_exempt  # Read-only; AG Grid datasource does not send CSRF tokens.
def grid_projects(request: HttpRequest) -> JsonResponse:
    """POST /grid/projects — returns {rows, lastRow} per the SSRM contract."""
    if not request.user.is_authenticated:
        return JsonResponse({"error": "forbidden"}, status=403)
    if not can(request.user, "project.read"):
        return JsonResponse({"error": "forbidden"}, status=403)

    try:
        data = json.loads(request.body)
    except (json.JSONDecodeError, UnicodeDecodeError):
        return JsonResponse({"error": "invalid request body"}, status=400)

    try:
        parsed = parse_ssrm_request(data)
    except SsrmError as exc:
        return JsonResponse({"error": str(exc)}, status=400)

    if parsed.match_nothing:
        return JsonResponse({"rows": [], "lastRow": 0})

    qs = Project.objects.filter(parsed.q_filter)
    total = qs.count()
    rows_qs = (
        qs.order_by(*parsed.order_fields)
        .values(*_PROJECTION)[parsed.offset : parsed.offset + parsed.limit]
    )

    rows = []
    for row in rows_qs:
        rows.append({
            "id": str(row["id"]),
            "code": row["code"],
            "name": row["name"],
            "status": row["status"],
            "compliance_score": row["compliance_score"],
            "start_date": row["start_date"].isoformat() if row["start_date"] else None,
            "target_completion_date": (
                row["target_completion_date"].isoformat()
                if row["target_completion_date"]
                else None
            ),
        })

    return JsonResponse({"rows": rows, "lastRow": total})
