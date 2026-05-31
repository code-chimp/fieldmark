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

ASP.NET Core Identity is wired against the `dotnet_auth` schema via `AuthDbContext`. Login (`Pages/Account/Login.cshtml`) and logout (`Pages/Account/Logout.cshtml`) Razor Pages landed in Story 1.11.

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

The seeder is called from `Program.cs` after the `--dump-routes` early-return so route-dump invocations do not touch the database. `Program.cs` runs `AuthDbContext.Database.MigrateAsync()` before the seeder at every startup (including tests). A `Role` value object in `FieldMark.Domain` and the `authz.Can` primitive are deferred to Story 1.12.

### Dev User Seeding

Six dev users are seeded idempotently via `FieldMark.Web/SeedData/DevUsersSeeder.cs` immediately after `RoleSeeder`:
- `marisol` (COMPLIANCE_OFFICER), `pat` (SITE_SUPERVISOR), `aisha` (ADMIN), `ravi` (INSPECTOR), `kenji` (EXECUTIVE), `testuser` (no role)

The seeder runs on every web-app startup (inside the existing `IServiceScope` block in `Program.cs`) and is also invocable standalone via `dotnet run --project FieldMark.Web -- --seed-dev-users`. Identity hashing uses `IPasswordHasher<IdentityUser<Guid>>` (PBKDF2, framework-native) — the seeder must never call the hasher directly. The manifest plaintext password (`FieldMark!2026`) is for dev-environment first-login only.

Manifest source: `docker/postgres/init/seed-uuids/dev-users.json` — resolved relative to `env.ContentRootPath` at runtime.

### Story 1.11 shipped

- `app.UseAuthentication()` / `app.UseAuthorization()` pipeline wiring — in `Program.cs`
- `Pages/Account/Login.cshtml` / `.cshtml.cs` — `[AllowAnonymous]`, 422 on bad creds
- `Pages/Account/Logout.cshtml` / `.cshtml.cs` — `[AllowAnonymous]`, POST clears session
- Fallback `RequireAuthenticatedUser` policy applied to all non-exempt pages
- `FieldMark.Web/Authentication/ClaimsPrincipalExtensions.cs` — `GetActorId()`, `GetConceptualRoles()`
- `FieldMark.Tests.Web` project — 10 integration tests with Testcontainers + `NoOpAntiforgery`

### What is still deferred

_(Nothing in Epic 1 remains deferred after Story 1.12.)_

## Authorization

The single .NET-side authorization decision primitive is `DomainPolicies.Can` in `FieldMark.Web/Authorization/DomainPolicies.cs`. Signature:

```csharp
DomainPolicies.Can(ClaimsPrincipal user, string action, Guid? entityId = null) : bool
```

**Rules:**
- Handlers and Razor page-model code call `Can`; templates receive pre-computed `permission` booleans — templates must never call `Can` directly.
- Role names are defined in `FieldMark.Domain/ValueObjects/Role.cs` as `Role.Admin`, `Role.ComplianceOfficer`, etc. Hard-coded role-name string literals anywhere else are a defect.
- Actions are registered via `DomainPolicies.RegisterAction(action, roles...)`. Epic 2+ stories register their actions at startup (typically in `Program.cs` or a per-aggregate `<Aggregate>Policies.Register()` helper). Story 1.12 ships the map empty — Epic 1 has no live action affordances.
- Entity-scope rules (e.g., "Site Supervisor can act only on assigned Violations") are deferred to Epic 2+ and will wire into `EvaluateEntityScope` inside `DomainPolicies.cs` without changing the call-site contract.

**ActionButton partial:** `Pages/Shared/_ActionButton.cshtml` with view model `ViewModels/Components/ActionButtonVm.cs`. The caller supplies pre-computed `Permission` (from `Can`) and `StateAllows` (from the entity's `can_*` predicate). The partial renders the absent / disabled / present trichotomy; callers never implement the trichotomy themselves.

Canonical snapshot reference: `fieldmark_shared/components/action_button.example.html`.

## Coding Standards

- Async by default for I/O; pass `CancellationToken` through handlers.
- Nullable reference types enabled; treat warnings as errors.
- EF Core entities use private setters and parameter-validating constructors.
- Domain entities receive infrastructure dependencies (e.g., `IClock`) via method parameters, not constructor DI.

## Razor Component Rules

These rules apply to every Razor partial in `Pages/Shared/Components/` and were ratified after Story 2.4's five review rounds.

**Null-safety — always extract model properties through `S()` before use.** Never access `@Model.PropertyName` directly inside markup. Extract every model property to a local variable at the top of the partial:

```razor
@{
    var title = S(Model.Title);
    var message = S(Model.Message);
    var tileId = S(Model.TileId);
}
```

`S()` (the null-safety string helper in each partial) returns `""` for null or whitespace inputs, preventing `RuntimeBinderException` when a dynamic model omits a key. This applies to ALL model reads — including `Id`, label, and timestamp properties, not just user-visible strings.

**No `Html.Raw` in component templates.** All user-visible strings flow through Razor's default HTML encoding (`@variable`). `@Html.Raw(...)` is prohibited in component wrappers — it bypasses auto-escaping and is an XSS vector.

**Every component test file must include an `Html.Raw` grep guard.** When a component test suite (`*SnapshotTests.cs`) is written, it must include a `[Fact]` asserting that `Html.Raw` does not appear in the wrapper `.cshtml` file. This applies to every component — not just the first three written. A component test file without this guard is incomplete.

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

## Home page

The Home page lives at `FieldMark.Web/Pages/Index.cshtml` (`IndexModel` in `Index.cshtml.cs`) and is served at `/`.

**Story 2.10 update:** `GET /` now redirects to `GET /dashboard` for authenticated users. The Compliance Dashboard is the landing page.

**Chrome composition order (AC #2, Story 1.13 — all three stacks must match):**
`<a class="fm-brand-lockup">` → `<div class="ml-auto flex items-center gap-3">` containing `_ThemeToggle` (3-button pill) then `_AvatarMenu`. Any new chrome control added to any stack must be added to all three in the same commit (FR58).

**Role → badge-token mapping** (locked in Story 1.13; source of truth is `FieldMark.Domain/ValueObjects/Role.cs`):

| Role | Token | Label |
|---|---|---|
| `ADMIN` | `danger` | Admin |
| `COMPLIANCE_OFFICER` | `info` | Compliance Officer |
| `INSPECTOR` | `warning` | Inspector |
| `SITE_SUPERVISOR` | `neutral` | Site Supervisor |
| `EXECUTIVE` | `success` | Executive |

The badge `<span class="badge badge-{token}" role="status">{label}</span>` is the first cross-stack visual proof of identity. Never hard-code tokens or labels outside `Role.cs`.

**Tooltip escaping:** Any Razor partial that emits a `data-tooltip` attribute must use `@Html.AttributeEncode(value)` or the Razor `@` expression (which HTML-encodes by default). Never use `@Html.Raw(...)` for tooltip values — raw entities would render as literal `&amp;` in the tooltip text.
