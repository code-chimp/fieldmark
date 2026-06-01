# Story 2.11: Project Detail anchor screen with header strip, TabStrip, and EntityRail

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.11
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — `domain.project` (58–73), `domain.project_trade_scope` (84–89), `domain.project_inspector` (93–97)

Depends on (all **done** unless noted):
- **Story 2.1** — `domain.project` + `job_site` + `project_trade_scope` + `project_inspector` mappings per stack. Go has `ProjectStore.LoadWithRelations(ctx,id) → (*Project, []JobSite, []ProjectTradeScope, []ProjectInspector, error)` ([projectstore.go:103](../../fieldmark-go/internal/data/postgres/projectstore.go)). .NET has `DbSet<Project|ProjectTradeScope|ProjectInspector>` (no navigation collections — query each by `project_id`). Django has `Project.trade_scopes` / `Project.inspector_assignments` related managers ([projects/models.py:134,154](../../fieldmark_py/projects/models.py)).
- **Story 2.3** — reference data read API + in-process cache (`TradeType` by id → `name`). Used to resolve assigned-trade display names.
- **Story 2.4** — **StatusBadge** wrapper (`_StatusBadge.cshtml` / `_status_badge.html` / `status_badge.html`; props `entity, value, severity`; `project:Active|OnHold|Closed` → labelled badge; `badge-unknown` fallback).
- **Story 2.5** — **ComplianceTile** wrapper + `#compliance-tile` target (props `score (int?, null→em-dash), label, id`; `role="status"` intrinsic). This story renders the **singular** `id="compliance-tile"` variant in the header strip (FR36) — **not** `compliance-tile-portfolio` (that is the dashboard).
- **Story 2.6** — **EntityRail** wrapper (`_EntityRail.cshtml` / `_entity_rail.html` / `entity_rail.html`; props `id, entity_type_label, entity_loaded, body_slot?, footer_slot?`). Renders the empty/loaded `<aside>` shell with `tabindex="-1"`, `role="region"`, `aria-live="polite"`. See [README](../../fieldmark_shared/components/entity_rail/README.md) §8 for the **canonical Project Detail layout** this story owns.
- **Story 2.7** — **TabStrip** wrapper (`_TabStrip.cshtml` / `_tab_strip.html` / `tab_strip.html`; props `id, aria_label, tabs[TabSpec{id,label,hx_get,hx_target,badge_count?}], active_index, hx_swap?`). Pairs with vendored `tabstrip/tabstrip.js` (already loaded in all three base layouts). See [README](../../fieldmark_shared/components/tab_strip/README.md) §9 for the **active-tab OOB-swap contract** this story consumes.
- **Story 2.8** — `Project.create`, `GET /projects/new`, `POST /projects/`; the `HX-Redirect`/`303` target is `/projects/<id>` (this screen). The stub it lands on is replaced here.
- **Story 2.9** — Projects list at `/projects` with an **`<aside id="project-detail">` rail**; the grid row-click (default `detail` mode) does `htmx.ajax('GET','/projects/'+id,{target:'#project-detail',swap:'innerHTML'})` ([ag-grid-panel.js:57](../../fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js)). **This is why `GET /projects/<id>` must be dual-mode — see Decision 1.** `project.read` action (all five roles) registered there.
- **Story 1.12** — `can(actor, action)` + `RegisterAction`. **Story 1.5** — base-layout chrome (`#flash-region`, header, footer). **Story 2.10** — dashboard grid uses `navigate` mode → full-page `/projects/<id>`.

## Story

As any authorized user,
I want to view a Project's current truth on a single page at `/projects/<id>` — a header strip (code + name + StatusBadge + ComplianceTile), an HTMX-driven TabStrip (Summary / Inspections / Violations / Audit), a Summary panel showing the project's metadata + a permission-gated action-button row, and a sticky co-present EntityRail at ≥ 1280px,
So that I can read status, score, and project facts without page navigation (FR11) — and the tab-swap orchestration (TabStrip §9 OOB), the EntityRail co-presence + responsive collapse (UX-DR18/24/30), and the affordance trichotomy (UX-DR10/21) are locked in for Stories 2.12 (place-on-hold), 2.13 (audit tab), and Epics 3–6 (inspections/violations/closure).

**Scope boundary.** This story produces, per stack:
- (a) `GET /projects/<id>` — the full Project Detail screen, **dual-mode** (Decision 1): full page on direct navigation; body partial on `HX-Request` (so the Story 2.9 list rail keeps working).
- (b) Four tab-content endpoints `GET /projects/<id>/tabs/{summary,inspections,violations,audit}` — each returns the tab-panel inner HTML **plus** an OOB `<nav id="project-detail-tabstrip">` with the new active tab (TabStrip §9).
- (c) The **Summary** tab content: code, name, start date, target completion date, description, assigned trades (names via Story 2.3 cache), assigned inspectors (display names via per-stack auth lookup), and the **Place on Hold / Resume / Close** ActionButton row (trichotomy — Decision 4).
- (d) The **header strip** (code + name + StatusBadge + ComplianceTile `#compliance-tile`), the **TabStrip** (`#project-detail-tabstrip`), the tab-panel container `#project-detail-tab-content`, and the empty **EntityRail** (`id="violation-detail"` — Decision 5).
- (e) The **`project-detail-grid`** CSS container (sticky rail ≥1280px / stacked <1280px — UX-DR30) added to `fieldmark_shared/src/_layout.css` under `/* EntityRail responsive collapse */` (the EntityRail README §8 reserves this for Story 2.11).
- (f) The `project.place_on_hold` / `project.resume` / `project.close` action registrations + read-only `can_*` status predicates on the Project entity (Decision 4).
- (g) A cross-stack contract doc `docs/reference/project-detail-contract.md` + per-stack conformance test (Decision 2 / AC10).
- (h) Per-stack tests + `make parity`.

**Out of scope:**
- **The actual place-on-hold / resume / close transitions** — entity transition methods, the `POST` handlers, audit writes, and three-region OOB. Those are **Story 2.12** (hold/resume) and **Epic 6** (close). This story renders the *affordances* (buttons) per the trichotomy; the buttons' `hx-post` targets resolve to 404 until those stories land (Decision 4 — accepted, epic-sanctioned deferral).
- **Real Inspections / Violations / Audit tab content.** Those lists need Epic 3 (Inspection), Epic 4 (Violation), and Story 2.13 (audit). This story renders **placeholder panels** for the three non-Summary tabs (a single InlineAlert-styled "Coming in a later story" message inside a correctly-labelled `role="tabpanel"`), so the tab mechanics + ARIA + parity are provable now. (Story 2.13 replaces the Audit panel; Epic 3/4 replace Inspections/Violations.)
- **EntityRail row-selection wiring** (loading an inspection/violation into the rail). The rail renders **empty** here; selecting a row to populate it is Epic 3 (inspection-detail) / Epic 4 (violation-detail). The rail's *presence, responsive collapse, focus surface, and survival across tab swaps* are in scope.
- **Compliance score recompute / live tile OOB updates** — the `#compliance-tile` renders the stored `compliance_score`; no mutation occurs on this read page.
- Any `domain.*` schema change (`pg_indexes` zero-diff).

---

## ⚠️ Decisions baked into this story (read first)

Each is implemented as written and listed in the Sign-off block for reviewer ratification.

1. **`GET /projects/<id>` is dual-mode.** Story 2.9's `/projects` list page contains `<aside id="project-detail">` and its grid row-click swaps `innerHTML` of `#project-detail` with the response of `GET /projects/<id>`. The dashboard (2.10) and direct navigation hit the **same** URL as a full page. Therefore:
   - **`HX-Request: true` header present** → return the **body partial**: the `<section id="project-header-strip">` + `<div class="tabs">`(TabStrip + `#project-detail-tab-content` with Summary pre-rendered) + `<aside id="violation-detail">`, **without** base chrome (no `<html>/<head>/<header>/<footer>`). This is what swaps into the list's `#project-detail` rail.
   - **No `HX-Request`** (direct nav / dashboard navigate / 2.8 redirect) → return the **full page** (base-layout chrome wrapping the same body partial).
   - Both paths render byte-identical body markup; only the chrome wrapper differs. Use each stack's idiomatic partial-vs-layout switch (Razor `if (Request.IsHtmx()) return Partial(...)`; Django `request.headers.get("HX-Request")` choosing template; Go `c.Get("HX-Request")` choosing `c.Render(..., layout)` vs fragment render with `""` layout — see [main.go:172](../../fieldmark-go/cmd/web/main.go) for the no-layout `c.Render(fragment, …, "")` idiom).

2. **Tab-content route scheme = `GET /projects/:id/tabs/{summary,inspections,violations,audit}`.** This matches the epic's explicit parity AC. The TabStrip canonical fixture used `/projects/__ID__/summary` as an *illustrative placeholder*; the wrapper accepts whatever `hx_get` the consumer passes, so this choice does **not** break Story 2.7's snapshot tests (they assert the wrapper renders the placeholder URLs, not these). Each tab endpoint returns the panel inner HTML **plus** the OOB tabstrip (TabStrip §9). This scheme + the dual-mode rule are documented in `docs/reference/project-detail-contract.md` (AC10).

3. **No "PM name" row.** The epic AC text lists "PM name", but [ag-grid-ssrm-contract.md:139](../../docs/reference/ag-grid-ssrm-contract.md) already ratified that **`domain.project` has no project-manager column, there is no `domain.user` table, and ADR-012 forbids domain→auth FKs** — so `pm_name` was dropped from the grid contract. The Summary tab is consistent with that ratified contract: it renders **assigned inspectors** (the only people-link, via `domain.project_inspector`) and **omits a Project Manager field** entirely. If a PM concept is ever introduced it follows the ssrm-contract Change Procedure (schema + three handlers + three tests).

4. **The action-button row is rendered via the trichotomy now; the transitions land later.** The epic AC requires the Place on Hold / Resume / Close row "using the affordance trichotomy" and requires Executive to see **no** buttons. To make the trichotomy real:
   - **Register three actions** (per stack, at composition time, alongside 2.8's `project.create` / 2.9's `project.read`): `project.place_on_hold`, `project.resume`, `project.close`, each granted to **`ADMIN`** only (mirrors `project.create`; the minimal authorized set — Story 2.12 may broaden to other roles and will ratify). Executive lacks all three → **absent** (satisfies the Executive AC).
   - **Add three pure, read-only `can_*` predicates** to the Project entity (no transition methods — those are 2.12/Epic 6): `CanPlaceOnHold()` ⇔ `Status == Active`; `CanResume()` ⇔ `Status == OnHold`; `CanClose()` ⇔ `Status == Active` (the *closure gate* — open-violation / required-inspection — is **not** evaluated here; it lands in Epic 6. This predicate is the coarse status-only gate, documented as such).
   - **ActionButton `hx-post` targets** are the canonical 2.12/Epic-6 endpoints: `/projects/<id>/place-on-hold`, `/projects/<id>/resume`, `/projects/<id>/close`, with `hx-target="#project-detail"` and `hx-swap="outerHTML"` (the three-region convention 2.12 will satisfy). **Until 2.12/Epic 6 ship those handlers, clicking a present button yields HTTP 404** — an accepted, epic-sanctioned deferral (the affordance is the deliverable here, not the transition). Do **not** invent stub handlers.
   - Trichotomy outcomes on this screen: Executive → **absent** (no permission); ADMIN on an `Active` project → **present** Place-on-Hold + present Close, **disabled** Resume (reason: "Project is not on hold"); ADMIN on an `OnHold` project → **present** Resume, **disabled** Place-on-Hold ("Project is already on hold") + disabled Close ("Only active projects can be closed"); ADMIN on a `Closed` project → all three **disabled**.

5. **The Project Detail EntityRail uses `id="violation-detail"`, rendered empty.** Per the EntityRail [README §8](../../fieldmark_shared/components/entity_rail/README.md) canonical Project Detail layout, the rail in `<main>` is `<aside id="violation-detail">`. This story renders the **empty-violation** variant (no entity selected). Row-selection that swaps an inspection/violation partial into the rail is Epic 3/4. The rail is a **sibling** of `#project-detail-tab-content` inside the `project-detail-grid` container (not a descendant) so a tab swap (which targets `#project-detail-tab-content`) never clears it (UX-DR24).

6. **Default tab = Summary, server-rendered inline on first load.** On the initial `GET /projects/<id>`, the `#project-detail-tab-content` panel already contains the fully-rendered Summary content (not an empty shell awaiting an HTMX fetch). This keeps the page functional with JS disabled (cat 3) and avoids a flash-of-empty-panel. Tab *clicks* fetch via HTMX. `GET /projects/<id>/tabs/summary` returns the same Summary panel markup (for when the user navigates back to Summary).

---

## Acceptance Criteria

### AC1 — `GET /projects/<id>` renders the Project Detail anchor screen

**Given** I am authenticated with `project.read` and navigate to `GET /projects/<id>` for an existing project
**When** the full page renders
**Then** each stack renders (inside Story 1.5 chrome) a single **Project Detail shell** carrying `id="project-detail"` (the canonical "main detail panel" target — the same element the Story 2.9 list row-click loads into) wrapping, in this order:

1. `<section id="project-header-strip">` with: the project **code** and **name** (`<h1>` — exactly one per page, UX-DR33), a **StatusBadge** (`entity="project"`, `value=<status>`) for the current status, and the **ComplianceTile** wrapper with `id="compliance-tile"`, `label="Compliance"`, `score=<project.compliance_score>` (FR36).
2. A `<div class="tabs">` wrapper (Basecoat outer container — TabStrip README §7) containing:
   - the **TabStrip** wrapper: `id="project-detail-tabstrip"`, `aria_label="Project Detail Tabs"`, `active_index=0`, four tabs in order **Summary / Inspections / Violations / Audit**, each `hx_get="/projects/<id>/tabs/<name>"`, `hx_target="#project-detail-tab-content"`, no badges;
   - the panel `<div id="project-detail-tab-content" role="tabpanel" aria-labelledby="tab-summary" tabindex="-1">` containing the **server-rendered Summary content** (Decision 6 / AC2).
3. `<aside id="violation-detail" …>` — the **empty** EntityRail (Decision 5), a sibling of `#project-detail-tab-content` inside the `project-detail-grid` container.

**And** the `#compliance-tile` id appears **exactly once** on the page (it is the header tile; there is no `-portfolio` variant here).
**And** the canonical ids `project-detail`, `project-header-strip`, `project-detail-tabstrip`, `project-detail-tab-content`, `compliance-tile`, and `violation-detail` are present and spelled exactly.

> **Reviewer note — the `#project-detail` shell wrapper (carried over from the 2026-05-31 design side-session; reconciled to the `innerHTML` swap).** The body partial (header strip + `.project-detail-grid` + rail) must be the **inner content** of a stable element carrying `id="project-detail"` that exists in **both** render modes (Decision 1): on the standalone full page the `<main>` (or a `<div>` directly inside it) carries `id="project-detail"`; in the Story 2.9 list-embedded mode the existing `<aside id="project-detail">` *is* that wrapper. This story only **establishes** the shell — it wires no whole-panel re-render. **Story 2.12's place-on-hold/resume three-region orchestration targets `#project-detail` with `hx-swap="innerHTML"`**, re-rendering the body partial (new StatusBadge + flipped ActionButton trichotomy) inside the persistent wrapper without destroying it. The `#compliance-tile` lives inside this shell **and** is re-rendered OOB by 2.12; keep the in-shell tile and the OOB tile **byte-equivalent** (no drift) so the single round trip paints consistently. (The earlier side-session draft used `hx-swap="outerHTML"` on the shell; this note adopts the `innerHTML` reconciliation ratified in Story 2.12 Decision 1.)

**Given** the project id does not exist or is not a valid UUID
**Then** the response is **HTTP 404** with no entity-state leakage (no project fields, no "exists but forbidden" signal — FR56).

### AC2 — Summary tab content (default panel)

**Given** the Summary tab is active (default)
**When** `#project-detail-tab-content` renders (inline on first load, and via `GET /projects/<id>/tabs/summary`)
**Then** it shows, using existing Basecoat surfaces (no new component invented):
- **code**, **name**, **start date**, **target completion date** (`—` when null — UX-DR17 empty distinction), **description** (`—` when null/blank);
- **Assigned trades** — the `name` of each `TradeType` referenced by `domain.project_trade_scope` for this project, resolved through the Story 2.3 reference cache; rendered as a list; create guarantees ≥1 trade so the empty case is not expected, but if zero rows render "No trades assigned";
- **Assigned inspectors** — the **display name** of each `user_id` in `domain.project_inspector`, resolved per stack via the same auth source the create form's inspector selector used (Django: `DevUserUuid.uuid → auth_user`, full name or username; .NET: `AuthDbContext` users by `Guid` id; Go: `fiber_auth.users` by id). When the project has **no** inspectors (inspectors are optional at create), render the empty state "No inspectors assigned" (cat 9);
- the **ActionButton row** (AC5).

**And** **no** "Project Manager" field is rendered (Decision 3).
**And** all project string fields (`code`, `name`, `description`) and resolved trade/inspector names render through each engine's **default HTML escaping** — no `Html.Raw` / `|safe` / `template.HTML` (per each stack's component rules); a per-stack XSS round-trip test (AC8) proves a `<script>` payload in `name`/`description` is escaped.

### AC3 — Tab switching: HTMX swap of the panel only, with OOB tabstrip update

**Given** I click the **Inspections**, **Violations**, or **Audit** tab
**When** HTMX fires `hx-get="/projects/<id>/tabs/<name>"` with `hx-target="#project-detail-tab-content"` `hx-swap="innerHTML"`
**Then** the response replaces **only** `#project-detail-tab-content` — the header strip, ComplianceTile, EntityRail, and `#flash-region` are **unaffected** (UX-DR24)
**And** the response **also** emits an out-of-band `<nav id="project-detail-tabstrip" hx-swap-oob="outerHTML" …>` re-rendering the full tablist with the clicked tab's `aria-selected="true"` (and `tabindex="0"`) and all others `false`/`-1` (TabStrip README §9)
**And** the swapped-in panel root carries `role="tabpanel"` + `aria-labelledby="<active-tab-id>"` (UX-DR33), and **focus moves to that panel root** after the swap (UX-DR31 tab-content convention — the panel has `tabindex="-1"`; move focus via `autofocus` on the inserted root or an `HX-Trigger`-driven `.focus()`; pick one mechanism and use it identically across stacks).

**Given** the Inspections / Violations / Audit panels
**Then** each renders a correctly-labelled `role="tabpanel"` placeholder (a single InlineAlert-styled `role="status"` message, e.g. "Inspections appear here once the inspection workflow lands.") — real content is Epic 3 / Epic 4 / Story 2.13 (out of scope).

### AC4 — EntityRail co-presence, survival across tab swaps, and responsive collapse

**Given** the page renders at **≥ 1280px**
**Then** the `project-detail-grid` container places `#project-detail-tab-content` and the `#violation-detail` rail **side by side**, the rail **sticky** on the right (~one-third width) (UX-DR18); the rail shows the EntityRail **empty-violation** variant ("Select an entity to see its detail here.").

**Given** the page renders at **< 1280px**
**Then** the rail **un-fixes and stacks below** the tab content (UX-DR30). The responsive rule lives in `fieldmark_shared/src/_layout.css` under `/* EntityRail responsive collapse */` (EntityRail README §8); the EntityRail component's own collapse rule already exists — this story adds the **grid-container** layout (`project-detail-grid`) that positions the two regions.

**Given** I switch tabs (AC3)
**Then** the `#violation-detail` rail's content is **not** cleared (it is a sibling of the swap target, not a descendant — Decision 5).

### AC5 — Action-button row uses the affordance trichotomy (UX-DR10/21)

**Given** the Summary tab renders the Place on Hold / Resume / Close row via the **ActionButton** component (`_ActionButton.cshtml` / `_action_button.html` / `action_button.html`)
**When** the caller supplies pre-computed `permission` (from `can(actor, "project.place_on_hold"|"project.resume"|"project.close")`) and `state_allows` (from the Project entity's `CanPlaceOnHold()` / `CanResume()` / `CanClose()` predicate — Decision 4)
**Then** the component renders the trichotomy itself (callers never branch): **absent** when `!permission`; **disabled** (with `data-tooltip`/`aria-describedby` reason) when `permission && !state_allows`; **present** (primary button, `hx-post`, `hx-target="#project-detail"`, `hx-swap="outerHTML"`, `hx-disabled-elt="this"`) when both true.
- Button ids: `place-on-hold-btn`, `resume-btn`, `close-btn`; `hx-post` targets `/projects/<id>/place-on-hold`, `/resume`, `/close` (handlers land in 2.12/Epic 6 — Decision 4).

**Given** I am **Executive** (FR43)
**When** I view any tab on `/projects/<id>`
**Then** **no** action buttons render anywhere (Executive has `project.read` but none of the three write actions → trichotomy collapses to **absent**) (UX-DR21).

**Given** the three actions are registered
**Then** `project.place_on_hold`, `project.resume`, `project.close` resolve to `ADMIN` only this story; a per-stack test asserts ADMIN → permitted, Executive/Inspector/SiteSupervisor/ComplianceOfficer → not permitted.

### AC6 — Authorization & the read-handler shape

**Given** the page is a **pure read**
**Then** the handler authorizes (`can(actor, "project.read")`) → loads the aggregate → renders. **No** `transaction.atomic` / `IDbContextTransaction` / `pgx.Tx`, and **no** `AuditEntry` write (reads are not audited — same discipline as Story 2.10, opposite of Story 2.8). [architecture.md read-handler shape](../planning-artifacts/architecture.md).

**Given** a user **without** `project.read`
**When** they request `GET /projects/<id>` or any tab endpoint
**Then** **HTTP 403** without entity-state leakage (FR7, FR56). (Executive and all five seeded roles **have** `project.read` per Story 2.9; the no-role `testuser` fixture → 403.)

**Given** an **unauthenticated** request to `GET /projects/<id>` or a tab endpoint
**Then** the Story 1.11 redirect-to-login fires first (302/303 → `/login`), unchanged.

### AC7 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

**Given category 1 (unknown vocabulary)**
**Then** `project.status` is DB-CHECK-constrained to `{Active, OnHold, Closed}` so StatusBadge always hits a known case; the StatusBadge `badge-unknown` fallback (from 2.4) is the safety net — no new handling needed.

**Given category 3 (JS fails / disabled)**
**Then** with JS off: the header strip, ComplianceTile, StatusBadge, the **server-rendered Summary panel** (Decision 6), and the empty EntityRail all render and are readable; the TabStrip tabs are visible `<button>`s (tab *switching* and arrow-key nav require JS — acceptable degradation, the default Summary content is already present). A `javaScriptEnabled:false` test asserts the header strip + Summary content render and the page is navigable.

**Given category 9 (empty / whitespace values)**
**Then** `target_completion_date` null → `—`; `description` null/blank → `—`; zero inspectors → "No inspectors assigned"; these are tested. (Date/score are never user-free-text; no derived-initials helper here.)

**Given category 6 (text overflow & special characters)**
**Then** long project `name`/`code`/`description` and trade/inspector names render through default escaping and the existing Basecoat truncation utilities; the EntityRail `entity-rail__entity-type` already carries `truncate`. No raw-entity rendering.

**Given categories 2 (fonts), 5 (stacking), 7 (reduced motion), 8 (forced colors)**
**Then** covered by Story 1.14 global rules (no new fonts; no unbounded stacking — the rail is a single region; reduced-motion + forced-colors handled in `_a11y.css`; StatusBadge/ComplianceTile pair color with text). An **axe-core** scan on `/projects/<id>` (Summary active) reports **zero** new WCAG 2.1 AA violations, including correct tablist/tabpanel/region roles.

### AC8 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

**Given category 3a (XSS round-trip on render)**
**Then** for the Summary tab, a per-stack test passes the bare payload `<script>alert(1)</script>` as the project **`name`** and **`description`**, renders the panel, and asserts both `Contains("&lt;script&gt;alert(1)&lt;/script&gt;")` **and** `NotContains("<script>")` for **each** field (both assertions required).

**Given category 1 (open redirect)**
**Then** N/A — no return-target parameter on this screen; all routes are server-fixed.

**Given category 6 (CSRF) & the read-only nature**
**Then** `GET /projects/<id>` and the four tab endpoints are safe methods needing no CSRF token. The action-button `hx-post`s carry each stack's existing CSRF posture (Story 1.6/1.11) but their **handlers do not exist yet** (2.12) — no CSRF surface is introduced by this story.

**Given categories 2, 4, 5, 7**
**Then** N/A — no new cookies, no dynamic regex, no filesystem writes, no new identity handling.

### AC9 — `make parity` and full gate

**Given** Story 1.3 route-parity tooling
**When** I run `make parity`
**Then** all three stacks' route dumps contain `GET /projects/:id`, `GET /projects/:id/tabs/summary`, `/tabs/inspections`, `/tabs/violations`, `/tabs/audit` (stack-idiomatic param syntax), diff clean; `pg_indexes` diff is **zero** (no schema change).

**Build/type/lint/test gates green per stack:**
- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean.
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean.
- **`fieldmark_shared`:** `pnpm run build` — `dist/fieldmark.css` regenerated and committed **only** for the new `project-detail-grid` rule in `src/_layout.css`; no other CSS drift.
- **E2E:** a Project Detail scenario (login → open `/projects/<id>` → assert header strip + four tabs + Summary content + empty rail → click Violations → assert panel swaps + tabstrip `aria-selected` flips + header/rail unchanged → as Executive assert zero action buttons).

### AC10 — Cross-stack contract (per the Cross-Stack Architecture Principle)

This story introduces a new cross-stack **orchestration contract** (route scheme + dual-mode response + tab-swap composition), so it ships the three required deliverables:

**Given** the contract doc
**When** I inspect `docs/reference/project-detail-contract.md`
**Then** it specifies: the route scheme (`GET /projects/:id` + `GET /projects/:id/tabs/{summary,inspections,violations,audit}`); the **dual-mode** rule (HX-Request → body partial; else full page — Decision 1); the canonical ids and their roles (the `project-detail` **shell wrapper** — the main-detail-panel / 2.12 whole-panel re-render target, present in both render modes, with `compliance-tile` nested inside it and kept byte-equivalent to any OOB tile re-render; `project-header-strip`, `compliance-tile`, `project-detail-tabstrip`, `project-detail-tab-content`, `violation-detail`); the tab order (Summary/Inspections/Violations/Audit); the **tab-swap response composition** (panel inner HTML targeting `#project-detail-tab-content` **+** OOB `#project-detail-tabstrip` with flipped `aria-selected`); the focus-move convention; and the 404 / 403 / unauthenticated behaviors.

**Given** each stack implements the contract natively
**Then** Razor (.NET), Django templates, Go `html/template` — no shared template fragment, no symlinked partial; the composition is re-implemented idiomatically per stack.

**Given** each stack's test suite
**When** I run the per-stack Project-Detail conformance test
**Then** it asserts: (1) `GET /projects/:id` with `HX-Request: true` returns the body partial (no `<html>`/`<head>`), without it returns the full page; (2) `GET /projects/:id/tabs/violations` returns a `role="tabpanel"` panel **and** an `hx-swap-oob` `#project-detail-tabstrip` with `aria-selected="true"` on `tab-violations`; (3) the canonical ids + tab order are present and identical across stacks.

---

## Tasks / Subtasks

- [x] **Task 1: Shared CSS — `project-detail-grid` responsive container** (AC: #4)
  - [x] 1.1 In `fieldmark_shared/src/_layout.css`, under a `/* EntityRail responsive collapse */` (or `/* Project Detail layout */`) comment, add `.project-detail-grid`: at `≥1280px` a two-column grid (content ~2/3, rail ~1/3) with the `#violation-detail` rail `position: sticky; top: …`; below `1280px` single-column (rail stacks under content). Reuse the existing EntityRail collapse rule from Story 2.6; this adds only the **container** layout. Keep it minimal — prefer Tailwind-compatible utilities where they exist; only hand-author what utilities can't express.
  - [x] 1.2 `cd fieldmark_shared && pnpm run build`; commit `src/_layout.css` + regenerated `dist/fieldmark.css`. Verify no unrelated CSS drift.

- [x] **Task 2: Project entity — read-only `can_*` status predicates** (AC: #5)
  - [x] 2.1 .NET `FieldMark.Domain/Entities/Project.cs`: `public bool CanPlaceOnHold() => Status == ProjectStatus.Active;` `CanResume() => Status == ProjectStatus.OnHold;` `CanClose() => Status == ProjectStatus.Active;` (pure; no state mutation; XML-doc that the Epic-6 closure gate is **not** evaluated here).
  - [x] 2.2 Django `projects/models.py` `Project`: `def can_place_on_hold(self) -> bool: return self.status == ProjectStatus.ACTIVE` (+ `can_resume`, `can_close`).
  - [x] 2.3 Go `internal/domain/entities/project.go`: `func (p Project) CanPlaceOnHold() bool { … }` (+ `CanResume`, `CanClose`). (These are the first behavior methods on the Go Project field-bag — see the package comment.)
  - [x] 2.4 Domain unit tests per stack: each predicate true/false across `{Active, OnHold, Closed}`.

- [x] **Task 3: Register the three write actions** (AC: #5)
  - [x] 3.1 .NET: `DomainPolicies.RegisterAction("project.place_on_hold", Role.Admin)` (+ resume, close) — at startup alongside 2.8/2.9 registrations (`Program.cs` or a `ProjectPolicies.Register()`).
  - [x] 3.2 Django: `register_action("project.place_on_hold", Role.ADMIN)` (+ resume, close) — module-level in `projects/views.py` next to `project.create`.
  - [x] 3.3 Go: `auth.RegisterAction("project.place_on_hold", domain.RoleAdmin)` (+ resume, close) — composition time (handler `init()` or `main.go`).

- [x] **Task 4: `GET /projects/<id>` dual-mode handler + page composition** (AC: #1, #2, #4, #6, #10)
  - [x] 4.1 Build the **detail view model** per stack: load Project + trade-scope ids + inspector ids (Go `ProjectStore.LoadWithRelations`; .NET query `Projects` + `ProjectTradeScopes`/`ProjectInspectors` by id; Django `Project.objects.get` + `.trade_scopes`/`.inspector_assignments`). Resolve trade **names** via the Story 2.3 reference cache; resolve inspector **display names** via each stack's auth source (Decision 3 / AC2). Compute the three `permission` booleans (`can(...)`) and three `state_allows` booleans (entity predicates) for the ActionButton row. 404 on missing/invalid id; 403 on `!can(project.read)`.
  - [x] 4.2 Compose the body partial: header strip (StatusBadge + ComplianceTile `#compliance-tile`), `<div class="tabs">` (TabStrip `#project-detail-tabstrip` active_index=0 + `#project-detail-tab-content` with Summary pre-rendered), `<aside id="violation-detail">` empty EntityRail, all inside `.project-detail-grid`.
  - [x] 4.3 Dual-mode switch (Decision 1): `HX-Request` → return body partial only; else wrap in base chrome. Same body markup both ways.
  - [x] 4.4 .NET `Pages/Projects/Detail.cshtml(.cs)` (replace stub) — note the stub's `@page "/projects/{id:guid}"` returns 404-via-`NotFound()` on parse-miss already; preserve. Django `projects/views.py` `project_detail` (rename/replace `project_detail_stub`) + `projects/templates`. Go `projects_detail_handler.go` (replace stub) + `pages/projects_detail.html` + tab fragments.

- [x] **Task 5: Tab-content endpoints + OOB tabstrip** (AC: #2, #3, #10)
  - [x] 5.1 Add routes `GET /projects/<id>/tabs/{summary,inspections,violations,audit}` per stack (Django `urls.py`, Go `main.go`, .NET — a second handler/page or a single handler switching on a route segment; pick the idiomatic option and keep parity-route names).
  - [x] 5.2 Each endpoint: authorize `project.read` (403) / 404 on bad id; render the **panel inner HTML** (`role="tabpanel"` + `aria-labelledby="<tab-id>"`, `tabindex="-1"`) **plus** an OOB `<nav id="project-detail-tabstrip" hx-swap-oob="outerHTML">` with the correct `active_index` (TabStrip §9). Summary panel = AC2 content; the other three = labelled placeholder panels (scope boundary).
  - [x] 5.3 Focus move to panel root after swap (UX-DR31) — one mechanism, identical across stacks.

- [x] **Task 6: `docs/reference/project-detail-contract.md` + per-stack conformance tests** (AC: #10)
  - [x] 6.1 Author the contract doc (route scheme, dual-mode, canonical ids, tab order, tab-swap composition, focus convention, 404/403/unauth).
  - [x] 6.2 Per-stack conformance test (the three assertions in AC10).

- [x] **Task 7: Per-stack page/handler tests** (AC: #1, #2, #5, #6, #7, #8)
  - [x] 7.1 200 render (authorized) asserting all canonical ids + tab order + StatusBadge + `#compliance-tile` once; 404 (bad id); 403 (no-role user); unauthenticated → login redirect.
  - [x] 7.2 Trichotomy: ADMIN on Active → present hold/close + disabled resume; ADMIN on OnHold → present resume + disabled hold/close; Executive → **zero** action buttons (assert absence of all three button ids).
  - [ ] 7.3 Edge cases (AC7): null target_completion_date/description → `—`; zero inspectors → "No inspectors assigned"; `javaScriptEnabled:false` (E2E) header+Summary render.
  - [x] 7.4 XSS round-trip (AC8): `<script>` in `name` and `description`, both assertions, both fields.
  - [x] 7.5 Action registration: ADMIN permitted, other four roles not, for all three actions.

- [ ] **Task 8: E2E + parity + gate** (AC: #9)
  - [ ] 8.1 E2E Project Detail scenario (full-page + tab swap + OOB tabstrip + Executive no-buttons + rail survives tab swap).
  - [ ] 8.2 `make parity` (five routes present, `pg_indexes` zero-diff) + `make test-all` green.

- [x] **Task 9: Story sign-off** (AC: all)
  - [x] 9.1 Populate the Sign-off block; record the six decisions; flip sprint-status to `review`.

## Dev Notes

### Critical context (read before writing code)

- **This is a read screen — no transaction, no audit.** Authorize → load → render. The mutating flow (tx + audit + three-region OOB) is Story 2.12. See Story 2.10 dev notes for the same read-handler discipline.
- **Dual-mode is non-negotiable (Decision 1).** Story 2.9's `/projects` list page is live and its row-click (`detail` mode) swaps `GET /projects/<id>` into `<aside id="project-detail">`. If you return a full `<html>` document there, the list page breaks (a whole page nested in a div). Detect `HX-Request` and return the body partial. Verify by actually exercising the `/projects` list row-click, not just the standalone page.
- **Reuse the four components — do not invent markup.** StatusBadge (2.4), ComplianceTile (2.5, `id="compliance-tile"`), TabStrip (2.7, `id="project-detail-tabstrip"`), EntityRail (2.6, `id="violation-detail"`, empty variant). ActionButton (1.12) for the action row. Each has a canonical contract README under `fieldmark_shared/components/`; the snapshot tests for those components already exist — your job is to **compose** them, passing the documented props. Do not re-emit their inner markup by hand.
- **TabStrip needs the `.tabs` outer wrapper from *you* (the consumer).** Basecoat 0.3.11's `.tabs` class wraps **both** the `[role="tablist"]` and the `[role="tabpanel"]` via descendant selectors. The TabStrip wrapper renders only the `<nav>`; you must wrap the TabStrip + `#project-detail-tab-content` in `<div class="tabs">` (TabStrip README §7).
- **OOB tabstrip on every tab response (Decision 2, TabStrip §9).** A tab click swaps `#project-detail-tab-content` (the `hx-target`) and **must also** return `<nav id="project-detail-tabstrip" hx-swap-oob="outerHTML" …>` with the new active tab, so `aria-selected` and the roving `tabindex` stay correct. Re-render the TabStrip wrapper with the new `active_index` and add the OOB attribute on its root.
- **tabstrip.js is already loaded** in all three base layouts (`<script src="/vendor/tabstrip/tabstrip.js" defer>`); it attaches to `nav[data-tabstrip]` and re-attaches on `htmx:after:swap` (HTMX 4.0 colon event name — see `fieldmark_shared/CLAUDE.md`). You do **not** add or modify JS. The OOB swap that replaces the `<nav>` triggers the re-attachment automatically.
- **EntityRail rail id = `violation-detail`, empty variant (Decision 5).** Render `entity_loaded=false`. It sits as a **sibling** of `#project-detail-tab-content` in `.project-detail-grid` so tab swaps (which target the panel) never touch it (UX-DR24, EntityRail README §8 layout).
- **Action row (Decision 4) — register actions + add entity predicates; do NOT build POST handlers.** The buttons' `hx-post` targets are the 2.12/Epic-6 endpoints and will 404 until then — that is the sanctioned deferral. The Executive "no buttons" AC is satisfied purely by Executive lacking the three write permissions (absent branch of the trichotomy). The disabled-reason strings are user-visible: "Project is already on hold" / "Project is not on hold" / "Only active projects can be closed".
- **No PM field (Decision 3).** The grid contract already dropped `pm_name`; the Summary tab must not reintroduce a Project-Manager concept. Inspectors are the only people-link.
- **Inspector display-name resolution crosses the auth boundary per stack.** Mirror the create form: Django `_get_reference_data` joins `DevUserUuid → auth_user`; .NET reads `AuthDbContext` Identity users by `Guid` id; Go reads `fiber_auth.users` by id. The `domain.project_inspector.user_id` holds the **canonical UUID**, which is the Identity/Go user id and the Django `DevUserUuid.uuid`. Keep this read out of any transaction.
- **Tab route scheme `/projects/:id/tabs/<name>` (Decision 2).** The TabStrip canonical fixture's `/projects/__ID__/summary` is a placeholder and does not constrain you. Use `/tabs/` to match the epic parity AC and the contract doc.

### Source tree — where things land

| Stack | Detail page + tabs | Entity predicates | Action registration |
|---|---|---|---|
| .NET | `FieldMark.Web/Pages/Projects/Detail.cshtml(.cs)` (replace stub) + tab handler(s) | `FieldMark.Domain/Entities/Project.cs` | `Program.cs` / `ProjectPolicies` |
| Django | `projects/views.py` (`project_detail` replacing `project_detail_stub`; tab views) + `projects/urls.py` + `templates/projects/detail.html` + tab partials | `projects/models.py` `Project` | module-level in `projects/views.py` |
| Go | `internal/web/handlers/projects_detail_handler.go` (replace stub) + tab handler + routes in `cmd/web/main.go` + `templates/pages/projects_detail.html` + tab fragments | `internal/domain/entities/project.go` | handler `init()` / `main.go` |

Shared: `fieldmark_shared/src/_layout.css` (+ rebuilt `dist/fieldmark.css`). New doc: `docs/reference/project-detail-contract.md`.

### Existing code to reuse (read before writing)

- **Component wrappers + READMEs:** StatusBadge ([_StatusBadge.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/Components/_StatusBadge.cshtml), [_status_badge.html](../../fieldmark_py/templates/components/_status_badge.html), [status_badge.html](../../fieldmark-go/internal/web/templates/components/status_badge.html)); ComplianceTile, TabStrip, EntityRail wrappers (paths above) + READMEs in `fieldmark_shared/components/`.
- **ActionButton:** [_ActionButton.cshtml](../../FieldMark/FieldMark.Web/Pages/Shared/_ActionButton.cshtml) + [ActionButtonVm](../../FieldMark/FieldMark.Web/ViewModels/Components/ActionButtonVm.cs); Go [action_button.go](../../fieldmark-go/internal/web/viewmodels/action_button.go) + `action_button.html`; Django `_action_button.html`. Trichotomy lives **inside** the component.
- **Data layer:** Go [projectstore.go](../../fieldmark-go/internal/data/postgres/projectstore.go) `LoadWithRelations`; .NET `FieldMarkDbContext` DbSets; Django related managers ([models.py](../../fieldmark_py/projects/models.py)).
- **Reference cache (trade names):** Story 2.3 (`reference/` per stack).
- **Auth lookup (inspector names):** create-form pattern — Django [`_get_reference_data`](../../fieldmark_py/projects/views.py); .NET `AuthDbContext`; Go `fiber_auth.users`.
- **`can()` / RegisterAction:** Story 1.12 (`DomainPolicies` / `fieldmark.authz` / `auth.Can`); `project.read` registered in Story 2.9.
- **Existing stubs to replace:** [Detail.cshtml(.cs)](../../FieldMark/FieldMark.Web/Pages/Projects/Detail.cshtml.cs), [project_detail_stub](../../fieldmark_py/projects/views.py), [projects_detail_handler.go](../../fieldmark-go/internal/web/handlers/projects_detail_handler.go) (+ their templates).

### Project Structure Notes

- Adds 5 routes to the parity inventory; replaces 3 stub detail routes/handlers; no `domain.*` schema change (`pg_indexes` zero-diff).
- One shared CSS rule (`project-detail-grid`) → `src/_layout.css` + rebuilt `dist/fieldmark.css`. No new vendor JS.
- New cross-stack contract doc (`docs/reference/project-detail-contract.md`) — first contract for the tab-swap orchestration; 2.12 extends it for the mutating three-region flow (which has its own `docs/how-to/three-region-oob-orchestration.md`).
- Canonical HTMX target ids introduced/used: `project-header-strip`, `compliance-tile` (existing, OOB-capable), `project-detail-tabstrip`, `project-detail-tab-content`, `violation-detail` (existing). Note these in the root inventory if the parity tooling tracks ids.

### References

- Epic AC: [epic-2 §Story 2.11](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md)
- DDL: [010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) (project / project_trade_scope / project_inspector)
- Component contracts: [entity_rail/README.md](../../fieldmark_shared/components/entity_rail/README.md) (§8 layout, §9 focus), [tab_strip/README.md](../../fieldmark_shared/components/tab_strip/README.md) (§7 `.tabs` wrapper, §9 OOB), [compliance_tile/README.md](../../fieldmark_shared/components/compliance_tile/README.md), [status_badge/README.md](../../fieldmark_shared/components/status_badge/README.md)
- Grid contract (no PM, dual-mode rail target `#project-detail`): [ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md)
- Read-handler shape (no tx, no audit): [architecture.md](../planning-artifacts/architecture.md); prior read story [2-10](2-10-compliance-dashboard-with-portfolio-tiles.md)
- Edge cases / security: [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md) cat 1/3/6/9, [security-defaults.md](../../docs/reference/security-defaults.md) cat 3a
- Prior component stories: [2-6 EntityRail](2-6-implement-entityrail-component-with-responsive-collapse.md), [2-7 TabStrip](2-7-implement-tabstrip-component-with-arrow-key-navigation.md), [2-5 ComplianceTile](2-5-implement-compliancetile-component-and-compliance-tile-oob-target.md), [2-9 list/rail](2-9-project-list-ag-grid-with-server-side-row-model.md)

## Dev Agent Record

### Agent Model Used

GPT-5 Codex

### Debug Log References

- `dotnet test FieldMark.Tests.Domain --filter ProjectActionPredicateTests` (pass)
- `uv run pytest projects/tests/test_project_action_predicates.py` (pass; required escalated run for uv cache access)
- `go test ./internal/domain/entities -run ProjectActionPredicates` (pass; required escalated run for Go build cache access)
- `make parity` (pass: routes + pg_indexes parity clean)
- `make test-all` (pass: .NET + Django + Go unit/integration gates)
- `make e2e` (fails in this environment: Playwright Chromium launch permission denied; not an app assertion failure)
- `dotnet test FieldMark/FieldMark.Tests.Web/FieldMark.Tests.Web.csproj --filter "ProjectsDetailPageTests|CanTests"` (pass)
- `make test-net` (pass)
- `make test-django` (pass; 305 passed, 9 skipped integration-conditional tests)
- `make test-go` (pass)

### Completion Notes List

- Completed Task 1: added shared `.project-detail-grid` layout rule and rebuilt `fieldmark_shared/dist/fieldmark.css`.
- Completed Task 2: added status-only action predicates to Project entities in .NET, Django, and Go.
- Completed Task 3: registered `project.place_on_hold`, `project.resume`, and `project.close` for ADMIN in all three stacks.
- Added per-stack predicate tests covering `{Active, OnHold, Closed}` state matrix.
- Implemented Project Detail dual-mode page in all three stacks with canonical ids and layout composition.
- Implemented `/projects/:id/tabs/:tab` endpoints in all three stacks, including OOB tabstrip updates and placeholder panels.
- Added `docs/reference/project-detail-contract.md` and verified route/index parity with `make parity`.
- Verified stack gates: `make test-net`, `make test-django`, `make test-go`.
- Fixed .NET test harness insert helper to avoid `DBNull` parameter mapping in `ExecuteSqlRawAsync` for nullable columns.
- Added cross-stack Project Detail conformance tests:
  - .NET: `FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`
  - Django: `fieldmark_py/projects/tests/test_project_detail.py` (integration-marked, skips when domain schema absent on default DB)
  - Go: `fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`

### File List

- _bmad-output/implementation-artifacts/sprint-status.yaml
- fieldmark_shared/src/_layout.css
- fieldmark_shared/dist/fieldmark.css
- FieldMark/FieldMark.Domain/Entities/Project.cs
- FieldMark/FieldMark.Web/Program.cs
- FieldMark/FieldMark.Tests.Domain/Entities/ProjectActionPredicateTests.cs
- fieldmark_py/projects/models.py
- fieldmark_py/projects/views.py
- fieldmark_py/projects/tests/test_project_action_predicates.py
- fieldmark-go/internal/domain/entities/project.go
- fieldmark-go/internal/domain/entities/project_actions_test.go
- fieldmark-go/cmd/web/main.go
- FieldMark/FieldMark.Web/Pages/Projects/Detail.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/Detail.cshtml.cs
- FieldMark/FieldMark.Web/Pages/Projects/_DetailBody.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/Tabs/_SummaryPanel.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/Tabs/_InspectionsPanel.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/Tabs/_ViolationsPanel.cshtml
- FieldMark/FieldMark.Web/Pages/Projects/Tabs/_AuditPanel.cshtml
- FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml
- FieldMark/FieldMark.Web/Program.cs
- fieldmark_py/projects/urls.py
- fieldmark_py/projects/views.py
- fieldmark_py/templates/projects/detail.html
- fieldmark_py/templates/projects/_detail_body.html
- fieldmark_py/templates/projects/_tab_response.html
- fieldmark_py/templates/projects/tabs/_summary_panel.html
- fieldmark_py/templates/projects/tabs/_placeholder_inspections.html
- fieldmark_py/templates/projects/tabs/_placeholder_violations.html
- fieldmark_py/templates/projects/tabs/_placeholder_audit.html
- fieldmark_py/templates/components/_tab_strip.html
- fieldmark-go/internal/web/handlers/projects_detail_handler.go
- fieldmark-go/internal/web/templates/pages/projects_detail.html
- fieldmark-go/internal/web/templates/pages/projects_detail_body.html
- fieldmark-go/internal/web/templates/pages/projects_detail_panels.html
- fieldmark-go/internal/web/templates/pages/projects_detail_tab_response.html
- fieldmark-go/internal/web/templates/components/tab_strip.go
- fieldmark-go/internal/web/templates/components/tab_strip.html
- docs/reference/project-detail-contract.md
### Review Findings

Round 1 — 2026-06-01

**Decision-needed (1):**

- [x] [Review][Patch] Go `StatusBadgeVM`: implement `ResolveStatusBadge(entity, value string) StatusBadgeVM` helper — align Go with the Entity+Value+Severity delegated-resolution pattern used by .NET and Django. Add a `ResolveStatusBadge` func in the viewmodels package (mirrors `StatusBadgeVM.Severity` deferred-work entry from Story 2-4); update the detail handler to call it instead of the inline `if/else` chain. Close the deferred-work entry. [`fieldmark-go/internal/web/viewmodels/components.go`, `fieldmark-go/internal/web/handlers/projects_detail_handler.go`]

**Patch items (18):**

- [x] [Review][Patch] `id="project-detail"` shell wrapper missing in all three body partials — `_DetailBody.cshtml`, `_detail_body.html`, and `projects_detail_body.html` all open directly with `<section id="project-header-strip">` — no outer element carries `id="project-detail"`. AC1's reviewer note is explicit: the body partial inner content must be wrapped in a stable element with `id="project-detail"` in both render modes, because Story 2.12 targets `#project-detail` with `hx-swap="innerHTML"`. The Story 2.9 list rail already has `<aside id="project-detail">` and is the target for the HTMX partial mode; the standalone full page needs the same shell for 2.12's re-render. Add `<div id="project-detail">…</div>` wrapper around body content in all three partials. [AC1, AC10]
- [x] [Review][Patch] Django whitespace-only description renders raw spaces, not em-dash [`fieldmark_py/templates/projects/tabs/_summary_panel.html:7`] — `{% if project.description %}` is truthy for `"   "` (whitespace-only). The Django CLAUDE.md explicitly warns about this falsy-check gap. .NET uses `string.IsNullOrWhiteSpace`; Go uses `strings.TrimSpace(...) != ""`. Fix: `{% if project.description and project.description.strip %}`. [AC2, AC7 cat-9]
- [x] [Review][Patch] Go and Django tab handlers return bare fragment on non-HTMX direct navigation — `_render_tab_response` (Django) has no `HX-Request` check; Go's `GetProjectsDetail` falls through to full-page render with Summary panel hardcoded, ignoring the requested tab. .NET correctly redirects to `/projects/:id`. Hard-rule requires stack symmetry on HTMX behaviors. Fix: add `if not request.headers.get("HX-Request"): return redirect(...)` guard in Django; add `if c.Get("HX-Request") != "true" && isTab { return c.Redirect(...) }` guard in Go tab path. [AC3, AC10]
- [x] [Review][Patch] ADMIN-on-Closed all-buttons-disabled test absent in all three stacks — Decision 4 states "ADMIN on a Closed project → all three disabled." No test in any stack verifies this. The domain-predicate tests cover the Closed state, but no integration-level test verifies that all three buttons render as *disabled* (not absent) when status is Closed. [AC5]
- [x] [Review][Patch] Django missing trichotomy tests (Admin-Active, Admin-OnHold, Closed, non-admin) [`fieldmark_py/projects/tests/test_project_detail.py`] — the Django conformance test has no trichotomy coverage at all. Hard-rule sibling component test parity requires all three stacks to have equivalent special-purpose test patterns. [AC5]
- [x] [Review][Patch] Go missing trichotomy tests (Admin-Active, Admin-OnHold, Closed, Executive no-buttons) [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`] — same parity gap as Django. [AC5]
- [x] [Review][Patch] Django missing XSS round-trip test for `name` and `description` [`fieldmark_py/projects/tests/test_project_detail.py`] — AC8 explicitly requires a per-stack test with bare `<script>alert(1)</script>` payload, asserting escaped form present AND raw form absent, for both fields. [AC8, hard-rule sibling parity]
- [x] [Review][Patch] Go missing XSS round-trip test for `name` and `description` [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`] — same gap as Django. [AC8, hard-rule sibling parity]
- [x] [Review][Patch] 403 (no-role user) conformance test absent in all three stacks — Task 7.1 explicitly requires a 403 test. No stack's conformance test asserts that a user without `project.read` receives HTTP 403. [AC6, AC10]
- [x] [Review][Patch] Contract doc missing `project-detail` shell wrapper id [`docs/reference/project-detail-contract.md`] — AC10 requires the contract to document the `project-detail` shell wrapper as the main-detail-panel / 2.12 whole-panel re-render target, present in both render modes. The contract's Canonical IDs list omits it. [AC10]
- [x] [Review][Patch] Contract doc missing non-HTMX tab URL behavior [`docs/reference/project-detail-contract.md`] — The contract specifies dual-mode for `GET /projects/:id` but is silent on what `GET /projects/:id/tabs/:tab` returns without `HX-Request`. This is the root cause of the cross-stack divergence in P3. Document expected behavior (redirect to `/projects/:id`). [AC10, AC3]
- [x] [Review][Patch] Contract doc missing Story 2.12 re-render target + compliance-tile byte-equivalence note [`docs/reference/project-detail-contract.md`] — AC10 requires documenting that Story 2.12 will target `#project-detail` with `hx-swap="innerHTML"` and that `#compliance-tile` inside the shell must remain byte-equivalent to any OOB tile re-render. [AC10]
- [x] [Review][Patch] .NET `AdminActive` test has duplicate `resume-btn` assertion; resume disabled-state never verified [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs:258-259`] — both assertions check `Contains("id=\"resume-btn\"")`. The second should verify `resume-btn` is in the *disabled* state (e.g., assert `aria-describedby="resume-btn-reason"` is present). [AC5]
- [x] [Review][Patch] .NET `claimMap` empty-string fallback produces blank inspector name [`FieldMark/FieldMark.Web/Pages/Projects/Detail.cshtml.cs:80`] — `g => g.First().ClaimValue ?? ""` coalesces null to `""`, but `claimMap.GetValueOrDefault(u.Id, u.UserName ?? ...)` only falls back when the key is absent, not when it maps to `""`. A `display_name` claim with empty value renders the inspector as a blank string. Fix: `claimMap.TryGetValue(u.Id, out var d) && !string.IsNullOrWhiteSpace(d) ? d : u.UserName ?? u.Id.ToString()`. [AC2]
- [x] [Review][Patch] Go test setup missing `resetForTests()` before `RegisterAction` [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`] — `makeProjectsDetailApp` registers actions into global mutable state without resetting first. `authz_test.go` calls `resetForTests()` explicitly; absence here causes stale-state failures under `-count=2` or parallel runs.
- [x] [Review][Patch] .NET `CanTests.cs` registers into global `DomainPolicies` without reset [`FieldMark/FieldMark.Tests.Web/Authorization/CanTests.cs`] — `ProjectActions_AreAdminOnly` calls `RegisterAction` inside the test body with no reset fixture. Flaky under non-deterministic ordering.
- [x] [Review][Patch] .NET whitespace-only description test case missing [`FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs`] — the empty-fields test uses `description: null` (NULL in DB). Cat-9 requires testing all three boundary cases: null, empty string, AND whitespace-only (`"   "`). Add a second row with `description: "   "` and assert it renders as `—`. [AC7 cat-9]
- [x] [Review][Patch] .NET `CreateProjectRowAsync` test helper injects `""` instead of `NULL` for null description — when `description` is null, the helper writes `description ?? string.Empty` to the SQL parameter, storing `""` instead of `NULL`. The handler renders `—` for both (via `IsNullOrWhiteSpace`), so the test passes, but the DB receives an empty string, violating null semantics on a nullable column. Fix: pass `(object?)null` when `description` is null.

**Deferred (7):**

- [x] [Review][Defer] `IsTabResponse` flag set before tab validation in .NET — future fragility, no current defect; `NotFound()` correctly fires before `Page()` in all current code paths [FieldMark/FieldMark.Web/Pages/Projects/Detail.cshtml.cs] — deferred, pre-existing
- [x] [Review][Defer] Go nil-pool `loadInspectorNames` silent swallow — pre-existing nil-pool pattern throughout Go test suite; inspector names absent with no error if Pool is nil at runtime [fieldmark-go/internal/web/handlers/projects_detail_handler.go] — deferred, pre-existing
- [x] [Review][Defer] Go `projects_detail_tab_response.html` has no fallback for unknown `PanelTemplate` value — handler switch guards it; low risk today [fieldmark-go/internal/web/templates/pages/projects_detail_tab_response.html] — deferred, pre-existing
- [x] [Review][Defer] .NET `ProjectActionPredicateTests` uses reflection to set `Status` — fragile to future encapsulation tightening; no current defect [FieldMark/FieldMark.Tests.Domain/Entities/ProjectActionPredicateTests.cs] — deferred, pre-existing
- [x] [Review][Defer] `compliance-tile` present-once assertion not tested — low risk on this read page; AC1 states it must appear exactly once [FieldMark/FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs] — deferred, pre-existing
- [x] [Review][Defer] `javaScriptEnabled:false` E2E test absent (Task 7.3 open) — Playwright Chromium environment constraint; no CI lane guarantee yet [AC7 cat-3] — deferred, pre-existing
- [x] [Review][Defer] E2E scenario + `make parity` unverified (Task 8 open) — Playwright environment constraint in implementation environment [AC9] — deferred, pre-existing

Round 2 — 2026-06-01

**Decision-needed (1):**

- [x] [Review][Patch] `tradeNameById` Active-only filter silently drops inactive trades — show deactivated trades with an "(inactive)" suffix instead. In all three stacks: build the trade-name lookup from ALL trade types (remove the `Where(t => t.Active)` / `filter(active=True)` guard); then when resolving each assigned trade's display name, append `" (inactive)"` if `!t.Active` / `not t.active`. This way a deactivated trade still appears in the list (it IS assigned) with a clear "(inactive)" label, and "No trades assigned" only renders when the project genuinely has zero `domain.project_trade_scope` rows. [AC2] [FieldMark.Web/Pages/Projects/Detail.cshtml.cs, fieldmark_py/projects/views.py, fieldmark-go/internal/web/handlers/projects_detail_handler.go]

**Patch items (4):**

- [x] [Review][Patch] Django `_render_tab_response` authz check fires after the HX-Request redirect — an authenticated user without `project.read` hitting `/projects/:id/tabs/:tab` without `HX-Request` gets a **302 redirect** to `/projects/:id` instead of **403**, leaking that the project ID exists. Move the `can(request.user, "project.read")` guard to execute before the `if request.headers.get("HX-Request") != "true"` redirect. [AC6, security-defaults §project existence leakage] [`fieldmark_py/projects/views.py:_render_tab_response`]
- [x] [Review][Patch] Go XSS test (`TestGetProjectsDetail_XssPayloadEscaped`) mutates package-level stub vars without cleanup — `stubProjectName` and `stubProjectDescription` (and possibly `stubProjectTargetDate`) are set to the XSS payload and left dirty; subsequent tests in the same binary inherit them. Add `t.Cleanup(func() { stubProjectName = "Project Detail Go"; stubProjectDescription = nil; stubProjectTargetDate = nil })` at the start of the test. [`fieldmark-go/internal/web/handlers/projects_detail_handler_test.go`]
- [x] [Review][Patch] Go unguarded `m["IsTabResponse"].(bool)` type assertion panics if key absent — the single-value form panics if `buildVM` ever returns a map without the `IsTabResponse` key. Use the two-value form: `isTab, _ := m["IsTabResponse"].(bool)`. [`fieldmark-go/internal/web/handlers/projects_detail_handler.go`]
- [x] [Review][Patch] `_TabStrip.cshtml` bare `catch` on `HxSwapOob` silently swallows typos and binding errors — a typo in the prop name (e.g. `HXSwapOob`) produces a `RuntimeBinderException` that is caught and discarded; the OOB attribute is simply absent and the tabstrip is never updated, with no diagnostic signal. Replace with a null-safe check: `var hxSwapOob = Model.GetType() == typeof(ExpandoObject) ? ... : null` or use `HasProperty(Model, "HxSwapOob")` before accessing. [`FieldMark/FieldMark.Web/Pages/Shared/Components/_TabStrip.cshtml`]

**Deferred (4):**

- [x] [Review][Defer] Inspector silent drop when `domain.project_inspector.user_id` has no matching row in the auth schema (cross-stack) — deleted users leave orphaned FK rows; the inspector is silently omitted with no fallback display or operator log. Pre-existing cross-stack gap; address when user-lifecycle management is in scope. [`FieldMark.Web/Pages/Projects/Detail.cshtml.cs:89`, `fieldmark_py/projects/views.py`, `fieldmark-go/internal/web/handlers/projects_detail_handler.go`] — deferred, pre-existing
- [x] [Review][Defer] `_SummaryPanel` has no `#project-action-form` slot — Story 2.12 Decision 1 designates this as Task 0 in that story; the omission is epic-sanctioned. — deferred, Story 2.12 Task 0
- [x] [Review][Defer] `HtmxMode` test ARIA assertion coverage weak — verifies `id="violation-detail"` presence but not `role="region"` or `aria-live` attributes on the rail, nor absence of `hx-swap-oob` on the main response. Low severity; address when next touching the conformance tests. [FieldMark.Tests.Web/Pages/ProjectsDetailPageTests.cs] — deferred, pre-existing
- [x] [Review][Defer] `autofocus` on OOB-swapped panel may not fire in all HTMX 4.0 / browser combinations — spec-compliant approach; no known failure in HTMX 4.0-beta2 + Chromium. If cross-browser issues arise, add `HX-Trigger: {"focusPanel": true}` fallback per UX-DR31. — deferred, monitor

Round 3 — 2026-06-01

**Patch items (1):**

- [x] [Review][Patch] Go unguarded `m["Tabs"].(components.TabStripArgs)` type assertion panics if key absent — the R3 fix applied the two-value form to `m["IsTabResponse"]` but left `m["Tabs"]` as a single-value assertion. Use `tabs, ok := m["Tabs"].(components.TabStripArgs); if !ok { return c.SendStatus(fiber.StatusInternalServerError) }`. [`fieldmark-go/internal/web/handlers/projects_detail_handler.go:2094`]

**Deferred (8):**

- [x] [Review][Defer] Go status-mutation tests lack `t.Cleanup` on `stubProjectStatus` — sequential execution safe; `makeProjectsDetailApp` resets stubs on entry; fragile to future reordering [projects_detail_handler_test.go] — deferred, pre-existing
- [x] [Review][Defer] `.NET` handler builds full Summary VM before non-HTMX tab redirect — wasted DB work thrown away on redirect path; no correctness defect [Detail.cshtml.cs] — deferred, pre-existing
- [x] [Review][Defer] All three stacks return custom text body on 403 — AC6 specifies HTTP 403 status only; body shape is not AC-required; consistent with prior stories — deferred, pre-existing
- [x] [Review][Defer] Django `{% include panel_template %}` variable path — value always hardcoded in the view; add a comment if refactored [fieldmark_py/templates/projects/_tab_response.html] — deferred, pre-existing
- [x] [Review][Defer] Django `prefetch_related` bypassed by `.values_list()` — two extra DB queries per request; fix by iterating `.all()` to use prefetch cache [fieldmark_py/projects/views.py:_build_project_detail_context] — deferred, pre-existing
- [x] [Review][Defer] No test for no-role user + non-HTMX GET on tab URL — authz-before-redirect fix (R2 P1) correct but untested for the combined unauthorized+non-HX path [all three stacks] — deferred, pre-existing
- [x] [Review][Defer] Go `buildVM` nil project not guarded after `LoadWithRelations` — nil+nil-error store return panics; store contract prevents this today [projects_detail_handler.go] — deferred, pre-existing
- [x] [Review][Defer] `.NET` `_DetailBody.cshtml` hardcodes `ActiveIndex = 0` — correct per Decision 6; footgun if non-Summary initial active tab is ever needed — deferred, pre-existing

## Sign-off

- Date of final review: 2026-06-01
- Total review-round count: 3
- Final reviewer verdict (PASS/FAIL): PASS
- Deferred-work entries created from this story: none during implementation (known pre-existing e2e environment permission issue remains)
- Decisions requiring ratification (confirm or overturn at review):
  1. **`GET /projects/<id>` is dual-mode** (HX-Request → body partial for the Story 2.9 `#project-detail` rail; else full page). _pending_
  2. **Tab route scheme `/projects/:id/tabs/<name>`** + each tab response = panel inner HTML + OOB `#project-detail-tabstrip`. _pending_
  3. **No "PM name" field** — consistent with the ratified ag-grid-ssrm-contract `pm_name` drop. _pending_
  4. **Action-button row rendered via trichotomy; transitions deferred** — register `project.place_on_hold`/`resume`/`close` (ADMIN only) + add read-only entity `can_*` predicates; `hx-post` targets 404 until 2.12/Epic 6. _pending_
  5. **EntityRail rail id = `violation-detail`, empty variant** (per EntityRail README §8). _pending_
  6. **Default tab = Summary, server-rendered inline on first load.** _pending_
