# CLAUDE.md — Django Stack

This file provides guidance to Claude Code (claude.ai/code) when working in the `fieldmark_py/` Django project. Read alongside the root `CLAUDE.md`.

## Commands

Run from the `fieldmark_py/` directory:

```bash
uv sync
uv run python manage.py runserver
uv run python manage.py migrate
uv run python manage.py makemigrations
uv run python -m pytest
uv run python -m pytest path/to/test_file.py::TestClass::test_method
uv run ruff check .
uv run mypy .
```

## Project Structure

Apps map to bounded contexts, not technical layers. Each app owns its models, views, urls, forms, templates, and static assets:

- `projects/` — Project aggregate
- `inspections/` — Inspection aggregate
- `violations/` — Violation aggregate and CorrectiveAction
- `audit/` — AuditEntry
- `compliance/` — rules engine and compliance scoring
- `reference/` — TradeType, ViolationCategory, ComplianceRule (admin-managed reference data)
- `grid/` — AG Grid JSON endpoints

## What Belongs Where

**Model methods** — all state-transition logic (`place_on_hold`, `close`, `approve_resolution`, etc.), `can_*` predicates, domain invariants, and typed exceptions. This is the only place domain behavior lives.

**Views** — thin orchestrators. The only permitted pattern:
1. Authorize (role check + ownership check)
2. `with transaction.atomic():`
3. Load aggregate via ORM
4. Call model method
5. Write `AuditEntry`
6. Recompute compliance score if affected
7. Render template (partial or full page)

**Forms** — input validation that mirrors entity invariants. Forms are not the source of truth; models enforce invariants independently.

**Django Admin** — platform tooling for reference data management. Not part of the product UX. Over-customization is discouraged.

## Coding Standards

- Wrap every mutating view in `transaction.atomic`.
- Use `TextChoices` for enum-like fields.
- Use `select_related` / `prefetch_related` aggressively to avoid N+1 queries.
- Class-based or function-based views are acceptable; stay consistent within an app.
- Structured logging via `structlog`.

## Hard Rules (Django-specific)

Root `CLAUDE.md` covers cross-stack rules (no service layers, no client-side state, real PostgreSQL in tests, infrastructure-owned `domain` schema). The Django-specific rules are:

- **No Django signals** — not for business logic, not for side effects, not ever without an ADR.
- **No business logic in views, forms, managers, or middleware.** It belongs on model methods.
- **No cross-app side effects.** An action in `violations/` must not trigger behavior in `inspections/` implicitly.
- **Django migrations are scoped to `django_auth` only.** Domain models must set `Meta.managed = False` and `db_table = 'domain"."<table>'` so Django never attempts to create or alter `domain.*` tables.

## Authentication

Django's built-in `auth`, `sessions`, `admin`, and `contenttypes` apps are the framework-native auth source. No custom user model — `AbstractUser` is not subclassed (Architecture D7).

**Schema mechanism:** All framework-managed tables land in `django_auth` via `OPTIONS["options"] = "-c search_path=django_auth,public"` on `DATABASES["default"]`. This supersedes the `fieldmark/routers.py` approach shown in the architecture directory diagram — a custom `DatabaseRouter` and per-model `db_table` overrides are not needed and must not be added. The divergence is intentional.

**Domain model isolation:** Domain models set `Meta.managed = False` and `db_table = 'domain"."<table>'` so Django never CREATEs or ALTERs `domain.*` tables (ADR-014). Django migrations are scoped to `django_auth` only.

**Conceptual roles:** Roles map to Django `Group` objects. The five canonical group names are `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`. Seed them with:

```bash
uv run python manage.py seed_groups
```

The command lives in `tools/management/commands/seed_groups.py` and is idempotent — safe to re-run. Do not wire it to `AppConfig.ready()` or a `post_migrate` signal.

**Story 1.11 shipped:** `login_view` (`GET`/`POST`, `@login_not_required`) and `logout_view` (`POST`, `@login_not_required`) are in `fieldmark/views.py`. `LoginRequiredMiddleware` enforces unauthenticated-redirect for all other views. `LOGIN_URL = "/login"`, `LOGIN_REDIRECT_URL = "/"`, `LOGOUT_REDIRECT_URL = "/login"` are set in `settings.py`.

## Authorization

The single Django-side authorization decision primitive is `fieldmark.authz.can` in `fieldmark/authz.py`. Signature:

```python
can(user, action: str, entity_id: uuid.UUID | None = None) -> bool
```

**Rules:**
- Views call `can(request.user, action)`; templates receive pre-computed `permission` booleans — templates must never call `can` directly.
- Role names are defined in `fieldmark/roles.py` as the `Role` StrEnum (`Role.ADMIN`, `Role.COMPLIANCE_OFFICER`, etc.). Hard-coded role-name string literals elsewhere are a defect.
- Actions are registered at module load time via `register_action(action, *roles)`. Do not wire registration to `AppConfig.ready()` or signals; use module-level statements in the handler package. Story 1.12 ships the map empty — Epic 1 has no live action affordances.
- Entity-scope rules are deferred to Epic 2+ and will wire into `_evaluate_entity_scope` inside `authz.py`.

**ActionButton template:** `templates/components/_action_button.html`. Include via `{% include "components/_action_button.html" with id=... permission=... state_allows=... ... %}`. The caller supplies pre-computed `permission` (from `can`) and `state_allows` (from the entity's `can_*` predicate). The template handles the trichotomy internally.

Canonical snapshot reference: `fieldmark_shared/components/action_button.example.html`.

## Authentication / User UUIDs

Django's `auth_user.id` is a `BIGSERIAL` AutoField. The project does not use a custom user model, so the canonical cross-stack UUIDs (from `docker/postgres/init/seed-uuids/dev-users.json`) **cannot** be `auth_user` primary keys. The chosen approach is a side table:

```
django_auth.dev_user_uuid (user_id BIGINT FK → auth_user.id, uuid UUID UNIQUE)
```

Model: `tools.models.DevUserUuid` — `OneToOneField(User, related_name="dev_uuid")`.

**Why a side table** (not a custom user model or a custom column on `auth_user`): A custom user model requires being committed before any migrations land — `auth_user` already has an integer PK from Story 1.8, so that option is closed. Adding a `uuid` column to `auth_user` mutates the framework table beyond its standard shape. The side table adds one row per user, leaves `auth_user` untouched, and lives in `django_auth` (clean ADR-012 story).

**Lookup at audit-write time (Epic 2+):** `request.user.dev_uuid.uuid` — the `dev_uuid` reverse accessor returns the `DevUserUuid` row, and `.uuid` is the `domain.audit_entry.actor_id` value to write. Do not use `request.user.pk` (integer) as an actor ID.

## Reference

- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (canonical request flow with Django code stub, decisions, patterns)
- `_bmad-output/planning-artifacts/prd/` — capability source of truth
- Root `CLAUDE.md` — cross-stack rules and canonical inventories (audit actions, HTMX target IDs, method names)

## Home page

The Home page template lives at `fieldmark_py/templates/pages/home.html` and is served by `fieldmark.views.home` at `/`.

**This page is intentionally empty in Epic 1.** It renders `<h1>FieldMark</h1>`, the role badge, and a placeholder paragraph only. Story 2.10 replaces it with the real Compliance Dashboard.

**Chrome composition order (AC #2, Story 1.13 — all three stacks must match):**
`<a class="fm-wordmark">` → `<div class="ml-auto flex items-center gap-3">` containing `_theme_toggle.html` then `_avatar_menu.html`. Any new chrome control added to any stack must be added to all three in the same commit (FR58).

**Role → badge-token mapping** (locked in Story 1.13; source of truth is `fieldmark/roles.py` — `LABELS` and `BADGE_TOKENS` dicts keyed by `Role` enum):

| Role | Token | Label |
|---|---|---|
| `ADMIN` | `danger` | Admin |
| `COMPLIANCE_OFFICER` | `info` | Compliance Officer |
| `INSPECTOR` | `warning` | Inspector |
| `SITE_SUPERVISOR` | `neutral` | Site Supervisor |
| `EXECUTIVE` | `success` | Executive |

The badge `<span class="badge badge-{{ role_badge_token }}" role="status">{{ role_label }}</span>` is the first cross-stack visual proof of identity. Never hard-code tokens or labels outside `roles.py`.

**Tooltip escaping:** Any template that emits a `data-tooltip` attribute must pass the value through Django's auto-escaping (`{{ value }}` — do not use `{{ value|safe }}`). The Django template engine auto-escapes by default, which correctly encodes `&`, `<`, `>` into entities. `{{ value|safe }}` bypasses this and must never be used for tooltip content.
