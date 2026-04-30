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
- Remain symmetrical (conceptually) with the Django implementation

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
- Persistence adapter

Contains:
- DbContext
- DbSet declarations
- EF Core configuration
- Mappings
- Database-specific concerns

Must NOT contain:
- Business rules
- Workflow logic
- UI logic

Dependencies:
- References FieldMark.Domain

Allowed NuGet packages:
- Microsoft.EntityFrameworkCore
- Microsoft.EntityFrameworkCore.Design
- Npgsql.EntityFrameworkCore.PostgreSQL

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

- PostgreSQL is the sole persistence engine
- Database is the **system of record**
- EF Core migrations own domain schema evolution
- Django migrations own only Django-specific tables

No dual ownership of domain tables is allowed.

---

## DbContext Registration Rule

- DbContext is defined in FieldMark.Data
- DbContext is registered in FieldMark.Web
- DbContext must never be referenced by FieldMark.Domain

Example location of registration:
- Program.cs (Web project)

---

## Authentication & Authorization Policy

Current status: **Deferred by design**

Rules:
- Authentication must not be scaffolded during project creation
- ASP.NET Core Identity must not be introduced until:
  - Domain schema is stable
  - Migration ownership is explicit
  - Architectural rules are locked

Django auth/admin is treated as platform tooling only.

Identity systems must not be shared across stacks for this project.

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
