# fieldmark-go

**Go (Fiber) implementation of FieldMark** — one of three server-authoritative backend stacks demonstrating that HTML-over-the-wire can deliver enterprise-grade interactivity without SPA architecture.

The other stacks are [`FieldMark/`](../FieldMark/README.md) (.NET Razor Pages) and [`fieldmark_py/`](../fieldmark_py/README.md) (Django Templates). All three implement the same domain, expose the same routes, and produce the same audit trail — the point is that the architecture is portable, not that any one framework is superior.

---

## What This Stack Demonstrates

Fiber is used here to show that backend authority is an **architectural posture**, not a feature of a specific framework. Where .NET and Django provide opinionated, batteries-included defaults, Go with Fiber requires explicit construction of the same structure — making the decisions visible rather than implicit.

> Fiber is an HTTP and rendering adapter. It is not the architecture itself.

---

## Architecture

This implementation follows the same non-negotiable principles as the other stacks:

1. **The backend owns truth.** Domain logic, workflow transitions, validation, and authorization are server-side only.
2. **Domain is centralized and explicit.** Business rules live on domain entities, not in handlers.
3. **Persistence is an adapter.** Data access does not drive design decisions.
4. **UI is a projection.** Templates render authoritative server state; they do not own it.
5. **Complexity must be earned.** No abstraction is added speculatively.

### Dependency direction (hard rule)

```
web → app → domain
data/postgres → app, domain
```

The `domain` package has no outbound dependencies. The `web` package never contains business rules. The `app` package coordinates — it does not own domain logic or render HTML.

### Request flow

```
Browser
  ↓  HTTP / HTMX request
web/handler
  ↓  invoke use case
app/service
  ↓  execute domain rules / transitions
domain/entity
  ↓  load or persist
data/postgres/store
  ↓  return DTO
web/handler
  ↓  render template
Browser ← HTML fragment or full page
```

---

## Project Structure

```
fieldmark-go/
├── cmd/
│   └── web/
│       └── main.go          # composition root; wires dependencies and starts Fiber
├── internal/
│   ├── domain/
│   │   ├── entities/        # Project, Inspection, Violation, CorrectiveAction — with state-transition methods
│   │   ├── valueobjects/    # typed wrappers (ProjectCode, Severity, etc.)
│   │   ├── enums/           # domain enum constants
│   │   └── errors/          # typed domain errors (InvalidStateTransition, etc.)
│   ├── app/
│   │   ├── services/        # use-case orchestration (thin; delegates to domain)
│   │   ├── ports/           # persistence interfaces the app layer depends on
│   │   └── dto/             # data transfer objects crossing app ↔ web boundary
│   ├── data/
│   │   └── postgres/
│   │       ├── db.go        # connection setup
│   │       ├── stores/      # ProjectStore, InspectionStore, ViolationStore, etc.
│   │       └── models/      # scan targets for SQL results
│   └── web/
│       ├── handlers/        # one file per route group; thin; no business logic
│       ├── middleware/       # auth, request-id, logging
│       ├── templates/
│       │   ├── layouts/     # base.html
│       │   ├── pages/       # full route surfaces (dashboard, project_detail, etc.)
│       │   ├── partials/    # shared markup (nav, header, footer)
│       │   └── fragments/   # HTMX swap targets (compliance_tile, violation_row, audit_log)
│       └── static/
│           ├── css/
│           ├── js/
│           └── vendor/      # vendored HTMX and AG Grid (no CDN dependency)
├── go.mod
└── go.sum
```

---

## Database

fieldmark-go connects to the shared PostgreSQL instance used by all three stacks.

### Schema ownership

The FieldMark domain schema is **infrastructure-owned**, not owned by this application:

| Schema | Owner |
|---|---|
| `domain` | Infrastructure SQL init scripts — not any framework |
| `fiber_auth` | This stack (when auth is implemented) |
| `django_auth` | Django stack |
| `dotnet_auth` | .NET stack |

Domain tables (`domain.projects`, `domain.inspections`, `domain.violations`, etc.) are created by SQL init scripts that run when the PostgreSQL container first starts — see [`docker/postgres/init/`](../docker/postgres/init/). This stack maps to those tables; it does not create or migrate them.

This means:
- **Do not use any ORM auto-migration tooling against the `domain` schema.** Treat it as a shared contract.
- Stores in `internal/data/postgres/stores/` query the `domain.*` tables via explicit SQL.
- Auth tables in `fiber_auth` are managed by this stack when authentication is added.

### Domain user references

Domain tables store user identifiers as opaque values (`created_by_user_id`, `assigned_to_user_id`, etc.) — they do not foreign-key any framework's auth tables. This keeps the domain portable across all three stacks.

### Local connection

```
Host:     localhost
Port:     5432
Database: fieldmark
User:     fieldmark
Password: fieldmark
```

---

## Prerequisites

- [Go 1.26+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) — for the shared PostgreSQL instance

---

## Getting Started

**1. Start the database (from the monorepo root):**

```bash
docker compose up -d
```

This starts PostgreSQL 17, runs the schema init scripts in `docker/postgres/init/`, and makes the database available on `:5432`.

**2. Install dependencies:**

```bash
go mod download
```

**3. Apply framework-local auth schema:**

```bash
go run ./cmd/migrate-fiber-auth
```

Creates the `fiber_auth.users` and `fiber_auth.user_roles` tables (idempotent — safe to re-run). Required once after `make reset`.

**4. Run the application:**

```bash
go run cmd/web/main.go
```

The server starts on [http://localhost:3000](http://localhost:3000).

---

## Current State

Standup is complete. The Fiber server starts, validates the Postgres connection, serves static assets, renders a full-page dashboard route and one HTMX fragment route (`/fragments/compliance-tile`), and hydrates a stub authentication actor onto every request from the `X-FieldMark-Actor` cookie/header or the `FIELDMARK_STUB_ACTOR` env var. Folder layout matches the structure above. Domain implementation begins with the first feature story.

---

## Guardrails

### Allowed
- Thin Fiber handlers that delegate immediately to `app/services`
- Explicit, domain-specific store interfaces (`ProjectStore`, `ViolationStore`)
- `html/template` for all server-rendered output
- HTMX fragments as the primary interactivity mechanism
- JSON responses only for AG Grid data endpoints

### Prohibited
- Business rules in handlers or middleware
- `fiber.Ctx` escaping the `web` layer
- Persistence details leaking into `domain` or `app`
- Generic repository abstractions
- Client-side state stores or workflow orchestration
- Auto-migration of the shared `domain` schema

If a pattern requires a paragraph of explanation, it probably should not exist in this codebase.

---

## Key Reference Documents

All in [`_bmad-output/planning-artifacts/research/`](../_bmad-output/planning-artifacts/research/):

| Document | Purpose |
|---|---|
| `domain-model.md` | Entity catalog, state machines, invariants, PostgreSQL schema |
| `architecture-decisions.md` | ADRs 011–014: ORM-first model, auth isolation, schema init, domain ownership |
| `dotnet-reference.md` | .NET stack guardrails (useful for cross-stack parity reference) |
| `django-reference.md` | Django stack guardrails (useful for cross-stack parity reference) |

Stack-specific architecture reference: [`_bmad-output/planning-artifacts/research/fiber-reference.md`](../_bmad-output/planning-artifacts/research/fiber-reference.md)
