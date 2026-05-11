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

### Domain Aggregates

Four aggregates: **Project**, **Inspection**, **Violation** (with **CorrectiveAction** inside it). Full state machines, invariants, and DDL live in `_bmad-output/planning-artifacts/architecture.md` and `_bmad-output/planning-artifacts/prd/`.

### Canonical Audit Action Strings

`ProjectCreated`, `ProjectClosed`, `ProjectPlacedOnHold`, `ProjectResumed`, `InspectionScheduled`, `InspectionStarted`, `InspectionCompleted`, `InspectionCancelled`, `ViolationOpened`, `ViolationAssigned`, `ViolationVoided`, `CorrectiveActionSubmitted`, `CorrectiveActionTakenForReview`, `CorrectiveActionApproved`, `CorrectiveActionRejected`.

Stored verbatim in `domain.audit_entry.action`. Adding a string is an ADR amendment — do not invent semantically-similar variants.

### Canonical HTMX Target IDs

`#project-detail`, `#project-list`, `#violation-detail`, `#violation-list`, `#inspection-list`, `#audit-log`, `#compliance-tile` (OOB only), `#corrective-action-form`, `#corrective-action-list`, `#flash-region` (aria-live).

Identical across all three stacks. Inventing a new id is an ADR amendment.

### HTMX Patterns

- Partials have exactly one root element with a stable `id`. IDs are identical across all three stacks: `#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log`.
- State-changing actions use `<button hx-post>`, never links.
- `hx-swap-oob` only for header-level tiles (compliance score, notification badge); document each use site.
- The server decides whether to render action buttons — absent vs. disabled is a server decision.
- Domain exceptions return HTTP 409 with the originating partial showing the error and unchanged state.
- Input validation failures (malformed field, missing required value) return HTTP 422 with the originating form partial re-rendered; field-level `aria-invalid` + `aria-describedby` on each invalid field; top InlineAlert with `role="alert"`. No state is mutated on 422.

### AG Grid

Server-side row model only. Endpoint contract: `{ "rows": [...], "lastRow": N }`. Row selection fires an HTMX request to load a detail panel — the grid never owns the detail view. No business logic inside grid configurations.

### Shared Front-End Assets

`fieldmark_shared/` is the single source of truth for all shared front-end assets:

- **CSS** — Tailwind v4 compiles `src/fieldmark.css` → `dist/fieldmark.css`. Commit `dist/fieldmark.css`; no build step needed after cloning.
- **Vendor JS** — `vendor/ag-grid/` and `vendor/htmx/` contain the canonical copies of AG Grid and HTMX.

All three stacks consume these via **symlinks** into their `vendor/` static directories — there are no committed copies of vendor files inside any stack. The layout is identical across stacks:

| Asset | .NET | Django | Go/Fiber |
|---|---|---|---|
| `fieldmark.css` | `wwwroot/vendor/fieldmark.css` | `static/vendor/fieldmark.css` | `internal/web/static/vendor/fieldmark.css` |
| AG Grid | `wwwroot/vendor/ag-grid` | `static/vendor/ag-grid` | `internal/web/static/vendor/ag-grid` |
| HTMX | `wwwroot/vendor/htmx` | `static/vendor/htmx` | `internal/web/static/vendor/htmx` |

To add a new shared JS library: add it to `fieldmark_shared/vendor/`, create directory symlinks in all three stacks, and update `fieldmark_shared/CLAUDE.md`.

## Hard Rules (all stacks)

Cannot be relaxed without an ADR amendment:

- **Backend authority.** Domain rules, transitions, validation, authorization — server only.
- **Infrastructure-owned domain schema.** The `domain` schema is created by SQL init scripts in `docker/postgres/init/`, not by any framework's migration tooling. Framework migrations touch only their own auth schema.
- **No fat service layers.** In .NET and Django, handlers/views call entity methods directly — no intermediate service class owns domain logic. Go uses a thin `app` coordination layer by design (explicit dependency wiring); it must not contain business rules.
- **No repository or Unit-of-Work abstractions.** Use `DbContext` / ORM / explicit SQL directly.
- **No CQRS, MediatR, or in-process buses.**
- **No client-side state stores.** No Redux, Zustand, Pinia, NgRx, Signals, or equivalents.
- **No AutoMapper.** Project to view models manually.
- **No SQLite in tests.** Real PostgreSQL only (Testcontainers / pytest-django).
- **AuditEntry writes are non-optional** and always in the same transaction as the triggering write.
- **Stack symmetry** on routes, HTMX IDs, AG Grid contracts, audit action strings, and method names.
- **Casing is canonical at the wire and DB layer; idiomatic in source.** PascalCase in Python and snake_case in C# are *both wrong*. Database columns, JSON fields, and enum values are `snake_case` / `SCREAMING_SNAKE_CASE` everywhere. Code identifiers follow the language's native idiom — never converted for "consistency."

Stack-specific rules are in each project's own `CLAUDE.md`.

## Key Reference Documents

- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (decisions, patterns, structure, validation)
- `_bmad-output/planning-artifacts/prd/` — product requirements source of truth (sharded; index at `prd/index.md`)

The `_bmad-output/planning-artifacts/research/` folder contains pre-kickoff priming material that informed the canonical PRD and architecture. It is not maintained going forward and is not authoritative — agents should not rely on it.
