# Epic 2: Project Lifecycle & Compliance Dashboard

Establish every cross-cutting pattern (canonical request flow, audit-in-same-transaction, AG Grid SSRM, three-region OOB orchestration, EntityRail + TabStrip layout, Phase-2 components) on the Project aggregate. After this epic an Admin/PM can create projects, the Compliance Dashboard renders the portfolio, the Project Detail anchor screen is live, and place-on-hold/resume transitions exercise the three-region pattern.

## Story 2.1: Map `domain.project` and supporting tables into each stack's data layer

As a developer building Project-related features in any stack,
I want each stack's data layer to read and write `domain.project`, `domain.job_site`, `domain.project_trade_scope`, and `domain.project_inspector` against the existing canonical DDL,
So that subsequent stories can implement Project behavior without inventing schema.

**Acceptance Criteria:**

**Given** the .NET stack
**When** I inspect `FieldMark.Data/Configuration/`
**Then** `ProjectConfiguration.cs`, `JobSiteConfiguration.cs`, `ProjectTradeScopeConfiguration.cs`, `ProjectInspectorConfiguration.cs` use `ToTable("<table>", "domain")` and the snake_case naming convention
**And** enum columns use `HasConversion<string>()` for the `SCREAMING_SNAKE_CASE` storage convention (per `domain-model.md` §9).

**Given** the Django stack
**When** I inspect `projects/models.py`
**Then** `Project`, `JobSite`, `ProjectTradeScope`, `ProjectInspector` declare `Meta.managed = False` and `db_table = 'domain"."<table>'`
**And** field types match the canonical DDL exactly.

**Given** the Go stack
**When** I inspect `internal/data/`
**Then** `ProjectStore` is a narrow interface with read methods only at this story (no writes yet) backed by a `pgx`-using implementation in `internal/data/projectstore.go`
**And** column lists in SQL match the canonical DDL.

**Given** all three mappings exist
**When** I run `make parity`
**Then** `pg_indexes` for `domain.*` shows zero diff against the canonical inventory.

**Given** each stack's domain unit-test project
**When** I run `make test-net`, `make test-django`, `make test-go`
**Then** a smoke test exists per stack that loads a seeded Project by ID and asserts every column maps round-trip.

---

## Story 2.2: Map `domain.audit_entry` and provide a per-stack `append_audit_entry()` helper

As a handler author across all three stacks,
I want a single helper that appends an AuditEntry within the current DB transaction,
So that FR39 (audit-on-every-mutation) is mechanically satisfied for every transition Epic 2+ introduces.

**Acceptance Criteria:**

**Given** each stack's data layer
**When** I inspect it
**Then** `domain.audit_entry` is mapped: .NET via `AuditEntryConfiguration.cs`; Django via `audit/models.py` with `Meta.managed = False`; Go via `AuditEntryStore` interface + pgx implementation in `internal/data/auditentrystore.go`.

**Given** any mutating handler
**When** it calls `append_audit_entry(actor_id, action, entity_type, entity_id, project_id?, before_state, after_state, metadata?)` (or the per-stack idiomatic equivalent)
**Then** the helper writes the row using the same `DbContext` / connection / `pgx.Tx` as the surrounding transaction (FR39, FR57)
**And** `before_state` / `after_state` are serialized as JSONB
**And** `action` is stored verbatim from a canonical-string enum (no inventing variants — FR40).

**Given** a handler aborts (rule violation / exception)
**When** the transaction rolls back
**Then** no audit entry is left orphaned (verified by a per-stack integration test against Testcontainers / pytest-django / Go integration build tag).

**Given** the canonical action string list from CLAUDE.md
**When** I inspect each stack's enum/constants
**Then** the same 14 strings are present with the same SCREAMING/PascalCase casing
**And** the cross-stack diff for audit action constants is clean.

**Given** the principle that cross-stack invariants live as documentation contracts (Epic 1 retro)
**When** I inspect `docs/reference/audit-actions.md`
**Then** the file exists and lists every canonical audit action string as the single source of truth
**And** each stack's native enum/constants module references the doc URL in a top-of-file comment
**And** the doc is the only place the list is authored — no shared code package, no symlinked manifest.

**Given** each stack's test suite
**When** I run a per-stack audit-action conformance test
**Then** the test reads the documented list (parsed from `docs/reference/audit-actions.md` or a checked-in fixture derived from it) and asserts the stack's native enum/constants set matches exactly — no extras, no missing entries.

---

## Story 2.3: Map reference data tables and expose a read API per stack

As an Administrator,
I want to view the catalog of Trade Types, Violation Categories, and Compliance Rules,
So that I can verify the canonical reference data is loaded and visible (FR52, FR53).

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its `reference` module (.NET `Domain/Entities/Reference/` + `Data/Configuration/Reference*.cs`; Django `reference/` app; Go `internal/domain/reference.go` + `internal/data/referencestore.go`)
**Then** `TradeType`, `ViolationCategory`, `ComplianceRule` map to their `domain.*` tables read-only
**And** rows are loaded on first request (not at process start) and cached in process memory with a TTL or invalidation hook
**And** no application restart is required to pick up changes (FR53 — verified by `UPDATE domain.compliance_rule …` then re-fetching).

**Given** I am authenticated as Administrator
**When** I navigate to `/admin/reference`
**Then** three sections render: Trade Types, Violation Categories, Compliance Rules
**And** each is a Basecoat Table (not AG Grid — UX-DR carve-out) showing all rows
**And** no Create/Edit/Delete affordances are rendered (FR67 is Growth).

**Given** I am authenticated as any non-Administrator role
**When** I navigate to `/admin/reference`
**Then** I receive HTTP 403 without entity-state leakage (FR56).

---

## Story 2.4: Implement Phase-2 markup-only components — StatusBadge, InlineAlert, AuditRow, DashboardTile

As a developer rendering Project Detail or the Compliance Dashboard,
I want four small Basecoat-compliant wrapper templates per stack with byte-identical output,
So that I can compose subsequent stories without inventing markup per screen.

**Acceptance Criteria:**

**Given** each stack's `partials/components/` directory
**When** I inspect it
**Then** wrappers for **StatusBadge** (UX-DR9), **InlineAlert** (UX-DR13), **AuditRow** (UX-DR12), and **DashboardTile** (UX-DR17) exist as markup-only templates with no logic.

**Given** the StatusBadge wrapper
**When** I render it with each entity-state combination from the Step 8 vocabulary (Project, Inspection, Violation+severity, CorrectiveAction)
**Then** the produced markup matches the canonical example in `fieldmark_shared/components/status_badge/` byte-for-byte across all three stacks
**And** color is never the sole information carrier (text label always present).

**Given** the InlineAlert wrapper
**When** I render it with `severity ∈ {danger, warning, info, success}`
**Then** danger/warning render `role="alert"` and info/success render `role="status"` (UX-DR13)
**And** the icon is Lucide-sourced and paired with text.

**Given** the AuditRow wrapper
**When** I render it with an action + actor + timestamp + before/after JSON
**Then** it lives inside an `aria-live="polite"` parent
**And** the disclosure for before/after JSON has `aria-expanded` and uses JetBrains Mono for the snippet (UX-DR12).

**Given** the DashboardTile wrapper
**When** I render it with a label + value + optional secondary text
**Then** it renders as a Basecoat Card with uppercase label + `text-3xl font-bold tnum` value (UX-DR17)
**And** when `role="status"` is set it announces value changes politely.

**Given** the canonical example gallery at `fieldmark_shared/components/`
**When** I run the per-stack snapshot tests
**Then** each stack's wrapper produces output byte-identical to the canonical example (UX-DR40 initial coverage).

**Given** the principle that cross-stack invariants live as documentation contracts (Epic 1 retro)
**When** I inspect each component's directory under `fieldmark_shared/components/` (e.g., `status_badge/`, `inline_alert/`, `audit_row/`, `dashboard_tile/`)
**Then** each directory contains a `canonical.html` example *and* a `README.md` describing the contract: required props, ARIA attributes, allowed class vocabulary, and the snapshot-equality requirement
**And** the per-stack snapshot test references the canonical file by path (no copy-paste of the expected markup into test source).

**Given** the components are markup-only wrappers
**When** I inspect each stack's wrapper implementation
**Then** the wrapper lives natively in that stack's idiomatic component location (.NET `Pages/Shared/Components/`, Django `templates/components/`, Go `internal/web/templates/components/`) — no shared template engine, no symlinked partial.

---

## Story 2.5: Implement ComplianceTile component and `#compliance-tile` OOB target

As a user of any FieldMark screen showing project compliance,
I want a tile that renders the project's compliance score with threshold-based color and announces changes politely to assistive technology,
So that the canonical anchor demo's OOB-update mechanism has its target in place.

**Acceptance Criteria:**

**Given** each stack's component wrapper
**When** I inspect it (UX-DR11)
**Then** ComplianceTile renders as `<section id="compliance-tile" role="status" aria-live="polite" aria-atomic="true">` containing an uppercase label, a `text-3xl font-bold tnum` numeric score, the semantic threshold word ("Healthy"/"Watch"/"Concern"/"Critical"), and the semantic color (UX-DR4 thresholds).

**Given** scores 0–100
**When** the tile renders each
**Then** the band-to-color mapping is exactly: ≥90 success, 70–89 warning (lighter), 50–69 warning (darker), <50 danger
**And** color is paired with both the numeric value and the threshold word.

**Given** a Project Detail or Compliance Dashboard render
**When** the page is served
**Then** `#compliance-tile` appears once per page (header strip on Project Detail; first dashboard tile on Compliance Dashboard rendered with id `#compliance-tile-portfolio`)
**And** Story 2.12 / 2.13's OOB swap can target `#compliance-tile` without inventing markup.

**Given** the three stacks
**When** I render ComplianceTile with identical inputs
**Then** the produced HTML is byte-identical (verified by snapshot against `fieldmark_shared/components/compliance_tile/`).

---

## Story 2.6: Implement EntityRail component with responsive collapse

As a Compliance Officer or Project Manager working on the Project Detail screen at desktop,
I want a sticky right-rail container that holds the currently selected entity's detail and collapses to stacked layout below tablet,
So that List+Detail co-presence works at desktop and falls back gracefully on smaller viewports.

**Acceptance Criteria:**

**Given** each stack's wrapper (UX-DR18)
**When** I inspect it
**Then** EntityRail renders as `<aside id="<entity>-detail" tabindex="-1" role="region" aria-label="<entity-type> detail">` with three slots: header strip (entity-type label + dismiss ×), body, action footer.

**Given** no entity is selected
**When** the rail renders
**Then** the empty state reads "Select an entity to see its detail here." with Basecoat Card styling and an `aria-label="Empty entity rail"`.

**Given** the viewport
**When** width ≥ 1280px
**Then** the rail is sticky on the right at the third of the content area
**And** at 768–1279px the rail un-fixes and renders stacked below the list
**And** at <768px the same stacked behavior holds (UX-DR30).

**Given** a row is selected in any list within Project Detail
**When** the HTMX request lands a partial into the rail
**Then** focus moves to the rail's root via `tabindex="-1"` and either `HX-Trigger`-driven focus script or autofocus (UX-DR31 primary-swap convention).

**Given** I switch tabs on Project Detail
**When** the tab content swaps
**Then** the rail's content is not cleared by the tab swap (rail is independent of tab content per UX-DR24).

---

## Story 2.7: Implement TabStrip component with arrow-key navigation

As any user of Project Detail,
I want a horizontal tab strip whose tabs swap content via HTMX with full ARIA tablist semantics,
So that I can move between Summary / Inspections / Violations / Audit without page navigation and with keyboard support.

**Acceptance Criteria:**

**Given** each stack's wrapper (UX-DR16)
**When** I inspect it
**Then** TabStrip renders as `<nav role="tablist">` containing `<button role="tab" aria-selected="<bool>" aria-controls="<panel-id>" hx-get="<url>" hx-target="<panel-id>" hx-swap="innerHTML">` per tab
**And** the corresponding tab-panel container carries `role="tabpanel"` + `aria-labelledby="<tab-id>"` (UX-DR33 strict heading + landmark rules).

**Given** a tab is active
**When** I inspect its rendering
**Then** `aria-selected="true"` is set, the underline indicator is applied, and the tab is bold per Basecoat
**And** after a tab click the active tab updates `aria-selected` via OOB swap returned in the same response.

**Given** focus is on a tab
**When** I press Left/Right arrow keys
**Then** focus cycles between tabs without activating them; Enter or Space activates the focused tab
**And** the keyboard JS is ≤ 15 LOC and vendored as `tabstrip.js` (UX-DR15-style budget reflected in Step 11 components budget).

**Given** the three stacks
**When** I render the TabStrip with identical inputs
**Then** the HTML output is byte-identical.

---

## Story 2.8: Project create form (PM/Admin)

As a Project Manager or Administrator,
I want to create a new Project with required metadata,
So that the application has Projects to manage, inspect, and report on.

**Acceptance Criteria:**

**Given** I am authenticated as `ADMIN` or another role with `project.create` permission (initially `ADMIN`)
**When** I navigate to `/projects/new`
**Then** a form renders using Basecoat inputs with `<label>` associations capturing: code (unique), name, start date, optional target completion date, trade scope assignments (multi-select from `domain.trade_type`), inspector assignments (multi-select from seeded inspector users) (FR9).

**Given** the form
**When** I submit invalid input (missing required fields, duplicate code, end-date before start-date)
**Then** the form re-renders with HTTP `422` containing a top InlineAlert with `role="alert"` and `aria-invalid="true"` + `aria-describedby` on each invalid field (UX-DR34, FR61).

**Given** the form
**When** I submit valid input
**Then** a single transaction performs: load reference rows → call `Project.create(...)` entity method → write `Project` to `domain.project` → write `project_trade_scope` and `project_inspector` rows → append `AuditEntry(action="ProjectCreated", actor=me, before=null, after=<project state>)` → commit (FR57)
**And** I am redirected via `HX-Redirect` (or framework-equivalent) to `/projects/<id>`.

**Given** I am not authorized to create projects
**When** I navigate to `/projects/new`
**Then** the link is `absent` from the Compliance Dashboard (FR6) and a direct request returns HTTP `403` without state leakage (FR7).

**Given** I am authenticated as any authorized role
**When** the route `POST /projects/` is called with GET method
**Then** the response is HTTP `405` (or framework-equivalent rejecting GET on mutating routes) (FR54).

**Given** `make parity`
**When** I run it after this story
**Then** the route inventory contains `GET /projects/new` and `POST /projects/` on all three stacks with diff clean.

Note: `ProjectCreated` has been added to the canonical audit action string list (now 15 strings total) by ADR amendment ratified during epic planning. The story may emit it directly.

---

## Story 2.9: Project list AG Grid with server-side row model

As a Project Manager, Compliance Officer, or Administrator,
I want to view the Projects list with server-side filtering, sorting, and pagination,
So that I can navigate the portfolio without client-side compute (FR48–FR51).

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its grid handler
**Then** `POST /grid/projects` accepts an AG Grid SSRM request payload (filterModel, sortModel, startRow, endRow) and returns JSON `{ "rows": [...], "lastRow": N }` with `snake_case` keys (FR49, Architecture D10).

**Given** the response rows
**When** I inspect each
**Then** they contain `id`, `code`, `name`, `status`, `compliance_score`, `start_date`, `target_completion_date`, `pm_name` — projected manually from `domain.project` join `domain.user`-equivalent (NFR6 — no AutoMapper).

**Given** the Project list page at `/projects`
**When** it renders
**Then** the AGGridPanel wrapper (UX-DR19) initializes AG Grid in SSRM mode pointing at `POST /grid/projects`
**And** filtering by status and sorting by `compliance_score` work end-to-end against the server (FR47, FR51 — no client-side filter).

**Given** I click a row
**When** the row-click handler fires
**Then** `htmx.ajax("GET", "/projects/<id>", { target: "#project-detail" })` runs and loads Project Detail (FR50)
**And** the grid does not render the detail itself (FR51).

**Given** the empty state (no projects)
**When** the grid renders
**Then** a custom no-rows overlay reads "No projects yet — create one to get started" with an ActionButton present for users with `project.create` permission, absent otherwise (UX-DR26).

**Given** `make parity`
**When** I run it
**Then** `POST /grid/projects` exists on all three stacks identically.

**Given** the principle that cross-stack invariants live as documentation contracts (Epic 1 retro)
**When** I inspect `docs/reference/ag-grid-ssrm-contract.md`
**Then** the file exists and specifies the SSRM wire format exhaustively: request shape (`filterModel`, `sortModel`, `startRow`, `endRow`, allowed filter operators), response shape (`{ "rows": [...], "lastRow": N }`), snake_case key convention, row-projection rules (no AutoMapper), and error behaviour for invalid inputs
**And** each stack's grid handler implements the contract natively — direct EF Core projection (.NET), direct ORM/raw SQL projection (Django), direct `pgx` query (Go) — with no shared codec or generated stubs.

**Given** each stack's test suite
**When** I run a per-stack SSRM conformance test
**Then** the test issues a canonical SSRM request (fixture loaded from the doc or a derived JSON) against `POST /grid/projects` and asserts the response shape, key casing, and lastRow semantics match the documented contract exactly.

---

## Story 2.10: Compliance Dashboard with portfolio tiles

As any authorized user landing on FieldMark,
I want a Compliance Dashboard showing portfolio aggregates and a Projects list,
So that I can see "where is risk?" at a glance and drill into a Project from there (FR44, FR46).

**Acceptance Criteria:**

**Given** I am authenticated and have `dashboard.view` permission
**When** I navigate to `/dashboard` (and the empty Home from Story 1.13 redirects here once Story 2.10 lands)
**Then** the page renders four tiles at the top: ComplianceTile (`#compliance-tile-portfolio`) showing the portfolio average compliance score, a DashboardTile for "Overdue Violations" broken down by severity sub-badges, a DashboardTile for "Active Projects" count, and a DashboardTile for "Inspections This Week" count (FR44).

**Given** the tile row
**When** the viewport collapses
**Then** at 1024–1279px the row reflows to 2×2; at <768px to single column (UX-DR30).

**Given** the tile row
**When** any tile's underlying value changes from a same-page action
**Then** the affected tile is OOB-swappable (`id` stable; `role="status"` set on tiles that update) — wiring of actual change events is deferred to the epics that introduce them (UX-DR45 not violated because tiles render correctly at every static state).

**Given** the dashboard
**When** the tile row is rendered
**Then** below it the Projects AG Grid (Story 2.9) fills the remaining viewport.

**Given** any tile, including the ComplianceTile, renders
**When** there is no data (no projects, no inspections)
**Then** the tile renders `—` rather than `0` to distinguish "empty" from "zero" (UX-DR17).

**Given** `make parity`
**When** I run it
**Then** `GET /dashboard` is present on all three stacks.

---

## Story 2.11: Project Detail anchor screen with header strip, TabStrip, and EntityRail

As any authorized user,
I want to view a Project's current truth on a single page with HTMX-driven tabs and a co-present detail rail,
So that I can read the status, score, inspections, violations, and audit log without navigation (FR11).

**Acceptance Criteria:**

**Given** I navigate to `/projects/<id>`
**When** the page renders
**Then** the page has: a header strip with the project's code + name + StatusBadge for current status + ComplianceTile (`#compliance-tile`, FR36); a TabStrip with tabs Summary / Inspections / Violations / Audit; a content region `#project-detail-tab-content` (the tab panel target); a sticky `EntityRail` on the right at ≥ 1280px (UX-DR18, UX-DR24).

**Given** the Summary tab is active (default)
**When** the content region renders
**Then** it shows the project's code, name, start date, target completion date, PM name, list of assigned trades, list of assigned inspectors, and a "Place on Hold" / "Resume" / "Close" ActionButton row using the affordance trichotomy (UX-DR10) — the actual transitions land in Stories 2.12 and Epic 6.

**Given** I click another tab
**When** HTMX fires `hx-get` for that tab's content endpoint
**Then** only `#project-detail-tab-content` swaps; the header strip, ComplianceTile, EntityRail, and FlashRegion are unaffected (UX-DR24)
**And** focus moves to the new tab-panel root (UX-DR31 tab-content convention).

**Given** I am Executive role (FR43)
**When** I navigate to `/projects/<id>`
**Then** no action buttons render anywhere on the page (UX-DR21 affordance trichotomy collapses to `absent` for all permission-gated actions).

**Given** the page at <1280px
**When** it renders
**Then** EntityRail un-fixes and stacks below the tab content (UX-DR30).

**Given** `make parity`
**When** I run it
**Then** `/projects/:id` route + tab endpoints (`/projects/:id/tabs/summary`, `/projects/:id/tabs/inspections`, `/projects/:id/tabs/violations`, `/projects/:id/tabs/audit`) exist on all three stacks.

---

## Story 2.12: Place-On-Hold and Resume transitions with three-region OOB orchestration

As a Project Manager,
I want to place an Active Project on hold and resume it back to Active, each with a recorded reason and an audit row visible immediately,
So that the canonical three-region orchestration pattern (UX-DR20) is verified before the anchor demo lands in Epic 5 (FR12, FR13).

**Acceptance Criteria:**

**Given** I am authorized (`project.place_on_hold` or `project.resume`)
**When** I click the corresponding ActionButton on Project Detail
**Then** an inline form expands requesting a reason; submission fires `POST /projects/<id>/place-on-hold` (or `/resume`) with the reason in the body (FR54 — POST only).

**Given** a valid request
**When** the handler runs the canonical flow
**Then** it: authorizes → begins transaction → loads `Project` aggregate → calls `project.place_on_hold(reason)` (or `.resume()`) → appends `AuditEntry(action="ProjectPlacedOnHold" or "ProjectResumed", actor, before, after, metadata={reason})` → commits (FR57, FR40)
**And** the response body contains the re-rendered `#project-detail` partial **plus** OOB `#compliance-tile` (re-rendered with the unchanged score — score isn't affected, but the OOB pattern is exercised) **plus** OOB `#audit-log` row prepended (UX-DR20, UX-DR23).

**Given** the request is unauthorized
**When** the handler resolves the authz check
**Then** HTTP `403` is returned without entity-state leakage (FR7, FR56)
**And** no `#compliance-tile` or `#audit-log` OOB updates are emitted (UX-DR22).

**Given** the entity method raises (e.g., already on hold)
**When** the exception bubbles to the handler
**Then** HTTP `409` is returned with the originating `#project-detail` partial re-rendered showing *current* state + inline InlineAlert (UX-DR22, FR55)
**And** `#compliance-tile` and `#audit-log` are **not** updated.

**Given** the project transitions
**When** I open the Audit tab
**Then** the new audit row is present at the top
**And** the action string is exactly `ProjectPlacedOnHold` or `ProjectResumed` (cross-stack-identical — Architecture canonical list).

**Given** an E2E Playwright scenario runs against all three stacks
**When** the place-on-hold action is exercised
**Then** all three stacks produce a single HTTP round trip (no follow-up requests) with the three-region updates in one paint
**And** the local-dev p95 timing is ≤ 200 ms per stack with cross-stack divergence ≤ 50 ms p95 (NFR1).

**Given** the principle that cross-stack patterns live as documentation contracts (Epic 1 retro)
**When** I inspect `docs/how-to/three-region-oob-orchestration.md`
**Then** the recipe exists with a worked example: when to use it (mutations that affect entity + header tile + audit log), the canonical response composition (main partial + OOB `#compliance-tile` + OOB `#audit-log` prepend), the negative cases (403 / 409 must NOT emit OOB swaps per UX-DR22), and the testing contract for response shape
**And** each stack's handler implements the recipe natively — Razor partial composition (.NET), Django template `include`/`extends` (Django), Go `html/template` blocks (Go) — with no shared template fragment or symlinked partial.

**Given** each stack's test suite
**When** I run a per-stack three-region conformance test
**Then** the test exercises a successful place-on-hold and asserts the response contains exactly the three documented regions (main `#project-detail`, OOB `#compliance-tile`, OOB `#audit-log`); a 403 response asserts zero OOB regions; a 409 response asserts the main partial re-renders with current state plus InlineAlert and zero OOB regions.

---

## Story 2.13: Project audit log tab

As any authorized user on Project Detail,
I want to view the project's full audit history most-recent-first,
So that I have forensic visibility into every domain mutation that touched the project (FR42, FR43).

**Acceptance Criteria:**

**Given** I navigate to the Audit tab on `/projects/<id>`
**When** the tab content swaps in
**Then** `#audit-log` renders as a list (Basecoat Table, not AG Grid — per UX-DR carve-out) of AuditRow components for every `domain.audit_entry` row where `project_id = <id>`, ordered `created_at DESC`.

**Given** I am Executive role (FR43)
**When** I view the Audit tab
**Then** every row renders as a read-only AuditRow with no action affordances (UX-DR21 collapses all actions to `absent`).

**Given** a single AuditEntry
**When** I expand its disclosure
**Then** the `before_state` and `after_state` JSONB columns render in JetBrains Mono with `aria-expanded` toggling correctly (UX-DR12).

**Given** the page first renders
**When** the audit log is initially fetched
**Then** at most 100 rows load on initial render (older rows behind a "Load more" affordance that fetches the next 100 via HTMX `hx-get` appending into `#audit-log`).

**Given** `make parity`
**When** I run it
**Then** the audit-log shape (HTML structure of AuditRow + the Load-more affordance) is byte-identical across stacks.

---

## Story 2.14: Reference data read pages for Administrator

As an Administrator,
I want to view the catalog of reference data,
So that I can confirm what's loaded without running SQL (FR52).

**Acceptance Criteria:**

**Given** Story 2.3 mapped the reference data
**When** I am authenticated as `ADMIN` and navigate to `/admin/reference/trade-types`, `/admin/reference/violation-categories`, `/admin/reference/compliance-rules`
**Then** each page renders a Basecoat Table listing rows with their canonical columns (e.g., `code`, `name`, `severity_weight` for compliance rules) — no CRUD affordances (FR67 is Growth, intentionally absent).

**Given** any non-Administrator
**When** they navigate to any `/admin/reference/*` route
**Then** HTTP `403` is returned (FR56) — links to these pages are absent from any rendered nav (FR6).

**Given** `make parity`
**When** I run it
**Then** the three reference routes exist on all three stacks and the diff is clean.

---

## Story 2.15: Harden Epic 2 — consolidated deferred-work pass

_Added 2026-06-03. The Epic 2 equivalent of Story 1.14 — one disciplined pass over the ~50 deferred items accumulated during Epic 2 (chiefly the 2.11/2.12 review marathons) plus the findings from manual AC testing after Story 2.13. Authored as a shell; manual-test findings are added before dev starts. Lands before the Epic 2 retrospective. Full detail: [2-15-harden-epic-2.md](../../implementation-artifacts/2-15-harden-epic-2.md)._

As the Project Lead closing out Epic 2,
I want every known deferred item and manual-AC-test finding resolved in one consolidated, cross-stack-symmetric pass,
So that Epic 3 begins on a clean baseline rather than dribbling Epic 2 debt across new stories.

**Acceptance Criteria (summary — see story file for the itemized D-* checklist):**

**Given** the deferred-work backlog for Epic 2 (Groups A–G in the story file: transaction/concurrency correctness, reason-handling consistency, cross-stack divergences, defensive guards, test hardening, dead-code/efficiency, parity robots/security symmetry)
**When** the hardening pass completes
**Then** every in-scope item is resolved with the fix mirrored across all three stacks unless explicitly stack-specific
**And** each resolved `deferred-work.md` entry is annotated "Resolved by Story 2.15."

**Given** the manual AC test pass run after Story 2.13
**When** findings are enumerated (Section H of the story file)
**Then** all are resolved under the same cross-stack discipline.

**Given** the robots.txt / security.txt parity exemption (Group G)
**When** the two routes are landed on .NET and the exemption filter is removed from `tools/parity/diff-routes.sh`
**Then** `make parity` is green **without** the exemption — now verifying all three stacks serve both routes.

**Given** all three stacks
**When** I run `make test-net`, `make test-django`, `make test-go`, and `make parity`
**Then** all gates are green.

**Out of scope** (own story / Epic 7 / tech-writer / Epic 3): concurrent-deletion-after-commit 404 (own cross-stack story), all E2E/JS-disabled items (Epic 7), Go nil-pool harness (verify retro A3), docs-governance trio (Paige / retro A2), inspector deleted-user fallback (resolve with Story 3.4a's `inspector_name` decision). See the story file's "Explicitly OUT of scope" section.
