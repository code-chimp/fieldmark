# FieldMark

**Construction Compliance & Inspection Management System**

FieldMark is a reference implementation of an enterprise-grade Construction Compliance & Inspection Management System (CCIMS), built to demonstrate that server-driven web architecture can deliver SPA-equivalent interactivity without the cognitive and architectural overhead of a single-page application.

The system is implemented across three stacks — .NET (Razor Pages + HTMX), Django (Django Templates + HTMX), and Go (Fiber + HTMX) — against a shared PostgreSQL database, with strict architectural symmetry enforced at every story boundary. AG Grid is the planned JavaScript island for data-dense views, but no client-side state stores, no duplicated business rules, and no frontend routing exist in any stack.

FieldMark is a teaching artifact for an upcoming talk on HTMX. It is not a product seeking market fit.

## The Thesis

Modern enterprise web applications routinely default to SPA architectures regardless of whether the application's interaction patterns actually require client-owned state. For applications whose core interactions are fundamentally request-response — submit an inspection, resolve a violation, view a dashboard — the SPA default trades architectural complexity for marginal interactivity gains. FieldMark makes that trade-off legible by implementing a non-trivial domain in a server-authoritative style across three different backend ecosystems and inviting direct comparison.

## Domain

FieldMark models the lifecycle of construction compliance: project managers oversee a portfolio of construction engagements, compliance officers schedule and perform inspections, inspections produce findings that spawn violations, site supervisors submit corrective actions, and a server-evaluated rules engine scores compliance and gates workflow transitions. The domain is rich enough to require state machines, role-based access, audit trails, and configurable compliance rules — without being so broad that the architectural message gets lost.

### Key workflows

- **Project lifecycle** — create, place on hold, resume, close (gated by compliance rules)
- **Inspection workflow** — schedule, start, complete with findings, cancel
- **Violation management** — open from findings, assign to supervisors, track corrective actions through submission, review, approval/rejection
- **Compliance scoring** — server-computed 0–100 score per project based on open and overdue violations, recalculated on every relevant state transition
- **Audit trail** — immutable, append-only log of every domain mutation, written in the same transaction as the change it records

## Repository Layout

```
fieldmark/
├── Makefile                Source of truth for local orchestration
├── FieldMark/              .NET solution (Razor Pages + HTMX)
│   ├── FieldMark.Domain/     Domain entities and behavior
│   ├── FieldMark.Data/       EF Core contexts and dotnet_auth migrations
│   ├── FieldMark.Web/        Razor Pages app, auth, seeders, route dump tooling
│   ├── FieldMark.Tests.Domain/       xUnit domain tests
│   ├── FieldMark.Tests.Integration/  xUnit PostgreSQL integration tests
│   └── FieldMark.Tests.Web/          xUnit web/rendering tests
├── fieldmark_py/           Django project (Templates + HTMX)
│   ├── fieldmark/            Project settings, auth/authz helpers, URLs, tests
│   ├── templates/            Shared layouts, fragments, and page templates
│   ├── static/               Symlinks to shared CSS/vendor assets
│   ├── projects/             Project app and dev-user seeding
│   ├── tools/                Stack-local management tooling
│   └── audit/, compliance/, grid/, inspections/, reference/, violations/
├── fieldmark-go/           Go project (Fiber + HTMX)
│   ├── cmd/web/              Web entry point
│   ├── cmd/migrate-fiber-auth/  fiber_auth DDL migrator
│   ├── cmd/seed/             Go auth/dev-user seeder
│   ├── internal/             App, data, domain, web handlers/templates/auth
│   └── tools/                Go tool pinning
├── fieldmark_shared/       Shared Tailwind/Basecoat CSS, examples, and vendor JS
│   ├── src/                  CSS source and partials
│   ├── dist/                 Committed compiled CSS
│   ├── components/           Canonical HTML examples used by stack tests
│   ├── scripts/              CSS build optimization
│   └── vendor/               Vendored HTMX and theme-toggle assets
├── fieldmark-landing/      Static project landing page
│   ├── index.html            Self-contained landing page
│   └── static/               Landing page fonts, images, and icons
├── docs/                   Canonical project docs (overview, architecture, rules)
├── e2e/                    Playwright browser tests and fixtures
├── tools/                  Domain verification, parity scripts, git-hook samples
│   └── parity/               Cross-stack route/index dump and diff scripts
├── docker/
│   └── postgres/
│       └── init/             Schema init SQL and shared seed UUID manifest
├── docker-compose.yml      PostgreSQL 17 for local development
└── README.md               This file
```

Each stack and shared package has its own README with focused setup instructions:

- [**.NET README**](FieldMark/README.md)
- [**Django README**](fieldmark_py/README.md)
- [**Go README**](fieldmark-go/README.md)
- [**Shared assets README**](fieldmark_shared/README.md)
- [**Landing page README**](fieldmark-landing/README.md)
- [**Documentation index**](docs/README.md)

## Tech Stack

| Component | .NET | Django | Go |
|---|---|---|---|
| Runtime | .NET 10 | Python 3.14+ | Go 1.26+ |
| Web framework | ASP.NET Core Razor Pages | Django 6.x | Fiber v3 |
| ORM / data access | EF Core 10 | Django ORM | Explicit SQL via stores |
| Database | PostgreSQL 17 | PostgreSQL 17 | PostgreSQL 17 |
| Interactivity | HTMX 4.0.0-beta2 | HTMX 4.0.0-beta2 | HTMX 4.0.0-beta2 |
| Shared CSS | Tailwind CSS 4.2.4 + Basecoat 0.3.11 | Tailwind CSS 4.2.4 + Basecoat 0.3.11 | Tailwind CSS 4.2.4 + Basecoat 0.3.11 |
| Data grids | AG Grid Community 35.x contract | AG Grid Community 35.x contract | AG Grid Community 35.x contract |

Shared browser assets are owned by `fieldmark_shared/` and consumed by the stacks through symlinks. HTMX is currently vendored there; AG Grid server-side-row-model contracts are part of the architecture and must stay identical when grid endpoints are implemented.

## Database Architecture

All three stacks share a single PostgreSQL database with schema-level isolation:

| Schema | Owner |
|---|---|
| `domain` | Infrastructure SQL init scripts — authoritative for all business data |
| `django_auth` | Django stack |
| `dotnet_auth` | .NET stack |
| `fiber_auth` | Go stack |

The `domain` schema is created by init scripts in `docker/postgres/init/` and is not owned or migrated by any single framework. Frameworks map to `domain.*` tables; they do not create or alter them. Framework-specific auth schemas are owned by their respective stacks.

## Architectural Rules

These are non-negotiable across all three stacks:

1. **Backend authority.** Domain rules, workflow transitions, validation, and authorization are server-side only.
2. **Rich domain model.** Behavior lives on entities. No CQRS, no generic repositories, no mediator patterns, no Clean/Onion layering.
3. **Stack symmetry.** All implementations are structurally equivalent — same routes, same HTMX target IDs, same audit action strings, same domain method names (modulo language casing conventions).
4. **HTML over the wire.** HTMX drives interactivity. JavaScript is restricted to AG Grid wiring and minimal UX glue.
5. **No client state stores.** No Redux, NgRx, Pinia, Zustand, or equivalents. Ever.
6. **Earned complexity.** No abstraction is added speculatively. If a pattern requires explanation during the demo, it likely should not exist.
7. **Infrastructure-owned domain schema.** The shared `domain` schema is created by SQL init scripts, not by any framework's migration tooling.

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) or [Docker Engine](https://docs.docker.com/engine/install/) — for PostgreSQL
- [.NET 10 SDK](https://dotnet.microsoft.com/download) — for the .NET stack
- [Python 3.14+](https://www.python.org/) with [uv](https://docs.astral.sh/uv/) — for the Django stack
- [Go 1.26+](https://go.dev/dl/) — for the Go stack
- [Node.js 20+](https://nodejs.org/) with [pnpm](https://pnpm.io/installation) — for CSS builds (`fieldmark_shared`) and e2e tests (later stories); Tailwind's Oxide engine requires Node ≥ 20
- `psql` — PostgreSQL client for the verification script. On macOS: `brew install libpq && brew link --force libpq`

## Getting Started

Start with the Makefile; it is the executable source of truth for local development:

```bash
make help
```

**1. Start PostgreSQL:**

```bash
make up
```

This starts PostgreSQL 17 on `localhost:5432` and runs the schema init scripts on first volume creation. Credentials: `fieldmark / fieldmark / fieldmark`.

**2. Apply auth migrations and seed dev users:**

Each stack manages its own auth schema. After starting Postgres, apply migrations then seed:

```bash
# .NET — applies dotnet_auth migrations and seeds roles + users
cd FieldMark && dotnet ef database update --context AuthDbContext --project FieldMark.Data --startup-project FieldMark.Web

# Django — applies django_auth migrations
cd fieldmark_py && uv run python manage.py migrate
uv run python manage.py seed_groups

# Go — applies fiber_auth DDL
cd fieldmark-go && go run ./cmd/migrate-fiber-auth

# Seed all three stacks at once (requires auth schemas to exist first)
make seed
```

`make seed` reads `docker/postgres/init/seed-uuids/dev-users.json` and writes the six dev users into all three stacks' auth schemas with identical UUIDs. It is idempotent — re-running is safe.

**3. Run the stacks** (each in its own terminal):

```bash
make run-net       # .NET Razor Pages  →  http://localhost:5000
make run-django    # Django Templates  →  http://localhost:8000
make run-go        # Go Fiber          →  http://localhost:3000
```

All three stacks connect to the same Postgres instance and can run simultaneously.

**4. Reset the database** (destroy volume and re-run init scripts):

```bash
make reset
```

Run `make help` for the full list of available targets.

### Verifying the database

After `make up` or `make reset`, verify the canonical schema:

```bash
./tools/verify-domain-schema.sh
```

Expected output:

```
OK domain schema verified (5 schemas, 12 tables, N trade types, N violation categories, 4 compliance rules)
```

Non-zero exit = schema drift. Investigate before running any stack.

### Per-stack setup

Each stack has its own README with additional dev instructions:

- [FieldMark/README.md](FieldMark/README.md) — .NET (Razor Pages + EF Core)
- [fieldmark_py/README.md](fieldmark_py/README.md) — Django (Templates + psycopg)
- [fieldmark-go/README.md](fieldmark-go/README.md) — Go (Fiber + pgx)

### Architecture rules

See [CLAUDE.md](CLAUDE.md) for the cross-stack architectural rules enforced at every story boundary, and each stack's own `CLAUDE.md` for stack-specific constraints.

## Status

Core scaffolding, shared PostgreSQL schema, design-system foundation, cross-stack base layout, theme toggle, auth schemas, login/logout, authorization primitives, dev-user seeding, and parity tooling are in place. Feature implementation is continuing across all three stacks.

## License

MIT License — Copyright (c) 2026 Tim Goshinski. See [LICENSE](LICENSE) for details.
