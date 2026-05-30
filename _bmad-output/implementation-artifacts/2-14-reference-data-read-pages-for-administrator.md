# Story 2.14: Reference data read pages for Administrator

Status: ready-for-dev

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

- [ ] **Task 1: .NET — three Razor Pages** (AC: #1, #2, #3, #4, #5, #6)
  - [ ] 1.1 Add `Pages/Admin/ReferenceTradeTypes.cshtml(.cs)` (`@page "/admin/reference/trade-types"`), `ReferenceViolationCategories.cshtml(.cs)` (`@page "/admin/reference/violation-categories"`), `ReferenceComplianceRules.cshtml(.cs)` (`@page "/admin/reference/compliance-rules"`). Each `[Authorize(Roles = "ADMIN")]`, each ctor-injects `IReferenceReader`, each `OnGetAsync` calls only its one list method. ComplianceRules projects to a `ParametersJson` row record (copy 2.3's `ComplianceRuleRow` shape).
  - [ ] 1.2 Each `.cshtml` renders the single-catalog `<h1>` + Basecoat `<table class="table">` (reuse the matching `<section>` markup from [Reference.cshtml](../../FieldMark/FieldMark.Web/Pages/Admin/Reference.cshtml)) + the admin sub-nav (AC5) + the empty-table `colspan` row (AC3).
  - [ ] 1.3 Tests in `FieldMark.Tests.Web/Pages/` (mirror `AdminReferencePageTests`): per route, ADMIN (`aisha`) → 200 + seeded rows (assert `ELEC` / a category code / `OPEN_VIOLATION_GATE` as appropriate) + correct column headers; each non-admin (`marisol`/`ravi`/`pat`/`kenji`) → 403 without leakage; empty-DB → empty-state row.

- [ ] **Task 2: Django — three views + templates + urls** (AC: #1, #2, #3, #4, #5, #6)
  - [ ] 2.1 Add three views in [reference/views.py](../../fieldmark_py/reference/views.py) (e.g. `trade_types`, `violation_categories`, `compliance_rules`), each guarded by the existing `_is_admin` → `PermissionDenied`, each calling only its one `queries.list_*()` function; compliance-rules view reuses the `ComplianceRuleRow`/`json.dumps` projection.
  - [ ] 2.2 Add three templates under `reference/templates/reference/` (extract the matching `<section>` from [index.html](../../fieldmark_py/reference/templates/reference/index.html) into a single-catalog page) + the admin sub-nav + empty-state (`{% empty %}` clause in the `{% for %}`).
  - [ ] 2.3 Add three `path("admin/reference/<catalog>", view, name=…)` entries in [fieldmark/urls.py](../../fieldmark_py/fieldmark/urls.py) **before** the `admin/` mount, after the existing `admin/reference` line; no trailing slash.
  - [ ] 2.4 Tests under `reference/tests/`: ADMIN → 200 + rows; each non-admin role → 403; empty-table state.

- [ ] **Task 3: Go — three handlers + templates + routes** (AC: #1, #2, #3, #4, #5, #6)
  - [ ] 3.1 Add three handler methods on `AdminReferenceHandlers` in [admin_reference.go](../../fieldmark-go/internal/web/handlers/admin_reference.go) (e.g. `TradeTypesIndex`, `ViolationCategoriesIndex`, `ComplianceRulesIndex`), each with the existing admin-role guard, each calling only its one `Reference.List*` method and reusing the existing row-projection helpers (`tradeTypeRows`, `violationCategoryRows`, the inline rule-row loop).
  - [ ] 3.2 Add three templates under `internal/web/templates/pages/` (single-catalog table) + admin sub-nav + empty-state (`{{if .Rows}}…{{else}}…{{end}}`).
  - [ ] 3.3 Register the three routes in [cmd/web/main.go](../../fieldmark-go/cmd/web/main.go) **inside both** the `if pool != nil` (real handler) **and** the `else` (stub) branches alongside the existing `/admin/reference` registration — required for `-dump-routes` parity (AC6).
  - [ ] 3.4 Handler tests in `internal/web/handlers/`: ADMIN actor → 200 + rows; non-admin actor → 403; empty store → empty-state.

- [ ] **Task 4: Parity + gate** (AC: #6, #9)
  - [ ] 4.1 `make parity` — assert the three new routes present + clean diff across stacks; `pg_indexes` zero-diff.
  - [ ] 4.2 `make test-all` green; per-stack build/lint/type gates clean.

- [ ] **Task 5: Story sign-off** (AC: all)
  - [ ] 5.1 Populate the Sign-off block; record the five decisions; flip sprint-status to `review`.

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

### Debug Log References

### Completion Notes List

### File List

## Sign-off

- Date of final review:
- Total review-round count:
- Final reviewer verdict (PASS/FAIL):
- Deferred-work entries created from this story:
- Decisions requiring ratification (recorded here; confirm or overturn at review):
  1. **Additive** — the 2.3 `/admin/reference` overview page + its tests are left byte-identical; the three new routes are purely additive (no repurpose).
  2. **Three explicit concrete routes** per stack (never a parameterized .NET route) for clean `make parity`.
  3. **Reuse 2.3's read API + column sets + JSONB disclosure verbatim** — no new data layer or markup invented.
  4. **Empty-table graceful `colspan` state** added per page (defensive; the only behavior beyond 2.3's section markup).
  5. **No new cross-stack `docs/reference/` contract** — invariant is the route set + column set + admin-403 posture + sub-nav, asserted by parity + per-stack tests.
