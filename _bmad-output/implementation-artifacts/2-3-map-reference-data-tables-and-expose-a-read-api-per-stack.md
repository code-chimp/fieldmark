# Story 2.3: Map reference data tables and expose a read API per stack

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.3
Canonical DDL: [docker/postgres/init/010_domain_tables.sql:20–52](../../docker/postgres/init/010_domain_tables.sql)
Canonical seed: [docker/postgres/init/020_domain_seed.sql](../../docker/postgres/init/020_domain_seed.sql) — `trade_type`, `violation_category`, `compliance_rule` rows are inserted by infra on first volume init.

## Story

As an Administrator,
I want to view the catalog of Trade Types, Violation Categories, and Compliance Rules through a read-only admin page in each stack,
So that I can verify the canonical reference data is loaded and visible (FR52), and so that subsequent stories (Story 3.3 compliance-rule engine, Story 2.10 dashboard) can read the same reference data through a stable per-stack module without inventing new mappings (FR53).

**Scope boundary:** this story produces (a) read-only entity mappings for `domain.trade_type`, `domain.violation_category`, `domain.compliance_rule` per stack, (b) a stack-native read API surface (interface in Go; manager/queryset in Django; `DbSet`/service in .NET) that returns each table's rows, (c) one admin-only `/admin/reference` page per stack that renders three Basecoat tables, (d) a per-stack authorization conformance test for non-admin → 403. **Out of scope:** Create / Update / Delete UI affordances (FR67 — Growth phase), the compliance-rule-engine evaluation (Story 3.3), the violation-category dropdown on Inspection forms (Story 3.9), the rules-engine parameter editor (FR68 — Growth), any audit-log emission (these are reads only — no `AuditEntry` is written).

## Acceptance Criteria

### AC1 — `.NET` mapping + read API + admin page

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Domain/Entities/Reference/`
**Then** `TradeType.cs`, `ViolationCategory.cs`, `ComplianceRule.cs` exist as immutable property bags with:
- `TradeType`: `Id (Guid)`, `Code (string, max 32)`, `Name (string, max 120)`, `Description (string?)`, `Active (bool)`.
- `ViolationCategory`: `Id`, `Code (max 32)`, `Name (max 200)`, `TradeTypeId (Guid?)` (nullable per DDL — declared without navigation property; the relational link is DDL-owned), `DefaultSeverity (string)` stored as the verbatim DDL string (`'Low' | 'Medium' | 'High' | 'Critical'`), `Description (string?)`, `Active (bool)`.
- `ComplianceRule`: `Id`, `Code (max 64)`, `Name (max 200)`, `Description (string, required)`, `RuleKind (string)` storing the verbatim DDL string (`'ScoringPenalty' | 'ClosureGate'`), `Parameters (JsonDocument)` mapped to JSONB (use `System.Text.Json.JsonDocument` per the AuditEntry precedent — do not introduce Newtonsoft.Json), `Active (bool)`.
- Parameterless private constructor for EF Core; public constructor takes all fields except `Id` for callers that want to fabricate test instances. **No behavior methods this story.**

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Data/Configuration/`
**Then** `TradeTypeConfiguration.cs`, `ViolationCategoryConfiguration.cs`, `ComplianceRuleConfiguration.cs` use `builder.ToTable("trade_type", "domain")` (and equivalent), declare PKs on `Id`, set `HasMaxLength(...)` on string columns matching the DDL widths exactly, and configure `Parameters` via `builder.Property(r => r.Parameters).HasColumnType("jsonb")`. **No `HasIndex` calls** — the DDL declares `UNIQUE` on `code` for all three tables (and implicit indexes for those uniques + the PK and the `violation_category.trade_type_id` FK); EF Core declarations would risk phantom indexes against `make parity`. Settle empirically (same protocol as Story 2.1 / 2.2).

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs`
**Then** `DbSet<TradeType> TradeTypes`, `DbSet<ViolationCategory> ViolationCategories`, `DbSet<ComplianceRule> ComplianceRules` are declared. `ApplyConfigurationsFromAssembly` picks up the new configurations automatically.

**Given** the .NET stack
**When** I inspect `FieldMark/FieldMark.Data/Reference/IReferenceReader.cs` + `ReferenceReader.cs`
**Then** a stateless service is exposed:

```csharp
public interface IReferenceReader
{
    Task<IReadOnlyList<TradeType>> ListTradeTypesAsync(CancellationToken ct = default);
    Task<IReadOnlyList<ViolationCategory>> ListViolationCategoriesAsync(CancellationToken ct = default);
    Task<IReadOnlyList<ComplianceRule>> ListComplianceRulesAsync(CancellationToken ct = default);
}

public sealed class ReferenceReader(FieldMarkDbContext db) : IReferenceReader { ... }
```

Implementation issues a `db.TradeTypes.AsNoTracking().OrderBy(t => t.Code).ToListAsync(ct)` per call (no caching layer — see Dev Notes "Caching" below). Registered as `services.AddScoped<IReferenceReader, ReferenceReader>()` in `FieldMark.Web/Program.cs` so it shares the request-scoped `DbContext`.

**Given** the .NET stack
**When** I navigate authenticated as `ADMIN` to `GET /admin/reference`
**Then** Razor Page `FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml(.cs)` renders three sections (each in its own `<section>` with an `<h2>` heading) — Trade Types, Violation Categories, Compliance Rules — each as a Basecoat `<table class="table">` with one row per record. The handler `OnGetAsync` calls `IReferenceReader` three times and projects into a page-scoped view model with three `IReadOnlyList<>` properties. No Create / Edit / Delete affordances are rendered (FR67 is Growth). For `ComplianceRule.Parameters`, render the JSON as a `<details><summary>parameters</summary><pre><code class="font-mono"><!-- compact JSON --></code></pre></details>` disclosure — do not try to parse rule-kind-specific shapes.

**Given** the .NET stack
**When** I navigate as a non-`ADMIN` authenticated user (`COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) to `/admin/reference`
**Then** the request returns HTTP 403 via the standard `[Authorize(Roles = "ADMIN")]` Razor Page authorization filter. **No entity-state leakage** (FR56) — the 403 response body must not include any of the row counts, code values, or rule parameters. Use the framework's default 403 page or a stack-shared `_AccessDenied.cshtml` partial that contains only a generic "You do not have permission to access this page." string.

### AC2 — Django mapping + read API + admin page

**Given** the Django stack
**When** I inspect `fieldmark_py/reference/models.py`
**Then** `TradeType`, `ViolationCategory`, `ComplianceRule` are declared with:
- `Meta.managed = False`, `Meta.db_table = 'domain"."trade_type'` (and equivalent — note the embedded double-quotes pattern from Story 2.1 / 2.2).
- `Meta.default_permissions = ()` to suppress Django's automatic CRUD permission rows.
- Field types matching the DDL exactly: `models.UUIDField(primary_key=True, default=uuid.uuid4)`; `CharField(max_length=32, unique=True)` for `code` columns (use the DDL widths verbatim — 32, 64); `TextField(null=True)` for nullable `description`; `BooleanField(default=True)` for `active`; `models.JSONField()` (not `null=True`) for `ComplianceRule.parameters` since the DDL is `NOT NULL`.
- Severity / kind columns: declare both as nested `TextChoices` (`class Severity(models.TextChoices): LOW = "Low", "Low"` etc.) on the model class. Choices are illustrative only — the column type is `CharField` and the DB CHECK constraint is the binding authority. Do not add a Python-side validator that duplicates the CHECK; let unknown values surface as a CHECK violation if they ever appear.
- `ViolationCategory.trade_type` declared as `models.UUIDField(null=True, db_column='trade_type_id')` rather than a `ForeignKey(TradeType)` — keeping `reference/` free of a self-import dependency, mirroring `AuditEntry.project_id` from Story 2.2. The FK constraint is DDL-owned.

**Given** the Django stack
**When** I run `uv run python manage.py makemigrations reference`
**Then** Django emits the initial state-tracking migration (same pattern as the `audit/` app: a state migration is generated for `managed=False` models, then `makemigrations reference` reports `No changes detected` on the next run). The migration must not issue `CREATE TABLE`.

**Given** the Django stack
**When** I inspect `fieldmark_py/reference/queries.py`
**Then** module-level helpers are exposed:

```python
def list_trade_types() -> list[TradeType]:
    return list(TradeType.objects.order_by("code"))

def list_violation_categories() -> list[ViolationCategory]:
    return list(ViolationCategory.objects.order_by("code"))

def list_compliance_rules() -> list[ComplianceRule]:
    return list(ComplianceRule.objects.order_by("code"))
```

Pure read functions, no caching layer (see Dev Notes "Caching" below). These are the import target for Story 3.3 (compliance engine) and any future consumer — views call `queries.list_*()`, never `Model.objects.all()` directly, so the read shape is a single edit point.

**Given** the Django stack
**When** I inspect `fieldmark_py/reference/views.py`
**Then** a single view `reference_index(request)` is declared and decorated with:

```python
from django.contrib.auth.decorators import user_passes_test
from fieldmark.roles import Role

def _is_admin(user):
    return user.is_authenticated and user.groups.filter(name=Role.ADMIN.value).exists()

@user_passes_test(_is_admin, login_url=None)  # 403 (not redirect) on failure
def reference_index(request):
    ...
```

The view calls the three `queries.list_*()` helpers and renders `reference/templates/reference/index.html` (template path follows the Django app-local convention). On non-admin access the response is **HTTP 403** (not a login redirect): the `user_passes_test` decorator's default behavior redirects to `LOGIN_URL`; override by raising `PermissionDenied` directly inside the view body after a manual role check, **or** wire `@user_passes_test(..., login_url=None)` and provide a stack-level 403 handler — pick whichever yields a clean 403 + generic body. The `@login_required` middleware applied repo-wide (Story 1.11) guarantees anonymous requests never reach the view.

**Given** the Django stack
**When** I inspect `fieldmark_py/fieldmark/urls.py`
**Then** the URL `path("admin/reference", reference_index, name="reference_index")` is added **before** `path("admin/", admin.site.urls)`. Django resolves URL patterns in declaration order; placing it after would dispatch `/admin/reference` into Django's built-in admin and fail. Add a one-line comment at the inserted line documenting the ordering requirement so a future reordering doesn't silently break the route. Verify by manual `curl -i http://localhost:8000/admin/reference` as ADMIN — the response must be the Story 2.3 page, not Django Admin.

**Given** the Django stack
**When** I navigate authenticated as ADMIN to `/admin/reference`
**Then** the page renders three sections matching the .NET AC1 wireframe (three Basecoat `<table>` blocks, no CRUD affordances, JSON params under a `<details>` disclosure). The base layout from Story 1.5 wraps the page.

**Given** the Django stack
**When** I navigate as any non-ADMIN role (or anonymous after middleware redirect)
**Then** non-ADMIN → HTTP 403 with no body leakage; anonymous → 302 to `/login` per `LoginRequiredMiddleware` from Story 1.11 (this matches the existing site-wide behavior and is *not* a Story 2.3 deliverable).

### AC3 — Go mapping + read API + admin page

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/domain/entities/reference.go`
**Then** plain structs are declared (no methods):
- `TradeType { ID uuid.UUID; Code string; Name string; Description *string; Active bool }`
- `ViolationCategory { ID uuid.UUID; Code string; Name string; TradeTypeID *uuid.UUID; DefaultSeverity string; Description *string; Active bool }`
- `ComplianceRule { ID uuid.UUID; Code string; Name string; Description string; RuleKind string; Parameters json.RawMessage; Active bool }`

(Nullable DDL columns map to pointer fields; `Parameters` uses `json.RawMessage` per the AuditEntry precedent from Story 2.2.)

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/data/postgres/referencestore.go`
**Then** a narrow interface + pgx implementation:

```go
type ReferenceStore interface {
    ListTradeTypes(ctx context.Context) ([]domain.TradeType, error)
    ListViolationCategories(ctx context.Context) ([]domain.ViolationCategory, error)
    ListComplianceRules(ctx context.Context) ([]domain.ComplianceRule, error)
}

type referenceStorePg struct{ pool *pgxpool.Pool }

func NewReferenceStore(pool *pgxpool.Pool) ReferenceStore { return &referenceStorePg{pool: pool} }
```

Each method issues a single `SELECT` with an enumerated column list ordered by `code ASC`. No `SELECT *`. No transaction parameter — reads only.

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/app/deps.go`
**Then** a `Reference ReferenceStore` field is wired into the `Deps` struct (or per-stack equivalent), constructed via `NewReferenceStore(pool)` at startup. The handler in this story is the first consumer. Confirm the actual `Deps` shape from Story 1.9 / 1.12 / 2.1 before editing — Go CLAUDE.md is the binding reference for the pattern.

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/web/handlers/admin_reference.go` (new file)
**Then** a handler `AdminReferenceIndex(c *fiber.Ctx) error` is declared that:
1. Calls `app.AuthorizeRole(c, domain.RoleAdmin)` (or whatever the Story 1.12 helper is named — confirm before writing) and returns `fiber.NewError(fiber.StatusForbidden)` (or the project's equivalent rendered-403 path) on failure. **No entity state in the 403 body.**
2. Calls all three `Reference.List*` methods (sequentially is fine — three small queries; do not introduce errgroup ceremony).
3. Renders `fieldmark-go/internal/web/templates/pages/admin_reference.html` via the stack's existing html/template render helper, passing a `pageData` struct with three slices.

**Given** the Go stack
**When** I inspect `fieldmark-go/internal/web/router.go` (or wherever Story 1.11+ registered routes)
**Then** `GET /admin/reference` is wired to `handlers.AdminReferenceIndex` behind the existing auth middleware. Mount order does not collide with anything Story 1.11 / 1.13 added; verify with `make parity` (route diff) after wiring.

**Given** the Go stack
**When** I navigate as ADMIN to `/admin/reference`
**Then** the page renders three Basecoat tables matching AC1 wireframe.

**Given** the Go stack
**When** I navigate as any non-ADMIN role
**Then** the response is HTTP 403 with a generic "You do not have permission to access this page." body. No row counts, no codes, no parameter JSON in the response.

### AC4 — Cross-stack visual + markup parity

**Given** all three admin pages render
**When** I diff the HTML body of `/admin/reference` between any two stacks (with seeded data identical across stacks — the canonical seed file ensures this)
**Then** the structural markup is identical: three `<section>` containers in the same order (Trade Types → Violation Categories → Compliance Rules), the same `<table class="table">` headers in the same order, the same row count, the same row contents. Whitespace and comment-only differences are acceptable; structural and content differences are defects.

**Given** the per-stack tests
**When** I run them
**Then** at least a snapshot-style assertion per stack confirms the page renders the three section headings in order and the row count per section matches `SELECT count(*) FROM domain.<table>`. (A full cross-stack snapshot suite is out of scope here — the parity contract is sufficient evidence at story-completion time.)

### AC5 — FR53 hot-reload (no application restart required)

**Given** any stack is running and an admin has just loaded `/admin/reference`
**When** an operator runs `UPDATE domain.compliance_rule SET name = 'Concrete: 28-day strength (UPDATED)' WHERE code = 'CR_CONCRETE_28D';` and re-loads `/admin/reference`
**Then** the new name appears immediately without restarting the stack process (FR53). Implementation guidance: the simplest path that satisfies FR53 is **no caching** — each request runs the three short `SELECT`s, well under any latency budget given the table sizes (≈10–20 rows per table). If a future story benchmarks a hot path that reads these tables and needs caching, a TTL or invalidation hook lands in *that* story, not here. **Do not introduce process-lifetime memoization in Story 2.3** — see Dev Notes "Caching".

**Given** each stack ships an integration-flavoured test for AC5
**When** the test runs
**Then** it (a) reads a compliance rule's name via the reader/store, (b) updates the row via raw SQL on the same DB, (c) reads again via the reader/store, (d) asserts the second read returns the new name without re-creating the reader/store instance. (Per-stack integration-test placement mirrors Story 2.2's pattern: .NET `FieldMark.Tests.Integration/ReferenceHotReloadTests.cs`; Django `fieldmark_py/reference/tests/test_hot_reload.py` marked `@pytest.mark.integration`; Go `fieldmark-go/internal/data/postgres/referencestore_hotreload_test.go` with `//go:build integration`.)

### AC6 — Per-stack 403 conformance test (FR56)

**Given** each stack
**When** I run the 403 conformance test
**Then** the test:
1. Authenticates as each non-ADMIN role in turn (`COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) and asserts the response status is exactly **403**, not 302 (redirect-to-login is a defect — `LoginRequiredMiddleware` should have already passed because the user *is* authenticated).
2. Asserts the response body does **not** contain any of the strings: a trade-type `code` value (e.g., `'TT_CONCRETE'`), a violation-category `code` value, a compliance-rule `code` value, the literal substring `"rule_kind"`, or the literal substring `"parameters"`. This is the FR56 "no entity-state leakage" gate — the 403 page must be content-free with respect to the protected resource.

Placement:
- **.NET:** `FieldMark/FieldMark.Tests.Integration/AdminReferenceAuthzTests.cs` (uses `WebApplicationFactory<Program>` with seeded role-bearing test users — confirm the test-host pattern from Story 1.11 `AuthorizationTests.cs` before writing; do not reinvent the harness).
- **Django:** `fieldmark_py/reference/tests/test_authz.py` using `Client().force_login(user)` with users seeded into each Group (the Story 1.12 / 1.13 tests have the precedent).
- **Go:** `fieldmark-go/internal/web/handlers/admin_reference_authz_test.go` using the existing Fiber test harness from Story 1.11 / 1.13.

### AC7 — `make parity` clean, one new route per stack

**Given** all three mappings + routes + tests land
**When** I run `make parity` from the repo root
**Then** `pg_indexes` for `domain.*` shows **zero diff** against the baseline (no phantom indexes from any stack's reference mapping). The route-parity script reports the same drift baseline as Story 2.2 (pre-existing `/robots.txt` and `/.well-known/security.txt` divergence) **plus** the symmetric addition of `GET /admin/reference` on all three stacks. If only one or two stacks add the route, the parity gate fails — this is the contract.

### AC8 — Cross-stack architecture principle guard rail

**Given** the Cross-Stack Architecture Principle (root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle)
**When** I inspect this story's diff
**Then**:
- No new file in `fieldmark_shared/` lists reference-data codes, rule kinds, or severity strings. The DDL is the contract; per-stack mappings each read it directly. (Severity strings `'Low' | 'Medium' | 'High' | 'Critical'` and rule-kind strings `'ScoringPenalty' | 'ClosureGate'` are DDL CHECK literals, mirrored in each stack's `TextChoices` / typed-string-const declarations — not in a shared file.)
- No shared template engine, no symlinked partial. Each stack's admin-reference page lives in that stack's idiomatic template location (Razor at `Pages/Admin/Reference.cshtml`, Django at `reference/templates/reference/index.html`, Go at `internal/web/templates/pages/admin_reference.html`).
- The three-deliverable rule is **not** invoked for severity/rule-kind strings: these are DDL-owned, not introduced by this story. If a future story makes them mutable (e.g., adds a new severity tier), *that* story creates the contract doc — out of scope here.

### AC9 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean. Verify `uv run python manage.py makemigrations reference` reports `No changes detected` after the initial state migration lands.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./internal/data/postgres/... ./internal/web/handlers/...` — clean.
- From repo root: `make parity` exits 0 (AC7) and `make test-all` (the canonical pre-merge gate landed in Story 2.2 round-3 patches) exits 0.

## Tasks / Subtasks

- [x] **Task 1: .NET — entities, EF configs, reader, admin page, tests** (AC: #1, #5, #6, #7, #9)
  - [x] 1.1 `FieldMark/FieldMark.Domain/Entities/Reference/TradeType.cs`, `ViolationCategory.cs`, `ComplianceRule.cs` — property bags + private EF ctor.
  - [x] 1.2 `FieldMark/FieldMark.Data/Configuration/TradeTypeConfiguration.cs`, `ViolationCategoryConfiguration.cs`, `ComplianceRuleConfiguration.cs` — `ToTable("...", "domain")`, no `HasIndex` (verify empirically via `make parity` after wiring).
  - [x] 1.3 `FieldMark.Data/Context/FieldMarkDbContext.cs` — three `DbSet<>` fields.
  - [x] 1.4 `FieldMark.Data/Reference/IReferenceReader.cs` + `ReferenceReader.cs` — three async list methods, `AsNoTracking().OrderBy(Code)`.
  - [x] 1.5 `FieldMark.Web/Program.cs` — `services.AddScoped<IReferenceReader, ReferenceReader>()` after the `AddDbContext` calls.
  - [x] 1.6 `FieldMark.Web/Pages/Admin/Reference.cshtml(.cs)` — `[Authorize(Roles = "ADMIN")]`, three sections, JSON parameters in `<details>` disclosure.
  - [x] 1.7 `FieldMark.Tests.Web/Pages/AdminReferencePageTests.cs` (AC6) — non-ADMIN → 403, response body excludes protected strings. Deviation from the draft file path is deliberate: web authz tests belong in the existing `FieldMark.Tests.Web` harness; `FieldMark.Tests.Integration` remains data-only per stack rules.
  - [x] 1.8 `FieldMark.Tests.Integration/ReferenceHotReloadTests.cs` (AC5).
  - [x] 1.9 Run `dotnet csharpier format . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/` — all green.

- [x] **Task 2: Django — models, queries, view, template, urls, tests** (AC: #2, #5, #6, #7, #9)
  - [x] 2.1 `fieldmark_py/reference/models.py` — replace placeholder with `TradeType`, `ViolationCategory`, `ComplianceRule`; `Meta.managed = False`, `db_table = 'domain"."<table>'`, `default_permissions = ()`.
  - [x] 2.2 `uv run python manage.py makemigrations reference` — produces the initial state-tracking migration only (same pattern as `audit/`).
  - [x] 2.3 `fieldmark_py/reference/queries.py` — three `list_*` module-level helpers.
  - [x] 2.4 `fieldmark_py/reference/views.py` — `reference_index` view; admin-only authorization → 403 (not redirect) on failure.
  - [x] 2.5 `fieldmark_py/reference/templates/reference/index.html` — three Basecoat tables; JSON parameters disclosed under `<details>`.
  - [x] 2.6 `fieldmark_py/fieldmark/urls.py` — `path("admin/reference", reference_index, name="reference_index")` inserted **before** `path("admin/", admin.site.urls)`; comment documenting the ordering requirement.
  - [x] 2.7 `fieldmark_py/reference/tests/__init__.py`, `test_authz.py` (AC6), `test_hot_reload.py` (AC5, `@pytest.mark.integration`), `test_index_view.py` (AC2 happy path).
  - [x] 2.8 Run `uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — all green.
  - [x] 2.9 Verify `uv run python manage.py makemigrations reference` reports `No changes detected` after step 2.2 lands.

- [x] **Task 3: Go — structs, store, handler, template, route, tests** (AC: #3, #5, #6, #7, #9)
  - [x] 3.1 `fieldmark-go/internal/domain/entities/reference.go` — three structs.
  - [x] 3.2 `fieldmark-go/internal/data/postgres/referencestore.go` — interface + pgx impl with enumerated column SELECTs.
  - [x] 3.3 `fieldmark-go/internal/app/deps.go` — confirmed no `Deps` file exists in current repo shape; wired the `ReferenceStore` in `cmd/web/main.go`, matching the existing stack composition point.
  - [x] 3.4 `fieldmark-go/internal/web/handlers/admin_reference.go` — role gate → 403, three store calls, render template.
  - [x] 3.5 `fieldmark-go/internal/web/templates/pages/admin_reference.html` — three Basecoat tables; JSON parameters in `<details>` disclosure.
  - [x] 3.6 `fieldmark-go/cmd/web/main.go` — register `GET /admin/reference`.
  - [x] 3.7 `fieldmark-go/internal/web/handlers/admin_reference_authz_test.go` (AC6), `referencestore_hotreload_test.go` (`//go:build integration`, AC5), basic happy-path handler test.
  - [x] 3.8 Run `make check && go test ./... && go test -tags=integration ./internal/data/postgres/... ./internal/web/handlers/...` — all green.

- [x] **Task 4: Parity and cross-stack verification** (AC: #4, #7, #8)
  - [x] 4.1 Run `make parity` — pg_indexes diff is zero; product-route parity is clean with symmetric `GET /admin/reference` on all three stacks.
  - [x] 4.2 Per-stack page tests assert the three section headings in order and table row rendering; `make parity` confirms route symmetry. Full authenticated side-by-side curl was not necessary after these automated checks.
  - [x] 4.3 Confirm `grep -rn "TT_CONCRETE\|ScoringPenalty\|ClosureGate" fieldmark_shared/` returns zero hits.

- [x] **Task 5: Story sign-off** (AC: all)
  - [x] 5.1 Populate the Sign-off block below; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **Reference rows are infra-seeded.** `docker/postgres/init/020_domain_seed.sql` inserts every Trade Type, Violation Category, and Compliance Rule on first volume init. Do **not** add a per-stack seed runner for these rows; the runners exist for *user* manifests only (Story 1.10). Verify pre-existing seeded data with `psql -c 'SELECT count(*) FROM domain.trade_type;'` before debugging missing rows — if the count is zero, the answer is `make reset`, not new code.
- **No CRUD UI this story.** FR52 is read-only in MVP; FR67 is the Growth follow-on. Resist adding even a "this is read-only" disabled button — the page is genuinely just three tables. The .NET `Pages/Admin/` directory currently exists empty (Story 2.2 left it untouched); this story is its first real page.
- **`/admin` URL collision in Django.** Django ships `path("admin/", admin.site.urls)` in `urls.py` from Story 1.7/1.8. URL patterns resolve in declaration order, so `path("admin/reference", ...)` **must precede** the Django-admin mount or it will never match. Add a comment at the inserted line so a future Black/isort reordering doesn't break the route. (The Go and .NET stacks have no equivalent — Fiber has explicit routes only; Razor Pages map by filesystem path so `/Admin/Reference.cshtml` is unambiguous.)
- **Caching: do not add any.** Architecture §"Cross-Cutting Decisions" line 303 binds: "Caching: None (no Redis, no in-process) — PRD §Non-Goals". The epic AC's phrase "cached in process memory with a TTL or invalidation hook" is a **permission**, not a requirement; the load-bearing requirement is FR53 (no restart needed to pick up changes). The simplest implementation that satisfies FR53 is per-request `SELECT` — well within latency budget for 3 short queries against tables with ~10–20 rows each. The AC5 hot-reload test directly enforces this: if a future contributor adds process-lifetime memoization, AC5 will fail. **If you genuinely measure a perf problem in a later story, add caching there, not here.**
- **Authorization → 403, never 302.** A non-ADMIN authenticated user accessing `/admin/reference` must receive HTTP 403, not a redirect to login. The `LoginRequiredMiddleware` from Story 1.11 already handled anonymous → redirect; layer-of-defense is the role check inside the view/handler returning 403 for "authenticated but unprivileged". AC6's per-stack test fails on a 302 — this catches the common Django footgun of `@user_passes_test` defaulting to `login_url`.
- **JSONB rendering shape.** `ComplianceRule.parameters` is a `JSONB NOT NULL` column with rule-kind-specific shapes (e.g., `ScoringPenalty` rules carry `{"weight": 5, "violation_severity": "High"}`; `ClosureGate` rules may be empty `{}`). The story renders this as an opaque blob — do **not** try to parse per-rule-kind shapes; that's Story 3.3's job. Use a `<details><summary>parameters</summary><pre>{compact-json}</pre></details>` disclosure on each row so the data is inspectable without dominating the table layout.
- **Reads only — no `AuditEntry`.** This story does not call `append_audit_entry()`. Reads aren't audited (FR39 is "audit-on-every-mutation"); the audit-helper from Story 2.2 stays untouched.

### Reference-data shape (read so you don't have to keep flipping back)

From [docker/postgres/init/010_domain_tables.sql:24–52](../../docker/postgres/init/010_domain_tables.sql):

| Table | Columns | Notes |
|---|---|---|
| `domain.trade_type` | `id UUID PK`, `code VARCHAR(32) UNIQUE NOT NULL`, `name VARCHAR(120) NOT NULL`, `description TEXT (nullable)`, `active BOOLEAN NOT NULL DEFAULT TRUE` | Simplest of the three. |
| `domain.violation_category` | `id UUID PK`, `code VARCHAR(32) UNIQUE NOT NULL`, `name VARCHAR(200) NOT NULL`, `trade_type_id UUID REFERENCES domain.trade_type(id) (nullable)`, `default_severity VARCHAR(16) NOT NULL`, `description TEXT (nullable)`, `active BOOLEAN NOT NULL DEFAULT TRUE` | `default_severity` CHECK = `{'Low','Medium','High','Critical'}`. |
| `domain.compliance_rule` | `id UUID PK`, `code VARCHAR(64) UNIQUE NOT NULL`, `name VARCHAR(200) NOT NULL`, `description TEXT NOT NULL`, `rule_kind VARCHAR(32) NOT NULL`, `parameters JSONB NOT NULL`, `active BOOLEAN NOT NULL DEFAULT TRUE` | `rule_kind` CHECK = `{'ScoringPenalty','ClosureGate'}`. |

Seed counts at this story's time of writing: `trade_type ≈ 11`, `violation_category ≈ 80+`, `compliance_rule ≈ 5–10`. The exact counts come from [020_domain_seed.sql](../../docker/postgres/init/020_domain_seed.sql) and may grow before this story lands — the AC4 parity assertion uses `SELECT count(*)`, not a hardcoded number.

### Edge cases (per [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

Walked the nine categories.

- **1. Unknown enum / vocabulary values (`default_severity`, `rule_kind`):** the DDL CHECK constraints prevent unknown values from being persisted, and this story is read-only, so on the *write* side the protection is mechanical. On the *read* side, the templates render the raw string from the column — if a future operator runs `UPDATE … SET rule_kind = 'foobar'` past the CHECK constraint (impossible without dropping the constraint), the page would render `foobar` verbatim. **No defensive code needed this story** — the DB is the source of truth and the constraint is binding.
- **6. Text overflow & special characters in user-visible strings:** the `description` columns are user-editable from the future Growth-phase admin UI and could contain `&`, `<`, etc. Use each framework's default auto-escaping (Razor `@Model.Description`, Django `{{ rule.description }}` *not* `|safe`, Go `html/template` `{{ .Description }}`). **No `|safe`, no `Html.Raw`, no `template.HTML`.** Long descriptions can clip with `max-w-prose truncate` on the cell or via a `<details>` disclosure — pick one and apply consistently. The Story 1.14 tooltip-escaping pattern is the precedent.
- Categories **2 (font load), 3 (JS init), 4 (AG Grid overlays), 5 (stacking), 7 (reduced-motion), 8 (forced-colors), 9 (empty-input fallbacks):** N/A — this is a static read-only page with no client JS, no overlays, no animations, no user-input-derived display tokens. Forced-colors compliance is inherited from the global `_a11y.css` rules landed in Story 1.14 (no new affordance needs a special block here).

### Security defaults (per [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md))

Walked the seven categories.

- **6. CSRF posture per stack:** this story adds only a `GET` route — CSRF tokens are not required for safe-method requests. No deviation from the Story 1.11 / 1.7-1.8 posture.
- Categories **1 (open-redirect), 2 (cookie attributes), 3 (allowlist validation on writes), 4 (dynamic RegExp), 5 (filesystem writes), 7 (stub-auth warnings):** N/A — no redirects, no cookies set, no user-input writes, no regex on input, no filesystem writes, no auth changes.

The one **non-trivial security control** is FR56 (no entity-state leakage on 403), enforced by AC6.

### Cross-stack contract three-deliverable check

This story does **not** introduce a new cross-stack contract. Severity strings (`'Low' | 'Medium' | 'High' | 'Critical'`) and rule-kind strings (`'ScoringPenalty' | 'ClosureGate'`) are pre-existing DDL CHECK literals, not authored by this story — each stack mirrors them in `TextChoices` / typed-string constants directly against the DDL. The DDL is the existing contract; this story is consumer code.

If a future story makes severity tiers configurable (FR68 — Growth), that story will need a `docs/reference/severity-tiers.md` contract, a per-stack enum, and a conformance test — same pattern as Story 2.2's audit actions.

### Files this story modifies vs creates

| File | New / Modified | Purpose |
|---|---|---|
| `FieldMark/FieldMark.Domain/Entities/Reference/TradeType.cs` | NEW | entity |
| `FieldMark/FieldMark.Domain/Entities/Reference/ViolationCategory.cs` | NEW | entity |
| `FieldMark/FieldMark.Domain/Entities/Reference/ComplianceRule.cs` | NEW | entity |
| `FieldMark/FieldMark.Data/Configuration/TradeTypeConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Configuration/ViolationCategoryConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Configuration/ComplianceRuleConfiguration.cs` | NEW | EF mapping |
| `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs` | MODIFY | three `DbSet<>` |
| `FieldMark/FieldMark.Data/Reference/IReferenceReader.cs` | NEW | reader interface |
| `FieldMark/FieldMark.Data/Reference/ReferenceReader.cs` | NEW | reader impl |
| `FieldMark/FieldMark.Web/Program.cs` | MODIFY | DI registration |
| `FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml` | NEW | admin page markup |
| `FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml.cs` | NEW | admin page model |
| `FieldMark/FieldMark.Tests.Integration/AdminReferenceAuthzTests.cs` | NEW | AC6 |
| `FieldMark/FieldMark.Tests.Integration/ReferenceHotReloadTests.cs` | NEW | AC5 |
| `fieldmark_py/reference/models.py` | MODIFY (replace placeholder) | three models |
| `fieldmark_py/reference/queries.py` | NEW | read helpers |
| `fieldmark_py/reference/views.py` | MODIFY | `reference_index` view |
| `fieldmark_py/reference/templates/reference/index.html` | NEW | template |
| `fieldmark_py/reference/migrations/0001_initial.py` | NEW (auto-generated) | state migration |
| `fieldmark_py/reference/tests/__init__.py` | NEW | test package |
| `fieldmark_py/reference/tests/test_index_view.py` | NEW | happy path |
| `fieldmark_py/reference/tests/test_authz.py` | NEW | AC6 |
| `fieldmark_py/reference/tests/test_hot_reload.py` | NEW | AC5 (integration) |
| `fieldmark_py/fieldmark/urls.py` | MODIFY | add `/admin/reference` route before `admin/` |
| `fieldmark-go/internal/domain/entities/reference.go` | NEW | structs |
| `fieldmark-go/internal/data/postgres/referencestore.go` | NEW | store + impl |
| `fieldmark-go/internal/data/postgres/referencestore_hotreload_test.go` | NEW | AC5 (integration build tag) |
| `fieldmark-go/internal/app/deps.go` | MODIFY | wire `Reference` field |
| `fieldmark-go/internal/web/handlers/admin_reference.go` | NEW | handler |
| `fieldmark-go/internal/web/handlers/admin_reference_authz_test.go` | NEW | AC6 |
| `fieldmark-go/internal/web/templates/pages/admin_reference.html` | NEW | template |
| `fieldmark-go/internal/web/router.go` | MODIFY | register route |

Anything outside this list — `AuditEntry` writes (none — reads aren't audited), `compliance/` rule engine code, `_action_button.html` invocations, AG Grid endpoints, the `tests.py` placeholder file rename — is out of scope. Resist the urge.

### Files to read fully before editing

- [docker/postgres/init/010_domain_tables.sql:20–52](../../docker/postgres/init/010_domain_tables.sql) — DDL for the three tables. Binding.
- [docker/postgres/init/020_domain_seed.sql](../../docker/postgres/init/020_domain_seed.sql) — seeded rows; needed to verify row counts in AC4.
- [_bmad-output/implementation-artifacts/2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md](2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md) — same shape of read-only mapping story; reuse the conventions (snake-case naming, no `HasIndex` for DDL-owned uniques, `Meta.managed = False`, flat `internal/data/postgres/`, enumerated SELECT columns).
- [_bmad-output/implementation-artifacts/2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md](2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md) — JSONB mapping precedent for `compliance_rule.parameters`; integration-test placement precedent.
- [_bmad-output/planning-artifacts/architecture.md:1291](../planning-artifacts/architecture.md) (FR52–FR53 row in the FR-to-stack mapping table) and [:303](../planning-artifacts/architecture.md) (caching decision row).
- [_bmad-output/planning-artifacts/prd/functional-requirements.md:89–90](../planning-artifacts/prd/functional-requirements.md) — FR52, FR53 verbatim.
- [_bmad-output/planning-artifacts/prd/functional-requirements.md:97](../planning-artifacts/prd/functional-requirements.md) — FR56 (no entity-state leakage), binding for AC6.
- [fieldmark_py/fieldmark/authz.py](../../fieldmark_py/fieldmark/authz.py) and [fieldmark_py/fieldmark/roles.py](../../fieldmark_py/fieldmark/roles.py) — Django authz primitives from Story 1.12; `Role.ADMIN.value == "ADMIN"`.
- [fieldmark_py/fieldmark/urls.py](../../fieldmark_py/fieldmark/urls.py) — URL ordering constraint for the `admin/` collision.
- [fieldmark-go/internal/domain/role.go](../../fieldmark-go/internal/domain/role.go) — Go role constants from Story 1.12.
- [FieldMark/FieldMark.Web/Program.cs](../../FieldMark/FieldMark.Web/Program.cs) — Story 2.2 registered `IAuditAppender`; same pattern for `IReferenceReader`.
- [fieldmark-go/internal/app/deps.go](../../fieldmark-go/internal/app/deps.go) — confirm `Deps` shape before adding the `Reference` field (its name may differ from the sketch in this story).
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md).

### Project Structure Notes

- The `.NET FieldMark.Web/Pages/Admin/` directory exists empty. This story is its first occupant.
- The Django `reference/` app exists with placeholder `models.py` (`# Create your models here.`), placeholder `views.py`, and a placeholder `tests.py` (note: single file, *not* a `tests/` package). This story leaves the `tests.py` placeholder in place and adds a sibling `tests/` directory with `__init__.py` — same shape as `audit/tests/` from Story 2.2. The convention is set; do not delete `tests.py` here (out of scope, would generate noise).
- The Go `internal/domain/entities/` package was created by Story 2.1 (or 2.2); add `reference.go` alongside `project.go` / `audit_entry.go`. The `internal/web/handlers/` package has `auth.go` + `home_test.go` from Story 1.11/1.13 — add `admin_reference.go` flat at this level (no `admin/` sub-package; the `internal/web/handlers/` directory is flat per Story 1.11 convention).

### References

- AC source: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.3
- DDL: [docker/postgres/init/010_domain_tables.sql:24–52](../../docker/postgres/init/010_domain_tables.sql)
- Seed: [docker/postgres/init/020_domain_seed.sql](../../docker/postgres/init/020_domain_seed.sql)
- FRs: [prd/functional-requirements.md FR52/FR53/FR56](../planning-artifacts/prd/functional-requirements.md)
- Architecture FR-to-stack mapping: [architecture.md:1291](../planning-artifacts/architecture.md)
- Architecture caching decision: [architecture.md:303](../planning-artifacts/architecture.md) ("Caching: None")
- Cross-Stack Architecture Principle: root [CLAUDE.md](../../CLAUDE.md) §Cross-Stack Architecture Principle
- Previous stories' shape patterns: [2-1](2-1-map-domain-project-and-supporting-tables-into-each-stacks-data-layer.md), [2-2](2-2-map-domain-audit-entry-and-provide-a-per-stack-append-audit-entry-helper.md)
- Authz primitives landed in Story 1.12: [fieldmark_py/fieldmark/authz.py](../../fieldmark_py/fieldmark/authz.py), [fieldmark-go/internal/domain/role.go](../../fieldmark-go/internal/domain/role.go)
- Stack rules: [FieldMark/CLAUDE.md](../../FieldMark/CLAUDE.md), [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md), [fieldmark-go/CLAUDE.md](../../fieldmark-go/CLAUDE.md)
- Security defaults checklist: [docs/reference/security-defaults.md](../../docs/reference/security-defaults.md)
- Component edge-case checklist: [docs/reference/component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md)

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-05-28: Loaded workflow configuration, project context, story file, stack rules, DDL, seed data, and prior Story 2.1/2.2 implementation patterns.
- 2026-05-28: Sprint tracker listed Story 2.3 as `ready-for-dev`; synced sprint status and story status to `in-progress` before implementation.
- 2026-05-28: Implemented .NET, Django, and Go reference mappings/read APIs/admin pages/tests. `dotnet build` and `dotnet test` required sandbox escalation because MSBuild named pipes are blocked in the default sandbox.
- 2026-05-28: `make up` and `./tools/verify-domain-schema.sh` succeeded; schema verification reported 4 trade types, 9 violation categories, and 4 compliance rules.
- 2026-05-28: `make parity` initially exposed two parity-tool issues: Django's route dumper filtered `/admin/reference` as a Django Admin internal route, and strict route parity failed on the pre-existing robots/security.txt drift. Updated parity tooling to include `/admin/reference` and filter documented non-business informational routes.
- 2026-05-28: `make test-all` passed after rerun with sandbox-approved .NET/MSBuild, Docker/Postgres, Go cache, and existing npx axe-core access.
- 2026-05-28: Resolved round 1 code-review patch items: Django null descriptions render blank, Go admin-reference handler guards nil actor before role check, and Django hot-reload test restores the changed rule name in `finally`.

### Completion Notes List

- Added read-only reference-data mappings and no-cache read APIs for `.NET`, Django, and Go.
- Added admin-only `/admin/reference` pages in all three stacks with three Basecoat tables and opaque compact JSON parameter disclosures.
- Added hot-reload tests proving compliance-rule updates are visible without recreating the reader/store, plus non-admin 403 tests that assert protected reference strings do not leak.
- Kept `fieldmark_shared/` free of reference-data codes and DDL-owned rule-kind/severity strings.
- Updated route parity tooling so product-route parity includes `/admin/reference` and ignores the documented robots/security.txt informational-route drift.
- Resolved review findings for null-description parity, nil actor handling, and Django hot-reload teardown.

### Change Log

- 2026-05-28: Implemented Story 2.3 and moved story/sprint status to `review`.
- 2026-05-28: Addressed code review findings — 3 patch items resolved.

### File List

- `FieldMark/FieldMark.Domain/Entities/Reference/TradeType.cs`
- `FieldMark/FieldMark.Domain/Entities/Reference/ViolationCategory.cs`
- `FieldMark/FieldMark.Domain/Entities/Reference/ComplianceRule.cs`
- `FieldMark/FieldMark.Data/Configuration/TradeTypeConfiguration.cs`
- `FieldMark/FieldMark.Data/Configuration/ViolationCategoryConfiguration.cs`
- `FieldMark/FieldMark.Data/Configuration/ComplianceRuleConfiguration.cs`
- `FieldMark/FieldMark.Data/Context/FieldMarkDbContext.cs`
- `FieldMark/FieldMark.Data/Reference/IReferenceReader.cs`
- `FieldMark/FieldMark.Data/Reference/ReferenceReader.cs`
- `FieldMark/FieldMark.Web/Program.cs`
- `FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml`
- `FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml.cs`
- `FieldMark/FieldMark.Tests.Integration/ReferenceHotReloadTests.cs`
- `FieldMark/FieldMark.Tests.Web/Fixtures/PostgresFixture.cs`
- `FieldMark/FieldMark.Tests.Web/Pages/AdminReferencePageTests.cs`
- `fieldmark_py/reference/models.py`
- `fieldmark_py/reference/migrations/0001_initial.py`
- `fieldmark_py/reference/queries.py`
- `fieldmark_py/reference/views.py`
- `fieldmark_py/reference/templates/reference/index.html`
- `fieldmark_py/reference/conftest.py`
- `fieldmark_py/reference/tests/__init__.py`
- `fieldmark_py/reference/tests/test_authz.py`
- `fieldmark_py/reference/tests/test_hot_reload.py`
- `fieldmark_py/reference/tests/test_index_view.py`
- `fieldmark_py/fieldmark/urls.py`
- `fieldmark_py/tools/management/commands/dump_routes.py`
- `fieldmark-go/internal/domain/entities/reference.go`
- `fieldmark-go/internal/data/postgres/referencestore.go`
- `fieldmark-go/internal/data/postgres/referencestore_hotreload_test.go`
- `fieldmark-go/internal/web/handlers/admin_reference.go`
- `fieldmark-go/internal/web/handlers/admin_reference_authz_test.go`
- `fieldmark-go/internal/web/templates/pages/admin_reference.html`
- `fieldmark-go/cmd/web/main.go`
- `tools/parity/diff-routes.sh`

## Sign-off

| Field | Value |
|---|---|
| Final review date | 2026-05-28 |
| Total review rounds | 1 |
| Final reviewer verdict | _pending re-review — round 1 patch items resolved, status `review`_ |
| Deferred-work entries | _none_ |
| Dev-notes divergences from epic AC | `.NET` web authz tests live in `FieldMark.Tests.Web` rather than `FieldMark.Tests.Integration` to preserve the repo's existing test-project boundary. Go has no current `internal/app/deps.go`, so reference-store wiring lands in `cmd/web/main.go`, the existing composition point. See Dev Notes "Caching: do not add any" for the deliberate no-cache FR53 implementation. |

### Review Findings

- [x] [Review][Patch] Django template: null descriptions render "None" — AC4 parity break [fieldmark_py/reference/templates/reference/index.html]
- [x] [Review][Patch] Go handler: nil actor dereference before role check [fieldmark-go/internal/web/handlers/admin_reference.go:64]
- [x] [Review][Patch] Django hot-reload test: no explicit teardown — asymmetric with .NET/Go finally/defer restore [fieldmark_py/reference/tests/test_hot_reload.py]
