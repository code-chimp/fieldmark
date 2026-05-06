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

## Hard Rules

- **No Django signals** — not for business logic, not for side effects, not ever without an ADR.
- **No business logic in views, forms, managers, or middleware.** It belongs on model methods.
- **No service layers** that duplicate or wrap model behavior.
- **No cross-app side effects.** An action in `violations/` must not trigger behavior in `inspections/` implicitly.
- **Django migrations are scoped to `django_auth` only** — Django's built-in auth, admin, and session tables. The shared `domain` schema is created by SQL init scripts in `docker/postgres/init/` and is not Django's to create or evolve. Domain models must set `managed = False` in their `Meta` class so Django never attempts to create or alter those tables.
- **No SQLite in tests.** Use real PostgreSQL via `pytest-django`.

## Reference

- `_bmad-output/planning-artifacts/research/django-reference.md` — full Django guardrails (authoritative)
- `_bmad-output/planning-artifacts/research/architecture-decisions.md` — ADRs and hard constraints
