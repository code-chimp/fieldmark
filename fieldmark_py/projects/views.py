"""Project views — Story 2.8 (create) + Story 2.9 (list).

See docs/reference/project-create-form-contract.md for the create form contract.
See docs/reference/ag-grid-ssrm-contract.md for the list page + grid contract.
"""

from __future__ import annotations

import re
import uuid
import json
import logging
from datetime import UTC, datetime

from django.contrib.auth import get_user_model
from django.contrib.auth.decorators import login_required
from django.db import IntegrityError, transaction
from django.http import HttpRequest, HttpResponse, HttpResponseForbidden
from django.shortcuts import redirect, render
from django.views.decorators.http import require_GET, require_http_methods, require_POST

from audit.actions import AuditAction
from audit.append import append_audit_entry
from fieldmark.authz import can, register_action
from fieldmark.roles import Role
from reference.models import TradeType
from tools.models import DevUserUuid

from .errors import InvalidProjectTransition
from .forms import ProjectCreateForm
from .models import Project, ProjectInspector, ProjectTradeScope

User = get_user_model()

# Register the project.create action for ADMIN role (Story 2.8).
register_action("project.create", Role.ADMIN)
# project.read registered in grid/views.py at import time (Story 2.9).
register_action("project.place_on_hold", Role.ADMIN)
register_action("project.resume", Role.ADMIN)
register_action("project.close", Role.ADMIN)

_REASON_MAX_LENGTH = 500
_CONTROL_CHAR_PATTERN = re.compile(r"[\x00-\x1F\x7F]")


def _get_reference_data():
    """Return (trade_type_choices, inspector_choices) for the form."""
    trade_types = TradeType.objects.filter(active=True).order_by("code")
    trade_choices = [(str(t.id), t.name) for t in trade_types]

    inspectors = (
        User.objects.filter(groups__name="INSPECTOR", is_active=True)
        .prefetch_related("dev_uuid")
        .order_by("username")
    )
    inspector_choices = []
    for u in inspectors:
        try:
            canonical_uuid = str(u.dev_uuid.uuid)
        except DevUserUuid.DoesNotExist:
            continue
        display = getattr(u, "get_full_name", lambda: "")() or u.username
        inspector_choices.append((canonical_uuid, display))

    return trade_choices, inspector_choices


@require_GET
@login_required
def project_create_get(request: HttpRequest) -> HttpResponse:
    """GET /projects/new — renders the empty create form."""
    if not can(request.user, "project.create"):
        return HttpResponseForbidden("You do not have permission to access this page.")

    trade_choices, inspector_choices = _get_reference_data()
    form = ProjectCreateForm(
        trade_type_choices=trade_choices,
        inspector_choices=inspector_choices,
    )
    return render(
        request,
        "projects/create.html",
        {"form": form, "trade_choices": trade_choices, "inspector_choices": inspector_choices},
    )


@require_POST
@login_required
def project_create_post(request: HttpRequest) -> HttpResponse:
    """POST /projects/ — creates the project or re-renders the form with errors."""
    if not can(request.user, "project.create"):
        return HttpResponseForbidden("You do not have permission to access this page.")

    trade_choices, inspector_choices = _get_reference_data()
    form = ProjectCreateForm(
        request.POST,
        trade_type_choices=trade_choices,
        inspector_choices=inspector_choices,
    )

    if not form.is_valid():
        return _render_422(request, form, trade_choices, inspector_choices)

    code = form.cleaned_data["code"]
    name = form.cleaned_data["name"]
    description = form.cleaned_data.get("description")
    start_date = form.cleaned_data["start_date"]
    target_completion_date = form.cleaned_data.get("target_completion_date")
    # Deduplicate preserving order — duplicate selections are semantically
    # redundant and would cause a composite-PK 23505 on project_trade_scope /
    # project_inspector if not removed here.
    trade_scope_ids: list[uuid.UUID] = list(dict.fromkeys(form.cleaned_data["trade_scope_ids"]))
    inspector_ids: list[uuid.UUID] = list(dict.fromkeys(form.cleaned_data["inspector_ids"]))

    # Build the after_state JSON snapshot (alphabetical keys, sorted UUID lists).
    sorted_trade_ids = sorted(str(tid) for tid in trade_scope_ids)
    sorted_inspector_ids = sorted(str(uid) for uid in inspector_ids)
    after_state = {
        "code": code,
        "compliance_score": 100,
        "description": description,
        "inspector_ids": sorted_inspector_ids,
        "name": name,
        "start_date": start_date.isoformat(),
        "status": "Active",
        "target_completion_date": target_completion_date.isoformat() if target_completion_date else None,
        "trade_scope_ids": sorted_trade_ids,
    }

    try:
        with transaction.atomic():
            # Reference validation inside the transaction so the reads share the
            # same MVCC snapshot as the writes (AC3 §single-transaction requirement).
            # 422 returns exit the with-block without an exception → Django commits
            # (no-op since no writes occurred), then we return the 422 response.
            valid_trade_ids = set(
                TradeType.objects.filter(active=True).values_list("id", flat=True)
            )
            invalid_trades = [tid for tid in trade_scope_ids if tid not in valid_trade_ids]
            if invalid_trades:
                form.add_error(
                    "trade_scope_ids",
                    "One or more selected trade types are no longer available. Please reselect.",
                )
                return _render_422(request, form, trade_choices, inspector_choices)

            if inspector_ids:
                # Django platform note: auth_user.pk is an integer; canonical cross-stack
                # UUIDs are stored in the DevUserUuid side table. Validation MUST go through
                # DevUserUuid — this is the Django-specific equivalent of .NET's
                # IdentityUser<Guid>.Id and Go's fiber_auth.users.id. Users without a
                # DevUserUuid row cannot appear in the form selector and therefore cannot
                # submit a valid inspector UUID through normal flow.
                valid_inspector_uuids = set(
                    DevUserUuid.objects.filter(
                        user__groups__name="INSPECTOR", user__is_active=True
                    ).values_list("uuid", flat=True)
                )
                invalid_inspectors = [uid for uid in inspector_ids if uid not in valid_inspector_uuids]
                if invalid_inspectors:
                    form.add_error(
                        "inspector_ids",
                        "One or more selected inspectors are no longer available. Please reselect.",
                    )
                    return _render_422(request, form, trade_choices, inspector_choices)

            project, scopes, inspectors = Project.create(
                code=code,
                name=name,
                description=description,
                start_date=start_date,
                target_completion_date=target_completion_date,
                trade_scope_ids=trade_scope_ids,
                inspector_ids=inspector_ids,
            )
            project.save()
            ProjectTradeScope.objects.bulk_create(scopes)
            if inspectors:
                ProjectInspector.objects.bulk_create(inspectors)

            # actor_id is the canonical cross-stack UUID via the dev_uuid side table.
            # Guard against missing dev_uuid row (should not happen for seeded users,
            # but avoids a 500 for users created outside the seed runner). Log an
            # error so operators can investigate. Use uuid5(NAMESPACE_DNS, user-PK)
            # as a deterministic non-nil fallback: entries from the same user still
            # share the same synthetic UUID, preserving per-user traceability in
            # audit logs without collapsing all unknown actors to the same all-zeros.
            try:
                actor_id = request.user.dev_uuid.uuid  # type: ignore[union-attr]
            except Exception:  # noqa: BLE001
                import logging
                logging.getLogger(__name__).error(
                    "projects.create: no dev_uuid for user_id=%s; "
                    "using synthetic actor_id — seed runner may need to be re-run",
                    request.user.pk,
                )
                actor_id = uuid.uuid5(
                    uuid.NAMESPACE_DNS,
                    f"django-user-{request.user.pk}",
                )

            append_audit_entry(
                actor_id=actor_id,
                action=AuditAction.PROJECT_CREATED,
                entity_type="Project",
                entity_id=project.id,
                project_id=project.id,
                before_state=None,
                after_state=after_state,
            )
    except IntegrityError as exc:
        # Use structured constraint inspection via the underlying psycopg exception.
        # psycopg3 surfaces pgcode and constraint_name on the DiagnosticsObject,
        # which is more robust than string-matching `str(exc)`.
        cause = getattr(exc, "__cause__", None)
        diag = getattr(cause, "diag", None)
        is_code_collision = (
            getattr(diag, "sqlstate", None) == "23505"
            and getattr(diag, "constraint_name", None) == "project_code_key"
        )
        if is_code_collision:
            form.add_error("code", "A project with this code already exists.")
            return _render_422(request, form, trade_choices, inspector_choices)
        raise

    is_htmx = request.headers.get("HX-Request") == "true"
    redirect_url = f"/projects/{project.id}"
    if is_htmx:
        response = HttpResponse("", status=200)
        response["HX-Redirect"] = redirect_url
        return response

    response = HttpResponse("", status=303)
    response["Location"] = redirect_url
    return response


def _project_tabs(project_id: uuid.UUID) -> list[dict[str, str]]:
    base = f"/projects/{project_id}/tabs"
    return [
        {"id": "tab-summary", "label": "Summary", "hx_get": f"{base}/summary", "hx_target": "#project-detail-tab-content", "badge_count": None},
        {"id": "tab-inspections", "label": "Inspections", "hx_get": f"{base}/inspections", "hx_target": "#project-detail-tab-content", "badge_count": None},
        {"id": "tab-violations", "label": "Violations", "hx_get": f"{base}/violations", "hx_target": "#project-detail-tab-content", "badge_count": None},
        {"id": "tab-audit", "label": "Audit", "hx_get": f"{base}/audit", "hx_target": "#project-detail-tab-content", "badge_count": None},
    ]


def _actor_display_map(user_ids: list[uuid.UUID]) -> dict[uuid.UUID, str]:
    if not user_ids:
        return {}
    rows = (
        DevUserUuid.objects.filter(uuid__in=user_ids)
        .select_related("user")
    )
    out: dict[uuid.UUID, str] = {}
    for row in rows:
        full = row.user.get_full_name() if hasattr(row.user, "get_full_name") else ""
        out[row.uuid] = full or row.user.username
    return out


def _build_project_detail_context(request: HttpRequest, project: Project) -> dict[str, object]:
    scope_ids = list(project.trade_scopes.values_list("trade_type_id", flat=True))
    trade_rows = TradeType.objects.filter(id__in=scope_ids)
    trade_by_id = {t.id: t for t in trade_rows}
    trade_names = [
        trade_by_id[sid].name if trade_by_id[sid].active else f"{trade_by_id[sid].name} (inactive)"
        for sid in scope_ids
        if sid in trade_by_id
    ]

    inspector_ids = list(project.inspector_assignments.values_list("user_id", flat=True))
    display_map = _actor_display_map(inspector_ids)
    inspector_names = [display_map[iid] for iid in inspector_ids if iid in display_map]

    can_hold = can(request.user, "project.place_on_hold")
    can_resume = can(request.user, "project.resume")
    can_close = can(request.user, "project.close")

    return {
        "project": project,
        "trade_names": trade_names,
        "inspector_names": inspector_names,
        "tabs": _project_tabs(project.id),
        "status_entity": "Project",
        "status_value": project.status,
        "can_place_on_hold": can_hold,
        "can_resume": can_resume,
        "can_close": can_close,
        "state_can_place_on_hold": project.can_place_on_hold(),
        "state_can_resume": project.can_resume(),
        "state_can_close": project.can_close(),
    }


def _load_project_or_404(id: uuid.UUID) -> Project | None:
    return (
        Project.objects
        .filter(pk=id)
        .prefetch_related("trade_scopes", "inspector_assignments")
        .first()
    )


@require_GET
@login_required
def project_detail(request: HttpRequest, id: uuid.UUID) -> HttpResponse:
    if not can(request.user, "project.read"):
        return HttpResponseForbidden("You do not have permission to access this page.")
    project = _load_project_or_404(id)
    if project is None:
        return HttpResponse("Not Found.", status=404)
    context = _build_project_detail_context(request, project)
    is_htmx = request.headers.get("HX-Request") == "true"
    template = "projects/_detail_body.html" if is_htmx else "projects/detail.html"
    return render(request, template, context)


def _render_tab_response(request: HttpRequest, id: uuid.UUID, active_index: int, panel_template: str) -> HttpResponse:
    if not can(request.user, "project.read"):
        return HttpResponseForbidden("You do not have permission to access this page.")
    if request.headers.get("HX-Request") != "true":
        return redirect(f"/projects/{id}")
    project = _load_project_or_404(id)
    if project is None:
        return HttpResponse("Not Found.", status=404)
    context = _build_project_detail_context(request, project)
    context["active_index"] = active_index
    context["panel_template"] = panel_template
    return render(request, "projects/_tab_response.html", context)


@require_GET
@login_required
def project_tab(request: HttpRequest, id: uuid.UUID, tab: str) -> HttpResponse:
    key = (tab or "").strip().lower()
    if key == "summary":
        return _render_tab_response(request, id, 0, "projects/tabs/_summary_panel.html")
    if key == "inspections":
        return _render_tab_response(request, id, 1, "projects/tabs/_placeholder_inspections.html")
    if key == "violations":
        return _render_tab_response(request, id, 2, "projects/tabs/_placeholder_violations.html")
    if key == "audit":
        return _render_tab_response(request, id, 3, "projects/tabs/_placeholder_audit.html")
    return HttpResponse("Not Found.", status=404)


@require_GET
@login_required
def project_list(request: HttpRequest) -> HttpResponse:
    """GET /projects — project list page with AG Grid SSRM panel.

    See docs/reference/ag-grid-ssrm-contract.md
    """
    if not can(request.user, "project.read"):
        return HttpResponseForbidden("You do not have permission to access this page.")

    can_create = can(request.user, "project.create")
    return render(
        request,
        "projects/index.html",
        {"can_create": can_create},
    )


def _render_422(request, form, _stale_trade_choices=None, _stale_inspector_choices=None) -> HttpResponse:
    """Re-render just the form partial with 422 status for HTMX swap.

    Always re-fetches reference data fresh so 422 re-renders never show options
    that became invalid between the initial GET and this POST (e.g., a trade type
    deactivated while the form was open). The stale-choices arguments are accepted
    but ignored for backwards compat with call sites inside the transaction block.
    """
    trade_choices, inspector_choices = _get_reference_data()
    return render(
        request,
        "projects/_create_form.html",
        {
            "form": form,
            "trade_choices": trade_choices,
            "inspector_choices": inspector_choices,
            "error_count": len(form.errors),
        },
        status=422,
    )


def _actor_id_from_request_user(request: HttpRequest) -> uuid.UUID:
    try:
        return request.user.dev_uuid.uuid  # type: ignore[union-attr]
    except Exception:  # noqa: BLE001
        logging.getLogger(__name__).error(
            "projects.transition: no dev_uuid for user_id=%s; "
            "using synthetic actor_id — seed runner may need to be re-run",
            request.user.pk,
        )
        return uuid.uuid5(uuid.NAMESPACE_DNS, f"django-user-{request.user.pk}")


def _validate_reason(reason: str, required: bool) -> str | None:
    value = (reason or "").strip()
    if required and not value:
        return "Reason is required."
    if value and len(value) > _REASON_MAX_LENGTH:
        return f"Reason must be {_REASON_MAX_LENGTH} characters or fewer."
    if value and _CONTROL_CHAR_PATTERN.search(value):
        return "Reason contains invalid control characters."
    return None


def _audit_row_context(action: str, actor_name: str, occurred_at: datetime, before_after_json: str) -> dict[str, object]:
    iso = occurred_at.astimezone(UTC).isoformat().replace("+00:00", "Z")
    return {
        "action": action,
        "actor_name": actor_name,
        "occurred_at": iso,
        "absolute": iso,
        "relative": "just now",
        "before_after_json": before_after_json,
        "expanded": False,
    }


def _render_transition_form(request: HttpRequest, id: uuid.UUID, *, action_path: str, submit_label: str, title: str, required: bool, reason: str = "", error: str | None = None, status: int = 200) -> HttpResponse:
    return render(
        request,
        "projects/_project_transition_form.html",
        {
            "project_id": id,
            "action_path": action_path,
            "submit_label": submit_label,
            "title": title,
            "required": required,
            "reason": reason,
            "error": error,
            "alert_error": error,
        },
        status=status,
    )


def _render_transition_success(request: HttpRequest, project: Project, *, action: AuditAction, reason: str, before_state: dict[str, object], after_state: dict[str, object]) -> HttpResponse:
    context = _build_project_detail_context(request, project)
    context["oob"] = {
        "audit": _audit_row_context(
            action=action.value,
            actor_name=request.user.get_username(),
            occurred_at=datetime.now(UTC),
            before_after_json=json.dumps({"before": before_state, "after": after_state}, sort_keys=True),
        ),
    }
    return render(request, "projects/_detail_transition_response.html", context)


def _render_transition_conflict(request: HttpRequest, project: Project, *, title: str, message: str) -> HttpResponse:
    context = _build_project_detail_context(request, project)
    context["transition_error"] = {"title": title, "message": message}
    return render(request, "projects/_detail_body.html", context, status=409)


@require_http_methods(["GET", "POST"])
@login_required
def project_place_on_hold(request: HttpRequest, id: uuid.UUID) -> HttpResponse:
    if request.method == "GET":
        if not can(request.user, "project.place_on_hold"):
            return HttpResponseForbidden("You do not have permission to access this page.")
        project = _load_project_or_404(id)
        if project is None:
            return HttpResponse("Not Found.", status=404)
        return _render_transition_form(
            request,
            id,
            action_path=f"/projects/{id}/place-on-hold",
            submit_label="Place on hold",
            title="Place project on hold",
            required=True,
        )

    if not can(request.user, "project.place_on_hold"):
        return HttpResponseForbidden("You do not have permission to access this page.")
    project = _load_project_or_404(id)
    if project is None:
        return HttpResponse("Not Found.", status=404)

    reason = request.POST.get("reason", "")
    err = _validate_reason(reason, required=True)
    if err is not None:
        return _render_transition_form(
            request,
            id,
            action_path=f"/projects/{id}/place-on-hold",
            submit_label="Place on hold",
            title="Place project on hold",
            required=True,
            reason=reason,
            error=err,
            status=422,
        )

    before_state: dict[str, object] = {}
    try:
        with transaction.atomic():
            project = Project.objects.select_for_update().get(pk=id)
            before_state = {"status": project.status}
            project.place_on_hold(reason)
            project.save(update_fields=["status"])
            append_audit_entry(
                actor_id=_actor_id_from_request_user(request),
                action=AuditAction.PROJECT_PLACED_ON_HOLD,
                entity_type="Project",
                entity_id=project.id,
                project_id=project.id,
                before_state=before_state,
                after_state={"status": project.status},
                metadata={"reason": reason.strip()},
            )
    except Project.DoesNotExist:
        return HttpResponse("Not Found.", status=404)
    except InvalidProjectTransition as ex:
        return _render_transition_conflict(
            request,
            project,
            title="Couldn't place project on hold",
            message=str(ex),
        )

    project = _load_project_or_404(id)
    if project is None:
        return HttpResponse("Not Found.", status=404)
    return _render_transition_success(
        request,
        project,
        action=AuditAction.PROJECT_PLACED_ON_HOLD,
        reason=reason.strip(),
        before_state=before_state,
        after_state={"status": project.status},
    )


@require_http_methods(["GET", "POST"])
@login_required
def project_resume(request: HttpRequest, id: uuid.UUID) -> HttpResponse:
    if request.method == "GET":
        if not can(request.user, "project.resume"):
            return HttpResponseForbidden("You do not have permission to access this page.")
        project = _load_project_or_404(id)
        if project is None:
            return HttpResponse("Not Found.", status=404)
        return _render_transition_form(
            request,
            id,
            action_path=f"/projects/{id}/resume",
            submit_label="Resume",
            title="Resume project",
            required=False,
        )

    if not can(request.user, "project.resume"):
        return HttpResponseForbidden("You do not have permission to access this page.")
    project = _load_project_or_404(id)
    if project is None:
        return HttpResponse("Not Found.", status=404)

    reason = request.POST.get("reason", "")
    err = _validate_reason(reason, required=False)
    if err is not None:
        return _render_transition_form(
            request,
            id,
            action_path=f"/projects/{id}/resume",
            submit_label="Resume",
            title="Resume project",
            required=False,
            reason=reason,
            error=err,
            status=422,
        )

    before_state: dict[str, object] = {}
    try:
        with transaction.atomic():
            project = Project.objects.select_for_update().get(pk=id)
            before_state = {"status": project.status}
            project.resume(reason.strip() or None)
            project.save(update_fields=["status"])
            append_audit_entry(
                actor_id=_actor_id_from_request_user(request),
                action=AuditAction.PROJECT_RESUMED,
                entity_type="Project",
                entity_id=project.id,
                project_id=project.id,
                before_state=before_state,
                after_state={"status": project.status},
                metadata={"reason": reason.strip()},
            )
    except Project.DoesNotExist:
        return HttpResponse("Not Found.", status=404)
    except InvalidProjectTransition as ex:
        return _render_transition_conflict(
            request,
            project,
            title="Couldn't resume project",
            message=str(ex),
        )

    project = _load_project_or_404(id)
    if project is None:
        return HttpResponse("Not Found.", status=404)
    return _render_transition_success(
        request,
        project,
        action=AuditAction.PROJECT_RESUMED,
        reason=reason.strip(),
        before_state=before_state,
        after_state={"status": project.status},
    )
