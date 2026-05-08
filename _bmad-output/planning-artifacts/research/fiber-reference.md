# FieldMark Go (Fiber) Architecture & Guardrails Reference

## Purpose

This document defines the **architectural intent, constraints, and guardrails** for the Go (Fiber) implementation of FieldMark. It is designed to be used as **priming context for agentic systems** (BMAD, coding agents, architectural agents) so the Fiber stack remains aligned with FieldMark's backend-authority philosophy and stays symmetrical in intent with the .NET and Django implementations.

This document is a peer to `dotnet-reference.md` and `django-reference.md`.

---

## Architectural Intent (Non‑Negotiable)

The Fiber implementation of FieldMark is:
- **Server‑authoritative**
- **Domain‑centric**
- **Workflow‑driven by backend rules**
- **Symmetrical in intent (not mechanics) with the .NET and Django implementations**

Fiber is treated as an **HTTP and rendering adapter**, not the architecture itself. Without explicit constraints, Fiber can drift toward handler-centric code with business logic in request handlers; this document exists to prevent that.

---

## Conceptual Parity Model

| Concept | .NET | Django | Fiber |
|---|---|---|---|
| Domain logic | `FieldMark.Domain` project | Model methods / domain modules | `internal/domain` |
| Persistence | EF Core | Django ORM | Explicit SQL via stores in `internal/data/postgres` |
| Composition | `Program.cs` | Django settings + views | `cmd/web/main.go` |
| Server-rendered UI | Razor Pages | Django templates | `html/template` via Fiber |
| Auth | `dotnet_auth` (deferred) | `django_auth` (built-in) | `fiber_auth` (deferred) |

Parity is **conceptual**, not idiomatic.

---

## Project Layout

The Fiber implementation is a peer to the other stacks under the monorepo root:

```
fieldmark/
├── FieldMark/        (.NET)
├── fieldmark_py/     (Django)
├── fieldmark-go/     (Fiber)
├── docker/
└── docker-compose.yml
```

Inside `fieldmark-go/`:

```
fieldmark-go/
├── cmd/
│   └── web/
│       └── main.go
├── internal/
│   ├── domain/
│   │   ├── entities/
│   │   ├── valueobjects/
│   │   ├── enums/
│   │   └── errors/
│   ├── app/
│   │   ├── services/
│   │   ├── ports/
│   │   └── dto/
│   ├── data/
│   │   └── postgres/
│   │       ├── db.go
│   │       ├── stores/
│   │       └── models/
│   └── web/
│       ├── handlers/
│       ├── middleware/
│       ├── templates/
│       │   ├── layouts/
│       │   ├── pages/
│       │   ├── partials/
│       │   └── fragments/
│       └── static/
│           ├── css/
│           ├── js/
│           └── vendor/
├── go.mod
└── go.sum
```

This structure is intentionally layered by responsibility, not by technical fashion. It is not flattened, inverted, or rearranged into "clean architecture" pyramids.

---

## Layer Responsibilities

### `internal/domain`

Role:
- Business concepts and behavior

Contains:
- Entities and aggregate roots
- Value objects
- Domain enums
- Domain invariants
- State transition methods (matching the canonical `snake_case` method names in `domain-model.md` §9)
- Domain-specific errors

Must NOT contain:
- Fiber imports
- SQL
- Template rendering
- Request/response structs
- Any `internal/web`, `internal/data`, or `internal/app` imports

Dependency rule:
- **No outbound package references** beyond the Go standard library and small, neutral helpers

### `internal/app`

Role:
- Use-case orchestration and the application boundary

Contains:
- Service-level workflows that compose domain calls
- Ports (interfaces) for persistence
- DTOs that cross the app/web boundary

Must NOT contain:
- HTML rendering
- `fiber.Ctx` references
- SQL queries

### `internal/data/postgres`

Role:
- Persistence adapter mapping to the infrastructure-owned `domain.*` schema and to the framework-local `fiber_auth.*` schema

Contains:
- Database connection setup (`db.go`)
- Store implementations (`ProjectStore`, `InspectionStore`, `ViolationStore`, etc.) that satisfy ports defined in `internal/app`
- Hand-written SQL or query-builder calls against `domain.*`
- Transaction coordination

Must NOT contain:
- UI logic
- Domain workflow rules
- Any DDL against `domain.*` (ADR-014: the Fiber stack owns nothing in `domain`)

### `internal/web`

Role:
- Composition root and server-rendered UI

Contains:
- Fiber route definitions, handlers, and middleware
- Template rendering (`html/template` via Fiber's HTML engine)
- Static asset serving
- HTMX fragment endpoints

Must NOT contain:
- Business rules
- Query logic beyond orchestration
- Direct database access (always goes through `internal/app` ports)

---

## Dependency Direction (Hard Rule)

Permitted:

```
web  → app
web  → domain
app  → domain
data → app   (implements ports)
data → domain
```

Forbidden:

```
domain → web
domain → data
domain → app
app    → web
app    → data   (concrete implementations; ports/interfaces only)
```

If a dependency violates this direction, it is architecturally invalid.

---

## Database & Persistence Policy

- PostgreSQL is the sole persistence engine, shared across all three stacks.
- The `domain` schema is **infrastructure-owned** (ADR-014). It is created by Postgres init scripts in `docker/postgres/init/` and evolved through hand-authored infrastructure SQL. Fiber does **not** create, alter, or drop tables in `domain`.
- Fiber data access uses **explicit SQL** against `domain.*`. There is no Go ORM in this project; small, narrow store interfaces (e.g. `ProjectStore`, `InspectionStore`, `ViolationStore`) are the only persistence abstraction.
- Stores must:
  - Use the database's `snake_case` table and column names directly (no ORM-driven naming).
  - Treat enum-like columns as strings storing `SCREAMING_SNAKE_CASE` values (see `domain-model.md` §9).
  - Run audit-entry inserts inside the same transaction as the triggering write.
- The `fiber_auth` schema is reserved for Fiber's auth tables when authentication is implemented; it is created by the same init scripts as the other schemas.

### Persistence — Acceptable

- Domain-specific store interfaces, kept narrow: `ProjectStore`, `InspectionStore`, `ViolationStore`, `CorrectiveActionStore`, `AuditStore`.
- A small `db.go` that owns the `*sql.DB` (or `*pgxpool.Pool`) lifecycle.

### Persistence — Not Acceptable

- Generic repositories
- Layered repository abstractions for their own sake
- "Clean architecture" ports everywhere with no clear need
- Auto-migration tooling against `domain.*`

---

## Authentication & Authorization Policy

Authentication is **framework-local** (ADR-012) and **deferred by design** for the Fiber stack.

Rules:
- The `fiber_auth` schema exists in the database (created by the Postgres init scripts) but Fiber auth is not yet implemented. This is an intentional standup-time deferral.
- When Fiber adopts authentication:
  - All auth tables map to `fiber_auth.*`.
  - Middleware enforces role checks aligned with the shared conceptual roles (Administrator, Compliance Officer, Inspector, Site Supervisor, Executive Viewer).
  - Domain tables in `domain.*` must not foreign-key any `fiber_auth.*` table; user references are stored as opaque UUIDs.
- No attempt is made to share identity with the .NET or Django stacks.

---

## HTMX & Templating Strategy

- HTMX is treated as a transport enhancer for HTML fragments, not as a client-side application framework.
- Templates use Go's standard `html/template` via Fiber's HTML template renderer.
- Allowed response types:
  - Full HTML page
  - Partial HTML fragment (HTMX swap target)
  - JSON only for JS islands such as AG Grid
- Template organization:
  - `templates/layouts/` — base layouts only
  - `templates/pages/` — full route surfaces (dashboard, project list, project detail, etc.)
  - `templates/fragments/` — HTMX swap targets (`compliance_tile.html`, `violation_row.html`, `audit_log.html`)
  - `templates/partials/` — shared markup (nav, header, footer)
- HTMX target IDs (`#project-detail`, `#compliance-tile`, `#violation-detail`, `#audit-log`) must match the IDs used in the .NET and Django implementations exactly.

Disallowed:
- Client-side state stores
- Client-owned workflow logic
- REST-first UI orchestration
- DSLs or component languages that obscure the markup

---

## Static Asset Strategy

- No public CDN dependency.
- HTMX and AG Grid are vendored locally under `internal/web/static/vendor/`.
- The shared Tailwind build output (`fieldmark_style/dist/fieldmark.css`) is symlinked into `internal/web/static/vendor/fieldmark.css`, matching the other two stacks.

---

## Explicitly Rejected Patterns

Unless reversed by an explicit ADR, the following are prohibited:

- Business logic in handlers, middleware, or templates
- `fiber.Ctx` escaping the web layer
- Generic repository abstractions
- Service layers that duplicate domain behavior
- Client-side workflow orchestration
- Framework-driven "magic" replacing explicit rules
- ORM-style migrations or auto-generated DDL against `domain.*`

---

## Agent Guardrail Rules (Derivable)

An implementing agent must:
- Reject business rules in `internal/web` handlers or middleware
- Reject `fiber.Ctx` references outside `internal/web`
- Reject persistence details leaking into `internal/domain`
- Reject any SQL or migration tooling that issues DDL against `domain.*` (ADR-014)
- Reject FK relationships from `domain.*` rows to `fiber_auth.*` rows (ADR-012)
- Reject generic repository abstractions or speculative ports
- Prefer HTML responses over JSON unless a JS island requires JSON
- Keep template files close to literal HTML
- Add structure before adding abstractions

If a pattern requires explanation, it is likely too complex for this project.

---

## Unit testing

Use the Go standard library **`testing`** package only—no third-party test framework is required. Unit tests validate **domain and application behavior**, not Fiber routing, middleware chains, or HTML rendering. Shared workflows and browser-visible parity are validated with **Playwright E2E tests**, not duplicated here.

**Layout:** place **`*_test.go`** next to the code under test, for example:

```
internal/domain/violation_test.go          # next to domain types
internal/app/compliance_service_test.go    # next to services
internal/data/postgres/stores/project_store_test.go   # when exercising SQL-backed stores
```

**Do test:** domain structs and methods; application services and rule enforcement; store behavior when persistence is in scope.

**Do not unit-test as primary subjects:** Fiber route registration; middleware; template rendering.

**Persistence in tests:** the Fiber stack does **not** mock the database—tests that hit Postgres use a **real PostgreSQL** instance (see root `CLAUDE.md`). Pure domain logic tests may run without a database.

**Commands:** `go test ./...` from `fieldmark-go/` (see stack `Makefile` / `CLAUDE.md` for QA runners combined with fmt/vet/staticcheck).

---

## Core Principle

**Fiber projects the domain; it must never invent it.**

---

## Status

This document defines the **locked architectural guardrails** for the Go (Fiber) implementation of FieldMark and is intended as a stable priming artifact for agentic design enforcement.
