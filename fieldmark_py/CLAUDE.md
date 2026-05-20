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

**Login and logout** views and the unauthenticated-redirect contract are Story 1.11's scope.

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
