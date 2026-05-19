# Story 1.7: Wire ASP.NET Core Identity to `dotnet_auth` schema with conceptual roles

Status: done

## Story

As an administrator using the .NET stack,
I want framework-native authentication backed by the `dotnet_auth` schema with the canonical password policy and five conceptual-role records seeded,
So that user identity is owned by .NET, is isolated from `domain.*`, and is ready for the login/logout flow that lands in Story 1.11.

## Acceptance Criteria

1. **Separate `AuthDbContext` configured for `dotnet_auth`.** A new `AuthDbContext : IdentityDbContext<IdentityUser<Guid>, IdentityRole<Guid>, Guid>` lives in `FieldMark.Data/Context/AuthDbContext.cs`. In `OnModelCreating`, after `base.OnModelCreating(modelBuilder)`, the context calls `modelBuilder.HasDefaultSchema("dotnet_auth")`. The context is configured to use `UseSnakeCaseNamingConvention()` in its `DbContextOptions` (registered in `Program.cs`). `FieldMarkDbContext` is **not** modified to carry Identity types — the two contexts remain independent (Architecture D6 + .NET CLAUDE.md `What Belongs Where`).

2. **All seven Identity tables land in `dotnet_auth` with snake_case names.** After `dotnet ef database update` is applied, the `dotnet_auth` schema contains exactly these tables:
   - `dotnet_auth.users`
   - `dotnet_auth.roles`
   - `dotnet_auth.user_roles`
   - `dotnet_auth.role_claims`
   - `dotnet_auth.user_claims`
   - `dotnet_auth.user_logins`
   - `dotnet_auth.user_tokens`

   No Identity table is created in any other schema. Column names are snake_case (e.g., `normalized_user_name`, `concurrency_stamp`, `email_confirmed`). Verified by `psql \dt dotnet_auth.*` and a column spot-check on `dotnet_auth.users`.

3. **Canonical password policy.** `Program.cs` registers Identity with options:
   - `RequireDigit = true`
   - `RequireLowercase = true`
   - `RequireUppercase = true`
   - `RequireNonAlphanumeric = false`
   - `RequiredLength = 10`

   No other Identity options are changed from defaults at this story. (Source: Architecture §Authentication & Security → D6, and Story 1.7 AC in epics.md.)

4. **Initial Identity migration is scoped exclusively to `dotnet_auth`.** `FieldMark.Data/Migrations/Auth/` contains migration files for an initial Identity migration (typical names: `<timestamp>_InitialIdentity.cs`, `<timestamp>_InitialIdentity.Designer.cs`, `AuthDbContextModelSnapshot.cs`). Every `migrationBuilder.CreateTable`, `CreateIndex`, and `AddForeignKey` call in the migration targets `schema: "dotnet_auth"`. Verified by `grep -rn 'schema:' FieldMark/FieldMark.Data/Migrations/Auth/` returning only `"dotnet_auth"` occurrences; `grep -rn '"domain"' FieldMark/FieldMark.Data/Migrations/Auth/` returns **zero** matches.

5. **Five conceptual-role records are seeded idempotently on first run.** When `make run-net` starts the application against a freshly-reset database (after `make reset`), the `dotnet_auth.roles` table contains exactly five rows with the following `name` / `normalized_name` values:
   - `ADMIN` / `ADMIN`
   - `COMPLIANCE_OFFICER` / `COMPLIANCE_OFFICER`
   - `INSPECTOR` / `INSPECTOR`
   - `SITE_SUPERVISOR` / `SITE_SUPERVISOR`
   - `EXECUTIVE` / `EXECUTIVE`

   Running the seeder a second time (restarting the app) produces **no** duplicates, **no** errors, and no row mutations (verified by capturing the row UUIDs after the first run and comparing after the second). Seeding logic is its own method/class invoked once from `Program.cs` after the host is built; it is **not** in a migration.

6. **No login/logout pages or Identity UI scaffolded.** This story wires the schema + options + role seeding only. The Identity Razor Pages UI (`/Identity/Account/Login`, `/Identity/Account/Manage`, etc.) is **not** registered. The `--dump-routes` output continues to list only the application's own routes; no `/identity/*` paths appear. Login, logout, and the unauthenticated-redirect contract land in Story 1.11. (Use `AddIdentityCore` + `AddRoles` + `AddEntityFrameworkStores` + `AddSignInManager` — **not** `AddDefaultIdentity` — to avoid implicit UI registration. See Dev Notes.)

7. **`make parity` exits 0.** After this story lands, `make parity` from repo root still exits 0. The route inventories of all three stacks remain identical to their HEAD-before-this-story state. The `pg_indexes` snapshot for the `domain` schema is unchanged (Identity touches only `dotnet_auth`, not `domain`). Verified by running `make parity` after `make reset && make run-net` succeeds and a `dotnet_auth` schema is populated.

8. **`AuthDbContext` is registered in DI but `FieldMarkDbContext` continues to work unchanged.** `Program.cs` registers both contexts via `AddDbContext` (or `AddDbContextPool`) against the same connection string. The existing `--dump-routes` flag still works (Story 1.3 contract preserved). `dotnet build` succeeds with `TreatWarningsAsErrors=true`. `dotnet test` for `FieldMark.Tests.Domain` continues to pass.

9. **CLAUDE.md `Authentication` section is updated.** The .NET stack `CLAUDE.md` (`FieldMark/CLAUDE.md`) currently states "ASP.NET Core Identity is **deferred by design**." That section is rewritten to reflect that Identity is now wired against `dotnet_auth`, document the `AuthDbContext` ↔ `FieldMarkDbContext` split, the migration-folder convention, and note that login/logout UI is added in Story 1.11.

## Tasks / Subtasks

- [x] Task 1: Add NuGet packages to `FieldMark.Data` and `FieldMark.Web` (AC: #1, #2, #3)
  - [x] 1.1 Add to `FieldMark.Data/FieldMark.Data.csproj`:
    - `Microsoft.AspNetCore.Identity.EntityFrameworkCore` (version matching .NET 10 / EF Core 10.0.7 — confirm latest 10.x with `dotnet add package`; do not pin to a 9.x version)
    - `EFCore.NamingConventions` (latest stable compatible with EF Core 10) — required for `UseSnakeCaseNamingConvention()` per Architecture D1
  - [x] 1.2 Run `dotnet restore` and confirm `dotnet build` is clean before writing code

- [x] Task 2: Create `AuthDbContext` in `FieldMark.Data` (AC: #1, #2)
  - [x] 2.1 Create `FieldMark/FieldMark.Data/Context/AuthDbContext.cs`
  - [x] 2.2 Inherit from `IdentityDbContext<IdentityUser<Guid>, IdentityRole<Guid>, Guid>` (Guid PKs — UUID at rest, per Architecture §Schema ownership "UUIDs generated in app code")
  - [x] 2.3 Override `OnModelCreating(ModelBuilder modelBuilder)`:
    - Call `base.OnModelCreating(modelBuilder)` first (required by `IdentityDbContext`)
    - Call `modelBuilder.HasDefaultSchema("dotnet_auth")` after
  - [x] 2.4 Constructor accepts `DbContextOptions<AuthDbContext>` and forwards to base
  - [x] 2.5 Namespace: `FieldMark.Data.Context` (matches `FieldMarkDbContext`)

- [x] Task 3: Register Identity, both DbContexts, and password policy in `Program.cs` (AC: #1, #3, #6, #8)
  - [x] 3.1 Register `AuthDbContext` via `builder.Services.AddDbContext<AuthDbContext>(options => options.UseNpgsql(connectionString).UseSnakeCaseNamingConvention())` — note `UseSnakeCaseNamingConvention()` is on the options builder, not in `OnModelCreating`
  - [x] 3.2 Apply `UseSnakeCaseNamingConvention()` to the existing `FieldMarkDbContext` registration too (it was missing — adding it now keeps both contexts consistent and aligns with Architecture D1)
  - [x] 3.3 Register Identity with the canonical password policy using `AddIdentityCore` (NOT `AddDefaultIdentity` — see AC #6):
    ```csharp
    builder.Services
        .AddIdentityCore<IdentityUser<Guid>>(options =>
        {
            options.Password.RequireDigit = true;
            options.Password.RequireLowercase = true;
            options.Password.RequireUppercase = true;
            options.Password.RequireNonAlphanumeric = false;
            options.Password.RequiredLength = 10;
        })
        .AddRoles<IdentityRole<Guid>>()
        .AddEntityFrameworkStores<AuthDbContext>()
        .AddSignInManager()
        .AddDefaultTokenProviders();
    ```
  - [x] 3.4 Do **not** add `app.UseAuthentication()` or scaffold Identity UI in this story — wiring the auth middleware and login pages is Story 1.11's scope. Keeping it absent here is what holds AC #6 + AC #7 (parity).
  - [x] 3.5 Confirm `--dump-routes` flag handling still precedes `app.Run()` and works unchanged

- [x] Task 4: Generate and apply the initial Identity migration (AC: #4)
  - [x] 4.1 From `FieldMark/`, run: `dotnet ef migrations add InitialIdentity --context AuthDbContext --project FieldMark.Data --startup-project FieldMark.Web --output-dir Migrations/Auth`
  - [x] 4.2 Open the generated migration file and verify every `CreateTable`, `CreateIndex`, `AddForeignKey` call has `schema: "dotnet_auth"`. If `EFCore.NamingConventions` is correctly wired, table and column names will already be snake_case. **Do not hand-edit the migration to "fix" anything except verify the schema target.**
  - [x] 4.3 Run: `dotnet ef database update --context AuthDbContext --project FieldMark.Data --startup-project FieldMark.Web`
  - [x] 4.4 Verify via `psql`: `\dt dotnet_auth.*` lists the seven tables; spot-check `\d dotnet_auth.users` for snake_case columns (`normalized_user_name`, `email_confirmed`, etc.)
  - [x] 4.5 Run `grep -rn '"domain"' FieldMark/FieldMark.Data/Migrations/Auth/` — must return zero matches

- [x] Task 5: Seed the five canonical roles idempotently (AC: #5)
  - [x] 5.1 Create `FieldMark/FieldMark.Web/SeedData/RoleSeeder.cs` with a static async method (e.g., `SeedAsync(IServiceProvider services, CancellationToken ct)`)
  - [x] 5.2 Inside, resolve `RoleManager<IdentityRole<Guid>>` from the provided scope and iterate the five canonical role names. For each, call `await roleManager.RoleExistsAsync(name)`; if false, `await roleManager.CreateAsync(new IdentityRole<Guid>(name) { Id = Guid.NewGuid() })`. This is the idempotent path — `RoleExistsAsync` short-circuits on subsequent runs.
  - [x] 5.3 Define the five names as a private `static readonly string[]` inside the seeder class. Names exactly: `ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`. Do **not** introduce a shared `Role` enum in `FieldMark.Domain` for this story — Domain has zero outbound references and that work belongs to Story 1.12 (`authz.Can` primitive). At this story the names live as strings in the seeder.
  - [x] 5.4 In `Program.cs`, after `var app = builder.Build();` and before `app.Run()`, call the seeder inside a fresh scope:
    ```csharp
    using (var scope = app.Services.CreateScope())
    {
        await FieldMark.Web.SeedData.RoleSeeder.SeedAsync(scope.ServiceProvider, CancellationToken.None);
    }
    ```
    Place the call **after** the `--dump-routes` early-return so route-dump invocations do not touch the database. Place it **before** `app.Run()`.
  - [x] 5.5 Change `Program.cs` top-level to `await` if not already (the new seed call is async)

- [x] Task 6: Verify build, tests, and parity (AC: #7, #8)
  - [x] 6.1 `cd FieldMark && dotnet build` — zero warnings (Directory.Build.props sets `TreatWarningsAsErrors=true`)
  - [x] 6.2 `cd FieldMark && dotnet test` — `FieldMark.Tests.Domain` still passes
  - [x] 6.3 `make reset && make run-net` (let it start, then Ctrl-C) — confirm migration applies on a fresh volume; query `dotnet_auth.roles` to confirm five rows; restart the app and re-query to confirm idempotence (same five rows, no duplicates)
  - [x] 6.4 `make parity` — exits 0; routes still identical across stacks; `pg_indexes` for `domain.*` unchanged from canonical snapshot
  - [x] 6.5 `cd FieldMark && dotnet run --project FieldMark.Web -- --dump-routes` — output unchanged from HEAD-before-this-story

- [x] Task 7: Update `FieldMark/CLAUDE.md` (AC: #9)
  - [x] 7.1 Rewrite the `## Authentication` section. New content (paraphrased — author it in your own voice; structure below):
    - State: "ASP.NET Core Identity is wired against the `dotnet_auth` schema via `AuthDbContext`. Login/logout pages are not yet added — Story 1.11."
    - Explain: `FieldMark.Data/Context/AuthDbContext.cs` owns the seven Identity tables under `dotnet_auth`. `FieldMarkDbContext` owns `domain.*` mappings (work for later stories). The two contexts are independent — never merge them.
    - Document migration folder: `FieldMark.Data/Migrations/Auth/`. Auth migrations only. `domain.*` is infrastructure-owned and NOT touched by any EF Core migration (re-state ADR-014).
    - Document the EF migration command shape (the `--context AuthDbContext --output-dir Migrations/Auth` flags).
    - Document the five conceptual roles and where they are seeded (`SeedData/RoleSeeder.cs`).
  - [x] 7.2 The existing `## Migration Ownership` section is already correct ("EF Core migrations are scoped to the `dotnet_auth` schema only"). Confirm it still reads correctly given the new wiring; no edit needed unless a sentence has been invalidated.

## Dev Notes

### Brownfield posture — what exists today (read before writing anything)

State of the .NET stack at HEAD of this branch:

- `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs` — single empty context, no DbSets, no schema configured. Independent from auth.
- `FieldMark/FieldMark.Data/FieldMark.Data.csproj` — references `Microsoft.EntityFrameworkCore` 10.0.7 + `Npgsql.EntityFrameworkCore.PostgreSQL` 10.0.1. **Does not** reference `Microsoft.AspNetCore.Identity.EntityFrameworkCore` or `EFCore.NamingConventions` — both must be added in Task 1.
- `FieldMark/FieldMark.Web/Program.cs` — composes a single `FieldMarkDbContext` against `FIELDMARK_DATABASE_URL`. Does **not** call `UseSnakeCaseNamingConvention()` yet. The `--dump-routes` flag handling (Story 1.3) lives after `app.MapRazorPages()`; preserve that early-return path. Add the role-seed call **after** the `--dump-routes` block so route-dump invocations don't touch the DB.
- `FieldMark/FieldMark.Data/Migrations/` does **not** exist yet. Story 1.7 creates `Migrations/Auth/`. There is no `Migrations/Domain/` and there must not be (ADR-014: `domain.*` is infrastructure-owned).
- `FieldMark/CLAUDE.md` lines 78–85 say Identity is "deferred by design" and "Do not introduce Identity during any story that does not explicitly call for it." Story 1.7 **is** that story — rewrite this section.
- `docker/postgres/init/001_schemas.sql` already creates the `dotnet_auth` schema. The schema *exists*; this story populates it.
- `tools/parity/canonical-pg-indexes.txt` is the canonical baseline for `domain.*` indexes. This story must not change it.

### Why `AddIdentityCore`, not `AddDefaultIdentity`

`AddDefaultIdentity` scaffolds Identity's Razor Pages UI under `/Identity/Account/*` (Login, Register, Manage, etc.). That would (a) add ~20 routes that exist only in .NET, breaking AC #7 (`make parity`) and AC #6, and (b) collide with the Basecoat-styled login form that Story 1.11 will hand-author. `AddIdentityCore` + `AddRoles<>` + `AddEntityFrameworkStores<>` + `AddSignInManager` gives you the services (UserManager, RoleManager, SignInManager, password hasher, token providers) without any UI registration. This is the supported decomposition — see `Microsoft.AspNetCore.Identity` source.

`AddDefaultTokenProviders()` is included because password-reset and email-confirmation flows in Story 1.11 will want them. Adding them now costs nothing and avoids a churn migration later.

### Why separate `AuthDbContext` (not one combined context)

Architecture §Project Structure & Boundaries (line 1036) explicitly calls for two DbContexts:

> `FieldMarkDbContext.cs` — maps ALL of `domain.*`; `HasDefaultSchema("domain")`
> `AuthDbContext.cs` — ASP.NET Core Identity; `HasDefaultSchema("dotnet_auth")`

Reasons (a) Identity types extend `IdentityDbContext<...>` which carries its own conventions and would pollute `FieldMarkDbContext` if merged; (b) migration ownership is cleanly bifurcated — `Migrations/Auth/` for the auth context, no `Migrations/Domain/` ever (ADR-014); (c) the cross-stack symmetry argument: Django uses a DB router to point auth tables at `django_auth`; Go owns `fiber_auth` separately when it lands. Two contexts in .NET mirrors that boundary.

### `UseSnakeCaseNamingConvention()` placement

Per the `EFCore.NamingConventions` package, the convention is applied at `DbContextOptions` configuration time:

```csharp
builder.Services.AddDbContext<AuthDbContext>(options =>
    options
        .UseNpgsql(connectionString)
        .UseSnakeCaseNamingConvention());
```

It is **not** something you call inside `OnModelCreating`. With this in place, the migration will be generated with snake_case table and column names automatically — you do not hand-roll a fluent config per Identity table.

Also apply it to `FieldMarkDbContext`'s registration in this story even though that context has no DbSets yet. Forgetting it now means future stories will generate camelCase columns and we'll have to either regenerate everything or write a corrective migration. Apply it once, here, while the context is still empty.

### Guid PKs for Identity

Architecture §Technical Constraints (line 84): "UUIDs generated in app code (not `gen_random_uuid()`)." Identity defaults to `string` PKs (which would store as `TEXT`). We use `IdentityDbContext<IdentityUser<Guid>, IdentityRole<Guid>, Guid>` so PKs are PostgreSQL `UUID` columns, generated as `Guid.NewGuid()` in C# at row insert. This matches Architecture §Schema ownership and aligns with the dev-user manifest format that Story 1.10 will introduce (UUIDv7 strings in JSON → `Guid` in .NET).

### Role names — strings now, enum later

Story 1.12 introduces the `authz.Can` primitive and a `Role` value object in `FieldMark.Domain/ValueObjects/Role.cs` per Architecture §Complete Repository Directory Structure (line 1027). At Story 1.7 we **do not** create that enum — Domain has zero outbound references and adding role constants there ahead of the authz machinery is premature. Keep the names as a `static readonly string[]` inside `RoleSeeder` for now. Story 1.12 will refactor.

### Idempotent seeding pattern

The required pattern is:

```csharp
internal static class RoleSeeder
{
    private static readonly string[] CanonicalRoles =
    {
        "ADMIN",
        "COMPLIANCE_OFFICER",
        "INSPECTOR",
        "SITE_SUPERVISOR",
        "EXECUTIVE",
    };

    internal static async Task SeedAsync(IServiceProvider services, CancellationToken ct)
    {
        var roleManager = services.GetRequiredService<RoleManager<IdentityRole<Guid>>>();
        foreach (var name in CanonicalRoles)
        {
            if (!await roleManager.RoleExistsAsync(name))
            {
                var role = new IdentityRole<Guid>(name) { Id = Guid.NewGuid() };
                var result = await roleManager.CreateAsync(role);
                if (!result.Succeeded)
                {
                    throw new InvalidOperationException(
                        $"Failed to seed role '{name}': {string.Join("; ", result.Errors.Select(e => e.Description))}");
                }
            }
        }
    }
}
```

Throw on failure — silently swallowing role-seed errors hides a misconfigured database from the developer on first run. The exception will surface in the .NET startup logs. Do **not** add retry logic; this runs once at startup against a local Postgres.

### `Program.cs` ordering — important for parity tooling

The `--dump-routes` block must remain the early-return path. Place the role-seed call **after** the `--dump-routes` check:

```csharp
var app = builder.Build();
// ... existing middleware ...

if (args.Contains("--dump-routes"))
{
    FieldMark.Web.Tools.DumpRoutes.Run(app);
    return;
}

using (var scope = app.Services.CreateScope())
{
    await FieldMark.Web.SeedData.RoleSeeder.SeedAsync(scope.ServiceProvider, CancellationToken.None);
}

app.Run();
```

Reason: `--dump-routes` is invoked by `tools/parity/dump-routes-net.sh` and must not require a live database. Story 1.3's review pass explicitly fixed this same shape on the Go side (DB connect moved after flag check); preserve the same property on .NET.

### EF Core CLI invocation specifics

The `dotnet ef` command needs to know **which** context to operate on (we now have two). Always pass `--context AuthDbContext` for auth-schema migrations. The `--output-dir Migrations/Auth` flag puts files in the right folder per Architecture §Repository Directory Structure (line 1042). Document this in the .NET CLAUDE.md per Task 7.

If `dotnet ef` is not installed globally, install: `dotnet tool install --global dotnet-ef --version 10.*` (or via the local `dotnet-tools.json` manifest already present in `FieldMark/`).

### What this story does NOT do (boundary clarification)

- **No login or logout pages.** Story 1.11 (`Login, logout, and unauthenticated-redirect across all three stacks`). Do not scaffold any `Pages/Account/*`.
- **No `app.UseAuthentication()` or `app.UseAuthorization()` policy wiring.** Identity services are registered but the auth middleware is not yet added to the request pipeline. Story 1.11 adds the pipeline; Story 1.12 adds policies. Adding `UseAuthentication()` now would have no observable effect since there are no protected routes — but it would also start emitting `AuthenticationScheme` defaults that Story 1.11's design wants control over. Leave it out.
- **No dev users.** Story 1.10 seeds the six canonical dev users from `docker/postgres/init/seed-uuids/dev-users.json`. The manifest doesn't exist yet — don't create it here.
- **No `Role` enum in `FieldMark.Domain`.** Story 1.12.
- **No `authz.Can` primitive or `ActionButton` trichotomy helper.** Story 1.12.
- **No domain entity mappings.** `FieldMarkDbContext` remains DbSet-free in this story. Adding `UseSnakeCaseNamingConvention()` to its options registration is the only change.

### Anti-patterns that must NOT slip in

- ❌ `AddDefaultIdentity<IdentityUser<Guid>>` — scaffolds UI; breaks AC #6 and AC #7.
- ❌ Hand-writing the migration instead of running `dotnet ef migrations add`. EF Core's generated migration is the contract; if `EFCore.NamingConventions` is wired correctly, snake_case appears for free.
- ❌ Adding role names as `const` strings in a Domain class. Domain has zero outbound references (.NET CLAUDE.md Hard Rules: "`FieldMark.Domain` has zero project or package references"). Role names belong in the seeder for now.
- ❌ Calling the seed method inside a migration's `Up()`. Migrations are schema, not data. The seeder runs from `Program.cs` against a fresh scope at startup.
- ❌ Modifying `tools/parity/canonical-pg-indexes.txt`. The auth schema is not in this snapshot (it filters on `schemaname='domain'` — verified in Story 1.3). Don't touch this file.
- ❌ Using `gen_random_uuid()` or `uuid-ossp` Postgres extensions. UUIDs are generated in C# (`Guid.NewGuid()`) per Architecture §Technical Constraints.
- ❌ Merging `FieldMarkDbContext` and `AuthDbContext`. Architecture and the .NET CLAUDE.md "What Belongs Where" both call for separation.
- ❌ Adding `Identity` references to `FieldMark.Domain` or `FieldMark.Web` directly. Identity belongs in `FieldMark.Data` (the auth context lives there) + `FieldMark.Web` (composition root for DI). Domain stays pure.

### Project Structure Notes

Files this story adds or modifies:

- **New:** `FieldMark/FieldMark.Data/Context/AuthDbContext.cs`
- **New:** `FieldMark/FieldMark.Web/SeedData/RoleSeeder.cs` (and parent `SeedData/` directory)
- **New:** `FieldMark/FieldMark.Data/Migrations/Auth/<timestamp>_InitialIdentity.cs` (+ `.Designer.cs` + `AuthDbContextModelSnapshot.cs`) — generated by `dotnet ef`
- **Update:** `FieldMark/FieldMark.Data/FieldMark.Data.csproj` — add `Microsoft.AspNetCore.Identity.EntityFrameworkCore` and `EFCore.NamingConventions` package references
- **Update:** `FieldMark/FieldMark.Web/Program.cs` — add `AuthDbContext` registration, `UseSnakeCaseNamingConvention()` on both contexts, Identity registration with password options, role seeding call (`await` at top-level)
- **Update:** `FieldMark/CLAUDE.md` — rewrite `## Authentication` section per Task 7

No conflicts with the unified project structure (Architecture §Complete Repository Directory Structure lines 1032–1090). All file locations match the prescribed layout.

### Testing Standards

Per Architecture §Test location and the .NET CLAUDE.md Project Structure:

- **Unit tests (`FieldMark.Tests.Domain`):** This story does not add domain logic. No new unit tests are required. Existing tests must continue to pass.
- **Integration tests (`FieldMark.Tests.Integration`):** Testcontainers + real Postgres. Story 1.7 is foundational wiring; integration tests for the auth flow land in Story 1.11 (login/logout). Do **not** speculatively add login-flow integration tests in this story — they'll need rewriting once 1.11 designs the Razor Pages contract.
- **Manual verification of role seeding** is captured in Task 6.3 (start → query → restart → re-query). That is sufficient for this story; the cross-stack E2E suite covers user-facing role-aware flows in later epics.

If you have spare cycles after the AC pass, an optional addition is a `FieldMark.Tests.Integration/AuthSchemaTests.cs` that spins up a Postgres container, runs the auth migration, and asserts the seven `dotnet_auth.*` tables exist. **Treat this as nice-to-have** — it duplicates what `make reset && make run-net` already verifies, and the Testcontainers fixture pattern hasn't been formalized yet (will be in a later epic).

### References

- [Architecture: Authentication & Security → D6](_bmad-output/planning-artifacts/architecture.md#authentication--security) — schema target, password rules, table names.
- [Architecture: D1 — EF Core driver and naming-convention package](_bmad-output/planning-artifacts/architecture.md#data-architecture) — `EFCore.NamingConventions` is the canonical snake_case mechanism.
- [Architecture: Repository Directory Structure → `FieldMark.Data` and `FieldMark.Web`](_bmad-output/planning-artifacts/architecture.md#complete-repository-directory-structure) — file locations for `AuthDbContext.cs`, `SeedData/`, `Migrations/Auth/`.
- [Architecture: D4 — Auth-schema migrations locked by ADR-012](_bmad-output/planning-artifacts/architecture.md#data-architecture) — `--output-dir Migrations/Auth` against `AuthDbContext`.
- [Architecture: Architectural Boundaries → Authentication / authorization](_bmad-output/planning-artifacts/architecture.md#architectural-boundaries) — opaque UUID refs; no FKs from `domain.*` to auth schemas.
- [docs/hard-rules.md](docs/hard-rules.md) — backend authority, infrastructure-owned domain schema, no repository/UoW abstractions.
- [FieldMark/CLAUDE.md](FieldMark/CLAUDE.md) — .NET-specific rules; Domain has zero outbound references; no MediatR/CQRS/Repository.
- [PRD FR1–FR8 — Authentication & Authorization](_bmad-output/planning-artifacts/prd/functional-requirements.md) — framework-local authentication; conceptual roles.
- [PRD architectural-constraints-prd-binding.md §Authentication & Authorization (ADR-012)](_bmad-output/planning-artifacts/prd/architectural-constraints-prd-binding.md) — schema isolation contract.
- [Story 1.3 implementation artifact](_bmad-output/implementation-artifacts/1-3-establish-tools-parity-and-make-parity-with-per-stack-dump-routes.md) — `--dump-routes` early-return pattern; `make parity` invariants.
- [docker/postgres/init/001_schemas.sql](docker/postgres/init/001_schemas.sql) — `dotnet_auth` schema already created.

### Previous Story Intelligence (Story 1.3 — last `done` story on the .NET side)

Key learnings to carry forward:

- **`--dump-routes` early-return is sacred.** Story 1.3 had to be patched to move the DB connect on the Go side after the flag check, and to replace `Environment.Exit(0)` with `return` in .NET. Don't re-introduce DB work before that early return — see "Program.cs ordering" above.
- **`TreatWarningsAsErrors=true` is enforced** via `Directory.Build.props`. Identity registration extensions return nullable `IdentityBuilder` in some overloads — make sure you don't introduce `CS86xx` warnings.
- **`make parity` is the cross-stack guard.** Any new route in any stack — even an Identity one — will fail it. AC #6 + #7 explicitly call out the parity invariant.
- **CSharpier is the formatter** (Story 1.3 dev notes / `dotnet-tools.json`). Run `dotnet csharpier format .` before committing.

### Git Intelligence

Recent commits (most relevant to this story):

- `d03f0fe feat: e1s3 establish tools parity` — touches `Program.cs`, adds `Tools/DumpRoutes.cs`. The `--dump-routes` flag is the parity hook this story must respect.
- `cbf47e9 feat: e1s2 verified sql init scripts` — the init scripts that create `dotnet_auth` schema (Story 1.2). This story consumes them.
- `a6fac88 feat: e1s1 confirm scaffolds` — the scaffold this story extends. `FieldMarkDbContext` ships empty from this commit.

No prior commit has added Identity, EF Core migrations, or any `dotnet_auth.*` table. This story is the first.

### Latest Technical Information

- **.NET 10** is in use (`TargetFramework=net10.0` everywhere). When adding NuGet packages, use the latest 10.x versions. `Microsoft.AspNetCore.Identity.EntityFrameworkCore` 10.x and `EFCore.NamingConventions` (latest compatible with EF Core 10.x — check `dotnet add package --version "?"` interactively if you're unsure).
- **EF Core 10.0.7** is the pinned version. `EFCore.NamingConventions` typically tracks EF Core major versions — make sure you pick a 10.x release.
- **Npgsql.EntityFrameworkCore.PostgreSQL 10.0.1** is already pinned. Identity stores compose with it transparently.
- The `dotnet-ef` global tool — if not installed, install via `dotnet tool install --global dotnet-ef --version 10.*`. The repository ships a `dotnet-tools.json` (already includes CSharpier) — adding `dotnet-ef` to that manifest is acceptable (and recommended for reproducibility) but not required to land this story.

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6

### Debug Log References

- **EF Core package version mismatch**: `Microsoft.AspNetCore.Identity.EntityFrameworkCore 10.0.8` requires EF Core `>= 10.0.8`, forcing a bump from 10.0.7 → 10.0.8 across all EF Core packages in both `FieldMark.Data` and `FieldMark.Web`.
- **Design DLL version conflict**: `Microsoft.EntityFrameworkCore.Tools 10.0.8` nuspec hard-depends on `Microsoft.EntityFrameworkCore.Design 8.0.0` (packaging quirk). This caused `MissingMethodException` when running `dotnet ef migrations add`. Fixed by adding an explicit `Microsoft.EntityFrameworkCore.Design 10.0.8` reference to `FieldMark.Web.csproj` with `PrivateAssets="all"` to force the correct version.
- **Table names not snake_case**: `EFCore.NamingConventions.UseSnakeCaseNamingConvention()` converts column and constraint names but does NOT strip the explicit `AspNet*` table names set by `IdentityDbContext.OnModelCreating()`. Added `ToTable()` overrides in `AuthDbContext.OnModelCreating()` for all seven Identity entities to produce `users`, `roles`, `user_roles`, etc. per AC #2.
- **CA1725 parameter name mismatch**: `IdentityDbContext`'s `OnModelCreating` uses `builder`, not `modelBuilder`. Fixed to satisfy `TreatWarningsAsErrors=true`.
- **dotnet-ef tool version**: Global `dotnet-ef` was 9.0.9; updated to 10.0.8 to match EF Core 10.0.8 project packages.

### Completion Notes List

- Added `Microsoft.AspNetCore.Identity.EntityFrameworkCore 10.0.8` and `EFCore.NamingConventions 10.0.1` to `FieldMark.Data`. Bumped all EF Core packages to 10.0.8 to match Identity's requirement.
- Added explicit `Microsoft.EntityFrameworkCore.Design 10.0.8` to `FieldMark.Web` to override the stale 8.0.0 version that `Microsoft.EntityFrameworkCore.Tools` nuspec incorrectly pins.
- Created `AuthDbContext` inheriting from `IdentityDbContext<IdentityUser<Guid>, IdentityRole<Guid>, Guid>` with `HasDefaultSchema("dotnet_auth")` and explicit `ToTable()` overrides for all 7 Identity entities.
- Registered both `FieldMarkDbContext` and `AuthDbContext` with `UseSnakeCaseNamingConvention()`. Registered Identity using `AddIdentityCore` (not `AddDefaultIdentity`) per AC #6 to avoid scaffolding UI routes.
- Generated `InitialIdentity` migration under `FieldMark.Data/Migrations/Auth/`. All 7 tables land in `dotnet_auth` with snake_case columns and clean names (`users`, `roles`, etc.). Zero `domain` schema references.
- Created `RoleSeeder.cs` with idempotent `SeedAsync` that seeds 5 canonical roles. Verified: first run creates rows; second run produces no duplicates and no mutations.
- `make parity` exits 0 — routes identical across stacks (4 routes), `pg_indexes` for `domain.*` unchanged (21 indexes).
- `--dump-routes` continues to list only 4 application routes; no `/identity/*` paths added.
- `dotnet build` zero warnings, `dotnet test` all pass.
- `FieldMark/CLAUDE.md` `## Authentication` section rewritten to document the new wiring.

### File List

- **New:** `FieldMark/FieldMark.Data/Context/AuthDbContext.cs`
- **New:** `FieldMark/FieldMark.Web/SeedData/RoleSeeder.cs`
- **New:** `FieldMark/FieldMark.Data/Migrations/Auth/20260519112023_InitialIdentity.cs`
- **New:** `FieldMark/FieldMark.Data/Migrations/Auth/20260519112023_InitialIdentity.Designer.cs`
- **New:** `FieldMark/FieldMark.Data/Migrations/Auth/AuthDbContextModelSnapshot.cs`
- **Updated:** `FieldMark/FieldMark.Data/FieldMark.Data.csproj` — added Identity.EntityFrameworkCore 10.0.8, EFCore.NamingConventions 10.0.1; bumped EF Core to 10.0.8
- **Updated:** `FieldMark/FieldMark.Web/FieldMark.Web.csproj` — added Design 10.0.8 explicit reference; bumped EF Core Tools to 10.0.8
- **Updated:** `FieldMark/FieldMark.Web/Program.cs` — AuthDbContext registration, UseSnakeCaseNamingConvention on both contexts, Identity registration with password options, role seed call
- **Updated:** `FieldMark/CLAUDE.md` — rewrote `## Authentication` section

## Change Log

- 2026-05-19: Story 1.7 implemented — wired ASP.NET Core Identity to `dotnet_auth` schema with `AuthDbContext`, generated `InitialIdentity` migration, seeded five canonical roles idempotently, updated CLAUDE.md authentication documentation.
