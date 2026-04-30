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
- **Symmetrical in intent (not mechanics) with the .NET implementation**

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

- PostgreSQL is the shared system of record
- Django migrations are permitted ONLY for:
  - Django framework tables
  - Auth tables
  - Admin support tables

Domain schema evolution is **not owned by Django**.

Rules:
- Django must not autonomously evolve domain tables once schema is established
- Shared domain tables must not be created by both EF Core and Django
- Dual ownership of tables is prohibited

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

Django’s built‑in auth:
- Is enabled for Admin and internal tooling
- Does not imply finalized product auth design

Rules:
- Auth usage must remain minimal and explicit
- No attempt is made to share identity with .NET
- External identity providers are out of scope

Auth is considered **orthogonal infrastructure**, not domain logic.

---

## Agent Guardrail Rules (Derivable)

An implementing agent must:
- Reject introduction of Django signals
- Reject business logic in views
- Prefer model methods for domain behavior
- Reject schema evolution ownership beyond Django’s internal tables
- Fail solutions that rely on implicit framework behaviors

If a solution relies on Django “magic” instead of explicit rules, it is invalid.

---

## Core Principle

**Django expresses the domain; it must never invent it.**

---

## Status

This document defines the **locked architectural guardrails** for the Django implementation of FieldMark and is intended as a stable priming artifact for agentic design enforcement.
