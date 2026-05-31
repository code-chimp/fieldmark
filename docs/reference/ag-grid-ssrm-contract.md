# AG Grid Server-Side Row Model — Wire Format Contract

> **Status:** populated by Story 2.9, 2026-05-30.
> This document supersedes the skeleton scaffolded by Epic 1 retrospective action item A4.

This document is the **single source of truth** for the AG Grid SSRM wire format used by every grid endpoint in FieldMark (`POST /grid/projects`, `POST /grid/inspections`, `POST /grid/violations`, etc.). Each stack implements the contract natively against its framework; per-stack conformance tests assert alignment.

See the root [CLAUDE.md](../../CLAUDE.md) **Cross-Stack Architecture Principle** for why this lives as documentation rather than as a shared codec.

---

## AG Grid Edition Note

FieldMark uses **AG Grid Enterprise** (UMD bundle `ag-grid-enterprise.min.js`, version `35.3.0`) to enable the true **Server-Side Row Model** (`rowModelType: 'serverSide'`). The Enterprise bundle includes Community and auto-registers all modules (`ServerSideRowModelModule`, Set Filter, etc.) — no individual `ModuleRegistry.registerModules` call is needed.

The demo runs **without a license key**. AG Grid renders an "unlicensed" watermark on the grid; this is a deliberate, accepted tradeoff to showcase Enterprise functionality in the teaching artifact. Do not add a license key. Do not fall back to the Community Infinite Row Model.

The wire contract below is what the `IServerSideDatasource.getRows(params)` sends and receives.

---

## Request Shape

The AG Grid SSRM datasource calls `getRows(params)` and POSTs `params.request` as JSON to the grid endpoint. The request body is AG Grid's `IServerSideGetRowsRequest`, using AG Grid's camelCase vendor vocabulary — these keys are **not** subject to FieldMark's snake_case-on-the-wire rule (they are vendor wire, like HTTP header names).

```jsonc
{
  "startRow": 0,          // inclusive, 0-based
  "endRow": 100,          // exclusive; block size = endRow - startRow (AG Grid default: 100)
  "sortModel": [          // ordered; first entry is primary sort
    { "colId": "compliance_score", "sort": "asc" }   // sort ∈ {"asc","desc"}
  ],
  "filterModel": {        // keyed by colId; empty object = no filter
    "status":           { "filterType": "set",    "values": ["Active", "OnHold"] },
    "compliance_score": { "filterType": "number", "type": "greaterThan", "filter": 70 },
    "code":             { "filterType": "text",   "type": "contains",    "filter": "BLDG" },
    "start_date":       { "filterType": "date",   "type": "inRange", "dateFrom": "2026-01-01", "dateTo": "2026-12-31" }
  }
  // AG Grid may also send: rowGroupCols, valueCols, pivotCols, pivotMode, groupKeys
  // The handler IGNORES these (no grouping/pivot in this story's flat grid).
}
```

### Column allowlist

`colId` values in `sortModel` and `filterModel` must be one of the following snake_case projected column names:

| colId | Type | Filter |
|---|---|---|
| `code` | text | `agTextColumnFilter` |
| `name` | text | `agTextColumnFilter` |
| `status` | set | `agSetColumnFilter` (Enterprise) |
| `compliance_score` | number | `agNumberColumnFilter` |
| `start_date` | date | `agDateColumnFilter` |
| `target_completion_date` | date | `agDateColumnFilter` |

A `colId` not in this allowlist → **400**. `id` is never a filterable/sortable column.

### Allowed filter operators per column type

**Set filter** (`status`): `filterModel.status = { "filterType": "set", "values": [...] }`. The server maps to `WHERE status = ANY($1)`. Every value in `values` must be ∈ `{Active, OnHold, Closed}` (the DDL CHECK enum). Any other value → 400. An empty `values: []` means "match nothing" (return zero rows, `lastRow: 0`), per AG Grid Set Filter semantics.

**Text filter** (`code`, `name`): `filterType: "text"`. Allowed `type` values: `equals`, `notEqual`, `contains`, `notContains`, `startsWith`, `endsWith`, `blank`, `notBlank`. Unknown `type` → 400.

**Number filter** (`compliance_score`): `filterType: "number"`. Allowed `type` values: `equals`, `notEqual`, `greaterThan`, `greaterThanOrEqual`, `lessThan`, `lessThanOrEqual`, `inRange` (uses `filter` + `filterTo`), `blank`, `notBlank`. Unknown `type` → 400.

**Date filter** (`start_date`, `target_completion_date`): `filterType: "date"`. Allowed `type` values: `equals`, `notEqual`, `greaterThan` (after), `lessThan` (before), `inRange` (uses `dateFrom` + `dateTo`), `blank`, `notBlank`. Date values are `YYYY-MM-DD` strings. Unknown `type` → 400.

### Sort direction values

`sort` must be `"asc"` or `"desc"`. Any other value → 400. Sort is allowed on every projected column. `sortModel` is honored in array order (first entry is primary sort).

### Pagination bounds

- `startRow ≥ 0`
- `endRow > startRow`
- `endRow - startRow ≤ 1000` (hard server cap — exceeding it → 400, **not** a silent clamp)

The default AG Grid block size is 100.

### Stable pagination ordering

To guarantee consistent block-to-block pagination, every `ORDER BY` clause appends `, id ASC` as a deterministic tiebreaker. When `sortModel` is empty, the default order is `code ASC, id ASC`.

---

## Response Shape

```jsonc
{
  "rows": [
    {
      "id": "f9e4c1b2-3d4e-5f6a-7b8c-9d0e1f2a3b4c",  // UUID string
      "code": "RIVERSIDE-01",
      "name": "Riverside Substation Upgrade",
      "status": "Active",                              // PascalCase per DDL CHECK — see note
      "compliance_score": 71,                         // integer 0–100
      "start_date": "2026-01-15",                     // YYYY-MM-DD
      "target_completion_date": null                  // YYYY-MM-DD or null
    }
  ],
  "lastRow": 247  // total rows matching filterModel (see semantics below)
}
```

### Casing rule (explicit)

Envelope keys (`rows`, `lastRow`) and request keys (`startRow`, `endRow`, `sortModel`, `filterModel`, etc.) are AG Grid's vendor vocabulary — **camelCase**. Row-object keys are domain wire — **snake_case** per the project's standard rule. `lastRow` is deliberately camelCase and is documented here as the one envelope-level exception to the snake_case rule.

### `status` values are PascalCase

`Active`, `OnHold`, `Closed` — matching the `domain.project.status` DDL CHECK constraint and Story 2.1 enum mappings (`ProjectStatus` in .NET, `ProjectStatus.TextChoices` in Django, `enums.ProjectStatus` in Go). The architecture.md SCREAMING_SNAKE example is superseded by the DDL.

### `lastRow` semantics

The server **always** returns the total count of rows matching `filterModel` — a `SELECT COUNT(*)` over the same `WHERE` clause as the data query. `lastRow` lets AG Grid size the scrollbar and know when to stop requesting blocks. The server never returns `-1` here because the count is always knowable.

The `IServerSideDatasource.getRows(params)` maps the envelope to:
```js
params.success({ rowData: resp.rows, rowCount: resp.lastRow });
```

---

## Row Projection Rules

Rows are **manually projected** per NFR6 — no AutoMapper, no Mapster (rejected NET-MAPSTER), no generic mapper. The canonical `POST /grid/projects` projection is seven columns read directly from `domain.project` with **no join**:

| Column | Wire key | Type | Notes |
|---|---|---|---|
| `id` | `id` | UUID string | Never filterable/sortable by colId |
| `code` | `code` | string | |
| `name` | `name` | string | |
| `status` | `status` | string | PascalCase: `Active` / `OnHold` / `Closed` |
| `compliance_score` | `compliance_score` | integer | 0–100; `0` serializes as `0` (not null, not blank) |
| `start_date` | `start_date` | `YYYY-MM-DD` string | |
| `target_completion_date` | `target_completion_date` | `YYYY-MM-DD` or `null` | Optional column; null is valid |

**`pm_name` is omitted.** `domain.project` has no project-manager column; there is no `domain.user` table; ADR-012 forbids domain→auth foreign keys. The only people-link is `domain.project_inspector` (inspectors, not a PM). If a PM concept is introduced (future schema change, infra-owned), the column is added here + three handlers + three conformance tests under the Change Procedure below.

Per-stack implementations:
- **.NET:** `IQueryable<Project>.Select(p => new ProjectGridRow { ... })` — explicit named projection
- **Django:** `.values("id", "code", "name", "status", "compliance_score", "start_date", "target_completion_date")` followed by date/UUID string formatting
- **Go:** explicit scan into `ProjectGridRow` struct with `json:"snake_case"` tags

Do not `SELECT *` and serialize the whole entity — that would leak `description`, `actual_closed_at`, `updated_at`, `created_at`.

---

## Error Behaviour

Malformed requests return **HTTP 400** with a JSON body:

```jsonc
{ "error": "<human-readable message>" }
```

This is **not** the `{rows, lastRow}` envelope. Triggers:

| Condition | Message |
|---|---|
| Invalid JSON body | `"invalid request body"` |
| Unknown `colId` in `sortModel` or `filterModel` | `"unknown column: <colId>"` |
| Disallowed filter operator for the column type | `"invalid operator '<type>' for column '<colId>'"` |
| `status` filter value outside `{Active,OnHold,Closed}` | `"invalid status value: <value>"` |
| `endRow ≤ startRow` | `"endRow must be greater than startRow"` |
| `startRow < 0` | `"startRow must be >= 0"` |
| `endRow - startRow > 1000` | `"page size exceeds maximum of 1000"` |
| Non-`asc`/`desc` sort direction | `"invalid sort direction: <value>"` |

A handler **must not** build SQL from an unvalidated `colId` or operator — the allowlist is the SQL-injection guard. No `colId` is ever interpolated into SQL; the allowlist maps each `colId` to a known-safe column token.

### CSRF posture for `POST /grid/projects`

`POST /grid/projects` is a **read** — idempotent query, no state change, no audit entry. Architecture note: reads are not audited (PRD). CSRF posture:

- **.NET:** exempt from antiforgery via `[IgnoreAntiforgeryToken]` with documented rationale (read-only, no side effects). The endpoint succeeds without an antiforgery token.
- **Django:** `@csrf_exempt` with documented rationale (read-only, no side effects). The AG Grid datasource does not send `csrfmiddlewaretoken`. `@require_POST` is the method gate.
- **Go:** no CSRF middleware (ADR-012); no change.

---

## Per-Stack Native Implementations

No shared codec, no generated stubs. Each stack is idiomatic.

- **.NET** — `FieldMark.Web/Pages/Grid/ProjectsGridModel.cs` (or `Pages/Grid/Projects.cshtml.cs`); SSRM request/validator in `FieldMark.Web/Grid/SsrmRequest.cs`. EF Core `IQueryable.Select` projection.
- **Django** — `fieldmark_py/grid/views.py` + `grid/urls.py`; SSRM parser in `grid/ssrm.py`. ORM `.values()` projection.
- **Go** — `fieldmark-go/internal/web/handlers/grid_projects_handler.go`; SSRM parser in `internal/web/ssrm.go`. `pgx` direct query, scan into struct.

Every grid handler file and the AGGridPanel init JS carry a top-of-file comment referencing:
`docs/reference/ag-grid-ssrm-contract.md`

---

## Conformance Test Contract

Each stack ships a conformance test that:

1. Issues a request body equivalent to the canonical fixture `docs/reference/fixtures/ssrm-canonical-request.json` against `POST /grid/projects`. Tests use the fixture data inline or by loading the file — either approach satisfies the contract as long as the request body matches the fixture exactly.
2. Asserts the response envelope shape (`rows` array, `lastRow` integer).
3. Asserts row-object keys are exactly `{id, code, name, status, compliance_score, start_date, target_completion_date}` — no extra columns.
4. Asserts key casing: row-object keys are `snake_case`; envelope key `lastRow` is camelCase.
5. Asserts `lastRow` is a non-negative integer equal to the total count matching the filter.

Per-stack test locations (as of Story 2.9):
- **.NET:** `FieldMark.Tests.Web/Pages/ProjectsListPageTests.cs` — conformance tests run against the real DB via `WebApplicationFactory` + `PostgresFixture`. Test names: `GridProjects_CanonicalFixture_ReturnsValidEnvelope`, `GridProjects_CanonicalFixture_RowKeysAreSnakeCase`, `GridProjects_LastRowEqualsTotalMatchingFilter`, `GridProjects_NullTargetDateSerializesAsNull`.
- **Django:** `fieldmark_py/grid/tests.py` (parser unit tests), `fieldmark_py/grid/test_grid_views.py` (view-level authz + 400 tests). Integration conformance tests (live DB) are in `fieldmark_py/projects/tests/test_project_list.py` (view-level) and run with `pytest -m integration` when the domain DB is available.
- **Go:** `fieldmark-go/internal/web/handlers/grid_projects_handler_test.go` (authz + 400 unit tests), `fieldmark-go/internal/web/ssrm_test.go` (parser unit tests). Integration conformance tests (live DB) are run with `go test -tags=integration ./...`.

---

## Change Procedure

Adding a projected column or filterable field:

1. Edit this document — add the column to the allowlist table and the row projection table.
2. Add the column to all three stack handlers (update the `SELECT`/`.values()`/scan).
3. Add the column to all three conformance tests (assert the new key appears, old tests remain green).
4. Run `make parity` — route diff must remain clean; `pg_indexes` diff must remain zero.
5. Open a PR that touches this doc + three handlers + three conformance tests atomically.
