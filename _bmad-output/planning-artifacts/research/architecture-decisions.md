# Architecture Decisions & Constraints — FieldMark

**Document type:** Architectural Decision Record + Hard Constraints
**Status:** Accepted — foundational constraint
**Project:** FieldMark (Construction Compliance & Inspection Management System)

This document serves as the **primary architectural constraint source** for downstream spec-driven and agentic design systems. Part 1 records the foundational architectural decision with its rationale. Part 2 operationalizes that decision into enforceable rules that deliberately narrow the solution space.

---

## Part 1: Architectural Decision Records

### Purpose

These ADRs record foundational architectural decisions that intentionally constrain the solution space. They are designed to support backend authority, reduce cognitive load, and align .NET and Django implementations so that architecture — not framework favoritism — is the focus of comparison.

### Architectural Goals

- Establish **backend authority** as the single source of truth
- Minimize duplicated logic between UI and server
- Ensure architectural **parity** across the .NET, Django, and Go (Fiber) stacks
- Reduce accidental complexity introduced by framework patterns
- Enable spec-driven and agent-assisted design without pattern drift

---

### ADR-011: ORM-First, Rich Domain Model (No CQRS / No Repositories)

**Decision**

The system shall use an **ORM-first architecture with rich domain models**, avoiding CQRS, MediatR, generic repository abstractions, and layered domain service architectures.

Business rules, invariants, and state transitions are implemented directly on ORM domain entities. Request handlers (Razor Pages / Django views) orchestrate workflows but do not own domain logic.

**Rationale**

- Django's architectural model centralizes behavior on models and request handlers
- EF Core supports rich entities with encapsulated invariants and lifecycle logic
- CQRS and repository patterns introduce abstraction asymmetry between stacks
- Eliminating architectural indirection reinforces backend authority
- Reduces the number of conceptual layers contributors must understand

**Consequences**

- Domain entities are not anemic data containers
- DbContext / Django ORM are accessed directly (no repository pass-throughs)
- Validation logic is implemented once, server-side
- Architectural explanations during demos are simplified

**Explicitly Rejected Alternatives**

- CQRS (Command / Query segregation)
- MediatR or message-based pipelines
- Generic Repository pattern layered over EF Core
- Clean Architecture / Onion Architecture layering

These were rejected to prevent architectural complexity from overshadowing the demo's core message.

**Relationship to Other ADRs**

- Complements ADR-003 (Django selection)
- Complements ADR-006 (Database-first modeling)
- Reinforces ADR-001 (Server-driven UI)

---

### ADR-012: Authentication Is Framework-Local; Authorization Is Domain-Driven

**Status**

Accepted

**Context**

FieldMark is implemented across multiple backend stacks: Django, ASP.NET Core (.NET), and Go (Fiber). Each framework has its own conventions, middleware, schema expectations, and preferred identity strategy. Django introduces authentication and admin tables early in the project lifecycle, while .NET and Fiber allow authentication to be deferred.

A shared authentication system across these frameworks would introduce unnecessary complexity: shared user-schema coupling, cross-framework identity synchronization, premature infrastructure, and loss of backend replaceability. The business domain still requires a consistent authorization model expressed in terms of domain-relevant roles and permissions.

**Decision**

Authentication is **framework-local**; authorization is **domain-driven**.

- Each backend owns its own authentication implementation.
- Each backend stores its auth tables in its own Postgres schema: `django_auth`, `dotnet_auth`, `fiber_auth`.
- The FieldMark domain does not assume a shared user model.
- Roles and permissions are defined conceptually at the product level (Administrator, Compliance Officer, Inspector, Site Supervisor, Executive Viewer) and implemented natively by each framework.
- Domain tables in the `domain` schema **do not foreign-key any auth table**. User references are stored as opaque identifiers (e.g. `created_by_user_id`, `assigned_to_user_id`) meaningful only to the calling backend.

**Consequences**

- Frameworks remain replaceable; auth concerns are cleanly isolated.
- Domain tables remain stable and portable.
- Django's built-in auth can be used without forcing premature symmetry.
- .NET and Fiber can add auth later without redesigning the domain.
- There is no shared login experience across frameworks (acceptable for a teaching artifact).
- User references in domain data are not relationally enforced across auth systems; integrity is the application's responsibility.
- Role mapping must be maintained separately in each implementation.

**Explicitly Rejected Alternatives**

- Shared `users` table across all frameworks — rejected for cross-framework coupling and portability loss.
- Centralized identity provider for the demo — out of scope and operationally unnecessary.
- Full auth implementation in all frameworks immediately — would prematurely harden framework-specific decisions.

**Relationship to Other ADRs**

- Pairs with ADR-014 (domain schema is framework-neutral).
- Depends on ADR-013 (auth schemas are pre-created by infrastructure).

---

### ADR-013: PostgreSQL Schemas Are Created by Infrastructure Init Scripts

**Status**

Accepted

**Context**

FieldMark uses a shared PostgreSQL database across multiple backend implementations and requires deterministic local setup, clear schema ownership, and strong separation between shared domain infrastructure and framework-specific support tables.

If schemas are created implicitly by application code or framework migrations, the system becomes fragile: startup order matters, local setup is error-prone, frameworks accidentally assume infrastructure ownership, and onboarding is harder. Django in particular expects database structures to exist prior to migration, and its defaults can otherwise fall back to the `public` schema.

**Decision**

Schemas are treated as **infrastructure** and are created by PostgreSQL init scripts mounted into the Postgres container via `docker-entrypoint-initdb.d`.

The init scripts create the structural invariants:

- `domain`
- `django_auth`
- `dotnet_auth`
- `fiber_auth`
- `infra` (reserved for cross-stack metadata; optional for now)

Application frameworks **must not** create schemas. Framework migrations may create tables only within their authorized schemas.

**Consequences**

- `docker compose up` is sufficient to establish schema boundaries.
- Schema ownership is explicit, reviewable, and identical across local, CI, and demo environments.
- New backends can be added without changing startup assumptions.
- Framework migrations become simpler and safer.
- Infrastructure SQL must be maintained manually; developers must understand the distinction between schema creation and table migrations.
- Postgres init scripts only run when the data volume is empty. Adding a schema later requires destroying and recreating the volume (`docker compose down -v && docker compose up -d`).

**Explicitly Rejected Alternatives**

- Create schemas through Django migrations — rejected because schemas are infrastructure, not framework-owned data.
- Create schemas through EF Core migrations — same reason, plus it would bias ownership toward .NET.
- Manual schema creation as a setup step — degrades developer experience and introduces drift.

**Relationship to Other ADRs**

- Enables ADR-012 (framework-local auth schemas) and ADR-014 (infrastructure-owned domain schema).

---

### ADR-014: Shared Domain Schema Is Infrastructure-Owned and Framework-Neutral

**Status**

Accepted

**Context**

FieldMark is explicitly a multi-backend architecture demonstration. Its shared business domain must remain authoritative regardless of whether the current implementation is accessed via Django, .NET, or Go (Fiber).

In single-stack applications, code-first migration systems such as EF Core and Django ORM migrations are appropriate. In a multi-backend environment, allowing one framework to generate and evolve the shared domain schema would implicitly grant ownership to that framework, distort naming and constraints around ORM defaults, weaken architecture neutrality, and create confusion about what is shared versus framework-local.

This decision **supersedes** the language in ADR-011 and earlier `dotnet-reference.md` / `django-reference.md` drafts that suggested EF Core or Django migrations own the domain schema.

**Decision**

The shared FieldMark domain schema is **owned by the architecture**, not by any framework. Shared domain tables in the `domain` schema are created and evolved using **infrastructure-level SQL migrations** committed alongside the Docker init scripts.

Examples of shared domain tables include `domain.project`, `domain.inspection`, `domain.finding`, `domain.violation`, `domain.corrective_action`, `domain.audit_entry`, and the reference-data tables (`domain.trade_type`, `domain.violation_category`, `domain.compliance_rule`).

Frameworks **map to** these tables but do not create or evolve them:

- EF Core entities map explicitly to `domain.*` via fluent configuration. Auto-migrations are scoped to `dotnet_auth`.
- Django models that represent `domain.*` tables use `Meta.managed = False` (or equivalent discipline). Django migrations are scoped to `django_auth`.
- Fiber data access uses explicit SQL against `domain.*`; Fiber owns nothing in `domain`.

**Consequences**

- The shared business model is explicit, reviewable, and stable across implementations.
- Framework neutrality and backend replaceability are preserved.
- Schema drift between stacks is reduced; the demo's teaching value improves.
- Domain SQL must be authored and maintained directly; the convenience of ORM auto-migration is reduced for shared tables.
- Change management for the domain becomes deliberate: update the ERD, author an infrastructure migration, review for cross-framework impact, update mappings, apply.

**Explicitly Rejected Alternatives**

- EF Core owns the shared domain schema — rejected; makes .NET the implicit design authority.
- Django owns the shared domain schema — rejected; makes Django the implicit design authority.
- Dual migration ownership across frameworks — rejected; brittle, confusing, and architecturally unsound.

**Related Rules**

- Framework auth schemas remain framework-owned (ADR-012).
- Domain tables must not foreign-key framework auth tables (ADR-012).
- Framework models must be updated to match infrastructure-owned schema changes; "silent" ORM auto-migration of domain tables is prohibited.

**Relationship to Other ADRs**

- Supersedes the earlier suggestion in ADR-011 that EF Core or Django migrations own the domain schema.
- Depends on ADR-013 for schema creation.
- Pairs with ADR-012 to keep auth and domain ownership cleanly separated.

---

### Cross-reference — data layer and identity (ADRs 012–014)

These three ADRs are the **normative spine** for Postgres layout and ownership across stacks:

| ADR | Topic |
|-----|--------|
| **012** | Framework-local authentication schemas and opaque user references on `domain.*` |
| **013** | Schemas created only by infrastructure init scripts — frameworks never create schemas |
| **014** | Shared `domain.*` DDL owned by infrastructure SQL, not by EF Core / Django / Fiber migrations |

**Planning primers** (orientation for agents; ADR text above remains authoritative):

- [`authentication-authorization-primer.md`](authentication-authorization-primer.md) — roles, boundaries, non-goals
- [`domain-schema-ownership-primer.md`](domain-schema-ownership-primer.md) — domain DDL change workflow and checklist

**Implementation:** SQL under `docker/postgres/init/` (mounted via Postgres `docker-entrypoint-initdb.d`). Application stacks map to existing tables only.

---

## Part 2: Hard Constraints & Guardrails

### Purpose

This section defines hard architectural constraints that downstream design agents, spec-driven systems (BMAD), and human contributors must adhere to. It operationalizes the ADR intent into enforceable rules.

If a solution requires explaining a pattern, that solution is likely invalid.

---

### Global Architectural Principles

- One system of record
- Backend owns all business rules and validation
- UI layers are projections, not authorities
- Complexity must be earned
- Architectural symmetry across stacks is mandatory

---

### .NET Implementation Constraints

**Allowed**

- ASP.NET Razor Pages
- EF Core with direct DbContext usage
- Rich domain entities with behavior
- Minimal APIs for data-only endpoints
- HTMX for incremental UI updates

**Disallowed**

- CQRS patterns (commands/queries, handler pipelines)
- MediatR or equivalent mediator frameworks
- Generic or abstract repository layers
- Anemic entity models
- Client-side validation as a source of truth

---

### Django Implementation Constraints

**Allowed**

- Django ORM as primary data access
- Business rules defined on models and forms
- Django views as workflow orchestrators
- Django templates with HTMX

**Disallowed**

- Service layers duplicating model behavior
- Client-side state machines
- Multiple competing domain representations

---

### Go (Fiber) Implementation Constraints

**Allowed**

- Fiber as the HTTP/rendering adapter
- Go's standard `html/template` (via Fiber's HTML engine) for server-rendered pages and HTMX fragments
- Explicit SQL against `domain.*` for persistence; small, narrow store interfaces (e.g. `ProjectStore`, `ViolationStore`)
- Thin handlers that orchestrate app services; HTMX fragments served from the web layer

**Disallowed**

- Business logic in handlers or middleware
- `fiber.Ctx` escaping the web layer
- Generic repository abstractions or "clean architecture" ports without a concrete need
- ORM-style migrations of `domain.*` tables (Fiber owns nothing in `domain`)
- Client-side workflow orchestration

See `fiber-reference.md` for the full layout, layer responsibilities, and dependency-direction rules.

---

### Data & Persistence Constraints

- PostgreSQL is the canonical datastore
- Schema is treated as a shared contract
- **The `domain` schema is infrastructure-owned and framework-neutral (ADR-014).** It is created by Postgres init scripts under `docker/postgres/init/`, evolved by hand-authored infrastructure SQL, and is **not** migrated by EF Core, Django, or any other framework. Framework ORMs map to `domain.*` tables but never create or alter them.
- **Auth schemas are framework-local (ADR-012).** EF Core migrations are scoped to `dotnet_auth`; Django migrations are scoped to `django_auth`; Fiber, when it adopts auth, owns `fiber_auth`. Domain tables must not foreign-key any auth table — user references are stored as opaque identifiers.
- All Postgres schemas (`domain`, `django_auth`, `dotnet_auth`, `fiber_auth`, `infra`) are created by infrastructure init scripts (ADR-013), never by application code or framework migrations.
- Migrations must preserve cross-stack compatibility within the schemas each framework owns.
- No stack-specific database features without parity consideration.
- **All database column and table names use `snake_case`.** Django uses `snake_case` natively but must override the default `<app>_<model>` table-name prefix by setting `Meta.db_table` explicitly (e.g. `db_table = "project"`). .NET configures EF Core globally via `UseSnakeCaseNamingConvention()` and sets table and schema names via `ToTable("project", "domain")` in fluent config. Go/Fiber writes explicit SQL against `domain.<table>`. A schema diff across the three stacks must produce zero naming differences. See `domain-model.md` §9 for the full naming convention reference.

---

### UI & Interaction Constraints

- HTMX is the primary mechanism for interactivity
- JavaScript is limited to UI islands (AG Grid)
- No frontend state stores (Redux, Signals, Stores, etc.)
- Navigation and workflow transitions are server-driven

---

### Unit testing & E2E boundaries

- **Unit tests** (per stack, idiomatic tools—see `dotnet-reference.md`, `django-reference.md`, `fiber-reference.md`) prove **domain rules, application orchestration, and authorization logic**. They must not substitute for **Playwright E2E** coverage of full user workflows, HTMX-driven UI state, or **cross-stack behavioral parity**.
- **Playwright E2E** validates end-to-end workflows and shared UX contracts; **unit tests must not duplicate E2E scenarios**.
- Framework adapters—Razor Pages, Django views/templates/routing, Fiber handlers/middleware/templating—are **not** the primary subject of unit tests; behavior lives in the domain and application layers.
- Prefer tests that are **fast and deterministic**; avoid gratuitous mocking of HTTP. For **database-backed tests**, this repository enforces **real PostgreSQL** (no SQLite); contributors run tests against local Postgres (e.g. Docker) as documented in root `CLAUDE.md` and stack guides.

---

### Agentic Design Guardrails

Any agent or automated design system must:

- Prefer existing domain behavior over new abstractions
- Reject solutions requiring CQRS, repositories, or mediators
- Centralize validation on the server
- Optimize for understandability over extensibility
