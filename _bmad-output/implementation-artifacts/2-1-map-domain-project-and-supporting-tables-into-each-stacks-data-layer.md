# Story 2.1: Map `domain.project` and supporting tables into each stack's data layer

Status: ready-for-dev

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

- [ ] **Task 1: Read upstream artifacts and confirm posture** (AC: all)
  - [ ] 1.1 Re-read [010_domain_tables.sql:24–95](../../docker/postgres/init/010_domain_tables.sql) — every column, every CHECK constraint, every FK behavior. The DDL is binding.
  - [ ] 1.2 Read [Story 1.7 (.NET Identity wiring)](1-7-wire-asp-net-core-identity-to-dotnet-auth-schema-with-conceptual-roles.md) and [Story 1.8 (Django auth)](1-8-wire-django-built-in-auth-to-django-auth-schema-with-conceptual-role-groups.md) for the existing `AuthDbContext` / `django_auth` schema separation pattern. The new `FieldMarkDbContext` mapping must not leak into the auth schema.
  - [ ] 1.3 Read [Story 1.10 (dev-users seed)](1-10-author-shared-uuid-dev-user-manifest-and-per-stack-idempotent-seed-runners.md) — the dev-user UUIDs are what `project_inspector.user_id` will reference in Story 2.8. AC5 smoke tests can either insert an arbitrary UUID for `user_id` (no FK to enforce) or pick one from the dev-users manifest.
  - [ ] 1.4 Read the existing integration-test fixtures end-to-end:
    - [FieldMark.Tests.Integration/PostgresContainerFixture.cs](../../FieldMark/FieldMark.Tests.Integration/PostgresContainerFixture.cs) and [DomainRollbackSmokeTests.cs](../../FieldMark/FieldMark.Tests.Integration/DomainRollbackSmokeTests.cs) — the AC5 .NET smoke piggybacks on the same `[Collection(PostgresCollection.Name)]` fixture.
    - [fieldmark_py/conftest.py](../../fieldmark_py/conftest.py) `domain_db` fixture — the Django smoke uses this cursor.
    - [fieldmark-go/internal/data/postgres/integration_test.go](../../fieldmark-go/internal/data/postgres/integration_test.go) — the Go smoke shares the `//go:build integration` tag and `openPool(t)` helper.

- [ ] **Task 2: .NET mapping** (AC: #1, #4, #5, #7)
  - [ ] 2.1 Add value object `FieldMark.Domain/ValueObjects/ProjectStatus.cs` as `public enum ProjectStatus { Active, OnHold, Closed }`. No serialization attributes (Domain rule). The enum-string mapping is configured in the EF Core layer.
  - [ ] 2.2 Add entity `FieldMark.Domain/Entities/Project.cs` — properties: `Id (Guid)`, `Code (string)`, `Name (string)`, `Description (string?)`, `Status (ProjectStatus)`, `StartDate (DateOnly)`, `TargetCompletionDate (DateOnly?)`, `ActualClosedAt (DateTimeOffset?)`, `ComplianceScore (int)`, `CreatedAt (DateTimeOffset)`, `UpdatedAt (DateTimeOffset)`. Private setters. Private parameterless ctor for EF Core. Public ctor optional (Story 2.8 will introduce `Project.Create`).
  - [ ] 2.3 Add entities `JobSite.cs`, `ProjectTradeScope.cs`, `ProjectInspector.cs` — property bags per DDL columns; same private-setter discipline.
  - [ ] 2.4 Decide: snake-case via `EFCore.NamingConventions` (register in `Program.cs` against `FieldMarkDbContext`) vs per-property `HasColumnName`. Recommended: register the convention package — `services.AddDbContextPool<FieldMarkDbContext>(opt => opt.UseNpgsql(...).UseSnakeCaseNamingConvention())`. Document the choice in dev notes (other stacks read these column names; convention drift would surface as Django/Go test breakage).
  - [ ] 2.5 Add `FieldMark.Data/Configuration/ProjectConfiguration.cs`:
    - `builder.ToTable("project", "domain")`
    - `builder.HasKey(p => p.Id)`
    - `builder.Property(p => p.Status).HasConversion<string>().HasMaxLength(16)`
    - `builder.Property(p => p.Code).HasMaxLength(32).IsRequired()`
    - `builder.Property(p => p.Name).HasMaxLength(200).IsRequired()`
    - `builder.HasIndex(p => p.Code).IsUnique()` — **wait**: the unique constraint exists in the DDL (`code VARCHAR(32) UNIQUE`). EF Core might pick this up as `HasIndex` and emit a parity diff. Run `make parity` after wiring; if a phantom index appears in the .NET side, switch to `builder.HasAlternateKey(p => p.Code)` (does not create an extra index) or remove the EF declaration entirely (the constraint is DDL-owned). Settle this empirically before merging.
  - [ ] 2.6 Add `JobSiteConfiguration.cs`, `ProjectTradeScopeConfiguration.cs`, `ProjectInspectorConfiguration.cs` — `ToTable` + composite keys + CASCADE relationships per AC1.
  - [ ] 2.7 Update `FieldMark.Data/Context/FieldMarkDbContext.cs`:
    - Add `DbSet<Project>`, `DbSet<JobSite>`, `DbSet<ProjectTradeScope>`, `DbSet<ProjectInspector>` properties.
    - Override `OnModelCreating` (if not already): `modelBuilder.HasDefaultSchema("domain"); modelBuilder.ApplyConfigurationsFromAssembly(typeof(FieldMarkDbContext).Assembly);`
  - [ ] 2.8 Register `FieldMarkDbContext` in `FieldMark.Web/Program.cs` with `AddDbContextPool` reading `FIELDMARK_DATABASE_URL` (or the `Default` connection string already used by `AuthDbContext` — confirm in current `Program.cs`).
  - [ ] 2.9 Confirm `dotnet ef migrations add` is NOT run against `FieldMarkDbContext` — domain schema is infrastructure-owned. If a migration is accidentally generated, delete it and document the near-miss in dev notes.
  - [ ] 2.10 Add `FieldMark.Tests.Integration/ProjectMappingSmokeTests.cs` per AC5 (one round-trip on `Project`; one `count`-query test that covers the three relation tables compiling and reading).
  - [ ] 2.11 Run `dotnet csharpier format .`, `dotnet build`, `dotnet test`, `dotnet test FieldMark.Tests.Integration` — all green.

- [ ] **Task 3: Django mapping** (AC: #2, #4, #5, #7)
  - [ ] 3.1 Verify `fieldmark_py/pyproject.toml` Django version supports `Meta.primary_key` composite PKs (Django 5.2+). If yes, use composite PK; if no, use the `primary_key=True` on `project` field option and document in dev notes.
  - [ ] 3.2 Add to `fieldmark_py/projects/models.py`:
    - `class ProjectStatus(models.TextChoices)` with `ACTIVE = "Active", "Active"`, `ON_HOLD = "OnHold", "OnHold"`, `CLOSED = "Closed", "Closed"`.
    - `class Project(models.Model)` with all DDL columns, `Meta.managed = False`, `Meta.db_table = 'domain"."project'`.
    - `class JobSite(models.Model)`, `class ProjectTradeScope(models.Model)`, `class ProjectInspector(models.Model)` — same `Meta` flags, correct `db_table`, FKs declared as `models.ForeignKey(Project, db_column="project_id", on_delete=models.DO_NOTHING)` (DO_NOTHING because the cascade is DDL-owned; Django must not attempt to enforce or override it).
  - [ ] 3.3 Run `uv run python manage.py makemigrations projects` — assert it outputs `No changes detected` (or any output is restricted to `django_auth`-scoped models, not `domain.*`). If it tries to migrate `domain.*`, the `Meta` flags are wrong.
  - [ ] 3.4 Add `fieldmark_py/projects/tests/__init__.py` and `fieldmark_py/projects/tests/test_mapping.py` per AC5. Mark `@pytest.mark.integration` and use the `domain_db` cursor from `conftest.py` to seed the row, then use the Django ORM to read it back (`Project.objects.using('default').get(pk=id)` — note both the raw insert and the ORM read share the *same* psycopg transaction by using the same connection; or insert via the ORM and read via the ORM, since both go through the same `default` connection).
  - [ ] 3.5 Run `uv run ruff check .`, `uv run mypy .`, `uv run pytest`, `uv run pytest -m integration` — all green.

- [ ] **Task 4: Go mapping** (AC: #3, #4, #5, #7)
  - [ ] 4.1 Add `fieldmark-go/internal/domain/enums/project_status.go` per AC3.
  - [ ] 4.2 Add `fieldmark-go/internal/domain/entities/project.go`, `job_site.go`, `project_trade_scope.go`, `project_inspector.go` — plain structs with exported fields. Use `uuid.UUID` (`github.com/google/uuid`, already in `go.sum`), `time.Time` for timestamps, `*time.Time` for nullable timestamps, `civil.Date` (or `time.Time` rounded to date) for `DATE` columns — pick `time.Time` for simplicity; document the convention in dev notes.
  - [ ] 4.3 Add `fieldmark-go/internal/data/postgres/errors.go` with `ErrProjectNotFound` sentinel.
  - [ ] 4.4 Add `fieldmark-go/internal/data/postgres/projectstore.go` with `ProjectStore` interface (Load, LoadWithRelations) and `projectStorePg` struct backed by `*pgxpool.Pool`. Explicit column lists; explicit `Rows.Scan`. Translate `pgx.ErrNoRows` → `ErrProjectNotFound` in `Load`.
  - [ ] 4.5 Add `fieldmark-go/internal/data/postgres/projectstore_test.go` with `//go:build integration` and the AC5 smoke (insert via `pool.Exec`, load via `projectStorePg.Load`, assert round-trip).
  - [ ] 4.6 No `app/deps.go` wiring this story — `Deps` plumbing for `ProjectStore` lands in Story 2.8 / 2.11 when a handler first needs it. (If this is contentious, wire `Deps.ProjectStore` now — but no handler consumes it yet; YAGNI.)
  - [ ] 4.7 Run `cd fieldmark-go && make check && go test -tags=integration ./internal/data/postgres/...` — all green.

- [ ] **Task 5: Parity and cross-stack verification** (AC: #4, #6, #7)
  - [ ] 5.1 Run `make parity` from repo root. `pg_indexes` diff for `domain.*` must remain at zero. If a new index appears (likely from .NET `HasIndex(p => p.Code)` — see Task 2.5), resolve before merging.
  - [ ] 5.2 Confirm no new file appears in `docs/reference/` or `docs/how-to/` (AC6 guard rail).
  - [ ] 5.3 Run `make test-net-integration test-django-integration test-go-integration` (Postgres must be `make up` first for Django + Go lanes; .NET spins its own container via Testcontainers).

- [ ] **Task 6: Story sign-off** (AC: all)
  - [ ] 6.1 Populate the Sign-off block at the bottom of this story (date, review-round count, reviewer verdict, deferred-work link if any).
  - [ ] 6.2 Update [sprint-status.yaml](sprint-status.yaml) `development_status` for `2-1-...` to `review` (handled by dev-story workflow; mentioned for completeness).

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

_to be filled at implementation time_

### Debug Log References

### Completion Notes List

### File List

## Sign-off

| Field | Value |
|---|---|
| Final review date | _pending_ |
| Total review rounds | _pending_ |
| Final reviewer verdict | _pending_ |
| Deferred-work entries | _pending_ |
| Dev-notes divergences from epic AC | DDL-vs-AC enum casing: implemented PascalCase per DDL; epic AC mentioned SCREAMING_SNAKE_CASE per `research/domain-model.md` §9 which is non-authoritative (root CLAUDE.md: research/ is not maintained). |
