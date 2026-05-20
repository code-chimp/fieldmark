# FieldMark — .NET Solution

The .NET implementation of FieldMark, built with ASP.NET Core Razor Pages, EF Core, and HTMX.

## Architecture

This solution follows a server-authoritative, domain-centric architecture. The domain owns all business rules, state transitions, and invariants. Persistence and UI are adapters. Detailed rationale is in `_bmad-output/planning-artifacts/research/dotnet-reference.md` and `_bmad-output/planning-artifacts/research/architecture-decisions.md`.

### Solution Structure

```
FieldMark.sln
├── FieldMark.Domain/       Domain entities, value objects, enums, exceptions
│                           No outbound project references. No EF Core. No NuGet packages.
│
├── FieldMark.Data/         EF Core persistence adapter
│   ├── Context/              FieldMarkDbContext
│   └── Configuration/        IEntityTypeConfiguration<T> per entity
│                           References: FieldMark.Domain
│
├── FieldMark.Web/          ASP.NET Core Razor Pages — composition root
│   ├── Pages/                Razor Pages grouped by domain area
│   │   └── Shared/           Layouts, partials
│   ├── wwwroot/
│   │   ├── css/
│   │   ├── js/               AG Grid wiring, UX helpers
│   │   └── lib/              HTMX, AG Grid (vendored)
│   └── Program.cs            DI registration, middleware, startup
│                           References: FieldMark.Domain, FieldMark.Data
│
├── FieldMark.Tests.Domain/       xUnit domain unit tests (no I/O)
│                               References: FieldMark.Domain
│
└── FieldMark.Tests.Integration/  xUnit integration tests
                                References: FieldMark.Domain, FieldMark.Data
                                Uses Testcontainers.PostgreSql for real DB
```

### Dependency Direction (Hard Rule)

```
Web  → Domain    ✓
Web  → Data      ✓
Data → Domain    ✓

Domain → Data    ✗  (never)
Domain → Web     ✗  (never)
```

If a dependency violates this direction, it is architecturally invalid.

## Tech Stack

| Layer | Choice | Version |
|---|---|---|
| Runtime | .NET | 10 |
| Web framework | ASP.NET Core Razor Pages | 10 |
| ORM | EF Core | 10.0.7 |
| Database driver | Npgsql | 10.0.1 |
| Database | PostgreSQL | 17 |
| Interactivity | HTMX | 4.x |
| Data grids | AG Grid Community | 35.x |

## Prerequisites

- [.NET 10 SDK](https://dotnet.microsoft.com/download)
- PostgreSQL 17 running locally (see root `docker-compose.yml`)

## Getting Started

**1. Start PostgreSQL** (from the repo root):

```bash
docker compose up -d
```

**2. Run auth migrations:**

```bash
dotnet ef database update --context AuthDbContext --project FieldMark.Data --startup-project FieldMark.Web
```

**3. Seed roles and dev users:**

```bash
dotnet run --project FieldMark.Web -- --seed-dev-users
```

This seeds the five conceptual roles and the six dev users (from the shared `docker/postgres/init/seed-uuids/dev-users.json` manifest) into `dotnet_auth`. The command is idempotent. See `make seed` in the root `Makefile` for the canonical cross-stack convenience that seeds all three stacks at once.

**4. Run the application:**

```bash
dotnet run --project FieldMark.Web
```

The app will be available at `http://localhost:5000`.

### Creating Migrations

When domain entities change:

```bash
dotnet ef migrations add <MigrationName> -p FieldMark.Data -s FieldMark.Web
```

### Schema Ownership

The `domain` schema is created and migrated by `docker/postgres/init/`. EF Core migrations in this stack apply only to the `dotnet_auth` schema. Running `dotnet ef migrations add` against a `DbContext` whose default schema is `"domain"` is a defect — it would violate ADR-014 (infrastructure-owned domain schema). Domain entities are mapped to existing tables via `ToTable` in `IEntityTypeConfiguration<T>`; they are not created or altered by EF Core.

## Architectural Constraints

The following patterns are **explicitly rejected** and will not be introduced without an ADR amendment:

- CQRS or MediatR
- Repository pattern or Unit-of-Work abstractions
- Clean / Onion / Hexagonal layering
- Client-side state management
- API-first SPA backend
- AutoMapper (project to view models manually)
- SQLite for tests (use real PostgreSQL via Testcontainers)

### What belongs where

**FieldMark.Domain** — entity classes with behavior. State transition methods (`place_on_hold`, `close`, `approve_resolution`, etc.), domain invariants, `can_*` predicates, and typed exceptions. No EF Core references, no data annotations, no serialization attributes.

**FieldMark.Data** — `FieldMarkDbContext`, `DbSet` declarations, `IEntityTypeConfiguration<T>` mappings. No business rules, no workflow logic.

**FieldMark.Web** — Razor Page handlers that are thin orchestrators: authorize, begin transaction, load aggregate, invoke domain method, append audit entry, recompute compliance score, commit, render template. If a handler is doing business logic, it belongs on the entity.

### Coding Standards

- Async by default for I/O; pass `CancellationToken` through.
- Nullable reference types enabled; treat warnings as errors.
- EF Core entities use private setters and parameter-validating constructors.
- Domain entities receive infrastructure dependencies (e.g., `IClock`) via method parameters, not via DI on the entity.
- `Program.cs` is the sole composition root.

## Request Flow

```
Browser
  │  HTMX request (hx-get / hx-post)
  ▼
Razor Page handler
  │  authorize (RBAC)
  │  load aggregate via DbContext
  │  invoke domain method on entity
  │  persist (single transaction, includes audit write + recomputation)
  │  render partial or full page
  ▼
Server-rendered HTML → HTMX swaps into DOM
```

For AG Grid views, minimal API endpoints return paginated JSON using the server-side row model. Row selection triggers an HTMX request to load a detail panel.

## Parity

This implementation must remain structurally equivalent to the Django and Go stacks at all times. Routes, HTMX target IDs, AG Grid endpoint contracts, audit entry shapes, and domain method names must match across all three stacks (modulo language casing conventions). A story is not done until all three stacks pass it.

## Related Documentation

- [Root README](../README.md) — project overview, thesis, domain summary
- [Django README](../fieldmark_py/README.md) — the parallel Python/Django implementation
- [Go README](../fieldmark-go/README.md) — the parallel Go/Fiber implementation
- [Domain Model](../_bmad-output/planning-artifacts/research/domain-model.md) — entities, state machines, schema
- [.NET Architecture Reference](../_bmad-output/planning-artifacts/research/dotnet-reference.md) — .NET-specific guardrails
- [Architecture Decisions](../_bmad-output/planning-artifacts/research/architecture-decisions.md) — ADRs and hard constraints
