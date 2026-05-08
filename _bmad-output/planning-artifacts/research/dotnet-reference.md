# FieldMark .NET Solution Architecture & Design Reference

## Purpose
This document is a **reference architecture and guardrail specification** for the FieldMark .NET solution. It is designed to serve as *priming context* for an implementing agent (e.g., BMAD, coding agents, architectural agents) so that enforcement rules, skills, and constraints can be derived mechanically and consistently.

This document intentionally encodes architectural intent through **structure, dependency direction, and explicit prohibitions**, not just guidelines.

---

## Architectural Intent (Non‑Negotiable)

FieldMark is a **server‑authoritative, domain‑centric system**.

The .NET implementation must:
- Centralize business rules in the domain
- Treat persistence and UI as adapters
- Avoid client‑side or framework‑driven orchestration
- Remain symmetrical (conceptually) with the Django and Go (Fiber) implementations

Complexity is opt‑in and must be justified explicitly.

---

## Solution Layout

```
FieldMark.sln
src/
  FieldMark.Web     (ASP.NET Core Razor Pages)
  FieldMark.Domain  (Domain model and behavior)
  FieldMark.Data    (EF Core persistence)
```

This structure is intentional and must not be flattened or inverted.

---

## Project Responsibilities

### FieldMark.Domain

Role:
- Business concepts and behavior

Contains:
- Entities and aggregate roots
- Value objects
- Domain enums
- Domain invariants
- State transition methods
- Domain exceptions

Must NOT contain:
- Entity Framework references
- Data annotations for persistence
- Serialization attributes
- Validation frameworks
- External NuGet packages

Dependency rule:
- **No outbound project references**

---

### FieldMark.Data

Role:
- Persistence adapter mapping to the infrastructure-owned `domain.*` schema and to the framework-local `dotnet_auth.*` schema

Contains:
- DbContext
- DbSet declarations
- EF Core configuration (fluent mappings, value converters, naming conventions)
- Mappings of domain entities to existing `domain.*` tables
- Database-specific concerns

Must NOT contain:
- Business rules
- Workflow logic
- UI logic
- Migrations that create or alter tables in the `domain` schema (see ADR-014)

Dependencies:
- References FieldMark.Domain

Allowed NuGet packages:
- Microsoft.EntityFrameworkCore
- Microsoft.EntityFrameworkCore.Design
- Npgsql.EntityFrameworkCore.PostgreSQL
- EFCore.NamingConventions (for `UseSnakeCaseNamingConvention()`)

---

### FieldMark.Web

Role:
- Composition root and server-rendered UI

Contains:
- Razor Pages (.cshtml + PageModel)
- Request orchestration
- Dependency injection configuration
- EF Core registration
- HTMX interaction points

Must NOT contain:
- Business rules
- Persistence logic beyond orchestration
- Repository abstractions
- CQRS pipelines

Dependencies:
- References FieldMark.Domain
- References FieldMark.Data

Allowed NuGet packages (initial):
- Microsoft.EntityFrameworkCore.Tools
- Npgsql.EntityFrameworkCore.PostgreSQL

---

## Dependency Direction (Hard Rule)

Permitted:
```
Web  → Domain
Web  → Data
Data → Domain
```

Forbidden:
```
Domain → Data
Domain → Web
```

If a dependency violates this direction, it is architecturally invalid.

---

## Database & Persistence Policy

- PostgreSQL is the sole persistence engine.
- The database is the **system of record**.
- The `domain` schema is **infrastructure-owned** (ADR-014). It is created by the Postgres init scripts in `docker/postgres/init/` and evolved by hand-authored infrastructure SQL. EF Core does **not** own, create, or migrate `domain.*` tables.
- EF Core migrations are scoped exclusively to the `dotnet_auth` schema (ADR-012). Any migration that touches `domain.*` is a defect.
- EF Core entities map to the existing `domain.*` tables via explicit fluent configuration:
  - `optionsBuilder.UseSnakeCaseNamingConvention()` is required globally.
  - Each entity uses `ToTable("project", "domain")` (or equivalent) so the schema and snake_case table name match the infrastructure-defined schema exactly.
  - Enum properties use `HasConversion<string>()` to match the `SCREAMING_SNAKE_CASE` storage form (see `domain-model.md` §9).
- No dual ownership of domain tables is allowed across stacks; `EnsureCreated()` and ad-hoc DDL against `domain.*` are forbidden.

---

## DbContext Registration Rule

- DbContext is defined in FieldMark.Data
- DbContext is registered in FieldMark.Web
- DbContext must never be referenced by FieldMark.Domain

Example location of registration:
- Program.cs (Web project)

---

## Authentication & Authorization Policy

Authentication is **framework-local** (ADR-012). The .NET stack will use ASP.NET Core Identity when it adopts authentication; Identity tables live in the `dotnet_auth` schema, which is the only schema EF Core migrations are permitted to touch.

Rules:
- Authentication must not be scaffolded during project creation; it is added deliberately as a later milestone.
- When ASP.NET Core Identity is introduced, all Identity tables map to `dotnet_auth.*` (e.g. `optionsBuilder.UseSnakeCaseNamingConvention()` plus explicit `ToTable("aspnet_users", "dotnet_auth")` and equivalents). Identity schemas must never spill into `domain` or any other framework's auth schema.
- Domain tables in `domain.*` must not foreign-key any `dotnet_auth.*` table. User references on domain rows are stored as opaque UUIDs (e.g. `created_by_user_id`).
- Authorization policies and roles align with the shared conceptual role vocabulary (Administrator, Compliance Officer, Inspector, Site Supervisor, Executive Viewer); the implementation is native ASP.NET Core authorization.
- Identity systems must not be shared across stacks. Django auth lives in `django_auth`, Fiber auth in `fiber_auth`.

---

## Explicitly Rejected Patterns

The following are prohibited unless an ADR explicitly reverses the decision:

- CQRS
- MediatR
- Repository pattern
- Unit‑of‑work abstractions
- Clean / Onion / Hexagonal layering
- Client‑side state management
- API‑first SPA backends

---

## Agent Guardrail Rules (Derivable)

An implementing agent must:
- Reject additions of EF Core to Domain
- Reject Domain references to Data or Web
- Reject introduction of repositories or mediators
- Reject any EF Core migration that creates, alters, or drops a table in the `domain` schema (ADR-014)
- Reject any model configuration that omits explicit `ToTable("...", "domain")` or that relies on `EnsureCreated()` for `domain.*`
- Reject FK relationships from `domain.*` entities to `dotnet_auth.*` Identity entities (ADR-012)
- Prefer modifying structure over adding abstractions
- Fail designs that require explanation of architectural patterns

If an agent solution requires explaining *why* it is structured a certain way, it is likely invalid.

---

## Core Architectural Principle

**The domain owns behavior.**
Everything else adapts to it.

---

## Status

This document defines the **baseline, locked architecture** for FieldMark’s .NET implementation and is intended to be used as a priming artifact for agentic design and enforcement systems.
