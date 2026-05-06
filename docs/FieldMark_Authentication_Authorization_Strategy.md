# FieldMark Authentication & Authorization Strategy

## Purpose

This document defines the **authentication and authorization (a8n) strategy** for the FieldMark demo across **Django, .NET, and Go (Fiber)** implementations.

It establishes shared architectural intent while explicitly allowing **framework-specific implementations**, in service of the broader FieldMark thesis:

> **One authoritative domain, multiple replaceable application projections.**

This document is normative and should be treated as a foundational architecture reference.

---

## Core Principles

1. **Authentication is framework‑local**
   - Each backend owns its own identity system
   - No shared users table
   - No attempt to unify auth across frameworks

2. **Authorization logic is domain‑driven**
   - Roles and permissions reflect business intent
   - UI frameworks do not invent authorization semantics

3. **Schemas enforce boundaries**
   - Auth data lives in framework‑specific schemas
   - Core domain data lives in a shared schema

4. **Domain data does not foreign‑key framework auth tables**
   - Domain tables store user identifiers as opaque values
   - This keeps the domain portable and framework‑agnostic

---

## Postgres Schema Strategy (Auth‑Related)

The FieldMark database uses schemas to isolate concerns.

```text
Database: fieldmark

Schemas:
- domain        (authoritative business data)
- django_auth   (Django users, groups, permissions)
- dotnet_auth   (ASP.NET Core Identity)
- fiber_auth    (Go/Fiber auth if implemented)
```

Schemas are treated as **infrastructure**, not application behavior.

---

## Role Vocabulary (Shared Conceptual Model)

The following roles are defined conceptually and apply across all backends:

- **Administrator**
  - System configuration
  - Compliance rule changes
  - User management

- **Compliance Officer**
  - Review and resolve violations
  - View audit history

- **Inspector**
  - Perform inspections
  - Record findings

- **Site Supervisor**
  - View project compliance status
  - Respond to violations

- **Executive Viewer**
  - Read‑only access to dashboards and reports

These roles represent **business meaning**, not framework constructs.

Each framework may implement them using its native authorization tools.

---

## Authorization Model

- Authorization is **role‑based**, not attribute‑based
- Permissions are coarse‑grained and explicit
- UI surfaces are protected by backend decisions

Examples:

- Only Administrators can modify compliance rules
- Only Compliance Officers can mark violations resolved
- Inspectors cannot access admin configuration

No client‑side authorization logic is permitted.

---

## Domain ↔ User Reference Strategy

Domain tables (in the `domain` schema) **must not foreign‑key framework auth tables**.

Instead:

- Store user references as opaque identifiers
- Examples:
  - `created_by_user_id`
  - `assigned_to_user_id`
  - `resolved_by_user_id`

These identifiers:
- originate from the active framework’s auth system
- are treated as immutable strings or UUIDs
- are meaningful only to the calling backend

This preserves:
- domain independence
- backend replaceability
- architectural cleanliness

---

## Framework Responsibilities

### Django

- Uses built‑in Django auth and admin
- Auth tables live in `django_auth`
- Django is responsible for mapping roles to business permissions
- Django does not extend or alter other framework schemas

### .NET

- Will use ASP.NET Core Identity when implemented
- Identity tables live in `dotnet_auth`
- Policies/roles align to the shared conceptual role model

### Fiber

- Auth implementation deferred until needed
- When implemented, tables live in `fiber_auth`
- Middleware enforces role checks

---

## Explicit Non‑Goals

- Single sign‑on between frameworks
- Shared users table
- Cross‑framework login
- External IdP integration (for demo scope)

These may be discussed conceptually, but are intentionally out of scope for FieldMark.

---

## Lifecycle Guidance

- Authentication is **architecturally decided now**
- Framework‑specific implementations are **added incrementally**
- Django auth exists earliest due to framework design
- .NET and Fiber auth are introduced when feature work begins

This avoids premature commitment while preventing late‑stage bolting‑on.

---

## Status

Accepted – FieldMark Authentication & Authorization Strategy
