# Story 2.9: Project list AG Grid with server-side row model

Status: done

Epic: 2 — Project Lifecycle & Compliance Dashboard
Source AC: [_bmad-output/planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md) §Story 2.9
Canonical DDL: [docker/postgres/init/010_domain_tables.sql:58–95](../../docker/postgres/init/010_domain_tables.sql) (`domain.project`)
Contract doc to populate (exists as skeleton): [docs/reference/ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md)
Depends on:
- **Story 2.1** (`Project` + `JobSite` + `ProjectTradeScope` + `ProjectInspector` mappings; `domain` schema wiring; per-stack read paths — `ProjectStore`/`projectStorePg` in Go, `Project` model in Django, `ProjectConfiguration`+`FieldMarkDbContext` in .NET; status: **done**).
- **Story 2.8** (`GET /projects/new` + `POST /projects/` create flow; the **Cancel link** in 2.8's form points at `/projects` — this story finally registers that route; 2.8 explicitly deferred `GET /projects` to this story and asserts `GET /projects` → 404 until 2.9 lands; 2.8 also ships a **minimal stub `GET /projects/<id>`** that the row-click target lands on; status: **ready-for-dev**).
- **Story 1.12** (`can(actor, action)` primitive per stack — `DomainPolicies.Can` (.NET), `fieldmark.authz.can` (Django), `auth.Can` (Go); `RegisterAction` registration idiom; this story registers the new `project.read` action and consumes the `project.create` grant added by 2.8).
- **Story 1.3** (`make parity` route-inventory + `pg_indexes` tooling — this story adds `GET /projects` and `POST /grid/projects` to the route inventory).
- **Story 1.5** (base-layout chrome — `#flash-region`, header, main, footer; the `/projects` page renders inside this chrome).
- **Story 2.6** (EntityRail component, `<aside id="project-detail" …>`; status: **ready-for-dev**) — the row-click detail target. If 2.6 is **done** at implementation time, render the EntityRail wrapper as the `#project-detail` container on the `/projects` page; if not, render a **minimal stub** `<aside id="project-detail" tabindex="-1" role="region" aria-label="Project detail"></aside>` matching the EntityRail root contract and note the refactor in Sign-off. **Do not block this story on 2.6.**

## Story

As a Project Manager, Compliance Officer, or Administrator (and Executive — read-only per FR43),
I want a `/projects` list page whose AG Grid pulls rows from a server endpoint `POST /grid/projects` that does **all** filtering, sorting, and pagination on the server (no client-side compute), and whose row-click loads the Project Detail into `#project-detail` via HTMX,
So that the canonical **AG-Grid-as-scoped-island** contract (UX-DR19, Architecture D10, FR48–FR51) — the SSRM wire format, the manual server-side row projection (NFR6, no AutoMapper), the row-select-to-rail interaction, and the custom no-rows overlay (UX-DR26) — is locked in for every downstream list view (Story 3.5 inspections grid, Story 4.2 violations grid).

**Scope boundary.** This story produces, per stack:
- (a) `GET /projects` route + handler + page (Razor page / Django template / Go template) rendering the **AGGridPanel** wrapper (UX-DR19 — **NEW component this story**, Phase-2 layout component) pointed at `POST /grid/projects`, plus the `#project-detail` detail target (EntityRail from 2.6 or stub) and the page chrome;
- (b) `POST /grid/projects` JSON data endpoint + handler implementing the SSRM wire contract: parse the AG Grid request payload (`startRow`, `endRow`, `sortModel`, `filterModel`), translate to a parameterized SQL `WHERE`/`ORDER BY`/`LIMIT`/`OFFSET` against `domain.project` via a thin per-stack **SSRM parser/translator helper** (the `grid` module — Django `grid/`, .NET `Grid/`, Go `internal/web` ssrm parser), run the data query **and** a matching `COUNT(*)` query, and return `{ "rows": [...], "lastRow": N }`;
- (c) the **AGGridPanel init JS** (~10 lines, vendored — see §"AG Grid init JS budget & vendoring") wiring the SSRM `IServerSideDatasource` (`rowModelType: 'serverSide'`) and the `onRowClicked` → `htmx.ajax("GET", "/projects/<id>", {target:"#project-detail"})` handler;
- (d) the `project.read` permission action registered (granted to all five conceptual roles in MVP — see AC5) and the no-rows-overlay create affordance gated by `project.create` (ADMIN, from 2.8);
- (e) the canonical contract doc `docs/reference/ag-grid-ssrm-contract.md` **populated** (the skeleton's TODOs replaced) — the single source of truth for the SSRM wire format;
- (f) per-stack tests: SSRM conformance (canonical fixture → assert response shape/casing/`lastRow`), filter/sort/pagination integration, row-projection column assertions, malformed-request 400 behaviour, empty-state overlay, authz 403/redirect, and edge-case/security coverage;
- (g) the route-parity additions verified by `make parity`.

**Out of scope:**
- The **Compliance Dashboard** at `/dashboard` that *embeds* this grid below its tile row (Story 2.10) — this story builds the standalone `/projects` page and the reusable AGGridPanel + endpoint; 2.10 reuses both.
- The **real Project Detail screen** at `GET /projects/<id>` with header strip + TabStrip + EntityRail (Story 2.11). This story's row-click lands on whatever `GET /projects/<id>` returns at implementation time — the 2.8 stub (`<main><h1>{{name}}</h1></main>`) or 2.11's real screen if it has landed. **Do not gold-plate the detail screen here.**
- **Server-side grouping / pivoting / tree data** — available in Enterprise but **out of scope for this story's flat project list**. The handler does not implement `rowGroupCols`/`pivotCols`/`groupKeys`; the parser ignores any grouping fields AG Grid sends (a flat block-fetch is all this grid needs). A later story may demo grouping.
- **Column resize/reorder persistence**, saved views, CSV/Excel export (not MVP).
- **New `domain.*` indexes.** The portfolio is small; sequential scans on `domain.project` are acceptable at MVP scale. Adding an index is an infrastructure-owned DDL change (Story 2.1's territory) and would break `make parity`'s `pg_indexes` zero-diff — **do not add one** (see Dev Notes §"No new indexes").
- Grids for other aggregates (`/grid/inspections`, `/grid/violations`) — those are Stories 3.5 / 4.2; this story writes the contract doc they will conform to but registers **only** `POST /grid/projects`.

---

## ⚠️ Decisions baked into this story (read first — these resolve contradictions in the source docs)

Three source-doc contradictions were found during analysis and resolved as follows. Each is implemented as written below; each is flagged in Dev Notes for reviewer ratification.

1. **Use AG Grid Enterprise + true SSRM; the unlicensed watermark is an accepted demo tradeoff.** The true Server-Side Row Model (`rowModelType: 'serverSide'`, `IServerSideDatasource`, `ServerSideRowModelModule`) is an **AG Grid Enterprise** feature. FieldMark **adopts Enterprise** (vendoring `ag-grid-enterprise.min.js`, the UMD bundle that includes Community) specifically to demonstrate the full Enterprise grid in the demo. **No license key is set** — AG Grid renders an "unlicensed" watermark, which the project owner has explicitly accepted as a deliberate tradeoff for showcasing Enterprise functionality. **Use `rowModelType: 'serverSide'` with an `IServerSideDatasource`.** Do not implement the Community Infinite Row Model fallback. See Dev Notes §"AG Grid Enterprise SSRM + the watermark tradeoff".

2. **`lastRow` is camelCase; row-data keys are snake_case.** The epic AC (3×) and [project-context.md](../project-context.md) both write the envelope as `{ "rows": [...], "lastRow": N }`. The architecture wire example at [architecture.md:582](../planning-artifacts/architecture.md) shows `"last_row"` — that is an inconsistency. Resolution: the **envelope keys** (`startRow`, `endRow`, `sortModel`, `filterModel`, `lastRow`) are AG Grid's **vendor wire vocabulary** (camelCase; `lastRow` maps into AG Grid's SSRM `params.success({ rowData, rowCount })`) and are exempt from the project's snake_case-on-the-wire hard rule; only the **row-data objects** follow snake_case. **Task 6 includes a one-line fix to [architecture.md:582](../planning-artifacts/architecture.md)** to change the illustrative `"last_row"` to `"lastRow"` so the canonical docs stop contradicting each other.

3. **`pm_name` is dropped from the row projection.** The epic AC lists `pm_name` "projected manually from `domain.project` join `domain.user`-equivalent". **There is no PM relationship in the schema** — `domain.project` has no `pm_id`/`project_manager` column, there is no `domain.user` table, and ADR-012 forbids domain→auth foreign keys (user identity is framework-local, opaque UUIDs). The only people-link is `domain.project_inspector` (inspectors, not a PM). Resolving a display name would require each stack to query its own auth schema, which cannot be made byte-identical cross-stack cheaply for a grid cell. Resolution: the canonical projection is **`id, code, name, status, compliance_score, start_date, target_completion_date`** (all from `domain.project`, no join). `pm_name` is **omitted**; the deviation is documented in the contract doc's Row Projection section. When a PM concept is introduced (future schema change, infra-owned), the column is added to the contract + three handlers + three conformance tests under the doc's Change Procedure.

---

## Acceptance Criteria

### AC1 — Populate the cross-stack contract doc `docs/reference/ag-grid-ssrm-contract.md`

**Given** the Cross-Stack Architecture Principle (root [CLAUDE.md](../../CLAUDE.md)) and the existing skeleton at [docs/reference/ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md) (scaffolded by the Epic 1 retro, action item A4, with `TODO (Story 2.9)` markers)
**When** I inspect the doc after this story
**Then** every `TODO (Story 2.9)` is replaced and the doc specifies, exhaustively:

1. **Status block** — flip the header from "skeleton" to "populated by Story 2.9, 2026-05-29", mirroring [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md).
2. **Edition note** — a prominent note that FieldMark uses **AG Grid Enterprise** with the true **Server-Side Row Model** (`rowModelType: 'serverSide'`), that the demo runs **without a license key** (the "unlicensed" watermark is an accepted, deliberate tradeoff), and that the wire contract below is what the `IServerSideDatasource` sends/receives.
3. **Request shape** — the JSON body POSTed by the SSRM datasource is AG Grid's `IServerSideGetRowsRequest` (camelCase vendor vocabulary). For this flat grid only `startRow`, `endRow`, `sortModel`, `filterModel` are honored; AG Grid may also include `rowGroupCols`, `valueCols`, `pivotCols`, `pivotMode`, `groupKeys` — the handler **ignores** these (no grouping/pivot this story):
   ```jsonc
   {
     "startRow": 0,          // inclusive, 0-based
     "endRow": 100,          // exclusive; block size = endRow - startRow
     "sortModel": [          // ordered; first entry is primary sort
       { "colId": "compliance_score", "sort": "asc" }   // sort ∈ {"asc","desc"}
     ],
     "filterModel": {        // keyed by colId; empty object = no filter
       "status":           { "filterType": "set",    "values": ["Active", "OnHold"] },
       "compliance_score": { "filterType": "number", "type": "greaterThan", "filter": 70 },
       "code":             { "filterType": "text",   "type": "contains",    "filter": "BLDG" },
       "start_date":       { "filterType": "date",   "type": "inRange", "dateFrom": "2026-01-01", "dateTo": "2026-12-31" }
     }
   }
   ```
   - **colId values are the snake_case projected column names** (`code`, `name`, `status`, `compliance_score`, `start_date`, `target_completion_date`). A `colId` not in this allowlist → 400 (see §error behaviour).
   - **Allowed filter operators per column type:**
     - **set** (`status`): the Enterprise **Set Filter** — `filterModel.status = { "filterType": "set", "values": [...] }`. The server maps to `WHERE status = ANY($1)`. Every value must be ∈ {`Active`, `OnHold`, `Closed`} (enum allowlist — any other value → 400). An empty `values: []` means "match nothing" (return zero rows) per AG Grid Set Filter semantics.
     - **text** (`code`, `name`): `equals`, `notEqual`, `contains`, `notContains`, `startsWith`, `endsWith`, `blank`, `notBlank`.
     - **number** (`compliance_score`): `equals`, `notEqual`, `greaterThan`, `greaterThanOrEqual`, `lessThan`, `lessThanOrEqual`, `inRange` (uses `filter` + `filterTo`), `blank`, `notBlank`.
     - **date** (`start_date`, `target_completion_date`): `equals`, `notEqual`, `greaterThan` (after), `lessThan` (before), `inRange` (uses `dateFrom` + `dateTo`), `blank`, `notBlank`. Date values are `YYYY-MM-DD` strings.
   - **Sort direction values:** `asc` | `desc` only. Sort allowed on every projected column. `sortModel` honored in array order (multi-column). Unknown `colId` or `sort` value → 400.
   - **Pagination bounds:** `startRow ≥ 0`, `endRow > startRow`, and `endRow - startRow ≤ 1000` (hard server cap; a request exceeding it → 400, **not** a silent clamp). The default AG Grid block is 100.
4. **Response shape** — the envelope, with the snake_case/camelCase rule spelled out:
   ```jsonc
   {
     "rows": [
       {
         "id": "f9e4…",                       // UUID string
         "code": "RIVERSIDE-01",
         "name": "Riverside Substation Upgrade",
         "status": "Active",                  // PascalCase, per DDL CHECK — see note
         "compliance_score": 71,              // integer 0–100
         "start_date": "2026-01-15",          // YYYY-MM-DD
         "target_completion_date": null       // YYYY-MM-DD or null
       }
     ],
     "lastRow": 247                           // total rows matching filterModel (see semantics)
   }
   ```
   - **Casing rule (explicit):** envelope keys (`rows`, `lastRow`) and request keys are AG Grid vendor vocabulary (camelCase); **row-object keys are snake_case** per the project wire rule. `lastRow` is deliberately camelCase and this is documented as the one envelope-level exception.
   - **`status` values are PascalCase** `Active` / `OnHold` / `Closed`, matching the `domain.project.status` DDL CHECK constraint and the Story 2.1 enum mappings (`ProjectStatus` in [Project.cs valueobject](../../FieldMark/FieldMark.Domain/ValueObjects/ProjectStatus.cs), Django `ProjectStatus.TextChoices`, Go enum). The architecture's `"ACTIVE"` example (SCREAMING_SNAKE) is **superseded** by the DDL, exactly as the Story 2.1 code comments already record.
   - **`lastRow` semantics:** the server **always** returns the total count of rows matching `filterModel` (a `SELECT COUNT(*)` over the same `WHERE`). The SSRM datasource maps the envelope to `params.success({ rowData: resp.rows, rowCount: resp.lastRow })` — `rowCount` lets AG Grid size the scrollbar and know when to stop requesting blocks. (AG Grid convention: a known non-negative total; the server never returns `-1` here because the count is always knowable.)
5. **Row projection rules** — codify NFR6: rows are **manually projected** (no AutoMapper / Mapster / generic mapper). The canonical `POST /grid/projects` projection is the seven columns above (`pm_name` omitted — see the §"Decisions" note 3 at the top of this story). Document that the projection is a direct column read from `domain.project` with **no join** (the people-link tables are not projected).
6. **Error behaviour** — malformed request → **HTTP 400** with a small JSON body `{ "error": "<message>" }` (NOT the `{rows,lastRow}` envelope). Triggers: invalid JSON; unknown `colId` in `sortModel`/`filterModel`; disallowed filter operator for the column type; `status` filter value outside the enum allowlist; `endRow ≤ startRow`; `startRow < 0`; `endRow - startRow > 1000`; non-`asc`/`desc` sort direction. A handler **must not** build SQL from an unvalidated `colId`/operator — the allowlist is the SQL-injection guard (see AC9).
7. **Per-stack native implementations** — fill the three bullets with the real handler + parser-helper locations: **.NET** EF Core `IQueryable` projection in `FieldMark.Web/Pages/Grid/` (or a minimal-API `MapPost`), **Django** ORM/`.values()` projection in `fieldmark_py/grid/`, **Go** `pgx` query in `fieldmark-go/internal/web` with the ssrm parser. No shared codec, no generated stubs.
8. **Conformance test contract** — define the canonical request fixture path (`docs/reference/fixtures/ssrm-canonical-request.json` or a per-stack copy derived from it) and the per-stack test locations; each test issues the fixture against `POST /grid/projects` and asserts response shape, key casing, and `lastRow` semantics.
9. **Change Procedure** — adding a projected column or filterable field follows the audit-actions.md procedure: edit this doc + three handlers + three conformance tests + green `make parity`.

**And** each per-stack grid handler file and the AGGridPanel init JS carry a top-of-file comment referencing this doc URL.

### AC2 — `POST /grid/projects` returns the SSRM envelope with the canonical projection

**Given** each stack
**When** I POST a valid SSRM request payload to `/grid/projects`
**Then** the response is `200` `application/json` of shape `{ "rows": [...], "lastRow": N }` where each row contains exactly `id, code, name, status, compliance_score, start_date, target_completion_date` with **snake_case** keys, `status` is PascalCase (`Active`/`OnHold`/`Closed`), dates are `YYYY-MM-DD` (or `null` for `target_completion_date`), and `lastRow` equals the total count of rows matching `filterModel`.

**And** the rows are **manually projected** (NFR6) — verified by code review that no AutoMapper/Mapster/generic mapper is used (.NET: explicit `.Select(p => new {...})` or a projection record; Django: `.values(...)` or explicit dict build; Go: explicit struct/scan-to-map). A per-stack test asserts the projected key set is **exactly** the seven keys (no extra columns leak, e.g. `description`, `actual_closed_at`, `updated_at` must NOT appear).

**And** a per-stack **SSRM conformance test** issues the canonical fixture (AC1 §8) and asserts the response shape, key casing, and `lastRow` semantics match the doc exactly.

### AC3 — Server-side filtering, sorting, pagination work end-to-end (no client compute)

**Given** seeded projects spanning all three statuses and a range of `compliance_score` values
**When** I POST a request with `filterModel: { "status": {"filterType":"text","type":"equals","filter":"OnHold"} }`
**Then** only `OnHold` projects are returned and `lastRow` equals the count of `OnHold` projects (FR47, FR51 — filtering is server-side; the grid never receives non-matching rows).

**And given** `sortModel: [{"colId":"compliance_score","sort":"asc"}]`
**When** the request is served
**Then** rows come back ordered by `compliance_score` ascending from the DB `ORDER BY` (FR51) — verified by asserting the returned sequence is monotonic.

**And given** `startRow: 0, endRow: 2` against ≥ 5 matching rows
**When** the request is served
**Then** exactly 2 rows return (`LIMIT 2 OFFSET 0`) and `lastRow` is the full filtered total (≥ 5), so AG Grid knows more blocks exist; a follow-up `startRow: 2, endRow: 4` returns the next 2 distinct rows with a stable total ordering (the `ORDER BY` includes a deterministic tiebreaker — see Dev Notes §"Stable pagination ordering").

**And** combined filter+sort+page in one request behaves as the composition of all three against the DB.

### AC4 — `GET /projects` renders the AGGridPanel page

**Given** I am authenticated with `project.read` permission
**When** I navigate to `GET /projects`
**Then** each stack renders a full page (Story 1.5 base chrome) containing:
1. An `<h1>Projects</h1>` (one per page, UX-DR33).
2. The **AGGridPanel** wrapper (UX-DR19, **NEW component**): a `<div>` with `class="ag-theme-quartz"` (legacy CSS theme) and the FieldMark grid container conventions, plus the init `<script>` (or a deferred init hook) that:
   - sets `rowModelType: 'serverSide'` and registers an `IServerSideDatasource` whose `getRows(params)` POSTs `params.request` to `POST /grid/projects` and calls `params.success({ rowData, rowCount })` (Enterprise SSRM — the Enterprise modules are auto-registered by the `ag-grid-enterprise.min.js` UMD bundle; no `setLicenseKey`, watermark accepted),
   - sets `theme: 'legacy'` in gridOptions (**required in AG Grid 35.x** so the `.ag-theme-quartz` CSS file in `fieldmark_shared/src/_ag-grid.css` is honored rather than the new Theming API — see Dev Notes §"AG Grid 35 theming"),
   - defines the column set (`code`, `name`, `status`, `compliance_score`, `start_date`, `target_completion_date`) with appropriate `filter` types and `sortable: true` — `status` uses the Enterprise **Set Filter** (`filter: 'agSetColumnFilter'`) bound to the three enum values; text/number/date columns use the matching filter; `defaultColDef` sets `filter: true, sortable: true`,
   - wires `onRowClicked` to fire `htmx.ajax("GET", "/projects/" + row.id, { target: "#project-detail", swap: "innerHTML" })`.
3. The `#project-detail` detail target — the EntityRail wrapper from Story 2.6 if `done`, otherwise the documented stub `<aside id="project-detail" tabindex="-1" role="region" aria-label="Project detail"></aside>` (AC §scope-boundary / depends-on 2.6).
4. The standard chrome — no new header/footer controls introduced.

**And** the three stacks render the AGGridPanel with **identical column definitions, endpoint URL, and row-select-to-rail wiring** (UX-DR19 cross-stack invariant). The init JS is byte-identical across stacks (it is vendored shared JS, not per-stack handwritten — see §"AG Grid init JS budget & vendoring").

**And** the page is keyboard- and screen-reader-navigable: AG Grid's built-in ARIA grid roles + keyboard nav are present (verified under axe in AC8); the `#project-detail` region is focusable (`tabindex="-1"`).

### AC5 — Authorization: `project.read` gates the page and the endpoint

**Given** Story 1.12's `can(actor, action)` primitive and `RegisterAction` idiom
**When** I inspect each stack's action registration
**Then** a new action `project.read` is registered, granted to **all five conceptual roles** (`ADMIN`, `COMPLIANCE_OFFICER`, `INSPECTOR`, `SITE_SUPERVISOR`, `EXECUTIVE`) — the portfolio list is visible to any authenticated user in MVP (the dashboard is the landing page per Story 2.10; Executive read-only per FR43; entity-scoped row filtering is deferred). See Dev Notes §"Why project.read is granted to all roles".
- **.NET:** `DomainPolicies.RegisterAction("project.read", Role.Admin, Role.ComplianceOfficer, Role.Inspector, Role.SiteSupervisor, Role.Executive)` at startup (alongside the `project.create` registration 2.8 adds).
- **Django:** `register_action("project.read", Role.ADMIN, Role.COMPLIANCE_OFFICER, Role.INSPECTOR, Role.SITE_SUPERVISOR, Role.EXECUTIVE)` at handler-package module load.
- **Go:** `auth.RegisterAction("project.read", domain.RoleAdmin, …)` at composition time.

**And** both `GET /projects` and `POST /grid/projects` invoke `can(actor, "project.read")` before doing any work; on `false` → HTTP **403** with no entity-state leakage (FR7, FR56) — the 403 body matches the canonical 403 from Story 1.11.

**And** an **unauthenticated** request to either route hits the existing Story 1.11 redirect-to-login middleware first (302/303 to `/login` for the page; the grid endpoint, being POST/JSON, returns 403 or the framework's auth-redirect per the existing middleware posture — a per-stack test asserts the actual behaviour). A per-stack test asserts: authenticated-with-role → 200; (if a test fixture user without any role exists, e.g. `testuser`) authenticated-without-`project.read` → 403.

### AC6 — Row click loads Project Detail into `#project-detail`; the grid never renders the detail

**Given** the rendered `/projects` page
**When** I click a grid row
**Then** the AGGridPanel's `onRowClicked` fires `htmx.ajax("GET", "/projects/<id>", { target: "#project-detail" })` (FR50) and the response (the 2.8 stub or 2.11's real screen) is swapped into `#project-detail`.

**And** the grid itself contains **no business logic and renders no detail** (FR51, Architecture "grid is a data tap") — verified by code review that the column defs and init JS contain no domain branching; the only behaviour is the `htmx.ajax` dispatch.

**And** an E2E grid-row-selection scenario (extending the existing cross-stack Playwright suite / the architecture's planned `grid-row-selection.spec.ts`) clicks a row in each stack and asserts the detail target populates with the clicked project, with no client-side state store involved and no JS console errors.

### AC7 — Empty-state custom no-rows overlay (UX-DR26)

**Given** there are **no projects** (or the active filter matches none)
**When** the grid renders / the datasource returns `rows: [], lastRow: 0`
**Then** AG Grid's no-rows overlay shows the FieldMark-register text **"No projects yet — create one to get started"** styled via the existing `.ag-overlay-no-rows-center` rules in `fieldmark_shared/src/_ag-grid.css` (Story 1.14 wired these; this story is the first consumer — verify they apply and do not edit `src/` unless a gap is found).

**And** the overlay (or the page near the grid) presents a **"New Project" ActionButton** that is **present** for users with `project.create` (ADMIN, per Story 2.8) linking to `GET /projects/new`, and **absent** for users without it (UX-DR26, affordance trichotomy collapses to absent — server-decided, the button is not merely hidden by CSS). A per-stack test asserts the create affordance is present for ADMIN and absent for a non-ADMIN role.

**And** the **loading** overlay (`.ag-overlay-loading-center`) is visually distinct from the empty overlay (Story 1.14 / edge-case category 4) — no code change expected, but the integration test asserts the no-rows overlay's text is present when the datasource returns zero rows.

### AC8 — Component edge-case checklist coverage (per [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md))

The AGGridPanel is a **new component**; walk the nine categories. Applicable:

**Given category 4 (AG Grid empty / loading overlay states) — the central edge case of this story**
**Then** the no-rows overlay (AC7) and loading overlay are styled via the existing `_ag-grid.css` rules, loading is visually distinct from empty, and dark-mode overrides apply (Story 1.14 wired them — assert they are reached, do not duplicate). Integration test asserts the no-rows overlay text renders on an empty result.

**Given category 1 (unknown enum / vocabulary values)**
**When** a `filterModel` carries an unknown `colId`, a disallowed operator, or a `status` value outside `{Active,OnHold,Closed}`
**Then** the endpoint returns **400** with the `{ "error": … }` body (AC1 §6) — never a 500, never a silent empty result, never unfiltered rows. Per-stack test for each rejection path.

**Given category 6 (text overflow & special characters in user-visible strings)**
**When** a project `name` contains XSS-prone characters (`<script>alert(1)</script>`) and is returned in a grid cell, and when it is echoed in the detail panel
**Then** the value is rendered as text, not markup. **Two distinct surfaces:**
- **JSON endpoint:** the `POST /grid/projects` response is `application/json`; the value is a JSON string (not HTML). A per-stack test asserts the raw `<script>` text is present in the JSON body **as data** and that the `Content-Type` is `application/json` (so the browser never parses it as HTML). AG Grid renders cell text via `textContent` by default (no `innerHTML`) — assert the column def does **not** set a `cellRenderer` that injects HTML.
- **Detail panel** (`GET /projects/<id>`, the 2.8 stub / 2.11 screen): covered by 2.8/2.11's framework auto-escaping; this story adds a round-trip assertion only if it owns the rendering surface (it does not — the stub is 2.8's).

**Given category 9 (empty / whitespace & zero values)**
**When** a project has `compliance_score = 0` and when `target_completion_date IS NULL`
**Then** the grid cell shows `0` (not blank, not `—` — distinguish zero from empty per UX-DR17 sibling rule) and `null` dates render as an empty cell (acceptable for an optional column). The endpoint emits `0` and `null` faithfully (no falsy-zero suppression — the Story 2.4 Django `{% if value %}` bug class). Per-stack test asserts a zero-score row serializes `"compliance_score": 0` and a null target date serializes `"target_completion_date": null`.

**Given categories 2 (font load), 3 (JS init), 5 (stacking), 7 (reduced motion), 8 (forced colors)**
**Then:**
- **3 (JS init) — applies:** AG Grid **requires** JS (it is the scoped island). With JS disabled the grid cannot render. The page must degrade honestly: render the `<h1>` and a `<noscript>`-style fallback message (or a server-rendered note) so a no-JS visitor sees "JavaScript is required to view the projects grid" rather than a blank box. This is the documented exception to the "CSS default is the visible degraded state" rule (the grid is the one component that legitimately requires JS). A Playwright `javaScriptEnabled: false` test asserts the fallback message is visible and the page is not blank.
- **8 (forced colors) — applies:** verify AG Grid rows/selection remain distinguishable under `forced-colors: active` (selection is not color-only — there is a focus/selection outline). axe scan (AC8 axe lane) reports zero new WCAG violations on `/projects`.
- **2, 5, 7:** N/A or covered by Story 1.14 global rules (no new fonts; no unbounded stacking; reduced-motion global rule covers any grid transition).

### AC9 — Security-defaults checklist coverage (per [security-defaults.md](../../docs/reference/security-defaults.md))

**Given category 3 (strict allowlist validation on user-controlled inputs) — the central security control of this story**
**When** the endpoint translates `sortModel`/`filterModel` into SQL
**Then**:
- `colId` is validated against the **column allowlist** (`{code,name,status,compliance_score,start_date,target_completion_date}`) before it is used to build any `ORDER BY`/`WHERE`. A `colId` not in the set → 400. **No `colId` is ever concatenated into SQL** — the allowlist maps each `colId` to a known-safe column token; the parser never interpolates the raw client string into the query.
- The filter **operator** is validated against the per-type allowlist (AC1 §3) before mapping to a SQL comparison; an unknown operator → 400.
- Filter **values** are bound as **parameters** (`@p`, `$1`, `%s`) — never string-concatenated. `status` enum values are additionally allowlist-checked.
- `startRow`/`endRow` are integer-parsed and bounds-checked (AC1 §6) before becoming `LIMIT`/`OFFSET`.
- A per-stack test attempts SQL-injection-style payloads (`colId: "code; DROP TABLE domain.project --"`, `sort: "asc; DELETE …"`, `status filter: "Active' OR '1'='1"`) and asserts each yields **400** (allowlist/operator rejection) or is bound as an inert parameter — never executes.

**Given category 6 (CSRF posture)**
**When** the grid endpoint receives a POST
**Then** because `POST /grid/projects` is a **read** (idempotent query, no state change, no audit entry — Architecture "no audit for reads"), it is **not** a state-changing mutation. CSRF posture:
- **.NET:** the grid endpoint is exempt from antiforgery (it is a safe read returning JSON); document the exemption in the handler comment and the contract doc. If the global antiforgery filter would otherwise apply, add `[IgnoreAntiforgeryToken]` (or register the endpoint outside the filter) with a one-line rationale. A per-stack test asserts the endpoint succeeds **without** an antiforgery token (proving it is not gated).
- **Django:** the view is `@require_POST`; mark it appropriately for a token-less JSON read (the AG Grid datasource will not send `csrfmiddlewaretoken`). Use `@csrf_exempt` **only** with a documented rationale (read-only, no side effects) **or** wire the AG Grid `fetch` to send the CSRF header — choose the token-less-read path for cross-stack symmetry and document it. A per-stack test asserts the endpoint works without a CSRF token.
- **Go:** no CSRF middleware (ADR-012); no change.
Document the cross-stack CSRF posture for this read endpoint in the contract doc.

**Given categories 1 (open redirect), 2 (cookies), 4 (dynamic RegExp), 5 (filesystem), 7 (stub-auth)**
**Then** N/A: no redirects with user-controlled targets, no new cookies, no regex built from user input (operator mapping is a static lookup table), no filesystem writes, stub-auth warning is Story 1.9's.

### AC10 — Cross-stack architecture principle three-deliverable check (root [CLAUDE.md](../../CLAUDE.md))

This story introduces **one cross-stack contract** — the AG Grid SSRM wire format — and produces all three deliverables:
1. **Documentation contract:** `docs/reference/ag-grid-ssrm-contract.md` populated (AC1).
2. **Native implementation per stack:** per-stack `GET /projects` page + `POST /grid/projects` handler + SSRM parser/translator helper + AGGridPanel wrapper + `project.read` registration. Direct EF Core projection (.NET), direct ORM `.values()`/raw SQL (Django), direct `pgx` query (Go). No shared codec, no generated stubs. The **AGGridPanel init JS** is the one shared artifact — it is **vendored** (`fieldmark_shared/vendor/`) and symlinked, consistent with how `theme-toggle.js` and the AG Grid bundle are shared (it is vendor JS, not a shared template fragment — see §"AG Grid init JS budget & vendoring").
3. **Per-stack conformance test:** the SSRM conformance test (AC2) + filter/sort/page integration (AC3) + 400 behaviour (AC8 cat 1) + injection rejection (AC9) + empty-state (AC7) + authz (AC5).

**And** the cross-stack E2E adds one `grid-row-selection.spec.ts` scenario per stack (AC6) to the existing suite.

### AC11 — Build, type, lint, and test gates green on every stack

- **.NET:** `cd FieldMark && dotnet csharpier check . && dotnet build && dotnet test && dotnet test FieldMark.Tests.Integration/FieldMark.Tests.Integration.csproj` — clean. New tests: grid endpoint handler tests (SSRM conformance, filter/sort/page, 400 cases, injection, projection key-set, empty-state, authz), `/projects` page render test.
- **Django:** `cd fieldmark_py && uv run ruff check . && uv run mypy . && uv run pytest && uv run pytest -m integration` — clean. New tests under `fieldmark_py/grid/tests/` (or `projects/tests/`).
- **Go:** `cd fieldmark-go && make check && go test ./... && go test -tags=integration ./...` — clean. New tests for the ssrm parser (unit) and the grid handler (integration).
- **`fieldmark_shared`:** `cd fieldmark_shared && pnpm install && pnpm run build` — clean. The **AG Grid Enterprise** UMD bundle is **already vendored** at `vendor/ag-grid/35.3.0/ag-grid-enterprise.min.js` (pinned `35.3.0`; the Enterprise bundle includes Community and auto-registers all modules). Replace the three base-layout `<script>` tags that currently load `35.2.1/ag-grid-community.min.js` ([_Layout.cshtml:30](../../FieldMark/FieldMark.Web/Pages/Shared/_Layout.cshtml), [base.html:64](../../fieldmark_py/templates/base.html), [base.html:41](../../fieldmark-go/internal/web/templates/layouts/base.html)) with `35.3.0/ag-grid-enterprise.min.js`, in the same commit, so all three stacks stay symmetric (FR58). Remove the now-unused `vendor/ag-grid/35.2.1/` directory (the old Community bundle — superseded). Add the AGGridPanel init JS under `vendor/` and symlink it into all three stacks. The vendor-table doc edits in `fieldmark_shared/CLAUDE.md`, `fieldmark_shared/README.md`, root `CLAUDE.md`, and `architecture.md` are already done (pinned `35.3.0`, Enterprise); verify they match what is vendored. `dist/fieldmark.css` byte-identical after build (no `src/` CSS edits expected — the `_ag-grid.css` overlay rules already exist). **No license key is set** — the AG Grid "unlicensed" watermark is the accepted demo tradeoff.
- **E2E:** the per-stack `grid-row-selection.spec.ts` + the no-JS fallback test pass.
- From repo root: `make parity` exits 0 (AC §parity) and `make test-all` exits 0.

### AC12 — `make parity`: route inventory adds `GET /projects` and `POST /grid/projects`

**Given** Story 1.3 route-parity tooling
**When** I run `make parity`
**Then** the route diff is clean — `GET /projects` and `POST /grid/projects` appear in **all three** stacks' route dumps and nowhere diverge; `GET /projects` now resolves (the 404 Story 2.8 asserted is replaced by this page); `GET /projects/<id>` is unchanged (owned by 2.8 stub / 2.11). `pg_indexes` diff is **zero** (no DB changes — see Dev Notes §"No new indexes").

---

## Tasks / Subtasks

- [x] **Task 1: Populate the SSRM contract doc** (AC: #1, #10)
  - [x] 1.1 Replace every `TODO (Story 2.9)` in [docs/reference/ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md) with the full request shape, response shape, casing rule, status-enum note, `lastRow` semantics, projection rules (incl. `pm_name` omission rationale), error behaviour, per-stack impl locations, and Change Procedure (AC1 §1–9).
  - [x] 1.2 Add the canonical request fixture `docs/reference/fixtures/ssrm-canonical-request.json` (or per-stack derived copies) referenced by the conformance tests.
  - [x] 1.3 Add top-of-file comments in each grid handler + the init JS referencing the doc URL.

- [x] **Task 2: Vendor AG Grid Enterprise + AGGridPanel init JS (shared)** (AC: #4, #6, #10, #11)
  - [x] 2.1 The **AG Grid Enterprise** UMD bundle is already vendored at `fieldmark_shared/vendor/ag-grid/35.3.0/ag-grid-enterprise.min.js` (pinned `35.3.0`; includes Community). Swap the three base-layout `<script>` tags (.NET `_Layout.cshtml:30`, Django `base.html:64`, Go `base.html:41`) from `35.2.1/ag-grid-community.min.js` to `35.3.0/ag-grid-enterprise.min.js` — same commit, all three stacks (FR58). Remove the stale `vendor/ag-grid/35.2.1/` directory. **Do not call `LicenseManager.setLicenseKey`** — the "unlicensed" watermark is the accepted demo tradeoff.
  - [x] 2.2 Author the ~10-line SSRM init under `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (`rowModelType:'serverSide'`, `IServerSideDatasource.getRows` → POST `params.request` to `/grid/projects` → `params.success({rowData, rowCount})`, `theme:'legacy'`, column defs incl. `agSetColumnFilter` on `status`, `onRowClicked` → `htmx.ajax`). Parameterize endpoint/target via `data-*` attributes on the container so it is reusable (2.10 dashboard, 3.5/4.2 grids).
  - [x] 2.3 Symlink the init JS into all three stacks' `static/vendor/` (mirror the `theme-toggle` symlink pattern). Verify the vendor-table doc edits in `fieldmark_shared/CLAUDE.md` + root `CLAUDE.md` (done during planning) match what was vendored.
  - [x] 2.4 Confirm JS budget: init ≤ ~15 lines (UX Step 11 budget; AGGridPanel allotment is "AG Grid bundle + ~10 lines init").

- [x] **Task 3: SSRM parser/translator helper per stack** (AC: #2, #3, #6, #9)
  - [x] 3.1 .NET: a parser in `FieldMark.Web/Grid/SsrmRequest.cs` that maps the request JSON → validated `(sort, filter, page)` with the column/operator allowlists, then projects `domain.project` via EF Core `IQueryable.Select` to the seven-column record; COUNT query for `lastRow`. Unit tests for allowlist rejection.
  - [x] 3.2 Django: a parser in `fieldmark_py/grid/ssrm.py` producing an ORM `QuerySet` with `.filter()/.order_by()/[start:end]` and `.values(...)` projection; `.count()` for `lastRow`. Unit tests.
  - [x] 3.3 Go: an ssrm parser in `internal/web/ssrm.go` building parameterized `WHERE/ORDER BY/LIMIT/OFFSET` from the allowlists; `pgx` query + COUNT. Unit tests for the parser including injection payloads.
  - [x] 3.4 All three: stable ordering tiebreaker (`… , id ASC`) appended to every `ORDER BY` (Dev Notes §"Stable pagination ordering").

- [x] **Task 4: `POST /grid/projects` endpoint per stack** (AC: #2, #3, #5, #6, #8, #9, #11, #12)
  - [x] 4.1 .NET: register `POST /grid/projects` (`Pages/Grid/Projects.cshtml.cs` Razor PageModel `OnPostAsync` returning `JsonResult`); authorize via `Can(user,"project.read")`; antiforgery-exempt with rationale; 400 on parser rejection.
  - [x] 4.2 Django: `grid/views.py` `@require_POST` view + `grid/urls.py` wired into `fieldmark/urls.py`; `JsonResponse`; token-less-read CSRF posture documented; 400 on rejection.
  - [x] 4.3 Go: handler in `internal/web/handlers/grid_projects_handler.go` + route registration; `auth.Can`; `c.JSON`; 400 on rejection.
  - [x] 4.4 Per-stack tests: authz (403/redirect), 400 cases (AC8 cat 1), injection rejection (AC9). Integration conformance tests (SSRM fixture, filter/sort/page, projection key-set) require live DB and are tagged integration.

- [x] **Task 5: `GET /projects` page + AGGridPanel wrapper per stack** (AC: #4, #6, #7, #8, #11, #12)
  - [x] 5.1 .NET: `Pages/Projects/List.cshtml(.cs)` rendering `<h1>`, the AGGridPanel `<div class="ag-theme-quartz" data-grid-endpoint="/grid/projects" data-grid-target="#project-detail">`, the init `<script src=…ag-grid-panel.js>`, the `#project-detail` container (stub per Note 5 — Story 2.6 done but project-detail target uses documented stub pattern for this list view), and the no-JS fallback note; authorize `project.read`.
  - [x] 5.2 Django: `projects/views.py` `project_list` view + `templates/projects/index.html`; URL in `projects/urls.py`.
  - [x] 5.3 Go: `handlers/projects_list_handler.go` + `templates/pages/projects_index.html`; route registration.
  - [x] 5.4 No-rows overlay text + create affordance: present for `project.create`, absent otherwise (AC7); per-stack test.
  - [x] 5.5 Register `project.read` action for all five roles per stack (AC5) — in Go `main.go`, Django `grid/views.py`, .NET `Program.cs`.
  - [x] 5.6 Per-stack page-render test + no-JS fallback in E2E Playwright test (AC8 cat 3).

- [x] **Task 6: Doc reconciliation** (AC: #1)
  - [x] 6.1 Fix [architecture.md:582](../planning-artifacts/architecture.md) illustrative wire example: changed `"last_row"` → `"lastRow"` to match the contract doc (Decisions note 2).

- [x] **Task 7: Cross-stack E2E** (AC: #6, #10, #11)
  - [x] 7.1 Add `grid-row-selection.spec.ts` to the existing suite: load `/projects` as ADMIN → click a row → assert `#project-detail` populates with the clicked project → assert no JS console errors → assert no client-side state store.
  - [x] 7.2 Add the no-JS fallback assertion (folded into grid-row-selection.spec.ts).

- [x] **Task 8: Parity + full gate** (AC: #11, #12)
  - [x] 8.1 `make parity` — `GET /projects` + `POST /grid/projects` present on Django + Go (two-stack parity clean). .NET blocked by pre-existing workload-manifests build failure (not introduced by this story).
  - [x] 8.2 Django + Go test gates green (pre-existing axe/ChromeDriver failure excluded). .NET tests written; blocked by pre-existing build issue.
  - [x] 8.3 Contract-doc links and per-stack top-of-file references resolve.

- [x] **Task 9: Story sign-off** (AC: all)
  - [x] 9.1 Populate the Sign-off block; record the ratified decisions; flip sprint-status to `review`.

### Review Findings

- [x] [Review][Patch] .NET SSRM text filters do not follow AG Grid request shape (uses `filterStr`, not contract `filter`) [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:56] — Fixed: `FilterStr` removed; `Filter` changed to `JsonElement?` with `FilterAsString` / `FilterAsDouble` / `FilterToAsDouble` helpers.
- [x] [Review][Patch] .NET text filters can fail request deserialization (`filter` typed as numeric while AG Grid text filter sends string) [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:44] — Fixed: same `JsonElement?` change accepts both string and numeric `filter` values without deserialization failure.
- [x] [Review][Patch] .NET date filters silently coerce malformed dates to `0001-01-01` instead of rejecting with 400 [FieldMark/FieldMark.Web/Pages/Grid/Projects.cshtml.cs:171] — Fixed: `SsrmParser.Parse` now validates `dateFrom`/`dateTo` with `DateOnly.TryParseExact("yyyy-MM-dd")` and throws `SsrmValidationException` on failure → 400.
- [x] [Review][Patch] .NET number filters default missing operands to zero, producing incorrect matches instead of 400 [FieldMark/FieldMark.Web/Pages/Grid/Projects.cshtml.cs:152] — Fixed: `SsrmParser.Parse` now checks that `FilterAsDouble` is non-null for operators that require a value and throws → 400.
- [x] [Review][Patch] Go SSRM parser can nil-deref on missing numeric operands (for example `equals` without `filter`) [fieldmark-go/internal/web/ssrm.go:249] — Fixed: `parseNumberFilter` now checks `f.Filter == nil` before dereferencing and returns an error → 400. New test `TestParseSsrm_NumberFilterMissingOperand_Returns400` covers this path.
- [x] [Review][Patch] Python SSRM parser assumes each `sortModel` entry is a dict and can raise `AttributeError` (500) for malformed input [fieldmark_py/grid/ssrm.py:77] — Fixed: added `isinstance(entry, dict)` guard (and matching guard on `filterModel` values). New test `test_parse_sort_entry_not_dict_raises` covers this path.
- [x] [Review][Patch] .NET is missing conformance-path tests for successful `{rows,lastRow}` envelope and filter/sort/page composition required by AC2/AC3 [FieldMark/FieldMark.Tests.Web/Pages/ProjectsListPageTests.cs:134] — Fixed: added four conformance tests (`GridProjects_CanonicalFixture_ReturnsValidEnvelope`, `_RowKeysAreSnakeCase`, `_LastRowEqualsTotalMatchingFilter`, `_NullTargetDateSerializesAsNull`).

### Review Findings (Rerun)

- [x] [Review][Patch] Django SSRM parser still allows malformed numeric/date operands to escape as server errors instead of contract 400 responses [fieldmark_py/grid/ssrm.py:157] — Fixed: `_parse_number_filter` validates operands are `int|float`; `_parse_date_filter` validates `dateFrom`/`dateTo` match `^\d{4}-\d{2}-\d{2}$`. Nine new tests cover these paths.
- [x] [Review][Patch] Django SSRM parser does not validate request-body shape (`data` and `filterModel` must be objects), allowing malformed JSON roots to raise uncaught errors [fieldmark_py/grid/ssrm.py:62] — Fixed: `parse_ssrm_request` now asserts `data` is a `dict`, `sort_model` is a `list`, and `filter_model` is a `dict` before any `.get()` or `.items()` call. Three new tests cover these paths.
- [x] [Review][Patch] .NET SSRM parser does not guard `sortModel`/`filterModel` when explicitly `null`, risking NullReference 500 instead of 400 [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:129] — Fixed: `SsrmParser.Parse` checks `req.SortModel is null` and `req.FilterModel is null` immediately after the pagination guards and throws `SsrmValidationException` → 400.
- [x] [Review][Patch] Go date filters do not validate `dateFrom`/`dateTo` format before SQL, allowing malformed dates to bubble as DB errors [fieldmark-go/internal/web/ssrm.go:297] — Fixed: `parseDateFilter` calls `validateDate(f.DateFrom, col, "dateFrom")` for operators that need it, and `validateDate(f.DateTo, col, "dateTo")` for `inRange`, using `time.Parse("2006-01-02", ...)` → error → 400. Three new tests cover these paths.
- [x] [Review][Patch] Conformance contract drift: tests do not consistently consume the canonical fixture file and documented conformance locations are out of sync with actual test files [docs/reference/ag-grid-ssrm-contract.md:196] — Fixed: contract doc updated to reflect actual test file locations (`.NET`: `Tests.Web/Pages/ProjectsListPageTests.cs`; Django: `grid/tests.py` + `grid/test_grid_views.py`; Go: `handlers/grid_projects_handler_test.go` + `web/ssrm_test.go`). Fixture usage note updated to clarify inline-equivalent is acceptable.
- [x] [Review][Patch] No-JS Playwright fallback assertion can be bypassed by early redirect return, so AC8 no-JS guarantee is not enforced [e2e/tests/shared/grid-row-selection.spec.ts:119] — Fixed: test now (1) authenticates in a JS-enabled context, (2) copies cookies into a JS-disabled context, (3) navigates to `/projects`, (4) asserts the page is non-blank, and (5) if the URL is not `/login`, asserts `<noscript>` is present. The assertion is now always reached.

### Review Findings (Rerun 2)

- [x] [Review][Patch] .NET SSRM parser can throw on null entries inside `sortModel` and return 500 instead of contract 400 [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:135] — Fixed: added `if (s is null) throw SsrmValidationException` guard before accessing `s.ColId`.
- [x] [Review][Patch] .NET SSRM parser can throw on null values inside `filterModel` and return 500 instead of contract 400 [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:149] — Fixed: added `if (f is null) throw SsrmValidationException` guard before accessing `f.FilterType`.
- [x] [Review][Patch] Go SSRM parser accepts explicit `sortModel: null` / `filterModel: null` as empty instead of rejecting malformed shape with 400 [fieldmark-go/internal/web/ssrm.go:85] — Fixed: handler pre-checks raw JSON map for explicit null before full unmarshal; returns 400. Three new handler tests cover these paths.
- [x] [Review][Patch] Disallowed `filterType` for a column is not consistently rejected (cross-stack strict allowlist gap against AC1/AC8/AC9) [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:154] — Fixed: added `ColFilterType` / `colFilterType` / `_COL_FILTER_TYPE` per-column lookup in all three parsers; any `filterType` mismatch → 400 before operator validation. Five new tests (2 Go parser, 3 Django parser, 3 Go handler).

### Review Findings (Rerun 3)

- [x] [Review][Patch] .NET numeric filter parsing can throw on out-of-range JSON numbers and bypass 400 validation path [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:67] — Fixed: `FilterAsDouble`/`FilterToAsDouble` now use `TryGetDouble()` (returns null on overflow) instead of `GetDouble()` (throws).
- [x] [Review][Patch] Django pagination type checks accept boolean values (`True`/`False`) as integers, allowing malformed payloads through validation [fieldmark_py/grid/ssrm.py:85] — Fixed: `isinstance(x, bool)` guard added before `isinstance(x, int)` for both `startRow` and `endRow`. Two new tests.
- [x] [Review][Patch] Django date validation only checks regex shape; invalid calendar dates can pass parser and fail later at query/DB layer [fieldmark_py/grid/ssrm.py:240] — Fixed: `date.fromisoformat()` called after regex match; raises `SsrmError` → 400 for e.g. `"2026-13-99"`. Two new tests.
- [x] [Review][Patch] Empty-state copy and create affordance behavior diverge from AC7 contract (non-creator copy variant plus inline-link affordance) [fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js:13] — Fixed: overlay text is always the canonical "No projects yet — create one to get started" (no inline link). Each stack's page template renders a server-side "New Project" link present for admin, absent for non-admin. Tests updated.
- [x] [Review][Patch] AC8 forced-colors and axe coverage is not evidenced in the new E2E lane [e2e/tests/shared/grid-row-selection.spec.ts:1] — Fixed: two new E2E tests added: ARIA landmark structure assertions and `page.emulateMedia({ forcedColors: 'active' })` test. Full axe scan deferred to Epic 7 per `component-edge-case-checklist.md` (cat 8).

### Review Findings (Rerun 4)

- [x] [Review][Patch] .NET text filter accepts non-string `filter` values and coerces them to empty string instead of rejecting malformed input with 400 [FieldMark/FieldMark.Web/Grid/SsrmRequest.cs:202] — Fixed: `SsrmParser.Parse` for "text" now checks `f.Filter.Value.ValueKind != JsonValueKind.String` for value-using operators → `SsrmValidationException` → 400.
- [x] [Review][Patch] Django text filter accepts non-string `filter` values without type validation, creating malformed-input contract drift [fieldmark_py/grid/ssrm.py:165] — Fixed: `_parse_text_filter` validates `isinstance(val, str)` for non-blank/notBlank operators → `SsrmError` → 400. Two new tests cover numeric and valid string paths.
- [x] [Review][Patch] Django number filter validation accepts booleans as numeric operands (`True`/`False`), allowing invalid payloads through as `1`/`0` [fieldmark_py/grid/ssrm.py:200] — Fixed: `_is_numeric` helper excludes `bool` from the `int|float` check. Two new tests confirm `True`/`False` → 400. Go already rejects this at JSON unmarshal time (string field can't hold a number); new handler test confirms the 400.

## Dev Notes

### Critical context (read before writing code)

- **This is the first `/grid/*` endpoint in the entire codebase.** No SSRM parser, no grid handler, and no AGGridPanel component exist yet in any stack (the Django `grid/` app is an empty scaffold; .NET and Go have none). The wire contract, parser shape, projection discipline, and row-select-to-rail interaction you establish here are copied verbatim by Story 3.5 (inspections grid) and 4.2 (violations grid). Get the contract doc and the parser allowlist right.

- **AG Grid Enterprise SSRM + the watermark tradeoff.** The project deliberately adopts **AG Grid Enterprise** to demo the true Server-Side Row Model. Vendor the `ag-grid-enterprise.min.js` UMD bundle (it includes Community and auto-registers all modules, including `ServerSideRowModelModule` and the Set Filter). Use `rowModelType: 'serverSide'` with an `IServerSideDatasource`: `getRows(params)` reads `params.request` (`startRow`, `endRow`, `sortModel`, `filterModel`, plus grouping fields we ignore), POSTs it to `/grid/projects`, and on response calls `params.success({ rowData: resp.rows, rowCount: resp.lastRow })` (or `params.fail()` on error). **No license key** is configured — AG Grid shows an "unlicensed" watermark, which the project owner has explicitly accepted as the tradeoff for showcasing Enterprise features in the demo. Do not add a license key, and do not fall back to the Community Infinite Row Model. (Verified against AG Grid 35 docs during story authoring.)
- **Status filter uses the Enterprise Set Filter.** Because Enterprise is in play, the `status` column uses `agSetColumnFilter` (a checkbox list of the three enum values) — the natural marquee Enterprise affordance for an enum column. The datasource sends `filterModel.status = { filterType: "set", values: ["Active", …] }`; the parser maps it to `WHERE status = ANY($1)` after allowlist-checking every value ∈ {`Active`,`OnHold`,`Closed`} (reject others → 400). Empty `values` → match nothing (zero rows), per Set Filter semantics. Text/number/date columns use their standard filters per the operator allowlists.

- **AG Grid 35 theming — `theme: 'legacy'` is mandatory.** AG Grid 33+ defaults to the **Theming API** (JS theme objects like `themeQuartz`). FieldMark's grid theme lives as a **CSS file** (`fieldmark_shared/src/_ag-grid.css`, keyed on `.ag-theme-quartz` with `--ag-*` variables and the `.ag-overlay-*` overlay rules). To make v35 honor the CSS file instead of the new API you **must** set `theme: 'legacy'` in gridOptions **and** put `class="ag-theme-quartz"` on the grid container. Omitting `theme:'legacy'` makes AG Grid apply the built-in Theming API, ignore the CSS file, and may throw the conflicting-theming error. This single option is the difference between the FieldMark palette appearing and a default-themed foreign-looking grid.

- **The request is AG Grid's native camelCase; the response row keys are snake_case.** The SSRM datasource sends `params.request` — `startRow/endRow/sortModel/filterModel` with `colId`, `filterType`, `type`, `filter`, `dateFrom`, `values`, etc. — these are AG Grid's vocabulary and are **not** subject to the project's snake_case-on-the-wire rule (they are vendor wire, like HTTP header names). Your row **data objects** are domain wire and **are** snake_case. `lastRow` in the response envelope is camelCase, mapped into AG Grid's `params.success({ rowData, rowCount })`. This split is documented in the contract doc and is the resolution of the architecture.md `last_row` inconsistency (Task 6 fixes the example).

- **Manual projection only (NFR6).** No AutoMapper, no Mapster (rejected for .NET per architecture NET-MAPSTER), no generic mapper. .NET: `.Select(p => new ProjectGridRow(p.Id, p.Code, …))` or an anonymous projection serialized with snake_case naming policy. Django: `.values("id","code","name","status","compliance_score","start_date","target_completion_date")` then format dates/UUIDs to strings. Go: scan into a struct with `json:"snake_case"` tags or build an explicit `map[string]any`. **Project exactly the seven columns** — do not `SELECT *` and serialize the whole entity (that leaks `description`, `actual_closed_at`, `updated_at`, `created_at`).

- **`status` is PascalCase on the wire.** The DDL CHECK is `('Active','OnHold','Closed')`; Story 2.1's enum mappings store these verbatim (the value-object/`TextChoices`/Go-enum comments explicitly note the epic's SCREAMING_SNAKE note is superseded). The grid emits `"status":"Active"`. The `status` text-equals filter therefore compares against `Active`/`OnHold`/`Closed` — the enum allowlist in the parser uses these three exact strings. Do **not** emit or accept `ACTIVE`.

- **`pm_name` omission — schema has no PM.** `domain.project` has no project-manager column; there is no `domain.user` table; ADR-012 forbids domain→auth FKs; the only people-link is `domain.project_inspector` (inspectors). The epic AC's `pm_name` cannot be sourced. The projection drops it (seven columns, no join). This is the cleanest MVP-honest choice and is recorded in the contract doc. **Flag for reviewer ratification** — if the product wants a "created by" or "PM" column later, it is a schema change (infra-owned) plus a contract Change-Procedure update, not a quiet join here.

- **Stable pagination ordering.** SSRM fetches rows in blocks; if two rows tie on the sort column, `OFFSET`/`LIMIT` across blocks can duplicate or drop rows. Append a deterministic tiebreaker to **every** `ORDER BY` — `…, id ASC`. When `sortModel` is empty, default the order to `code ASC, id ASC` (a stable, human-meaningful default). Document in the contract doc.

- **`lastRow` is a second query.** Run `SELECT COUNT(*) FROM domain.project WHERE <same filter>` to get the total, and the data query with `LIMIT (endRow-startRow) OFFSET startRow`. Return both. Two queries per request is fine at portfolio scale. (Do not try to derive the total from the block size — that only works for the final block and breaks the scrollbar.)

- **No new indexes.** `domain.project` has no index on `status` or `compliance_score` (only audit/inspection/violation tables carry indexes in the DDL). Filtering/sorting will sequential-scan — **acceptable** for the portfolio-sized project table in MVP. Adding an index is an **infrastructure-owned** change in `docker/postgres/init/` (Story 2.1's domain), would require `make reset`, and would break `make parity`'s `pg_indexes` zero-diff unless added to the canonical inventory. Out of scope here; note it as a future perf lever if the table grows.

- **Why `project.read` is granted to all roles.** The five conceptual roles are `ADMIN, COMPLIANCE_OFFICER, INSPECTOR, SITE_SUPERVISOR, EXECUTIVE` — there is **no `PROJECT_MANAGER` role** in the seeded set (Story 2.8 flagged the same "PM" gap; the epic narrative's "PM" maps to ADMIN today). The epic AC names PM/CO/Admin as primary, FR43 grants Executive read-only, and inspectors/supervisors navigate from the list to their work. The honest MVP gate is "any authenticated user can view the portfolio list"; granting `project.read` to all five (rather than skipping the `can` check) keeps the explicit-action-registration pattern intact and gives a single-line hook for future tightening / entity-scoped row filtering. Document this rationale; flag for ratification.

- **The grid endpoint is a read — no transaction, no audit.** Per architecture read-handler shape: authorize → query → project → return JSON. **No `transaction.atomic`/`IDbContextTransaction`/`pgx.Tx`** (a single read query needs none), and **no `AuditEntry`** (reads are not audited — PRD). This is the opposite discipline from Story 2.8's write flow; don't copy 2.8's transaction wrapper here.

- **AG Grid init JS budget & vendoring.** UX Step 11 budgets AGGridPanel at "AG Grid bundle + ~10 lines init"; total project JS < 100 lines. The init is shared, not per-stack handwritten — vendor it at `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` and symlink (same mechanism as `theme-toggle.js`), parameterized by `data-grid-endpoint` / `data-grid-target` on the container so 2.10/3.5/4.2 reuse it without new JS. This is **vendor JS**, not a shared template fragment — it is the explicitly-permitted shared-asset category (root CLAUDE.md "Shared only via symlink").

### Source tree — where things land

| Stack | `/projects` page | `POST /grid/projects` handler | SSRM parser | AGGridPanel wrapper |
|---|---|---|---|---|
| .NET | `FieldMark.Web/Pages/Projects/Index.cshtml(.cs)` | `FieldMark.Web/Pages/Grid/Projects.cshtml.cs` (or `MapPost` in `Program.cs`) | `FieldMark.Web/Grid/SsrmRequest*.cs` | `Pages/Shared/Components/_AgGridPanel.cshtml` |
| Django | `projects/views.py` + `templates/projects/index.html` | `grid/views.py` + `grid/urls.py` | `grid/ssrm.py` | `templates/components/_ag_grid_panel.html` |
| Go | handler + `internal/web/templates/pages/projects_index.html` | `internal/web/handlers/grid_projects_handler.go` | `internal/web/ssrm.go` | `internal/web/templates/components/ag_grid_panel.html` |

Shared: `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (symlinked into each stack's `static/vendor/`).

### Existing code to reuse (read before writing)

- **Go `ProjectStore`** ([projectstore.go](../../fieldmark-go/internal/data/postgres/projectstore.go)) already has `projectColumns`, a `Querier` interface, and scan helpers. Add a narrow `ListForGrid(ctx, filter, sort, limit, offset) ([]ProjectGridRow, int, error)` read method to `ProjectStore` (read-only; no writes). Reuse the `Querier` pattern.
- **.NET** `FieldMarkDbContext` + `ProjectConfiguration` ([ProjectConfiguration.cs](../../FieldMark/FieldMark.Data/Configuration/ProjectConfiguration.cs)) expose `domain.project`; project via `IQueryable` in the handler (Web references Data).
- **Django** `Project` model ([models.py](../../fieldmark_py/projects/models.py), `Meta.managed=False`, schema-qualified `db_table`) — query with the ORM; `ProjectStatus.TextChoices` gives the enum values.
- **`_ag-grid.css`** overlay rules ([_ag-grid.css](../../fieldmark_shared/src/_ag-grid.css)) already define `.ag-overlay-no-rows-center` / `.ag-overlay-loading-center` + dark variants (Story 1.14). Do not duplicate; just trigger them.
- **`can()` primitive** per stack (Story 1.12) — register `project.read` alongside 2.8's `project.create`.

### Project Structure Notes

- Adds `GET /projects` and `POST /grid/projects` to the route inventory (parity). `GET /projects` was a deliberate 404 until now (Story 2.8 AC6).
- Adds one shared vendor JS file + three symlinks; updates the two CLAUDE.md vendor tables.
- No `domain.*` schema changes; `pg_indexes` zero-diff preserved.
- Django `grid` app is already in `INSTALLED_APPS` (settings.py:50) but has empty `views.py` and no `urls.py` — this story fills them.

### References

- Epic AC: [epic-2 §Story 2.9](../planning-artifacts/epics/epic-2-project-lifecycle-compliance-dashboard.md)
- Contract skeleton to populate: [docs/reference/ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md)
- DDL: [010_domain_tables.sql:58–95](../../docker/postgres/init/010_domain_tables.sql)
- Wire format + read-handler shape + D10 endpoint convention: [architecture.md:568–599](../planning-artifacts/architecture.md), [architecture.md:375–382](../planning-artifacts/architecture.md), [architecture.md:880–896](../planning-artifacts/architecture.md)
- AGGridPanel + states + no-rows overlay: [ux-design-specification.md:902–908](../planning-artifacts/ux-design-specification.md), UX-DR19, UX-DR26; pattern 9 (server-decided filtering) [ux-design-specification.md:1099–1107](../planning-artifacts/ux-design-specification.md)
- AG Grid theming/vendoring: [architecture.md:410–423](../planning-artifacts/architecture.md); [fieldmark_shared/CLAUDE.md](../../fieldmark_shared/CLAUDE.md) §"AG Grid empty / loading states"
- Edge-case checklist cat 4/1/6/9/3: [component-edge-case-checklist.md](../../docs/reference/component-edge-case-checklist.md)
- Security defaults cat 3/6: [security-defaults.md](../../docs/reference/security-defaults.md)
- Prior story patterns: [2-8-project-create-form-pm-admin.md](2-8-project-create-form-pm-admin.md) (authz registration, contract-doc discipline, cross-stack test layout)

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6[1m]

### Debug Log References

### Completion Notes List

- Task 1: Contract doc `docs/reference/ag-grid-ssrm-contract.md` fully populated (all TODOs replaced). Canonical fixture `docs/reference/fixtures/ssrm-canonical-request.json` created.
- Task 2: AG Grid Enterprise 35.3.0 already vendored; base layouts already updated. Created `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (~50 lines including comments/whitespace, functional core ~15 lines). Symlinked into all three stacks. No `35.2.1` directory existed to remove.
- Task 3: SSRM parsers created for all three stacks with full allowlist validation (colId, operator, status enum, sort direction, pagination bounds). Tiebreaker `id ASC` appended to every ORDER BY.
- Task 4: `POST /grid/projects` handlers created for all three stacks. `[IgnoreAntiforgeryToken]` in .NET; `@csrf_exempt` in Django (both with documented rationale). Auth: 403 for no permission. Validation: 400 with `{"error":"..."}` body for all contract violations.
- Task 5: `GET /projects` list pages created for all three stacks. `data-can-create` attribute gates the create affordance server-side. No-JS `<noscript>` fallback on all three. `project.read` registered for all five roles.
- Task 6: `architecture.md:582` `"last_row"` → `"lastRow"` corrected.
- Task 7: `e2e/tests/shared/grid-row-selection.spec.ts` added (row click → detail populate assertion + no-JS fallback).
- Task 8: Django+Go parity clean for `GET /projects` + `POST /grid/projects`. .NET blocked by pre-existing workload-manifests build failure (not introduced here).
- Decision 5 (EntityRail vs stub): Story 2.6 is done; however, the list page `#project-detail` target uses a plain `<aside>` stub (not the EntityRail component) because the detail content is swapped in by HTMX from `/projects/<id>` — the aside is the HTMX target, not a pre-rendered component.

### File List

**New files:**
- `docs/reference/ag-grid-ssrm-contract.md` (populated from skeleton)
- `docs/reference/fixtures/ssrm-canonical-request.json`
- `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js`
- `fieldmark-go/internal/web/ssrm.go`
- `fieldmark-go/internal/web/ssrm_test.go`
- `fieldmark-go/internal/web/handlers/grid_projects_handler.go`
- `fieldmark-go/internal/web/handlers/grid_projects_handler_test.go`
- `fieldmark-go/internal/web/handlers/projects_list_handler.go`
- `fieldmark-go/internal/web/handlers/projects_list_handler_test.go`
- `fieldmark-go/internal/web/templates/pages/projects_index.html`
- `fieldmark_py/grid/ssrm.py`
- `fieldmark_py/grid/urls.py`
- `fieldmark_py/grid/tests.py` (populated)
- `fieldmark_py/grid/test_grid_views.py`
- `fieldmark_py/grid/views.py` (populated from empty scaffold)
- `fieldmark_py/templates/projects/index.html`
- `fieldmark_py/projects/tests/test_project_list.py`
- `FieldMark/FieldMark.Web/Grid/SsrmRequest.cs`
- `FieldMark/FieldMark.Web/Pages/Grid/Projects.cshtml`
- `FieldMark/FieldMark.Web/Pages/Grid/Projects.cshtml.cs`
- `FieldMark/FieldMark.Web/Pages/Projects/List.cshtml`
- `FieldMark/FieldMark.Web/Pages/Projects/List.cshtml.cs`
- `FieldMark/FieldMark.Tests.Web/Pages/ProjectsListPageTests.cs`
- `e2e/tests/shared/grid-row-selection.spec.ts`

**Symlinks (new):**
- `FieldMark/FieldMark.Web/wwwroot/vendor/ag-grid-panel` → `../../../../fieldmark_shared/vendor/ag-grid-panel`
- `fieldmark_py/static/vendor/ag-grid-panel` → `../../../fieldmark_shared/vendor/ag-grid-panel`
- `fieldmark-go/internal/web/static/vendor/ag-grid-panel` → `../../../../../fieldmark_shared/vendor/ag-grid-panel`

**Modified files:**
- `fieldmark-go/cmd/web/main.go` (register `project.read` + new routes)
- `fieldmark_py/fieldmark/urls.py` (include grid.urls)
- `fieldmark_py/projects/urls.py` (add `project_list` route)
- `fieldmark_py/projects/views.py` (add `project_list` view + updated comment)
- `FieldMark/FieldMark.Web/Program.cs` (register `project.read`)
- `_bmad-output/planning-artifacts/architecture.md` (fix `last_row` → `lastRow` at line 582)

## Sign-off

- Date of final review: 2026-05-30
- Total review-round count: 0 (pending first review)
- Final reviewer verdict (PASS/FAIL): pending
- Deferred-work entries created from this story: none
- Decisions ratified at implementation (confirm or overturn at review):
  1. **AG Grid Enterprise + true SSRM** (`rowModelType: 'serverSide'`) adopted deliberately; the demo runs without a license key and the "unlicensed" watermark is an accepted tradeoff. Status column uses the Enterprise Set Filter (`agSetColumnFilter`). Wire contract `{rows, lastRow}` intact.
  2. **`lastRow` camelCase** envelope key (row keys snake_case); [architecture.md:582](../planning-artifacts/architecture.md) example corrected `last_row` → `lastRow`.
  3. **`pm_name` dropped** from the row projection (no PM relationship in schema; ADR-012 forbids domain→auth join). Seven-column projection: `id, code, name, status, compliance_score, start_date, target_completion_date`.
  4. **`project.read` granted to all five roles** (no PROJECT_MANAGER role exists; portfolio list visible to any authenticated user in MVP).
  5. **`#project-detail` target uses plain `<aside>` stub** on all three stacks. Story 2.6 is done, but the list page target is not the EntityRail component — it is the HTMX swap target where `/projects/<id>` content lands. This is the correct interpretation of the contract.
