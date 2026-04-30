# Architectural Decision Record (Updated)

## Project
Construction Compliance & Inspection Management System (CCIMS)

## Purpose
This ADR records foundational architectural decisions that intentionally constrain the solution space. These decisions are designed to support backend authority, reduce cognitive load, and align .NET and Django implementations so that architecture—not framework favoritism—is the focus of comparison.

This ADR is intentionally written to act as a **primary architectural constraint source** for downstream spec‑driven and agentic design systems.

---

## Architectural Goals

- Establish **backend authority** as the single source of truth
- Minimize duplicated logic between UI and server
- Ensure architectural **parity** between .NET and Django stacks
- Reduce accidental complexity introduced by framework patterns
- Enable spec‑driven and agent‑assisted design without pattern drift

---

## ADR‑011: ORM‑First, Rich Domain Model (No CQRS / No Repositories)

### Decision

The system shall use an **ORM‑first architecture with rich domain models**, avoiding CQRS, MediatR, generic repository abstractions, and layered domain service architectures.

Business rules, invariants, and state transitions are implemented directly on ORM domain entities. Request handlers (Razor Pages / Django views) orchestrate workflows but do not own domain logic.

---

### Rationale

- Django’s architectural model centralizes behavior on models and request handlers
- EF Core supports rich entities with encapsulated invariants and lifecycle logic
- CQRS and repository patterns introduce abstraction asymmetry between stacks
- Eliminating architectural indirection reinforces backend authority
- Reduces the number of conceptual layers contributors must understand

---

### Consequences

- Domain entities are not anemic data containers
- DbContext / Django ORM are accessed directly (no repository pass‑throughs)
- Validation logic is implemented once, server‑side
- Architectural explanations during demos are simplified

---

### Explicitly Rejected Alternatives

- CQRS (Command / Query segregation)
- MediatR or message‑based pipelines
- Generic Repository pattern layered over EF Core
- Clean Architecture / Onion Architecture layering

These were rejected to prevent architectural complexity from overshadowing the demo’s core message.

---

## Relationship to Other ADRs

- Complements ADR‑003 (Django selection)
- Complements ADR‑006 (Database‑first modeling)
- Reinforces ADR‑001 (Server‑driven UI)

---

## Status

Accepted – foundational constraint

