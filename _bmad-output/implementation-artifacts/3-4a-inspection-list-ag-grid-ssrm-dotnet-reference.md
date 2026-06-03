# Story 3.4a: Inspection list AG Grid (SSRM) — .NET reference

Status: ready-for-dev

Epic: 3 — Inspection Workflow & Violation Genesis
Split group: **3.4 Inspection list AG Grid with SSRM endpoint** (a/b/c) — this is **`.a` the .NET reference**. See [docs/how-to/cross-stack-story-splitting.md](../../docs/how-to/cross-stack-story-splitting.md).
Source AC: [epic-3 §Story 3.4](../planning-artifacts/epics/epic-3-inspection-workflow-violation-genesis.md) — the **Canonical Acceptance Criteria** there are the contract all three stacks satisfy; this story implements them in **.NET only**.
Contract doc (extended by this story): [docs/reference/ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md) — populated by 2.9 for `POST /grid/projects`; this story **adds the `POST /grid/inspections` section** (does not re-author the wire format).
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — `domain.inspection` (101–123; `inspector_id` is an **opaque user ref, no FK**), `domain.trade_type`.
Precedent: [2-9 Project list AG Grid SSRM](2-9-project-list-ag-grid-with-server-side-row-model.md) — the .NET parser `FieldMark.Web/Grid/SsrmRequest.cs`, the vendored `ag-grid-panel.js`, and the `{rows,lastRow}` envelope all come from there.

## Split-group context (read first)

- **This is `.a` — the first split story in Epic 3** and the highest-churn-risk one (the 2-9 AG Grid pattern took 5+ review rounds when done as one 3-stack story). Its job is to **settle the inspections SSRM contract in .NET, reviewed clean**, so 3.4b (Django + Go) and 3.4c (parity & DoD) port against a frozen target.
- **`make parity` is NOT a gate here.** It will be red (`POST /grid/inspections` exists only in .NET). Parity is asserted on **3.4c**. Do not touch `fieldmark_py/` or `fieldmark-go/`.
- **Hard dependency chain:** `3.4a → 3.4b → 3.4c`. The reviewer treats the column allowlist, the row projection (esp. the `inspector_name` resolution — Decision 1), the project-scoping mechanism, and the contract-doc section as the **frozen contract** 3.4b copies.
- **Definition of "done" for this story:** all .NET ACs pass, `make test-net` green, the `/grid/inspections` contract-doc section landed. Three-stack invariant is **not** claimed here — that's 3.4c.

## Dependencies

**Done / assumed-in-place:**
- **2.9** — the SSRM machinery: the populated [ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md), the .NET parser `FieldMark.Web/Grid/SsrmRequest.cs` (column/operator allowlists, `(sort, filter, page)` validation, 400 behaviour), the vendored & parameterized `fieldmark_shared/vendor/ag-grid-panel/ag-grid-panel.js` (driven by `data-*` attributes: `data-endpoint`, row-click target), and the AG Grid **Enterprise** `35.3.0` bundle. **Reuse all of it** — this story configures a second grid, it does not rebuild the plumbing.
- **3.1** — `InspectionConfiguration.cs` mapping `domain.inspection` (read). _Blocks this story — it projects inspection rows._
- **1.10** — the shared **dev-user manifest** (`dev-users.json`: Marisol, Pat, Aisha, Ravi, Kenji with identical UUIDs + names across stacks). The cross-stack-clean source for `inspector_name` (Decision 1).
- **2.6** — EntityRail (`#inspection-detail` rail container is the row-click target).
- **2.11** — Project Detail Inspections tab (this grid renders inside it).
- **1.12 / 1.11** — `can(actor, action)`; canonical 403 body + unauthenticated→`/login` redirect.

**⚠️ Forward dependency (note, not a blocker):**
- The row-click wires `htmx.ajax("GET", "/inspections/<id>", { target: "#inspection-detail" })`. The **`GET /inspections/<id>` handler + detail partial land in Story 3.6** (built after 3.4 per the Epic 3 build order). So in this story the row-click is **wired** (the vendored panel's `onRowClicked` targets `#inspection-detail`, the empty EntityRail from 2.6), but clicking returns 404 until 3.6 lands. This mirrors how 2-9's row-click targeted `/projects/<id>`. The grid **data** endpoint (`POST /grid/inspections`) is the core deliverable and is fully testable here. Full row-click→detail E2E is verified when 3.6 lands (and proven cross-stack in 3.6c).

## Story

As any authorized user on the Inspections tab of Project Detail — in the .NET stack —
I want a server-side AG Grid of a project's inspections (filter/sort/pagination all server-side) whose row-click loads the inspection detail into the EntityRail,
So that the **reference implementation** of the inspections SSRM grid (request scoping, column allowlist, manual row projection incl. the `inspector_name` cross-stack-identity resolution, the `{rows,lastRow}` envelope, 400/403 behaviour) is settled and reviewed before it is ported to Django and Go (FR48, FR50, FR51).

**Scope boundary — this story produces, in .NET only:**

- (a) `POST /grid/inspections` JSON endpoint + handler: parse the AG Grid SSRM request (reusing `SsrmRequest.cs`), enforce the **inspections column allowlist** (Decision 2), translate to parameterized SQL `WHERE`/`ORDER BY`/`LIMIT`/`OFFSET` against `domain.inspection` (+ `domain.trade_type` join), **always scoped to the request's `project_id`** (Decision 3), run the data query **and** a matching `COUNT(*)`, and return `{ "rows": [...], "lastRow": N }` with the canonical projection (Decision 1).
- (b) The AGGridPanel container on the Inspections tab pointed at `POST /grid/inspections` via `data-*` attributes (reusing the vendored `ag-grid-panel.js` — **no new shared JS**), with the inspections column defs (incl. `agSetColumnFilter` on `status` and `outcome`) and `onRowClicked` → `#inspection-detail`.
- (c) The **contract-doc section** for `POST /grid/inspections` appended to `ag-grid-ssrm-contract.md`: project-scoping param, column allowlist, set-filter enums (`status`, `outcome`), the row projection + the `inspector_name`/`trade_name` resolution, and that it inherits the shared envelope/casing/error rules.
- (d) .NET tests: SSRM conformance (canonical inspections fixture → shape/casing/`lastRow`), filter/sort/pagination integration, projection column assertions, project-scoping isolation (a project's grid never leaks another project's inspections), malformed-request 400, authz 403/redirect.

**Out of scope:**
- **Django and Go** — Story 3.4b.
- **`make parity`, cross-stack conformance, byte-identical snapshots** — Story 3.4c.
- **`GET /inspections/<id>` + the detail partial** — Story 3.6 (this story only wires the row-click target).
- **Filters/date-range default scoping UI (inspector default = me, etc.)** — Story 3.5 (this story ships the grid + the allowlist those filters use; the *default-filter* logic is 3.5).
- **Schedule / start / complete / cancel** — Stories 3.7–3.10.
- Any `domain.*` schema change (`pg_indexes` zero-diff; .NET migrations are auth-schema only).

---

## ⚠️ Decisions baked into this story

1. **`inspector_name` is resolved from the shared dev-user manifest, not from any auth schema — RECOMMENDED, NEEDS LEAD CONFIRMATION.**
   `domain.inspection.inspector_id` is an opaque UUID (DDL: "opaque user ref; no FK"); there is **no `domain.user` table** and ADR-012 forbids domain→auth foreign keys. This is the **same cross-stack-identity problem 2-9 hit with `pm_name`** — but unlike pm_name (which 2-9 dropped), the inspections grid genuinely needs the inspector's display name.
   Resolving it from each stack's **auth schema** (`dotnet_auth`/`django_auth`/`fiber_auth`) cannot be made byte-identical cross-stack cheaply (different identity stores). The clean, cross-stack-identical source is the **shared dev-user manifest** (Story 1.10's `dev-users.json`) — identical UUIDs→names in every stack by construction.
   **Recommended resolution:** project `inspector_id` from `domain.inspection`, and resolve `inspector_name` via a small in-process lookup over the shared manifest (the same UUIDs the seed runners used). The projection is then byte-identical across stacks because the name source is the shared file, not the auth store.
   **This is the load-bearing decision of the split group.** Confirm before dev: if the Lead prefers (a) project `inspector_id` only and render the raw UUID, or (b) a future `domain`-side identity table, the column changes here and 3.4b/3.4c inherit it. The contract-doc section records the chosen resolution as the frozen contract.

2. **Inspections column allowlist** (filterable/sortable `colId`s — the SQL-injection guard, per 2-9 AC9). `id` is never filterable/sortable.

   | colId | Type | Filter | Source |
   |---|---|---|---|
   | `trade_name` | text | `agTextColumnFilter` | `domain.trade_type.name` (join) |
   | `inspector_name` | text | `agTextColumnFilter` | manifest lookup (Decision 1) — _filtering/sorting on a manifest-derived field: see Dev Note below_ |
   | `scheduled_for` | date | `agDateColumnFilter` | `domain.inspection.scheduled_for` |
   | `status` | set | `agSetColumnFilter` | `{Scheduled, InProgress, Completed, Cancelled}` |
   | `outcome` | set | `agSetColumnFilter` | `{Pass, Fail, Conditional}` + blank/notBlank for NULL |
   | `completed_at` | date | `agDateColumnFilter` | `domain.inspection.completed_at` (nullable) |

   **Set-filter enum allowlists** (values outside → 400, same discipline as 2-9's `{Active,OnHold,Closed}`): `status` ∈ the four states; `outcome` ∈ the three outcomes (NULL handled via `blank`/`notBlank`, not as a set value).

   **Dev Note — `inspector_name` filter/sort:** because the name is manifest-derived (not a SQL column), server-side **filter and sort on `inspector_name` operate on `inspector_id` translated through the manifest** (e.g. resolve the filter text → matching UUIDs → `WHERE inspector_id = ANY(...)`; sort by a name→UUID ordering map). Document this in the contract section so 3.4b implements the identical translation. _(If the Lead picks Decision-1 option (a) raw-UUID, `inspector_name` leaves the allowlist and only `inspector_id` is filter/sortable.)_

3. **`/grid/inspections` is project-scoped; `project_id` is a required request parameter and the WHERE always pins it.** Unlike the global `/grid/projects`, this grid lists one project's inspections. The `project_id` arrives as a request param (query string or a field the panel posts via `data-*`); the handler **always** adds `WHERE project_id = $projectId` to both the data and COUNT queries (it is not a user-supplied `filterModel` entry and cannot be widened by the client). Authz: `can(actor, "inspection.read")` (or `project.read` if that's the canonical grant — confirm against 2.11's registrations) **for that project**; on false → 403, no leakage. **Project-scoping isolation is an explicit test (AC4):** project A's grid never returns project B's inspections even with a permissive `filterModel`.

4. **Reuse the vendored AGGridPanel JS verbatim — no new shared artifact.** 2-9 parameterized `ag-grid-panel.js` via `data-*` (`data-endpoint`, row-click target) precisely so 3.4/4.2 reuse it. This story adds a **container** with `data-endpoint="/grid/inspections"`, the project id, and `data-detail-target="#inspection-detail"`, plus the inspections **column defs** (the column defs are per-grid config passed to the panel, not a new JS module). If the panel needs a minor extension to pass `project_id`, that extension is shared (vendored) and 3.4c asserts all three stacks use the same panel.

5. **`POST /grid/inspections` is a read** — no audit entry, no state change. CSRF posture matches 2-9's `/grid/projects` (read endpoint; see 2-9 AC for the per-stack CSRF stance). Response `Content-Type: application/json`; cell text rendered via AG Grid `textContent` (no `cellRenderer` injecting HTML) — the XSS guard from 2-9 AC carries over.

6. **Manual projection, no AutoMapper (NFR6).** EF Core `IQueryable.Select` to a seven-field record (`id, trade_name, inspector_name, scheduled_for, status, outcome, completed_at`), then the manifest lookup fills `inspector_name`. `status`/`outcome` are PascalCase verbatim from the DDL CHECK enums; `scheduled_for`/`completed_at` serialize as `YYYY-MM-DD` (or `null`).

---

## Acceptance Criteria (.NET)

### AC1 — `POST /grid/inspections` returns the SSRM envelope with the canonical projection
**Given** I am authorized to read the project and the Inspections tab is active
**When** I POST a valid SSRM request (with the project's id) to `/grid/inspections`
**Then** the response is `200 application/json` of shape `{ "rows": [...], "lastRow": N }` where each row contains exactly `id, trade_name, inspector_name, scheduled_for, status, outcome, completed_at` with **snake_case** keys; `status` is PascalCase (`Scheduled`/`InProgress`/`Completed`/`Cancelled`); `outcome` is PascalCase or `null`; dates are `YYYY-MM-DD` or `null`; `inspector_name`/`trade_name` resolve per Decisions 1–2; and `lastRow` equals the total count of rows matching `filterModel` **within this project**.

### AC2 — Server-side filter / sort / pagination (no client compute)
**Given** the grid
**When** a `filterModel` of `status: {filterType:"set", values:["Scheduled"]}` is posted **Then** only `Scheduled` inspections of this project return and `lastRow` is their count.
**When** a `sortModel` of `[{colId:"scheduled_for", sort:"desc"}]` is posted **Then** rows are ordered by `scheduled_for` desc with a deterministic tiebreaker (stable pagination — see Dev Notes).
**When** `startRow:0,endRow:2` then `startRow:2,endRow:4` are posted **Then** disjoint successive blocks return with a stable total `lastRow`.
**And** filtering/sorting on `inspector_name` operates via the manifest→UUID translation (Decision 2 Dev Note).

### AC3 — Malformed request → 400 (allowlist is the injection guard)
**Given** a `filterModel`/`sortModel` with an unknown `colId`, a disallowed operator for the column type, a `status`/`outcome` value outside its enum, a non-`asc`/`desc` sort, or out-of-bounds paging (`endRow ≤ startRow`, `startRow < 0`, `endRow-startRow > 1000`)
**When** posted **Then** `400` with `{ "error": "<message>" }` (not the envelope); **no SQL is built from an unvalidated colId/operator**.

### AC4 — Project-scoping isolation
**Given** projects A and B each with inspections
**When** I POST `/grid/inspections` scoped to A — even with a permissive/empty `filterModel`
**Then** only A's inspections return; B's never appear; `lastRow` counts A only. (Explicit cross-project leakage test.)

### AC5 — Authz: 403 / redirect, no leakage
**Given** a requester without read permission on the project **When** they POST `/grid/inspections` **Then** `403` with the canonical 1.11 body, no entity-state leakage.
**Given** an unauthenticated request **Then** 302/303 → `/login` (1.11).

### AC6 — Grid page wiring + empty state
**Given** the Inspections tab renders **Then** it initializes the vendored AGGridPanel (`rowModelType:'serverSide'`, Enterprise auto-registered, no license key — watermark accepted) against `/grid/inspections`, with the inspections column defs and `onRowClicked` → `htmx.ajax("GET","/inspections/<id>",{target:"#inspection-detail"})`.
**Given** the project has no inspections **When** the datasource returns `rows:[], lastRow:0` **Then** a custom no-rows overlay renders (e.g. "No inspections yet").
_(Row-click navigation is wired but `GET /inspections/<id>` lands in 3.6 — see forward-dependency note.)_

### AC7 — Contract doc section
**Given** [ag-grid-ssrm-contract.md](../../docs/reference/ag-grid-ssrm-contract.md)
**Then** a `POST /grid/inspections` section is appended specifying: the `project_id` scoping param, the column allowlist + set-filter enums (Decision 2), the row projection + `inspector_name`/`trade_name` resolution (Decision 1), and that envelope/casing/`lastRow`/error rules are inherited from the shared format. A canonical inspections request fixture path is defined for the conformance test (consumed by 3.4c).

---

## Task plan
0. Confirm **Decision 1** (`inspector_name` from manifest vs raw-UUID vs future identity table) with Lead — freezes the projection contract.
1. Append the `/grid/inspections` section to `ag-grid-ssrm-contract.md` (AC7) + canonical request fixture.
2. Extend/confirm `SsrmRequest.cs` allowlist config for the inspections columns + set-filter enums (AC2/AC3).
3. `POST /grid/inspections` handler: project-scoped EF Core `IQueryable.Select` projection (+ `trade_type` join) + COUNT; manifest lookup for `inspector_name` + the name→UUID filter/sort translation (AC1/AC2/AC4).
4. Inspections AGGridPanel container + column defs on the Inspections tab; `onRowClicked` → `#inspection-detail`; empty-state overlay (AC6).
5. Tests: conformance (fixture), filter/sort/pagination, projection columns, project-scoping isolation, 400, 403/redirect (AC1–AC6).
6. `make test-net` green.

## Definition of done (this story)
- [ ] AC1–AC7 pass in .NET.
- [ ] `make test-net` green.
- [ ] `/grid/inspections` contract section + canonical fixture landed; **Decision 1 ratified and recorded** in the contract doc.
- [ ] Reviewed clean. **Only then does 3.4b start.**
- [ ] _Not claimed here:_ `make parity`, cross-stack conformance, snapshots → **Story 3.4c**.

## Sign-off / contract handoff to 3.4b
_To be completed at review:_
- **Endpoint:** `POST /grid/inspections` (project-scoped via `project_id` request param; WHERE always pins it)
- **Row projection (frozen):** `id, trade_name, inspector_name, scheduled_for, status, outcome, completed_at` (snake_case keys; `status`/`outcome` PascalCase enums; dates `YYYY-MM-DD`/null)
- **`inspector_name` resolution (frozen — Decision 1):** _record the ratified choice_
- **Column allowlist + set-filter enums (frozen):** per Decision 2
- **Row-click target:** `#inspection-detail` (handler lands in 3.6)
- **Inherited from 2.9:** `{rows,lastRow}` envelope, casing rule, 400 error body, read/no-audit posture, vendored `ag-grid-panel.js`
