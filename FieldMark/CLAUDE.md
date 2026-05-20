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

ASP.NET Core Identity is wired against the `dotnet_auth` schema via `AuthDbContext`. Login and logout pages are not yet added — that work lands in Story 1.11.

### AuthDbContext and FieldMarkDbContext are independent

`FieldMark.Data/Context/AuthDbContext.cs` owns the seven Identity tables under `dotnet_auth`:
- `dotnet_auth.users`, `dotnet_auth.roles`, `dotnet_auth.user_roles`
- `dotnet_auth.role_claims`, `dotnet_auth.user_claims`, `dotnet_auth.user_logins`, `dotnet_auth.user_tokens`

`FieldMarkDbContext` owns `domain.*` mappings (added in later stories). **Never merge these two contexts** — Identity types extend `IdentityDbContext<...>` and the migration ownership is cleanly bifurcated.

### Migration folder convention

Auth migrations live in `FieldMark.Data/Migrations/Auth/`. The `domain.*` schema is infrastructure-owned and managed by SQL init scripts; it is **not** touched by any EF Core migration (ADR-014).

To add an auth migration, always specify the context and output directory:

```bash
dotnet ef migrations add <Name> --context AuthDbContext --project FieldMark.Data --startup-project FieldMark.Web --output-dir Migrations/Auth
dotnet ef database update --context AuthDbContext --project FieldMark.Data --startup-project FieldMark.Web
```

### Role seeding

Five conceptual roles are seeded idempotently at startup via `FieldMark.Web/SeedData/RoleSeeder.cs`:
- `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`

The seeder is called from `Program.cs` after the `--dump-routes` early-return so route-dump invocations do not touch the database. A `Role` value object in `FieldMark.Domain` and the `authz.Can` primitive are deferred to Story 1.12.

### Dev User Seeding

Six dev users are seeded idempotently via `FieldMark.Web/SeedData/DevUsersSeeder.cs` immediately after `RoleSeeder`:
- `marisol` (COMPLIANCE_OFFICER), `pat` (SITE_SUPERVISOR), `aisha` (ADMIN), `ravi` (INSPECTOR), `kenji` (EXECUTIVE), `testuser` (no role)

The seeder runs on every web-app startup (inside the existing `IServiceScope` block in `Program.cs`) and is also invocable standalone via `dotnet run --project FieldMark.Web -- --seed-dev-users`. Identity hashing uses `IPasswordHasher<IdentityUser<Guid>>` (PBKDF2, framework-native) — the seeder must never call the hasher directly. The manifest plaintext password (`FieldMark!2026`) is for dev-environment first-login only.

Manifest source: `docker/postgres/init/seed-uuids/dev-users.json` — resolved relative to `env.ContentRootPath` at runtime.

### What is still deferred

- `app.UseAuthentication()` / `app.UseAuthorization()` pipeline wiring — Story 1.11
- Login and logout Razor Pages — Story 1.11
- `authz.Can` primitive and `ActionButton` trichotomy — Story 1.12

## Coding Standards

- Async by default for I/O; pass `CancellationToken` through handlers.
- Nullable reference types enabled; treat warnings as errors.
- EF Core entities use private setters and parameter-validating constructors.
- Domain entities receive infrastructure dependencies (e.g., `IClock`) via method parameters, not constructor DI.

## Hard Rules (.NET-specific)

Root `CLAUDE.md` covers the cross-stack rules (no CQRS/MediatR, no repositories, no AutoMapper, no client-side state, real PostgreSQL in tests). The .NET-specific rules are:

- `FieldMark.Domain` has zero project or package references. Adding any is architecturally invalid.
- No MVC controllers — Razor Pages only. No Blazor.
- No Clean / Onion / Hexagonal layering on top of the existing 4-project layout.
- No Mapster either — manual projection in LINQ. (AutoMapper is forbidden by the root rules; Mapster was evaluated and rejected — see architecture.md → §Core Architectural Decisions → NET-MAPSTER.)

## Agent Behaviour Rules

- **Prefer modifying structure over adding abstractions.** If a solution requires a new pattern or layer, question whether the problem is being misstated.
- **If a solution requires explaining why it is structured a certain way, it is likely invalid.** Complexity must be earned and self-evident.
- Reject any design that adds EF Core to Domain, points Domain at Data or Web, or introduces repositories or mediators.

## Reference

- `_bmad-output/planning-artifacts/architecture.md` — architectural source of truth (canonical request flow with .NET code stub, decisions, patterns)
- `_bmad-output/planning-artifacts/prd/` — capability source of truth
- Root `CLAUDE.md` — cross-stack rules and canonical inventories (audit actions, HTMX target IDs, method names)
