# Story 2.14: Reference data read pages for Administrator

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.14
Canonical DDL: [docker/postgres/init/010_domain_tables.sql:20–52](../../docker/postgres/init/010_domain_tables.sql) — `domain.trade_type` (24–30), `domain.violation_category` (32–41), `domain.compliance_rule` (43–52)
Canonical seed: [docker/postgres/init/020_domain_seed.sql](../../docker/postgres/init/020_domain_seed.sql) — reference rows inserted by infra on first volume init.
Depends on:
- **Story 2.3** (reference mappings + per-stack read API + the existing `/admin/reference` overview page + admin-only authz; status: **done**). This story **reuses 2.3's read API verbatim** — `IReferenceReader` (.NET), `reference/queries.py` (Django), `ReferenceStore` (Go) — and **invents no new data layer**.
- **Story 1.12** (`can()` / role gating — the admin-only authz pattern reused here; `ADMIN` role check is already wired in 2.3's handlers).
- **Story 1.5** (base-layout chrome — the new pages render inside it).
- **Story 1.10** (dev-user manifest — test fixtures: `aisha` = ADMIN; `marisol`/`ravi`/`pat`/`kenji` = non-admin roles).

## Story

As an Administrator,
I want a dedicated read-only page for each reference catalog at `/admin/reference/trade-types`, `/admin/reference/violation-categories`, and `/admin/reference/compliance-rules`,
So that I can browse each catalog on its own deep-linkable page (FR52) without scrolling the combined overview, while non-administrators are denied access (FR56) and no Create/Edit/Delete affordances exist (FR67 is Growth).

**Scope boundary.** This story produces, per stack:
- (a) **three new GET routes** — `/admin/reference/trade-types`, `/admin/reference/violation-categories`, `/admin/reference/compliance-rules` — each rendering a single Basecoat `<table class="table">` of that catalog's rows inside the Story 1.5 chrome;
- (b) **admin-only gating** on each route, reusing 2.3's per-stack pattern exactly (non-`ADMIN` → HTTP 403, no entity-state leakage);
- (c) a **shared admin sub-nav** on the three new pages (links to the two sibling catalog pages + back to the `/admin/reference` overview) so the catalogs are discoverable;
- (d) per-stack tests: 200 + correct columns + seeded rows for ADMIN; 403 for each non-admin role; empty-table graceful state; `make parity` clean.

**Out of scope:**
- **Any change to the existing `/admin/reference` overview page** (Story 2.3). It stays byte-identical — its passing tests are untouched (see Decisions note 1). The three new pages are **purely additive**.
- **Create / Update / Delete affordances** — FR67 is Growth phase; no forms, no buttons, no audit (reads only — no `AuditEntry`).
- **The reference data layer** — mappings and read APIs were delivered by Story 2.3 (`done`); this story adds **only** presentation routes that call the existing read methods.
- **Pagination / filtering / sorting** — reference tables are tiny seeded catalogs (a handful of rows each); render all rows in document order (`ORDER BY code`, already the read API's contract). No AG Grid (UX-DR carve-out — Basecoat Table only).
- **New `domain.*` schema or indexes** — `pg_indexes` zero-diff; no DDL change.
- **A new cross-stack `docs/reference/` contract** — this story composes existing contracts (see AC8 / Decisions note 5).

---

## ⚠️ Decisions baked into this story (read first)

These resolve ambiguities the epic AC leaves open; each is implemented as written and flagged in the Sign-off block for reviewer ratification.

1. **Additive, not a repurpose — the 2.3 overview page is left exactly as-is.** Story 2.3 built `/admin/reference` as a single page with three `<section>`s and has passing per-stack tests (e.g. .NET `AdminReferencePageTests` asserts three `<tbody>` + the three section headings). This story **adds** three dedicated routes and does **not** edit, redirect, or delete the overview page or its tests. Rationale: lowest regression surface; the overview remains a valid at-a-glance view, the three pages are the deep-link view. The minor redundancy (same rows visible two ways) is an accepted tradeoff over touching three stacks' shipped templates + tests. (Alternative considered: convert `/admin/reference` into a hub that links out and moves the tables to the sub-pages — rejected for this story because it regresses 2.3's tests across all three stacks for no MVP-functional gain.)

2. **Three explicit concrete routes per stack — never a parameterized route.** The parity tool ([tools/parity/diff-routes.sh](../../tools/parity/diff-routes.sh)) dumps `METHOD /path` lowercase and diffs literally. A .NET parameterized `@page "/admin/reference/{catalog}"` would dump as `/admin/reference/{catalog}`, which cannot match Django/Go concrete paths → parity DRIFT. Each stack registers the three literal paths below, identically.

3. **Reuse 2.3's read API and column sets verbatim — invent no markup or queries.** Each new page calls **one** existing read method and renders the **same columns** 2.3's overview already renders for that catalog (AC2 table). The `compliance_rule.parameters` JSONB renders via the same `<details><summary>parameters</summary><pre><code class="font-mono">…</code></pre>` disclosure 2.3 uses (compact JSON, no rule-kind-specific parsing).

4. **Empty-table graceful state.** Although the infra seed always loads rows, each page renders a single full-width "No <catalog> defined." row when the read returns zero rows (defensive; covers a fresh volume where seed has not run). This is the only behavior the three pages add beyond 2.3's section markup.

5. **No new cross-stack data contract; the invariant is the route set + page composition.** This story adds **no** JSON/wire/form contract. The cross-stack invariant is: the three literal route strings, the per-catalog column set + order, the admin-403 posture, and the Basecoat-Table shape — asserted by `make parity` + per-stack structure tests. No `docs/reference/` doc is required (composition of Story 2.3's existing mapping + read API + authz contracts).

---

## Acceptance Criteria

### AC1 — Three dedicated admin catalog pages render for ADMIN

**Given** I am authenticated as `ADMIN`
**When** I navigate to each route below
**Then** each stack returns **HTTP 200**, renders the Story 1.5 chrome, a single `<h1>` naming the catalog (one per page, UX-DR33), and a single Basecoat `<table class="table">` with one `<tbody>` row per seeded record, ordered by `code`:

| Route (literal, lowercase) | `<h1>` | Read method (Story 2.3) |
|---|---|---|
| `GET /admin/reference/trade-types` | `Trade Types` | `ListTradeTypes` / `list_trade_types()` / `ListTradeTypes` |
| `GET /admin/reference/violation-categories` | `Violation Categories` | `ListViolationCategories` / `list_violation_categories()` / `ListViolationCategories` |
| `GET /admin/reference/compliance-rules` | `Compliance Rules` | `ListComplianceRules` / `list_compliance_rules()` / `ListComplianceRules` |

**And** each page calls **only its one** read method (not all three) — the page is single-catalog.
**And** no Create / Edit / Delete affordances render anywhere on any of the three pages (FR67 is Growth).

### AC2 — Each catalog renders its canonical columns (reuse 2.3's column sets)

**Given** each new page
**When** the table renders
**Then** the column headers + cell values match Story 2.3's overview exactly for that catalog:

| Catalog | Columns (in order) |
|---|---|
| Trade Types | `Code`, `Name`, `Description`, `Active` |
| Violation Categories | `Code`, `Name`, `Trade Type ID`, `Default Severity`, `Description`, `Active` |
| Compliance Rules | `Code`, `Name`, `Description`, `Rule Kind`, `Parameters`, `Active` |

**And** `Description` renders an empty cell when the source value is `NULL` (`.NET` empty string / Django `default_if_none:""` / Go `optionalString` — the 2.3 helpers).
**And** the Compliance Rules `Parameters` cell renders the JSONB as a `<details><summary>parameters</summary><pre><code class="font-mono">{compact-json}</code></pre>` disclosure (identical to 2.3 — .NET `JsonSerializer.Serialize(RootElement)`, Django `json.dumps(…, separators=(",",":"))`, Go `string(rule.Parameters)`).
**And** `Active` renders as the boolean's text form (`True`/`False` per stack convention, matching 2.3) — color is never the sole carrier (cat 8).

### AC3 — Empty-table graceful state (edge-case cat 9)

**Given** a catalog read returns zero rows
**When** the page renders
**Then** the `<tbody>` contains a single row reading "No trade types defined." / "No violation categories defined." / "No compliance rules defined." spanning all columns (`colspan`), instead of an empty `<tbody>`.
**And** a per-stack test asserts this against an empty `domain.*` table (truncate-in-transaction or a no-seed fixture) — distinct from the seeded-rows path.

### AC4 — Authorization: non-admin → 403 on every new route (FR56)

**Given** Story 2.3's admin-gating pattern
**When** I inspect each stack
**Then** each of the three new routes enforces `ADMIN`-only access using the **same mechanism 2.3 uses**, with **no new authz primitive invented**:
- **.NET:** `[Authorize(Roles = "ADMIN")]` on each new PageModel (as on `ReferenceModel`).
- **Django:** the `_is_admin(request)` guard from [reference/views.py](../../fieldmark_py/reference/views.py) raising `PermissionDenied` (→ 403) at the top of each new view.
- **Go:** the `actor == nil || actor.Role != string(domain.RoleAdmin)` check from [admin_reference.go](../../fieldmark-go/internal/web/handlers/admin_reference.go) returning `fiber.StatusForbidden` + the canonical message string, in each new handler.

**And** an authenticated non-`ADMIN` user (`COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) requesting **any** of the three routes receives **HTTP 403** with **no reference-data leakage** in the body (no codes, no row data — same assertion as 2.3's `AdminReference_NonAdmin_Returns403WithoutReferenceState`).
**And** an unauthenticated request hits the Story 1.11 auth-required redirect first (→ `/login`), unchanged.
**And** a per-stack test exercises ADMIN → 200 and each non-admin role → 403 for all three routes.

### AC5 — Nav links absent for non-admins (FR6); admin sub-nav present

**Given** FR6 (no affordance leakage)
**When** a non-admin renders any page
**Then** **no** link to any `/admin/reference*` route appears in their rendered chrome (trivially satisfied — there are no `/admin/reference` links in the base layout/nav today; do **not** add any nav-bar links).

**Given** I am ADMIN on any of the three new pages
**When** the page renders
**Then** a small in-page admin sub-nav (e.g. a `<nav aria-label="Reference catalogs">` with links) offers: the two **sibling** catalog pages + a "← Reference overview" link to `/admin/reference`. These links live **only inside the three new admin pages** (which are themselves admin-gated), so they never render for non-admins. The sub-nav markup + link set is identical across stacks.

### AC6 — `make parity` and full gate

**Given** Story 1.3 route-parity tooling
**When** I run `make parity`
**Then** all three new routes (`get /admin/reference/trade-types`, `get /admin/reference/violation-categories`, `get /admin/reference/compliance-rules`) appear in **all three** stacks' route dumps with a **clean diff**; `get /admin/reference` (2.3) remains present; `pg_indexes` diff is **zero** (no schema change). `make test-all` exits 0.

**Critical per-stack registration notes (parity will fail otherwise):**
- **Go:** routes are registered inside the `if pool != nil { … } else { … }` block in [cmd/web/main.go:150–158](../../fieldmark-go/cmd/web/main.go). `go run ./cmd/web -dump-routes` runs **without** a pool → the `else` branch executes. Register the three new routes in **both** branches (real handler when `pool != nil`; stub `func(c) error { return nil }` in `else`), exactly as 2.3 does for `/admin/reference`.
- **Django:** the three new `path("admin/reference/…")` entries **must precede** `path("admin/", admin.site.urls)` in [fieldmark/urls.py](../../fieldmark_py/fieldmark/urls.py) (patterns resolve in declaration order — same constraint as the existing `admin/reference` line). Use no trailing slash, matching the existing `admin/reference` entry.
- **.NET:** three explicit `@page "/admin/reference/<catalog>"` directives (one per Razor Page) — never a route parameter (Decisions note 2).

**Build/type/lint/test gates green per stack:**
- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test` — clean. New: three `Pages/Admin/Reference*.cshtml(.cs)` pages + page tests.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest` — clean. New: three views + three templates + three url paths + tests under `reference/tests/`.
- **Go:** `cd fieldmark-go && make check && go test ./...` — clean. New: three handler methods + three templates + route registration + handler tests.
- **`fieldmark_shared`:** **no CSS change expected** (Basecoat `.table` + existing utilities already cover this). `dist/fieldmark.css` byte-identical. If a sub-nav utility is genuinely missing, prefer an existing Tailwind utility over a `src/` edit.

### AC7 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

**Given category 9 (empty / whitespace values)**
**Then** the empty-table state is per AC3; `NULL` `description` renders an empty cell per AC2 (the 2.3 helpers). No derived-token helper is introduced.

**Given category 6 (text overflow & special characters)**
**Then** all cell values are framework-escaped on render (Razor / Django autoescape / Go `html/template`) — reference data is operator-seeded, not user-entered, so a dedicated XSS round-trip test is **not** required (per [security-defaults.md §3a](../../docs/reference/security-defaults.md)); note this explicitly in dev notes. Long `description`/`name` values wrap within the Basecoat table cell (no fixed-width clipping introduced).

**Given category 8 (forced colors)**
**Then** `Active` and `Default Severity` are conveyed as **text**, never color-only. axe scan on each page reports zero new WCAG 2.1 AA violations.

**Given categories 1, 2, 3, 4, 5, 7**
**Then** N/A — no unknown-vocabulary component (plain table cells), no new fonts, no JS init (pages are static server-rendered markup — degrade perfectly with JS off), no AG Grid, no unbounded stacking, reduced-motion is the Story 1.14 global rule. A `javaScriptEnabled:false` assertion (the three pages fully render with JS off) is sufficient for cat 3.

### AC8 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

**Given the central concern is authorization** (cat — broken access control)
**Then** AC4 is the security AC: every route is admin-gated and 403s without leakage for all four non-admin roles.

**Given categories 1–7**
**Then** N/A — no open-redirect (no user-controlled redirect target), no new cookies, no user-controlled writes (read-only), no dynamic regex, no filesystem writes, no CSRF surface (`GET` safe methods, no forms), stub-auth warning is Story 1.9's. State this explicitly.

### AC9 — Cross-stack parity of the page composition (per [CLAUDE.md](../../CLAUDE.md) Cross-Stack Architecture Principle)

**Given** the Cross-Stack Architecture Principle
**Then** each stack's three pages are **native** (Razor Pages; Django views + templates; Go handlers + `html/template`) — no shared template fragment, no symlinked partial. The cross-stack invariant is the **route set (AC1) + column set/order (AC2) + admin-403 posture (AC4) + sub-nav link set (AC5)**, asserted by `make parity` + per-stack structure tests. This story introduces **no** new `docs/reference/` contract (composition of Story 2.3's existing contracts — Decisions note 5).

## Tasks / Subtasks

- [x] **Task 1: .NET — three Razor Pages** (AC: #1, #2, #3, #4, #5, #6)
  - [x] 1.1 Add `Pages/Admin/ReferenceTradeTypes.cshtml(.cs)` (`@page "/admin/reference/trade-types"`), `ReferenceViolationCategories.cshtml(.cs)` (`@page "/admin/reference/violation-categories"`), `ReferenceComplianceRules.cshtml(.cs)` (`@page "/admin/reference/compliance-rules"`). Each `[Authorize(Roles = "ADMIN")]`, each ctor-injects `IReferenceReader`, each `OnGetAsync` calls only its one list method. ComplianceRules projects to a `ParametersJson` row record (copy 2.3's `ComplianceRuleRow` shape).
  - [x] 1.2 Each `.cshtml` renders the single-catalog `<h1>` + Basecoat `<table class="table">` (reuse the matching `<section>` markup from [Reference.cshtml](../../FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml)) + the admin sub-nav (AC5) + the empty-table `colspan` row (AC3).
  - [x] 1.3 Tests in `FieldMark.Tests.Web/Pages/` (mirror `AdminReferencePageTests`): per route, ADMIN (`aisha`) → 200 + seeded rows (assert `ELEC` / a category code / `OPEN_VIOLATION_GATE` as appropriate) + correct column headers; each non-admin (`marisol`/`ravi`/`pat`/`kenji`) → 403 without leakage; empty-DB → empty-state row.

- [x] **Task 2: Django — three views + templates + urls** (AC: #1, #2, #3, #4, #5, #6)
  - [x] 2.1 Add three views in [reference/views.py](../../fieldmark_py/reference/views.py) (e.g. `trade_types`, `violation_categories`, `compliance_rules`), each guarded by the existing `_is_admin` → `PermissionDenied`, each calling only its one `queries.list_*()` function; compliance-rules view reuses the `ComplianceRuleRow`/`json.dumps` projection.
  - [x] 2.2 Add three templates under `reference/templates/reference/` (extract the matching `<section>` from [index.html](../../fieldmark_py/reference/templates/reference/index.html) into a single-catalog page) + the admin sub-nav + empty-state (`{% empty %}` clause in the `{% for %}`).
  - [x] 2.3 Add three `path("admin/reference/<catalog>", view, name=…)` entries in [fieldmark/urls.py](../../fieldmark_py/fieldmark/urls.py) **before** the `admin/` mount, after the existing `admin/reference` line; no trailing slash.
  - [x] 2.4 Tests under `reference/tests/`: ADMIN → 200 + rows; each non-admin role → 403; empty-table state.

- [x] **Task 3: Go — three handlers + templates + routes** (AC: #1, #2, #3, #4, #5, #6)
  - [x] 3.1 Add three handler methods on `AdminReferenceHandlers` in [admin_reference.go](../../fieldmark-go/internal/web/handlers/admin_reference.go) (e.g. `TradeTypesIndex`, `ViolationCategoriesIndex`, `ComplianceRulesIndex`), each with the existing admin-role guard, each calling only its one `Reference.List*` method and reusing the existing row-projection helpers (`tradeTypeRows`, `violationCategoryRows`, the inline rule-row loop).
  - [x] 3.2 Add three templates under `internal/web/templates/pages/` (single-catalog table) + admin sub-nav + empty-state (`{{if .Rows}}…{{else}}…{{end}}`).
  - [x] 3.3 Register the three routes in [cmd/web/main.go](../../fieldmark-go/cmd/web/main.go) **inside both** the `if pool != nil` (real handler) **and** the `else` (stub) branches alongside the existing `/admin/reference` registration — required for `-dump-routes` parity (AC6).
  - [x] 3.4 Handler tests in `internal/web/handlers/`: ADMIN actor → 200 + rows; non-admin actor → 403; empty store → empty-state.

- [x] **Task 4: Parity + gate** (AC: #6, #9)
  - [x] 4.1 `make parity` — assert the three new routes present + clean diff across stacks; `pg_indexes` zero-diff.
  - [x] 4.2 `make test-all` green; per-stack build/lint/type gates clean.

- [x] **Task 5: Story sign-off** (AC: all)
  - [x] 5.1 Populate the Sign-off block; record the five decisions; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **This is a pure read — no transaction, no audit.** Authorize → call the one read method → render. Do **not** open a transaction and write **no** `AuditEntry` (reads are never audited). Same discipline as Story 2.3 and the dashboard (2.10).

- **Reuse Story 2.3's read API — do not add a data layer.** The mappings (`TradeType`/`ViolationCategory`/`ComplianceRule`), the read surfaces (`IReferenceReader` / `reference/queries.py` / `ReferenceStore`), and the JSONB projection helpers all exist and are `done`. Each new page is a thin presentation layer over one existing read method. If you find yourself writing a query or a model, stop — it already exists.

- **Reuse the existing column markup.** The exact `<table class="table">` blocks already live in the 2.3 overview ([Reference.cshtml](../../FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml), [index.html](../../fieldmark_py/reference/templates/reference/index.html), [admin_reference.html](../../fieldmark-go/internal/web/templates/pages/admin_reference.html)). Lift the relevant `<section>`'s table into the single-catalog page; keep headers/cells identical so AC2 holds and the two views stay consistent.

- **Admin gating is copy-the-pattern, not invent-a-primitive.** .NET `[Authorize(Roles="ADMIN")]`; Django `_is_admin`→`PermissionDenied`; Go `actor.Role != string(domain.RoleAdmin)`→`fiber.StatusForbidden`. Use the **same 403 message string** Go already returns (`"You do not have permission to access this page."`) for consistency.

- **Parity is the #1 trap for this story (route registration).** Three failure modes, all caught by `make parity`:
  1. **.NET parameterized route** → dumps `/admin/reference/{catalog}`, drifts from concrete Django/Go paths. Use three literal `@page` routes.
  2. **Go `else`-branch omission** → `-dump-routes` runs pool-less, so a route registered only in `if pool != nil` is invisible to the dump → DRIFT. Register in both branches.
  3. **Django ordering / trailing slash** → place before `admin/`, match the no-trailing-slash convention of the existing `admin/reference` line.
  Route strings are lowercase kebab-case **plural**: `/admin/reference/trade-types`, `/admin/reference/violation-categories`, `/admin/reference/compliance-rules`.

- **Do not touch the 2.3 overview page or its tests (Decisions note 1).** `/admin/reference` and `AdminReferencePageTests` (+ Django/Go equivalents) stay green. This story is additive.

- **Reference data is operator-seeded, not user input.** Framework autoescaping is sufficient; no XSS round-trip test is required (per [security-defaults.md §3a](../../docs/reference/security-defaults.md)). The one real security control is authz (AC4).

### Canonical columns (from DDL — [010_domain_tables.sql:24–52](../../docker/postgres/init/010_domain_tables.sql))

- `domain.trade_type`: `id, code (≤32, unique), name (≤120), description (TEXT, null), active (bool, default true)` → display `Code, Name, Description, Active`.
- `domain.violation_category`: `id, code (≤32, unique), name (≤200), trade_type_id (UUID, null, FK), default_severity (∈ Low/Medium/High/Critical), description (TEXT, null), active` → display `Code, Name, Trade Type ID, Default Severity, Description, Active`.
- `domain.compliance_rule`: `id, code (≤64, unique), name (≤200), description (TEXT, NOT NULL), rule_kind (∈ ScoringPenalty/ClosureGate), parameters (JSONB, NOT NULL), active` → display `Code, Name, Description, Rule Kind, Parameters (disclosure), Active`.

### Source tree — where things land

| Stack | New pages | Route registration | Reused read API |
|---|---|---|---|
| .NET | `FieldMark.Web/Pages/Admin/ReferenceTradeTypes.cshtml(.cs)`, `ReferenceViolationCategories.cshtml(.cs)`, `ReferenceComplianceRules.cshtml(.cs)` | `@page` directive in each `.cshtml` | `FieldMark.Data/Reference/IReferenceReader` |
| Django | three templates in `reference/templates/reference/` + three views in `reference/views.py` | `fieldmark/urls.py` (before `admin/`) | `reference/queries.py` |
| Go | three templates in `internal/web/templates/pages/` + three methods in `internal/web/handlers/admin_reference.go` | `cmd/web/main.go` (both `if pool`/`else` branches) | `internal/data/postgres/referencestore.go` (`ReferenceStore`) |

No shared asset changes. No `docs/reference/` contract change.

### Existing code to reuse (read before writing)

- **2.3 overview page + tables** — [Reference.cshtml(.cs)](../../FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml), [reference/views.py](../../fieldmark_py/reference/views.py) + [index.html](../../fieldmark_py/reference/templates/reference/index.html), [admin_reference.go](../../fieldmark-go/internal/web/handlers/admin_reference.go) + [admin_reference.html](../../fieldmark-go/internal/web/templates/pages/admin_reference.html).
- **Read APIs** — [IReferenceReader.cs](../../FieldMark/FieldMark.Data/Reference/IReferenceReader.cs), [reference/queries.py](../../fieldmark_py/reference/queries.py), `ReferenceStore` in [referencestore.go](../../fieldmark-go/internal/data/postgres/referencestore.go).
- **Authz patterns + 403 test** — 2.3's `[Authorize(Roles="ADMIN")]`, `_is_admin`, Go role guard; [AdminReferencePageTests.cs](../../FieldMark/FieldMark.Tests.Web/Pages/AdminReferencePageTests.cs) for the 200/403 test shape.
- **Dev-user fixtures** — Story 1.10 manifest: `aisha` (ADMIN), `marisol`/`ravi`/`pat`/`kenji` (non-admin roles).

### Project Structure Notes

- Adds three `GET /admin/reference/*` routes to the parity inventory; `GET /admin/reference` (2.3) stays.
- No `domain.*` schema change; `pg_indexes` zero-diff.
- No shared JS/CSS change expected; no new vendor asset.
- No new cross-stack `docs/reference/` contract (composition of Story 2.3's contracts).

### References

- Epic AC: [epic-2 §Story 2.14](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md)
- DDL: [010_domain_tables.sql:24–52](../../docker/postgres/init/010_domain_tables.sql) (trade_type / violation_category / compliance_rule)
- Read-handler shape (no tx, no audit): [architecture.md](../planning-artifacts/architecture.md) §read-handler
- Parity tooling: [tools/parity/diff-routes.sh](../../tools/parity/diff-routes.sh), [dump-routes-fiber.sh](../../tools/parity/dump-routes-fiber.sh)
- Edge cases / security: [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) cat 9/6/8, [security-defaults.md](../../docs/reference/security-defaults.md) (authz; §3a operator-seeded data)
- Prior story: [2-3-…md](2-3-map-reference-data-tables-and-expose-a-read-api-per-stack.md) (mappings, read API, overview page, admin authz)

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- 2026-06-01: `make test-go` passed, including new `admin_reference_catalogs_test.go`.
- 2026-06-01: `cd fieldmark_py && uv run pytest reference/tests/test_catalog_views.py reference/tests/test_catalog_authz.py` passed (18 tests).
- 2026-06-01: `make test-net` failed due pre-existing unrelated .NET compile/analyzer issues in existing files (`_ProjectCreateForm.cshtml`, `_TabStrip.cshtml`, `Projects/*.cshtml.cs`, `Grid/Projects.cshtml.cs`).
- 2026-06-01: `make test-django` failed on pre-existing unrelated integration/dashboard failures outside reference catalog changes.
- 2026-06-01: `make parity` failed because .NET route dump cannot build due the same pre-existing .NET compile issues.
- 2026-06-01: Review follow-ups P1/P2/P3 implemented.
- 2026-06-01: `make parity` now passes with route + index parity clean.
- 2026-06-01: `make test-net` now compiles and runs, but still has pre-existing failing web-page tests outside Story 2.14 scope (`ProjectsListPageTests`, `ProjectsCreatePageTests`, `HomePageTests`).
- 2026-06-01: `make test-django` still has pre-existing integration/dashboard failures outside Story 2.14 scope (`projects/tests/test_mapping.py`, `fieldmark/tests/test_dashboard_page.py`).
- 2026-06-01: `make test-all` passed end-to-end (.NET, Django, Go unit + integration).
- 2026-06-01: Round 2 patch checks passed (`uv run pytest reference/tests/test_catalog_views.py` and `go test ./internal/web/handlers -run TestAdminReferenceCatalogsAdminRendersPages`).

### Completion Notes List

- Implemented three dedicated admin catalog pages per stack:
  - .NET Razor pages/routes: `/admin/reference/trade-types`, `/admin/reference/violation-categories`, `/admin/reference/compliance-rules`.
  - Django views/templates/routes for the same three paths.
  - Go handlers/templates/route registration for the same three paths (both pool and no-pool route registration branches).
- Added admin-only sub-nav on each dedicated catalog page linking to overview + sibling pages.
- Added empty-state row behavior on each dedicated table (`No ... defined.` with proper `colspan`).
- Added/updated tests for admin success, non-admin 403 no-leakage, and empty-state behavior in all three stacks.
- Addressed review action items:
  - P1 fixed: Django catalog pages now render page-specific 3-link sub-navs (overview + siblings, no self-link).
  - P2 fixed: .NET empty-state test now uses DI override (`EmptyReferenceReader`) instead of mutating shared seeded tables.
  - P3 fixed: Django violation category trade type cell now applies `default_if_none`.
- P4 fixed: full AC6 gate is now green (`make parity` and `make test-all`).

### File List

- _bmad-output/implementation-artifacts/sprint-status.yaml
- FieldMark/FieldMark.Web/Pages/Admin/ReferenceTradeTypes.cshtml
- FieldMark/FieldMark.Web/Pages/Admin/ReferenceTradeTypes.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Admin/ReferenceViolationCategories.cshtml
- FieldMark/FieldMark.Web/Pages/Admin/ReferenceViolationCategories.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Admin/ReferenceComplianceRules.cshtml
- FieldMark/FieldMark.Web/Pages/Admin/ReferenceComplianceRules.cshtml.cs
- FieldMark/FieldMark.Tests.Web/Pages/AdminReferenceCatalogPagesTests.cs
- fieldmark_py/reference/views.py
- fieldmark_py/fieldmark/urls.py
- fieldmark_py/reference/templates/reference/_subnav.html
- fieldmark_py/reference/templates/reference/trade_types.html
- fieldmark_py/reference/templates/reference/violation_categories.html
- fieldmark_py/reference/templates/reference/compliance_rules.html
- fieldmark_py/tools/management/commands/dump_routes.py
- tools/parity/dump-routes-net.sh
- FieldMark/FieldMark.Web/Pages/Projects/Shared/_ProjectCreateForm.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/Create.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Projects/Index.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Grid/Projects.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml
- FieldMark/FieldMark.Tests.Web/Pages/ProjectsCreatePageTests.cs
- FieldMark/FieldMark.Tests.Web/Pages/ProjectsListPageTests.cs
- FieldMark/FieldMark.Tests.Web/Components/TabStripSnapshotTests.cs
- fieldmark_py/reference/tests/test_catalog_views.py
- fieldmark_py/reference/tests/test_catalog_authz.py
- fieldmark-go/internal/web/handlers/admin_reference.go
- fieldmark-go/internal/web/templates/pages/admin_reference_trade_types.html
- fieldmark-go/internal/web/templates/pages/admin_reference_violation_categories.html
- fieldmark-go/internal/web/templates/pages/admin_reference_compliance_rules.html
- fieldmark-go/internal/web/handlers/admin_reference_catalogs_test.go
- fieldmark-go/cmd/web/main.go
### Review Findings

- [x] [Review][Patch] P1 — Sub-nav self-link asymmetry: Django `_subnav.html` shows 4 links on every page (includes a link to the current page); .NET and Go correctly show 3 links (2 siblings + overview, exclude self). Spec AC5 requires "link set identical across stacks." Fix: remove the self-link for the current page from Django, either via page-specific sub-navs (3 links each) or a conditional include using a `current_page` context variable. [`fieldmark_py/reference/templates/reference/_subnav.html`]
- [x] [Review][Patch] P2 — .NET empty-state test TRUNCATEs shared seed tables (`domain.compliance_rule`, `domain.violation_category`, `domain.trade_type`) without re-seeding. xUnit does not guarantee method order within a class; if `EmptyState` runs before `Admin_RendersExpectedPage`, the seeded rows (`ELEC`, `OPEN_VIOLATION_GATE`) are gone and the admin-renders test fails. Fix: wrap the truncate in a transaction that rolls back after the assertion, or inject an empty DI override instead of mutating shared state. [`FieldMark/FieldMark.Tests.Web/Pages/AdminReferenceCatalogPagesTests.cs:84-107`]
- [x] [Review][Patch] P3 — Django `violation_categories.html` renders `{{ category.trade_type }}` without a null guard. The `trade_type_id` FK is nullable in the schema (`UUID NULL`); a category with no trade type renders literal "None" instead of an empty cell. Fix: `{{ category.trade_type|default_if_none:"" }}`. Violates AC2. [`fieldmark_py/reference/templates/reference/violation_categories.html:23`]
- [x] [Review][Patch] P4 — AC6 gate not green: `make test-all` and `make parity` both fail due to pre-existing .NET compile issues in files outside this story (`_ProjectCreateForm.cshtml`, `_TabStrip.cshtml`, `Projects/*.cshtml.cs`, `Grid/Projects.cshtml.cs`). AC6 is unconditional — the gate must be green before the story ships. Action: investigate and resolve the pre-existing .NET compile errors so the gate passes.
#### Round 2 (re-run after P1–P4 resolved)

- [x] [Review][Patch] R1 — Django `compliance_rules.html` renders `{{ rule.description }}` without `|default_if_none:""`. The schema has `description TEXT NOT NULL` but AC2 requires the defensive null-guard (same pattern applied in P3 to `violation_categories.html`). Also: `ComplianceRuleRow.description: str` type annotation should be `str | None` or the view should coerce `None → ""` before building the dataclass. Violates AC2. [`fieldmark_py/reference/templates/reference/compliance_rules.html:26`]
- [x] [Review][Patch] R2 — Django `test_catalog_views.py` and Go `admin_reference_catalogs_test.go` do not assert sub-nav self-link absence per path. The .NET test (`AdminReferenceCatalogPagesTests.cs:49–66`) does per-path `NotContain` assertions. A future regression re-introducing the self-link would pass Django and Go tests undetected. Violates AC5/AC9 cross-stack test parity. Fix: add `assert "/admin/reference/trade-types" not in html` (and sibling paths) to the Django parameterized test; add equivalent `strings.Contains`-negation checks to Go. [`fieldmark_py/reference/tests/test_catalog_views.py`] [`fieldmark-go/internal/web/handlers/admin_reference_catalogs_test.go`]
- [x] [Review][Defer] D-R1 — `_subnav.html` is dead code (no template includes it after the P1 inline-nav fix). The file contains the 4-link self-referential sub-nav — accidentally `{% include %}`ing it in a future template would silently reintroduce AC5 bug on all 3 pages. Delete before merge. — deferred [`fieldmark_py/reference/templates/reference/_subnav.html`]
- [x] [Review][Defer] D-R2 — .NET `NotContain("/admin/reference/trade-types")` assertion checks the full rendered body text. A tighter assertion `html.Should().NotContain("href=\"/admin/reference/trade-types\"")` would scope the check to the nav href only and not be at risk of false-failing if the URL appears in a meta tag or breadcrumb in a future layout change. — deferred, test quality [`FieldMark/FieldMark.Tests.Web/Pages/AdminReferenceCatalogPagesTests.cs:51`]
- [x] [Review][Defer] D-R3 — Django `index.html` (Story 2.3 overview) `{{ category.trade_type }}` has no `|default_if_none:""` guard (same gap as P3). Out of scope — spec Decisions note 1 explicitly prohibits touching the 2.3 overview page. Track as pre-existing. — deferred, pre-existing, out of story scope [`fieldmark_py/reference/templates/reference/index.html`]

#### Round 1

- [x] [Review][Patch] P1 — Sub-nav self-link asymmetry: Django `_subnav.html` shows 4 links on every page (includes a link to the current page); .NET and Go correctly show 3 links (2 siblings + overview, exclude self). Spec AC5 requires "link set identical across stacks." Fix: Django sub-nav should show only the 2 sibling links + overview. [`fieldmark_py/reference/templates/reference/_subnav.html`] ✅ resolved
- [x] [Review][Patch] P2 — .NET empty-state test TRUNCATEs shared seed tables (`domain.compliance_rule`, `domain.violation_category`, `domain.trade_type`) without re-seeding. Fix: use `WithWebHostBuilder` + `EmptyReferenceReader` DI override. [`FieldMark/FieldMark.Tests.Web/Pages/AdminReferenceCatalogPagesTests.cs:84-107`] ✅ resolved
- [x] [Review][Patch] P3 — Django `violation_categories.html` renders `{{ category.trade_type }}` without a null guard. The FK `trade_type_id` is nullable. Fix: `{{ category.trade_type|default_if_none:"" }}`. Violates AC2. [`fieldmark_py/reference/templates/reference/violation_categories.html:23`] ✅ resolved
- [x] [Review][Patch] P4 — AC6 gate not green: `make test-all` and `make parity` both failing. ✅ resolved
- [x] [Review][Defer] D1 — Go call-counter assertion (`store.tradeCalls != 1 || store.categoryCalls != 1 || store.ruleCalls != 1`) is cumulative across all 3 sub-tests and placed outside `t.Run` scope. The assertion is correct for sequential execution and the handlers are obviously single-purpose, but it does not strictly isolate per-handler method calls and would be a data race if sub-tests were ever parallelized. — deferred, pre-existing [`fieldmark-go/internal/web/handlers/admin_reference_catalogs_test.go:114`]
- [x] [Review][Defer] D2 — `actor == nil` guard in Go handlers is unreachable dead code: `auth.ActorFromCtx` always returns a non-nil `*app.Actor` (falls back to `app.Anonymous()`). The route-level `RequireAuth()` middleware redirects anonymous users before the handler runs. Defensive but misleading to future maintainers. — deferred, pre-existing [`fieldmark-go/internal/web/handlers/admin_reference.go:92-94` and equivalent]
- [x] [Review][Defer] D3 — AC7 cat 3 Playwright `javaScriptEnabled:false` assertion not implemented. Pages are purely server-rendered and degrade correctly with JS off; this is a test coverage gap only, not a production defect. — deferred, pre-existing

## Sign-off

- Date of final review: 2026-06-01
- Total review-round count: 1
- Final reviewer verdict (PASS/FAIL): PASS (ready for code review)
- Deferred-work entries created from this story:
- Decisions requiring ratification (recorded here; confirm or overturn at review):
  1. **Additive** — the 2.3 `/admin/reference` overview page + its tests are left byte-identical; the three new routes are purely additive (no repurpose).
  2. **Three explicit concrete routes** per stack (never a parameterized .NET route) for clean `make parity`.
  3. **Reuse 2.3's read API + column sets + JSONB disclosure verbatim** — no new data layer or markup invented.
  4. **Empty-table graceful `colspan` state** added per page (defensive; the only behavior beyond 2.3's section markup).
  5. **No new cross-stack `docs/reference/` contract** — invariant is the route set + column set + admin-403 posture + sub-nav, asserted by parity + per-stack tests.
