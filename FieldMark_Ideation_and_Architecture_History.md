# FieldMark Architecture Ideation & Design History

## Purpose of This Document

This document captures the **historical context, rationale, and key architectural decisions** that emerged during the initial ideation and foundation‑building phase of the FieldMark project.

It is intended to serve as:
- A shared memory of *why* decisions were made
- A grounding artifact for future contributors and agentic systems
- Context for evaluating future trade‑offs without re‑litigating fundamentals

FieldMark is treated not as a demo toy, but as a **reference implementation** demonstrating backend authority, architectural restraint, and cross‑stack parity.

---

## Core Problem Being Addressed

Modern enterprise applications frequently suffer from:
- Split authority between frontend and backend
- Duplication of business rules across stacks
- Excessive architectural ceremony (CQRS, mediators, service layers)
- High cognitive load driven by SPA‑centric designs

The FieldMark project explores an alternative:

**A server‑authoritative system where the backend owns truth, workflows, and validation, while the UI acts as a projection of state.**

This is illustrated using two parallel implementations:
- ASP.NET Core Razor Pages + EF Core
- Django + Django ORM

---

## Guiding Architectural Principles

The following principles shape *all* downstream decisions:

1. **Backend Authority**  
   Business rules, state transitions, and validation live on the server.

2. **Single System of Record**  
   One PostgreSQL database shared across implementations.

3. **Conceptual Parity, Not Idiomatic Symmetry**  
   .NET and Django differ mechanically but align architecturally.

4. **Intentional Constraint**  
   Patterns are rejected unless they earn their complexity.

5. **Agent‑First Architecture**  
   Structure enforces discipline better than instructions alone.

---

## Technology & Stack Decisions

### Backend Frameworks

- **ASP.NET Core Razor Pages**  
  Chosen for page‑centric, server‑driven workflows with minimal ceremony.

- **Django**  
  Chosen for its opinionated ORM, built‑in admin, and enterprise credibility.

Flask was explicitly rejected due to lack of enforced structure and parity.

---

### Persistence

- **PostgreSQL (v17 preferred)**  
  Neutral, enterprise‑grade, well‑supported by EF Core and Django.

- Postgres runs in Docker and is treated as shared infrastructure.

- Database is the **authoritative system of record**.

---

### UI Strategy

- **Server‑Rendered HTML + HTMX**
- No SPA frameworks
- Minimal JavaScript islands (e.g., AG Grid)
- Third‑party JS libraries are vendored and served locally (no public CDN dependency)

---

## Domain Modeling Philosophy

A shared domain model was designed before feature implementation.

Key domain concepts:
- Project
- Inspection
- InspectionItem
- Violation
- CorrectiveAction
- ComplianceSnapshot
- AuditEvent

Domain modeling emphasizes:
- Explicit workflow state
- Materialized compliance snapshots
- Append‑only audit trails
- UUID primary keys for parity and resilience

(User identity, permissions, attachments, and notifications were explicitly deferred.)

---

## .NET Solution Architecture

### Project Structure

```
FieldMark.Web     – Razor Pages UI & composition root
FieldMark.Domain  – Entities, invariants, business behavior
FieldMark.Data    – EF Core persistence
```

### Dependency Direction (Hard Rule)

Permitted:
- Web → Domain
- Web → Data
- Data → Domain

Forbidden:
- Domain → Data
- Domain → Web

Domain is dependency‑free and NuGet‑free by design.

---

### Explicitly Rejected .NET Patterns

- CQRS
- MediatR
- Repository pattern
- Unit‑of‑work abstractions
- Clean / Onion / Hexagonal layering
- SPA‑first backends

These were rejected to avoid fragmentation of authority and unnecessary indirection.

---

## Django Architecture Guardrails

Django is treated as an **adapter**, not a design authority.

Key constraints:
- Business logic must live in model methods or explicit domain modules
- Views orchestrate; they do not contain rules
- Django signals are prohibited
- Django migrations own only framework tables (auth, admin)
- Domain schema evolution is owned elsewhere

Django Admin is considered **platform tooling**, not product UX.

---

## Schema & Migration Ownership

- EF Core migrations own the domain schema
- Django migrations manage Django’s internal tables only
- Dual ownership of domain tables is forbidden
- Schema is treated as a shared contract between stacks

---

## Authentication Strategy

Authentication is **deliberately deferred**.

Rationale:
- Early auth scaffolding introduces irreversible schema and structure decisions
- Django’s built‑in auth is used only for admin/platform needs
- No identity sharing between .NET and Django
- External IdP integration is out of scope for the demo phase

---

## Front‑Loading “Plumbing” Work

Before agentic development begins, the following groundwork was prioritized:

- Monorepo layout finalized
- Dockerized shared Postgres running
- ORM plumbing validated in both stacks
- Static asset strategy finalized (no CDN dependency)
- Base layout templates created
- Domain ERD created and checked in
- Architectural guardrail documents written

This was done to **avoid burning agent tokens on trivial or reversible work**.

---

## Agentic Development Strategy

The project is designed for **spec‑driven, agent‑assisted development** using systems like BMAD.

Key strategies:
- Encode constraints in structure, not prose
- Provide reference documents as priming artifacts
- Start agents on narrow, safe tasks
- Defer feature complexity until rules are proven stable

Human effort is reserved for irreversible decisions; agents handle reversible implementation work.

---

## Current Project State

As of this checkpoint:

- Architecture is locked
- Domain model is defined
- Infrastructure is running
- No business features have been implemented
- The project is ready for BMAD kickoff

This document represents the **baseline historical context** against which all future work should be evaluated.

---

## Status

Accepted – foundational context document
