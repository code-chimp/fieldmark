# Project Brief — FieldMark

**Project codename:** FieldMark
**Domain:** Construction Compliance & Inspection Management System (CCIMS)
**Document type:** BMAD Project Brief
**Status:** Draft v1
**Owner:** Tim Goshinski

---

## 1. Executive Summary

FieldMark is a reference implementation of an enterprise-grade Construction Compliance & Inspection Management System, built specifically to demonstrate that server-driven web architecture — Razor Pages + HTMX on .NET, and Django + HTMX on Python — can deliver SPA-equivalent interactivity without the cognitive and architectural overhead of a single-page application.

The product itself is realistic: project managers, compliance officers, and site supervisors track inspections, record violations, and resolve corrective actions against a server-evaluated rules engine. The architectural thesis is the actual deliverable: that backend authority over workflow, validation, and state — combined with HTML over the wire and judiciously placed JavaScript islands (AG Grid) — produces an application that is faster to build, easier to reason about, and cheaper to maintain than the SPA alternatives it will be compared against.

FieldMark is intentionally not a product seeking product-market fit. It is a teaching artifact for an upcoming talk on HTMX, with the secondary goal of seeding parity comparisons against Angular, React, and Flask/Django implementations of the same domain.

---

## 2. Problem Statement

### The architectural problem

Modern enterprise web applications routinely default to SPA architectures (Angular, React) regardless of whether the application's interaction patterns actually require client-owned state. This default carries durable, compounding costs:

- **Duplicated business rules** between server validation and client validation, which drift over time
- **Two systems of record** (server domain + client store) that must be kept synchronized
- **Cognitive surface area** spread across build tools, state management libraries, routing libraries, data-fetching libraries, and component frameworks
- **Onboarding cost** for contributors who must understand the patterns of both halves

For applications whose core interactions are fundamentally request-response — submit an inspection, resolve a violation, view a dashboard — the SPA default trades architectural complexity for marginal interactivity gains.

### The demonstration problem

Architectural arguments are unpersuasive in the abstract. Audiences need to see a non-trivial, realistic application that exhibits SPA-level interactivity without SPA-level architecture. Existing HTMX demos tend to be either too small (todo lists) to be credible at enterprise scale, or too narrowly scoped to demonstrate workflow orchestration, rule enforcement, and rich data interactions in a single coherent system.

FieldMark closes that gap with a domain rich enough to require state machines, rule evaluation, role-based access, and dense data grids — without being so broad that the architectural message gets lost.

---

## 3. Proposed Solution

A web application implementing the CCIMS domain twice with strict architectural symmetry:

- **.NET stack:** ASP.NET Core, Razor Pages, EF Core with rich domain entities, HTMX for interactivity, AG Grid as a JS island.
- **Python stack:** Django, Django ORM with rich models, Django templates with HTMX, AG Grid as a JS island.
- **Shared:** PostgreSQL schema, identical workflows, identical state machines, identical rule semantics.

The two stacks are built in parallel against the same epic and story sequence, producing a real-time parity comparison rather than a port done after the fact. Both stacks reject CQRS, MediatR, generic repositories, layered service architectures, and client-side state stores per ADR-011.

The architectural thesis is operationalized at every layer: the server owns the rules engine, the server owns workflow transitions, the server owns validation, the server owns the audit trail. The client requests HTML.

---

## 4. Target Users

### Audience for the talk (true primary users of the artifact)

Mid-to-senior software engineers, tech leads, and architects evaluating whether HTMX is appropriate for their context. They are skeptical, have built or maintained SPAs, and want to see something credible at enterprise scale before they take the architectural argument seriously.

### Personas inside the application's domain

| Persona | Role in the product | Primary interactions |
|---|---|---|
| Project Manager | Monitors compliance status across a portfolio of construction projects | Dashboard drill-down, project list filtering, inspection scheduling oversight |
| Compliance Officer / Inspector | Performs inspections and records violations on site | Inspection workflow, violation capture, severity classification |
| Site Supervisor | Resolves violations and submits corrective action evidence | Violation detail, evidence submission, resolution requests |
| Executive / Oversight | Reads-only view of risk and trend data | Dashboard, reports, no write operations |
| Reference-Data Admin | Maintains inspection codes, violation categories, rule metadata | Admin-only CRUD; minimal UX (Django Admin / Razor admin pages) |

The personas exist to make workflows realistic. They are not the audience the demo is built for.

---

## 5. Goals & Success Metrics

### Project goals (talk outcomes)

1. Convince a skeptical engineering audience that HTMX is viable at enterprise scale.
2. Make the architectural cost of SPA defaults legible by direct contrast.
3. Produce a reusable reference implementation other teams can clone and study.

### Product goals (in-application)

1. Demonstrate SPA-level reactivity using server-rendered HTML and HTMX partial swaps.
2. Centralize all domain logic, workflows, and state management on the backend.
3. Show rich third-party JS controls (AG Grid) integrated as islands without ceding architectural authority.
4. Maintain a non-trivial, realistic domain that holds up to scrutiny.

### Success metrics

| Metric | Target |
|---|---|
| Client-side JavaScript files written by hand | ≤ 5, all narrowly scoped (AG Grid wiring, minor UX helpers) |
| Business rules duplicated between client and server | 0 |
| Lines of state-management code (Redux/NgRx/Pinia equivalents) | 0 |
| HTTP requests directly traceable to user interactions | 100% |
| Architectural delta between .NET and Django implementations | Limited to language idioms and framework-specific syntax; no structural divergence |
| Onboarding time to add a new feature, measured against a contributor unfamiliar with the project | Comparable for both stacks |

These are explicitly architectural success metrics, not user satisfaction metrics. The application's "users" are simulated.

---

## 6. MVP Scope

### Must-have (in scope for the talk)

- Project lifecycle management (create, view, list, lifecycle states)
- Inspection scheduling and workflow execution
- Violation capture, lifecycle, and resolution
- Corrective action submission with rule-gated approval
- Compliance rules engine with server-side evaluation and recalculation
- Compliance dashboard with HTMX partial refresh and drill-down
- AG Grid integration with server-side row model on at least two views
- Audit log per project, immutable, append-only
- Role-based access control covering all four primary personas
- PostgreSQL schema with EF Core / Django migrations preserving cross-stack parity

### Should-have (if time permits)

- Reference data administration (inspection codes, violation categories)
- Multi-stack parity test suite that runs the same scenarios against both implementations

### Out of scope (initial phase)

- Mobile-native applications
- Offline-first behavior
- Real-time collaboration (multi-user simultaneous editing)
- Production-quality file uploads (placeholders only)
- Notification system (email, webhooks, in-app)
- GIS / site mapping
- Advanced analytics or BI integrations

### Future enhancements (post-talk)

- Regulatory rule configuration UI
- Notification system
- Parallel implementations in Angular and React for direct comparison

---

## 7. Technical Considerations (High-Level)

Detailed technical decisions live in `architecture-decisions.md`. At the brief level, the constraints that matter are:

- Backend authority is non-negotiable. Any solution that places business rules on the client is rejected.
- ORM-first, rich domain model. No CQRS, no repositories, no MediatR, no Clean/Onion layering.
- HTMX is the only mechanism for client-server interactivity beyond AG Grid's data fetching.
- No frontend state stores under any circumstances.
- PostgreSQL is the canonical datastore. Migrations must remain compatible across both stacks.
- Architectural symmetry between .NET and Django implementations is mandatory.

---

## 8. Constraints & Assumptions

### Constraints

- The talk fixes the timeline; scope must compress to fit, not the other way around.
- Both stacks must reach feature parity at every story boundary, not at the end.
- The architectural rules in `architecture-decisions.md` are non-negotiable. BMAD agents must reject solutions that violate them rather than relaxing the constraints.
- A single contributor (the author) is doing the implementation; agent-assisted development is expected.

### Assumptions

- The audience is sophisticated enough to follow architectural arguments without simplification.
- HTMX 1.x or 2.x is stable enough to be the primary interactivity layer in production-shaped code.
- AG Grid's server-side row model is mature on both .NET and Django backends.
- PostgreSQL feature usage will stay within the cross-stack-portable subset.

---

## 9. Risks & Open Questions

### Risks

| Risk | Mitigation |
|---|---|
| Demo overload — too many features, unclear narrative | Pre-script a single anchor workflow (resolve a violation) as the spine of the talk |
| SPA-mimicry — accidentally building HTMX in a way that hides authority instead of exposing it | Keep interactions HTTP-visible; use `hx-get`/`hx-post` directly rather than over-abstracting |
| Stack divergence — one stack pulls ahead and the other becomes a port | Story-level parity gates: a story is not done until both stacks pass it |
| Rule engine over-engineering — the rules become the project | Cap rule complexity at "required inspections passed before a project can close" plus "violation must be Resolved before close"; nothing combinatorial |
| Domain over-modeling — entities multiply | Hold to the entity list in `domain-model.md`; new entities require explicit justification |
| AG Grid as a wedge — JS island grows into a JS app | Keep AG Grid configuration declarative and server-fed; no client-side row computation |

### Open questions

- Will the talk include a live coding segment, or is everything pre-built? (Affects how heavily the dev workflow is rehearsed.)
- Do we need a hosted demo environment, or is local-only sufficient?
- Should the audit log be append-only at the database level (insert-only constraints) or only enforced at the application level?
- Are we committing to the same PostgreSQL version on both stacks, and which version?
- Authentication: do we mock identity for the demo or wire up actual OIDC against a local IdP?

---

## 10. Next Steps

1. Review and approve this brief.
2. Validate `domain-model.md` against the brief — entities, lifecycles, and rules align.
3. Validate `architecture-decisions.md` against the brief — constraints and ADR align.
4. Run the spec-driven workflow (BMAD) against the `_bmad-output/planning-artifacts/research/` folder to generate PRD, architecture, and epics.
5. Begin sprint planning from the generated epic breakdown, parity-first.

---

## Appendix A — Source Document Inventory

| Document | Role | Status |
|---|---|---|
| `prd.md` | Original PRD; authoritative for product scope | Existing |
| `architecture-decisions.md` | ADR-011 + hard constraints for downstream agents | Existing |
| `ux-guide.md` | Screen inventory, UX principles | Existing |
| `project-brief.md` | This document | Existing |
| `domain-model.md` | Entities, state machines, schema, ERD | Existing |
| `dotnet-reference.md` | .NET project structure, patterns, agent guardrails | Existing |
| `django-reference.md` | Django project structure, patterns, agent guardrails | Existing |
| `archive/architecture.md` | Prior architecture doc (reference only; BMAD will regenerate) | Archived |
| `archive/epics.md` | Prior epics doc (reference only; BMAD will regenerate) | Archived |
