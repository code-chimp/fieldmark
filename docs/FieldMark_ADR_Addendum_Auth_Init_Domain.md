# FieldMark ADR Addendum: Authentication, Database Initialization, and Shared Domain Ownership

## Purpose

This document contains draft Architectural Decision Records (ADRs) that extend the existing FieldMark ADR set.

These ADRs formalize the recent decisions around:

- authentication and authorization boundaries
- PostgreSQL schema initialization
- shared domain schema ownership in a multi-backend system

These entries are intended to be added to the existing ADR set and treated as normative architecture decisions.

---

## ADR-012: Authentication Is Framework-Local; Authorization Is Domain-Driven

### Status
Accepted

### Context

FieldMark is implemented across multiple backend stacks:

- Django
- ASP.NET Core (.NET)
- Go (Fiber)

Each framework has its own conventions, middleware, schema expectations, and preferred identity implementation strategy. Django introduces authentication and administration tables early in the project lifecycle, while .NET and Fiber allow authentication to be deferred.

A shared authentication system across these frameworks would introduce unnecessary complexity, including:

- shared user schema coupling
- cross-framework identity synchronization
- premature infrastructure complexity
- loss of backend replaceability

At the same time, the business domain still requires a consistent authorization model expressed in terms of domain-relevant roles and permissions.

### Decision

Authentication will be **framework-local**, while authorization will be **domain-driven**.

This means:

- each backend owns its own authentication implementation
- each backend stores its auth tables in its own Postgres schema
- the FieldMark domain does not assume a shared user model
- roles and permissions are defined conceptually at the product level and implemented natively by each framework

Framework-specific auth schemas are:

- `django_auth`
- `dotnet_auth`
- `fiber_auth`

The domain schema must not foreign-key framework auth tables. Instead, domain records store user references as opaque identifiers.

### Consequences

Positive:

- frameworks remain replaceable
- authentication concerns are cleanly isolated
- domain tables remain stable and portable
- Django's built-in auth can be used without forcing symmetry too early
- .NET and Fiber can add auth later without redesigning the domain

Negative:

- there is no shared login experience across frameworks
- user references in domain data are not relationally enforced across auth systems
- role mapping must be maintained separately in each implementation

### Alternatives Rejected

1. Shared users table across all frameworks
   - rejected due to cross-framework coupling and portability loss

2. Centralized identity provider for the demo
   - rejected as out of scope and operationally unnecessary

3. Full auth implementation in all frameworks immediately
   - rejected because it would prematurely harden framework-specific decisions

---

## ADR-013: PostgreSQL Schemas Are Created by Infrastructure Init Scripts

### Status
Accepted

### Context

FieldMark uses a shared PostgreSQL database across multiple backend implementations. The project requires deterministic local setup, clear schema ownership, and strong separation between:

- shared domain infrastructure
- framework-specific auth and support tables

If schemas are created implicitly by application code or migrations, the system becomes fragile:

- startup order matters
- local setup becomes error-prone
- frameworks can accidentally assume infrastructure ownership
- onboarding becomes more complex

Django in particular expects database structures to exist prior to migration, and its defaults can otherwise fall back to the `public` schema.

### Decision

Schemas are treated as **infrastructure** and will be created by PostgreSQL init scripts mounted through Docker Compose.

These scripts are responsible for creating structural invariants such as:

- `domain`
- `django_auth`
- `dotnet_auth`
- `fiber_auth`
- `infra` (optional)

Application frameworks must not create schemas.

Framework migrations may create tables only within their authorized schemas.

### Consequences

Positive:

- local setup becomes deterministic
- `docker compose up` is sufficient to establish schema boundaries
- schema ownership is explicit and reviewable
- future backends can be added without changing startup assumptions
- framework migrations become simpler and safer

Negative:

- infrastructure SQL must be maintained manually
- developers must understand the distinction between schema creation and table migrations

### Alternatives Rejected

1. Create schemas through Django migrations
   - rejected because schemas are infrastructure, not framework-owned data

2. Create schemas through EF Core migrations
   - rejected for the same reason, and because it would bias ownership toward .NET

3. Manual schema creation as a setup step
   - rejected because it degrades developer experience and introduces drift

---

## ADR-014: Shared Domain Schema Is Infrastructure-Owned and Framework-Neutral

### Status
Accepted

### Context

FieldMark is explicitly a multi-backend architecture demonstration. Its shared business domain must remain authoritative regardless of whether the current implementation is accessed via:

- Django
- ASP.NET Core (.NET)
- Go (Fiber)

In single-stack applications, code-first migration systems such as Entity Framework Core and Django ORM migrations are often appropriate. In a multi-backend environment, however, allowing one framework to generate and evolve the shared domain schema would:

- implicitly grant ownership to that framework
- distort naming and constraints around ORM defaults
- weaken architecture neutrality
- create confusion about what is shared vs framework-local

### Decision

The shared FieldMark domain schema is **owned by the architecture**, not by any framework.

Shared domain tables are created and evolved using **infrastructure-level SQL migrations**, not framework migrations.

Examples of shared domain tables include:

- `domain.projects`
- `domain.inspections`
- `domain.inspection_items`
- `domain.violations`
- `domain.corrective_actions`
- `domain.compliance_snapshots`
- `domain.audit_events`

Frameworks map to these tables but do not create or evolve them.

This means:

- EF Core maps to `domain.*` tables
- Django models map to `domain.*` tables
- Fiber data access assumes `domain.*` already exists

### Consequences

Positive:

- the shared business model becomes explicit and reviewable
- framework neutrality is preserved
- backend replaceability is strengthened
- schema drift is reduced
- teaching value of the demo improves significantly

Negative:

- domain SQL must be authored and maintained directly
- ORM auto-migration convenience is reduced for shared business tables
- change management becomes more deliberate

### Alternatives Rejected

1. EF Core owns the shared domain schema
   - rejected because it makes .NET the implicit design authority

2. Django owns the shared domain schema
   - rejected because it makes Django the implicit design authority

3. Dual migration ownership across frameworks
   - rejected because it is brittle, confusing, and architecturally unsound

### Related Rules

- framework auth schemas remain framework-owned
- domain tables must not foreign-key framework auth tables
- framework models must be updated to match infrastructure-owned schema changes

---

## Integration Notes

These ADRs should be cross-referenced with:

- FieldMark Authentication & Authorization Strategy
- FieldMark Docker Compose & Postgres Init Strategy
- FieldMark Shared Domain Schema Ownership Strategy
- existing FieldMark architecture and constraints documents

---

## Status of This Addendum

Drafted and ready for inclusion in the FieldMark ADR set.
