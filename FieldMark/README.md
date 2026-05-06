# FieldMark — .NET Solution

The .NET implementation of FieldMark, built with ASP.NET Core Razor Pages, EF Core, and HTMX.

## Architecture

This solution follows a server-authoritative, domain-centric architecture. The domain owns all business rules, state transitions, and invariants. Persistence and UI are adapters. Detailed rationale is in `docs/FieldMark_DotNet_Architecture_Reference.md` and `docs/architecture.md` at the repo root.

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
└── FieldMark.Web/          ASP.NET Core Razor Pages — composition root
    ├── Pages/                Razor Pages grouped by domain area
    │   └── Shared/           Layouts, partials
    │   (Project/, Admin/, etc. added as domain is implemented)
    ├── wwwroot/
    │   ├── css/
    │   ├── js/               AG Grid wiring, UX helpers
    │   └── lib/              HTMX, Bootstrap, jQuery
    └── Program.cs            DI registration, middleware, startup
                            References: FieldMark.Domain, FieldMark.Data
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
| Interactivity | HTMX | 2.x |
| Data grids | AG Grid Community | 32.x |

## Prerequisites

- [.NET 10 SDK](https://dotnet.microsoft.com/download)
- PostgreSQL 17 running locally (see root `docker-compose.yml`)

## Getting Started

**1. Start PostgreSQL** (from the repo root):

```bash
docker compose up -d
```

**2. Run migrations:**

```bash
dotnet ef database update -p FieldMark.Data -s FieldMark.Web
```

**3. Run the application:**

```bash
cd FieldMark.Web
dotnet run
```

The app will be available at `https://localhost:5001` (or the port shown in console output).

### Creating Migrations

When domain entities change:

```bash
dotnet ef migrations add <MigrationName> -p FieldMark.Data -s FieldMark.Web
```

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

This implementation must remain structurally equivalent to the Django stack at all times. Routes, HTMX target IDs, AG Grid endpoint contracts, audit entry shapes, and domain method names must match (modulo language casing conventions). A story is not done until both stacks pass it.

## Related Documentation

- [Root README](../README.md) — project overview, thesis, domain summary
- [Django README](../fieldmark_py/README.md) — the parallel Python implementation
- [Architecture](../docs/architecture.md) — full architecture with data flow patterns
- [Domain Model](../docs/domain-model.md) — entities, state machines, schema
- [.NET Architecture Reference](../docs/FieldMark_DotNet_Architecture_Reference.md) — .NET-specific guardrails
