# Story 3.7a: Schedule Inspection — .NET reference

Status: ready-for-dev

Epic: 3 — Inspection Workflow & Violation Genesis
Split group: **3.7 Schedule Inspection** (a/b/c) — this is **`.a` the .NET reference**. See [docs/how-to/cross-stack-story-splitting.md](../../docs/how-to/cross-stack-story-splitting.md).
Source AC: [epic-3 §Story 3.7](../planning-artifacts/epics/epic-3-inspection-workflow-violation-genesis.md) — the **Canonical Acceptance Criteria** there are the contract all three stacks satisfy; this story implements them in **.NET only**.
Canonical DDL: [docker/postgres/init/010_domain_tables.sql](../../docker/postgres/init/010_domain_tables.sql) — `domain.inspection` (101–123), `domain.audit_entry`.

## Split-group context (read first)

- **This is `.a`.** Its job is to **establish the reference**: the route shape, the entity factory, the canonical write-handler flow, the form-contract field names, the markup, and the contract-doc updates — in .NET, reviewed clean — so that **3.7b** (Django + Go port) and **3.7c** (parity & DoD) have a settled target to match.
- **`make parity` is NOT a gate on this story.** It will be red (the route exists only in .NET). Parity is asserted on **3.7c**. Do not add Django/Go code here.
- **Hard dependency chain:** `3.7a → 3.7b → 3.7c`. 3.7b does not start until this story is reviewed clean. The reviewer should treat any field-name / route / audit-string / markup decision here as the **frozen contract** 3.7b copies.
- **Definition of "done" for this story:** all .NET ACs below pass, `make test-net` green, contract artifacts (form-contract field list, audit-action entry) landed. The three-stack invariant is **not** claimed here — that's 3.7c.

## Dependencies

**Done / assumed-in-place:**
- **3.1** — `InspectionConfiguration.cs` + `FindingConfiguration.cs` mapping `domain.inspection` / `domain.finding` (read **and** insert-capable for inspection). _If 3.1 is not yet `done`, this story is blocked — it writes an inspection row._
- **2.8** — the canonical **.NET write-handler shape** (authorize → `IDbContextTransaction` → load → entity method → `IAuditAppender.Append(...)` → `SaveChanges` → commit) and the **422/InlineAlert re-render** pattern. `Project.create(...)` is the precedent entity factory; `project.schedule_inspection(...)` is the next.
- **2.2** — `domain.audit_entry` mapping + `IAuditAppender.Append(...)` called **inside the caller's open transaction**.
- **3.4** — **Inspections list AG Grid (the `#inspection-list` target).** Built **before** this story (Epic 3 reordered 2026-06-03 so infrastructure precedes transitions). The schedule success path re-renders the **real** `#inspection-list` target that 3.4 establishes — so AC3/AC5 are **DOM-verifiable** here, not emit-per-contract. _If 3.4 (at least 3.4a) is not yet `done`, this story is blocked on the list target._
- **2.11** — Project Detail anchor screen: the **Inspections tab** target on `/projects/<id>` and the TabStrip. This story renders into the Inspections tab.
- **2.5 / 2.4** — ComplianceTile + InlineAlert wrappers (InlineAlert used for the 409 path).
- **1.12 / 1.11** — `can(actor, action)` primitive; canonical 403 body + unauthenticated→`/login` redirect.

## Story

As a Compliance Officer or Administrator,
I want to schedule an Inspection on a Project for a specific trade, inspector, and scheduled time — in the .NET stack —
So that the **reference implementation** of the schedule flow (route, entity factory, canonical transaction, audit append, OOB list re-render, 403/409 paths, form-contract field names) is settled and reviewed before it is ported to Django and Go (FR16).

**Scope boundary — this story produces, in .NET only:**

- (a) A **domain factory method** `Project.ScheduleInspection(TradeType trade, Guid inspectorId, DateTimeOffset scheduledFor)` on the `Project` entity that returns a new `Inspection` in `status = "Scheduled"`, raising a **typed domain exception** when the trade is not in the project's scope, the inspector is not assigned to the project, or `scheduledFor` is in the past (Decision 3).
- (b) Two routes: `GET /projects/{id}/inspections/new` (returns the **inline schedule-form fragment**) and `POST /projects/{id}/inspections/` (runs the canonical mutating flow). _Note: the epic wrote `POST /projects/:id/inspections/`; the `GET` reveal mirrors 2.12's two-step form pattern._
- (c) The **success response composition**: re-rendered `#inspection-list` (main, the request's `hx-target`) **plus** OOB `#audit-log` row fragment. **No `#compliance-tile` OOB** — scheduling does not affect score (Decision 2).
- (d) The **403** (zero OOB) and **409** (originating form partial re-rendered with current state + InlineAlert, zero OOB) responses.
- (e) The **inline schedule form** (`#inspection-schedule-form`) revealed by the "Schedule Inspection" button: trade `<select>` (project's assigned trades), inspector `<select>` (project's assigned inspectors), `scheduled_for` datetime input — with server-side validation.
- (f) **Contract artifacts** (consumed by 3.7b/3.7c): the frozen **form-contract field-name list** recorded in this story's Sign-off (per CLAUDE.md form-contract corollary), and the **`InspectionScheduled`** audit-action entry confirmed/added in [docs/reference/audit-actions.md](../../docs/reference/audit-actions.md).
- (g) .NET unit + integration tests (AC-mapped).

**Out of scope:**
- **Django and Go** — Story 3.7b. Do not touch `fieldmark_py/` or `fieldmark-go/`.
- **`make parity`, cross-stack conformance tests, byte-identical snapshots, E2E, NFR timing** — Story 3.7c.
- **The Inspections AG Grid itself** — Story 3.4 (this story re-renders the `#inspection-list` target that 3.4 establishes).
- **Inspection detail / start / complete / cancel** — Stories 3.7–3.10.
- **Compliance-score recompute** — scheduling has no score impact.
- Any `domain.*` schema change (`pg_indexes` zero-diff for .NET migrations — auth schema only).

---

## ⚠️ Decisions baked into this story

1. **Canonical-truth corrections over epic prose.** The epic AC loosely wrote `status='SCHEDULED'` and `scheduled_at`. The **DDL is canonical**: status value is PascalCase **`'Scheduled'`** (CHECK constraint `{Scheduled, InProgress, Completed, Cancelled}`), and the column is **`scheduled_for`** (not `scheduled_at`). This story uses `Scheduled` / `scheduled_for`. 3.7b/3.7c inherit these verbatim. (The 3.4 grid's row field is a *projection alias*; it must also read `scheduled_for`.)
2. **Two-region success response, not three.** Scheduling does not change compliance score, so the response is **main `#inspection-list` + OOB `#audit-log`** only — **no `#compliance-tile`**. This is a deliberate divergence from the full three-region pattern and must be asserted (AC3): the success body contains exactly these two regions and **zero** `#compliance-tile`.
3. **Typed domain exception → HTTP 409.** `Project.ScheduleInspection(...)` raises a **named** exception (`InvalidInspectionScheduleException`) — not a bare `ArgumentException` — so the handler catches only it and maps to 409. The user-visible messages are part of the contract 3.7b matches: `"Trade is not in this project's scope"` / `"Inspector is not assigned to this project"` / `"Scheduled time must be in the future"`.
4. **Two-step reveal mirrors 2.8/2.12.** The present-state "Schedule Inspection" button `hx-get`s `/projects/{id}/inspections/new` into the Inspections-tab action slot; the returned `<form id="inspection-schedule-form">` `hx-post`s `/projects/{id}/inspections/`, `hx-target="#inspection-list"`, `hx-swap="innerHTML"`, with a no-JS fallback (`method="post" action="…"`) and the antiforgery token.
5. **Authorization:** `inspection.schedule` registered to `{ADMIN, COMPLIANCE_OFFICER}` (epic: "Compliance Officer or Administrator"). If that registration does not yet exist in the .NET authz wiring, this story adds it (and the same grant is mirrored by 3.7b). The authz tests assert those two roles permitted, the other three 403.
6. **`InspectionScheduled` audit action.** Confirm it exists in `docs/reference/audit-actions.md` + `audit-actions.json` + the .NET audit-action enum. If absent, add it via the canonical change procedure (doc first, then the .NET enum) — and note in Sign-off that 3.7b must add it to the Django + Go enums and 3.7c's conformance test must cover it.

---

## Acceptance Criteria (.NET)

### AC1 — Schedule-form reveal (`GET`)
**Given** I am authenticated as `ADMIN` or `COMPLIANCE_OFFICER` and `can(actor, "inspection.schedule")` is true
**When** I click "Schedule Inspection" on the Inspections tab of `/projects/{id}`
**Then** HTMX fires `hx-get="/projects/{id}/inspections/new"` into the tab action slot
**And** the response is a fragment containing `<form id="inspection-schedule-form">` with: a `<label>`+`<select name="trade_type_id">` populated from the project's assigned trades, a `<label>`+`<select name="inspector_id">` from the project's assigned inspectors, a `<label>`+`<input type="datetime-local" name="scheduled_for" required>`, the antiforgery token, a `<button type="submit">`, and a Cancel control that clears the slot
**And** the form carries `hx-post="/projects/{id}/inspections/"`, `hx-target="#inspection-list"`, `hx-swap="innerHTML"`, `hx-disabled-elt="find button[type=submit]"`, and the no-JS fallback `method="post" action="/projects/{id}/inspections/"`.

**Given** I lack the permission → the "Schedule Inspection" button is **absent** (trichotomy); a direct `GET` to the reveal endpoint by an unauthorized user returns **403** (AC4 shape).

### AC2 — `Project.ScheduleInspection(...)` domain factory
**Given** the `Project` entity
**When** `ScheduleInspection(trade, inspectorId, scheduledFor)` is called with a trade in scope, an assigned inspector, and a future time
**Then** it returns a new `Inspection` with `status = "Scheduled"`, `scheduled_for = scheduledFor`, `project_id`, `trade_type_id`, `inspector_id` set, and `started_at/completed_at/outcome` null
**And** it raises `InvalidInspectionScheduleException` with the Decision-3 message when: trade not in `project_trade_scope`; inspector not in `project_inspector`; `scheduledFor <= now`.
**And** it is a **pure domain method** — no DB, no audit, no transaction inside the entity (same discipline as `Project.create`).
**And** unit tests cover the happy path + each of the three raise conditions.

### AC3 — `POST` happy path: canonical flow + two-region response
**Given** authorized + valid input
**When** `POST /projects/{id}/inspections/` runs
**Then** the handler executes **in order**: authorize (`can`) → open **exactly one** `IDbContextTransaction` → load `Project` (with trade scope + inspectors) → `project.ScheduleInspection(...)` → insert the `domain.inspection` row → `IAuditAppender.Append(action: "InspectionScheduled", actor, entityType: "Inspection", entityId: <new id>, projectId: <id>, before: null, after: <inspection state>, metadata: null)` **in the same transaction** → `SaveChanges` → commit
**And** the response body contains **exactly two regions**: main re-rendered `#inspection-list` (`hx-swap="innerHTML"` target) showing the new inspection, **plus** OOB `#audit-log` AuditRow with `hx-swap-oob="afterbegin"`
**And** the body contains **no `#compliance-tile`** (assert absent) and HTTP status is **200** with exactly one round trip.

### AC4 — `403` unauthorized: zero OOB, no leakage
**Given** any of `{EXECUTIVE, INSPECTOR, SITE_SUPERVISOR}` (or no-role user)
**When** they `GET` the reveal or `POST` the schedule endpoint
**Then** HTTP **403** with the canonical 1.11 body, no entity state leaked, and **zero** OOB regions (`hx-swap-oob` absent).
**And** an unauthenticated request → 302/303 `/login` (1.11), unchanged.

### AC5 — `409` invalid input: originating form + InlineAlert, zero OOB
**Given** authorized but invalid input (trade out of scope / inspector not assigned / past time)
**When** `POST` runs and `ScheduleInspection` raises
**Then** HTTP **409**, the response re-renders the `#inspection-schedule-form` fragment with the submitted values preserved + an **InlineAlert** (`role="alert"`) carrying the exception message + `aria-invalid="true"` / `aria-describedby` on the offending field(s)
**And** **no** inspection row is written (transaction rolled back), **no** audit entry, and **zero** OOB regions.

### AC6 — Audit + transaction integrity (data layer)
**Given** a successful schedule
**When** I query `domain.audit_entry`
**Then** exactly one row exists with `action="InspectionScheduled"`, `entity_type="Inspection"`, `entity_id=<new inspection id>`, `project_id=<id>`, `before_state=null`, `after_state` = the inspection's JSONB state.
**Given** a handler abort (e.g. forced exception after the inspection insert) → the transaction rolls back: no inspection row, no audit row (integration test against the real-DB harness, per Epic 1 retro A3).

---

## Task plan

0. **Confirm Story 3.4 (Inspections list, `#inspection-list` target) is `done`** — this story re-renders that real target (Epic 3 reorder resolved the earlier sequencing concern).
1. Add/confirm `inspection.schedule` → `{ADMIN, COMPLIANCE_OFFICER}` in .NET authz registration; `Can` predicate for the button trichotomy.
2. Confirm/add `InspectionScheduled` to `audit-actions.md` + `.json` + .NET enum (Decision 6).
3. `Project.ScheduleInspection(...)` + `InvalidInspectionScheduleException` (+ unit tests, AC2).
4. `GET /projects/{id}/inspections/new` handler + `#inspection-schedule-form` Razor partial (AC1).
5. `POST /projects/{id}/inspections/` handler — canonical flow, two-region response composition (AC3), 409 path (AC5).
6. `#inspection-list` re-render partial + OOB `#audit-log` AuditRow fragment.
7. Integration tests: happy path, 403, 409, rollback (AC3–AC6).
8. `make test-net` green.

## Definition of done (this story)
- [ ] AC1–AC6 pass in .NET.
- [ ] `make test-net` green.
- [ ] Form-contract field names frozen and recorded in Sign-off (`trade_type_id`, `inspector_id`, `scheduled_for`).
- [ ] `InspectionScheduled` present in `audit-actions.md` + `.json` + .NET enum.
- [ ] Reviewed clean. **Only then does 3.7b start.**
- [ ] _Not claimed here:_ `make parity`, cross-stack conformance, snapshots, E2E, NFR timing → **Story 3.7c**.

## Sign-off / contract handoff to 3.7b
_To be completed at review:_
- **Routes:** `GET /projects/{id}/inspections/new`, `POST /projects/{id}/inspections/`
- **Form field names (frozen):** `trade_type_id`, `inspector_id`, `scheduled_for`
- **HTMX contract:** `hx-target="#inspection-list"`, `hx-swap="innerHTML"`; reveal slot id `#inspection-schedule-form`
- **Audit action:** `InspectionScheduled`
- **Exception messages (frozen):** `"Trade is not in this project's scope"` / `"Inspector is not assigned to this project"` / `"Scheduled time must be in the future"`
- **Response contract:** success = main `#inspection-list` + OOB `#audit-log` (afterbegin), **no** `#compliance-tile`; 403 = zero OOB; 409 = form + InlineAlert, zero OOB.
