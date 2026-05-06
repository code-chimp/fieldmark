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
- Ensure architectural **parity** between .NET and Django stacks
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

### Data & Persistence Constraints

- PostgreSQL is the canonical datastore
- Schema is treated as a shared contract
- Migrations must preserve cross-stack compatibility
- No stack-specific database features without parity consideration
- **All database column and table names use `snake_case`.** Django uses `snake_case` natively but must override the default `<app>_<model>` table-name prefix by setting `Meta.db_table` explicitly (e.g. `db_table = "project"`). .NET configures EF Core globally via `UseSnakeCaseNamingConvention()` and sets table names via `ToTable("project")` in fluent config. A schema diff between the two stacks must produce zero naming differences. See `domain-model.md` §9 for the full naming convention reference.

---

### UI & Interaction Constraints

- HTMX is the primary mechanism for interactivity
- JavaScript is limited to UI islands (AG Grid)
- No frontend state stores (Redux, Signals, Stores, etc.)
- Navigation and workflow transitions are server-driven

---

### Agentic Design Guardrails

Any agent or automated design system must:

- Prefer existing domain behavior over new abstractions
- Reject solutions requiring CQRS, repositories, or mediators
- Centralize validation on the server
- Optimize for understandability over extensibility
