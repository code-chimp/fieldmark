# FieldMark Shared Domain Schema Ownership Strategy

## Purpose

This document defines **how the shared FieldMark domain schema is owned, created, evolved, and consumed** in a multi‑backend environment consisting of:

- Django
- ASP.NET Core (.NET)
- Go (Fiber)

Its primary goal is to ensure that:

> **The FieldMark business domain is authoritative, framework‑neutral, and stable across implementations.**

This document is normative and should be used to guide database design, migrations, framework configuration, and future architectural decisions.

---

## Core Architectural Principles

1. **The domain is not owned by a framework**
   - No single backend (Django, .NET, Go) is the system of record for the domain schema
   - Frameworks must conform to the domain, not define it

2. **The database schema is part of the architecture**
   - The structure of the domain tables expresses business intent
   - It must be explicit, reviewable, and stable

3. **Infrastructure precedes application code**
   - Schemas and shared tables are created before any application runs
   - Framework migrations operate only within pre‑defined boundaries

4. **Frameworks are replaceable projections**
   - Any backend may be added, removed, or rewritten
   - The shared domain schema remains intact

---

## Database Ownership Model

### Database

```text
Database: fieldmark
```

A single Postgres database is used for the demo to keep deployment and local setup simple.

---

## Schema Ownership

```text
Schemas:
- domain        (authoritative business data)
- django_auth   (Django authentication & admin)
- dotnet_auth   (ASP.NET Core Identity)
- fiber_auth    (Go/Fiber auth if implemented)
- infra         (optional: migrations metadata, audit infra)
```

### Ownership Rules

- `domain` schema is **owned by the architecture**, not by a framework
- `*_auth` schemas are **owned by their respective frameworks**
- Frameworks must never create or alter schemas

---

## Shared Domain Tables

### Examples (non‑exhaustive)

```text
domain.projects
domain.inspections
domain.inspection_items
domain.violations
domain.corrective_actions
domain.compliance_snapshots
domain.audit_events
```

These tables:

- represent core business concepts
- exist independently of backend implementation
- may be accessed by multiple frameworks

---

## How Shared Domain Tables Are Created

### Authoritative Method

✅ **Infrastructure‑level SQL migrations**

Shared domain tables are created using:

- manually authored SQL files
- reviewed and committed as part of infrastructure
- executed via Docker / database bootstrap tooling

They are **not created by**:

- Django migrations
- Entity Framework Core migrations
- Go ORM tooling

---

## Why Not Code‑First for the Shared Domain

While code‑first approaches (EF Core, Django ORM) are productive in single‑stack applications, they are inappropriate for a shared multi‑backend domain because:

- they implicitly grant ownership to one framework
- naming and constraints reflect ORM defaults rather than business intent
- parallel frameworks cannot safely co‑evolve the schema
- the architecture becomes harder to reason about and explain

For FieldMark, **clarity and neutrality outweigh the convenience of auto‑generated schemas**.

---

## Acceptable Use of Framework ORMs

Framework ORMs are still encouraged for:

- mapping to the shared domain schema
- querying and persistence
- validation and invariants at the application level

### .NET

- EF Core entities map explicitly to `domain.*` tables
- Schema and table names are configured via mapping
- Auto‑migrations are disabled or scoped to framework‑owned schemas

### Django

- Django models may represent `domain.*` tables
- Django migrations do **not** create or modify those tables
- Managed = False (or equivalent discipline) is used where appropriate

### Go / Fiber

- Data access code assumes the domain schema exists
- SQL or query tooling treats the domain schema as authoritative

---

## Domain ↔ Authentication Boundary

Domain tables:

- do not foreign‑key framework auth tables
- do not assume a specific user model

Instead, they store user references as opaque identifiers:

```text
created_by_user_id
assigned_to_user_id
resolved_by_user_id
```

This ensures:

- the domain remains framework‑agnostic
- authentication systems are replaceable
- migrations remain simpler and safer

---

## Change Management for the Domain Schema

### Making a Change

1. Update the domain ERD and documentation
2. Author a new infrastructure SQL migration
3. Review for cross‑framework impact
4. Update framework models/mappings
5. Apply via Docker / DB bootstrap process

### Prohibited Changes

- silent schema drift via ORM auto‑migration
- framework‑specific extensions to domain tables
- breaking changes without coordinated updates

---

## Developer and Agent Rules

All contributors (human or agent) must follow these rules:

1. Do not generate shared domain tables via framework migrations
2. Do not modify domain tables from application code
3. Treat the domain schema as a contract
4. Update mappings when the domain changes
5. Escalate unclear ownership decisions

---

## Architectural Benefits

This strategy:

- reinforces backend authority
- keeps the demo credible and teachable
- avoids accidental framework lock‑in
- simplifies adding new backend implementations
- mirrors real‑world enterprise architecture practices

---

## Status

Accepted – FieldMark Shared Domain Schema Ownership Strategy
