# Architectural Constraints (PRD-Binding)

This section elevates non-negotiable architectural rules from `architecture-decisions.md` to PRD-binding status. Every story, every implementation decision, every cross-stack diff must respect these. Violating any of them is a defect, not a tradeoff.

## Backend Authority (ADR-011 spirit)

- Domain rules, validation, state transitions, and authorization decisions are server-side only.
- The client requests HTML; the server decides what is rendered, including whether action buttons are present, disabled, or absent.
- No business rule may be expressed in client-side code in any of the three stacks.
- Validation may exist on the client only as a UX courtesy; it is never authoritative and never the only enforcement.

## Stack Symmetry

The three stacks (.NET / Razor Pages, Django, Go / Fiber) must hold the following identical, modulo language casing:

- Route inventory.
- HTMX target IDs (`#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log`).
- AG Grid endpoint contract (`{ "rows": [...], "lastRow": N }`).
- Audit action strings (`ProjectClosed`, `ViolationOpened`, `CorrectiveActionApproved`, etc.).
- Domain method names (`start`, `complete`, `cancel`, `place_on_hold`, `resume`, `close`, `assign`, `submit_corrective_action`, `approve_resolution`, `reject_resolution`, `void`).

Any divergence is a defect. A story is not done until all three stacks pass it.

## Domain Schema Ownership (ADR-013, ADR-014)

- The `domain` schema is owned by infrastructure, not by any framework. It is created by Postgres init scripts in `docker/postgres/init/` and evolved by hand-authored SQL.
- EF Core, Django ORM, and Go data access **map to** `domain.*` tables but never create or alter them.
- Framework migrations are scoped to their own auth schemas (`dotnet_auth`, `django_auth`, `fiber_auth`) only.
- Schema drift between any stack and the canonical DDL is a build-blocking defect.

## Authentication & Authorization (ADR-012)

- Authentication is framework-local. Each stack owns its own auth schema and identity tables.
- Authorization is domain-driven. Roles (Admin, Compliance Officer, Inspector, Site Supervisor, Executive) are defined conceptually at the product level and implemented natively per stack.
- `domain.*` rows reference users only as opaque UUIDs. There are no foreign keys from `domain.*` to any auth schema.

## Forbidden Patterns (Across All Stacks)

The following are non-negotiably out of scope. Any story or implementation that requires them must trigger an ADR amendment, not a quiet exception:

- CQRS, MediatR, or in-process command/query buses.
- Generic Repository or Unit-of-Work abstractions over the ORM.
- Clean Architecture / Onion / hexagonal layering.
- AutoMapper or equivalent reflection-driven mappers (project to view models manually).
- Client-side state stores (Redux, NgRx, Pinia, Zustand, Signals stores, etc.).
- Client-owned routing or workflow orchestration.
- Fat service layers in .NET or Django that own domain logic. Go uses a thin `app` layer for explicit dependency wiring; it must not contain business rules.
- SQLite (or any non-PostgreSQL substitute) in tests. Real PostgreSQL only — Testcontainers in .NET, pytest-django in Python, equivalent in Go.

## Auditability

- An `AuditEntry` is written for every domain mutation, in the same database transaction as the change it records.
- Audit entries are append-only at the application level and (in production) at the schema level via revoked privileges.
- The audit row's actor, action, entity reference, and before/after state JSON are sufficient to reconstruct any change.
- Compliance score recomputation, when triggered by a write, occurs in the same transaction as that write.

## Interaction Architecture

- HTMX is the sole interactivity mechanism beyond AG Grid's data-fetching path. JavaScript outside AG Grid wiring and minimal UX glue is a defect.
- HTMX partials have exactly one root element with a stable, cross-stack-identical `id`.
- State-changing actions use `<button hx-post>`, never anchor tags.
- `hx-swap-oob` is permitted only for header-level tiles (compliance score, notification badge); each use site is documented at the call site.
- Domain rule violations return HTTP 409 with the originating partial re-rendered showing the error and unchanged state.

## Testing Boundaries

- Unit tests prove domain rules, application orchestration, and authorization. They do not stand in for E2E coverage.
- Playwright drives end-to-end scenarios across all three stacks; the same scenario passes in all three.
- Framework adapters (Razor Pages, Django views/templates, Fiber handlers) are not the primary subject of unit tests — behavior lives in the domain.
- Database-backed tests run against real PostgreSQL.

## Why These Are PRD-Binding, Not Just Architectural

This product's deliverable is the architectural argument itself. A feature that ships but violates these constraints fails the product, not just the architecture — the PRD success criteria are explicitly architectural. Lifting these into the PRD makes them visible to every downstream agent and contributor working from the PRD alone.
