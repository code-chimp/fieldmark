# Story 2.10: Compliance Dashboard with portfolio tiles

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.10
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — `domain.project` (58–95), `domain.inspection` (101–122), `domain.violation` (139–167)
Depends on:
- **Story 2.4** (DashboardTile component — `_DashboardTile.cshtml` / `_dashboard_tile.html` / `dashboard_tile.html`; props `tile_id, label, value, secondary, value_color, role_status`; empty value → `—`; status: **done**).
- **Story 2.5** (ComplianceTile component + the `compliance-tile-portfolio` variant — `_ComplianceTile.cshtml` / `_compliance_tile.html` / `compliance_tile.html`; props `score (int?, null→em-dash), label, id`; threshold bands baked in; status: **done**).
- **Story 2.9** (Projects AG Grid + **AGGridPanel** wrapper + `POST /grid/projects` + `#project-detail` convention; status: **ready-for-dev**). This story **reuses the AGGridPanel** below the tile row and **extends it** with a row-click `navigate` mode (see AC4 / Decisions note 4). If 2.9 is not yet `done` at implementation time, this story must coordinate so the shared AGGridPanel JS exists before the dashboard can embed it.
- **Story 1.12** (`can(actor, action)` primitive + `RegisterAction` — registers the new `dashboard.view` action).
- **Story 1.13** (empty role-aware Home at `GET /` — this story changes Home to **redirect to `/dashboard`**; the three stack `CLAUDE.md` "Home page" sections currently say "Story 2.10 replaces it" and must be updated).
- **Story 1.5** (base-layout chrome — header, `#flash-region`, footer; the dashboard renders inside it).
- **Story 2.1** (`domain.project` mapping — the project-based aggregates read through it; status: **done**).

## Story

As any authenticated user landing on FieldMark,
I want a Compliance Dashboard at `/dashboard` showing four portfolio aggregate tiles (portfolio compliance score, overdue violations by severity, active projects, inspections this week) above the Projects AG Grid, with `GET /` redirecting here,
So that I can answer "where is risk?" at a glance and drill into a Project from the grid (FR44, FR46) — and the OOB-swappable tile pattern (UX-DR45), the responsive tile-row reflow (UX-DR30), and the empty-vs-zero rendering rule (UX-DR17) are locked in for every downstream story that updates a tile from a same-page action (2.12 place-on-hold, Epic 5 anchor demo).

**Scope boundary.** This story produces, per stack:
- (a) `GET /dashboard` route + handler + page rendering: the **header tile row** (ComplianceTile `#compliance-tile-portfolio` + three DashboardTiles) and, below it, the **AGGridPanel** (Story 2.9) filling the remaining viewport;
- (b) a thin **read-only "dashboard stats" query** per stack computing the four aggregate values (ORM on `domain.project`; **direct read-only SQL** on `domain.violation` / `domain.inspection` — those aggregates are not entity-mapped until Epic 3, but the tables exist in the canonical DDL — see Decisions note 1);
- (c) the `GET /` → `GET /dashboard` **redirect** (302) wired into each stack's existing Home handler;
- (d) the `dashboard.view` permission action registered (all five conceptual roles) and both routes gated by it;
- (e) the responsive tile-row CSS (4-up ≥1280 / 2×2 768–1279 / 1-col <768, UX-DR30) using Tailwind grid utilities;
- (f) per-stack tests: aggregate-value correctness (incl. empty-vs-zero per tile), tile-id + ARIA structure, responsive markup, authz 403/redirect, Home-redirect, `make parity`.

**Out of scope:**
- **OOB-swap wiring of the tiles** — this story renders each tile with a **stable `id` and `role="status"`** so a later story can OOB-replace it, but wires **no** live change events (UX-DR45 is not violated because the tiles render correctly at every static state). The producers land in their own stories (2.12 place-on-hold re-renders `#compliance-tile`; Epic 5 anchor demo).
- **The Project Detail screen** the grid drills into (`GET /projects/<id>` — Story 2.11 / the 2.8 stub). The dashboard's grid row-click navigates there; this story does not build that screen.
- **Editing/refreshing tiles via a button, auto-refresh polling, or websockets** — tiles are server-rendered on page load only.
- **A new `domain.*` index** for the "inspections this week" `scheduled_for` filter — at MVP scale the table is empty/small; adding an index is infra-owned and breaks `pg_indexes` parity (see Dev Notes §"No new indexes").
- **Severity badge *chips*** in the Overdue Violations tile — DashboardTile's `secondary` is a single text line; the severity breakdown is rendered as text (see Decisions note 3).
- Inspection/Violation **entity mappings** (Epic 3) — the dashboard reads counts via direct SQL, it does not introduce ORM models for those aggregates.

---

## ⚠️ Decisions baked into this story (read first)

These resolve ordering/contract ambiguities; each is implemented as written and flagged in the Sign-off block for reviewer ratification.

1. **Dashboard aggregates read via direct read-only SQL; no Epic-3 entity mapping pulled forward.** The Overdue-Violations and Inspections-This-Week tiles need counts from `domain.violation` / `domain.inspection`. Those aggregates are not entity-mapped until Epic 3 (Stories 3.1, 3.2), but the **tables exist** in the canonical DDL. The dashboard is a read-only aggregate view, so each stack issues a small **direct `COUNT`/`AVG` query** (ORM for `domain.project`; raw SQL for `violation`/`inspection`) — **no** new `Inspection`/`Violation` ORM models. When Epic 3 seeds data, the tiles light up automatically. (At this story's point in time the violation/inspection tables are empty, so those two tiles render `—`.)

2. **Portfolio compliance score = `ROUND(AVG(compliance_score))` over non-`Closed` projects.** The "portfolio average" is the live portfolio: `WHERE status <> 'Closed'`. `AVG` of an empty set is `NULL` → the ComplianceTile renders its no-data em-dash (`score=null`). Rounded to the nearest integer (ComplianceTile takes `int?`).

3. **Overdue-Violations severity breakdown is `secondary` text, not badge chips.** The epic says "broken down by severity sub-badges". DashboardTile's `secondary` slot is a single text line; rendering real StatusBadge chips would break its byte-snapshot contract and require a 2.4 component extension. MVP renders the breakdown as text in `secondary` — e.g. `"2 Critical, 1 High"` (only non-zero severities, ordered Critical → High → Medium → Low). Color is **not** the sole carrier (text labels present — satisfies edge-case cat 8). Chips are a possible later enhancement (extend `dashboard_tile` canonical), out of scope here.

4. **The dashboard grid row-click navigates to `/projects/<id>` (full page).** Story 2.9's AGGridPanel row-click does `htmx.ajax` into `#project-detail` (rail). The dashboard has **no** rail (the grid fills the viewport), and "drill into a Project" (FR46) means going to the dedicated Project Detail screen. This story **extends the shared AGGridPanel JS** with a `data-grid-rowclick="navigate"` mode (`window.location = "/projects/" + id`); 2.9's default `detail` mode is unchanged. The dashboard uses `navigate`.

5. **`dashboard.view` granted to all five roles; `GET /` redirects to `/dashboard`.** "Any authorized user landing on FieldMark" (Executive read-only per FR43). Same all-roles posture as Story 2.9's `project.read`. `GET /` returns **302** to `/dashboard` for authenticated users (unauthenticated still hits the Story 1.11 login redirect first).

---

## Acceptance Criteria

### AC1 — `GET /dashboard` renders the four-tile header row

**Given** I am authenticated with `dashboard.view`
**When** I navigate to `GET /dashboard`
**Then** each stack renders a full page (Story 1.5 chrome) with `<h1>Compliance Dashboard</h1>` (one per page, UX-DR33) and a tile row containing **exactly these four tiles in this order**, each with the canonical stable `id`:

| # | Component | `id` | Label | Value source |
|---|---|---|---|---|
| 1 | **ComplianceTile** | `compliance-tile-portfolio` | `Portfolio Compliance` | `ROUND(AVG(compliance_score))` over `domain.project WHERE status <> 'Closed'`; `NULL` → no-data em-dash |
| 2 | **DashboardTile** | `overdue-violations-tile` | `Overdue Violations` | count of `domain.violation WHERE status IN ('Open','InProgress') AND due_at < now()`; `secondary` = severity breakdown text |
| 3 | **DashboardTile** | `active-projects-tile` | `Active Projects` | count of `domain.project WHERE status = 'Active'` |
| 4 | **DashboardTile** | `inspections-week-tile` | `Inspections This Week` | count of `domain.inspection WHERE scheduled_for ∈ [start_of_iso_week_utc, next_week_utc)` |

**And** tiles 2–4 are rendered via the **DashboardTile wrapper** (Story 2.4) and tile 1 via the **ComplianceTile wrapper** (Story 2.5) — **no new tile markup is invented** (reuse the existing components and their canonical class vocabulary).

**And** all four tiles carry `role="status"` (ComplianceTile sets it intrinsically; pass `role_status=true` to each DashboardTile) so a later story can OOB-replace them. The tiles emit **no** HTMX producer attributes this story (no `hx-*`).

### AC2 — Aggregate values are correct, with the empty-vs-zero distinction (UX-DR17)

**Given** the dashboard stats read
**When** the four values are computed
**Then** each tile distinguishes **empty (no data → `—`)** from **zero (data exists, count is 0 → `0`)**:

| Tile | Renders `—` when | Renders a number when |
|---|---|---|
| Portfolio Compliance | no non-`Closed` projects exist (`AVG` is `NULL`) | the rounded average (e.g. `73`) |
| Overdue Violations | **no** violations exist at all (`SELECT COUNT(*) FROM domain.violation = 0`) | the overdue count (incl. `0` if violations exist but none are overdue) |
| Active Projects | **no** projects exist at all (`COUNT(*) FROM domain.project = 0`) | the active count (incl. `0` if projects exist but none are `Active`) |
| Inspections This Week | **no** inspections exist at all (`COUNT(*) FROM domain.inspection = 0`) | the this-week count (incl. `0` if inspections exist but none this week) |

**And** the empty signal flows through the components correctly: ComplianceTile receives `score=null` (its no-data variant); DashboardTile receives an empty/`None` value (its `—` variant). **Do not** pass `0` where `—` is intended, and **do not** suppress a real `0` (the Story 2.4 falsy-zero bug class — Django must use the `_dashboard_tile.html` zero-safe pattern, never `{% if value %}`).

**And** at this story's point in time (no seeded inspections/violations) the Overdue-Violations and Inspections-This-Week tiles render `—`; a per-stack test asserts this against an empty DB **and** asserts the number path against a fixture that inserts rows directly into `domain.violation` / `domain.inspection` (the tables exist even though no ORM model does — insert via raw SQL in the test).

**And** the Overdue-Violations `secondary` text renders the non-zero severity breakdown ordered Critical → High → Medium → Low (e.g. `"2 Critical, 1 High"`); empty when the overdue count is `0`.

### AC3 — `GET /` redirects to `/dashboard`

**Given** Story 1.13's Home page at `GET /`
**When** an authenticated user requests `GET /`
**Then** each stack returns **HTTP 302** with `Location: /dashboard` (the empty Epic-1 Home is retired). An unauthenticated request still hits the Story 1.11 login-redirect middleware first (→ `/login`), unchanged.

**And** the three stack `CLAUDE.md` "Home page" sections (which say "intentionally empty in Epic 1… Story 2.10 replaces it") are updated to describe the redirect.

**And** a per-stack test asserts `GET /` (authenticated) → 302 → `/dashboard`, and `GET /dashboard` renders 200.

### AC4 — Projects AG Grid fills the remaining viewport (reuse Story 2.9, navigate on row-click)

**Given** the tile row is rendered
**When** I look below it
**Then** the **AGGridPanel** (Story 2.9) renders the Projects grid pointed at `POST /grid/projects`, filling the remaining viewport (UX-DR19) — **the same component and endpoint as `/projects`**, not a reimplementation.

**And** the dashboard sets the AGGridPanel's **`data-grid-rowclick="navigate"`** mode: a row-click navigates the browser to `/projects/<id>` (the Project Detail screen / 2.8 stub) — there is **no** `#project-detail` rail on the dashboard. This requires extending the shared AGGridPanel JS (from 2.9) with the `navigate` branch (`window.location = "/projects/" + id`) alongside its existing `detail` (htmx.ajax) branch; `detail` remains the default so `/projects` is unaffected.

**And** the grid contains no business logic and renders no detail (FR51), identical to Story 2.9.

### AC5 — Responsive tile-row reflow (UX-DR30)

**Given** the four-tile row
**When** the viewport changes width
**Then** the row reflows: **4-up at ≥ 1280px**, **2×2 at 768–1279px**, **single column at < 768px** — implemented with Tailwind grid utilities (e.g. `grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4`), **identical class set across all three stacks**. A per-stack snapshot/structure test asserts the grid-column utility classes are present and identical.

**And** the AG Grid below the tiles remains full-width at every breakpoint (its own acknowledged-poor mobile behavior per UX-DR30 is unchanged).

### AC6 — Authorization: `dashboard.view` gates the page

**Given** Story 1.12's `can()` primitive and `RegisterAction` idiom
**When** I inspect each stack
**Then** a new `dashboard.view` action is registered, granted to **all five conceptual roles** (`ADMIN, COMPLIANCE_OFFICER, INSPECTOR, SITE_SUPERVISOR, EXECUTIVE`) — see Decisions note 5.
- **.NET:** `DomainPolicies.RegisterAction("dashboard.view", Role.Admin, Role.ComplianceOfficer, Role.Inspector, Role.SiteSupervisor, Role.Executive)` at startup.
- **Django:** `register_action("dashboard.view", Role.ADMIN, …)` at module load.
- **Go:** `auth.RegisterAction("dashboard.view", domain.RoleAdmin, …)` at composition time.

**And** `GET /dashboard` invokes `can(actor, "dashboard.view")` before doing any work; on `false` → HTTP **403** with no entity-state leakage (FR7, FR56), matching the canonical 403 from Story 1.11. (`GET /` redirects before any authz beyond authentication.) A per-stack test asserts an authenticated role → 200; a role lacking `dashboard.view` (if a no-role fixture user exists, e.g. `testuser`) → 403.

### AC7 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

**Given category 9 (empty / whitespace & zero values) — the central edge case**
**Then** every tile's empty-vs-zero behaviour is per AC2; tested for **all four** tiles against an empty DB (`—`) and against data-exists-but-count-zero (`0`). The Django stack uses the zero-safe `_dashboard_tile.html` pattern (per [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md) §`{% if value %}`).

**Given category 4 (AG Grid empty / loading overlays)**
**Then** the embedded grid's overlays behave exactly as Story 2.9 (no-rows overlay text, loading overlay) — reused, not re-styled.

**Given category 1 (unknown vocabulary)**
**Then** ComplianceTile out-of-range score → no-data em-dash (handled by 2.5); violation `severity` values are DB-CHECK-constrained to `{Low,Medium,High,Critical}` so the breakdown text never sees an unknown token.

**Given category 3 (JS init)**
**Then** the **tiles are server-rendered static markup** (zero JS — degrade perfectly with JS off); the **grid requires JS** and reuses Story 2.9's no-JS fallback message. A `javaScriptEnabled:false` test asserts the four tiles still render their values and the grid shows its fallback.

**Given category 8 (forced colors)**
**Then** the compliance color is paired with the threshold word + numeric value, and severity is conveyed as text — no color-only signal. axe scan on `/dashboard` reports zero new WCAG violations.

**Given categories 2, 5, 7**
**Then** N/A or covered by Story 1.14 global rules (no new fonts; no unbounded stacking; reduced-motion global).

### AC8 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

**Given category 1 (open redirect)**
**Then** `GET /` → `/dashboard` is a **fixed server-decided target** (not user-controlled) — no open-redirect surface. N/A beyond noting it.

**Given category 6 (CSRF) and the read-only nature**
**Then** the dashboard is a **read** (no transaction, no audit — Architecture read-handler shape); `GET /dashboard` and `GET /` are safe methods needing no CSRF token. The embedded `POST /grid/projects` keeps its Story 2.9 posture.

**Given categories 2, 3, 4, 5, 7**
**Then** N/A — no user input on the dashboard page itself (no filters/forms), no new cookies, no dynamic regex, no filesystem writes, stub-auth warning is Story 1.9's. The only user-controlled inputs reach the **grid endpoint**, which Story 2.9 already validates (allowlist).

### AC9 — `make parity` and full gate

**Given** Story 1.3 route-parity tooling
**When** I run `make parity`
**Then** `GET /dashboard` appears in all three stacks' route dumps; `GET /` remains present (now a redirect); `pg_indexes` diff is **zero** (no schema change — no new index; see Dev Notes). `make test-all` exits 0.

**Build/type/lint/test gates green per stack:**
- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean. New: `Pages/Dashboard/Index.cshtml(.cs)`, dashboard-stats read, dashboard handler/page tests, Home-redirect test.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean.
- **`fieldmark_shared`:** `pnpm run build` — `dist/fieldmark.css` byte-identical unless a documented responsive utility is genuinely missing (prefer existing Tailwind utilities; no `src/` edits expected). The AGGridPanel `navigate` mode is a JS edit in `vendor/ag-grid-panel/ag-grid-panel.js` (no CSS rebuild needed).
- **E2E:** a dashboard scenario (login → land on `/dashboard` via `/` redirect → assert four tiles present with correct ids → click a grid row → assert navigation to `/projects/<id>`).

### AC10 — Cross-stack parity of the dashboard composition

**Given** the Cross-Stack Architecture Principle
**Then** the dashboard is composed from **existing contracts** (DashboardTile/ComplianceTile canonical components + Story 2.9's SSRM grid) — this story introduces **no new wire/data contract**, so no new `docs/reference/` doc is required. The cross-stack invariant is the **page composition**: the four tile `id`s (AC1 table), their order, the `role="status"` on each, and the responsive grid utility classes (AC5) are **identical** across stacks — asserted by per-stack structure/snapshot tests. Each stack's dashboard page + stats read is native (Razor + EF/raw SQL; Django template + ORM/raw SQL; Go template + pgx) — no shared template fragment.

## Tasks / Subtasks

- [ ] **Task 1: Extend shared AGGridPanel with `navigate` row-click mode** (AC: #4)
  - [ ] 1.1 In `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (Story 2.9), add a `data-grid-rowclick` branch: `"navigate"` → `window.location = "/projects/" + id`; default `"detail"` → existing `htmx.ajax` into `data-grid-target`. Keep ≤ ~15 LOC budget.

- [ ] **Task 2: Dashboard-stats read per stack** (AC: #1, #2, #7)
  - [ ] 2.1 .NET: a read-only stats reader (e.g. `FieldMark.Web/Dashboard/DashboardStats.cs`) — `domain.project` aggregates via EF LINQ on `DbSet<Project>`; `violation`/`inspection` counts via `Database.SqlQueryRaw<int>` / scalar SQL. Returns `(int? portfolioScore, int? overdueTotal, string overdueBreakdown, int? activeProjects, int? inspectionsThisWeek)` with the empty-vs-zero nullability per AC2.
  - [ ] 2.2 Django: stats in the dashboard view — `Project.objects` for project aggregates; `with connection.cursor()` raw SQL for `violation`/`inspection`. Compute the em-dash-vs-zero nullability in Python; pass to the template.
  - [ ] 2.3 Go: stats via pgx (a small `DashboardStore` or inline in the handler) — `domain.project` AVG/COUNT + `violation`/`inspection` COUNT. Return a stats struct with pointer/`sql.Null*` fields for the empty distinction.
  - [ ] 2.4 All three: ISO-week-UTC bounds for "this week" (Monday 00:00 UTC → next Monday); severity breakdown ordered Critical→High→Medium→Low, non-zero only.
  - [ ] 2.5 Unit/integration tests: empty DB → all-`—` (portfolio, active 0-projects, overdue, inspections); seeded fixtures (raw-SQL inserts) → correct numbers incl. zero-with-data.

- [ ] **Task 3: `GET /dashboard` page + tile row per stack** (AC: #1, #5, #6, #10)
  - [ ] 3.1 .NET: `Pages/Dashboard/Index.cshtml(.cs)` — authorize `dashboard.view`; render the four tiles via `_ComplianceTile`/`_DashboardTile` partials with the canonical ids + `role_status=true`; AGGridPanel below with `data-grid-rowclick="navigate"`; responsive grid classes.
  - [ ] 3.2 Django: dashboard view in `fieldmark/views.py` + URL in `fieldmark/urls.py`; `templates/dashboard/index.html` including the component partials + AGGridPanel.
  - [ ] 3.3 Go: dashboard handler + route registration in `cmd/web/main.go`; `internal/web/templates/pages/dashboard.html`.
  - [ ] 3.4 Register `dashboard.view` for all five roles per stack.
  - [ ] 3.5 Per-stack tests: 200 render, four tile ids + order + `role="status"`, responsive class set, 403 for a role without `dashboard.view`.

- [ ] **Task 4: `GET /` → `/dashboard` redirect** (AC: #3)
  - [ ] 4.1 .NET: `IndexModel.OnGet` returns `RedirectToPage("/Dashboard/Index")` (or `Redirect("/dashboard")`). Keep the page registered.
  - [ ] 4.2 Django: `home` view (`fieldmark/views.py:84`) returns `redirect("/dashboard")`.
  - [ ] 4.3 Go: `app.Get("/")` handler (`cmd/web/main.go:136`) redirects to `/dashboard`.
  - [ ] 4.4 Update the three stack `CLAUDE.md` "Home page" sections to describe the redirect (no longer "intentionally empty").
  - [ ] 4.5 Per-stack test: authenticated `GET /` → 302 → `/dashboard`.

- [ ] **Task 5: E2E + parity + gate** (AC: #9)
  - [ ] 5.1 E2E dashboard scenario (redirect → tiles → row-click navigate).
  - [ ] 5.2 `make parity` (`GET /dashboard` present, `pg_indexes` zero-diff) + `make test-all` green.

- [ ] **Task 6: Story sign-off** (AC: all)
  - [ ] 6.1 Populate the Sign-off block; record the five decisions; flip sprint-status to `review`.

### Review Findings

- [x] [Review][Patch] Django home tests stale — 7+ tests assert 200 + rendered body after `home` was changed to `redirect("/dashboard")`; all will fail at `make test-all` [AC3, AC9] [`fieldmark_py/fieldmark/tests/test_home_page.py`]
- [x] [Review][Patch] Go home_test fixture wires a rendering route, not a redirect — `buildHomeApp` renders `pages/home` while `TestHome*` assertions expect 302 and `/dashboard`; tests will fail [AC3, AC9] [`fieldmark-go/internal/web/handlers/home_test.go`]
- [x] [Review][Patch] Portfolio compliance score rounding diverges between stacks — Go uses `+0.5` truncation (round-half-up) while .NET `Math.Round` and Python `round()` use banker's rounding; `x.5` inputs produce different scores on the Go stack [AC10] [`fieldmark-go/internal/web/handlers/dashboard_handler.go`]
- [x] [Review][Patch] Django inspections-this-week uses SQL `date_trunc` instead of app-code ISO-week boundaries — spec constraint requires bounds computed in app code and passed as parameters; Django delegates to PostgreSQL `date_trunc('week', timezone('UTC', now()))` [Key Constraint] [`fieldmark_py/fieldmark/views.py:115`]
- [x] [Review][Patch] `rows.Err()` unchecked in Go after severity query loop — mid-stream network error terminates `rows.Next()` silently, returning incorrect partial severity counts [AC2] [`fieldmark-go/internal/web/handlers/dashboard_handler.go`]
- [x] [Review][Patch] Missing per-stack dashboard page integration tests — no `GET /dashboard` test for any stack asserting 200 (authorized role), 403 (unauthorized), or responsive grid-column classes [AC5, AC6]
- [x] [Review][Patch] Missing aggregate-value correctness tests (empty-vs-zero) for all four tiles — no test exercises `DashboardStatsReader` / `dashboard()` view / Go stats read against controlled DB states (table empty → `—`; data exists, count zero → `0`) [AC2, AC7]
- [x] [Review][Patch] Go dead nil-actor guard after `auth.Can()` 403 branch — `if actor == nil { actor = app.Anonymous() }` is unreachable because `Can()` returns `false` for nil and the 403 return fires first; misleads future maintainers [`fieldmark-go/internal/web/handlers/dashboard_handler.go:29`]
- [x] [Review][Patch] AG Grid navigate mode: use `window.location.href` and `encodeURIComponent(id)` — bare `window.location = '/projects/' + e.data.id` omits encoding and uses the less-explicit property form [`fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js:50`]
- [x] [Review][Patch] `BuildOverdueBreakdownAsync` throws `KeyNotFoundException` on unknown severity value — pre-seeded dict lookup has no `TryGetValue` guard; a severity value outside `{Critical,High,Medium,Low}` causes a 500 [AC7] [`FieldMark/FieldMark.Web/Dashboard/DashboardStatsReader.cs`]
- [x] [Review][Patch] `DashboardStatsReader` opens EF Core connection via `GetDbConnection().OpenAsync` without `db.Database.OpenConnectionAsync()` — bypasses EF Core connection lifecycle management; can fail if EF has already opened the connection [`FieldMark/FieldMark.Web/Dashboard/DashboardStatsReader.cs`]
- [x] [Review][Defer] Go home chrome tests exercise dead fixture — `buildHomeApp` wires `pages/home` rendering, not the redirect; tests pass but do not exercise production behavior [`fieldmark-go/internal/web/handlers/home_test.go`] — deferred, pre-existing test fixture design
- [x] [Review][Defer] Go nil-pool `/dashboard` branch returns empty HTTP 200 — dev-only stub consistent with other no-pool routes in `main.go` [`fieldmark-go/cmd/web/main.go:161`] — deferred, pre-existing stub pattern
- [x] [Review][Defer] `make parity` route-dump check is a no-op — tooling not yet scaffolded (pre-existing infrastructure gap) [`Makefile`] — deferred, pre-existing

#### Rerun findings

- [x] [Review][Patch] Missing `role="status"` assertions in all three stacks' per-stack tests — `DashboardPageTests.cs`, `test_dashboard_page.py`, `dashboard_handler_test.go` all assert tile IDs but none assert `role="status"` on each tile [AC1]
- [x] [Review][Patch] Go template-source test is too weak — `TestDashboardTemplate_ContainsTileIdsAndResponsiveGridClasses` reads raw `dashboard.html` and checks Go template variable names (`PortfolioTile`, etc.), not rendered canonical IDs (`id="compliance-tile-portfolio"`, etc.); a handler wiring bug would not be caught [AC1, AC5] [`fieldmark-go/internal/web/handlers/dashboard_handler_test.go`]
- [x] [Review][Patch] Unauthenticated `GET /dashboard` behavior untested in all three stacks — no test asserts that an anonymous/unauthenticated request to `/dashboard` redirects to `/login` (vs. returning 403); behavior is correct via middleware but unverified [AC6, AC9]
- [x] [Review][Patch] Go `readStats` uses `time.Now().UTC()` directly — week-boundary bounds are not injectable; spec constraint requires "a per-stack test pins a fixed `now()` and asserts identical bounds"; Go cannot satisfy this without a clock parameter [Spec constraint / AC9] [`fieldmark-go/internal/web/handlers/dashboard_handler.go`]
- [x] [Review][Defer] AG Grid `detail` mode silently drops row-click when `data-grid-target` is absent (no `console.warn`) — pre-existing Story 2.9 behavior, not introduced by this diff [`fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js`] — deferred, pre-existing

#### Rerun 2 findings

- [x] [Review][Patch] Go 403 test passes vacuously — `dashboard.view` is never registered in the test package `init()`, so `auth.Can()` returns `false` for ALL actors; `TestDashboard_NoPermissionRole_Returns403` does not verify role-based denial, only that an unregistered action blocks everyone [AC6] [`fieldmark-go/internal/web/handlers/dashboard_handler_test.go`]
- [x] [Review][Patch] Go rendered-HTML test missing `role="status"` on compliance tile — `TestDashboardTemplate_ContainsRenderedTileIdsAndRoleStatus` checks `id="compliance-tile-portfolio"` without `role="status"`, while the other three DashboardTile needles include the combined assertion; the compliance tile template does emit `role="status"` so adding it would pass [AC1, AC10] [`fieldmark-go/internal/web/handlers/dashboard_handler_test.go:85`]
- [x] [Review][Patch] Go nil-pool panic once `dashboard.view` is registered — once Finding 1 above is fixed by adding `RegisterAction` in test `init()`, any authorized-actor test path will reach `h.readStats()` which calls `h.Pool.QueryRow(...)` on a nil pool; a nil-pool guard or stub response is needed in `readStats` to maintain the existing deferred stub pattern [`fieldmark-go/internal/web/handlers/dashboard_handler.go`]
- [x] [Review][Patch] Django `test_dashboard_page.py` DB-touching tests lack `@pytest.mark.django_db` decorator — `test_dashboard_authenticated_admin_renders_200`, `test_dashboard_no_role_returns_403`, `test_dashboard_renders_tile_ids_and_responsive_classes` use the `db` fixture (which grants access) but inconsistently omit the explicit decorator; `test_dashboard_unauthenticated_redirects_to_login` lacks both decorator and `db` fixture [`fieldmark_py/fieldmark/tests/test_dashboard_page.py`]
- [x] [Review][Patch] Django dead variable `active_projects` in `dashboard` view — `active_projects = None if project_count == 0 else active_count` is computed but never referenced; `dashboard_context_from_raw` re-derives it from the same inputs; remove the dead local [`fieldmark_py/fieldmark/views.py`]

#### Rerun 3 findings

- [x] [Review][Patch] Django `test_home_unauthenticated_redirects_to_login` missing `@pytest.mark.django_db` — every other test in `test_home_page.py` carries the decorator; this one does not and also lacks the `db` fixture; safe today because the unauthenticated redirect fires before any DB access, but fragile if middleware ever touches the session store [`fieldmark_py/fieldmark/tests/test_home_page.py`]

## Dev Notes

### Critical context (read before writing code)

- **The dashboard is a pure read — no transaction, no audit.** Authorize → run the stats queries → render. Do **not** wrap in `transaction.atomic`/`IDbContextTransaction`/`pgx.Tx`, and write **no** `AuditEntry` (reads are not audited). Opposite discipline from Story 2.8's write flow.

- **Reuse the existing tile components — do not invent markup.** ComplianceTile (2.5) takes `score (int?)`, `label`, `id` and bakes in the threshold bands + ARIA; pass `score=portfolioAvg` (or `null`), `label="Portfolio Compliance"`, `id="compliance-tile-portfolio"`. DashboardTile (2.4) takes `tile_id`, `label`, `value`, `secondary`, `value_color`, `role_status`; pass `role_status=true` and the empty value as `null`/empty (→ `—`). The components already handle em-dash, zero-safety, and forced-colors text pairing — leaning on them is the whole point of Phase-2.

- **Empty-vs-zero is the #1 review trap (Story 2.4 history).** `—` means "the source set is empty"; `0` means "data exists, the count is zero". See AC2's table for each tile's exact rule. Django: the `_dashboard_tile.html` zero-safe pattern is mandatory — `{% if value %}` is falsy for `0` and would wrongly render `—`. .NET/Go: pass `int?`/`*int`/`sql.NullInt` and map `null → —`, `0 → "0"`.

- **Aggregates via direct SQL on existing tables (Decisions note 1).** `domain.violation` and `domain.inspection` are in the DDL but have **no** ORM model yet (Epic 3 maps them). The dashboard does **not** add those models — it issues read-only `COUNT` SQL. .NET: `context.Database.SqlQueryRaw<int>("SELECT count(*) FROM domain.violation WHERE …")`. Django: `with connection.cursor() as cur: cur.execute(…)`. Go: `pool.QueryRow(ctx, "SELECT count(*) …")`. The `domain.project` aggregates (portfolio avg, active count) **do** use the existing Story 2.1 mapping (EF `DbSet<Project>` / Django `Project.objects` / Go store) where convenient, or plain SQL — either is fine; keep it read-only.

- **Canonical aggregate queries** (parameterize timestamps; bind, don't concat):
  - Portfolio score: `SELECT ROUND(AVG(compliance_score)) FROM domain.project WHERE status <> 'Closed'` → `NULL` when no rows → ComplianceTile `score=null`.
  - Active Projects: `SELECT count(*) FROM domain.project WHERE status = 'Active'`; em-dash when `SELECT count(*) FROM domain.project = 0`.
  - Overdue Violations: `SELECT severity, count(*) FROM domain.violation WHERE status IN ('Open','InProgress') AND due_at < now() GROUP BY severity` (uses `idx_violation_due`); em-dash when `SELECT count(*) FROM domain.violation = 0`. Sum the group counts for the value; build the `secondary` string from non-zero severities Critical→High→Medium→Low.
  - Inspections This Week: `SELECT count(*) FROM domain.inspection WHERE scheduled_for >= $1 AND scheduled_for < $2` with `$1` = Monday 00:00:00 UTC of the current ISO week, `$2` = `$1 + 7 days`; em-dash when `SELECT count(*) FROM domain.inspection = 0`.

- **"This week" = ISO week in UTC** for deterministic cross-stack parity (Monday start). Compute the bounds in app code (not in SQL date functions, which differ subtly per stack) and bind them as parameters. Document the choice; a per-stack test pins a fixed `now()` and asserts identical bounds.

- **Reuse Story 2.9's AGGridPanel — extend, don't fork.** The dashboard embeds the exact same `<div class="ag-theme-quartz" data-grid-endpoint="/grid/projects" …>` + init JS. The only addition is the `navigate` row-click mode (Task 1). Do **not** copy the grid init into a dashboard-specific script.

- **No new indexes.** "Inspections this week" filters `scheduled_for`, which has no index (the DDL indexes are `idx_inspection_project_status`, `idx_violation_due`, audit indexes). At MVP scale (empty/few inspections) a seq scan is fine. Adding an index is infra-owned (`docker/postgres/init/`), needs `make reset`, and breaks `pg_indexes` zero-diff. Out of scope; note as a future perf lever.

- **`dashboard.view` for all roles; `GET /` redirect.** Mirror Story 2.9's `project.read` all-roles posture (no `PROJECT_MANAGER` role exists; the dashboard is the landing page for every authenticated user). The `GET /` redirect target is a fixed string — not an open-redirect surface.

- **Tile `role="status"` now, OOB wiring later.** Each tile gets a stable `id` (AC1) and `role="status"` so 2.12 / Epic 5 can OOB-replace it in a mutation response — but this story emits **no** `hx-swap-oob` and **no** producer attributes. UX-DR45 is satisfied because the tiles are correct at every static state.

### Source tree — where things land

| Stack | `/dashboard` page | Stats read | `GET /` redirect |
|---|---|---|---|
| .NET | `FieldMark.Web/Pages/Dashboard/Index.cshtml(.cs)` | `FieldMark.Web/Dashboard/DashboardStats.cs` | `Pages/Index.cshtml.cs` `OnGet` → redirect |
| Django | `fieldmark/views.py` dashboard view + `templates/dashboard/index.html` + `fieldmark/urls.py` | inline in the view (ORM + raw cursor) | `fieldmark/views.py:84` `home` → `redirect("/dashboard")` |
| Go | handler + `internal/web/templates/pages/dashboard.html` + route in `cmd/web/main.go` | pgx in handler / small `DashboardStore` | `cmd/web/main.go:136` `app.Get("/")` → redirect |

Shared: `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (extend with `navigate` mode).

### Existing code to reuse (read before writing)

- **ComplianceTile** wrapper + `compliance-tile-portfolio` variant — [_ComplianceTile.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/Components/_ComplianceTile.cshtml), [_compliance_tile.html](../../fieldmark_py/templates/components/_compliance_tile.html), [compliance_tile.html](../../fieldmark-go/internal/web/templates/components/compliance_tile.html); contract [README](../../fieldmark_shared/components/compliance_tile/README.md).
- **DashboardTile** wrapper — [_DashboardTile.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/Components/_DashboardTile.cshtml), [_dashboard_tile.html](../../fieldmark_py/templates/components/_dashboard_tile.html) (zero-safe pattern), [dashboard_tile.html](../../fieldmark-go/internal/web/templates/components/dashboard_tile.html); contract [README](../../fieldmark_shared/components/dashboard_tile/README.md).
- **AGGridPanel** + `POST /grid/projects` — Story 2.9 ([2-9-…md](2-9-project-list-ag-grid-with-server-side-row-model.md)).
- **`can()` / RegisterAction** — Story 1.12 (`DomainPolicies` / `fieldmark.authz` / `auth.Can`); register `dashboard.view` alongside 2.8's `project.create` and 2.9's `project.read`.
- **Home handlers** — [Index.cshtml.cs](../../FieldMark/FieldMark.Web/Pages/Index.cshtml.cs), [fieldmark/views.py](../../fieldmark_py/fieldmark/views.py) (`home`), [cmd/web/main.go](../../fieldmark-go/cmd/web/main.go) (`app.Get("/")`).

### Project Structure Notes

- Adds `GET /dashboard` to the route inventory (parity); `GET /` stays (now 302→`/dashboard`).
- No `domain.*` schema change; `pg_indexes` zero-diff.
- Extends one shared JS file (AGGridPanel `navigate` mode); no new vendor asset.
- Updates the three stack `CLAUDE.md` Home-page sections.
- No new cross-stack `docs/reference/` contract (composition of existing contracts).

### References

- Epic AC: [epic-2 §Story 2.10](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md)
- DDL: [010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) (project / inspection / violation)
- Dashboard layout + tile row: [ux-design-specification.md:545](../planning-artifacts/ux-design-specification.md), [:620](../planning-artifacts/ux-design-specification.md); DashboardTile [:880–885](../planning-artifacts/ux-design-specification.md); responsive UX-DR30 [:1166](../planning-artifacts/ux-design-specification.md)
- Component contracts: [compliance_tile/README.md](../../fieldmark_shared/components/compliance_tile/README.md), [dashboard_tile/README.md](../../fieldmark_shared/components/dashboard_tile/README.md)
- Read-handler shape (no tx, no audit): [architecture.md:880–887](../planning-artifacts/architecture.md)
- Edge cases / security: [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) cat 9/4/1/3/8, [security-defaults.md](../../docs/reference/security-defaults.md)
- Prior stories: [2-9-…md](2-9-project-list-ag-grid-with-server-side-row-model.md) (AGGridPanel, all-roles read gate), [2-4 / 2-5 components], Django zero-safe pattern [fieldmark_py/CLAUDE.md](../../fieldmark_py/CLAUDE.md)

## Dev Agent Record

### Agent Model Used

GPT-5 (Codex)

### Debug Log References

- 2026-05-31: Implemented shared AG Grid row-click `navigate` mode in `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js`.
- 2026-05-31: Added initial `/dashboard` routing/rendering and `dashboard.view` registration in `.NET`, `Django`, and `Go` stacks.
- 2026-05-31: Updated `GET /` handlers toward dashboard redirect behavior across stacks.
- 2026-05-31: Ran `make test-go`, `make test-django`, and `make test-net`; gates are not yet green.
- 2026-05-31: Detected unexpected workspace changes unrelated to this story (`.agents/skills/luokai0-karpathy-coder/`, `.claude/skills/karpathy-coder`) and halted per repository safety rules.

### Completion Notes List

- Story remains **in-progress**; acceptance criteria are not fully validated yet.
- Dashboard foundation is partially implemented across shared assets and all three stacks.
- Halted before completion because unexpected non-story workspace changes were detected.

### File List

- `FieldMark/FieldMark.Web/Dashboard/DashboardStatsReader.cs` (new)
- `FieldMark/FieldMark.Web/Pages/Dashboard/Index.cshtml` (new)
- `FieldMark/FieldMark.Web/Pages/Dashboard/Index.cshtml.cs` (new)
- `FieldMark/FieldMark.Web/Program.cs` (modified)
- `FieldMark/FieldMark.Web/Pages/Index.cshtml.cs` (modified)
- `fieldmark_py/fieldmark/views.py` (modified)
- `fieldmark_py/fieldmark/urls.py` (modified)
- `fieldmark_py/templates/dashboard/index.html` (new)
- `fieldmark-go/cmd/web/main.go` (modified)
- `fieldmark-go/internal/web/handlers/dashboard_handler.go` (new)
- `fieldmark-go/internal/web/templates/pages/dashboard.html` (new)
- `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (modified)
- `FieldMark/CLAUDE.md` (modified)
- `fieldmark_py/CLAUDE.md` (modified)
- `fieldmark-go/CLAUDE.md` (modified)
- `FieldMark/FieldMark.Tests.Web/Pages/HomePageTests.cs` (modified)
- `fieldmark-go/internal/web/handlers/home_test.go` (modified)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified)
## Sign-off

- Date of final review: 2026-05-31
- Total review-round count: 5 (round 1 + reruns 1–4)
- Final reviewer verdict (PASS/FAIL): **PASS**
- Deferred-work entries created from this story: 5 (see `deferred-work.md` — Go home chrome tests, Go nil-pool 200, `make parity` no-op, AG Grid `detail` mode silent no-op, Go nil-pool authorized-200 test)
- Decisions requiring ratification (recorded here; confirm or overturn at review):
  1. **Dashboard aggregates via direct read-only SQL** on `domain.violation`/`domain.inspection` (tables exist; no Epic-3 ORM mapping pulled forward). Those two tiles render `—` until Epic 3 seeds data. **RATIFIED.**
  2. **Portfolio score = `ROUND(AVG(compliance_score))` over non-`Closed` projects**; `NULL` → em-dash. **RATIFIED.**
  3. **Overdue-Violations severity breakdown is `secondary` text**, not badge chips (DashboardTile contract has one text line; chips would need a 2.4 component extension). **RATIFIED.**
  4. **Dashboard grid row-click navigates to `/projects/<id>`** via a new `navigate` mode on the shared AGGridPanel (no rail on the dashboard). **RATIFIED.**
  5. **`dashboard.view` granted to all five roles; `GET /` 302→`/dashboard`.** **RATIFIED.**
