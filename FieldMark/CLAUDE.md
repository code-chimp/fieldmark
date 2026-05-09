# CLAUDE.md — .NET Stack

This file provides guidance to Claude Code (claude.ai/code) when working in the `FieldMark/` .NET solution. Read alongside the root `CLAUDE.md`.

## Commands

Run from the `FieldMark/` directory:

```bash
dotnet build
dotnet run --project FieldMark.Web
dotnet watch run --project FieldMark.Web
dotnet test
dotnet test --filter "FullyQualifiedName~<TestName>"
dotnet ef database update -p FieldMark.Data -s FieldMark.Web
dotnet ef migrations add <Name> -p FieldMark.Data -s FieldMark.Web

# Formatting (CSharpier — installed as local tool in .config/dotnet-tools.json)
dotnet tool restore                  # restore tools after a fresh clone
dotnet csharpier format .            # format all C# files in place
dotnet csharpier check .             # CI: fail if any file would be reformatted
```

## Project Structure

Five projects. Production dependency graph is strictly one-way; test projects reference only what they test:

```
FieldMark.Domain              — entities with behavior; zero outbound references
      ↑
FieldMark.Data                — EF Core persistence adapter; references Domain only
      ↑
FieldMark.Web                 — Razor Pages composition root; references Domain and Data

FieldMark.Tests.Domain        — xUnit unit tests; references Domain only; no I/O
FieldMark.Tests.Integration   — xUnit integration tests; references Domain + Data;
                                uses Testcontainers.PostgreSql for a real DB
```

`Domain → anything` is always invalid. If a dependency points outward from Domain, the design is wrong.

Test projects must never reference `FieldMark.Web`. Integration tests spin up a real PostgreSQL container via Testcontainers — SQLite is prohibited.

## What Belongs Where

**FieldMark.Domain** contains:
- Entities and aggregate roots
- Value objects, domain enums
- State-transition methods and `can_*` predicates
- Domain invariants and typed exceptions

**FieldMark.Domain** must NOT contain:
- Any EF Core reference
- Data annotations for persistence (`[Key]`, `[Required]`, etc.)
- Serialization attributes
- Validation frameworks (FluentValidation, DataAnnotations, etc.)
- Any external NuGet package

**FieldMark.Data** contains:
- `FieldMarkDbContext`, `DbSet` declarations
- `IEntityTypeConfiguration<T>` per entity
- Database-specific concerns and migrations

**FieldMark.Data** must NOT contain business rules, workflow logic, or UI concerns.

Allowed packages in Data: `Microsoft.EntityFrameworkCore`, `Microsoft.EntityFrameworkCore.Design`, `Npgsql.EntityFrameworkCore.PostgreSQL`.

**FieldMark.Web** — Razor Page handlers are thin orchestrators. Follow the eight-step canonical flow from the root `CLAUDE.md`. If a handler does business logic, it belongs on the entity.

Allowed packages in Web: `Microsoft.EntityFrameworkCore.Tools`, `Npgsql.EntityFrameworkCore.PostgreSQL`.

`Program.cs` is the sole DI registration point. `DbContext` is defined in Data and registered in Web — it must never appear in Domain.

## Migration Ownership

EF Core migrations are scoped to the `dotnet_auth` schema only — ASP.NET Core Identity tables and any .NET-specific infrastructure. The shared `domain` schema is created and evolved by SQL init scripts in `docker/postgres/init/` and is not owned by any framework. Do not generate EF Core migrations that create or alter `domain.*` tables. Dual ownership of any domain table is prohibited.

## Authentication

ASP.NET Core Identity is **deferred by design**. Do not scaffold it until:
- Domain schema is stable
- Migration ownership is explicit and settled
- Architectural rules are locked

Do not introduce Identity during any story that does not explicitly call for it.

## Coding Standards

- Async by default for I/O; pass `CancellationToken` through handlers.
- Nullable reference types enabled; treat warnings as errors.
- EF Core entities use private setters and parameter-validating constructors.
- Domain entities receive infrastructure dependencies (e.g., `IClock`) via method parameters, not constructor DI.

## Hard Rules

- `FieldMark.Domain` has zero project or package references. Adding any is architecturally invalid.
- No MVC controllers — Razor Pages only. No Blazor.
- No CQRS, MediatR, or any in-process command bus.
- No repository or Unit-of-Work abstractions. Use `FieldMarkDbContext` directly.
- No Clean / Onion / Hexagonal layering.
- No AutoMapper. Project to view models manually.
- No client-side state stores. No API-first SPA backend.
- Tests use real PostgreSQL via Testcontainers. SQLite is prohibited.

## Agent Behaviour Rules

- **Prefer modifying structure over adding abstractions.** If a solution requires a new pattern or layer, question whether the problem is being misstated.
- **If a solution requires explaining why it is structured a certain way, it is likely invalid.** Complexity must be earned and self-evident.
- Reject any design that adds EF Core to Domain, points Domain at Data or Web, or introduces repositories or mediators.

## Reference

- `_bmad-output/planning-artifacts/research/dotnet-reference.md` — full .NET guardrails (authoritative)
- `_bmad-output/planning-artifacts/research/architecture-decisions.md` — ADRs and hard constraints
