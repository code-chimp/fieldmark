# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repository Is

FieldMark is a construction compliance and inspection management system implemented **across three parallel stacks** — .NET (Razor Pages + HTMX), Django (Templates + HTMX), and Go (Fiber + HTMX) — against a shared PostgreSQL 17 database. It is a teaching artifact demonstrating server-authoritative architecture as an alternative to SPAs. A story is never done until all three stacks pass it.

## Infrastructure

```bash
docker compose up -d    # PostgreSQL 17 on localhost:5432 (fieldmark/fieldmark/fieldmark)
```

Postgres init scripts in `docker/postgres/init/` run automatically on first startup and create the required schemas: `domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`. If the container was previously started without these mounts, destroy the volume and restart: `docker compose down -v && docker compose up -d`.

Stack-specific commands are in `FieldMark/CLAUDE.md`, `fieldmark_py/CLAUDE.md`, and `fieldmark-go/CLAUDE.md`.

## Architecture

### Three-Stack Constraint

All three stacks expose identical routes, HTMX target IDs, AG Grid endpoint contracts, audit entry shapes, and domain method names (modulo language casing conventions). Diverging any of these across stacks is a defect.

Canonical state-transition method names: `start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`, `assign`, `submit_corrective_action`, `approve_resolution`, `reject_resolution`, `void`.

### Database Schema Ownership

The database uses schema-level isolation:

| Schema | Owner |
|---|---|
| `domain` | Infrastructure SQL init scripts — not any framework |
| `django_auth` | Django stack |
| `dotnet_auth` | .NET stack |
| `fiber_auth` | Go stack |

**The `domain` schema is not owned by any framework.** EF Core, Django ORM, and Go data access code all map to `domain.*` tables but do not create or migrate them. Framework migrations are scoped to their own auth schemas only. Domain tables store user references as opaque identifiers — no foreign keys to any auth table.

### The Canonical Request Flow

Every mutating handler in all three stacks follows this sequence — nothing more:

1. Authorize (role check + ownership check where applicable)
2. Begin transaction
3. Load aggregate root
4. Call the entity method (raises a typed exception on rule violation)
5. Append `AuditEntry` in the same transaction
6. Recompute `compliance_score` on the project if the action affects it
7. Commit
8. Render template partial (HTMX) or full page

If a handler is doing anything outside this list, the logic belongs on the entity.

### Domain Model

Four aggregates, all in `_bmad-output/planning-artifacts/research/domain-model.md` (the schema authority):

- **Project** — root aggregate; owns inspections, violations, and audit entries. `compliance_score` (0–100) is server-computed on every relevant write.
- **Inspection** — `Scheduled → InProgress → Completed`; also `Scheduled → Cancelled`. Completing with findings spawns `Violation` entities.
- **Violation** — `Open → InProgress → Resolved` (terminal) or `Voided` (terminal). Resolution only through a Compliance Officer–approved `CorrectiveAction`. No reopen path.
- **CorrectiveAction** — inside the Violation aggregate. `Submitted → UnderReview → Approved` or `Rejected`.

### HTMX Patterns

- Partials have exactly one root element with a stable `id`. IDs are identical across all three stacks: `#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log`.
- State-changing actions use `<button hx-post>`, never links.
- `hx-swap-oob` only for header-level tiles (compliance score, notification badge); document each use site.
- The server decides whether to render action buttons — absent vs. disabled is a server decision.
- Domain exceptions return HTTP 409 with the originating partial showing the error and unchanged state.

### AG Grid

Server-side row model only. Endpoint contract: `{ "rows": [...], "lastRow": N }`. Row selection fires an HTMX request to load a detail panel — the grid never owns the detail view. No business logic inside grid configurations.

### Styling

`fieldmark_style/` is the sole CSS source. Tailwind v4 compiles `src/fieldmark.css` → `dist/fieldmark.css`, which is symlinked into all three apps. The compiled `dist/` is committed to the repository — no build step is required after cloning.

## Hard Rules (all stacks)

Cannot be relaxed without an ADR amendment:

- **Backend authority.** Domain rules, transitions, validation, authorization — server only.
- **Infrastructure-owned domain schema.** The `domain` schema is created by SQL init scripts in `docker/postgres/init/`, not by any framework's migration tooling. Framework migrations touch only their own auth schema.
- **No service layer** between handler and entity.
- **No repository or Unit-of-Work abstractions.** Use `DbContext` / ORM / explicit SQL directly.
- **No CQRS, MediatR, or in-process buses.**
- **No client-side state stores.** No Redux, Zustand, Pinia, NgRx, Signals, or equivalents.
- **No AutoMapper.** Project to view models manually.
- **No SQLite in tests.** Real PostgreSQL only (Testcontainers / pytest-django).
- **AuditEntry writes are non-optional** and always in the same transaction as the triggering write.
- **Stack symmetry** on routes, HTMX IDs, AG Grid contracts, audit action strings, and method names.

Stack-specific rules are in each project's own `CLAUDE.md`.

## Key Reference Documents

All in `_bmad-output/planning-artifacts/research/`:

- `project-brief.md` — executive summary, thesis, personas, MVP scope
- `prd.md` — original product requirements (scope authority)
- `domain-model.md` — ERD overview, entity catalog, state machines, invariants, compliance scoring, PostgreSQL schema
- `architecture-decisions.md` — ADR-011 through ADR-014 + hard constraints and guardrails for agents
- `ux-guide.md` — screen inventory, UX principles, wireframe patterns
- `dotnet-reference.md` — .NET project structure, patterns, agent guardrails
- `django-reference.md` — Django project structure, patterns, agent guardrails
- `fiber-reference.md` — Go/Fiber project structure, patterns, agent guardrails
