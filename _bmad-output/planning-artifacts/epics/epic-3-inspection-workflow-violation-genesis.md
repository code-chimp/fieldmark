# Epic 3: Inspection Workflow & Violation Genesis

Compliance Officer schedules Inspections; Inspector starts and completes them with findings; Fail findings auto-open Violations in the same DB transaction. Server-side compliance rule evaluation lit up here for the first time; score recomputes on inspection completion and violation opening.

## Story 3.1: Map `domain.inspection` and `domain.finding` into each stack's data layer

As a developer building Inspection features,
I want each stack's data layer to read and write `domain.inspection` and `domain.finding` against the existing canonical DDL,
So that subsequent stories can implement Inspection state transitions.

**Acceptance Criteria:**

**Given** the .NET stack
**When** I inspect `FieldMark.Data/Configuration/`
**Then** `InspectionConfiguration.cs` and `FindingConfiguration.cs` map the tables to `domain` schema with snake_case + enum-string conventions.

**Given** the Django stack
**When** I inspect `inspections/models.py`
**Then** `Inspection` and `Finding` models declare `Meta.managed = False` and target `domain.inspection` / `domain.finding` (FR-mapping only — DDL untouched).

**Given** the Go stack
**When** I inspect `internal/data/inspectionstore.go`
**Then** `InspectionStore` interface with `pgx`-backed implementation reads Inspections and their Findings (eager-loaded via SQL JOIN).

**Given** all mappings exist
**When** I run `make parity`
**Then** `pg_indexes` for `domain.*` is unchanged
**And** per-stack domain unit tests load a seeded Inspection with its Findings and assert round-trip fidelity.

---

## Story 3.2: Map `domain.violation` into each stack's data layer (write-capable for auto-open)

As a developer implementing Story 3.9 (auto-open Violations on Fail finding),
I want each stack's data layer to INSERT into `domain.violation` within the inspection-completion transaction,
So that Fail findings spawn Violations atomically per FR20.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its data layer
**Then** `ViolationStore` (Go) / `ViolationConfiguration.cs` (.NET) / `violations/models.py` Meta.managed=False (Django) is in place with read + insert capability
**And** the inserted row carries the canonical fields: `finding_id`, `severity`, `due_date`, `status='OPEN'`, `created_at`.

**Given** the Violation aggregate has no entity methods at this story
**When** I grep each stack's domain layer
**Then** only `Violation.open_from_finding(finding, severity, due_date_offset_days)` exists (the auto-open factory); assign/void/etc. land in Epic 4.

**Given** `make parity`
**When** I run it
**Then** `pg_indexes` is unchanged and no routes touch `/violations/*` yet (route parity stays clean).

---

## Story 3.3: Implement compliance rule engine and scoring helper

As a developer implementing any state transition that affects compliance,
I want a single per-stack pure function that recomputes a Project's compliance score from its current Violations and Inspections,
So that FR34 (server-side rule evaluation) and FR35 (same-transaction recompute) are mechanically satisfied by every handler that needs them.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its compliance module (.NET `Domain/Entities/Project.RecomputeComplianceScore`; Django `compliance/scoring.py`; Go `internal/domain/compliance_rules.go`)
**Then** a function `recompute_compliance_score(project, violations, completed_inspections, rules) -> int` exists that returns a value in `[0, 100]` (`CHECK` constraint already enforces this) (FR34).

**Given** the function
**When** it is called
**Then** it reads severity weights and due-offset values from the reference data (`domain.compliance_rule`), not from hardcoded constants (FR38 dynamic eval).

**Given** identical input fixtures (same Project + same Violations + same Inspections + same rule parameters)
**When** the function runs on each of the three stacks
**Then** the returned score is identical across stacks (cross-stack determinism — verified by a shared test fixture).

**Given** unit tests
**When** I run `make test-net`, `make test-django`, `make test-go`
**Then** each stack has unit tests covering: all-Pass Inspections + zero Violations → 100; one Open Critical Violation → < 50; one Resolved Violation does not lower score; Voided Violations do not affect score (FR26 invariant).

**Given** the function does not own transactional concerns
**When** it is called from a handler
**Then** the handler is responsible for persisting the new score within its surrounding transaction (FR57).

---

## Story 3.4: Schedule Inspection (Compliance Officer / Administrator)

As a Compliance Officer or Administrator,
I want to schedule an Inspection on a Project for a specific trade, inspector, and scheduled time,
So that Inspections exist for inspectors to perform (FR16).

**Acceptance Criteria:**

**Given** I am authorized (`inspection.schedule`)
**When** I click "Schedule Inspection" on the Inspections tab of Project Detail
**Then** an inline form expands (`#corrective-action-form`-style pattern — separate id `#inspection-schedule-form`) capturing trade (select from project's assigned trades), inspector (select from project's assigned inspectors), scheduled time (datetime).

**Given** a valid submission
**When** the handler runs `POST /projects/<id>/inspections/`
**Then** the canonical flow executes: authorize → begin txn → load Project → call `project.schedule_inspection(trade, inspector, when)` (entity factory) → write `domain.inspection` row with `status='SCHEDULED'` → append `AuditEntry(action="InspectionScheduled", ...)` → commit
**And** the response body re-renders `#inspection-list` plus OOB `#audit-log` row (no `#compliance-tile` update — scheduling doesn't affect score).

**Given** an unauthorized request
**When** the handler resolves authz
**Then** HTTP `403` without state leakage (FR7, FR56).

**Given** invalid input (trade not in project scope, inspector not assigned, scheduled-time in the past)
**When** the entity method raises
**Then** HTTP `409` with the originating form partial + InlineAlert (UX-DR22, FR55).

**Given** `make parity`
**When** I run it
**Then** `POST /projects/:id/inspections/` exists on all three stacks.

---

## Story 3.5: Inspection list AG Grid with SSRM endpoint

As any authorized user on the Inspections tab of Project Detail,
I want to see inspections with server-side filter/sort/pagination,
So that I can navigate inspection history at scale (FR48).

**Acceptance Criteria:**

**Given** the Inspections tab is active on `/projects/<id>`
**When** the tab content renders
**Then** AGGridPanel initializes against `POST /grid/inspections` with the project id passed as a request parameter
**And** the response is `{ "rows": [...], "lastRow": N }` with rows containing `id`, `trade_name`, `inspector_name`, `scheduled_at`, `status`, `outcome`, `completed_at` (FR49).

**Given** I click a row
**When** the row-click handler fires
**Then** `htmx.ajax("GET", "/inspections/<id>", { target: "#inspection-detail" })` loads inspection detail into EntityRail (FR50)
**And** the grid does not own the detail rendering (FR51).

**Given** `make parity`
**When** I run it
**Then** `POST /grid/inspections` exists identically on all three stacks.

---

## Story 3.6: Inspection list filters and date-range scoping

As an Inspector,
I want to filter Inspections by status and date range and see only inspections assigned to me by default,
So that my queue is meaningful (FR21).

**Acceptance Criteria:**

**Given** I am Inspector role
**When** I land on the Inspections tab
**Then** the default filter `inspector_id = me` is applied server-side
**And** the grid's filter UI exposes status (`SCHEDULED`/`IN_PROGRESS`/`COMPLETED`/`CANCELLED`), outcome (`PASS`/`FAIL`/`CONDITIONAL`/`NULL`), and a date range on `scheduled_at`.

**Given** I am Compliance Officer or Administrator
**When** I land on the Inspections tab
**Then** the default filter shows all inspections for the project
**And** the inspector filter is exposed as an additional dropdown.

**Given** the AG Grid filter model
**When** filters are applied
**Then** they are forwarded to the server in the SSRM payload and applied in SQL — never client-side (FR51, UX-DR28).

---

## Story 3.7: Inspection detail rendered in EntityRail

As any authorized user viewing the Inspections tab,
I want clicking a row to load the Inspection's details into the EntityRail,
So that I see the Inspection's findings and available actions without leaving the tab.

**Acceptance Criteria:**

**Given** I select a row in the Inspections AG Grid
**When** HTMX fires `GET /inspections/<id>`
**Then** the response is an HTML partial rooted at `<section id="inspection-detail" tabindex="-1" role="region" aria-label="Inspection detail">` that loads into the EntityRail
**And** focus moves to `#inspection-detail` (UX-DR31 primary-swap convention).

**Given** the rendered inspection detail
**When** I inspect it
**Then** it shows: trade name, inspector name, scheduled at, status StatusBadge, outcome StatusBadge (or `—` if not completed), notes, a Findings list (one card per finding with severity + description), and an ActionButton row using the trichotomy for "Start Inspection", "Complete Inspection", "Cancel Inspection" — actual transitions land in Stories 3.8 / 3.9 / 3.10.

**Given** at <1280px
**When** the rail collapses
**Then** the inspection detail renders stacked below the list per UX-DR30.

---

## Story 3.8: Start Inspection (Scheduled → InProgress)

As an Inspector,
I want to start an Inspection assigned to me,
So that I can record findings against it (FR17).

**Acceptance Criteria:**

**Given** I am Inspector and the Inspection is in `SCHEDULED` state and assigned to me
**When** I click "Start" on the Inspection detail
**Then** `POST /inspections/<id>/start` runs the canonical flow: authorize (`inspection.start` + assignment scope check) → begin txn → load Inspection → call `inspection.start(actor)` → status → `IN_PROGRESS`, `started_at = now()` → append `AuditEntry(action="InspectionStarted", ...)` → commit
**And** the response body re-renders `#inspection-detail` plus OOB `#audit-log` row (no score recompute yet — completion is what affects score).

**Given** the Inspection is not in `SCHEDULED` state
**When** I click "Start"
**Then** the ActionButton is `disabled-with-tooltip` (UX-DR21) — direct request via DevTools returns HTTP `409` with the originating partial + InlineAlert.

**Given** I am not the assigned inspector
**When** the Start button considers my permission
**Then** the button is `absent` (FR6) — direct request returns HTTP `403`.

---

## Story 3.9: Complete Inspection with findings and auto-open Violations atomically

As an Inspector,
I want to complete an InProgress Inspection with an outcome and zero or more findings,
So that Fail findings automatically open Violations and the project's compliance score updates in the same transaction (FR18, FR20, FR35).

**Acceptance Criteria:**

**Given** the Inspection is in `IN_PROGRESS` and I am the assigned inspector
**When** I open the "Complete" form on Inspection detail
**Then** I see outcome radio (`PASS` / `FAIL` / `CONDITIONAL`), notes textarea, and a repeating finding sub-form (severity select from `domain.violation_category.default_severity`, description, optional category).

**Given** a valid submission
**When** `POST /inspections/<id>/complete` runs
**Then** the canonical flow executes inside one transaction (FR57): authorize → load Inspection (with Findings) → call `inspection.complete(outcome, notes, findings)` → for each `FAIL`-class Finding: call `Violation.open_from_finding(finding, severity, due_date_offset_from_rule)` → INSERT each new `domain.violation` row → append `AuditEntry(action="InspectionCompleted", ...)` plus one `AuditEntry(action="ViolationOpened", entity=violation_id, ...)` per spawned Violation (FR40) → recompute Project compliance score via Story 3.3's function → UPDATE `domain.project.compliance_score` → commit.

**Given** the request succeeds
**When** the response is returned
**Then** the body contains `#inspection-detail` re-rendered (status `COMPLETED`, outcome badge, no Complete button) **plus** OOB `#compliance-tile` with the new score **plus** OOB `#audit-log` with the new rows appended (multiple rows possible if violations were spawned) **plus** OOB `#violation-list` re-rendered if the user is on the Violations tab — three-region orchestration pattern fully exercised (UX-DR20).

**Given** the entity method raises (e.g., Inspection not InProgress, outcome+findings inconsistent — `PASS` outcome with `FAIL` findings)
**When** the exception bubbles
**Then** HTTP `409` with the originating partial + InlineAlert; no DB state changed; no OOB updates emitted (UX-DR22).

**Given** the transaction aborts partway (e.g., DB connection loss after inserting some Violations)
**When** the handler resolves
**Then** the entire transaction rolls back — no orphan Violations, no orphan AuditEntries, no partial score update (FR57, NFR5).

**Given** an E2E Playwright scenario runs against all three stacks
**When** complete-with-fail-findings is exercised
**Then** the three-region paint happens in a single HTTP response on all three stacks with cross-stack timing divergence ≤ 50 ms p95 (NFR1).

---

## Story 3.10: Cancel Scheduled Inspection with reason

As a Compliance Officer or Administrator,
I want to cancel a Scheduled Inspection with a recorded reason,
So that abandoned schedules don't clutter the queue (FR19).

**Acceptance Criteria:**

**Given** the Inspection is in `SCHEDULED` and I am authorized (`inspection.cancel`)
**When** I click "Cancel" on Inspection detail
**Then** an inline form expands capturing a reason; `POST /inspections/<id>/cancel` runs the canonical flow → status → `CANCELLED`, `cancelled_at`, `cancellation_reason` set → append `AuditEntry(action="InspectionCancelled", metadata={reason})` → commit
**And** the response re-renders `#inspection-detail` + OOB `#audit-log` row (no score impact).

**Given** the Inspection is not `SCHEDULED` (already `IN_PROGRESS` / `COMPLETED` / `CANCELLED`)
**When** the entity method raises
**Then** HTTP `409` with originating partial + InlineAlert (UX-DR22)
**And** the Cancel button is `disabled-with-tooltip` whenever the precondition fails (UX-DR21).
