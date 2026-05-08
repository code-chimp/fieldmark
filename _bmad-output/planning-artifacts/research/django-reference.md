# FieldMark Django Architecture & Guardrails Reference

## Purpose
This document defines the **architectural intent, constraints, and guardrails** for the Django implementation of FieldMark. It is designed to be used as **priming context for agentic systems** (BMAD, coding agents, architectural agents) to ensure that Django development remains aligned with the overall FieldMark architecture and its backend‑authority philosophy.

This document is intentionally **shorter and more prohibitive** than the .NET reference, because Django’s defaults are powerful and must be explicitly constrained to preserve architectural parity.

---

## Architectural Intent (Non‑Negotiable)

The Django implementation of FieldMark is:
- **Server‑authoritative**
- **Domain‑centric**
- **Workflow‑driven by backend rules**
- **Symmetrical in intent (not mechanics) with the .NET and Go (Fiber) implementations**

Django is used as an *adapter framework*, not as an implicit design authority.

---

## Conceptual Parity Model

| Concept | .NET | Django |
|------|-----|--------|
| Domain logic | Domain project | Model methods / domain modules |
| Persistence | EF Core | Django ORM |
| Composition | Program.cs | Django settings + views |
| Admin tooling | Deferred | Django Admin |

Parity is **conceptual**, not idiomatic.

---

## Project & App Structure

Expected structure (illustrative):

```
fieldmark_py/
  fieldmark/
    settings.py
    urls.py
    wsgi.py

  inspections/
    models.py
    views.py
    admin.py

  manage.py
```

Apps represent **bounded contexts**, not technical layers.

---

## Domain & Business Logic Rules

### Allowed Locations

Business rules may exist only in:
- Django model methods
- Explicit domain helper modules within an app

Examples:
- State transition methods
- Validation enforcing invariants
- Calculated properties derived from state

### Forbidden Locations

Business logic must NOT live in:
- Views
- Django signals
- Middleware
- Custom managers performing orchestration

Views orchestrate requests; **models own behavior**.

---

## Database & Migration Ownership

- PostgreSQL is the shared system of record.
- The `domain` schema is **infrastructure-owned** (ADR-014). It is created by Postgres init scripts in `docker/postgres/init/` and evolved through hand-authored infrastructure SQL. Django migrations do **not** create or alter `domain.*` tables.
- Django migrations are scoped exclusively to the `django_auth` schema (ADR-012). This covers Django's built-in auth tables (`auth_user`, `auth_group`, `auth_permission`, etc.), admin support tables, sessions, and content types — all configured to live in `django_auth`.
- Django models that represent `domain.*` rows must use `Meta.managed = False` together with an explicit `Meta.db_table = "domain\".\"<table_name>"` (or the project's adopted convention for cross-schema references) so Django can read and write but never `CREATE`, `ALTER`, or `DROP`.
- The `DATABASE_ROUTERS` / `search_path` configuration must ensure migrations target only `django_auth`; CI must fail on any generated migration that touches another schema.
- Dual ownership of `domain.*` tables across stacks is prohibited; running `manage.py migrate` must never produce DDL against `domain`.

---

## Django Admin Policy

Django Admin is treated as:
- Platform tooling
- Operator support
- Configuration convenience

Admin UI is:
- Not part of the product UX
- Not required to mirror .NET UI
- Not a driver of domain behavior

Over‑customization of Admin is discouraged unless justified.

---

## Explicitly Rejected Django Patterns

Unless reversed by an explicit ADR, the following are prohibited:

- Django signals
- Fat views containing business logic
- Service layers duplicating model behavior
- Custom managers implementing workflows
- Cross‑app side effects
- Implicit transactional coupling

Framework “magic” must not replace explicit domain rules.

---

## Dependency & Coupling Rules

- Apps may depend on shared domain concepts
- Apps must not reach into each other’s internals
- No hidden coupling via signals or global state

Changes to one app must not implicitly trigger behavior in another.

---

## Authentication & Authorization Policy

Authentication is **framework-local** (ADR-012). Django uses its built-in auth and admin; all auth/admin tables live in the `django_auth` schema.

Rules:
- Django's built-in auth is enabled for Admin and internal tooling and is the Django stack's primary identity system.
- All auth, admin, sessions, and content-type tables are configured to live in `django_auth` — not `public`, not `domain`.
- Domain models in `domain.*` must not declare a `ForeignKey` to `auth_user` or any `django_auth.*` table. User references on domain rows are stored as opaque `UUIDField` values (e.g. `created_by_user_id`) populated from `request.user.id` at the view layer.
- Django authorization (groups, permissions, decorators) maps to the shared conceptual role vocabulary: Administrator, Compliance Officer, Inspector, Site Supervisor, Executive Viewer.
- No attempt is made to share identity with the .NET or Fiber stacks. External identity providers are out of scope.

Auth is considered **orthogonal infrastructure**, not domain logic.

---

## Agent Guardrail Rules (Derivable)

An implementing agent must:
- Reject introduction of Django signals
- Reject business logic in views
- Prefer model methods for domain behavior
- Reject any model representing a `domain.*` table that omits `Meta.managed = False` (ADR-014)
- Reject any generated migration whose SQL targets a schema other than `django_auth`
- Reject `ForeignKey` declarations from `domain.*` models to `auth_user` or any other `django_auth.*` table (ADR-012)
- Fail solutions that rely on implicit framework behaviors

If a solution relies on Django "magic" instead of explicit rules, it is invalid.

---

## Core Principle

**Django expresses the domain; it must never invent it.**

---

## Status

This document defines the **locked architectural guardrails** for the Django implementation of FieldMark and is intended as a stable priming artifact for agentic design enforcement.
