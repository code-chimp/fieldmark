# Epic 4: Violation Lifecycle & Assignment

Site Supervisors see their queue; Compliance Officers assign and reassign; overdue marking lights up; Administrators can void; reopening of terminal states is blocked.

## Story 4.1: Extend Violation data layer with write operations (assign, void, reassign)

As a developer implementing Violation workflow,
I want each stack's data layer to support UPDATE on `domain.violation` for assignment, reassignment, and void,
So that Epic 4's entity methods have persistence.

**Acceptance Criteria:**

**Given** each stack
**When** I inspect its Violation data path
**Then** it supports updating `assigned_to`, `status`, `voided_at`, `void_reason`
**And** `Violation` aggregate methods `assign(supervisor_id)`, `void(reason, actor)`, `reject_resubmission()` exist on the entity (entity methods own the rules; data layer is dumb).

**Given** the canonical DDL CHECK constraints
**When** I attempt to violate them (e.g., voided without reason)
**Then** the DB rejects the write (defense-in-depth — NFR5) and the entity method's pre-check raises before reaching the DB anyway.

---

## Story 4.2: Violation list AG Grid with SSRM endpoint

As any authorized user on the Violations tab of Project Detail,
I want a server-side row-model grid of the project's Violations,
So that I can browse Violations with filter and sort (FR48).

**Acceptance Criteria:**

**Given** the Violations tab is active
**When** the tab content renders
**Then** AGGridPanel initializes against `POST /grid/violations` with `project_id` as a request parameter
**And** rows contain `id`, `category_name`, `severity`, `status`, `assigned_to_name`, `due_date`, `is_overdue` (bool — see Story 4.6), `created_at`.

**Given** I click a row
**When** the row-click handler fires
**Then** `GET /violations/<id>` loads the Violation detail into `#violation-detail` in the EntityRail (FR50)
**And** the grid does not render the detail itself.

**Given** `make parity`
**When** I run it
**Then** `POST /grid/violations` exists on all three stacks identically.

---

## Story 4.3: Violation detail rendered in EntityRail

As any authorized user,
I want to see a selected Violation's detail in the EntityRail,
So that I can view severity, status, assignment, and audit context co-present with the list.

**Acceptance Criteria:**

**Given** I select a row in the Violations grid
**When** HTMX fires `GET /violations/<id>`
**Then** the response is a partial rooted at `<section id="violation-detail" tabindex="-1" role="region" aria-label="Violation detail">` loaded into the EntityRail
**And** focus moves to `#violation-detail` (UX-DR31).

**Given** the rendered Violation detail
**When** I inspect it
**Then** it shows: Severity StatusBadge (with Critical/High bump per UX-DR3), Status StatusBadge, the origin Finding excerpt, due date with `is_overdue` styling if applicable, assignee name, action button row using the trichotomy (`Assign`, `Void`, plus CA-related buttons for Epic 5).

**Given** I am Executive role (FR43)
**When** I view the rail
**Then** all action buttons are `absent` (UX-DR21).

---

## Story 4.4: Assign and reassign Violation to Site Supervisor

As a Compliance Officer,
I want to assign or reassign a Violation to a Site Supervisor,
So that a remediator owns the Violation and the audit trail records each reassignment (FR23).

**Acceptance Criteria:**

**Given** I am authorized (`violation.assign`)
**When** I click "Assign" or "Reassign" on the Violation detail
**Then** an inline form expands with a select of users in the `SITE_SUPERVISOR` role and an optional note; `POST /violations/<id>/assign` runs the canonical flow → call `violation.assign(supervisor_id)` → if the Violation was `OPEN` and is now in scope of a remediator, status stays `OPEN` (assignment doesn't auto-transition); reassignment while in `IN_PROGRESS` is a self-transition that emits an audit entry → append `AuditEntry(action="ViolationAssigned", before={assigned_to: old}, after={assigned_to: new}, metadata={note?})` → commit (FR23).

**Given** the response
**When** it returns
**Then** `#violation-detail` is re-rendered with the new assignee plus OOB `#audit-log` row (no `#compliance-tile` update — assignment doesn't affect score).

**Given** the Violation is in a terminal state (`RESOLVED` / `VOIDED`)
**When** the entity method is called
**Then** it raises and the handler returns HTTP `409` with originating partial + InlineAlert.

---

## Story 4.5: Site Supervisor's assigned-violation queue

As a Site Supervisor,
I want to see Violations assigned to me filterable by status and overdue flag,
So that I know what I need to work on (FR24).

**Acceptance Criteria:**

**Given** I am Site Supervisor and navigate to `/violations/mine`
**When** the page renders
**Then** an AGGridPanel renders a grid filtered server-side by `assigned_to = me`
**And** filter UI exposes status and "Overdue only" toggle (consults Story 4.6's overdue marking).

**Given** the empty queue
**When** I have no assigned violations
**Then** the no-rows overlay reads "No violations assigned to you. Good work." (UX-DR26).

**Given** I am any non-supervisor role
**When** I navigate to `/violations/mine`
**Then** I am redirected to `/dashboard` (the page is gated by role — FR6 collapses the link to absent in nav for non-supervisors).

---

## Story 4.6: Server-rendered overdue marking

As a user viewing any Violation list or detail,
I want overdue Violations visually marked,
So that urgency is communicated without color being the sole carrier (FR25).

**Acceptance Criteria:**

**Given** a Violation with `due_date < now()` and `status NOT IN ('RESOLVED', 'VOIDED')`
**When** any view that includes it renders
**Then** the server computes `is_overdue = true` and emits an "Overdue" text badge alongside the status StatusBadge with `--color-danger` styling — color paired with text (NFR3, WCAG 1.4.1).

**Given** a Violation not yet due, or in terminal state
**When** the same view renders
**Then** no Overdue badge is rendered and `is_overdue = false` in the AG Grid row payload.

**Given** the overdue computation
**When** I grep client code for it
**Then** zero client-side overdue calculation exists — `is_overdue` is always server-computed (FR34, UX-DR28).

**Given** the system time advances past a Violation's `due_date`
**When** the next render occurs
**Then** the Violation begins displaying as overdue without any background job or app restart (purely render-time computation).

---

## Story 4.7: Void Violation (Administrator) with score recompute

As an Administrator,
I want to void a non-terminal Violation with a recorded reason,
So that mistakenly opened or no-longer-applicable Violations can be removed from compliance scoring (FR26, FR35).

**Acceptance Criteria:**

**Given** the Violation is in a non-terminal state and I am authorized (`violation.void`)
**When** I click "Void" on the Violation detail
**Then** an inline form expands capturing a reason; `POST /violations/<id>/void` runs the canonical flow → call `violation.void(reason, actor)` → status → `VOIDED`, `void_reason`, `voided_at` set → append `AuditEntry(action="ViolationVoided", metadata={reason})` → recompute Project compliance score via Story 3.3 → UPDATE `domain.project.compliance_score` → commit (FR57).

**Given** the response
**When** it returns
**Then** `#violation-detail` is re-rendered with terminal state plus OOB `#compliance-tile` with the new score plus OOB `#audit-log` row (UX-DR20 three-region pattern).

**Given** the Violation is already in a terminal state
**When** the entity method is called
**Then** it raises and the handler returns HTTP `409` (UX-DR22) — Void button is `disabled-with-tooltip` whenever the precondition fails (UX-DR21).

**Given** the Voided Violation
**When** the compliance score is recomputed
**Then** it is excluded from scoring inputs (FR26 invariant — verified by Story 3.3's unit tests).

---

## Story 4.8: Block reopening of Resolved or Voided Violations

As a user who would otherwise try to re-open a terminal Violation,
I want the system to prevent it,
So that audit history remains a faithful append-only ledger and users follow the "new finding → new violation" path (FR27).

**Acceptance Criteria:**

**Given** a Violation in `RESOLVED` or `VOIDED` state
**When** I view it on Violation detail
**Then** no transition affordance returns the Violation to `OPEN` or `IN_PROGRESS` — all such ActionButtons are `absent` (UX-DR21, FR6).

**Given** a direct HTTP-level attempt to call any reopening endpoint (which does not exist)
**When** the request is made
**Then** the server returns HTTP `404` (no such route) (FR7).

**Given** I want to revisit the underlying issue
**When** I navigate from the Resolved/Voided Violation detail
**Then** a contextual note in the Violation detail says "To revisit this issue, complete an Inspection with a new Fail finding."
