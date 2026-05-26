# Story 2.1: Map `domain.project` and supporting tables into each stack's data layer

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.1
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) lines 24–95

## Story

As a developer building Project-related features in any stack,
I want each stack's data layer to read (and minimally probe) `domain.project`, `domain.job_site`, `domain.project_trade_scope`, and `domain.project_inspector` against the existing canonical DDL,
So that subsequent Epic 2 stories (2.2 audit helper, 2.8 create form, 2.9 grid, 2.11 detail, 2.12 transitions) can implement Project behavior without inventing schema.

**Scope boundary:** this is a *data-layer mapping* story. Out of scope: handlers, routes, view models, templates, write methods beyond what AC #5's smoke test exercises (write methods land in Story 2.8 / 2.12). Entity behavior methods (`Project.create`, `place_on_hold`, etc.) are introduced in their consuming stories — this story creates the *types* and *table mappings* only.

## Acceptance Criteria

### AC1 — .NET mapping (`FieldMark.Data/Configuration/`)

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Data/Configuration/`
**Then** `ProjectConfiguration.cs`, `JobSiteConfiguration.cs`, `ProjectTradeScopeConfiguration.cs`, `ProjectInspectorConfiguration.cs` exist as `IEntityTypeConfiguration<T>` implementations using `builder.ToTable("<table>", "domain")` (e.g. `"project"`, `"job_site"`, `"project_trade_scope"`, `"project_inspector"`)
**And** column names are `snake_case` (via per-property `HasColumnName` *or* the `EFCore.NamingConventions` `UseSnakeCaseNamingConvention()` hook registered in `Program.cs` against `FieldMarkDbContext` — pick one, document the choice in dev notes)
**And** enum-typed columns (`Project.Status`) use `HasConversion<string>()` so the storage form is a string and the in-memory form is a `ProjectStatus` enum.

**Status-enum storage values:** the persisted strings MUST match the existing DDL CHECK constraint on `domain.project.status` exactly — `'Active'`, `'OnHold'`, `'Closed'` (PascalCase, per [010_domain_tables.sql:71](../../docker/postgres/init/010_domain_tables.sql)). The epic AC mentions "`SCREAMING_SNAKE_CASE`" per `domain-model.md` §9; the DDL is binding (hard-rule: infrastructure owns `domain` schema). Implement what the DDL says, note the AC-vs-DDL divergence in dev notes, and do not "fix" the DDL.

**Composite keys:** `ProjectTradeScope` PK is `(project_id, trade_type_id)`; `ProjectInspector` PK is `(project_id, user_id)`. Configure via `builder.HasKey(x => new { x.ProjectId, x.TradeTypeId })` / `new { x.ProjectId, x.UserId }`.

**FK behavior:** `ON DELETE CASCADE` is in the DDL for `job_site.project_id`, `project_trade_scope.project_id`, `project_inspector.project_id`. Mirror with `.OnDelete(DeleteBehavior.Cascade)` so the relational model matches; do not weaken to `Restrict`.

**No FK to auth schemas:** `project_inspector.user_id` is an opaque UUID with no relational FK to any `*_auth` table (ADR-012, see DDL comment at line 91). Map as a plain `Guid` property — do *not* configure a navigation property to any Identity user type.

**Indexes:** the DDL declares no indexes on `domain.project` / `job_site` / `project_trade_scope` / `project_inspector` beyond their PKs. Do not introduce EF Core `HasIndex(...)` calls — that would create a phantom index expectation against `make parity`.

**Entities (also in scope):** add `FieldMark/FieldMark.Domain/Entities/Project.cs`, `JobSite.cs`, `ProjectTradeScope.cs`, `ProjectInspector.cs` and `FieldMark/FieldMark.Domain/ValueObjects/ProjectStatus.cs`. Entities are property bags with private setters and a parameter-validating constructor (or `EF`-only ctor). **No behavior methods this story.** Domain stays free of EF Core references (root .NET CLAUDE.md hard rule).

**DbContext wiring:** add `DbSet<Project> Projects`, `DbSet<JobSite> JobSites`, `DbSet<ProjectTradeScope> ProjectTradeScopes`, `DbSet<ProjectInspector> ProjectInspectors` to `FieldMarkDbContext` and ensure `OnModelCreating` picks up the new `IEntityTypeConfiguration<T>` classes (typically `modelBuilder.ApplyConfigurationsFromAssembly(typeof(FieldMarkDbContext).Assembly)`).

**DI registration:** if `FieldMarkDbContext` is not yet registered in `FieldMark.Web/Program.cs`, register it via `AddDbContextPool<FieldMarkDbContext>` reading `FIELDMARK_DATABASE_URL` from env (same convention as `AuthDbContext`). The connection string targets the canonical `fieldmark` database — *not* a per-schema search_path. The `domain` schema is reached via `HasDefaultSchema("domain")` on `FieldMarkDbContext` (see architecture.md line 1051).

### AC2 — Django mapping (`projects/models.py`)

**Given** the Django stack
**When** I inspect `fieldmark_py/projects/models.py`
**Then** `Project`, `JobSite`, `ProjectTradeScope`, `ProjectInspector` are declared as Django models with:
- `class Meta: managed = False`
- `db_table = 'domain"."project'` (and equivalent for the other three tables — note the embedded double-quotes that force Postgres to read `domain` as the schema name)
- field types match the canonical DDL exactly: `UUIDField(primary_key=True)`, `CharField(max_length=...)` matching the DDL widths (32 for `code`, 200 for `name`, 16 for `status`), `TextField(null=True)` for nullable text, `DateField`, `DateTimeField`, `IntegerField` (with `MinValueValidator(0) / MaxValueValidator(100)` on `compliance_score` mirroring the DDL CHECK constraint)
- `Project.status` uses `TextChoices` (Django idiom — Django CLAUDE.md §Coding Standards) with the same `'Active' / 'OnHold' / 'Closed'` literal values as the DDL CHECK constraint; the enum class is `ProjectStatus(models.TextChoices)`.

**Composite-PK pain point:** Django models require a single primary key. `domain.project_trade_scope` and `domain.project_inspector` have composite PKs. Options:
1. Declare the model with `unique_together = (("project", "trade_type"),)` (or equivalent) and let Django invent a virtual integer PK — **rejected**, because `Meta.managed = False` means Django won't create that integer column and reads will fail when the table doesn't have an `id`.
2. Use Django 5.2+ `Meta.primary_key = ('project_id', 'trade_type_id')` composite PK — preferred if the project's Django version supports it (check `pyproject.toml`).
3. Pick one field as `primary_key=True` (e.g. `project = models.ForeignKey(..., primary_key=True)`) and document the asymmetry — last resort.

Pick option 2 if supported; if not, pick option 3 and document. Test via the AC5 smoke that round-trip read works on a seeded row.

**Module location:** place `JobSite`, `ProjectTradeScope`, `ProjectInspector` in `fieldmark_py/projects/models.py` (same app — they all belong to the Project aggregate per Django CLAUDE.md §Project Structure "Apps map to bounded contexts").

**No migration generated:** running `uv run python manage.py makemigrations projects` after adding these models MUST produce no migration output for the `domain.*` tables (because `managed = False`). Verify this manually before commit — if Django emits a migration that touches `domain.*`, the `Meta` flags are wrong. (Django CLAUDE.md hard rule: "Django migrations are scoped to `django_auth` only.")

### AC3 — Go mapping (`internal/data/postgres/`)

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/data/postgres/`
**Then** the package contains:
- `fieldmark-go/internal/domain/entities/project.go` — `Project`, `JobSite`, `ProjectTradeScope`, `ProjectInspector` structs (plain field bags, no methods this story — same scope rule as .NET)
- `fieldmark-go/internal/domain/enums/project_status.go` — `ProjectStatus` string-typed enum with constants `ProjectStatusActive = "Active"`, `ProjectStatusOnHold = "OnHold"`, `ProjectStatusClosed = "Closed"` (matching DDL CHECK literals exactly)
- `fieldmark-go/internal/data/postgres/projectstore.go` — `ProjectStore` interface and a `projectStorePg` struct that satisfies it via `*pgxpool.Pool` (per Go CLAUDE.md §Layer Responsibilities: narrow per-aggregate Store interface, concrete pgx implementation, no generic `Repository[T]`)

**Interface — read methods only for this story:**
```go
type ProjectStore interface {
    Load(ctx context.Context, id uuid.UUID) (*domain.Project, error)
    LoadWithRelations(ctx context.Context, id uuid.UUID) (*domain.Project, []domain.JobSite, []domain.ProjectTradeScope, []domain.ProjectInspector, error)
}
```
No `Save`, no `Create`, no `Update`, no `Delete`. Writes land in Story 2.8 (`Create`) and 2.12 (`Save` for transitions).

**SQL column lists:** every `SELECT` must enumerate columns explicitly — no `SELECT *`. Column lists must match the canonical DDL exactly (`id, code, name, description, status, start_date, target_completion_date, actual_closed_at, compliance_score, created_at, updated_at` for `domain.project`). Drift here breaks `make parity` indirectly via Story 2.9's grid response shape.

**Composite-PK reads:** `LoadWithRelations` issues separate `SELECT` queries against `domain.job_site WHERE project_id = $1`, `domain.project_trade_scope WHERE project_id = $1`, `domain.project_inspector WHERE project_id = $1`. Do not introduce a JOIN-and-deduplicate pattern; the three side queries are clearer and pgx handles them without N+1 concerns at this read volume.

**No ORM, no sqlc, no scanner generation:** explicit `pgx` `Rows.Scan(&p.ID, &p.Code, …)` per row. The Go CLAUDE.md is explicit ("explicit SQL via pgx; … No generic `Repository[T]`").

**Sentinel error for not-found:** export `var ErrProjectNotFound = errors.New("project not found")` in `internal/data/postgres/errors.go` (create the file if absent — used by Load when `pgx.ErrNoRows` returns). Handler stories (2.11) will translate this to HTTP 404.

**Subdir convention:** the existing `internal/data/postgres/integration_test.go` lives at `internal/data/postgres/`. Place the new files at the same level (`internal/data/postgres/projectstore.go`, `projectstore_test.go`). Do **not** create a `stores/` sub-package — the architecture directory diagram (line 1207) shows a flat `internal/data/` shape; the actual repo uses `internal/data/postgres/` flat, and the existing integration test confirms it. Honor what's there.

### AC4 — `make parity` clean

**Given** all three mappings exist
**When** I run `make parity`
**Then** `pg_indexes` for `domain.*` shows zero diff against the canonical inventory (no new indexes introduced from any stack's mapping)
**And** the routes diff also stays clean (this story introduces zero new routes).

Parity tooling lives at [tools/parity/](../../tools/parity/). Index snapshot lives at [_bmad-output/implementation-artifacts/_parity-snapshots/](_parity-snapshots/) — review the existing baseline before running.

### AC5 — Per-stack round-trip smoke test

**Given** each stack's integration test lane (from Epic 1 retro action item A3)
**When** I run `make test-net-integration`, `make test-django-integration`, `make test-go-integration`
**Then** a smoke test per stack:
1. Inserts a `domain.project` row via raw SQL (using the existing transactional fixture — `PostgresContainerFixture` for .NET, `domain_db` cursor fixture for Django, `pool` from `openPool(t)` for Go) with a unique `code`, all required columns populated.
2. Loads that row through the new mapping (`FieldMarkDbContext.Projects.SingleAsync(p => p.Id == id)` / `Project.objects.get(pk=id)` / `projectStorePg.Load(ctx, id)`).
3. Asserts every column round-trips: `id`, `code`, `name`, `description`, `status` (as the enum value, not the string), `start_date`, `target_completion_date`, `actual_closed_at`, `compliance_score`, `created_at`, `updated_at`.
4. Rolls back (the existing fixture pattern; no data persists between tests).

**One smoke per stack is sufficient — do not over-test.** Round-trip on `Project` covers the enum-converter and date/timestamp mapping that are the genuinely-novel parts; `JobSite`/`ProjectTradeScope`/`ProjectInspector` are plain-typed and don't add test surface beyond verifying their constructors compile and a single `SELECT count(*)` works (one extra `count`-query test per stack is fine; full round-trip is not required for them).

**Naming:** .NET `ProjectMappingSmokeTests.cs`, Django `audit/tests/test_project_mapping.py` *or* a new `projects/tests/test_mapping.py` (preferred — keep Project tests in the projects app; mirror the existing `audit/tests/test_db_rollback.py` pattern), Go `projectstore_test.go` with build tag `//go:build integration`.

### AC6 — Documentation contract guard rail

**Given** the Cross-Stack Architecture Principle (root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle, ratified Epic 1 retro 2026-05-25)
**When** I inspect this story's diff
**Then** **no new file appears in `docs/reference/` or `docs/how-to/`** — this story introduces no new cross-stack contract beyond what the DDL already encodes. The DDL itself is the contract; each stack's mapping is the native implementation; AC5 smoke tests are the conformance gate.

This AC exists to prevent a well-meaning but wrong instinct to "document the mapping contract" — that would duplicate the DDL. (Stories 2.2 / 2.4 / 2.9 / 2.12 DO introduce new contracts and ship the matching `docs/` files. Story 2.1 does not.)

### AC7 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.
- **Go:** `cd fieldmark-go && make check && go test -tags=integration ./internal/data/postgres/...` — clean.
- From repo root: `make parity` exits 0 (AC4).

## Tasks / Subtasks

- [x] **Task 1: Read upstream artifacts and confirm posture** (AC: all)
  - [x] 1.1 Re-read [010_domain_tables.sql:24–95](../../docker/postgres/init/010_domain_tables.sql) — every column, every CHECK constraint, every FK behavior. The DDL is binding.
  - [x] 1.2 Read [Story 1.7 (.NET Identity wiring)](1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) and [Story 1.8 (Django auth)](1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) for the existing `AuthDbContext` / `django_auth` schema separation pattern. The new `FieldMarkDbContext` mapping must not leak into the auth schema.
  - [x] 1.3 Read [Story 1.10 (dev-users seed)](1-10-author-shared-uuid-dev-user-manifest-and-per-stack-idempotent-seed-runners.md) — the dev-user UUIDs are what `project_inspector.user_id` will reference in Story 2.8. AC5 smoke tests can either insert an arbitrary UUID for `user_id` (no FK to enforce) or pick one from the dev-users manifest.
  - [x] 1.4 Read the existing integration-test fixtures end-to-end:
    - [FieldMark.Tests.Integration/PostgresContainerFixture.cs](../../FieldMark/FieldMark.Tests.Integration/PostgresContainerFixture.cs) and [DomainRollbackSmokeTests.cs](../../FieldMark/FieldMark.Tests.Integration/DomainRollbackSmokeTests.cs) — the AC5 .NET smoke piggybacks on the same `[Collection(PostgresCollection.Name)]` fixture.
    - [fieldmark_py/conftest.py](../../fieldmark_py/conftest.py) `domain_db` fixture — the Django smoke uses this cursor.
    - [fieldmark-go/internal/data/postgres/integration_test.go](../../fieldmark-go/internal/data/postgres/integration_test.go) — the Go smoke shares the `//go:build integration` tag and `openPool(t)` helper.

- [x] **Task 2: .NET mapping** (AC: #1, #4, #5, #7)
  - [x] 2.1 Added `FieldMark.Domain/ValueObjects/ProjectStatus.cs` as `public enum ProjectStatus { Active, OnHold, Closed }`. No serialization attributes.
  - [x] 2.2 Added `FieldMark.Domain/Entities/Project.cs` — properties per AC; private setters; private parameterless ctor for EF Core.
  - [x] 2.3 Added `JobSite.cs`, `ProjectTradeScope.cs`, `ProjectInspector.cs` — property bags per DDL columns.
  - [x] 2.4 Snake-case via `EFCore.NamingConventions`. `Program.cs` (lines 66–68) already registers `FieldMarkDbContext` with `.UseSnakeCaseNamingConvention()` — convention was already in place from prior wiring; we did not duplicate it with per-property `HasColumnName`.
  - [x] 2.5 Added `ProjectConfiguration.cs`. Settled the DDL-vs-EF uniqueness question empirically: `HasAlternateKey(p => p.Code)` rather than `HasIndex(...).IsUnique()` — `make parity` pg-indexes remained clean (21 indexes, no .NET-only index introduced).
  - [x] 2.6 Added `JobSiteConfiguration.cs`, `ProjectTradeScopeConfiguration.cs`, `ProjectInspectorConfiguration.cs` — composite keys via `HasKey(x => new { ... })`; CASCADE on the `Project` FK via `OnDelete(DeleteBehavior.Cascade)`. `ProjectInspector.UserId` is a plain `Guid` with no navigation property (ADR-012).
  - [x] 2.7 `FieldMark.Data/Context/FieldMarkDbContext.cs` updated — added 4 `DbSet`s and `OnModelCreating` calling `HasDefaultSchema("domain")` + `ApplyConfigurationsFromAssembly`.
  - [x] 2.8 `Program.cs` already registers `FieldMarkDbContext` with `AddDbContext` (note: spec said `AddDbContextPool`, but the existing convention pairs `AuthDbContext` and `FieldMarkDbContext` with `AddDbContext` from prior wiring — leaving consistent rather than altering an out-of-scope pattern).
  - [x] 2.9 No EF Core migration generated against `FieldMarkDbContext` — domain schema is infrastructure-owned.
  - [x] 2.10 Added `FieldMark.Tests.Integration/ProjectMappingSmokeTests.cs` — `Project` round-trip + relation-table count probe. **Result:** 4/4 integration tests green (2 new + 2 pre-existing).
  - [x] 2.11 `dotnet csharpier check .` clean; `dotnet build` green; `dotnet test` green across Domain (19), Integration (4), Web (28).

- [x] **Task 3: Django mapping** (AC: #2, #4, #5, #7)
  - [x] 3.1 Django 6.0.4 supports `models.CompositePrimaryKey(...)` (Django 5.2+ feature); used it.
  - [x] 3.2 `fieldmark_py/projects/models.py` populated with `ProjectStatus` TextChoices + 4 models. `Meta.managed = False` + `db_table = 'domain"."<table>'` for all four. FKs declared `on_delete=DO_NOTHING` because the cascade is DDL-owned.
  - [x] 3.3 **AC divergence:** `makemigrations projects` reports four `CreateModel` operations even though models are `managed = False`. Django generates state-only migrations for unmanaged models; the schema editor still no-ops at runtime against `domain.*`. The AC wording ("MUST produce no migration output") is impossible to satisfy without further apparatus (custom `MIGRATION_MODULES = {"projects": None}` setting, or empty per-app `migrations/__init__.py`). Confirmed at runtime: no migration file was committed and `migrate projects` is not invoked. See Sign-off "Dev-notes divergences from epic AC" for the resolution.
  - [x] 3.4 Added `fieldmark_py/projects/tests/test_mapping.py` (AC5) + `projects/tests/conftest.py` (overrides `django_db_setup` to use the live `make up` Postgres — same posture as `audit/tests/test_db_rollback.py`, which sidesteps pytest-django's test-DB creation so the canonical init scripts remain the schema source of truth).
  - [x] 3.5 `uv run ruff check .` clean; `uv run mypy .` clean (93 files); `uv run pytest` 51/51 green; `uv run pytest -m integration` 5/5 green (2 new + 3 pre-existing).

- [x] **Task 4: Go mapping** (AC: #3, #4, #5, #7)
  - [x] 4.1 Added `internal/domain/enums/project_status.go` with `ProjectStatus` const block.
  - [x] 4.2 Added `internal/domain/entities/{project,job_site,project_trade_scope,project_inspector}.go` — `time.Time` for both DATE and TIMESTAMPTZ columns; `*time.Time` / `*string` for nullables; `uuid.UUID` for ids.
  - [x] 4.3 Added `internal/data/postgres/errors.go` with `ErrProjectNotFound`.
  - [x] 4.4 Added `internal/data/postgres/projectstore.go` — `ProjectStore` interface, `projectStorePg` impl, explicit column lists, explicit `Rows.Scan`; `pgx.ErrNoRows` → `ErrProjectNotFound`. `LoadWithRelations` issues three side queries (no JOINs) per AC.
  - [x] 4.5 Added `internal/data/postgres/projectstore_test.go` with `//go:build integration` — round-trip + `LoadWithRelations` smoke + not-found case (2 tests).
  - [x] 4.6 No `app/deps.go` wiring (deferred to Story 2.8 / 2.11 per AC).
  - [x] 4.7 `make check` green (fmt-check, vet, staticcheck, test); `go test -tags=integration ./internal/data/postgres/...` green.

- [x] **Task 5: Parity and cross-stack verification** (AC: #4, #6, #7)
  - [x] 5.1 `bash tools/parity/diff-pg-indexes.sh` clean: `OK pg_indexes parity verified (21 indexes)`. Settled the EF-Core unique-index concern via `HasAlternateKey` (Task 2.5) — no phantom index.
  - [x] 5.2 No new file in `docs/reference/` or `docs/how-to/`. AC6 guard rail intact.
  - [x] 5.3 All three integration lanes green: `make test-integration` reports "✓ Integration tests passed across .NET, Django, Go".
  - [x] **Note on `make parity` overall exit:** `make parity` exits non-zero because of a *pre-existing* routes drift (`/robots.txt` and `/.well-known/security.txt` are exposed by Django/Fiber but not .NET). Verified pre-existing by stashing Story 2.1 changes and re-running. This story introduces zero new routes (confirmed). Logged to [deferred-work.md](deferred-work.md) under "Deferred from: Story 2.1 (2026-05-26)".

- [x] **Task 6: Story sign-off** (AC: all)
  - [x] 6.1 Sign-off block populated.
  - [x] 6.2 Sprint status flipped to `in-progress` at start, will be flipped to `review` by the dev-story workflow on Step 9.

### Review Findings

- [x] [Review][Patch] Scope Django live-DB setup so existing `projects/tests` unit tests keep pytest-django isolation [fieldmark_py/projects/tests/conftest.py:23]
  - **Fix:** Deleted `projects/tests/conftest.py`. The `django_db_setup` session-fixture override now lives inside `projects/tests/test_mapping.py` itself (lines 28–41), where pytest fixture-scope rules confine it to tests in that file. Future unit tests added under `projects/tests/` will use pytest-django's default test-database setup unmodified.
- [x] [Review][Patch] Fix Go integration cleanup ordering so the inserted project row is deleted before the pool closes [fieldmark-go/internal/data/postgres/projectstore_test.go:18]
  - **Fix:** Replaced `defer pool.Close()` with `t.Cleanup(pool.Close)` registered BEFORE the row-delete `t.Cleanup`. Because `t.Cleanup` callbacks run in LIFO order, the row-delete now executes against an open pool, then pool.Close runs. Applied to both tests in the file.
- [x] [Review][Patch] Resolve unmanaged Django model migration drift so `makemigrations projects --check` does not keep reporting `0001_initial.py` [fieldmark_py/projects/models.py:30]
  - **Fix:** Added `MIGRATION_MODULES = {"projects": None}` to `fieldmark_py/fieldmark/settings.py` so Django will not generate state-only migrations for the unmanaged domain-mapping app. `uv run python manage.py makemigrations --check --dry-run` now reports "No changes detected" with exit 0. The block is set up for future domain-only apps (audit, inspections, …) to be added the same way.
- [x] [Review][Patch] Read Go project relations from a single stable database snapshot [fieldmark-go/internal/data/postgres/projectstore.go:75]
  - **Fix:** `LoadWithRelations` now opens a single `pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly}` transaction via `pool.BeginTx` and runs the project + three relation queries inside it. All four reads see the same snapshot; a concurrent writer can no longer make the project row disagree with its job_sites / trade_scopes / inspectors. A small `rowReader` interface keeps the scan helpers usable from either the pool (for `Load`) or the tx (for `LoadWithRelations`).
- [x] [Review][Patch] Change the .NET project mapping smoke to roll back its inserted row instead of commit-plus-delete cleanup [FieldMark/FieldMark.Tests.Integration/ProjectMappingSmokeTests.cs:41]
  - **Fix:** Rewrote the smoke to hold one `NpgsqlConnection` + open transaction across the raw INSERT and the EF Core read. `DbContextOptionsBuilder.UseNpgsql(conn)` + `ctx.Database.UseTransactionAsync(tx)` enlists EF Core in the same transaction so the read sees the uncommitted insert; final `tx.RollbackAsync()` means no row ever reaches disk. Added a fresh-connection post-check that counts rows by id and asserts zero — proof the rollback was honored. Pattern now mirrors `DomainRollbackSmokeTests`.

### Review Findings (Rerun 2026-05-27)

- [x] [Review][Patch] Replace the `MIGRATION_MODULES = {"projects": None}` workaround because `makemigrations projects --check --dry-run` now exits with a `ValueError` instead of cleanly proving no migration drift [fieldmark_py/fieldmark/settings.py:140]
  - **Fix:** Removed the `MIGRATION_MODULES` block from `settings.py`; let Django generate the state-only `projects/migrations/0001_initial.py` and committed it. Because every `CreateModel` operation carries `options.managed = False`, `sqlmigrate projects 0001` confirms every operation is `-- (no-op)` — the schema editor will not touch `domain.*` at runtime, but the autodetector is satisfied. Now: `makemigrations projects --check --dry-run` → `No changes detected in app 'projects'` exit 0; `makemigrations --check --dry-run` → `No changes detected` exit 0. The runtime invariant ("Django never CREATEs or ALTERs `domain.*` tables") is preserved by `managed = False` in the migration ops themselves.
- [x] [Review][Patch] Change the Go project mapping smoke to roll back its inserted row instead of commit-plus-delete cleanup [fieldmark-go/internal/data/postgres/projectstore_test.go:36]
  - **Fix:** Two parts. (1) In `projectstore.go`, promoted the previously-internal `rowReader` to an exported `Querier` interface and added a `LoadProjectFrom(ctx, q Querier, id)` helper so callers holding a `pgx.Tx` can drive the same production scan code. (2) Rewrote the round-trip smoke in `projectstore_test.go` to `pool.Begin(ctx)` a writable transaction, INSERT via the tx, call `postgres.LoadProjectFrom(ctx, tx, id)` for the read, then `tx.Rollback`. A post-rollback count-by-id against the pool asserts zero — proof the rollback was honored. `LoadWithRelations` retains its own committed-row smoke (separate test) since it opens its own snapshot tx internally. A `var _ postgres.Querier = (pgx.Tx)(nil)` compile-time check pins the contract.

## Dev Notes

### Critical context (read before writing code)

- **DDL is binding, not the AC.** The epic AC says enum storage is `SCREAMING_SNAKE_CASE`; the canonical DDL at [010_domain_tables.sql:71](../../docker/postgres/init/010_domain_tables.sql) says `'Active', 'OnHold', 'Closed'` (PascalCase). The DDL wins. Implement PascalCase string storage. Add a one-line dev note in the story sign-off block explaining the divergence so the next retro picks it up.
- **No new behavior methods.** `Project.create`, `place_on_hold`, `resume`, `close`, `RecomputeComplianceScore` are introduced by their consuming stories (2.8, 2.12, Epic 6). This story is *mapping only*.
- **No new handlers, routes, view models, templates.** Zero web-layer surface area.
- **No write methods in Go's `ProjectStore` interface.** Reads only. Writes land in 2.8 / 2.12 when the handler stories know what shape they need.
- **`make parity` is the integration smoke that catches schema drift.** If your mapping causes any new entry in either `dump-routes` or `dump-pg-indexes` outputs, you've over-mapped.
- **Composite PK in Django is the only genuinely-novel pain point** in this story. Resolve it deliberately in Task 3.1; don't paper over it with an unmanaged synthetic `id`.
- **EF Core `HasIndex` vs DDL `UNIQUE` constraint** is the next-most-likely surprise. Prefer `HasAlternateKey` or omit the EF declaration entirely; if you let `HasIndex` slip in, `make parity` will catch it but the diff will not be obvious to read.

### Files to read fully before editing

- [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — the canonical schema. Every column, every CHECK, every FK behavior is binding.
- [FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs](../../FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs) — currently a near-empty placeholder; this story is the first to populate it.
- [FieldMark/FieldMark.Data/Configuration/PlaceHolder.cs](../../FieldMark/FieldMark.Data/Configuration/PlaceHolder.cs) — confirm this is a literal placeholder and delete it if so, replacing with the new Configuration classes.
- [fieldmark_py/projects/models.py](../../fieldmark_py/projects/models.py) — currently empty (just placeholder comments).
- [fieldmark_py/audit/models.py](../../fieldmark_py/audit/models.py) — currently empty; Story 2.2 will populate it. Don't pre-empt that work.
- [fieldmark-go/internal/data/postgres/integration_test.go](../../fieldmark-go/internal/data/postgres/integration_test.go) — pattern for the Go smoke test.
- [_bmad-output/implementation-artifacts/epic-1-retro-2026-05-25.md](epic-1-retro-2026-05-25.md) §"Significant Discovery — Architectural Principle Ratified" — three-deliverable rule is *not* triggered by this story (AC6 explains).

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `FieldMark/FieldMark.Domain/ValueObjects/ProjectStatus.cs` | NEW | enum |
| `FieldMark/FieldMark.Domain/Entities/Project.cs` | NEW | entity property bag |
| `FieldMark/FieldMark.Domain/Entities/JobSite.cs` | NEW | entity property bag |
| `FieldMark/FieldMark.Domain/Entities/ProjectTradeScope.cs` | NEW | entity property bag |
| `FieldMark/FieldMark.Domain/Entities/ProjectInspector.cs` | NEW | entity property bag |
| `FieldMark/FieldMark.Data/Configuration/ProjectConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Configuration/JobSiteConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Configuration/ProjectTradeScopeConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Configuration/ProjectInspectorConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Configuration/PlaceHolder.cs` | DELETE if placeholder | replaced by the real configs |
| `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs` | MODIFY | add DbSets, OnModelCreating |
| `FieldMark/FieldMark.Web/Program.cs` | MODIFY (if needed) | register `FieldMarkDbContext` |
| `FieldMark/FieldMark.Tests.Integration/ProjectMappingSmokeTests.cs` | NEW | AC5 .NET smoke |
| `fieldmark_py/projects/models.py` | MODIFY | replace placeholder with Project + 3 relation models + ProjectStatus |
| `fieldmark_py/projects/tests/__init__.py` | NEW | enable tests subpackage |
| `fieldmark_py/projects/tests/test_mapping.py` | NEW | AC5 Django smoke |
| `fieldmark-go/internal/domain/enums/project_status.go` | NEW | enum |
| `fieldmark-go/internal/domain/entities/project.go` | NEW | struct |
| `fieldmark-go/internal/domain/entities/job_site.go` | NEW | struct |
| `fieldmark-go/internal/domain/entities/project_trade_scope.go` | NEW | struct |
| `fieldmark-go/internal/domain/entities/project_inspector.go` | NEW | struct |
| `fieldmark-go/internal/data/postgres/errors.go` | NEW | ErrProjectNotFound sentinel |
| `fieldmark-go/internal/data/postgres/projectstore.go` | NEW | interface + pgx impl |
| `fieldmark-go/internal/data/postgres/projectstore_test.go` | NEW | AC5 Go smoke |

Anything outside this list — handlers, routes, view models, templates, `app/deps.go` wiring, behavior methods, `docs/reference/*`, `fieldmark_shared/components/*` — is out of scope for Story 2.1. Resist the urge.

### Project Structure Notes

- The architecture diagram (architecture.md line 1207) shows Go's `internal/data/` as a flat package. The actual repo nests at `internal/data/postgres/` (where the existing integration test lives). Honor the actual repo layout, not the diagram.
- The architecture diagram shows .NET entities at `FieldMark/FieldMark.Domain/Entities/`; the directory does not yet exist (no entities have been added in Epic 1). Create it.
- The Django `audit/tests/` subpackage already exists (Epic 1 retro A3); the `projects/tests/` subpackage does not. Create it with an empty `__init__.py`.

### Edge cases (per [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories. **None apply to this story** — it introduces no user-facing components, no JS, no fonts, no tooltips, no toaster, no overlays. The checklist is a component checklist; this is a data-layer story.

If a future reviewer thinks Category 1 (unknown enum values) applies because `ProjectStatus` is enum-like: the resolution lives at the *DDL CHECK constraint*, not at the mapping layer. Postgres rejects unknown values; the mapping layer never sees them. The application enum is closed by definition (no "fallback unknown" value); deserialization of an unknown DB value would raise (`InvalidEnumArgumentException` / `ValueError` / unrecognized string in Go) — which is the correct failure mode for impossible data. Document this if it comes up.

### Security defaults (per [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories. **None apply.** This story handles no forms, cookies, redirects, user input, or filesystem writes. Security defaults re-enter the picture in Story 2.8 (project create form: input validation, CSRF, return-target).

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.1
- DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) lines 24–95
- Hard rules: [docs/reference/hard-rules.md](../../docs/reference/hard-rules.md)
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)
- Project structure: [_bmad-output/planning-artifacts/architecture.md](../planning-artifacts/architecture.md) §Project Structure & Boundaries (lines 945–1242)
- Naming conventions: [architecture.md:552–620](../planning-artifacts/architecture.md) §Naming Patterns
- Epic 1 retro (mapping-relevant background): [epic-1-retro-2026-05-25.md](epic-1-retro-2026-05-25.md) — A3 (integration harness), three-deliverable rule (Stories 2.2 / 2.4 / 2.9 / 2.12, *not* 2.1)
- Previous story (hardening): [1-14-harden-design-system-foundation-and-build-tooling-against-known-edge-cases.md](1-14-harden-design-system-foundation-and-build-tooling-against-known-edge-cases.md)

## Dev Agent Record

### Agent Model Used

claude-opus-4-7 (Claude Code dev-story workflow, 2026-05-26)

### Debug Log References

- Initial `dotnet build` failed with FluentAssertions 8.x renaming `BeGreaterOrEqualTo` → `BeGreaterThanOrEqualTo`. Fixed in `ProjectMappingSmokeTests.cs`.
- Initial Django smoke failed with `relation "domain.project" does not exist` because pytest-django created an empty test database. Resolved by overriding `django_db_setup` in `projects/tests/conftest.py` to skip test-DB creation and reuse the live `make up` Postgres — same posture as `audit/tests/test_db_rollback.py`.
- `ruff check` flagged DJ001 (`null=True` on string fields) and DJ008 (missing `__str__`). Added `__str__` methods to all four models; suppressed DJ001 with `# noqa` + comment because the DDL declares those columns nullable and the NULL-vs-empty-string distinction must round-trip.

### Completion Notes List

- Cross-stack shape-parity review (per persistent fact): re-read all three implementations side-by-side after the third stack landed. Verified parity in (a) enum storage strings (`Active`/`OnHold`/`Closed` in all three), (b) column lists, (c) nullable-column choices, (d) composite PKs, (e) no FK to auth schemas. Idiomatic per-stack asymmetries (Go sentinel error vs .NET `SingleAsync` throw vs Django `DoesNotExist`) are acceptable under the Cross-Stack Architecture Principle — each stack's read API stays native.
- AC6 docs guard rail honored: zero new files in `docs/reference/` or `docs/how-to/`. The DDL remains the contract; each stack's mapping is the native implementation; the three AC5 smoke tests are the conformance gate.
- pg_indexes parity stayed clean (21 indexes). The EF Core unique-index trap on `domain.project.code` was avoided by configuring `HasAlternateKey` instead of `HasIndex(...).IsUnique()`.
- `make parity` overall fails on a pre-existing routes drift (`/robots.txt`, `/.well-known/security.txt` on Django + Fiber but not .NET). Confirmed pre-existing via `git stash` + re-run. Logged to deferred-work.md.
- Test totals: .NET — 19 unit + 4 integration + 28 web (51 total, all green). Django — 51 unit + 5 integration (all green). Go — `make check` green + 2 new integration tests green.

### File List

**New:**
- `fieldmark_py/projects/migrations/0001_initial.py` (state-only, all ops marked `managed=False`; sqlmigrate confirms `(no-op)`; added in rerun review).
- `FieldMark/FieldMark.Domain/ValueObjects/ProjectStatus.cs`
- `FieldMark/FieldMark.Domain/Entities/Project.cs`
- `FieldMark/FieldMark.Domain/Entities/JobSite.cs`
- `FieldMark/FieldMark.Domain/Entities/ProjectTradeScope.cs`
- `FieldMark/FieldMark.Domain/Entities/ProjectInspector.cs`
- `FieldMark/FieldMark.Data/Configuration/ProjectConfiguration.cs`
- `FieldMark/FieldMark.Data/Configuration/JobSiteConfiguration.cs`
- `FieldMark/FieldMark.Data/Configuration/ProjectTradeScopeConfiguration.cs`
- `FieldMark/FieldMark.Data/Configuration/ProjectInspectorConfiguration.cs`
- `FieldMark/FieldMark.Tests.Integration/ProjectMappingSmokeTests.cs`
- `fieldmark_py/projects/tests/conftest.py`
- `fieldmark_py/projects/tests/test_mapping.py`
- `fieldmark-go/internal/domain/enums/project_status.go`
- `fieldmark-go/internal/domain/entities/project.go`
- `fieldmark-go/internal/domain/entities/job_site.go`
- `fieldmark-go/internal/domain/entities/project_trade_scope.go`
- `fieldmark-go/internal/domain/entities/project_inspector.go`
- `fieldmark-go/internal/data/postgres/errors.go`
- `fieldmark-go/internal/data/postgres/projectstore.go`
- `fieldmark-go/internal/data/postgres/projectstore_test.go`

**Modified:**
- `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs` — added 4 `DbSet`s and `OnModelCreating`.
- `FieldMark/FieldMark.Tests.Integration/ProjectMappingSmokeTests.cs` — rewritten to share one connection + transaction across raw insert and EF Core read; rolls back instead of commit-plus-delete (review round 1).
- `fieldmark_py/projects/models.py` — replaced placeholder with `ProjectStatus` + 4 models.
- `fieldmark_py/projects/tests/test_mapping.py` — moved `django_db_setup` override into the test file so the live-DB posture is file-scoped (review round 1).
- `fieldmark_py/fieldmark/settings.py` — added `MIGRATION_MODULES` block in review round 1; removed in rerun (replaced with committed state-only migration).
- `fieldmark-go/internal/data/postgres/projectstore.go` — promoted `rowReader` to exported `Querier`; added `LoadProjectFrom` helper for tx-driven reads (rerun review).
- `fieldmark-go/internal/data/postgres/projectstore_test.go` — round-trip smoke now uses `pool.Begin` + `LoadProjectFrom(tx,...)` + `Rollback`; added `LoadWithRelations` smoke and `Querier` compile-time check (rerun review).
- `fieldmark-go/internal/data/postgres/projectstore.go` — `LoadWithRelations` now reads from a single REPEATABLE READ / READ ONLY transaction (review round 1).
- `fieldmark-go/internal/data/postgres/projectstore_test.go` — replaced `defer pool.Close()` with `t.Cleanup(pool.Close)` ordered before the row-delete cleanup (review round 1).
- `_bmad-output/implementation-artifacts/sprint-status.yaml` — flipped 2-1 to `in-progress` then `review` then `in-progress` (round 1) then `review`.
- `_bmad-output/implementation-artifacts/deferred-work.md` — logged pre-existing routes drift under Story 2.1 section.

**Deleted:**
- `FieldMark/FieldMark.Data/Configuration/PlaceHolder.cs` — placeholder, replaced by the four real configs.
- `fieldmark_py/projects/tests/conftest.py` — removed; the live-DB override is now file-scoped inside `test_mapping.py` (review round 1).

### Change Log

- 2026-05-26 — Story 2.1 implemented: `domain.project` + 3 relation tables mapped into each stack's data layer; AC5 smoke tests landed in all three lanes; pg_indexes parity preserved.
- 2026-05-27 — Addressed code review findings: 5 patch items resolved (Django conftest scoping, Go test cleanup ordering, Django migration drift via `MIGRATION_MODULES`, Go `LoadWithRelations` single-snapshot read, .NET smoke rollback pattern).
- 2026-05-27 — Rerun review: 2 patch items resolved (Django migration drift re-fixed by committing the no-op state migration instead of the `MIGRATION_MODULES` workaround; Go smoke rewritten to use tx + rollback via new exported `Querier` interface and `LoadProjectFrom` helper).

## Sign-off

| Field | Value |
|---|---|
| Final review date | _pending review_ |
| Total review rounds | 2 (5 patches resolved round 1, 2 patches resolved rerun, all 2026-05-27) |
| Final reviewer verdict | _pending review_ |
| Deferred-work entries | [deferred-work.md → "Deferred from: Story 2.1 (2026-05-26)"](deferred-work.md) — pre-existing routes-parity drift (`/robots.txt`, `/.well-known/security.txt`). |
| Dev-notes divergences from epic AC | (1) DDL-vs-AC enum casing: implemented PascalCase per DDL; epic AC mentioned SCREAMING_SNAKE_CASE per `research/domain-model.md` §9 which is non-authoritative (root CLAUDE.md: research/ is not maintained). (2) AC2 §"No migration generated": initially flagged as a divergence; review round 1 used `MIGRATION_MODULES = {"projects": None}` but the rerun review found that broke `makemigrations projects --check` with a `ValueError`. **Final resolution (rerun review):** committed the auto-generated state-only `projects/migrations/0001_initial.py`. Every operation in the file carries `managed = False`, so `sqlmigrate projects 0001` is `-- (no-op)` everywhere — schema editor never touches `domain.*` (AC intent preserved) and `makemigrations --check` is clean (autodetector satisfied). (3) AC1 §`AddDbContextPool`: existing Program.cs uses `AddDbContext` consistently for both `AuthDbContext` and `FieldMarkDbContext`; not changed in this story (out-of-scope refactor risk). |
