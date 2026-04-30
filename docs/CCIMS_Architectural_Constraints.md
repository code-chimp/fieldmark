# Architectural Constraints & Design Guardrails

## Purpose
This document defines **hard architectural constraints** that downstream design agents, spec‑driven systems (BMAD, SpecFlow), and human contributors must adhere to.

It operationalizes intent from the ADRs into enforceable rules that deliberately narrow the solution space.

This document is designed to be directly consumable by agentic systems.

---

## Global Architectural Principles

- One system of record
- Backend owns all business rules and validation
- UI layers are projections, not authorities
- Complexity must be earned
- Architectural symmetry across stacks is mandatory

---

## .NET Implementation Constraints

### Allowed

- ASP.NET Razor Pages
- EF Core with direct DbContext usage
- Rich domain entities with behavior
- Minimal APIs for data‑only endpoints
- HTMX for incremental UI updates

### Disallowed

- CQRS patterns (commands/queries, handler pipelines)
- MediatR or equivalent mediator frameworks
- Generic or abstract repository layers
- Anemic entity models
- Client‑side validation as a source of truth

---

## Django Implementation Constraints

### Allowed

- Django ORM as primary data access
- Business rules defined on models and forms
- Django views as workflow orchestrators
- Django templates with HTMX

### Disallowed

- Service layers duplicating model behavior
- Client‑side state machines
- Multiple competing domain representations

---

## Data & Persistence Constraints

- PostgreSQL is the canonical datastore
- Schema is treated as a shared contract
- Migrations must preserve cross‑stack compatibility
- No stack‑specific database features without parity consideration

---

## UI & Interaction Constraints

- HTMX is the primary mechanism for interactivity
- JavaScript is limited to UI islands (AG Grid)
- No frontend state stores (Redux, Signals, Stores, etc.)
- Navigation and workflow transitions are server‑driven

---

## Agentic Design Guardrails

Any agent or automated design system must:

- Prefer existing domain behavior over new abstractions
- Reject solutions requiring CQRS, repositories, or mediators
- Centralize validation on the server
- Optimize for understandability over extensibility

If a solution requires explaining a pattern, that solution is likely invalid.

---

## Reference

See ADR‑011 for architectural rationale.

