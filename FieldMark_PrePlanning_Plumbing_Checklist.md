# FieldMark Pre‑Planning Plumbing Checklist

## Purpose

This checklist enumerates the **foundational plumbing work** that should be completed **before** kicking off formal project planning, BMAD workflows, or agent‑driven feature development for FieldMark.

Completing these items up front:
- Minimizes token burn on trivial or reversible work
- Reduces architectural drift
- Gives agentic systems strong structural signals
- Ensures agents focus on *features*, not environment setup

This checklist is intentionally opinionated and reflects the decisions captured in the accompanying architecture and history documents.

---

## 1. Repository & Structure

✔ Monorepo root created and committed

✔ Top‑level folders present:
- dotnet/ (or FieldMark/)
- fieldmark_py/
- docker/
- docs/

✔ .NET solution scaffolded:
- FieldMark.Web (Razor Pages)
- FieldMark.Domain
- FieldMark.Data

✔ Django project initialized under fieldmark_py/

✔ Empty but intention‑revealing folders created (Domain/Entities, Django apps, etc.)

✔ .gitignore configured for:
- build artifacts
- virtual environments
- postgres_data/

---

## 2. Infrastructure & Environment

✔ docker‑compose.yml present at repo root

✔ PostgreSQL container running successfully

✔ Postgres version chosen and pinned (v17 preferred)

✔ Shared database name/user resolved and documented

✔ Both .NET and Django verified to connect to Postgres

✔ No application containers added yet (DB only)

---

## 3. Dependency Plumbing

### .NET

✔ FieldMark.Domain has **no NuGet packages**

✔ FieldMark.Data includes:
- Microsoft.EntityFrameworkCore
- Microsoft.EntityFrameworkCore.Design
- Npgsql.EntityFrameworkCore.PostgreSQL

✔ FieldMark.Web includes:
- Microsoft.EntityFrameworkCore.Tools
- Npgsql.EntityFrameworkCore.PostgreSQL

✔ DbContext defined in Data

✔ DbContext registered in Web only

✔ No repositories, CQRS, or mediator packages installed

---

### Django

✔ uv project initialized

✔ Runtime dependencies installed:
- django
- psycopg[binary]

✔ Dev dependencies installed:
- ruff
- black
- mypy
- django‑stubs

✔ Django connects to shared Postgres

✔ Initial Django migrations applied (framework tables only)

✔ No domain schema migrations applied yet

---

## 4. Static Assets & UI Plumbing

✔ Decision made to avoid public CDNs

✔ Third‑party JS vendoring strategy finalized:
- LibMan → wwwroot (ASP.NET)
- Vendored files → static/ (Django)

✔ HTMX local copy added to both projects

✔ Base layout templates created in both stacks

✔ Shared visual structure agreed (navigation, slots, placeholders)

---

## 5. Domain & Data Model Preparation

✔ Core domain boundaries agreed (Project, Inspection, Violation, etc.)

✔ UUID primary key strategy chosen

✔ UTC‑only datetime policy agreed

✔ Status/state fields modeled explicitly

✔ Domain ERD authored and checked in

✔ Out‑of‑scope domains explicitly listed (Auth, Attachments, Notifications)

---

## 6. Architectural Guardrails (Documents)

✔ .NET Architecture & Design Reference created

✔ Django Architecture & Guardrails Reference created

✔ Cross‑stack architectural decisions documented

✔ Explicitly rejected patterns listed (CQRS, signals, repos, etc.)

✔ Authentication explicitly deferred and documented

✔ Migration ownership policy documented

---

## 7. Agent Readiness

✔ Architecture documents stored in repo for priming

✔ Domain ERD available for agent reasoning

✔ Naming and convention decisions locked

✔ Negative space documented (what will NOT be built)

✔ Foundation commit/tag created

---

## 8. Explicitly NOT Done Yet (By Design)

The following should remain **intentionally incomplete** prior to planning:

- No domain feature logic implemented
- No EF Core migrations for domain tables
- No Django domain migrations
- No authentication or Identity
- No background jobs
- No APIs beyond page workflows
- No notification systems

Agents should handle these **after** planning rules are in place.

---

## Completion Criteria

When all checked items above are complete, the project is considered:

✅ **Plumbing‑complete**
✅ **Architecture‑locked**
✅ **Ready for BMAD or agentic planning**

At this point, agents may be safely tasked with feature delivery.

---

## Status

Foundation Checklist – Accepted
